package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/inference-gateway/adl-cli/internal/prompt"
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
	fmt.Println("\n🚀 A2A Agent Project Initialization")
	fmt.Println("=====================================")
	fmt.Println()

	projectDir := promptString("Project directory (relative or absolute path)", ".")

	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	} else {
		projectName = getProjectNameFromGit()
		if projectName == "" {
			if projectDir == "." {
				cwd, _ := os.Getwd()
				projectName = filepath.Base(cwd)
			} else {
				projectName = filepath.Base(projectDir)
			}
		}
	}
	if projectDir != "." && projectDir != "" {
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}
	}

	fmt.Printf("\n📋 ADL Schema Setup")
	fmt.Println("-------------------")

	var adl *adlData
	var adlFile string

	useExisting := promptBool("Use an existing ADL schema file", false)

	if useExisting {
		for {
			existingFile := promptString("Path to existing ADL schema file (relative or absolute)", "")
			if existingFile == "" {
				fmt.Println("⚠️  ADL file path is required. Please provide a path to the existing schema file.")
				continue
			}

			if !filepath.IsAbs(existingFile) {
				cwd, _ := os.Getwd()
				existingFile = filepath.Join(cwd, existingFile)
			}

			if _, err := os.Stat(existingFile); os.IsNotExist(err) {
				fmt.Printf("⚠️  ADL file does not exist: %s\n", existingFile)
				fmt.Println("Please provide a valid path to an existing ADL schema file.")
				continue
			}

			existingADL, err := readADLFile(existingFile)
			if err != nil {
				fmt.Printf("⚠️  Failed to read ADL file: %v\n", err)
				fmt.Println("Please provide a valid ADL schema file.")
				continue
			}

			adl = existingADL
			adlFile = filepath.Join(projectDir, "agent.yaml")

			if err := writeADLFile(adl, adlFile); err != nil {
				return fmt.Errorf("failed to write ADL file: %w", err)
			}

			fmt.Printf("✅ Using existing ADL schema from: %s\n", existingFile)
			break
		}
	} else {
		fmt.Printf("\n")
		adl = collectADLInfo(projectName)
		adlFile = filepath.Join(projectDir, "agent.yaml")

		if err := writeADLFile(adl, adlFile); err != nil {
			return fmt.Errorf("failed to write ADL file: %w", err)
		}
	}

	fmt.Printf("\n✅ ADL file created: %s\n", adlFile)

	fmt.Println("🔨 Generating project structure...")

	overwrite := false
	if projectDir == "." || (projectDir != "" && dirExists(projectDir)) {
		overwrite = promptBool("Overwrite existing files", true)
	}

	if err := generateCmd.Flags().Set("file", adlFile); err != nil {
		return fmt.Errorf("failed to set file flag: %w", err)
	}
	if err := generateCmd.Flags().Set("output", projectDir); err != nil {
		return fmt.Errorf("failed to set output flag: %w", err)
	}
	if err := generateCmd.Flags().Set("template", adl.getTemplate()); err != nil {
		return fmt.Errorf("failed to set template flag: %w", err)
	}
	if err := generateCmd.Flags().Set("overwrite", fmt.Sprintf("%t", overwrite)); err != nil {
		return fmt.Errorf("failed to set overwrite flag: %w", err)
	}

	if err := runGenerate(generateCmd, []string{}); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	fmt.Println()
	fmt.Printf("🎉 Project '%s' initialized successfully!\n", projectName)
	if projectDir == "." {
		fmt.Printf("📁 Project location: current directory\n")
		fmt.Println()
		fmt.Println("📝 Next steps:")
		fmt.Println("   1. Implement the TODO placeholders in the generated files")
		fmt.Println("   2. Run 'task build' to build your agent")
		fmt.Println("   3. Run 'task run' to start your agent server")
	} else {
		fmt.Printf("📁 Project location: %s\n", projectDir)
		fmt.Println()
		fmt.Println("📝 Next steps:")
		fmt.Printf("   1. cd %s\n", projectDir)
		fmt.Println("   2. Implement the TODO placeholders in the generated files")
		fmt.Println("   3. Run 'task build' to build your agent")
		fmt.Println("   4. Run 'task run' to start your agent server")
	}

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
			TypeScript *struct {
				PackageName string `yaml:"packageName"`
				NodeVersion string `yaml:"nodeVersion"`
			} `yaml:"typescript,omitempty"`
			Rust *struct {
				PackageName string `yaml:"packageName"`
				Version     string `yaml:"version"`
				Edition     string `yaml:"edition"`
			} `yaml:"rust,omitempty"`
		} `yaml:"language,omitempty"`
	} `yaml:"spec"`
}

func (a *adlData) getTemplate() string {
	if a.Spec.Agent == nil || a.Spec.Agent.Provider == "none" {
		return "minimal"
	}
	return "ai-powered"
}

func collectADLInfo(projectName string) *adlData {
	adl := &adlData{
		APIVersion: "adl.dev/v1",
		Kind:       "Agent",
	}

	fmt.Println("📋 Agent Metadata")
	fmt.Println("-----------------")
	adl.Metadata.Name = promptString("Agent name", projectName)
	adl.Metadata.Description = promptString("Agent description", "A helpful AI agent")
	adl.Metadata.Version = promptString("Version", "0.1.0")

	fmt.Println("\n🤖 Agent Type")
	fmt.Println("--------------")
	agentType := promptChoice("Agent type", []string{"ai-powered", "minimal"}, "ai-powered")

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

		provider := promptChoice("AI Provider", []string{"openai", "anthropic", "azure", "ollama", "deepseek"}, "openai")
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

		adl.Spec.Agent.Model = promptString("Model", defaultModel)

		for {
			systemPrompt := promptString("System prompt", "You are a helpful AI assistant.")
			if systemPrompt != "" {
				adl.Spec.Agent.SystemPrompt = systemPrompt
				break
			}
			fmt.Println("⚠️  System prompt is required for AI-powered agents. Please provide a system prompt.")
		}

		if maxTokensStr := promptString("Max tokens (optional, press enter to skip)", ""); maxTokensStr != "" {
			if maxTokens, err := strconv.Atoi(maxTokensStr); err == nil {
				adl.Spec.Agent.MaxTokens = maxTokens
			}
		}

		if tempStr := promptString("Temperature (0.0-2.0, optional)", ""); tempStr != "" {
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

	adl.Spec.Capabilities.Streaming = promptBool("Enable streaming", true)
	adl.Spec.Capabilities.PushNotifications = promptBool("Enable push notifications", false)
	adl.Spec.Capabilities.StateTransitionHistory = promptBool("Enable state transition history", false)

	fmt.Println("\n🔧 Tools")
	fmt.Println("--------")
	addTools := promptBool("Add tools to your agent", false)

	if addTools {
		for {
			tool := struct {
				Name        string                 `yaml:"name"`
				Description string                 `yaml:"description"`
				Schema      map[string]interface{} `yaml:"schema"`
			}{}

			tool.Name = promptString("Tool name (e.g., 'get_weather')", "")
			if tool.Name == "" {
				break
			}

			tool.Description = promptString("Tool description", "")

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

			if !promptBool("Add another tool", false) {
				break
			}
		}
	}

	fmt.Println("\n🌐 Server Configuration")
	fmt.Println("-----------------------")
	portStr := promptString("Server port", "8080")
	if port, err := strconv.Atoi(portStr); err == nil {
		adl.Spec.Server.Port = port
	} else {
		adl.Spec.Server.Port = 8080
	}
	adl.Spec.Server.Debug = promptBool("Enable debug mode", false)

	fmt.Println("\n🐹 Go Configuration")
	fmt.Println("-------------------")

	adl.Spec.Language = &struct {
		Go *struct {
			Module  string `yaml:"module"`
			Version string `yaml:"version"`
		} `yaml:"go,omitempty"`
		TypeScript *struct {
			PackageName string `yaml:"packageName"`
			NodeVersion string `yaml:"nodeVersion"`
		} `yaml:"typescript,omitempty"`
		Rust *struct {
			PackageName string `yaml:"packageName"`
			Version     string `yaml:"version"`
			Edition     string `yaml:"edition"`
		} `yaml:"rust,omitempty"`
	}{}

	adl.Spec.Language.Go = &struct {
		Module  string `yaml:"module"`
		Version string `yaml:"version"`
	}{}

	defaultModule := getDefaultGoModule(adl.Metadata.Name)
	adl.Spec.Language.Go.Module = promptString("Go module", defaultModule)
	adl.Spec.Language.Go.Version = promptString("Go version", "1.24")

	return adl
}

func parseGitRemote() (owner, repo string) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", ""
	}

	remoteURL := strings.TrimSpace(string(output))

	if strings.HasPrefix(remoteURL, "https://") {
		re := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
		matches := re.FindStringSubmatch(remoteURL)
		if len(matches) >= 3 {
			return matches[1], matches[2]
		}
	} else if strings.HasPrefix(remoteURL, "git@") {
		re := regexp.MustCompile(`git@github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
		matches := re.FindStringSubmatch(remoteURL)
		if len(matches) >= 3 {
			return matches[1], matches[2]
		}
	}

	return "", ""
}

func getProjectNameFromGit() string {
	_, repo := parseGitRemote()
	return repo
}

func getDefaultGoModule(projectName string) string {
	owner, _ := parseGitRemote()
	if owner == "" {
		return fmt.Sprintf("github.com/example/%s", projectName)
	}
	return fmt.Sprintf("github.com/%s/%s", owner, projectName)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func promptString(promptText, defaultValue string) string {
	input, err := prompt.ReadString(promptText, defaultValue)
	if err != nil {
		fmt.Println()
		os.Exit(0)
	}
	return input
}

func promptBool(promptText string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	promptWithDefault := fmt.Sprintf("%s [y/n]", promptText)
	input, err := prompt.ReadString(promptWithDefault, defaultStr)
	if err != nil {
		fmt.Println()
		os.Exit(0)
	}

	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes"
}

func promptChoice(promptText string, choices []string, defaultValue string) string {
	promptWithChoices := fmt.Sprintf("%s (%s)", promptText, strings.Join(choices, "/"))
	input, err := prompt.ReadString(promptWithChoices, defaultValue)
	if err != nil {
		fmt.Println()
		os.Exit(0)
	}

	input = strings.TrimSpace(input)
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

func readADLFile(filePath string) (*adlData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var adl adlData
	if err := yaml.Unmarshal(data, &adl); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if adl.APIVersion == "" {
		return nil, fmt.Errorf("missing apiVersion in ADL file")
	}
	if adl.Kind == "" {
		return nil, fmt.Errorf("missing kind in ADL file")
	}
	if adl.Metadata.Name == "" {
		return nil, fmt.Errorf("missing metadata.name in ADL file")
	}

	return &adl, nil
}
