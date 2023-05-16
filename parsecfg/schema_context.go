package parsecfg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// SchemaContext generates a cty.Value object filled with known types but unknown values
// to be used for validation and data saturation during trigger and workflow evaluation
func SchemaContext() hcl.EvalContext {
	evalVars := map[string]cty.Value{
		"number":  cty.UnknownVal(cty.Number),
		"string":  cty.UnknownVal(cty.String),
		"boolean": cty.UnknownVal(cty.Bool),
	}
	var listFunc = function.New(&function.Spec{
		Description: `used to generate a list type for variables`,
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.DynamicPseudoType,
				AllowDynamicType: true,
				AllowUnknown:     true,
				AllowMarked:      false,
			},
		},
		Type: function.StaticReturnType(cty.List(cty.DynamicPseudoType)),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			listType := args[0]
			return cty.UnknownVal(cty.List(listType.Type())), nil
		},
	})
	var objectFunc = function.New(&function.Spec{
		Description: "used to generate a nested object",
		Params: []function.Parameter{
			{
				Name:             "data",
				Type:             cty.Map(cty.DynamicPseudoType),
				AllowDynamicType: true,
				AllowMarked:      false,
			},
		},
		Type: function.StaticReturnType(cty.Map(cty.DynamicPseudoType)),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			mapVal := args[0]
			outputMap := make(map[string]cty.Value)
			for k, v := range mapVal.AsValueMap() {
				outputMap[k] = v
			}
			return cty.MapVal(outputMap), nil
		},
	})
	evalFuncs := map[string]function.Function{
		"list":   listFunc,
		"object": objectFunc,
	}
	return hcl.EvalContext{
		Variables: evalVars,
		Functions: evalFuncs,
	}
}
