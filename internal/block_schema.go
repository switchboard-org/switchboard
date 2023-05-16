package internal

import "github.com/zclconf/go-cty/cty"

type SchemaBlock struct {
	Name     string               `hcl:"name,label"`
	IsList   bool                 `hcl:"name"`
	Format   map[string]cty.Value `hcl:"format"`
	Variants []VariantBlock       `hcl:"variant,block"`
}

type VariantBlock struct {
	Name   string               `hcl:"name,label"`
	Key    string               `hcl:"key"`
	Format map[string]cty.Value `hcl:"format"`
}
