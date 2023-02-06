package internal

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
	"reflect"
	"testing"
)

func getDecodedVariableStepConfig(fileName string) variableStepConfig {
	var configOutput variableStepConfig
	err := hclsimple.DecodeFile(fileName, nil, &configOutput)
	if err != nil {
		panic(err)
	}
	return configOutput
}

func Test_partialVariableConfig_CoalesceValue(t *testing.T) {
	type args struct {
		varType      cty.Type
		defaultValue cty.Value
		newValue     cty.Value
	}

	decodedConfig := getDecodedVariableStepConfig("../fixtures/variable_config/variables.hcl")
	partial := decodedConfig.Variables[0]
	tests := []struct {
		name    string
		config  partialVariableConfig
		args    args
		want    cty.Value
		wantErr bool
	}{
		{
			"override default",
			partial,
			args{
				varType:      cty.String,
				defaultValue: cty.StringVal("test"),
				newValue:     cty.StringVal("override"),
			},
			cty.StringVal("override"),
			false,
		},
		{
			"use default if no override",
			partial,
			args{
				varType:      cty.String,
				defaultValue: cty.StringVal("test"),
				newValue:     cty.NilVal,
			},
			cty.StringVal("test"),
			false,
		},
		{
			"throw error if no override or default",
			partial,
			args{
				varType:      cty.String,
				defaultValue: cty.NilVal,
				newValue:     cty.NilVal,
			},
			cty.NilVal,
			true,
		},
		{
			"throw error if override is invalid type",
			partial,
			args{
				varType:      cty.String,
				defaultValue: cty.StringVal("test"),
				newValue:     cty.BoolVal(true),
			},
			cty.NilVal,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pv := tt.config
			got, err := pv.CoalesceValue(tt.args.varType, tt.args.defaultValue, tt.args.newValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("CoalesceValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CoalesceValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_partialVariableConfig_EvaluationContext(t *testing.T) {
	decodedConfig := getDecodedVariableStepConfig("../fixtures/variable_config/variables.hcl")
	partial := decodedConfig.Variables[0]
	t.Run("should return expected variables and functions used for variable processing", func(t *testing.T) {
		got := partial.EvaluationContext()
		if got.Variables["number"] != cty.StringVal("number") ||
			got.Variables["string"] != cty.StringVal("string") ||
			got.Variables["boolean"] != cty.StringVal("boolean") {
			t.Errorf("EvaluationContext() did not provider the expected variables")
		}
		calledListFunc, err := got.Functions["list"].Call([]cty.Value{cty.StringVal("string")})
		if err != nil || calledListFunc.AsString() != "list(string)" {
			t.Errorf("EvaluationContext() did not provide a function named list, or it is implemented incorrectly")
		}
	})

}

func Test_partialVariableConfig_GetDefaultValue(t *testing.T) {
	type fields struct {
		Name   string
		Type   hcl.Expression
		Remain hcl.Body
	}
	type args struct {
		varType cty.Type
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   cty.Value
		want1  []error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pv := &partialVariableConfig{
				Name:   tt.fields.Name,
				Type:   tt.fields.Type,
				Remain: tt.fields.Remain,
			}
			got, got1 := pv.GetDefaultValue(tt.args.varType)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDefaultValue() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetDefaultValue() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_partialVariableConfig_ValidateAndGetType(t *testing.T) {
	type fields struct {
		Name   string
		Type   hcl.Expression
		Remain hcl.Body
	}
	tests := []struct {
		name   string
		fields fields
		want   cty.Type
		want1  []error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pv := &partialVariableConfig{
				Name:   tt.fields.Name,
				Type:   tt.fields.Type,
				Remain: tt.fields.Remain,
			}
			got, got1 := pv.ValidateAndGetType()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateAndGetType() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ValidateAndGetType() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_variableStepConfig_Decode(t *testing.T) {
	type fields struct {
		Variables []partialVariableConfig
		Remain    hcl.Body
	}
	type args struct {
		body hcl.Body
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &variableStepConfig{
				Variables: tt.fields.Variables,
				Remain:    tt.fields.Remain,
			}
			v.Decode(tt.args.body)
		})
	}
}

func Test_variableStepConfig_GetAllVariables(t *testing.T) {
	type fields struct {
		Variables []partialVariableConfig
		Remain    hcl.Body
	}
	type args struct {
		overrides map[string]cty.Value
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]cty.Value
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &variableStepConfig{
				Variables: tt.fields.Variables,
				Remain:    tt.fields.Remain,
			}
			if got := v.GetAllVariables(tt.args.overrides); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}
