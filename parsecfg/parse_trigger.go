package parsecfg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// triggerBlockParser is responsible for parsing trigger blocks.
type triggerBlockParser struct {
	triggerConfigs triggersStepConfig
}

// triggerStepConfig is the configuration for a trigger block.
type triggersStepConfig struct {
	Triggers []triggerConfig `hcl:"trigger,block"`
	Remain   hcl.Body        `hcl:",remain"`
}

// triggerConfig is the configuration for an individual trigger block.
type triggerConfig struct {
	Name     string    `hcl:"name,label"`
	Provider string    `hcl:"provider"`
	Function string    `hcl:"function"`
	Schema   cty.Value `hcl:"schema"`
}
