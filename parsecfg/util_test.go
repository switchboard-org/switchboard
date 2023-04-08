package parsecfg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
	"reflect"
	"testing"
)

func Test_findAllFiles(t *testing.T) {
	type args struct {
		root string
		ext  string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "load all hcl files but nothing else",
			args: args{
				root: "../fixtures/basic",
				ext:  ".hcl",
			},
			want: []string{"../fixtures/basic/fake_one.hcl", "../fixtures/basic/fake_two.hcl", "../fixtures/basic/more/more.hcl"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findAllFiles(tt.args.root, tt.args.ext); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findAllFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getVariableDataFromJSONFile(t *testing.T) {
	type args struct {
		varFile string
	}
	tests := []struct {
		name string
		args args
		want map[string]cty.Value
	}{
		{
			name: "load and parsecfg provided json variable file",
			args: args{varFile: "./../fixtures/variable.json"},
			want: map[string]cty.Value{"test": cty.StringVal("variable")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVariableDataFromJSONFile(tt.args.varFile); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVariableDataFromJSONFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loadAllHclFilesInDir(t *testing.T) {
	spec := hcldec.ObjectSpec{
		"fakes": &hcldec.BlockSpec{
			TypeName: "fake",
			Nested: hcldec.ObjectSpec{
				"value": &hcldec.AttrSpec{
					Name:     "value",
					Type:     cty.String,
					Required: true,
				},
			},
		},
		"others": &hcldec.BlockSpec{
			TypeName: "other",
			Nested: hcldec.ObjectSpec{
				"value": &hcldec.AttrSpec{
					Name:     "value",
					Type:     cty.String,
					Required: true,
				},
			},
		},
		"mores": &hcldec.BlockSpec{
			TypeName: "more",
			Nested: hcldec.ObjectSpec{
				"value": &hcldec.AttrSpec{
					Name:     "value",
					Type:     cty.String,
					Required: true,
				},
			},
		},
	}

	tests := []struct {
		name string
	}{
		{
			name: "load all config files in current and children directories",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diag := loadAllHclFilesInDir(fmt.Sprint("../fixtures/basic"))
			if diag.HasErrors() {
				for _, err := range diag.Errs() {
					t.Logf(err.Error())
				}
				t.Fail()
			}
			_, diag = hcldec.Decode(result, spec, nil)
			if diag.HasErrors() {
				for _, err := range diag.Errs() {
					t.Logf(err.Error())
				}
				t.Fail()
			}
		})
	}
}
