package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/inference-gateway/adl-cli/internal/prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new A2A agent project interactively",
	Long: `Initialize a new A2A agent project with an interactive wizard.

This command guides you through creating an Agent Definition Language (ADL) file
with your agent specifications. Use 'adl generate' afterwards to create the project code.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().Bool("defaults", false, "Use default values for all prompts")
	initCmd.Flags().String("path", "", "Project directory path")
	initCmd.Flags().String("name", "", "Agent name")
	initCmd.Flags().String("description", "", "Agent description")
	initCmd.Flags().String("version", "", "Agent version")
	initCmd.Flags().String("type", "", "Agent type (ai-powered/minimal)")
	initCmd.Flags().String("provider", "", "AI provider (openai/anthropic/groq/mistral/ollama/deepseek/cloudflare)")
	initCmd.Flags().String("model", "", "AI model")
	initCmd.Flags().String("system-prompt", "", "System prompt")
	initCmd.Flags().Int("max-tokens", 0, "Maximum tokens")
	initCmd.Flags().Float64("temperature", 0.0, "Temperature (0.0-2.0)")
	initCmd.Flags().Bool("streaming", false, "Enable streaming")
	initCmd.Flags().Bool("notifications", false, "Enable push notifications")
	initCmd.Flags().Bool("history", false, "Enable state transition history")
	initCmd.Flags().Int("port", 0, "Server port")
	initCmd.Flags().Bool("debug", false, "Enable debug mode")
	initCmd.Flags().String("language", "", "Programming language (go/rust/typescript)")
	initCmd.Flags().String("go-module", "", "Go module path")
	initCmd.Flags().String("go-version", "", "Go version")
	initCmd.Flags().String("rust-package-name", "", "Rust package name")
	initCmd.Flags().String("rust-version", "", "Rust version")
	initCmd.Flags().String("rust-edition", "", "Rust edition")
	initCmd.Flags().String("typescript-name", "", "TypeScript package name")
	initCmd.Flags().Bool("flox", false, "Enable Flox environment")
	initCmd.Flags().Bool("devcontainer", false, "Enable DevContainer environment")
	initCmd.Flags().String("deployment", "", "Deployment type (kubernetes, defaults to empty for no deployment)")

	if err := viper.BindPFlags(initCmd.Flags()); err != nil {
		fmt.Printf("Warning: failed to bind flags: %v\n", err)
	}

	viper.SetEnvPrefix("ADL")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

func runInit(cmd *cobra.Command, args []string) error {
	useDefaults, _ := cmd.Flags().GetBool("defaults")

	fmt.Println("\n🚀 A2A Agent Project Initialization")
	fmt.Println("=====================================")
	fmt.Println()

	projectDir := promptWithConfig("path", useDefaults, "Project directory (relative or absolute path)", ".")

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

	fmt.Println("\n📋 ADL Schema Setup")
	fmt.Println("-------------------")

	var adl *adlData
	var adlFile string

	useExisting := conditionalPromptBool(useDefaults, "Use an existing ADL schema file", false)

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
		adl = collectADLInfo(cmd, projectName, useDefaults)
		adlFile = filepath.Join(projectDir, "agent.yaml")

		if err := writeADLFile(adl, adlFile); err != nil {
			return fmt.Errorf("failed to write ADL file: %w", err)
		}
	}

	fmt.Printf("\n✅ ADL file created: %s\n", adlFile)

	fmt.Println()
	fmt.Printf("🎉 Project '%s' initialized successfully!\n", projectName)
	fmt.Printf("📁 ADL manifest location: %s\n", adlFile)
	fmt.Println()
	fmt.Println("📝 Next steps:")
	if projectDir == "." {
		fmt.Println("   1. Run 'adl generate' to generate the project code")
		fmt.Println("   2. Implement the TODO placeholders in the generated files")
		fmt.Println("   3. Run 'task build' to build your agent")
		fmt.Println("   4. Run 'task run' to start your agent server")
	} else {
		fmt.Printf("   1. cd %s\n", projectDir)
		fmt.Println("   2. Run 'adl generate' to generate the project code")
		fmt.Println("   3. Implement the TODO placeholders in the generated files")
		fmt.Println("   4. Run 'task build' to build your agent")
		fmt.Println("   5. Run 'task run' to start your agent server")
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
		Card *struct {
			ProtocolVersion    string   `yaml:"protocolVersion,omitempty"`
			URL                string   `yaml:"url,omitempty"`
			PreferredTransport string   `yaml:"preferredTransport,omitempty"`
			DefaultInputModes  []string `yaml:"defaultInputModes,omitempty"`
			DefaultOutputModes []string `yaml:"defaultOutputModes,omitempty"`
			DocumentationURL   string   `yaml:"documentationUrl,omitempty"`
			IconURL            string   `yaml:"iconUrl,omitempty"`
		} `yaml:"card,omitempty"`
		Agent *struct {
			Provider     string  `yaml:"provider"`
			Model        string  `yaml:"model"`
			SystemPrompt string  `yaml:"systemPrompt,omitempty"`
			MaxTokens    int     `yaml:"maxTokens,omitempty"`
			Temperature  float64 `yaml:"temperature,omitempty"`
		} `yaml:"agent,omitempty"`
		Artifacts *struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"artifacts,omitempty"`
		Services []string `yaml:"services,omitempty"`
		Skills   []struct {
			ID          string         `yaml:"id"`
			Name        string         `yaml:"name"`
			Description string         `yaml:"description"`
			Tags        []string       `yaml:"tags"`
			Schema      map[string]any `yaml:"schema"`
			Inject      []string       `yaml:"inject,omitempty"`
		} `yaml:"skills,omitempty"`
		Server struct {
			Port   int    `yaml:"port"`
			Scheme string `yaml:"scheme,omitempty"`
			Debug  bool   `yaml:"debug"`
			Auth   *struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"auth,omitempty"`
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
		SCM *struct {
			Provider       string `yaml:"provider"`
			URL            string `yaml:"url,omitempty"`
			GithubApp      bool   `yaml:"github_app,omitempty"`
			IssueTemplates bool   `yaml:"issue_templates,omitempty"`
		} `yaml:"scm,omitempty"`
		Sandbox *struct {
			Flox *struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"flox,omitempty"`
			DevContainer *struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"devcontainer,omitempty"`
		} `yaml:"sandbox,omitempty"`
		Deployment *struct {
			Type string `yaml:"type,omitempty"`
		} `yaml:"deployment,omitempty"`
		Hooks *struct {
			Post []string `yaml:"post,omitempty"`
		} `yaml:"hooks,omitempty"`
	} `yaml:"spec"`
}

func collectADLInfo(cmd *cobra.Command, projectName string, useDefaults bool) *adlData {
	adl := &adlData{
		APIVersion: "adl.dev/v1",
		Kind:       "Agent",
	}

	fmt.Println("📋 Agent Metadata")
	fmt.Println("-----------------")
	adl.Metadata.Name = promptWithConfig("name", useDefaults, "Agent name", projectName)
	adl.Metadata.Description = promptWithConfig("description", useDefaults, "Agent description", "A helpful AI agent")
	adl.Metadata.Version = promptWithConfig("version", useDefaults, "Version", "0.1.0")

	fmt.Println("\n🤖 Agent Type")
	fmt.Println("--------------")
	agentType := conditionalPromptChoice(useDefaults, "Agent type", []string{"ai-powered", "minimal"}, "ai-powered")

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

		provider := conditionalPromptChoice(useDefaults, "AI Provider", []string{"openai", "anthropic", "ollama", "deepseek", "mistral", "cloudflare", "cohere", "groq"}, "")
		adl.Spec.Agent.Provider = provider

		adl.Spec.Agent.Model = conditionalPrompt(useDefaults, "Model", "")

		systemPrompt := conditionalPrompt(useDefaults, "System prompt", "You are a helpful AI assistant.")
		adl.Spec.Agent.SystemPrompt = systemPrompt

		if maxTokensStr := conditionalPrompt(useDefaults, "Max tokens (optional, press enter to skip)", ""); maxTokensStr != "" {
			if maxTokens, err := strconv.Atoi(maxTokensStr); err == nil {
				adl.Spec.Agent.MaxTokens = maxTokens
			}
		}

		if tempStr := conditionalPrompt(useDefaults, "Temperature (0.0-2.0, optional)", ""); tempStr != "" {
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

	adl.Spec.Capabilities.Streaming = conditionalPromptBool(useDefaults, "Enable streaming", true)
	adl.Spec.Capabilities.PushNotifications = conditionalPromptBool(useDefaults, "Enable push notifications", false)
	adl.Spec.Capabilities.StateTransitionHistory = conditionalPromptBool(useDefaults, "Enable state transition history", false)

	fmt.Println("\n📂 Artifacts Configuration")
	fmt.Println("--------------------------")
	enableArtifacts := conditionalPromptBool(useDefaults, "Enable artifacts support (filesystem/MinIO storage)", false)

	if enableArtifacts {
		adl.Spec.Artifacts = &struct {
			Enabled bool `yaml:"enabled"`
		}{
			Enabled: true,
		}
		fmt.Println("ℹ️  Artifacts storage can be configured via A2A_ARTIFACT_* environment variables")
	}

	fmt.Println("\n🔌 Dependencies")
	fmt.Println("---------------")
	addDependencies := conditionalPromptBool(useDefaults, "Add dependencies for dependency injection", useDefaults)

	if addDependencies {
		if useDefaults {
			adl.Spec.Services = append(adl.Spec.Services, "logger")
			fmt.Printf("Add dependencies for dependency injection [y/n] [n]: y\n")
			fmt.Printf("Dependency name (e.g., 'logger', 'database') []: logger\n")
			fmt.Printf("✅ Added default logger dependency\n")
		}

		if !useDefaults {
			for {
				service := promptString("Service name (e.g., 'logger', 'database', 'cache', empty to finish)", "")
				if service == "" {
					break
				}

				if !isValidIdentifier(service) {
					fmt.Printf("⚠️  Invalid service name. Use only letters, numbers, and underscores, starting with a letter or underscore.\n")
					continue
				}

				duplicate := false
				for _, existing := range adl.Spec.Services {
					if existing == service {
						fmt.Printf("⚠️  Service '%s' already exists\n", service)
						duplicate = true
						break
					}
				}

				if !duplicate {
					adl.Spec.Services = append(adl.Spec.Services, service)
					fmt.Printf("✅ Added service: %s\n", service)
					fmt.Printf("💡 You will need to implement this in internal/%s package with a New%s function\n", service, titleCase(service))
				}

				if !promptBool("Add another service", false) {
					break
				}
			}
		}
	}

	fmt.Println("\n🔧 Skills")
	fmt.Println("---------")
	addSkills := conditionalPromptBool(useDefaults, "Add skills to your agent", false)

	if addSkills {
		for {
			skill := struct {
				ID          string         `yaml:"id"`
				Name        string         `yaml:"name"`
				Description string         `yaml:"description"`
				Tags        []string       `yaml:"tags"`
				Schema      map[string]any `yaml:"schema"`
				Inject      []string       `yaml:"inject,omitempty"`
			}{}

			skill.Name = promptString("Skill name (e.g., 'get_weather')", "")
			if skill.Name == "" {
				break
			}
			skill.ID = skill.Name

			skill.Description = promptString("Skill description", "")

			tagsStr := promptString("Skill tags (comma-separated, e.g., 'weather,api,data')", "")
			if tagsStr != "" {
				skill.Tags = strings.Split(tagsStr, ",")
				for i, tag := range skill.Tags {
					skill.Tags[i] = strings.TrimSpace(tag)
				}
			} else {
				skill.Tags = []string{"general"}
			}

			skill.Schema = map[string]any{
				"type": "object",
				"properties": map[string]any{
					"input": map[string]any{
						"type":        "string",
						"description": "Input parameter for " + skill.Name,
					},
				},
				"required": []string{"input"},
			}

			if len(adl.Spec.Services) > 0 {
				fmt.Printf("\nAvailable services for skill '%s':\n", skill.Name)
				for i, svc := range adl.Spec.Services {
					fmt.Printf("  %d. %s\n", i+1, svc)
				}

				addSkillServices := promptBool("Inject services into this skill", false)
				if addSkillServices {
					for {
						svcChoice := promptString("Enter service name (or empty to finish)", "")
						if svcChoice == "" {
							break
						}

						found := false
						for _, svc := range adl.Spec.Services {
							if svc == svcChoice {
								found = true
								break
							}
						}

						if found {
							alreadyAdded := false
							for _, existing := range skill.Inject {
								if existing == svcChoice {
									alreadyAdded = true
									break
								}
							}

							if !alreadyAdded {
								skill.Inject = append(skill.Inject, svcChoice)
								fmt.Printf("✅ Added service: %s\n", svcChoice)
							} else {
								fmt.Printf("⚠️  Service %s already added\n", svcChoice)
							}
						} else {
							fmt.Printf("⚠️  Service '%s' not found\n", svcChoice)
						}
					}
				}
			}

			adl.Spec.Skills = append(adl.Spec.Skills, skill)

			if !promptBool("Add another skill", false) {
				break
			}
		}
	}

	fmt.Println("\n🌐 Server Configuration")
	fmt.Println("-----------------------")
	portStr := conditionalPrompt(useDefaults, "Server port", "8080")
	if port, err := strconv.Atoi(portStr); err == nil {
		adl.Spec.Server.Port = port
	} else {
		adl.Spec.Server.Port = 8080
	}
	adl.Spec.Server.Scheme = conditionalPrompt(useDefaults, "Server scheme (http/https)", "http")
	adl.Spec.Server.Debug = conditionalPromptBool(useDefaults, "Enable debug mode", false)

	authEnabled := conditionalPromptBool(useDefaults, "Enable server authentication", false)
	if authEnabled {
		adl.Spec.Server.Auth = &struct {
			Enabled bool `yaml:"enabled"`
		}{
			Enabled: true,
		}
	}

	fmt.Println("\n🎴 Agent Card Configuration")
	fmt.Println("---------------------------")

	addCard := conditionalPromptBool(useDefaults, "Configure agent card (protocol, transport, modes)", false)
	if addCard {
		adl.Spec.Card = &struct {
			ProtocolVersion    string   `yaml:"protocolVersion,omitempty"`
			URL                string   `yaml:"url,omitempty"`
			PreferredTransport string   `yaml:"preferredTransport,omitempty"`
			DefaultInputModes  []string `yaml:"defaultInputModes,omitempty"`
			DefaultOutputModes []string `yaml:"defaultOutputModes,omitempty"`
			DocumentationURL   string   `yaml:"documentationUrl,omitempty"`
			IconURL            string   `yaml:"iconUrl,omitempty"`
		}{}

		adl.Spec.Card.ProtocolVersion = conditionalPrompt(useDefaults, "Protocol version", "0.3.0")
		adl.Spec.Card.PreferredTransport = conditionalPrompt(useDefaults, "Preferred transport", "JSONRPC")

		defaultInputModes := conditionalPrompt(useDefaults, "Default input modes (comma-separated)", "text,voice")
		if defaultInputModes != "" {
			modes := strings.Split(defaultInputModes, ",")
			for i, mode := range modes {
				modes[i] = strings.TrimSpace(mode)
			}
			adl.Spec.Card.DefaultInputModes = modes
		}

		defaultOutputModes := conditionalPrompt(useDefaults, "Default output modes (comma-separated)", "text,audio")
		if defaultOutputModes != "" {
			modes := strings.Split(defaultOutputModes, ",")
			for i, mode := range modes {
				modes[i] = strings.TrimSpace(mode)
			}
			adl.Spec.Card.DefaultOutputModes = modes
		}

		scheme := adl.Spec.Server.Scheme
		if scheme == "" {
			scheme = "http"
		}
		cardURL := conditionalPrompt(useDefaults, "Agent service URL", fmt.Sprintf("%s://%s.example.com:%d", scheme, adl.Metadata.Name, adl.Spec.Server.Port))
		adl.Spec.Card.URL = cardURL
	}

	fmt.Println("\n💻 Language Configuration")
	fmt.Println("-------------------------")

	language := promptWithConfig("language", useDefaults, "Programming language", "go")

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

	switch language {
	case "go":
		adl.Spec.Language.Go = &struct {
			Module  string `yaml:"module"`
			Version string `yaml:"version"`
		}{}
		defaultModule := getDefaultGoModule(adl.Metadata.Name)
		adl.Spec.Language.Go.Module = promptWithConfig("go-module", useDefaults, "Go module", defaultModule)
		adl.Spec.Language.Go.Version = promptWithConfig("go-version", useDefaults, "Go version", "1.25")

	case "rust":
		adl.Spec.Language.Rust = &struct {
			PackageName string `yaml:"packageName"`
			Version     string `yaml:"version"`
			Edition     string `yaml:"edition"`
		}{}
		adl.Spec.Language.Rust.PackageName = promptWithConfig("rust-package-name", useDefaults, "Rust package name", adl.Metadata.Name)
		adl.Spec.Language.Rust.Version = promptWithConfig("rust-version", useDefaults, "Rust version", "1.89.0")
		adl.Spec.Language.Rust.Edition = promptWithConfig("rust-edition", useDefaults, "Rust edition", "2024")

	case "typescript":
		adl.Spec.Language.TypeScript = &struct {
			PackageName string `yaml:"packageName"`
			NodeVersion string `yaml:"nodeVersion"`
		}{}
		adl.Spec.Language.TypeScript.PackageName = promptWithConfig("typescript-name", useDefaults, "TypeScript package name", adl.Metadata.Name)
		adl.Spec.Language.TypeScript.NodeVersion = "20"

	default:
		// Default to Go
		adl.Spec.Language.Go = &struct {
			Module  string `yaml:"module"`
			Version string `yaml:"version"`
		}{}
		defaultModule := getDefaultGoModule(adl.Metadata.Name)
		adl.Spec.Language.Go.Module = promptWithConfig("go-module", useDefaults, "Go module", defaultModule)
		adl.Spec.Language.Go.Version = promptWithConfig("go-version", useDefaults, "Go version", "1.25")
	}

	fmt.Println("\n🏗️ Sandbox Configuration")
	fmt.Println("------------------------")

	floxEnabled := promptBoolWithConfig("flox", useDefaults, "Enable Flox environment", false)
	devcontainerEnabled := promptBoolWithConfig("devcontainer", useDefaults, "Enable DevContainer environment", false)

	if floxEnabled || devcontainerEnabled {
		sandboxConfig := &struct {
			Flox *struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"flox,omitempty"`
			DevContainer *struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"devcontainer,omitempty"`
		}{}

		if floxEnabled {
			sandboxConfig.Flox = &struct {
				Enabled bool `yaml:"enabled"`
			}{
				Enabled: true,
			}
		}

		if devcontainerEnabled {
			sandboxConfig.DevContainer = &struct {
				Enabled bool `yaml:"enabled"`
			}{
				Enabled: true,
			}
		}

		adl.Spec.Sandbox = sandboxConfig
	}

	fmt.Println("\n🚀 Deployment Configuration")
	fmt.Println("---------------------------")

	deploymentType := promptWithConfig("deployment", useDefaults, "Deployment type (kubernetes, leave empty for no deployment)", "")
	if deploymentType != "" {
		adl.Spec.Deployment = &struct {
			Type string `yaml:"type,omitempty"`
		}{
			Type: deploymentType,
		}
	}

	fmt.Println("\n📋 Source Control Management")
	fmt.Println("-----------------------------")

	scmProvider := conditionalPrompt(useDefaults, "SCM provider", "github")
	if scmProvider != "" {
		adl.Spec.SCM = &struct {
			Provider       string `yaml:"provider"`
			URL            string `yaml:"url,omitempty"`
			GithubApp      bool   `yaml:"github_app,omitempty"`
			IssueTemplates bool   `yaml:"issue_templates,omitempty"`
		}{
			Provider: scmProvider,
		}

		if scmProvider == "github" {
			owner, repo := parseGitRemote()
			var defaultURL string
			if owner != "" && repo != "" {
				defaultURL = fmt.Sprintf("https://github.com/%s/%s", owner, repo)
			} else {
				defaultURL = fmt.Sprintf("https://github.com/example/%s", adl.Metadata.Name)
			}

			scmURL := conditionalPrompt(useDefaults, "Repository URL", defaultURL)
			adl.Spec.SCM.URL = scmURL
			adl.Spec.SCM.GithubApp = conditionalPromptBool(useDefaults, "Enable GitHub App integration", true)

			if useDefaults {
				adl.Spec.SCM.IssueTemplates = true
				fmt.Printf("Enable issue templates [y/n] [y]: y\n")
			} else {
				adl.Spec.SCM.IssueTemplates = promptBool("Enable issue templates", true)
			}
		}
	}

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

func conditionalPrompt(useDefaults bool, promptText, defaultValue string) string {
	if useDefaults {
		fmt.Printf("%s [%s]: %s\n", promptText, defaultValue, defaultValue)
		return defaultValue
	}
	return promptString(promptText, defaultValue)
}

func promptWithConfig(key string, useDefaults bool, promptText, defaultValue string) string {
	if viper.IsSet(key) && viper.GetString(key) != "" {
		value := viper.GetString(key)
		fmt.Printf("%s [%s]: %s\n", promptText, defaultValue, value)
		return value
	}
	return conditionalPrompt(useDefaults, promptText, defaultValue)
}

func promptBoolWithConfig(key string, useDefaults bool, promptText string, defaultValue bool) bool {
	if viper.IsSet(key) {
		value := viper.GetBool(key)
		valueStr := "n"
		if value {
			valueStr = "y"
		}
		defaultStr := "n"
		if defaultValue {
			defaultStr = "y"
		}
		fmt.Printf("%s [y/n] [%s]: %s\n", promptText, defaultStr, valueStr)
		return value
	}
	return conditionalPromptBool(useDefaults, promptText, defaultValue)
}

func conditionalPromptBool(useDefaults bool, promptText string, defaultValue bool) bool {
	if useDefaults {
		defaultStr := "n"
		if defaultValue {
			defaultStr = "y"
		}
		fmt.Printf("%s [y/n] [%s]: %s\n", promptText, defaultStr, defaultStr)
		return defaultValue
	}
	return promptBool(promptText, defaultValue)
}

func conditionalPromptChoice(useDefaults bool, promptText string, choices []string, defaultValue string) string {
	if useDefaults {
		fmt.Printf("%s (%s) [%s]: %s\n", promptText, strings.Join(choices, "/"), defaultValue, defaultValue)
		return defaultValue
	}
	return promptChoice(promptText, choices, defaultValue)
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

// isValidIdentifier checks if a string is a valid identifier (letters, numbers, underscore, starting with letter/underscore)
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	first := rune(s[0])
	if !unicode.IsLetter(first) && first != '_' {
		return false
	}

	for _, r := range s[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}

	return true
}

// titleCase converts a string to title case (first letter uppercase)
func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
