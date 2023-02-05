/*
Package main is the entrypoint for the switchboard CLI. The Switchboard CLI
is the main tool used for validating, processing, and deploying workflows
created using the switchboard workflow specification. All global settings
and workflows are build inside of HCL (Hashicorp Configuration Language) files.
*/
package main

import "switchboard/cmd"

func main() {
	cmd.Execute()
}
