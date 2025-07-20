package main

import (
	"github.com/inference-gateway/a2a-cli/cmd"
)

// Version is set at build time
var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
