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