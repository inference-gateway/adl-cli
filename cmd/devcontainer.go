package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inference-gateway/a2a-cli/internal/devcontainer"
	"github.com/spf13/cobra"
)

// devcontainerCmd represents the devcontainer command
var devcontainerCmd = &cobra.Command{
	Use:   "devcontainer",
	Short: "Generate VS Code devcontainer configuration",
	Long: `Generate VS Code devcontainer configuration files for your A2A agent project.

This command creates:
- .devcontainer/devcontainer.json - VS Code devcontainer configuration
- .devcontainer/Dockerfile - Language-specific development environment

The generated Dockerfile includes the A2A CLI and all necessary tools for
developing A2A agents in the specified programming language.`,
	RunE: runDevcontainer,
}

var (
	devcontainerADLFile string
	devcontainerOutputDir string
)

func init() {
	developmentCmd.AddCommand(devcontainerCmd)

	devcontainerCmd.Flags().StringVarP(&devcontainerADLFile, "file", "f", "agent.yaml", "ADL file to read language configuration from")
	devcontainerCmd.Flags().StringVarP(&devcontainerOutputDir, "output", "o", ".", "Output directory for devcontainer files")
}

func runDevcontainer(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(devcontainerADLFile); os.IsNotExist(err) {
		return fmt.Errorf("ADL file '%s' does not exist", devcontainerADLFile)
	}

	absADLFile, err := filepath.Abs(devcontainerADLFile)
	if err != nil {
		return fmt.Errorf("failed to resolve ADL file path: %w", err)
	}

	absOutputDir, err := filepath.Abs(devcontainerOutputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output directory path: %w", err)
	}

	generator := devcontainer.New()

	fmt.Printf("Generating devcontainer configuration from '%s' to '%s'\n", absADLFile, absOutputDir)

	if err := generator.Generate(absADLFile, absOutputDir); err != nil {
		return fmt.Errorf("devcontainer generation failed: %w", err)
	}

	fmt.Println("✅ Devcontainer configuration generated successfully!")
	fmt.Printf("📁 Files created in: %s/.devcontainer/\n", absOutputDir)
	fmt.Println("📝 Next steps:")
	fmt.Println("   1. Open project in VS Code")
	fmt.Println("   2. Install the 'Dev Containers' extension")
	fmt.Println("   3. Run 'Dev Containers: Reopen in Container' command")

	return nil
}
