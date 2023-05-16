package parsecfg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/switchboard-org/switchboard/internal"
)

type providerBlocksParser struct {
	config        providerBlocksConfig
	pluginManager internal.PluginManager
}

type providerBlocksConfig struct {
	Providers []providerBlockConfig `hcl:"provider,block"`
	Remain    hcl.Body              `hcl:",remain"`
}

type providerBlockConfig struct {
	//will be used for RequiredProvider if RequiredProvider not explicitly set
	Name string `hcl:"name,label"`
	//field that maps to SwitchboardBlock.RequiredProvider. By convention, will use Name field for this
	RequiredProvider *string  `hcl:"required_provider"`
	Remain           hcl.Body `hcl:",remain"`
}

func (p *providerBlocksParser) parse(ctx *hcl.EvalContext) ([]internal.ProviderBlock, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics
	var output []internal.ProviderBlock
	for _, provider := range p.config.Providers {
		providerBlock := internal.ProviderBlock{
			BlockName: provider.Name,
		}
		hclRange := provider.Remain.MissingItemRange()
		if provider.RequiredProvider == nil {
			providerBlock.ProviderName = provider.Name
		}
		pluginProvider, err := p.pluginManager.ProviderInstance(provider.Name)
		if err != nil {
			diagnostics = diagnostics.Append(simpleDiagnostic("could not get plugin provider instance", err.Error(), &hclRange))
			continue
		}
		pluginInitSchema, err := pluginProvider.InitSchema()
		if err != nil {
			diagnostics = diagnostics.Append(simpleDiagnostic("could not get schema for provider plugin", err.Error(), &hclRange))
			continue
		}
		decodedSchema := pluginInitSchema.Decode()
		val, diag := hcldec.Decode(provider.Remain, decodedSchema, ctx)
		if diag.HasErrors() {
			diagnostics = diagnostics.Extend(diag)
			continue
		}
		providerBlock.InitPayload = val
		output = append(output, providerBlock)
	}
	return output, diagnostics
}
