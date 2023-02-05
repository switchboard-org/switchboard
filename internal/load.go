package internal

// Load is responsible for validating all config settings, and will throw an error
// if anything is incorrectly configured.
func Load(workingDir string, varFile string) {
	rawBody := loadAllHclFilesInDir(workingDir)

	//validate custom variable blocks
	variableOverrides := getVariableDataFromJSONFile(varFile)
	var variableConfig variableStepConfig
	variableConfig.Decode(rawBody)
	variableConfig.GetAllVariables(variableOverrides)
}
