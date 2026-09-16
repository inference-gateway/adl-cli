package main

import (
	cmd "github.com/inference-gateway/adl-cli/cmd"
)

// Version is set at build time
var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
