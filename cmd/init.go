package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize your switchboard configuration",
	Long:  "Initializes your configuration, including retrieving plugins and creating temporary directories",
	Run:   initcfg,
}

func initcfg(cmd *cobra.Command, args []string) {
	diag := parser.Init()
	if diag.HasErrors() {
		for _, err := range diag.Errs() {
			log.Println(err)
		}
		return
	}
}
