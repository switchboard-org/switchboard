package cmd

import (
	"github.com/spf13/cobra"
	"github.com/switchboard-org/switchboard/server"
	"log"
)

var cmdServe = &cobra.Command{
	Use:   "run",
	Short: "Start a server that runs your workflows",
	Long:  "Parses your configuration and runs a server that registers triggers and process workflows",
	Run:   serve,
}

func serve(cmd *cobra.Command, args []string) {
	//this is a long-running call. Only exits on failure or when shutdown request received
	err := server.StartServer(parser)
	if err != nil {
		log.Println(err)
	}
}
