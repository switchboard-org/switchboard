package internal

import (
	"github.com/zclconf/go-cty/cty"
	"log"
	"testing"
)

func TestValidateValueAgainstSpec_RequiredFieldOrKeyMissing(t *testing.T) {
	// Create a test case with a sample value and spec
	value := cty.ObjectVal(map[string]cty.Value{
		"otherKey": cty.StringVal("value1"),
	})
	spec := MapSpec{
		"key1": &PrimitiveSpec{
			required:  false,
			isKey:     true,
			fieldType: cty.String,
		},
		"key2": &PrimitiveSpec{
			required:  true,
			isKey:     false,
			fieldType: cty.String,
		},
		"otherKey": &PrimitiveSpec{
			fieldType: cty.String,
			isKey:     true,
			required:  false,
		},
	}

	// Call the function being tested
	errors := ValidateValueAgainstSpec(value, &spec, "")

	// Assert the expected error count
	if len(errors) != 2 {
		t.Errorf("Expected 2 error, but got %d", len(errors))
	}

	// Assert the error message
	expectedErrorMessage := "missing a required string value at 'key1' key"
	if errors[0].Error() != expectedErrorMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrorMessage, errors[0].Error())
	}

	expectedErrorMessageTwo := "missing a required string value at 'key2' key"
	if errors[1].Error() != expectedErrorMessageTwo {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrorMessageTwo, errors[2].Error())
	}
}

func TestValidateValueAgainstSpec_NestedValueIsMissing(t *testing.T) {
	// Create a test case with a sample value and spec
	value := cty.ObjectVal(map[string]cty.Value{
		"nested": cty.ObjectVal(map[string]cty.Value{
			"key1": cty.StringVal("hello"),
		}),
	})
	spec := MapSpec{
		"nested": &ObjectSpec{
			required: true,
			keyMap: map[string]Spec{
				"key1": &PrimitiveSpec{
					required:  true,
					isKey:     false,
					fieldType: cty.String,
				},
				"key2": &PrimitiveSpec{
					required:  true,
					isKey:     false,
					fieldType: cty.String,
				},
			},
		},
	}

	// Call the function being tested
	errors := ValidateValueAgainstSpec(value, &spec, "")

	// Assert the expected error count
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, but got %d", len(errors))
	}

	// Assert the error message
	expectedErrorMessage := "missing a required string value at 'nested.key2' key"
	if errors[0].Error() != expectedErrorMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrorMessage, errors[0].Error())
	}
}

func TestValidateValueAgainstSpec_InvalidNestedListObjects(t *testing.T) {
	// Create a test case with a sample value and spec
	value := cty.ObjectVal(map[string]cty.Value{
		"list": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"key1": cty.StringVal("hello"),
				"key2": cty.NullVal(cty.String),
				"key3": cty.NullVal(cty.String),
			}),
			//this is the only valid one
			cty.ObjectVal(map[string]cty.Value{
				"key1": cty.StringVal("hello"),
				"key2": cty.StringVal("hola"),
				"key3": cty.NullVal(cty.String),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"key1": cty.NullVal(cty.String),
				"key2": cty.StringVal("hello"),
				"key3": cty.NullVal(cty.String),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"key1": cty.NullVal(cty.String),
				"key2": cty.NullVal(cty.String),
				"key3": cty.StringVal("hello"),
			}),
		}),
	})
	spec := MapSpec{
		"list": &ListSpec{
			required: true,
			innerTypeSpec: &MapSpec{
				"key1": &PrimitiveSpec{
					required:  true,
					isKey:     false,
					fieldType: cty.String,
				},
				"key2": &PrimitiveSpec{
					required:  true,
					isKey:     false,
					fieldType: cty.String,
				},
			},
		},
	}

	// Call the function being tested
	errors := ValidateValueAgainstSpec(value, &spec, "")

	// Assert the expected error count
	if len(errors) != 4 {
		t.Errorf("Expected 3 error, but got %d", len(errors))
	}

	log.Println(errors)
	// Assert the error message
	expectedErrorMessage := "missing a required string value at 'list[0].key2' key"
	if errors[0].Error() != expectedErrorMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrorMessage, errors[0].Error())
	}
}

func TestValidateValueAgainstSpec_NullValueForNonNullableField(t *testing.T) {
	// Create a test case with a sample value and spec
	value := cty.ObjectVal(map[string]cty.Value{
		"key1": cty.NullVal(cty.String), // Null value for a non-nullable field
	})
	spec := &ObjectSpec{
		required: true,
		keyMap: map[string]Spec{
			"key1": &PrimitiveSpec{
				required:  true,
				isKey:     false,
				fieldType: cty.String,
			},
		},
	}

	// Call the function being tested
	errors := ValidateValueAgainstSpec(value, spec, "")

	// Assert the expected error count
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, but got %d", len(errors))
	}

	// Assert the error message
	expectedErrorMessage := "missing a required string value at 'key1' key"
	if errors[0].Error() != expectedErrorMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrorMessage, errors[0].Error())
	}
}

func TestSchemaFormatValueToSpec_InvalidFormatValue(t *testing.T) {
	// Create a test case with an invalid format value
	formatValue := cty.StringVal("invalid") // Invalid format value type

	// Call the function being tested
	spec := SchemaFormatValueToSpec(formatValue)

	// Assert the expected spec type
	if spec != nil {
		t.Errorf("Expected nil spec, but got %T", spec)
	}
}

func TestSchemaFormatValueToSpec_ValidBasicFormat(t *testing.T) {
	formatValue := cty.ObjectVal(map[string]cty.Value{
		FORMAT_TYPE:     cty.StringVal(STRING),
		FORMAT_REQUIRED: cty.BoolVal(false),
	})
	spec := SchemaFormatValueToSpec(formatValue)
	if spec == nil {
		t.Errorf("Expected a spec, but got nil")
	}
}

// newFormatNode is also a function in parse_schemas.go ... should probably simplify
func newFormatNode(valType string, required bool, children cty.Value) cty.Value {
	obj := map[string]cty.Value{
		FORMAT_TYPE:     cty.StringVal(valType),
		FORMAT_REQUIRED: cty.BoolVal(required),
	}
	if !children.IsNull() {
		obj[FORMAT_CHILDREN] = children
	}
	return cty.ObjectVal(obj)
}

func TestSchemaFormatValueToSpec_ComplexValue(t *testing.T) {
	nestedObj := cty.ObjectVal(map[string]cty.Value{
		"requiredKey": newFormatNode(STRING, true, cty.NullVal(cty.String)),
	})
	fullConfigVal := cty.ObjectVal(map[string]cty.Value{
		"stringKey":         newFormatNode(STRING, false, cty.NullVal(cty.String)),
		"boolKey":           newFormatNode(BOOLEAN, true, cty.NullVal(cty.Bool)),
		"numKey":            newFormatNode(NUMBER, false, cty.NullVal(cty.Number)),
		"nested":            newFormatNode(OBJECT, false, nestedObj),
		"listWithObj":       newFormatNode(LIST, false, nestedObj),
		"listWithPrimitive": newFormatNode(LIST, false, newFormatNode(STRING, false, cty.NullVal(cty.String))),
	})

	spec := SchemaFormatValueToSpec(fullConfigVal)
	if spec == nil {
		t.Errorf("Expected a spec, but got nil")
	}
	if !spec.Valid() {
		t.Errorf("Expected valid spec, but was invalid")
	}
}

func TestSchemaFormatValueToSpec_ComplexInvalidValue(t *testing.T) {
	nestedObj := cty.ObjectVal(map[string]cty.Value{
		"badVal": cty.StringVal("bad"),
	})
	fullConfigVal := cty.ObjectVal(map[string]cty.Value{
		"stringKey":         newFormatNode(STRING, false, cty.NullVal(cty.String)),
		"boolKey":           newFormatNode(BOOLEAN, true, cty.NullVal(cty.Bool)),
		"numKey":            newFormatNode(NUMBER, false, cty.NullVal(cty.Number)),
		"nested":            newFormatNode(OBJECT, false, nestedObj),
		"listWithObj":       newFormatNode(LIST, false, nestedObj),
		"listWithPrimitive": newFormatNode(LIST, false, newFormatNode(STRING, false, cty.NullVal(cty.String))),
	})

	spec := SchemaFormatValueToSpec(fullConfigVal)
	if spec == nil {
		t.Errorf("Expected a spec, but got nil")
	}
	if spec.Valid() {
		t.Errorf("Expected invalid")
	}
}

func TestShallowMergeMapSpecs(t *testing.T) {
	mapOne := MapSpec{
		"test": &PrimitiveSpec{
			fieldType: cty.String,
			isKey:     false,
			required:  false,
		},
	}
	mapTwo := MapSpec{
		"test": &PrimitiveSpec{
			fieldType: cty.String,
			isKey:     false,
			required:  true,
		},
		"other": &PrimitiveSpec{
			fieldType: cty.String,
			isKey:     false,
			required:  false,
		},
	}
	spec := ShallowMergeMapSpecs(mapOne, mapTwo)
	if spec == nil {
		t.Errorf("Expected a spec but got nil")
	}
	children := spec.Children()
	for k, v := range children {
		if k == "test" && !v.IsRequired() {
			t.Errorf("Expected 'test' value to be required")
		}
	}
	if len(children) != 2 {
		t.Errorf("Expected map length of 2 but got %v", len(children))
	}
}
