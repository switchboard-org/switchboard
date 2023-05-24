package parsecfg

import (
	"errors"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/switchboard-org/switchboard/internal"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"golang.org/x/exp/slices"
	"strings"
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
	Remain   hcl.Body       `hcl:",remain"` //only used for debugging purposes (nothing actually remains)
}

type variantBlock struct {
	Name   string    `hcl:"name,label"`
	Key    string    `hcl:"key"`
	Format cty.Value `hcl:"format"`
	Remain hcl.Body  `hcl:",remain"` //only used for debugging purposes (nothing actually remains)
}

func (p *schemaBlockParser) parse() ([]internal.SchemaBlock, hcl.Diagnostics) {
	var output []internal.SchemaBlock
	var diagFinal hcl.Diagnostics

	for _, partial := range p.schemaConfigs.Schemas {
		partialValDiags := validateFormatValue(partial.Format, partial.Remain.MissingItemRange(), "", false, true)
		if partialValDiags.HasErrors() {
			diagFinal = diagFinal.Extend(partialValDiags)
		}
		var variants []internal.VariantBlock
		for _, variant := range partial.Variants {
			formatSpec := internal.SchemaFormatValueToSpec(variant.Format)
			variants = append(variants, internal.VariantBlock{
				Name:   variant.Name,
				Key:    variant.Key,
				Format: formatSpec,
			})
			variantValDiags := validateFormatValue(variant.Format, variant.Remain.MissingItemRange(), "", false, false)
			if variantValDiags.HasErrors() {
				diagFinal = diagFinal.Extend(variantValDiags)
			}
		}

		formatSpec := internal.SchemaFormatValueToSpec(partial.Format)
		schemaConfig := internal.SchemaBlock{
			Name:     partial.Name,
			IsList:   partial.IsList,
			Format:   formatSpec,
			Variants: variants,
		}
		output = append(output, schemaConfig)
	}
	if diagFinal.HasErrors() {
		return nil, diagFinal
	}
	return output, nil
}

// validateFormatValue is a recursive function that checks whether the provided user config data for the 'format'
// attribute has an appropriate structure
func validateFormatValue(format cty.Value, cfgRange hcl.Range, keyPath string, isSchematicVal bool, isRootFormat bool) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if !format.Type().IsObjectType() {
		diags = append(diags, validationFormatStandardDiag(keyPath, &cfgRange))
		return diags
	}
	// the base of a format object, and nested objects (either in a list or object) will not be newFormatNode value, but a key/val object as defined by the user
	if !isSchematicVal {
		iter := format.ElementIterator()
		for iter.Next() {
			k, v := iter.Element()
			elementDiag := validateFormatValue(v, cfgRange, genKeyPathString(keyPath, k.AsString()), true, isRootFormat)
			if elementDiag.HasErrors() {
				diags = diags.Extend(elementDiag)
			}
		}
		return diags
	}
	// must be newFormatNode past this point as it was not explicitly defined otherwise
	if !internal.IsFormatConstraintNode(format) {
		diags = append(diags, validationFormatStandardDiag(keyPath, &cfgRange))
		return diags
	}

	// schema keys can only be used at the first level of the root format in the schema (not allowed in variant formats)
	if format.Type().HasAttribute(internal.FORMAT_KEY) && (!isRootFormat || len(strings.Split(keyPath, ".")) > 1) {
		diags = diags.Append(simpleDiagnostic("invalid 'format' value", "cannot use key() function outside of the base level of the root schema format", &cfgRange))
	}

	if format.Type().HasAttribute(internal.FORMAT_CHILDREN) {
		children := format.GetAttr(internal.FORMAT_CHILDREN)
		childrenDiags := validateFormatValue(children, cfgRange, keyPath, internal.IsFormatConstraintNode(children), isRootFormat)
		if childrenDiags.HasErrors() {
			diags = diags.Extend(childrenDiags)
		}
	}

	return diags
}

func validationFormatStandardDiag(keyPath string, cfgRange *hcl.Range) *hcl.Diagnostic {
	return simpleDiagnostic("invalid 'format' value", fmt.Sprintf("format value at key path '%s' is invalid. Available functions: object(), list(). Available variables: string, number, bool.", keyPath), cfgRange)
}

func genKeyPathString(basePath string, newKey string) string {
	if basePath == "" {
		return newKey
	}
	return basePath + "." + newKey
}

// newFormatNode returns a structured cty.Value object that includes details on a particular schema entry.
// only object and list (unless primitive) type nodes will have children.
func newFormatNode(valType string, required bool, children cty.Value) cty.Value {
	obj := map[string]cty.Value{
		internal.FORMAT_TYPE:     cty.StringVal(valType),
		internal.FORMAT_REQUIRED: cty.BoolVal(required),
	}
	if !children.IsNull() {
		obj[internal.FORMAT_CHILDREN] = children
	}
	return cty.ObjectVal(obj)
}

/*
SECTION BELOW FOR EVAL FUNCTIONS ONLY
*/

func complexTypeFunc(args []cty.Value) (cty.Type, error) {
	propSchema := args[0]
	baseSchema := newFormatNode(internal.LIST, false, propSchema)
	return internal.SchemaFormatSpecType(baseSchema), nil
}

// complexImplFunc wraps an outer type (object or list) around an inner type, which can be a primitive type resulting
// from a string, number, bool variable being use, or an object literal with all keys having a proper node value
// (i.e. the result of using one of the  provided variables or functions)
func complexImplFunc(outerType string) func(args []cty.Value, retType cty.Type) (cty.Value, error) {
	return func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		childFormat := args[0]

		// catches the use of raw primitive types
		if !internal.IsAnyFormatNode(childFormat) {
			return cty.NilVal, errors.New("parameter must be a string/number/bool variable or an object literal where all key values are a variable or function")
		}

		if slices.Contains([]string{internal.OBJECT, internal.LIST}, childFormat.GetAttr(internal.FORMAT_TYPE).AsString()) {
			return cty.NilVal, errors.New("object() and list() functions are not allowed")
		}

		return cty.ObjectVal(map[string]cty.Value{
			internal.FORMAT_TYPE:     cty.StringVal(outerType),
			internal.FORMAT_REQUIRED: cty.BoolVal(false),
			internal.FORMAT_CHILDREN: childFormat,
		}), nil

	}
}

func schemaEvalContext(parent *hcl.EvalContext) *hcl.EvalContext {
	evalVars := map[string]cty.Value{
		internal.NUMBER:  newFormatNode(internal.NUMBER, false, cty.NullVal(cty.String)),
		internal.STRING:  newFormatNode(internal.STRING, false, cty.NullVal(cty.String)),
		internal.BOOLEAN: newFormatNode(internal.BOOLEAN, false, cty.NullVal(cty.String)),
	}
	evalVars = internal.MergeMaps(parent.Variables, evalVars)

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
		Type: complexTypeFunc,
		Impl: complexImplFunc(internal.LIST),
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
		Type: complexTypeFunc,
		Impl: complexImplFunc(internal.OBJECT),
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
			return internal.SchemaFormatSpecType(args[0]), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			format := args[0]
			if !internal.IsFormatConstraintNode(format) {
				return cty.NilVal, errors.New("provided value must be a newFormatNode object")
			}
			children := cty.NilVal
			if format.Type().HasAttribute(internal.FORMAT_CHILDREN) {
				children = format.GetAttr(internal.FORMAT_CHILDREN)
			}
			newSchema := newFormatNode(format.GetAttr(internal.FORMAT_TYPE).AsString(), true, children)
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
		Type: func(args []cty.Value) (cty.Type, error) {
			typeMap := map[string]cty.Type{
				internal.FORMAT_TYPE:     cty.String,
				internal.FORMAT_REQUIRED: cty.Bool,
				internal.FORMAT_KEY:      cty.Bool,
			}
			return cty.Object(typeMap), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			schema := args[0]
			if !internal.IsFormatConstraintNode(schema) {
				return cty.NilVal, errors.New("first parameter is invalid. Use string or number variable only")
			}
			if !slices.Contains([]string{internal.STRING, internal.NUMBER}, schema.GetAttr(internal.FORMAT_TYPE).AsString()) {
				return cty.NilVal, errors.New("first parameter is invalid. Use string or number variable only")
			}
			newSchema := map[string]cty.Value{
				internal.FORMAT_TYPE:     schema.GetAttr(internal.FORMAT_TYPE),
				internal.FORMAT_REQUIRED: schema.GetAttr(internal.FORMAT_REQUIRED),
			}
			if schema.Type().HasAttribute(internal.FORMAT_CHILDREN) {
				newSchema[internal.FORMAT_CHILDREN] = schema.GetAttr(internal.FORMAT_CHILDREN)
			}
			newSchema[internal.FORMAT_KEY] = cty.BoolVal(true)
			return cty.ObjectVal(newSchema), nil
		},
	})
	evalFuncs := map[string]function.Function{
		"list":   listFunc,
		"object": objFunc,
		"req":    requiredFunc,
		"key":    keyFunc,
	}
	evalFuncs = internal.MergeMaps(parent.Functions, evalFuncs)
	return &hcl.EvalContext{
		Variables: evalVars,
		Functions: evalFuncs,
	}
}
