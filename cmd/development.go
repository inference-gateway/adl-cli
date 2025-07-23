package cmd

import (
	"github.com/spf13/cobra"
)

// developmentCmd represents the development command
var developmentCmd = &cobra.Command{
	Use:   "development",
	Short: "Commands for development environment setup",
	Long: `Development commands help you set up and manage development environments
for your A2A agents. This includes generating devcontainer configurations,
cloud development environments, and other development tooling.`,
}

func init() {
	rootCmd.AddCommand(developmentCmd)
}
