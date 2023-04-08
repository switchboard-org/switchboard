/*
Package main is the entrypoint for the switchboard CLI. The Switchboard CLI
is the main tool used for validating, processing, and deploying workflows
created using the switchboard workflow specification. All global settings
and workflows are build inside of HCL (Hashicorp Configuration Language) files.
*/
package main

import "github.com/switchboard-org/switchboard/cmd"

// Version will get set to an appropriate tag during the release build...
// (https://belief-driven-design.com/build-time-variables-in-go-51439b26ef9/)
var Version = "development"

func main() {
	cmd.Execute(Version)
}
