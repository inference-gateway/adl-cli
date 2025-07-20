package cmd

import (
	"fmt"
	"os"

	"github.com/inference-gateway/a2a-cli/internal/schema"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [adl-file]",
	Short: "Validate an ADL file against the schema",
	Long: `Validate an Agent Definition Language (ADL) file against the official schema.

This command checks if your ADL file follows the correct structure and
contains all required fields for generating a valid A2A agent.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	adlFile := "agent.yaml"
	if len(args) > 0 {
		adlFile = args[0]
	}

	// Check if file exists
	if _, err := os.Stat(adlFile); os.IsNotExist(err) {
		return fmt.Errorf("ADL file '%s' does not exist", adlFile)
	}

	fmt.Printf("Validating '%s'...\n", adlFile)

	// Validate the file
	validator := schema.NewValidator()
	if err := validator.ValidateFile(adlFile); err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ '%s' is valid!\n", adlFile)
	return nil
}
