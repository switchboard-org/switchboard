package parsecfg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/switchboard-org/switchboard/providers"
	"reflect"
	"testing"
)

type MockDownloader struct {
	packages []providers.Package
}

func NewTestDownloader(packages []providers.Package) providers.Downloader {
	return &MockDownloader{
		packages: packages,
	}
}

func (d *MockDownloader) DownloadedProviders() ([]providers.Package, error) {
	return d.packages, nil
}

func (d *MockDownloader) DownloadProvider(_ string, _ string) error {
	return nil
}

func (d *MockDownloader) ProviderPath(_ string, _ string) string {
	return ".mock-packages"
}

func Test_switchboardBlockParser_init(t *testing.T) {
	type fields struct {
		config     switchboardBlockStepConfig
		downloader providers.Downloader
	}
	type args struct {
		currentVersion string
		ctx            *hcl.EvalContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *SwitchboardBlock
		want1  hcl.Diagnostics
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &switchboardBlockParser{
				config:     tt.fields.config,
				downloader: tt.fields.downloader,
			}
			got, got1 := c.init(tt.args.currentVersion, tt.args.ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("init() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("init() got1 = %v, want %v", got1, tt.want1)
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
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *SwitchboardBlock
		want1  hcl.Diagnostics
	}{
		// TODO: Add test cases.
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
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_switchboardBlockParser_parseRequiredBlocks(t *testing.T) {
	type fields struct {
		config     switchboardBlockStepConfig
		downloader providers.Downloader
	}
	type args struct {
		remain hcl.Body
		ctx    *hcl.EvalContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []requiredProviderData
		want1  hcl.Diagnostics
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &switchboardBlockParser{
				config:     tt.fields.config,
				downloader: tt.fields.downloader,
			}
			got, got1 := c.parseRequiredBlocks(tt.args.remain, tt.args.ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRequiredBlocks() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseRequiredBlocks() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_switchboardBlockParser_parseRequiredPackageBlockStep(t *testing.T) {
	type fields struct {
		config     switchboardBlockStepConfig
		downloader providers.Downloader
	}
	type args struct {
		block requiredProviderContentsConfig
		ctx   *hcl.EvalContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   RequiredProviderBlock
		want1  hcl.Diagnostics
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &switchboardBlockParser{
				config:     tt.fields.config,
				downloader: tt.fields.downloader,
			}
			got, got1 := c.parseRequiredPackageBlockStep(tt.args.block, tt.args.ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRequiredPackageBlockStep() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseRequiredPackageBlockStep() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_switchboardBlockParser_parseRequiredPackageBlocksStep(t *testing.T) {
	type fields struct {
		config     switchboardBlockStepConfig
		downloader providers.Downloader
	}
	type args struct {
		providersBody hcl.Body
		ctx           *hcl.EvalContext
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   requiredProviderBlocksStepConfig
		want1  hcl.Diagnostics
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &switchboardBlockParser{
				config:     tt.fields.config,
				downloader: tt.fields.downloader,
			}
			got, got1 := c.parseRequiredPackageBlocksStep(tt.args.providersBody, tt.args.ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRequiredPackageBlocksStep() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseRequiredPackageBlocksStep() got1 = %v, want %v", got1, tt.want1)
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
		config     switchboardBlockStepConfig
		downloader providers.Downloader
	}

	decodedConfig := getDecodedSwitchboardStepConfig("../fixtures/switchboard_config/basic.hcl")
	fieldsWithCorrectPackages := fields{
		config: decodedConfig,
		downloader: NewTestDownloader([]providers.Package{
			{
				Name:    "test",
				Version: "1.0.0",
			},
			{
				Name:    "test_two",
				Version: "1.0.0",
			},
		}),
	}
	fieldsWithMissingPackage := fields{
		config: decodedConfig,
		downloader: NewTestDownloader([]providers.Package{
			{
				Name:    "test",
				Version: "1.0.0",
			},
		}),
	}
	fieldsWithWrongPackageVersion := fields{
		config: decodedConfig,
		downloader: NewTestDownloader([]providers.Package{
			{
				Name:    "test",
				Version: "1.1.0",
			},
			{
				Name:    "test_two",
				Version: "2.0.0",
			},
		}),
	}
	type args struct {
		packages []requiredProviderData
	}
	testArgs := args{
		packages: []requiredProviderData{
			{
				block: RequiredProviderBlock{
					Name:    "test",
					Source:  "github.com/switchboard-org/test",
					Version: "1.0.0",
				},
				blockRange: testRange(),
			},
			{
				block: RequiredProviderBlock{
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
			c := &switchboardBlockParser{
				config:     tt.fields.config,
				downloader: tt.fields.downloader,
			}
			got := c.verifyPresenceOfPackages(tt.args.packages)
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
