package internal

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/switchboard-org/switchboard/internal/shared"
)

// LoadConfig goes through the user provided config & calculates any expressions, returning a nearly completely
// decoded config struct. Some values will remain expressions as they are only known in a separate process.
// It will short circuit with any errors and return to the caller if necessary.
func LoadConfig(workingDir string, varFile string) (*shared.SwitchboardConfig, hcl.Diagnostics) {
	var switchboardConfig shared.SwitchboardConfig
	rawBody, diag := loadAllHclFilesInDir(workingDir)
	if diag.HasErrors() {
		return nil, diag
	}
	//process config variable blocks, and override when necessary
	variableOverrides := getVariableDataFromJSONFile(varFile)
	var variableConfig variableStepConfig
	diag = gohcl.DecodeBody(rawBody, nil, variableConfig)
	if diag.HasErrors() {
		return nil, diag
	}
	vars, diag := variableConfig.CalculatedVariables(variableOverrides)
	if diag.HasErrors() {
		return nil, diag
	}
	switchboardConfig.Variables = vars

	//process config switchboard global step
	//check if providers are downloaded
	//remain := variableConfig.Remain
	return &switchboardConfig, nil
}
