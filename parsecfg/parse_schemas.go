package parsecfg

import (
	"errors"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/switchboard-org/switchboard/internal"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type schemaBlockParser struct {
	schemaConfigs schemasStepConfig
}

type schemasStepConfig struct {
	Schemas []schemaBlock `hcl:"schema,block"`
	Remain  hcl.Body      `hcl:",remain"`
}

type schemaBlock struct {
	Name     string         `hcl:"name,label"`
	IsList   *bool          `hcl:"is_list"`
	Format   cty.Value      `hcl:"format"`
	Variants []variantBlock `hcl:"variant,block"`
}

type variantBlock struct {
	Name   string    `hcl:"name,label"`
	Key    string    `hcl:"key"`
	Format cty.Value `hcl:"format"`
}

const (
	STRING          = "string"
	NUMBER          = "number"
	BOOLEAN         = "boolean"
	LIST            = "list"
	OBJECT          = "object"
	SCHEMA_TYPE     = "type"
	SCHEMA_REQUIRED = "required"
	SCHEMA_NESTED   = "nestedSchema"
)

func (p *schemaBlockParser) parse() ([]internal.SchemaBlock, hcl.Diagnostics) {
	var output []internal.SchemaBlock
	var diagFinal hcl.Diagnostics

	for _, partial := range p.schemaConfigs.Schemas {
		var variants []internal.VariantBlock
		for _, variant := range partial.Variants {
			variants = append(variants, internal.VariantBlock{
				Name:   variant.Name,
				Key:    variant.Key,
				Format: variant.Format,
			})
		}
		schemaConfig := internal.SchemaBlock{
			Name:     partial.Name,
			IsList:   partial.IsList,
			Format:   partial.Format,
			Variants: variants,
		}
		output = append(output, schemaConfig)
	}
	if diagFinal.HasErrors() {
		return nil, diagFinal
	}
	return output, nil
}

// schematicType is a recursive function that returns the full cty.Type from the root of the provided schema.
// All leafs will be structured to what schematic returns.
func schematicType(schema cty.Value) cty.Type {
	if !schema.CanIterateElements() {
		return cty.NilType
	}
	// if it's not a schematic, it must be a key/val object that includes user defined schema structure
	if !isSchematic(schema) {
		resultMap := make(map[string]cty.Type)
		keyValMap := schema.AsValueMap()
		for k, v := range keyValMap {
			resultMap[k] = schematicType(v)
		}
		return cty.Object(resultMap)
	}

	// schema is not a map type, let's return the cty.Type structure of the schematic.
	baseObjectMap := map[string]cty.Type{
		SCHEMA_TYPE:     cty.String,
		SCHEMA_REQUIRED: cty.Bool,
	}

	if schema.Type().IsObjectType() && schema.Type().HasAttribute(SCHEMA_NESTED) {
		baseObjectMap[SCHEMA_NESTED] = schematicType(schema.GetAttr(SCHEMA_NESTED))
	}
	return cty.Object(baseObjectMap)

}

// schematic returns a structured cty.Value object that includes details on a particular schema entry
func schematic(valType string, required bool, nestedSchema cty.Value) cty.Value {
	obj := map[string]cty.Value{
		SCHEMA_TYPE:     cty.StringVal(valType),
		SCHEMA_REQUIRED: cty.BoolVal(required),
	}
	if !nestedSchema.IsNull() {
		obj[SCHEMA_NESTED] = nestedSchema
	}
	return cty.ObjectVal(obj)
}

func isSchematic(value cty.Value) bool {
	valType := value.Type()
	if value.IsNull() || !value.IsKnown() || !value.CanIterateElements() || !valType.IsObjectType() {
		return false
	}
	return valType.HasAttribute(SCHEMA_TYPE) && valType.HasAttribute(SCHEMA_REQUIRED)
}

func nestedTypeFunc(args []cty.Value) (cty.Type, error) {
	propSchema := args[0]
	baseSchema := schematic(LIST, false, propSchema)
	return schematicType(baseSchema), nil
}

func nestedImplFunc(nestedType string) func(args []cty.Value, retType cty.Type) (cty.Value, error) {
	return func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		schema := args[0]
		if nestedType == OBJECT && isSchematic(schema) {
			return cty.NilVal, errors.New("paramater for object function must be a map of key/vals")
		}
		if !schema.Type().IsObjectType() && !schema.Type().IsMapType() {
			return cty.NilVal, errors.New(fmt.Sprintf("%s parameter must be a key/val map or supported schematic. (HINT: ", nestedType))
		}
		return cty.ObjectVal(map[string]cty.Value{
			SCHEMA_TYPE:     cty.StringVal(nestedType),
			SCHEMA_REQUIRED: cty.BoolVal(false),
			SCHEMA_NESTED:   schema,
		}), nil

	}

}

func schemaEvalContext() *hcl.EvalContext {
	evalVars := map[string]cty.Value{
		NUMBER:  schematic(NUMBER, false, cty.NilVal),
		STRING:  schematic(STRING, false, cty.NilVal),
		BOOLEAN: schematic(BOOLEAN, false, cty.NilVal),
	}

	var listFunc = function.New(&function.Spec{
		Description: `used to generate a list type for variables`,
		Params: []function.Parameter{
			{
				Name:             "schema",
				Type:             cty.DynamicPseudoType,
				AllowDynamicType: true,
				AllowUnknown:     true,
				AllowMarked:      false,
			},
		},
		Type: nestedTypeFunc,
		Impl: nestedImplFunc(LIST),
	})
	var objFunc = function.New(&function.Spec{
		Description: `used to generate an object type for variables`,
		Params: []function.Parameter{
			{
				Name:             "schema",
				Type:             cty.DynamicPseudoType,
				AllowDynamicType: true,
				AllowUnknown:     true,
				AllowMarked:      false,
			},
		},
		Type: nestedTypeFunc,
		Impl: nestedImplFunc(OBJECT),
	})
	var requiredFunc = function.New(&function.Spec{
		Description: "used to mark a schema element as required",
		Params: []function.Parameter{
			{
				Name:             "schema",
				Type:             cty.DynamicPseudoType,
				AllowDynamicType: true,
				AllowMarked:      false,
			},
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			return schematicType(args[0]), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			schema := args[0]
			if !isSchematic(schema) {
				return cty.NilVal, errors.New("provided value must be a schematic object")
			}
			nestedSchema := cty.NilVal
			if schema.Type().HasAttribute(SCHEMA_NESTED) {
				nestedSchema = schema.GetAttr(SCHEMA_NESTED)
			}
			newSchema := schematic(schema.GetAttr(SCHEMA_TYPE).AsString(), true, nestedSchema)
			return newSchema, nil
		},
	})
	var keyFunc = function.New(&function.Spec{
		Description: "used to identify which value is the key for the schema variants",
		Params: []function.Parameter{
			{
				Name:             "value_type",
				Type:             cty.DynamicPseudoType,
				AllowDynamicType: true,
				AllowMarked:      true,
				AllowUnknown:     true,
				AllowNull:        false,
			},
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			param := args[0]
			out := param.Mark("schema_key")
			return out, nil
		},
	})
	evalFuncs := map[string]function.Function{
		"list":   listFunc,
		"object": objFunc,
		"req":    requiredFunc,
		"key":    keyFunc,
	}
	return &hcl.EvalContext{
		Variables: evalVars,
		Functions: evalFuncs,
	}
}
