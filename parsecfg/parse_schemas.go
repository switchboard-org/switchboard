package parsecfg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type schemaBlockParser struct {
	schemaConfigs schemasStepConfig
}

type schemasStepConfig struct {
	Schemas []schemaBlock `hcl:"schema,block"`
	Remain  hcl.Body      `hcl:",remain"`
}

type schemaBlock struct {
	Name     string               `hcl:"name,label"`
	IsList   bool                 `hcl:"name"`
	Format   map[string]cty.Value `hcl:"format"`
	Variants []variantBlock       `hcl:"variant,block"`
}

type variantBlock struct {
	Name   string               `hcl:"name,label"`
	Key    string               `hcl:"key"`
	Format map[string]cty.Value `hcl:"format"`
}
