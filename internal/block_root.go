package internal

//this file contains the config container and all root block types

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// RootSwitchboardConfig is a container for all fully parsed and decoded
// static config settings as provided by the user. They are the result of processing all
// hcl block types in isolation, in an appropriate order.
// Most values come as discrete primitives from the config, while others are computed expressions.
//
// Certain attributes are not known at the config parsing stage, such as variables derived from triggers
// or workflow steps. These values will be evaluated during individual workflow cycles.
type RootSwitchboardConfig struct {
	Variables   []VariableBlock
	Switchboard SwitchboardBlock
	Providers   []ProviderBlock
	Schemas     []SchemaBlock
}

// EvalContext is the high level evaluation context object used for evaluating expressions throughout
// the various blocks of the parent configuration. Note: The RootSwitchboardConfig.Variables should
// already be calculated and set before this can be used to evaluate other root blocks
// on the RootSwitchboardConfig object.
func (conf *RootSwitchboardConfig) EvalContext() *hcl.EvalContext {
	var evalContext hcl.EvalContext
	evalContextVariables := make(map[string]cty.Value)
	for _, value := range conf.Variables {
		evalContextVariables[value.Name] = value.Value
	}
	if conf.Schemas != nil || len(conf.Schemas) != 0 {
		schemaMap := make(map[string]cty.Value)
		// the key of the schema will be the name, and the value the index of the object
		for i, schema := range conf.Schemas {
			schemaMap[schema.Name] = cty.NumberIntVal(int64(i))
		}
		evalContextVariables["schemas"] = cty.ObjectVal(schemaMap)
	}

	evalContext.Variables = evalContextVariables
	evalContext.Functions = generalContextFunctions()
	return &evalContext
}
