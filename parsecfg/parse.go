package parsecfg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/switchboard-org/switchboard/providers"
)

type Parser interface {
	Parse() (*RootSwitchboardConfig, hcl.Diagnostics)
	Init() hcl.Diagnostics
}

type DefaultParser struct {
	workingDir string
	varFile    string
	version    string
}

func NewDefaultParser(workingDir string, varFile string, version string) Parser {
	return &DefaultParser{
		workingDir: workingDir,
		varFile:    varFile,
		version:    version,
	}
}

// Parse runs through the user provided config & calculates any expressions, returning a near completely
// decoded config struct. Some values will remain expressions as they are only known in a separate process.
// It will short circuit with any errors and return to the caller if necessary.
func (p *DefaultParser) Parse() (*RootSwitchboardConfig, hcl.Diagnostics) {
	var switchboardConfig RootSwitchboardConfig
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
	//process config switchboard global step
	//check if providers are downloaded
	//remain := variableConfig.Remain
	return &switchboardConfig, nil
}

// Init is responsible for pulling down any required providers for the CLI to use,
// and anything else that needs to be setup before parsing can be done
func (p *DefaultParser) Init() hcl.Diagnostics {
	var switchboardConfig RootSwitchboardConfig
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

func (p *DefaultParser) parseVariableBlocks(body hcl.Body) ([]VariableBlock, hcl.Diagnostics) {
	var diag hcl.Diagnostics
	variableOverrides := getVariableDataFromJSONFile(p.varFile)
	var variablesParser variableBlocksParser
	diag = gohcl.DecodeBody(body, nil, &variablesParser)
	if diag.HasErrors() {
		return nil, diag
	}
	return variablesParser.parse(variableOverrides)
}

func (p *DefaultParser) parseSwitchboardBlock(body hcl.Body, ctx *hcl.EvalContext, init bool) (*SwitchboardBlock, hcl.Diagnostics) {
	switchboardStepParser := switchboardBlockParser{
		downloader: providers.NewDefaultDownloader(),
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
