package internal

import "github.com/zclconf/go-cty/cty"

// SwitchboardConfig is a container for all fully parsed and decoded
// config values, after parsing and decoding each root attribute in isolation.
// Some values are computed vs. coming directly from the config files.
type SwitchboardConfig struct {
	Variables []VariableConfig `hcl:"variable,block"`
}

type VariableConfig struct {
	Name  string    `hcl:"name,label"`
	Type  string    `hcl:"type"`
	Value cty.Value `hcl:"value"`
}
