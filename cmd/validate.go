package cmd

import (
	"github.com/spf13/cobra"
	"switchboard/internal"
)

var cmdValidate = &cobra.Command{
	Use:   "validate",
	Short: "Validate your workflow configuration",
	Long:  "Validates your workflow configuration files, including variables, providers, triggers, and steps",
	Run:   Validate,
}

func Validate(cmd *cobra.Command, args []string) {
	internal.Load(workingDir, varDefFile)
}
