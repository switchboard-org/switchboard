package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"switchboard/internal"
)

var (
	workingDir = internal.CurrentWorkingDir()
	varDefFile string
	rootCmd    = &cobra.Command{
		Use:   "switchboard",
		Short: "Switchboard is a workflow automation scripting tool",
		Long:  `Switchboard is an open-source, configuration-based, highly extensible, parallelized workflow automation tool built for developers who want to build workflow with ease, without losing the control they care about. See the docs at github.com/switchboard-org/switchboard`,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&varDefFile, "var-file", "./.switchboard/variables.json", "env file with variable values set")
}

func Execute() {
	rootCmd.AddCommand(cmdValidate)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
