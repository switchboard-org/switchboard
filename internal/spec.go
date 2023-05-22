package internal

import (
	"errors"
	"fmt"
	"github.com/zclconf/go-cty/cty"
)

type Spec interface {
	IsRequired() bool
	IsKey() bool
	Type() cty.Type
	Children() map[string]Spec
}

type ObjectSpec struct {
	required bool
	keyMap   map[string]Spec
}

func (s *ObjectSpec) IsRequired() bool {
	return s.required
}

func (s *ObjectSpec) IsKey() bool {
	return false
}

func (s *ObjectSpec) Type() cty.Type {
	var specs map[string]cty.Type
	for k, v := range s.keyMap {
		specs[k] = v.Type()
	}
	return cty.Object(specs)
}

func (s *ObjectSpec) Children() map[string]Spec {
	return s.keyMap
}

type PrimitiveSpec struct {
	fieldType cty.Type
	isKey     bool
	required  bool
}

func (s *PrimitiveSpec) IsRequired() bool {
	return s.required
}

func (s *PrimitiveSpec) IsKey() bool {
	return s.isKey
}

func (s *PrimitiveSpec) Type() cty.Type {
	return s.fieldType
}

func (s *PrimitiveSpec) Children() map[string]Spec {
	return nil
}

type ListSpec struct {
	required bool
	//innerTypeSpec represents what is IN the list
	innerTypeSpec Spec
}

func (s *ListSpec) IsRequired() bool {
	return s.required
}

func (s *ListSpec) IsKey() bool {
	return false
}

func (s *ListSpec) Type() cty.Type {
	return cty.List(s.innerTypeSpec.Type())
}

func (s *ListSpec) Children() map[string]Spec {
	return s.innerTypeSpec.Children()
}

// MapSpec allows us to solve the unique case of the root of all format
// objects and children being a key/val instead of a standard format schema-looking object.
type MapSpec map[string]Spec

func (s *MapSpec) IsRequired() bool {
	return false
}

func (s *MapSpec) IsKey() bool {
	return false
}

func (s *MapSpec) Type() cty.Type {
	attr := make(map[string]cty.Type)
	for k, v := range *s {
		attr[k] = v.Type()
	}
	return cty.Object(attr)
}

func (s *MapSpec) Children() map[string]Spec {
	return *s
}

// ValidateValueAgainstSpec will go through an entire value and ensure it meets the constraints of the spe
func ValidateValueAgainstSpec(val cty.Value, spec Spec, keyPath string) []error {
	var outputErrorList []error

	if val.IsNull() && (spec.IsRequired() || spec.IsKey()) {
		return append(outputErrorList, errors.New(fmt.Sprintf("%s value at key path '%s' is required.", spec.Type().FriendlyName(), keyPath)))
	}

	children := spec.Children()

	// if children are nil, we have nothing else to validate
	if children == nil {
		return outputErrorList
	}

	if spec.Type().IsObjectType() {
		for k, v := range children {
			errs := ValidateValueAgainstSpec(val.GetAttr(k), v, fmt.Sprintf("%s.%s", keyPath, k))
			outputErrorList = append(outputErrorList, errs...)
		}
		return outputErrorList
	}

	if spec.Type().IsListType() {
		oSpec := ObjectSpec{
			required: true,
			keyMap:   children,
		}
		iter := val.ElementIterator()
		for iter.Next() {
			i, v := iter.Element()
			errs := ValidateValueAgainstSpec(v, &oSpec, fmt.Sprintf("%s[%s]", keyPath, i))
			outputErrorList = append(outputErrorList, errs...)
		}
		return outputErrorList
	}

	return outputErrorList
}

// SchemaFormatValueToSpec transforms a format value to a Spec struct. This function should
// only be used after validating the format structure, given cty.Value is inherently dynamic
func SchemaFormatValueToSpec(val cty.Value) Spec {
	if !IsAnyFormatNode(val) {
		return nil
	}

	if IsFormatKeyValNode(val) {
		outputObj := MapSpec{}
		iter := val.ElementIterator()
		for iter.Next() {
			k, v := iter.Element()
			outputObj[k.AsString()] = SchemaFormatValueToSpec(v)
		}
		return &outputObj
	}

	typeStr := val.GetAttr(FORMAT_TYPE).AsString()
	isReq := val.GetAttr(FORMAT_REQUIRED).True()
	isKey := false
	if val.Type().HasAttribute(FORMAT_KEY) {
		val.GetAttr(FORMAT_KEY).True()
	}
	switch typeStr {
	case NUMBER:
		return &PrimitiveSpec{
			required:  isReq,
			isKey:     isKey,
			fieldType: cty.Number,
		}
	case BOOLEAN:
		return &PrimitiveSpec{
			required:  isReq,
			isKey:     isKey,
			fieldType: cty.Bool,
		}
	case STRING:
		return &PrimitiveSpec{
			required:  isReq,
			isKey:     isKey,
			fieldType: cty.String,
		}
	case LIST:
		return &ListSpec{
			required:      isReq,
			innerTypeSpec: SchemaFormatValueToSpec(val.GetAttr(FORMAT_CHILDREN)),
		}
	case OBJECT:
		objSpec := SchemaFormatValueToSpec(val.GetAttr(FORMAT_CHILDREN))
		return &ObjectSpec{
			required: isReq,
			keyMap:   objSpec.Children(),
		}
	}

	return nil
}
