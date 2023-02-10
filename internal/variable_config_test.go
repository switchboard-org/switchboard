package internal

import (
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/switchboard-org/switchboard/internal/shared"
	"github.com/zclconf/go-cty/cty"
	"reflect"
	"strings"
	"testing"
)

func Test_partialVariableConfig_coalescedValue(t *testing.T) {
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
			got, err := pv.coalescedValue(tt.args.varType, tt.args.defaultValue, tt.args.newValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("coalescedValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("coalescedValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_partialVariableConfig_evaluationContext(t *testing.T) {
	decodedConfig := getDecodedVariableStepConfig("../fixtures/variable_config/variables.hcl")
	partial := decodedConfig.Variables[0]
	t.Run("should return expected variables and functions used for variable processing", func(t *testing.T) {
		got := partial.evaluationContext()
		if got.Variables["number"] != cty.StringVal("number") ||
			got.Variables["string"] != cty.StringVal("string") ||
			got.Variables["boolean"] != cty.StringVal("boolean") {
			t.Errorf("evaluationContext() did not provider the expected variables")
		}
		calledListFunc, err := got.Functions["list"].Call([]cty.Value{cty.StringVal("string")})
		if err != nil || calledListFunc.AsString() != "list(string)" {
			t.Errorf("evaluationContext() did not provide a function named list, or it is implemented incorrectly")
		}
	})

}

func Test_partialVariableConfig_defaultValue(t *testing.T) {
	type args struct {
		varType cty.Type
	}

	decodedConfig := getDecodedVariableStepConfig("../fixtures/variable_config/variables.hcl")
	partialStringWithDefault := decodedConfig.Variables[0]
	partialBoolWithoutDefault := decodedConfig.Variables[1]
	tests := []struct {
		name                 string
		fields               partialVariableConfig
		args                 args
		want                 cty.Value
		wantErrorCount       int
		errorMessageIncludes []string
	}{
		{
			"decode default into value",
			partialStringWithDefault,
			args{cty.String},
			cty.StringVal("https://my-server.com"),
			0,
			nil,
		},
		{
			"decode default even if none provided (optionality), but as null value",
			partialBoolWithoutDefault,
			args{cty.Bool},
			cty.NullVal(cty.Bool),
			0,
			nil,
		},
		{
			"return diagnostic if default value is not the same type as expected",
			partialStringWithDefault,
			args{cty.Bool},
			cty.UnknownVal(cty.Bool),
			1,
			[]string{"Incorrect attribute value type"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pv := tt.fields
			got, got1 := pv.defaultValue(tt.args.varType)
			if !got.RawEquals(tt.want) {
				t.Errorf("defaultValue() got = %v, want %v", got.GoString(), tt.want.GoString())
			}
			if len(got1.Errs()) != tt.wantErrorCount {
				t.Errorf("defaultValue() expected error count = %v, want %v", len(got1.Errs()), tt.wantErrorCount)
			}
			if len(tt.errorMessageIncludes) > 0 {
				for i, msg := range got1.Errs() {
					if !strings.Contains(msg.Error(), tt.errorMessageIncludes[i]) {
						t.Errorf("calculatedType() expected error message '%s' at index '%v' to contain '%s', but did not", msg, i, tt.errorMessageIncludes[i])
					}
				}
			}
		})
	}
}

func Test_partialVariableConfig_calculatedType(t *testing.T) {
	decodedConfig := getDecodedVariableStepConfig("../fixtures/variable_config/variables.hcl")
	partialStringWithDefault := decodedConfig.Variables[0]

	badDecodedConfig := getDecodedVariableStepConfig("../fixtures/variable_config/variables_invalid_types.hcl")
	partialWithTypeSetToString := badDecodedConfig.Variables[0]
	partialWithTypeSetToInvalidExpression := badDecodedConfig.Variables[1]
	tests := []struct {
		name                 string
		fields               partialVariableConfig
		want                 cty.Type
		wantErrorCount       int
		errorMessageIncludes []string
	}{
		{
			"return error if 'type' field is not an hcl expression",
			partialWithTypeSetToString,
			cty.NilType,
			1,
			[]string{"Incorrect attribute value type"},
		},
		{
			"return error if 'type' field is an invalid hcl expression",
			partialWithTypeSetToInvalidExpression,
			cty.NilType,
			1,
			[]string{"Unknown variable"},
		},
		{
			"return the string value of the type if correct expression used",
			partialStringWithDefault,
			cty.String,
			0,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pv := tt.fields
			got, got1 := pv.calculatedType()
			if !got.Equals(tt.want) {
				t.Errorf("calculatedType() got = %v, want %v", got.GoString(), tt.want.GoString())
			}
			if len(got1.Errs()) != tt.wantErrorCount {
				t.Errorf("calculatedType() expected error count = %v, want %v", len(got1.Errs()), tt.wantErrorCount)
			}
			if len(tt.errorMessageIncludes) > 0 {
				for i, msg := range got1.Errs() {
					if !strings.Contains(msg.Error(), tt.errorMessageIncludes[i]) {
						t.Errorf("calculatedType() expected error message '%s' at index '%v' to contain '%s', but did not", msg, i, tt.errorMessageIncludes[i])
					}
				}
			}
		})
	}
}

func Test_variableStepConfig_CalculatedVariables(t *testing.T) {

	variableOverrides := getVariableDataFromJSONFile("../fixtures/variable_config/overrides.json")
	decodedConfig := getDecodedVariableStepConfig("../fixtures/variable_config/variables.hcl")
	type args struct {
		overrides map[string]cty.Value
	}
	tests := []struct {
		name           string
		fields         variableStepConfig
		args           args
		want           []shared.Variable
		wantErrorCount int
	}{
		{
			"successfully decode variables and set overrides",
			decodedConfig,
			args{
				variableOverrides,
			},
			[]shared.Variable{
				{
					Name:  "service_address",
					Type:  cty.String,
					Value: cty.StringVal("https://my-server.com"),
				},
				{
					Name:  "service_active",
					Type:  cty.Bool,
					Value: cty.BoolVal(true),
				},
				{
					Name:  "service_password",
					Type:  cty.Number,
					Value: cty.NumberIntVal(1),
				},
				{
					Name:  "service_user",
					Type:  cty.String,
					Value: cty.StringVal("joe"),
				},
				{
					Name:  "service_other",
					Type:  cty.Number,
					Value: cty.NumberIntVal(1),
				},
			},
			0,
		},
		{
			"should return multiple diagnostics for all issues (no provided default or override, in this case)",
			decodedConfig,
			args{
				nil,
			},
			nil,
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &variableStepConfig{
				Variables: tt.fields.Variables,
				Remain:    tt.fields.Remain,
			}
			got, diag := v.CalculatedVariables(tt.args.overrides)
			if len(got) != len(tt.want) {
				t.Errorf("CalculatedVariables() expected to have a map with %v keys. Got %v", len(tt.want), len(got))
			}
			for _, k := range tt.want {
				match := false
				for _, k2 := range got {
					if k2.Name == k.Name {
						match = true
						if k2.Type != k.Type {
							t.Errorf("CalculatedVariables() expected variable '%s' to have a type = %s. Got %s", k.Name, k.Type.FriendlyName(), k2.Type.FriendlyName())
						}
						if !k2.Value.RawEquals(k.Value) {
							t.Errorf("CalculatedVariables() expected variable '%s' value to = %s got = %s", k.Name, k.Value.GoString(), k2.Value.GoString())
						}
					}
				}
				if !match {
					t.Errorf("CalculatedVariables() expected to have a shared.Variable record with Name = %s.", k.Name)
				}
			}
			if len(diag.Errs()) != tt.wantErrorCount {
				t.Errorf("CalculatedVariables() expected to have %v errors. Got %v", tt.wantErrorCount, len(diag.Errs()))
			}
		})
	}
}

func getDecodedVariableStepConfig(fileName string) variableStepConfig {
	var configOutput variableStepConfig
	err := hclsimple.DecodeFile(fileName, nil, &configOutput)
	if err != nil {
		panic(err)
	}
	return configOutput
}
