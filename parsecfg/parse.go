package parsecfg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/switchboard-org/switchboard/internal"
	"github.com/switchboard-org/switchboard/providers"
)

type Parser interface {
	Parse() (*internal.RootSwitchboardConfig, hcl.Diagnostics)
	Init() hcl.Diagnostics
}

type DefaultParser struct {
	workingDir    string
	varFile       string
	version       string
	pluginManager internal.PluginManager
}

func NewDefaultParser(workingDir string, varFile string, version string) Parser {
	return &DefaultParser{
		workingDir:    workingDir,
		varFile:       varFile,
		version:       version,
		pluginManager: internal.NewDefaultPluginManager(),
	}
}

// Parse runs through the user provided config & calculates any expressions, returning a near completely
// decoded config struct. Some values will remain expressions as they are only known in a separate process.
// It will short circuit with any errors and return to the caller if necessary.
func (p *DefaultParser) Parse() (*internal.RootSwitchboardConfig, hcl.Diagnostics) {
	var switchboardConfig internal.RootSwitchboardConfig
	rawBody, diag := loadAllHclFilesInDir(p.workingDir)
	if diag.HasErrors() {
		return nil, diag
	}
	vars, diag := p.parseVariableBlocks(rawBody)
	if diag.HasErrors() {
		return nil, diag
	}
	switchboardConfig.Variables = vars

	switchboardBlock, diag := p.parseSwitchboardBlock(rawBody, switchboardConfig.EvalContext(), false)

	if diag.HasErrors() {
		return nil, diag
	}
	switchboardConfig.Switchboard = *switchboardBlock

	//load providers which will be used to validate a number of different blocks (provider, trigger, workflow actions, etc.)
	for _, requiredProvider := range switchboardBlock.RequiredProviders {
		err := p.pluginManager.LoadPlugin(requiredProvider)
		if err != nil {
			diag = diag.Append(simpleDiagnostic("could not load plugin", err.Error(), nil))
		}
	}
	providerBlocks, diag := p.parseProviderBlocks(rawBody, switchboardConfig.EvalContext())
	if diag.HasErrors() {
		return nil, diag
	}
	switchboardConfig.Providers = providerBlocks
	defer p.pluginManager.KillAllPlugins()

	schemaBlocks, diag := p.parseSchemaBlocks(rawBody, switchboardConfig.EvalContext())
	if diag.HasErrors() {
		return nil, diag
	}
	switchboardConfig.Schemas = schemaBlocks
	//process config switchboard global step
	//check if providers are downloaded
	//remain := variableConfig.Remain
	return &switchboardConfig, nil
}

// Init is responsible for pulling down any required providers for the CLI to use,
// and anything else that needs to be setup before parsing can be done
func (p *DefaultParser) Init() hcl.Diagnostics {
	var switchboardConfig internal.RootSwitchboardConfig
	rawBody, diag := loadAllHclFilesInDir(p.workingDir)
	if diag.HasErrors() {
		return diag
	}
	vars, diag := p.parseVariableBlocks(rawBody)
	if diag.HasErrors() {
		return diag
	}
	switchboardConfig.Variables = vars

	_, diag = p.parseSwitchboardBlock(rawBody, switchboardConfig.EvalContext(), true)
	return diag
}

func (p *DefaultParser) parseVariableBlocks(body hcl.Body) ([]internal.VariableBlock, hcl.Diagnostics) {
	var diag hcl.Diagnostics
	variableOverrides := getVariableDataFromJSONFile(p.varFile)
	var variablesParser variableBlocksParser
	diag = gohcl.DecodeBody(body, nil, &variablesParser)
	if diag.HasErrors() {
		return nil, diag
	}
	return variablesParser.parse(variableOverrides)
}

func (p *DefaultParser) parseSwitchboardBlock(body hcl.Body, ctx *hcl.EvalContext, init bool) (*internal.SwitchboardBlock, hcl.Diagnostics) {
	switchboardStepParser := switchboardBlockParser{
		downloader: providers.NewDefaultDownloader(),
		osManager:  internal.NewDefaultOsManager(),
	}
	diag := gohcl.DecodeBody(body, ctx, &switchboardStepParser.config)
	if diag.HasErrors() {
		return nil, diag
	}
	if init {
		return switchboardStepParser.init(p.version, ctx)
	}
	return switchboardStepParser.parse(p.version, ctx, true)
}

func (p *DefaultParser) parseProviderBlocks(body hcl.Body, ctx *hcl.EvalContext) ([]internal.ProviderBlock, hcl.Diagnostics) {
	providersStepParser := providerBlocksParser{
		pluginManager: p.pluginManager,
	}
	diag := gohcl.DecodeBody(body, ctx, &providersStepParser.config)
	if diag.HasErrors() {
		return nil, diag
	}
	return providersStepParser.parse(ctx)
}

func (p *DefaultParser) parseSchemaBlocks(body hcl.Body, ctx *hcl.EvalContext) ([]internal.SchemaBlock, hcl.Diagnostics) {
	schemaStepParser := schemaBlockParser{}
	diag := gohcl.DecodeBody(body, schemaEvalContext(ctx), &schemaStepParser.schemaConfigs)
	if diag.HasErrors() {
		return nil, diag
	}
	return schemaStepParser.parse()
}
