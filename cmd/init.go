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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/inference-gateway/adl-cli/internal/prompt"
	"github.com/inference-gateway/adl-cli/internal/tui"
)

// aiProviders are the LLM providers the inference-gateway supports, in the order
// shown to users. This mirrors the gateway's Provider enum
// (inference-gateway/inference-gateway openapi.yaml). Note the ADL schema's
// spec.agent.provider enum may lag the gateway; providers not yet in that enum
// will fail `adl validate` until the schema is updated upstream.
var aiProviders = []string{
	"openai",
	"anthropic",
	"google",
	"groq",
	"mistral",
	"deepseek",
	"cohere",
	"cloudflare",
	"moonshot",
	"ollama",
	"ollama_cloud",
}

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
	initCmd.Flags().String("provider", "", "AI provider (openai/anthropic/google/groq/mistral/deepseek/cohere/cloudflare/moonshot/ollama/ollama_cloud; empty = choose at runtime)")
	initCmd.Flags().String("model", "", "AI model (empty = choose at runtime)")
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
	initCmd.Flags().Bool("docker-compose", false, "Enable Docker Compose environment")
	initCmd.Flags().Bool("ai", false, "Enable AI assistant docs (CLAUDE.md/AGENTS.md) and claude-code in sandboxes")
	initCmd.Flags().Bool("ci", false, "Enable CI workflow generation")
	initCmd.Flags().Bool("cd", false, "Enable CD pipeline generation")
	initCmd.Flags().String("deployment", "", "Deployment type (kubernetes, defaults to empty for no deployment)")

	if err := viper.BindPFlags(initCmd.Flags()); err != nil {
		fmt.Printf("Warning: failed to bind flags: %v\n", err)
	}

	viper.SetEnvPrefix("ADL")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

// runInit dispatches to the interactive huh wizard when stdout/stdin are a
// terminal and the user did not pass --defaults; otherwise it runs the plain,
// non-interactive flow. The non-interactive flow is what the test suite exercises
// (it always passes --defaults) and what CI / piped invocations hit, so its
// output is kept byte-for-byte stable.
func runInit(cmd *cobra.Command, args []string) error {
	useDefaults, _ := cmd.Flags().GetBool("defaults")
	if !useDefaults && tui.IsTTY() {
		return runInitInteractive(args)
	}
	return runInitNonInteractive(args, useDefaults)
}

func runInitNonInteractive(args []string, useDefaults bool) error {
	tui.PrintBanner()

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

	tui.Println(tui.Header("ADL Schema Setup"))

	var adl *adlData
	var adlFile string

	useExisting := conditionalPromptBool(useDefaults, "Use an existing ADL schema file", false)

	if useExisting {
		for {
			existingFile := promptString("Path to existing ADL schema file (relative or absolute)", "")
			if existingFile == "" {
				tui.Println(tui.Note("An ADL file path is required. Please provide a path to the existing schema file."))
				continue
			}

			if !filepath.IsAbs(existingFile) {
				cwd, _ := os.Getwd()
				existingFile = filepath.Join(cwd, existingFile)
			}

			if _, err := os.Stat(existingFile); os.IsNotExist(err) {
				tui.Println(tui.Note(fmt.Sprintf("ADL file does not exist: %s", existingFile)))
				continue
			}

			existingADL, err := readADLFile(existingFile)
			if err != nil {
				tui.Println(tui.Note(fmt.Sprintf("Failed to read ADL file: %v", err)))
				continue
			}

			adl = existingADL
			adlFile = filepath.Join(projectDir, "agent.yaml")

			if err := writeADLFile(adl, adlFile); err != nil {
				return fmt.Errorf("failed to write ADL file: %w", err)
			}

			tui.Println(tui.Bullet(fmt.Sprintf("Using existing ADL schema from: %s", existingFile)))
			break
		}
	} else {
		adl = buildADL(collectAnswersNonInteractive(projectName, useDefaults))
		adlFile = filepath.Join(projectDir, "agent.yaml")

		if err := writeADLFile(adl, adlFile); err != nil {
			return fmt.Errorf("failed to write ADL file: %w", err)
		}
	}

	printInitSummary(adl, adlFile, projectDir)

	return nil
}

// vendorBlock mirrors spec.language.<lang>.vendor in the ADL schema. The
// `deps` and `devdeps` keys are intentionally rendered without `omitempty`
// so the scaffolded manifest shows them as empty lists - that's the only
// way first-time users discover where to drop `<package>@<version>`
// entries without consulting the schema. The matching language-specific
// generator (go.mod / Cargo.toml) treats nil and empty equivalently.
type vendorBlock struct {
	Deps    []string `yaml:"deps"`
	Devdeps []string `yaml:"devdeps"`
}

// serviceBlock mirrors a single entry under spec.services in the ADL schema.
// Services are keyed by name (the map key) and the schema requires every entry
// to be an object with type/interface/factory/description - rendering the bare
// service name as a YAML array fails `adl validate`. See issue #190.
type serviceBlock struct {
	Type        string `yaml:"type"`
	Interface   string `yaml:"interface"`
	Factory     string `yaml:"factory"`
	Description string `yaml:"description"`
}

// orchestratorToggle is a single coding-agent on/off switch under
// spec.development.ai.orchestrators.<agent>.
type orchestratorToggle struct {
	Enabled bool `yaml:"enabled"`
}

// orchestratorsBlock mirrors spec.development.ai.orchestrators: each
// coding agent is toggled independently and defaults to disabled.
type orchestratorsBlock struct {
	Claudecode *orchestratorToggle `yaml:"claudecode,omitempty"`
	Codex      *orchestratorToggle `yaml:"codex,omitempty"`
	Gemini     *orchestratorToggle `yaml:"gemini,omitempty"`
	Opencode   *orchestratorToggle `yaml:"opencode,omitempty"`
	Infer      *orchestratorToggle `yaml:"infer,omitempty"`
}

// aiBlock mirrors spec.development.ai: coding-agent orchestrators nested
// under an `orchestrators` key.
type aiBlock struct {
	Orchestrators *orchestratorsBlock `yaml:"orchestrators,omitempty"`
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
		Services map[string]serviceBlock `yaml:"services,omitempty"`
		Tools    []struct {
			ID          string         `yaml:"id"`
			Name        string         `yaml:"name"`
			Description string         `yaml:"description"`
			Tags        []string       `yaml:"tags"`
			Schema      map[string]any `yaml:"schema"`
			Inject      []string       `yaml:"inject,omitempty"`
		} `yaml:"tools,omitempty"`
		Skills []struct {
			ID          string   `yaml:"id"`
			Version     string   `yaml:"version,omitempty"`
			Source      string   `yaml:"source,omitempty"`
			Bare        bool     `yaml:"bare,omitempty"`
			Name        string   `yaml:"name,omitempty"`
			Description string   `yaml:"description,omitempty"`
			Tags        []string `yaml:"tags,omitempty"`
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
				Module  string       `yaml:"module"`
				Version string       `yaml:"version"`
				Vendor  *vendorBlock `yaml:"vendor,omitempty"`
			} `yaml:"go,omitempty"`
			TypeScript *struct {
				PackageName string       `yaml:"packageName"`
				NodeVersion string       `yaml:"nodeVersion"`
				Vendor      *vendorBlock `yaml:"vendor,omitempty"`
			} `yaml:"typescript,omitempty"`
			Rust *struct {
				PackageName string       `yaml:"packageName"`
				Version     string       `yaml:"version"`
				Edition     string       `yaml:"edition"`
				Vendor      *vendorBlock `yaml:"vendor,omitempty"`
			} `yaml:"rust,omitempty"`
		} `yaml:"language,omitempty"`
		SCM *struct {
			Provider       string `yaml:"provider"`
			URL            string `yaml:"url,omitempty"`
			GithubApp      bool   `yaml:"github_app,omitempty"`
			IssueTemplates bool   `yaml:"issue_templates"`
			Dependabot     bool   `yaml:"dependabot"`
			CI             bool   `yaml:"ci"`
			CD             bool   `yaml:"cd"`
		} `yaml:"scm,omitempty"`
		Development *struct {
			Sandbox *struct {
				Flox *struct {
					Enabled bool `yaml:"enabled"`
				} `yaml:"flox,omitempty"`
				DevContainer *struct {
					Enabled bool `yaml:"enabled"`
				} `yaml:"devcontainer,omitempty"`
				DockerCompose *struct {
					Enabled bool `yaml:"enabled"`
				} `yaml:"dockerCompose,omitempty"`
			} `yaml:"sandbox,omitempty"`
			AI   *aiBlock `yaml:"ai,omitempty"`
			Deps []string `yaml:"deps"`
		} `yaml:"development,omitempty"`
		Deployment *struct {
			Type string `yaml:"type,omitempty"`
		} `yaml:"deployment,omitempty"`
		Hooks *struct {
			Post []string `yaml:"post,omitempty"`
		} `yaml:"hooks,omitempty"`
	} `yaml:"spec"`
}

// toolAnswer captures one tool collected during init, before it is expanded into
// the generated JSON schema by buildADL.
type toolAnswer struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	Inject      []string
}

// skillAnswer captures one skill collected during init.
type skillAnswer struct {
	ID          string
	Version     string
	Source      string
	Bare        bool
	Name        string
	Description string
	Tags        []string
}

// answers is the language-agnostic bag of resolved choices produced by either
// collector (interactive wizard or non-interactive prompts/flags). buildADL is
// the single place that turns these answers into the on-disk adlData/agent.yaml,
// so both input paths emit byte-identical manifests for identical answers.
type answers struct {
	Name        string
	Description string
	Version     string

	AgentType    string
	Provider     string
	Model        string
	SystemPrompt string
	MaxTokens    int
	Temperature  float64

	Streaming              bool
	PushNotifications      bool
	StateTransitionHistory bool

	ArtifactsEnabled bool

	Services []string
	Tools    []toolAnswer
	Skills   []skillAnswer

	Port        int
	Scheme      string
	Debug       bool
	AuthEnabled bool

	CardEnabled        bool
	ProtocolVersion    string
	PreferredTransport string
	InputModes         []string
	OutputModes        []string
	CardURL            string

	Language        string
	GoModule        string
	GoVersion       string
	RustPackageName string
	RustVersion     string
	RustEdition     string
	TSPackageName   string

	FloxEnabled          bool
	DevcontainerEnabled  bool
	DockerComposeEnabled bool

	DeploymentType string

	ScmProvider    string
	ScmURL         string
	GithubApp      bool
	IssueTemplates bool
	Dependabot     bool
	CI             bool
	CD             bool

	// Coding-agent orchestrators (spec.development.ai.orchestrators). Only
	// claudecode has a CLI flag (--ai); the rest are wizard/manifest-only.
	Claudecode bool
	Codex      bool
	Gemini     bool
	Opencode   bool
	Infer      bool
}

// defaultToolSchema returns the placeholder JSON schema generated for a freshly
// scaffolded tool. The user fills in the real parameters afterwards.
func defaultToolSchema(name string) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"input": map[string]any{
				"type":        "string",
				"description": "Input parameter for " + name,
			},
		},
		"required": []string{"input"},
	}
}

// splitAndTrim splits a comma-separated list and trims whitespace from each
// element, preserving the original (non-filtering) behaviour of the wizard.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

// servicesFromNames expands the bare service names collected during init into
// the object shape spec.services requires (see serviceBlock). Each name becomes
// a map entry with conventional Go identifiers - <Name>Service for the interface
// and New<Name>Service for the factory - matching what the generator emits in
// internal/<name>/<name>.go. Returns nil when there are no services so the
// `omitempty` tag drops the key entirely for service-less agents.
func servicesFromNames(names []string) map[string]serviceBlock {
	if len(names) == 0 {
		return nil
	}
	services := make(map[string]serviceBlock, len(names))
	for _, name := range names {
		title := titleCase(name)
		services[name] = serviceBlock{
			Type:        "service",
			Interface:   title + "Service",
			Factory:     "New" + title + "Service",
			Description: title + " service",
		}
	}
	return services
}

// buildADL turns resolved answers into the agent.yaml document. It is pure (no
// prompts, no I/O) so it can be golden-tested in isolation and so the
// interactive and non-interactive paths can never drift in manifest shape.
func buildADL(ans answers) *adlData {
	adl := &adlData{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
	}

	adl.Metadata.Name = ans.Name
	adl.Metadata.Description = ans.Description
	adl.Metadata.Version = ans.Version

	if ans.AgentType == "ai-powered" {
		adl.Spec.Agent = &struct {
			Provider     string  `yaml:"provider"`
			Model        string  `yaml:"model"`
			SystemPrompt string  `yaml:"systemPrompt,omitempty"`
			MaxTokens    int     `yaml:"maxTokens,omitempty"`
			Temperature  float64 `yaml:"temperature,omitempty"`
		}{
			Provider:     ans.Provider,
			Model:        ans.Model,
			SystemPrompt: ans.SystemPrompt,
			MaxTokens:    ans.MaxTokens,
			Temperature:  ans.Temperature,
		}
	}

	adl.Spec.Capabilities = &struct {
		Streaming              bool `yaml:"streaming"`
		PushNotifications      bool `yaml:"pushNotifications"`
		StateTransitionHistory bool `yaml:"stateTransitionHistory"`
	}{
		Streaming:              ans.Streaming,
		PushNotifications:      ans.PushNotifications,
		StateTransitionHistory: ans.StateTransitionHistory,
	}

	if ans.ArtifactsEnabled {
		adl.Spec.Artifacts = &struct {
			Enabled bool `yaml:"enabled"`
		}{
			Enabled: true,
		}
	}

	adl.Spec.Services = servicesFromNames(ans.Services)

	for _, t := range ans.Tools {
		adl.Spec.Tools = append(adl.Spec.Tools, struct {
			ID          string         `yaml:"id"`
			Name        string         `yaml:"name"`
			Description string         `yaml:"description"`
			Tags        []string       `yaml:"tags"`
			Schema      map[string]any `yaml:"schema"`
			Inject      []string       `yaml:"inject,omitempty"`
		}{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Tags:        t.Tags,
			Schema:      defaultToolSchema(t.Name),
			Inject:      t.Inject,
		})
	}

	for _, s := range ans.Skills {
		adl.Spec.Skills = append(adl.Spec.Skills, struct {
			ID          string   `yaml:"id"`
			Version     string   `yaml:"version,omitempty"`
			Source      string   `yaml:"source,omitempty"`
			Bare        bool     `yaml:"bare,omitempty"`
			Name        string   `yaml:"name,omitempty"`
			Description string   `yaml:"description,omitempty"`
			Tags        []string `yaml:"tags,omitempty"`
		}{
			ID:          s.ID,
			Version:     s.Version,
			Source:      s.Source,
			Bare:        s.Bare,
			Name:        s.Name,
			Description: s.Description,
			Tags:        s.Tags,
		})
	}

	adl.Spec.Server.Port = ans.Port
	adl.Spec.Server.Scheme = ans.Scheme
	adl.Spec.Server.Debug = ans.Debug
	if ans.AuthEnabled {
		adl.Spec.Server.Auth = &struct {
			Enabled bool `yaml:"enabled"`
		}{
			Enabled: true,
		}
	}

	if ans.CardEnabled {
		adl.Spec.Card = &struct {
			ProtocolVersion    string   `yaml:"protocolVersion,omitempty"`
			URL                string   `yaml:"url,omitempty"`
			PreferredTransport string   `yaml:"preferredTransport,omitempty"`
			DefaultInputModes  []string `yaml:"defaultInputModes,omitempty"`
			DefaultOutputModes []string `yaml:"defaultOutputModes,omitempty"`
			DocumentationURL   string   `yaml:"documentationUrl,omitempty"`
			IconURL            string   `yaml:"iconUrl,omitempty"`
		}{
			ProtocolVersion:    ans.ProtocolVersion,
			PreferredTransport: ans.PreferredTransport,
			DefaultInputModes:  ans.InputModes,
			DefaultOutputModes: ans.OutputModes,
			URL:                ans.CardURL,
		}
	}

	adl.Spec.Language = &struct {
		Go *struct {
			Module  string       `yaml:"module"`
			Version string       `yaml:"version"`
			Vendor  *vendorBlock `yaml:"vendor,omitempty"`
		} `yaml:"go,omitempty"`
		TypeScript *struct {
			PackageName string       `yaml:"packageName"`
			NodeVersion string       `yaml:"nodeVersion"`
			Vendor      *vendorBlock `yaml:"vendor,omitempty"`
		} `yaml:"typescript,omitempty"`
		Rust *struct {
			PackageName string       `yaml:"packageName"`
			Version     string       `yaml:"version"`
			Edition     string       `yaml:"edition"`
			Vendor      *vendorBlock `yaml:"vendor,omitempty"`
		} `yaml:"rust,omitempty"`
	}{}

	switch ans.Language {
	case "rust":
		adl.Spec.Language.Rust = &struct {
			PackageName string       `yaml:"packageName"`
			Version     string       `yaml:"version"`
			Edition     string       `yaml:"edition"`
			Vendor      *vendorBlock `yaml:"vendor,omitempty"`
		}{
			PackageName: ans.RustPackageName,
			Version:     ans.RustVersion,
			Edition:     ans.RustEdition,
			Vendor:      &vendorBlock{Deps: []string{}, Devdeps: []string{}},
		}
	case "typescript":
		adl.Spec.Language.TypeScript = &struct {
			PackageName string       `yaml:"packageName"`
			NodeVersion string       `yaml:"nodeVersion"`
			Vendor      *vendorBlock `yaml:"vendor,omitempty"`
		}{
			PackageName: ans.TSPackageName,
			NodeVersion: "24",
			Vendor:      &vendorBlock{Deps: []string{}, Devdeps: []string{}},
		}
	default:
		adl.Spec.Language.Go = &struct {
			Module  string       `yaml:"module"`
			Version string       `yaml:"version"`
			Vendor  *vendorBlock `yaml:"vendor,omitempty"`
		}{
			Module:  ans.GoModule,
			Version: ans.GoVersion,
			Vendor:  &vendorBlock{Deps: []string{}, Devdeps: []string{}},
		}
	}

	ensureDevelopment(adl)
	adl.Spec.Development.Sandbox = &struct {
		Flox *struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"flox,omitempty"`
		DevContainer *struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"devcontainer,omitempty"`
		DockerCompose *struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"dockerCompose,omitempty"`
	}{
		Flox: &struct {
			Enabled bool `yaml:"enabled"`
		}{
			Enabled: ans.FloxEnabled,
		},
		DevContainer: &struct {
			Enabled bool `yaml:"enabled"`
		}{
			Enabled: ans.DevcontainerEnabled,
		},
		DockerCompose: &struct {
			Enabled bool `yaml:"enabled"`
		}{
			Enabled: ans.DockerComposeEnabled,
		},
	}

	if ans.DeploymentType != "" {
		adl.Spec.Deployment = &struct {
			Type string `yaml:"type,omitempty"`
		}{
			Type: ans.DeploymentType,
		}
	}

	if ans.ScmProvider != "" {
		adl.Spec.SCM = &struct {
			Provider       string `yaml:"provider"`
			URL            string `yaml:"url,omitempty"`
			GithubApp      bool   `yaml:"github_app,omitempty"`
			IssueTemplates bool   `yaml:"issue_templates"`
			Dependabot     bool   `yaml:"dependabot"`
			CI             bool   `yaml:"ci"`
			CD             bool   `yaml:"cd"`
		}{
			Provider: ans.ScmProvider,
		}

		if ans.ScmProvider == "github" {
			adl.Spec.SCM.URL = ans.ScmURL
			adl.Spec.SCM.GithubApp = ans.GithubApp
			adl.Spec.SCM.IssueTemplates = ans.IssueTemplates
			adl.Spec.SCM.Dependabot = ans.Dependabot
		}

		adl.Spec.SCM.CI = ans.CI
		adl.Spec.SCM.CD = ans.CD
	}

	ensureDevelopment(adl)
	adl.Spec.Development.AI = &aiBlock{
		Orchestrators: &orchestratorsBlock{
			Claudecode: &orchestratorToggle{Enabled: ans.Claudecode},
			Codex:      &orchestratorToggle{Enabled: ans.Codex},
			Gemini:     &orchestratorToggle{Enabled: ans.Gemini},
			Opencode:   &orchestratorToggle{Enabled: ans.Opencode},
			Infer:      &orchestratorToggle{Enabled: ans.Infer},
		},
	}

	return adl
}

// collectAnswersNonInteractive reproduces the original linear prompt flow,
// writing into an answers value instead of directly into adlData. With
// --defaults every helper echoes its default and returns without reading stdin,
// so the output is identical to the pre-refactor behaviour the tests pin.
func collectAnswersNonInteractive(projectName string, useDefaults bool) answers {
	var ans answers

	tui.Println(tui.Header("Agent Metadata"))
	ans.Name = promptWithConfig("name", useDefaults, "Agent name", projectName)
	ans.Description = promptWithConfig("description", useDefaults, "Agent description", "A helpful AI agent")
	ans.Version = promptWithConfig("version", useDefaults, "Version", "0.1.0")

	tui.Println(tui.Header("Agent Type"))
	ans.AgentType = conditionalPromptChoice(useDefaults, "Agent type", []string{"ai-powered", "minimal"}, "ai-powered")

	if ans.AgentType == "ai-powered" {
		tui.Println(tui.Header("AI Configuration"))
		tui.Println(tui.Note("Leave the provider and model empty to stay vendor-neutral and select them at runtime via environment variables."))
		ans.Provider = promptChoiceWithConfig("provider", useDefaults, "AI Provider (optional, empty to choose at runtime)", aiProviders, "")
		ans.Model = promptWithConfig("model", useDefaults, "Model (optional, empty to choose at runtime)", "")
		ans.SystemPrompt = conditionalPrompt(useDefaults, "System prompt", "You are a helpful AI assistant.")

		if maxTokensStr := conditionalPrompt(useDefaults, "Max tokens (optional, press enter to skip)", ""); maxTokensStr != "" {
			if maxTokens, err := strconv.Atoi(maxTokensStr); err == nil {
				ans.MaxTokens = maxTokens
			}
		}

		if tempStr := conditionalPrompt(useDefaults, "Temperature (0.0-2.0, optional)", ""); tempStr != "" {
			if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
				ans.Temperature = temp
			}
		}
	}

	tui.Println(tui.Header("Capabilities"))
	ans.Streaming = conditionalPromptBool(useDefaults, "Enable streaming", true)
	ans.PushNotifications = conditionalPromptBool(useDefaults, "Enable push notifications", false)
	ans.StateTransitionHistory = conditionalPromptBool(useDefaults, "Enable state transition history", false)

	tui.Println(tui.Header("Artifacts Configuration"))
	ans.ArtifactsEnabled = conditionalPromptBool(useDefaults, "Enable artifacts support (filesystem/MinIO storage)", false)
	if ans.ArtifactsEnabled {
		tui.Println(tui.Note("Artifacts storage can be configured via A2A_ARTIFACT_* environment variables"))
	}

	tui.Println(tui.Header("Dependencies"))
	addDependencies := conditionalPromptBool(useDefaults, "Add dependencies for dependency injection", useDefaults)

	if addDependencies {
		if useDefaults {
			ans.Services = append(ans.Services, "logger")
			echoRow("Dependency name", "logger")
			tui.Println(tui.Bullet("Added default logger dependency"))
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
				for _, existing := range ans.Services {
					if existing == service {
						fmt.Printf("⚠️  Service '%s' already exists\n", service)
						duplicate = true
						break
					}
				}

				if !duplicate {
					ans.Services = append(ans.Services, service)
					fmt.Printf("✅ Added service: %s\n", service)
					fmt.Printf("💡 You will need to implement this in internal/%s package with a New%sService function\n", service, titleCase(service))
				}

				if !promptBool("Add another service", false) {
					break
				}
			}
		}
	}

	tui.Println(tui.Header("Tools"))
	tui.Println(tui.Note("Tools are function-call entrypoints the agent can invoke (each becomes a generated handler)."))
	addTools := conditionalPromptBool(useDefaults, "Add tools to your agent", false)

	if addTools {
		for {
			var tool toolAnswer
			tool.Name = promptString("Tool name (e.g., 'get_weather')", "")
			if tool.Name == "" {
				break
			}
			tool.ID = tool.Name
			tool.Description = promptString("Tool description", "")

			if tagsStr := promptString("Tool tags (comma-separated, e.g., 'weather,api,data')", ""); tagsStr != "" {
				tool.Tags = splitAndTrim(tagsStr)
			} else {
				tool.Tags = []string{"general"}
			}

			if len(ans.Services) > 0 {
				fmt.Printf("\nAvailable services for tool '%s':\n", tool.Name)
				for i, svc := range ans.Services {
					fmt.Printf("  %d. %s\n", i+1, svc)
				}

				if promptBool("Inject services into this tool", false) {
					for {
						svcChoice := promptString("Enter service name (or empty to finish)", "")
						if svcChoice == "" {
							break
						}

						found := false
						for _, svc := range ans.Services {
							if svc == svcChoice {
								found = true
								break
							}
						}

						if found {
							alreadyAdded := false
							for _, existing := range tool.Inject {
								if existing == svcChoice {
									alreadyAdded = true
									break
								}
							}

							if !alreadyAdded {
								tool.Inject = append(tool.Inject, svcChoice)
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

			ans.Tools = append(ans.Tools, tool)

			if !promptBool("Add another tool", false) {
				break
			}
		}
	}

	tui.Println(tui.Header("Skills"))
	tui.Println(tui.Note("Skills are markdown playbooks (with YAML frontmatter) loaded into the system prompt at startup."))
	tui.Println(tui.Note("You can pull skills from the registry, or scaffold a blank one to author yourself ('bare')."))
	addSkills := conditionalPromptBool(useDefaults, "Add markdown skills to your agent", false)

	if addSkills {
		for {
			var skill skillAnswer
			skill.ID = promptString("Skill id (kebab-case, e.g. 'data-analysis')", "")
			if skill.ID == "" {
				break
			}

			source := promptChoice("Source", []string{"registry", "bare"}, "registry")
			switch source {
			case "bare":
				skill.Bare = true
				defaultName := skill.ID
				skill.Name = promptString("Skill name", defaultName)
				if skill.Name == "" {
					skill.Name = defaultName
				}
				skill.Description = promptString("Skill description", "")
				if tagsStr := promptString("Skill tags (comma-separated, optional)", ""); tagsStr != "" {
					skill.Tags = splitAndTrim(tagsStr)
				}
			default:
				skill.Version = promptString("Pin to version (optional, e.g. '0.1.0')", "")
			}

			ans.Skills = append(ans.Skills, skill)

			if !promptBool("Add another skill", false) {
				break
			}
		}
	}

	tui.Println(tui.Header("Server Configuration"))
	portStr := conditionalPrompt(useDefaults, "Server port", "8080")
	if port, err := strconv.Atoi(portStr); err == nil {
		ans.Port = port
	} else {
		ans.Port = 8080
	}
	ans.Scheme = conditionalPrompt(useDefaults, "Server scheme (http/https)", "http")
	ans.Debug = conditionalPromptBool(useDefaults, "Enable debug mode", false)
	ans.AuthEnabled = conditionalPromptBool(useDefaults, "Enable server authentication", false)

	tui.Println(tui.Header("Agent Card Configuration"))
	ans.CardEnabled = conditionalPromptBool(useDefaults, "Configure agent card (protocol, transport, modes)", false)
	if ans.CardEnabled {
		ans.ProtocolVersion = conditionalPrompt(useDefaults, "Protocol version", "0.3.0")
		ans.PreferredTransport = conditionalPrompt(useDefaults, "Preferred transport", "JSONRPC")

		if modes := conditionalPrompt(useDefaults, "Default input modes (comma-separated)", "text,voice"); modes != "" {
			ans.InputModes = splitAndTrim(modes)
		}

		if modes := conditionalPrompt(useDefaults, "Default output modes (comma-separated)", "text,audio"); modes != "" {
			ans.OutputModes = splitAndTrim(modes)
		}

		scheme := ans.Scheme
		if scheme == "" {
			scheme = "http"
		}
		ans.CardURL = conditionalPrompt(useDefaults, "Agent service URL", fmt.Sprintf("%s://%s.example.com:%d", scheme, ans.Name, ans.Port))
	}

	tui.Println(tui.Header("Language Configuration"))
	ans.Language = promptWithConfig("language", useDefaults, "Programming language", "typescript")

	switch ans.Language {
	case "rust":
		ans.RustPackageName = promptWithConfig("rust-package-name", useDefaults, "Rust package name", ans.Name)
		ans.RustVersion = promptWithConfig("rust-version", useDefaults, "Rust version", "1.89.0")
		ans.RustEdition = promptWithConfig("rust-edition", useDefaults, "Rust edition", "2024")
	case "typescript":
		ans.TSPackageName = promptWithConfig("typescript-name", useDefaults, "TypeScript package name", ans.Name)
	default:
		ans.GoModule = promptWithConfig("go-module", useDefaults, "Go module", getDefaultGoModule(ans.Name))
		ans.GoVersion = promptWithConfig("go-version", useDefaults, "Go version", "1.26.2")
	}

	tui.Println(tui.Header("Sandbox Configuration"))
	ans.FloxEnabled = promptBoolWithConfig("flox", useDefaults, "Enable Flox environment", false)
	ans.DevcontainerEnabled = promptBoolWithConfig("devcontainer", useDefaults, "Enable DevContainer environment", false)
	ans.DockerComposeEnabled = promptBoolWithConfig("docker-compose", useDefaults, "Enable Docker Compose environment", false)

	tui.Println(tui.Header("Deployment Configuration"))
	ans.DeploymentType = promptWithConfig("deployment", useDefaults, "Deployment type (kubernetes, leave empty for no deployment)", "")

	tui.Println(tui.Header("Source Control Management"))
	ans.ScmProvider = conditionalPrompt(useDefaults, "SCM provider", "github")
	if ans.ScmProvider != "" {
		if ans.ScmProvider == "github" {
			owner, repo := parseGitRemote()
			var defaultURL string
			if owner != "" && repo != "" {
				defaultURL = fmt.Sprintf("https://github.com/%s/%s", owner, repo)
			} else {
				defaultURL = fmt.Sprintf("https://github.com/example/%s", ans.Name)
			}

			ans.ScmURL = conditionalPrompt(useDefaults, "Repository URL", defaultURL)
			ans.GithubApp = conditionalPromptBool(useDefaults, "Enable GitHub App integration", true)

			if useDefaults {
				ans.IssueTemplates = false
				echoRow("Enable issue templates", "no")
			} else {
				ans.IssueTemplates = promptBool("Enable issue templates", false)
			}

			if useDefaults {
				ans.Dependabot = false
				echoRow("Enable Dependabot configuration", "no")
			} else {
				ans.Dependabot = promptBool("Enable Dependabot configuration", false)
			}
		}

		ans.CI = promptBoolWithConfig("ci", useDefaults, "Enable CI workflow generation", false)
		ans.CD = promptBoolWithConfig("cd", useDefaults, "Enable CD pipeline generation", false)
	}

	tui.Println(tui.Header("AI Assistant Documentation"))
	ans.Claudecode = promptBoolWithConfig("ai", useDefaults, "Enable Claude Code (CLAUDE.md + claude-code in sandboxes)", false)

	return ans
}

// ensureDevelopment lazily initialises adl.Spec.Development so that callers
// can populate adl.Spec.Development.Sandbox / adl.Spec.Development.AI without
// repeating the nil check at every assignment site.
//
// Deps is seeded as an empty (but non-nil) slice so the YAML encoder emits
// `deps: []` rather than omitting the key - first-time users need that
// breadcrumb to discover where to drop cross-cutting sandbox tools (e.g.
// `kubectl@^1.36.1`, `terraform@^1.15.3`).
func ensureDevelopment(adl *adlData) {
	if adl.Spec.Development != nil {
		if adl.Spec.Development.Deps == nil {
			adl.Spec.Development.Deps = []string{}
		}
		return
	}
	adl.Spec.Development = &struct {
		Sandbox *struct {
			Flox *struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"flox,omitempty"`
			DevContainer *struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"devcontainer,omitempty"`
			DockerCompose *struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"dockerCompose,omitempty"`
		} `yaml:"sandbox,omitempty"`
		AI   *aiBlock `yaml:"ai,omitempty"`
		Deps []string `yaml:"deps"`
	}{
		Deps: []string{},
	}
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

// echoRow prints a resolved field as a styled "label › value" row. It is used
// only on the non-interactive (--defaults / flag-driven) path; the interactive
// readline branches below never call it.
func echoRow(label, value string) {
	tui.Println(tui.Row(label, value))
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func conditionalPrompt(useDefaults bool, promptText, defaultValue string) string {
	if useDefaults {
		echoRow(promptText, defaultValue)
		return defaultValue
	}
	return promptString(promptText, defaultValue)
}

func promptWithConfig(key string, useDefaults bool, promptText, defaultValue string) string {
	if viper.IsSet(key) && viper.GetString(key) != "" {
		value := viper.GetString(key)
		echoRow(promptText, value)
		return value
	}
	return conditionalPrompt(useDefaults, promptText, defaultValue)
}

// promptChoiceWithConfig is the choice-list analogue of promptWithConfig: a
// flag/env value (via viper) wins and is echoed without prompting; otherwise it
// falls back to the interactive choice prompt, or the default in --defaults
// mode. This is what lets `adl init --defaults --provider <p>` honor the flag
// instead of silently returning the empty default (issue #191), while keeping
// the empty default so an unset provider stays vendor-neutral.
func promptChoiceWithConfig(key string, useDefaults bool, promptText string, choices []string, defaultValue string) string {
	if viper.IsSet(key) && viper.GetString(key) != "" {
		value := viper.GetString(key)
		echoRow(promptText, value)
		return value
	}
	return conditionalPromptChoice(useDefaults, promptText, choices, defaultValue)
}

func promptBoolWithConfig(key string, useDefaults bool, promptText string, defaultValue bool) bool {
	if viper.IsSet(key) {
		value := viper.GetBool(key)
		echoRow(promptText, yesNo(value))
		return value
	}
	return conditionalPromptBool(useDefaults, promptText, defaultValue)
}

func conditionalPromptBool(useDefaults bool, promptText string, defaultValue bool) bool {
	if useDefaults {
		echoRow(promptText, yesNo(defaultValue))
		return defaultValue
	}
	return promptBool(promptText, defaultValue)
}

func conditionalPromptChoice(useDefaults bool, promptText string, choices []string, defaultValue string) string {
	if useDefaults {
		echoRow(promptText, defaultValue)
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
