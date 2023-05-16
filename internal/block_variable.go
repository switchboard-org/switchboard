package internal

import "github.com/zclconf/go-cty/cty"

// VariableBlock contains the final variable value as calculated by the Load
// command, which may contain a mixture of default and override values, as provided by the user.
type VariableBlock struct {
	Name  string
	Type  cty.Type
	Value cty.Value
}
