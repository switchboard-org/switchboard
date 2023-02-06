package shared

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// SwitchboardConfig is a container for all fully parsed and decoded
// static shared settings as provided by the user. They are the result of processing all
// hcl block types in isolation, in an appropriate order.
// Some values are set discretely set in the shared, while others are computed via expressions.
//
// Some values cannot be known in the cli's context, such as variables derived from triggers
// or workflow steps. These values will remain as hcl.Expression or hcl.Body types until the
// hcl shared is parsed and decoded during individual workflow runs.
type SwitchboardConfig struct {
	Variables   []Variable  `hcl:"variable,block"`
	Switchboard Switchboard `hcl:"switchboard,block"`
	Providers   []Provider  `hcl:"provider,block"`
}

// Variable contains the final variable value as calculated by the Load
// command, which may contain a mixture of default and override values, as provided by the user.
type Variable struct {
	Name  string    `hcl:"name,label"`
	Type  cty.Type  `hcl:"type"`
	Value cty.Value `hcl:"value"`
}

type Switchboard struct {
	Version string `hcl:"version"`
}

type Provider struct {
	Name    string `hcl:"name,label"`
	Source  string `hcl:"source"`
	Version string `hcl:"version"`
	//remaining values defined by each provider
	Remain hcl.Body `hcl:",remain"`
}

// GetEvaluationContext provides the entire evaluation context needed for running a workflow.
// Specifically, the evaluation context includes all variables, as set in the SwitchboardConfig.Variables
// list, as well as a number of useful functions that can be used inside of trigger, workflow, and step blocks.
func (c *SwitchboardConfig) GetEvaluationContext() {

}
