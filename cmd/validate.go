package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var cmdValidate = &cobra.Command{
	Use:   "validate",
	Short: "Validate your workflow configuration",
	Long:  "Validates your workflow configuration files, including variables, providers, triggers, and steps",
	Run:   validate,
}

func validate(cmd *cobra.Command, args []string) {
	_, diag := parser.Parse()
	if diag.HasErrors() {
		for _, err := range diag.Errs() {
			log.Println(err)
		}
		return
	}
}
