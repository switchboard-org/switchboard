package parsecfg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/switchboard-org/switchboard/internal"
	"github.com/switchboard-org/switchboard/providers"
	"reflect"
	"testing"
)

type MockDownloader struct {
	packages           []providers.Package
	packagesDownloaded int
}

func NewTestDownloader(packages []providers.Package) *MockDownloader {
	return &MockDownloader{
		packages:           packages,
		packagesDownloaded: 0,
	}
}

func (d *MockDownloader) DownloadedProviders() ([]providers.Package, error) {
	return d.packages, nil
}

func (d *MockDownloader) DownloadProvider(_ string, _ string) error {
	d.packagesDownloaded += 1
	return nil
}

func (d *MockDownloader) ProviderPath(_ string, _ string) string {
	return ".mock-packages"
}

func (d *MockDownloader) GetDownloadCount() int {
	return d.packagesDownloaded
}

type TestOsManager struct{}

func NewTestOsManager() internal.OsManager {
	return &TestOsManager{}
}

func (d *TestOsManager) GetCurrentWorkingDirectory() string {
	return internal.CurrentWorkingDir()
}

func (d *TestOsManager) CreateDirectoryIfNotExists(path string) error {
	return nil
}

func Test_switchboardBlockParser_init(t *testing.T) {
	type fields struct {
		config     switchboardBlockStepConfig
		downloader MockDownloader
		osManager  internal.OsManager
	}
	type args struct {
		currentVersion string
		ctx            *hcl.EvalContext
	}
	matchingFields := fields{
		config: getDecodedSwitchboardStepConfig("../fixtures/switchboard_config/basic.hcl"),
		downloader: *NewTestDownloader([]providers.Package{
			{
				Name:    "provider-test",
				Version: "1.0.0",
			},
			{
				Name:    "provider-test-two",
				Version: "1.0.0",
			},
		}),
		osManager: NewTestOsManager(),
	}
	missingFields := fields{
		config:     getDecodedSwitchboardStepConfig("../fixtures/switchboard_config/basic.hcl"),
		downloader: *NewTestDownloader([]providers.Package{}),
		osManager:  NewTestOsManager(),
	}

	basicArgs := args{
		currentVersion: "1.0.0",
		ctx:            nil,
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantDiagCount     int
		wantDownloadCount *int
	}{
		{
			name:          "should run successfully if packages are present",
			fields:        matchingFields,
			args:          basicArgs,
			wantDiagCount: 0,
		},
		{
			name:              "should succeed and download packages that are missing",
			fields:            missingFields,
			args:              basicArgs,
			wantDiagCount:     0,
			wantDownloadCount: internal.Ptr(2),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &switchboardBlockParser{
				config:     tt.fields.config,
				downloader: &tt.fields.downloader,
				osManager:  tt.fields.osManager,
			}
			_, got1 := c.init(tt.args.currentVersion, tt.args.ctx)
			if len(got1.Errs()) != tt.wantDiagCount {
				t.Errorf("init() error count = %v, want %v", len(got1.Errs()), tt.wantDiagCount)
			}
			if tt.wantDownloadCount != nil {
				if tt.fields.downloader.GetDownloadCount() != *tt.wantDownloadCount {
					t.Errorf("init() packages downloaded: %v, wanted: %v", tt.fields.downloader.GetDownloadCount(), *tt.wantDownloadCount)
				}
			}
		})
	}
}

func Test_switchboardBlockParser_parse(t *testing.T) {
	type fields struct {
		config     switchboardBlockStepConfig
		downloader providers.Downloader
	}
	type args struct {
		currentVersion        string
		ctx                   *hcl.EvalContext
		shouldVerifyDownloads bool
	}
	matchedFields := fields{
		config: getDecodedSwitchboardStepConfig("../fixtures/switchboard_config/basic.hcl"),
		downloader: NewTestDownloader([]providers.Package{
			{
				Name:    "provider-test",
				Version: "1.0.0",
			},
			{
				Name:    "provider-test-two",
				Version: "1.0.0",
			},
		}),
	}
	missingFields := fields{
		config:     getDecodedSwitchboardStepConfig("../fixtures/switchboard_config/basic.hcl"),
		downloader: NewTestDownloader([]providers.Package{}),
	}

	fullOutput := internal.SwitchboardBlock{
		Version: "~> 1.0",
		RequiredProviders: []internal.RequiredProviderBlock{
			{
				Name:    "test",
				Source:  "github.com/switchboard-org/provider-test",
				Version: "1.0.0",
			},
			{
				Name:    "test_two",
				Source:  "github.com/switchboard-org/provider-test-two",
				Version: "1.0.0",
			},
		},
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		want          *internal.SwitchboardBlock
		wantDiagCount int
	}{
		{
			name:   "should parse and verify packages are present and succeed",
			fields: matchedFields,
			args: args{
				currentVersion:        "1.0.0",
				ctx:                   nil,
				shouldVerifyDownloads: true,
			},
			want:          &fullOutput,
			wantDiagCount: 0,
		},
		{
			name:   "should parse and not verify packages if verify set to false",
			fields: missingFields,
			args: args{
				currentVersion:        "1.0.0",
				ctx:                   nil,
				shouldVerifyDownloads: false,
			},
			want:          &fullOutput,
			wantDiagCount: 0,
		},
		{
			name:   "should fail if packages missing and verifying package presence",
			fields: missingFields,
			args: args{
				currentVersion:        "1.0.0",
				ctx:                   nil,
				shouldVerifyDownloads: true,
			},
			want:          nil,
			wantDiagCount: 2,
		},
		{
			name:   "should fail if version constraint not met",
			fields: matchedFields,
			args: args{
				currentVersion:        "2.0.0",
				ctx:                   nil,
				shouldVerifyDownloads: false,
			},
			want:          nil,
			wantDiagCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &switchboardBlockParser{
				config:     tt.fields.config,
				downloader: tt.fields.downloader,
			}
			got, got1 := c.parse(tt.args.currentVersion, tt.args.ctx, tt.args.shouldVerifyDownloads)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() got = %v, want %v", got, tt.want)
			}
			if len(got1.Errs()) != tt.wantDiagCount {
				t.Errorf("parse() error count = %v, want %v", len(got1.Errs()), tt.wantDiagCount)
			}
		})
	}
}

func Test_switchboardBlockParser_parseRequiredBlocks(t *testing.T) {
	type args struct {
		remain hcl.Body
		ctx    *hcl.EvalContext
	}
	decodedConfig := getDecodedSwitchboardStepConfig("../fixtures/switchboard_config/basic.hcl")
	tests := []struct {
		name string

		args  args
		want  []internal.RequiredProviderBlock
		want1 hcl.Diagnostics
	}{
		{
			name: "should parse required blocks",
			args: args{
				remain: decodedConfig.Switchboard.Remain,
			},
			want: []internal.RequiredProviderBlock{
				{
					Name:    "test",
					Source:  "github.com/switchboard-org/provider-test",
					Version: "1.0.0",
				},
				{
					Name:    "test_two",
					Source:  "github.com/switchboard-org/provider-test-two",
					Version: "1.0.0",
				},
			},
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseRequiredBlocks(tt.args.remain, tt.args.ctx)
			var gotBlocks []internal.RequiredProviderBlock
			for _, d := range got {
				gotBlocks = append(gotBlocks, d.block)
			}
			if !reflect.DeepEqual(gotBlocks, tt.want) {
				t.Errorf("parseRequiredBlocks() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseRequiredBlocks() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_switchboardBlockParser_parseVersion(t *testing.T) {
	type fields struct {
		config     switchboardBlockStepConfig
		downloader providers.Downloader
	}

	decodedConfig := getDecodedSwitchboardStepConfig("../fixtures/switchboard_config/basic.hcl")
	testFields := fields{
		config:     decodedConfig,
		downloader: NewTestDownloader([]providers.Package{}),
	}
	type args struct {
		expectedVersion hcl.Expression
		currentVersion  string
		ctx             *hcl.EvalContext
	}
	exactMatchVersionArgs := args{
		expectedVersion: decodedConfig.Switchboard.Version,
		currentVersion:  "1.0.0",
		ctx:             nil,
	}
	validMinorVersionArgs := args{
		expectedVersion: decodedConfig.Switchboard.Version,
		currentVersion:  "1.1.0",
		ctx:             nil,
	}
	invalidVersionArgs := args{
		expectedVersion: decodedConfig.Switchboard.Version,
		currentVersion:  "2.0.0",
		ctx:             nil,
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       string
		errorCount int
	}{
		{
			"succeeds with exact version match",
			testFields,
			exactMatchVersionArgs,
			"~> 1.0",
			0,
		},
		{
			"succeeds with minor version match",
			testFields,
			validMinorVersionArgs,
			"~> 1.0",
			0,
		},
		{
			"returns error with invalid version match",
			testFields,
			invalidVersionArgs,
			"~> 1.0",
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &switchboardBlockParser{
				config:     tt.fields.config,
				downloader: tt.fields.downloader,
			}
			got, got1 := c.parseVersion(tt.args.expectedVersion, tt.args.currentVersion, tt.args.ctx)
			if got != tt.want {
				t.Errorf("parseVersion() got = %v, want %v", got, tt.want)
			}
			errCount := len(got1.Errs())
			if errCount != tt.errorCount {
				t.Errorf("Expected %v errors but got %v", tt.errorCount, errCount)
			}
		})
	}
}

func Test_switchboardBlockParser_verifyPresenceOfPackages(t *testing.T) {
	type fields struct {
		config   switchboardBlockStepConfig
		packages []providers.Package
	}

	decodedConfig := getDecodedSwitchboardStepConfig("../fixtures/switchboard_config/basic.hcl")
	fieldsWithCorrectPackages := fields{
		config: decodedConfig,
		packages: []providers.Package{
			{
				Name:    "test",
				Version: "1.0.0",
			},
			{
				Name:    "test_two",
				Version: "1.0.0",
			},
		},
	}
	fieldsWithMissingPackage := fields{
		config: decodedConfig,
		packages: []providers.Package{
			{
				Name:    "test",
				Version: "1.0.0",
			},
		},
	}
	fieldsWithWrongPackageVersion := fields{
		config: decodedConfig,
		packages: []providers.Package{
			{
				Name:    "test",
				Version: "1.1.0",
			},
			{
				Name:    "test_two",
				Version: "2.0.0",
			},
		},
	}
	type args struct {
		packages []requiredProviderData
	}
	testArgs := args{
		packages: []requiredProviderData{
			{
				block: internal.RequiredProviderBlock{
					Name:    "test",
					Source:  "github.com/switchboard-org/test",
					Version: "1.0.0",
				},
				blockRange: testRange(),
			},
			{
				block: internal.RequiredProviderBlock{
					Name:    "test_two",
					Source:  "github.com/switchboard-org/test_two",
					Version: "1.0.0",
				},
				blockRange: testRange(),
			},
		},
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		errorCount int
	}{
		{
			"succeeds when all packages are present",
			fieldsWithCorrectPackages,
			testArgs,
			0,
		},
		{
			"returns an error diagonostic when a package is missing",
			fieldsWithMissingPackage,
			testArgs,
			1,
		},
		{
			"returns two errors when package versions are incorrect",
			fieldsWithWrongPackageVersion,
			testArgs,
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := verifyPresenceOfPackages(tt.fields.packages, tt.args.packages)
			errCount := len(got.Errs())
			if errCount != tt.errorCount {
				t.Errorf("Expected %v errors but got %v", tt.errorCount, errCount)
			}
		})
	}
}

func getDecodedSwitchboardStepConfig(fileName string) switchboardBlockStepConfig {
	var configOutput switchboardBlockStepConfig
	err := hclsimple.DecodeFile(fileName, nil, &configOutput)
	if err != nil {
		panic(err)
	}
	return configOutput
}

func testRange() hcl.Range {
	return hcl.Range{
		Filename: "test/file.hcl",
		Start: hcl.Pos{
			Line:   1,
			Column: 1,
			Byte:   1,
		},
		End: hcl.Pos{
			Line:   10,
			Column: 10,
			Byte:   10,
		},
	}
}
