package internal

import (
	"errors"
	"fmt"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

type Spec interface {
	IsRequired() bool
	IsKey() bool
	Type() cty.Type
	Children() map[string]Spec
	Valid() bool
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
	specs := make(map[string]cty.Type)
	for k, v := range s.keyMap {
		specs[k] = v.Type()
	}
	return cty.Object(specs)
}

func (s *ObjectSpec) Children() map[string]Spec {
	return s.keyMap
}

func (s *ObjectSpec) Valid() bool {
	if s.keyMap == nil {
		return false
	}
	for _, v := range s.keyMap {
		if v == nil || !v.Valid() {
			return false
		}
	}
	return true
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

func (s *PrimitiveSpec) Valid() bool {
	return true
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

func (s *ListSpec) Valid() bool {
	if s.innerTypeSpec == nil {
		return false
	}
	for _, v := range s.innerTypeSpec.Children() {
		if v == nil || !v.Valid() {
			return false
		}
	}
	return true
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

func (s *MapSpec) Valid() bool {
	if s == nil {
		return false
	}
	for _, v := range *s {
		if v == nil || !v.Valid() {
			return false
		}
	}
	return true
}

// ValidateValueAgainstSpec will go through an entire value and ensure it meets the constraints of the spe
func ValidateValueAgainstSpec(val cty.Value, spec Spec, keyPath string) []error {
	var outputErrorList []error
	if val.IsNull() && (spec.IsRequired() || spec.IsKey()) {
		return append(outputErrorList, errors.New(fmt.Sprintf("missing a required %s value at '%s' key", spec.Type().FriendlyName(), keyPath)))
	}

	children := spec.Children()

	if spec.Type().IsPrimitiveType() || children == nil {
		return outputErrorList
	}

	if spec.Type().IsObjectType() {
		for k, v := range children {
			newVal := cty.NullVal(cty.DynamicPseudoType)
			if val.Type().HasAttribute(k) {
				newVal = val.GetAttr(k)
			}
			nextKeyPath := fmt.Sprintf("%s.%s", keyPath, k)
			nextKeyPath = strings.TrimLeft(nextKeyPath, ".")
			errs := ValidateValueAgainstSpec(newVal, v, nextKeyPath)
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
		count := 0
		for iter.Next() {
			//list values
			_, v := iter.Element()
			errs := ValidateValueAgainstSpec(v, &oSpec, fmt.Sprintf("%s[%v]", keyPath, count))
			outputErrorList = append(outputErrorList, errs...)
			count++
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
		isKey = val.GetAttr(FORMAT_KEY).True()
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
		if objSpec == nil {
			return &ObjectSpec{
				required: isReq,
				keyMap:   nil,
			}
		}
		return &ObjectSpec{
			required: isReq,
			keyMap:   objSpec.Children(),
		}
	}

	return nil
}

// ShallowMergeMapSpecs takes two specs and merges them, with the rightSpec taking precedent
// over any duplicates on the left.
func ShallowMergeMapSpecs(leftSpec MapSpec, rightSpec MapSpec) MapSpec {
	finalSpec := MapSpec{}
	for k, v := range leftSpec.Children() {
		finalSpec[k] = v
	}
	for k, v := range rightSpec.Children() {
		finalSpec[k] = v
	}
	return finalSpec
}
