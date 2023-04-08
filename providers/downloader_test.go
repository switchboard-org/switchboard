package providers

import (
	"os"
	"reflect"
	"testing"
)

func Test_downloader_distName(t *testing.T) {
	type fields struct {
		os            string
		arch          string
		packageFolder string
	}
	setFields := fields{
		os:            "linux",
		arch:          "amd64",
		packageFolder: ".",
	}
	type args struct {
		location string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			"creates proper package dist name",
			setFields,
			args{
				"github.com/switchboard-org/provider-test",
			},
			"provider-test_linux_x86_64.tar.gz",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				os:            tt.fields.os,
				arch:          tt.fields.arch,
				packageFolder: tt.fields.packageFolder,
			}
			got, err := d.distName(tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("distName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("distName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloader_downloadPackage(t *testing.T) {
	type fields struct {
		os            string
		arch          string
		packageFolder string
	}
	setFields := fields{
		os:            "linux",
		arch:          "amd64",
		packageFolder: "./packages",
	}
	type args struct {
		location string
		version  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"downloads appropriate dist",
			setFields,
			args{
				location: "github.com/switchboard-org/provider-stripe",
				version:  "0.0.3",
			},
			false,
		},
		{
			"throw error for nonexistent dist",
			setFields,
			args{
				location: "github.com/switchboard-org/provider-stripe",
				version:  "0.0.1",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				os:            tt.fields.os,
				arch:          tt.fields.arch,
				packageFolder: tt.fields.packageFolder,
			}
			if err := d.downloadPackage(tt.args.location, tt.args.version); (err != nil) != tt.wantErr {
				t.Errorf("downloadPackage() error = %v, wantErr %v", err, tt.wantErr)
			}
			os.RemoveAll(tt.fields.packageFolder)
		})
	}
}

func Test_downloader_downloadedPackageList(t *testing.T) {
	type fields struct {
		os            string
		arch          string
		packageFolder string
	}
	setFields := fields{
		os:            "linux",
		arch:          "amd64",
		packageFolder: "./packages",
	}

	tests := []struct {
		name    string
		fields  fields
		want    []Package
		wantErr bool
	}{
		{
			"list two packages downloaded",
			setFields,
			[]Package{
				{
					Name:    "provider-other",
					Version: "1.0.0",
				},
				{
					Name:    "provider-stripe",
					Version: "0.0.3",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		os.MkdirAll("./packages/provider-stripe/0.0.3", 0750)
		os.MkdirAll("./packages/provider-other/1.0.0", 0750)
		os.Create("./packages/provider-stripe/0.0.3/plugin")
		os.Create("./packages/provider-other/1.0.0/plugin")
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				os:            tt.fields.os,
				arch:          tt.fields.arch,
				packageFolder: tt.fields.packageFolder,
			}
			got, err := d.downloadedPackageList()
			if (err != nil) != tt.wantErr {
				t.Errorf("downloadedPackageList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("downloadedPackageList() got = %v, want %v", got, tt.want)
			}
		})
		os.RemoveAll(setFields.packageFolder)
	}
}

func Test_downloader_packageIsDownloaded(t *testing.T) {
	type fields struct {
		os            string
		arch          string
		packageFolder string
	}
	setFields := fields{
		os:            "linux",
		arch:          "amd64",
		packageFolder: "./packages",
	}
	type args struct {
		location string
		version  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			"returns true if downloaded",
			setFields,
			args{
				"github.com/switchboard-org/provider-stripe",
				"0.0.3",
			},
			true,
			false,
		},
		{
			"returns false if version is not matched",
			setFields,
			args{
				"github.com/switchboard-org/provider-stripe",
				"0.0.4",
			},
			false,
			false,
		},
		{
			"returns false if package is not installed",
			setFields,
			args{
				"github.com/switchboard-org/provider-other",
				"0.0.4",
			},
			false,
			false,
		},
	}
	os.MkdirAll("./packages/provider-stripe/0.0.3", 0750)
	os.Create("./packages/provider-stripe/0.0.3/plugin")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				os:            tt.fields.os,
				arch:          tt.fields.arch,
				packageFolder: tt.fields.packageFolder,
			}
			got, err := d.packageIsDownloaded(tt.args.location, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("packageIsDownloaded() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("packageIsDownloaded() got = %v, want %v", got, tt.want)
			}
		})
	}
	os.RemoveAll(setFields.packageFolder)
}

func Test_downloader_packagePath(t *testing.T) {
	type fields struct {
		os            string
		arch          string
		packageFolder string
	}
	type args struct {
		location string
		version  string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			"should return error with invalid arch",
			fields{
				os:            "darwin",
				arch:          "amd64",
				packageFolder: "./packages",
			},
			args{
				location: "github.com/switchboard-org/provider-stripe",
				version:  "0.0.3",
			},
			"./packages/provider-stripe/0.0.3/plugin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				os:            tt.fields.os,
				arch:          tt.fields.arch,
				packageFolder: tt.fields.packageFolder,
			}
			if got := d.packagePath(tt.args.location, tt.args.version); got != tt.want {
				t.Errorf("packagePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloader_processorArchitecture(t *testing.T) {
	type fields struct {
		os            string
		arch          string
		packageFolder string
	}

	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			"should return error with invalid arch",
			fields{
				os:            "",
				arch:          "badarch",
				packageFolder: "./",
			},
			"",
			true,
		},
		{
			"should succeed if valid arch arch",
			fields{
				os:            "",
				arch:          "amd64",
				packageFolder: "./",
			},
			"x86_64",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				os:            tt.fields.os,
				arch:          tt.fields.arch,
				packageFolder: tt.fields.packageFolder,
			}
			got, err := d.processorArchitecture()
			if (err != nil) != tt.wantErr {
				t.Errorf("processorArchitecture() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("processorArchitecture() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloader_systemOs(t *testing.T) {
	type fields struct {
		os            string
		arch          string
		packageFolder string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			"should return error with invalid os",
			fields{
				os:            "bados",
				arch:          "",
				packageFolder: "./",
			},
			"",
			true,
		},
		{
			"should return error with invalid arch",
			fields{
				os:            "darwin",
				arch:          "",
				packageFolder: "./",
			},
			"darwin",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				os:            tt.fields.os,
				arch:          tt.fields.arch,
				packageFolder: tt.fields.packageFolder,
			}
			got, err := d.systemOs()
			if (err != nil) != tt.wantErr {
				t.Errorf("systemOs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("systemOs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
