package internal

import (
	"github.com/zclconf/go-cty/cty"
)

const (
	STRING          = "string"
	NUMBER          = "number"
	BOOLEAN         = "bool"
	LIST            = "list"
	OBJECT          = "object"
	FORMAT_TYPE     = "format_type"
	FORMAT_REQUIRED = "format_required"
	FORMAT_CHILDREN = "format_children"
	FORMAT_KEY      = "format_key"
)

type SchemaBlock struct {
	Name     string         `hcl:"name,label"`
	IsList   *bool          `hcl:"name"`
	Format   Spec           `hcl:"format"`
	Variants []VariantBlock `hcl:"variant,block"`
}

type VariantBlock struct {
	Name   string `hcl:"name,label"`
	Key    string `hcl:"key"`
	Format Spec   `hcl:"format"`
}

// IsAnyFormatNode is a simple check to make sure a particular value conforms to one of the format node structures.
// This function does not deep check any nested children, just the top level.
func IsAnyFormatNode(value cty.Value) bool {
	return IsFormatConstraintNode(value) || IsFormatKeyValNode(value)
}

// IsFormatConstraintNode checks to see whether a value conforms to the structure of a constraint node, which
// will/should be true for all leaf nodes in a value object
func IsFormatConstraintNode(value cty.Value) bool {
	valType := value.Type()
	if value.IsNull() || !value.IsKnown() || !value.CanIterateElements() || !valType.IsObjectType() {
		return false
	}
	return valType.HasAttribute(FORMAT_TYPE) && valType.HasAttribute(FORMAT_REQUIRED)
}

// IsFormatKeyValNode checks to see whether a value conforms to the structure of a key/val node where all values
// conform to the constraint node type. A KeyValNode should exist at the top level of the format value, and in
// some FORMAT_CHILDREN values (those that are "nested" objects)
func IsFormatKeyValNode(value cty.Value) bool {
	valType := value.Type()
	if IsFormatConstraintNode(value) || value.IsNull() || !value.IsKnown() || !value.CanIterateElements() || !valType.IsObjectType() {
		return false
	}
	iter := value.ElementIterator()
	for iter.Next() {
		k, v := iter.Element()
		if !k.Type().Equals(cty.String) {
			return false
		}
		// all vals in key/val MUST be a normal node
		if !IsFormatConstraintNode(v) {
			return false
		}
	}
	return true
}

// SchemaFormatSpecType is a recursive function that returns the full cty.Type from the root of the provided schema.
// All leafs will be structured to what schematic returns.
func SchemaFormatSpecType(schema cty.Value) cty.Type {
	if !schema.CanIterateElements() {
		return cty.NilType
	}
	// if it's not a schematic, it must be a key/val object that includes user defined schema structure
	if !IsFormatConstraintNode(schema) {
		resultMap := make(map[string]cty.Type)
		keyValMap := schema.AsValueMap()
		for k, v := range keyValMap {
			resultMap[k] = SchemaFormatSpecType(v)
		}
		return cty.Object(resultMap)
	}

	// schema is not a map type, let's return the cty.Type structure of the schematic.
	baseObjectMap := map[string]cty.Type{
		FORMAT_TYPE:     cty.String,
		FORMAT_REQUIRED: cty.Bool,
	}

	if schema.Type().IsObjectType() && schema.Type().HasAttribute(FORMAT_CHILDREN) {
		baseObjectMap[FORMAT_CHILDREN] = SchemaFormatSpecType(schema.GetAttr(FORMAT_CHILDREN))
	}

	if schema.Type().IsObjectType() && schema.Type().HasAttribute(FORMAT_KEY) {
		baseObjectMap[FORMAT_KEY] = cty.Bool
	}

	return cty.Object(baseObjectMap)
}
