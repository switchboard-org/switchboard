package internal

import (
	"errors"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type VariableStepConfig struct {
	Variables []PartialVariableConfig `hcl:"variable,block"`
	Remain    hcl.Body                `hcl:",remain"`
}

type PartialVariableConfig struct {
	Name   string         `hcl:"name,label"`
	Type   hcl.Expression `hcl:"type"`
	Remain hcl.Body       `hcl:",remain"`
}

func (v *VariableStepConfig) Decode(body hcl.Body) {
	diag := gohcl.DecodeBody(body, nil, v)

	if diag.HasErrors() {
		displayErrorsAndThrow(diag.Errs())
	}
}

// GetAllVariables takes the provided variable configuration blocks that each have an optional default value,
// along with a map of discrete override values, and will return a map with coalesced values, with overrides superseding defaults.
// Will throw an error if a variable has no default or override.
func (v *VariableStepConfig) GetAllVariables(overrides map[string]cty.Value) map[string]cty.Value {
	output := make(map[string]cty.Value)
	//variables will be built up in here and thrown at end if needed.
	var errorList []error

	for _, partial := range v.Variables {
		//check that var types are valid
		varType, errs := partial.ValidateAndGetType()
		if len(errs) > 0 {
			errorList = append(errorList, errs...)
			continue
		}
		//check that defaults are correct type
		defaultVal, errs := partial.GetDefaultValue(varType)
		if len(errs) > 0 {
			errorList = append(errorList, errs...)
			continue
		}

		overrideVal := overrides[partial.Name]
		//override defaults if necessary
		variableValue, err := partial.CoalesceValue(varType, defaultVal, overrideVal)
		if err != nil {
			errorList = append(errorList, err)
			continue
		}
		output[partial.Name] = variableValue
	}
	if len(errorList) > 0 {
		displayErrorsAndThrow(errorList)
	}
	return output
}

// EvaluationContext provides a number of context variables and functions
// used in defining custom variables in the configuration.
func (pv *PartialVariableConfig) EvaluationContext() hcl.EvalContext {
	evalVars := map[string]cty.Value{
		"number":  cty.StringVal("number"),
		"string":  cty.StringVal("string"),
		"boolean": cty.StringVal("boolean"),
	}
	var listFunc = function.New(&function.Spec{
		Description: `used to generate a list type for variables`,
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.String,
				AllowDynamicType: false,
				AllowMarked:      false,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			return cty.StringVal(fmt.Sprint("list(", args[0].AsString(), ")")), nil
		},
	})
	evalFuncs := map[string]function.Function{
		"list": listFunc,
	}
	return hcl.EvalContext{
		Variables: evalVars,
		Functions: evalFuncs,
	}
}

func (pv *PartialVariableConfig) ValidateAndGetType() (cty.Type, []error) {
	varList := pv.Type.Variables()
	if len(varList) == 0 {
		return cty.NilType, []error{errors.New(fmt.Sprintf("%s: Incorrect attribute value type; Inappropriate value for attribute \"type\": variable required (hint: don't use a primitive)", pv.Type.Range()))}
	}
	varEvalContext := pv.EvaluationContext()
	typeValue, diag := pv.Type.Value(&varEvalContext)
	if diag.HasErrors() {
		return cty.NilType, diag.Errs()
	}
	var variableType cty.Type

	switch typeValue.AsString() {
	case "string":
		variableType = cty.String
	case "number":
		variableType = cty.Number
	case "boolean":
		variableType = cty.Bool
	case "list(number)":
		variableType = cty.List(cty.Number)
	case "list(string)":
		variableType = cty.List(cty.String)
	case "list(boolean)":
		variableType = cty.List(cty.Bool)
	}
	//shouldn't be possible to get here as the variables/functions in the evaluation context shouldn't allow it.
	return variableType, nil
}

func (pv *PartialVariableConfig) GetDefaultValue(varType cty.Type) (cty.Value, []error) {
	spec := hcldec.AttrSpec{
		Name:     "default",
		Type:     varType,
		Required: false,
	}
	result, diag := hcldec.Decode(pv.Remain, &spec, nil)
	if diag.HasErrors() {
		return cty.NilVal, diag.Errs()
	}
	return result, nil
}

func (pv *PartialVariableConfig) CoalesceValue(varType cty.Type, defaultValue cty.Value, newValue cty.Value) (cty.Value, error) {
	if newValue.IsNull() && defaultValue.IsNull() {
		return cty.NilVal, errors.New(fmt.Sprintf("%s: No default or override value provided for '%s'", pv.Type.Range(), pv.Name))
	}

	if !newValue.IsNull() {
		if !varType.Equals(newValue.Type()) {
			return cty.NilVal, errors.New(fmt.Sprintf("%s: Incorrect override value type; Incorrect value for variable \"%s\": Expected '%s', Got '%s'", pv.Type.Range(), pv.Name, varType.FriendlyName(), newValue.Type().FriendlyName()))
		}
		return newValue, nil
	}
	return defaultValue, nil
}
