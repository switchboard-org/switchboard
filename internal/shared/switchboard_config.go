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

// EvalContext is the high level evaluation context object used for evaluating expressions throughout
// the various blocks of the parent configuration. Note: The SwitchboardConfig.Variables should
// already be calculated and set before this can be used to evaluate other root blocks
// on the SwitchboardConfig object.
func (conf *SwitchboardConfig) EvalContext() hcl.EvalContext {
	var evalContext hcl.EvalContext
	var evalContextVariables map[string]cty.Value
	for _, value := range conf.Variables {
		evalContextVariables[value.Name] = value.Value
	}

	evalContext.Variables = evalContextVariables
	evalContext.Functions = generalContextFunctions()
	return evalContext
}

// Variable contains the final variable value as calculated by the Load
// command, which may contain a mixture of default and override values, as provided by the user.
type Variable struct {
	Name  string    `hcl:"name,label"`
	Type  cty.Type  `hcl:"type"`
	Value cty.Value `hcl:"value"`
}

// Switchboard contains the primary global configuration elements of all workflows,
// including the required providers, log settings, retry settings, and more.
type Switchboard struct {
	Version           string             `hcl:"version"`
	RequiredProviders []RequiredProvider `hcl:"required_providers,block"`
}

// RequiredProvider tells us where a provider should be pulled from, and which version it
// should use.
type RequiredProvider struct {
	Name    string `hcl:"name,label"`
	Source  string `hcl:"source"`
	Version string `hcl:"version"`
}

// Provider block lets a user configure various settings for a particular provider, such
// as auth contexts and other provider-specific settings. A Provider block will be mapped to a
// RequiredProvider block by matching "name" block labels.
type Provider struct {
	Name           string                  `hcl:"name,label"`
	Authorizations []ProviderAuthorization `hcl:"authorization,block"`
	//providers can have their own schemas for provider blocks. They will fall into the Remain field
	Remain hcl.Body `hcl:",remain"`
}

type ProviderAuthorization struct {
	Name string `hcl:"name,label"`
}
