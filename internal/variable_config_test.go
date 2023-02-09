package internal

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
	"reflect"
	"strings"
	"testing"
)

func Test_partialVariableConfig_coalesceValue(t *testing.T) {
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
		want           map[string]cty.Value
		wantErrorCount int
	}{
		{
			"successfully decode variables and set overrides",
			decodedConfig,
			args{
				variableOverrides,
			},
			map[string]cty.Value{
				"service_address":  cty.StringVal("https://my-server.com"),
				"service_active":   cty.BoolVal(true),
				"service_password": cty.NumberIntVal(1),
				"service_user":     cty.StringVal("joe"),
				"service_other":    cty.NumberIntVal(1),
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
			for k, v := range tt.want {
				if gotVal, ok := got[k]; !ok {
					t.Errorf("CalculatedVariables() expected returned map to have '%s' key, but none exists", k)
				} else {
					if !gotVal.RawEquals(v) {
						t.Errorf("CalculatedVariables() expected map key '%s' to = %s got = %s", k, v.GoString(), gotVal.GoString())
					}
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

func getParsedBody(fileName string) hcl.Body {
	parser := hclparse.NewParser()
	parsedFile, diag := parser.ParseHCLFile(fileName)
	if diag.HasErrors() {
		panic(diag.Errs())
	}
	return parsedFile.Body

}
