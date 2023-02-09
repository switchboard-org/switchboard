package internal

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"switchboard/internal/shared"
)

// Load goes through the user provided config & calculates any expressions, returning a nearly completely
// decoded config struct. Some values will remain expressions as they are only known in a separate process.
// It will short circuit with any errors and return to the caller if necessary.
func Load(workingDir string, varFile string) (*shared.SwitchboardConfig, hcl.Diagnostics) {
	rawBody := loadAllHclFilesInDir(workingDir)

	//validate custom variable blocks
	variableOverrides := getVariableDataFromJSONFile(varFile)
	var variableConfig variableStepConfig
	diag := gohcl.DecodeBody(rawBody, nil, variableConfig)
	if diag.HasErrors() {
		return nil, diag
	}
	_, diag = variableConfig.CalculatedVariables(variableOverrides)
	if diag.HasErrors() {
		return nil, diag
	}
	//check if providers are downloaded
	//remain := variableConfig.Remain
	return nil, nil
}
