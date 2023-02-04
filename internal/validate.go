package internal

import (
	"github.com/spf13/cobra"
)

func Validate(cmd *cobra.Command, args []string) {
	rawBody := loadAllHclFilesIntoParsedBody()

	//validate custom variable blocks
	varDefFile := cmd.Flag("var-file").Value.String()
	variableOverrides := getVariableDataFromJSONFile(varDefFile)
	var variableConfig VariablesConfig
	variableConfig.Decode(rawBody)
	variableConfig.GetAllVariables(variableOverrides)

	//validate switchboard block (make sure all provider binaries are pulled down)

	//validate all triggers and workflows
}
