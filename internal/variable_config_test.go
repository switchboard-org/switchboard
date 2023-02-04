package internal

import (
	"github.com/hashicorp/hcl/v2"
	"reflect"
	"testing"
)

func Test_getVariableDataFromFile(t *testing.T) {
	type args struct {
		varFile string
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			name: "basic test loading json variable file",
			args: args{varFile: "./../fixtures/variable.json"},
			want: map[string]any{"test": "variable"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVariableDataFromFile(tt.args.varFile); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVariableDataFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_initVariables(t *testing.T) {
	type args struct {
		body         hcl.Body
		variableData map[string]any
	}
	tests := []struct {
		name string
		args args
		want []Variable
	}{
		{
			"load basic variable file",
			args{},
			[]Variable{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := initVariables(tt.args.body, tt.args.variableData)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initVariables() got = %v, want %v", got, tt.want)
			}
		})
	}
}
