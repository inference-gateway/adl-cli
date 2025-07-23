package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new A2A agent project interactively",
	Long: `Initialize a new A2A agent project with an interactive wizard.

This command guides you through creating an Agent Definition Language (ADL) file
and generates the initial project structure.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("🚀 A2A Agent Project Initialization")
	fmt.Println("=====================================")
	fmt.Println()

	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	} else {
		projectName = promptString(scanner, "Project name", "my-agent")
	}

	projectDir := filepath.Join(".", projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	adl := collectADLInfo(scanner, projectName)

	adlFile := filepath.Join(projectDir, "agent.yaml")
	if err := writeADLFile(adl, adlFile); err != nil {
		return fmt.Errorf("failed to write ADL file: %w", err)
	}

	fmt.Printf("\n✅ ADL file created: %s\n", adlFile)

	fmt.Println("🔨 Generating project structure...")

	if err := generateCmd.Flags().Set("file", adlFile); err != nil {
		return fmt.Errorf("failed to set file flag: %w", err)
	}
	if err := generateCmd.Flags().Set("output", projectDir); err != nil {
		return fmt.Errorf("failed to set output flag: %w", err)
	}
	if err := generateCmd.Flags().Set("template", adl.getTemplate()); err != nil {
		return fmt.Errorf("failed to set template flag: %w", err)
	}

	if err := runGenerate(generateCmd, []string{}); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	fmt.Println()
	fmt.Printf("🎉 Project '%s' initialized successfully!\n", projectName)
	fmt.Printf("📁 Project location: %s\n", projectDir)
	fmt.Println()
	fmt.Println("📝 Next steps:")
	fmt.Printf("   1. cd %s\n", projectName)
	fmt.Println("   2. Implement the TODO placeholders in the generated files")
	fmt.Println("   3. Run 'task build' to build your agent")
	fmt.Println("   4. Run 'task run' to start your agent server")

	return nil
}

type adlData struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
	} `yaml:"metadata"`
	Spec struct {
		Capabilities *struct {
			Streaming              bool `yaml:"streaming"`
			PushNotifications      bool `yaml:"pushNotifications"`
			StateTransitionHistory bool `yaml:"stateTransitionHistory"`
		} `yaml:"capabilities,omitempty"`
		Agent *struct {
			Provider     string  `yaml:"provider"`
			Model        string  `yaml:"model"`
			SystemPrompt string  `yaml:"systemPrompt,omitempty"`
			MaxTokens    int     `yaml:"maxTokens,omitempty"`
			Temperature  float64 `yaml:"temperature,omitempty"`
		} `yaml:"agent,omitempty"`
		Tools []struct {
			Name        string                 `yaml:"name"`
			Description string                 `yaml:"description"`
			Schema      map[string]interface{} `yaml:"schema"`
		} `yaml:"tools,omitempty"`
		Server struct {
			Port  int  `yaml:"port"`
			Debug bool `yaml:"debug"`
		} `yaml:"server"`
		Language *struct {
			Go *struct {
				Module  string `yaml:"module"`
				Version string `yaml:"version"`
			} `yaml:"go,omitempty"`
		} `yaml:"language,omitempty"`
	} `yaml:"spec"`
}

func (a *adlData) getTemplate() string {
	if a.Spec.Agent == nil || a.Spec.Agent.Provider == "none" {
		return "minimal"
	}
	return "ai-powered"
}

func collectADLInfo(scanner *bufio.Scanner, projectName string) *adlData {
	adl := &adlData{
		APIVersion: "a2a.dev/v1",
		Kind:       "Agent",
	}

	fmt.Println("📋 Agent Metadata")
	fmt.Println("-----------------")
	adl.Metadata.Name = promptString(scanner, "Agent name", projectName)
	adl.Metadata.Description = promptString(scanner, "Agent description", "A helpful AI agent")
	adl.Metadata.Version = promptString(scanner, "Version", "1.0.0")

	fmt.Println("\n🤖 Agent Type")
	fmt.Println("--------------")
	agentType := promptChoice(scanner, "Agent type", []string{"ai-powered", "minimal"}, "ai-powered")

	if agentType == "ai-powered" {
		fmt.Println("\n🧠 AI Configuration")
		fmt.Println("-------------------")

		adl.Spec.Agent = &struct {
			Provider     string  `yaml:"provider"`
			Model        string  `yaml:"model"`
			SystemPrompt string  `yaml:"systemPrompt,omitempty"`
			MaxTokens    int     `yaml:"maxTokens,omitempty"`
			Temperature  float64 `yaml:"temperature,omitempty"`
		}{}

		provider := promptChoice(scanner, "AI Provider", []string{"openai", "anthropic", "azure", "ollama", "deepseek"}, "openai")
		adl.Spec.Agent.Provider = provider

		var defaultModel string
		switch provider {
		case "openai":
			defaultModel = "gpt-4o-mini"
		case "anthropic":
			defaultModel = "claude-3-haiku-20240307"
		case "azure":
			defaultModel = "gpt-4o"
		case "ollama":
			defaultModel = "llama3.1"
		case "deepseek":
			defaultModel = "deepseek-chat"
		}

		adl.Spec.Agent.Model = promptString(scanner, "Model", defaultModel)

		for {
			systemPrompt := promptString(scanner, "System prompt", "You are a helpful AI assistant.")
			if systemPrompt != "" {
				adl.Spec.Agent.SystemPrompt = systemPrompt
				break
			}
			fmt.Println("⚠️  System prompt is required for AI-powered agents. Please provide a system prompt.")
		}

		if maxTokensStr := promptString(scanner, "Max tokens (optional, press enter to skip)", ""); maxTokensStr != "" {
			if maxTokens, err := strconv.Atoi(maxTokensStr); err == nil {
				adl.Spec.Agent.MaxTokens = maxTokens
			}
		}

		if tempStr := promptString(scanner, "Temperature (0.0-2.0, optional)", ""); tempStr != "" {
			if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
				adl.Spec.Agent.Temperature = temp
			}
		}
	}

	fmt.Println("\n⚡ Capabilities")
	fmt.Println("---------------")
	adl.Spec.Capabilities = &struct {
		Streaming              bool `yaml:"streaming"`
		PushNotifications      bool `yaml:"pushNotifications"`
		StateTransitionHistory bool `yaml:"stateTransitionHistory"`
	}{}

	adl.Spec.Capabilities.Streaming = promptBool(scanner, "Enable streaming", true)
	adl.Spec.Capabilities.PushNotifications = promptBool(scanner, "Enable push notifications", false)
	adl.Spec.Capabilities.StateTransitionHistory = promptBool(scanner, "Enable state transition history", false)

	fmt.Println("\n🔧 Tools")
	fmt.Println("--------")
	addTools := promptBool(scanner, "Add tools to your agent", false)

	if addTools {
		for {
			tool := struct {
				Name        string                 `yaml:"name"`
				Description string                 `yaml:"description"`
				Schema      map[string]interface{} `yaml:"schema"`
			}{}

			tool.Name = promptString(scanner, "Tool name (e.g., 'get_weather')", "")
			if tool.Name == "" {
				break
			}

			tool.Description = promptString(scanner, "Tool description", "")

			tool.Schema = map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"input": map[string]interface{}{
						"type":        "string",
						"description": "Input parameter for " + tool.Name,
					},
				},
				"required": []string{"input"},
			}

			adl.Spec.Tools = append(adl.Spec.Tools, tool)

			if !promptBool(scanner, "Add another tool", false) {
				break
			}
		}
	}

	fmt.Println("\n🌐 Server Configuration")
	fmt.Println("-----------------------")
	portStr := promptString(scanner, "Server port", "8080")
	if port, err := strconv.Atoi(portStr); err == nil {
		adl.Spec.Server.Port = port
	} else {
		adl.Spec.Server.Port = 8080
	}
	adl.Spec.Server.Debug = promptBool(scanner, "Enable debug mode", false)

	fmt.Println("\n🐹 Go Configuration")
	fmt.Println("-------------------")

	adl.Spec.Language = &struct {
		Go *struct {
			Module  string `yaml:"module"`
			Version string `yaml:"version"`
		} `yaml:"go,omitempty"`
	}{}

	adl.Spec.Language.Go = &struct {
		Module  string `yaml:"module"`
		Version string `yaml:"version"`
	}{}

	adl.Spec.Language.Go.Module = promptString(scanner, "Go module", fmt.Sprintf("github.com/example/%s", adl.Metadata.Name))
	adl.Spec.Language.Go.Version = promptString(scanner, "Go version", "1.24")

	return adl
}

func promptString(scanner *bufio.Scanner, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if input == "" {
		return defaultValue
	}
	return input
}

func promptBool(scanner *bufio.Scanner, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	fmt.Printf("%s [y/n] (default: %s): ", prompt, defaultStr)
	scanner.Scan()
	input := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes"
}

func promptChoice(scanner *bufio.Scanner, prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s (%s) [%s]: ", prompt, strings.Join(choices, "/"), defaultValue)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if input == "" {
		return defaultValue
	}

	for _, choice := range choices {
		if input == choice {
			return input
		}
	}

	return defaultValue
}

func writeADLFile(adl *adlData, filePath string) error {
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(adl); err != nil {
		_ = encoder.Close()
		return err
	}

	if err := encoder.Close(); err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(buf.String()), 0644)
}
