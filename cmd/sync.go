package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inference-gateway/a2a-cli/internal/generator"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update generated code while preserving implementations",
	Long: `Synchronize generated code with the ADL file while preserving your business logic.

This command updates the generated scaffolding code but preserves any
implementations you've added to the TODO placeholders.`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVarP(&adlFile, "file", "f", "agent.yaml", "ADL file to sync from")
	syncCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory containing the project")
}

func runSync(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(adlFile); os.IsNotExist(err) {
		return fmt.Errorf("ADL file '%s' does not exist", adlFile)
	}

	absADLFile, err := filepath.Abs(adlFile)
	if err != nil {
		return fmt.Errorf("failed to resolve ADL file path: %w", err)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output directory path: %w", err)
	}

	gen := generator.New(generator.Config{
		Template:  "",
		Overwrite: false,
		SyncMode:  true,
	})

	fmt.Printf("Syncing A2A agent from '%s' to '%s'\n", absADLFile, absOutputDir)

	if err := gen.Generate(absADLFile, absOutputDir); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	fmt.Println("✅ A2A agent synced successfully!")
	fmt.Println("🔄 Generated code updated while preserving your implementations")

	return nil
}
