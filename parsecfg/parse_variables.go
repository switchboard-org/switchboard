package parsecfg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type variableBlocksParser struct {
	Variables []partialVariableConfig `hcl:"variable,block"`
	Remain    hcl.Body                `hcl:",remain"`
}

type partialVariableConfig struct {
	Name   string         `hcl:"name,label"`
	Type   hcl.Expression `hcl:"type"`
	Remain hcl.Body       `hcl:",remain"`
}

// parse takes the provided variable configuration blocks that each have an optional default value,
// along with a map of discrete override values, and will return a map with coalesced values, with overrides superseding defaults.
// Will throw an error if a variable has no default or override.
func (v *variableBlocksParser) parse(overrides map[string]cty.Value) ([]VariableBlock, hcl.Diagnostics) {
	var output []VariableBlock
	//variables will be built up in here and thrown at end if needed.
	var diagFinal hcl.Diagnostics

	for _, partial := range v.Variables {
		//check that var types are valid
		varType, diag := partial.calculatedType()
		if diag.HasErrors() {
			diagFinal = diagFinal.Extend(diag)
			continue
		}
		//check that defaults are correct type
		defaultVal, diag := partial.defaultValue(varType)
		if diag.HasErrors() {
			diagFinal = diagFinal.Extend(diag)
			continue
		}

		overrideVal := overrides[partial.Name]
		//override defaults if necessary
		variableValue, diag := partial.coalescedValue(varType, defaultVal, overrideVal)
		if diag.HasErrors() {
			diagFinal = diagFinal.Extend(diag)
			continue
		}
		variableConfig := VariableBlock{
			Name:  partial.Name,
			Type:  varType,
			Value: variableValue,
		}
		output = append(output, variableConfig)
	}
	if diagFinal.HasErrors() {
		return nil, diagFinal
	}
	return output, diagFinal
}

// evaluationContext provides a number of context variables and functions
// used in defining custom variables in the configuration. Only use for parsing variables
func (pv *partialVariableConfig) evaluationContext() hcl.EvalContext {
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

func (pv *partialVariableConfig) calculatedType() (cty.Type, hcl.Diagnostics) {
	varList := pv.Type.Variables()
	if len(varList) == 0 {
		varRange := pv.Type.Range()
		diagnostic := hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "Incorrect attribute value type",
			Detail:      fmt.Sprintf("%s: Incorrect attribute value type; Inappropriate value for attribute \"type\": variable required (hint: don't use a primitive)", pv.Type.Range()),
			Subject:     &varRange,
			Context:     nil,
			Expression:  pv.Type,
			EvalContext: nil,
			Extra:       nil,
		}
		diag := hcl.Diagnostics{}
		return cty.NilType, diag.Append(&diagnostic)
	}
	varEvalContext := pv.evaluationContext()
	typeValue, diag := pv.Type.Value(&varEvalContext)
	if diag.HasErrors() {
		return cty.NilType, diag
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

func (pv *partialVariableConfig) defaultValue(varType cty.Type) (cty.Value, hcl.Diagnostics) {
	spec := hcldec.AttrSpec{
		Name:     "default",
		Type:     varType,
		Required: false,
	}
	return hcldec.Decode(pv.Remain, &spec, nil)
}

func (pv *partialVariableConfig) coalescedValue(varType cty.Type, defaultValue cty.Value, newValue cty.Value) (cty.Value, hcl.Diagnostics) {
	varRange := pv.Type.Range()
	diagnostic := hcl.Diagnostic{
		Severity:    hcl.DiagError,
		Summary:     "Incorrect attribute value type",
		Detail:      fmt.Sprintf("No default or override value provided for '%s'", pv.Name),
		Subject:     &varRange,
		Context:     nil,
		Expression:  pv.Type,
		EvalContext: nil,
		Extra:       nil,
	}
	if newValue.IsNull() && defaultValue.IsNull() {
		diagnostic.Summary = "No default or override value provided"
		diagnostic.Detail = fmt.Sprintf("No default or override value provided for '%s'", pv.Name)
		diag := hcl.Diagnostics{}
		return cty.NilVal, diag.Append(&diagnostic)
	}

	if !newValue.IsNull() {
		if !varType.Equals(newValue.Type()) {
			diagnostic.Summary = "Incorrect override value type"
			diagnostic.Detail = fmt.Sprintf("Incorrect override value type; Incorrect value for variable \"%s\": Expected '%s', Got '%s'", pv.Name, varType.FriendlyName(), newValue.Type().FriendlyName())
			diag := hcl.Diagnostics{}
			return cty.NilVal, diag.Append(&diagnostic)
		}
		return newValue, nil
	}
	return defaultValue, nil
}
