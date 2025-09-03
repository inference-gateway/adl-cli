package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inference-gateway/adl-cli/internal/generator"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate A2A agent code from ADL file",
	Long:  `Generate complete A2A agent project structure from an Agent Definition Language (ADL) file.`,
	RunE:  runGenerate,
}

var (
	adlFile            string
	outputDir          string
	template           string
	overwrite          bool
	generateCI         bool
	generateCD         bool
	deploymentType     string
	enableFlox         bool
	enableDevContainer bool
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&adlFile, "file", "f", "agent.yaml", "ADL file to generate from")
	generateCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated code")
	generateCmd.Flags().StringVarP(&template, "template", "t", "minimal", "Template to use (minimal)")
	generateCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files")
	generateCmd.Flags().BoolVar(&generateCI, "ci", false, "Generate CI workflow configuration")
	generateCmd.Flags().BoolVar(&generateCD, "cd", false, "Generate CD pipeline configuration with semantic-release")
	generateCmd.Flags().StringVar(&deploymentType, "deployment", "", "Deployment type (kubernetes, defaults to empty for no deployment)")
	generateCmd.Flags().BoolVar(&enableFlox, "flox", false, "Enable Flox environment")
	generateCmd.Flags().BoolVar(&enableDevContainer, "devcontainer", false, "Enable DevContainer environment")
}

func runGenerate(cmd *cobra.Command, args []string) error {
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
		Template:           template,
		Overwrite:          overwrite,
		Version:            version,
		GenerateCI:         generateCI,
		GenerateCD:         generateCD,
		DeploymentType:     deploymentType,
		EnableFlox:         enableFlox,
		EnableDevContainer: enableDevContainer,
	})

	fmt.Printf("Generating A2A agent from '%s' to '%s'\n", absADLFile, absOutputDir)
	fmt.Printf("Using template: %s\n", template)
	if generateCI {
		fmt.Printf("CI workflow generation: enabled\n")
	}
	if generateCD {
		fmt.Printf("CD pipeline generation: enabled\n")
	}
	if deploymentType != "" {
		fmt.Printf("Deployment type: %s\n", deploymentType)
	} else {
		fmt.Printf("Deployment left empty - no deployment files generated\n")
	}
	if enableFlox {
		fmt.Printf("Flox environment: enabled\n")
	}
	if enableDevContainer {
		fmt.Printf("DevContainer environment: enabled\n")
	}

	if err := gen.Generate(absADLFile, absOutputDir); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Println("✅ A2A agent generated successfully!")
	fmt.Printf("📁 Project location: %s\n", absOutputDir)

	fmt.Println()
	fmt.Println("📝 Next steps:")
	fmt.Println("   1. Implement the TODO placeholders in the generated files")
	fmt.Println("   2. Run 'task build' to build your agent")
	fmt.Println("   3. Run 'task run' to start your agent server")

	return nil
}
