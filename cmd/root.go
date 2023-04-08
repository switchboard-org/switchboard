package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/switchboard-org/switchboard/internal"
	"github.com/switchboard-org/switchboard/parsecfg"
	"os"
)

var (
	workingDir        = internal.CurrentWorkingDir()
	parser            parsecfg.Parser
	varDefinitionFile string
	rootCmd           = &cobra.Command{
		Use:   "switchboard",
		Short: "Switchboard is a workflow automation scripting tool",
		Long:  `Switchboard is an open-source, configuration-based, highly extensible, parallelized workflow automation tool built for developers who want to build workflow with ease, without losing the control they care about. See the docs at github.com/switchboard-org/switchboard`,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&varDefinitionFile, "var-file", "./variables.json", "env file with variable values set")
}

// Execute is the primary entrypoint for the CLI
func Execute(version string) {
	rootCmd.Version = version
	parser = parsecfg.NewDefaultParser(workingDir, varDefinitionFile, version)
	rootCmd.AddCommand(cmdValidate)
	rootCmd.AddCommand(cmdServe)
	rootCmd.AddCommand(cmdInit)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
