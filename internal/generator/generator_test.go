package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inference-gateway/adl-cli/internal/schema"
	"github.com/inference-gateway/adl-cli/internal/templates"
	"github.com/inference-gateway/adl-cli/internal/vendor"
)

func TestGenerator_Generate(t *testing.T) {
	validADL := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-agent",
			Description: "Test agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-agent",
					Version: "1.26.2",
				},
			},
		},
	}

	tmpDir, err := os.MkdirTemp("", "a2a-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	_ = filepath.Join(tmpDir, "agent.yaml")

	gen := New(Config{
		Template:  "minimal",
		Overwrite: true,
		Version:   "test-version",
	})

	outputDir := filepath.Join(tmpDir, "output")
	_ = gen
	_ = outputDir
	_ = validADL
}

func TestGenerator_validateADL(t *testing.T) {
	tests := []struct {
		name    string
		adl     *schema.ADL
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid minimal ADL",
			adl: &schema.ADL{
				APIVersion: "adl.inference-gateway.com/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 8080,
					},
					Language: schema.Language{
						Go: &schema.GoConfig{
							Module:  "github.com/example/test-agent",
							Version: "1.26.2",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing Go module",
			adl: &schema.ADL{
				APIVersion: "adl.inference-gateway.com/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 8080,
					},
					Language: schema.Language{
						Go: &schema.GoConfig{
							Version: "1.26.2",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "spec.language.go.module is required",
		},
		{
			name: "invalid port",
			adl: &schema.ADL{
				APIVersion: "adl.inference-gateway.com/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 0,
					},
					Language: schema.Language{
						Go: &schema.GoConfig{
							Module:  "github.com/example/test-agent",
							Version: "1.26.2",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "spec.server.port is required and must be greater than 0",
		},
		{
			name: "rust with redis cargo feature is accepted",
			adl: &schema.ADL{
				APIVersion: "adl.inference-gateway.com/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 8080,
					},
					Language: schema.Language{
						Rust: &schema.RustConfig{
							PackageName: "rust-agent",
							Version:     "1.94.1",
							Edition:     "2024",
							Features:    []string{"redis"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple languages specified",
			adl: &schema.ADL{
				APIVersion: "adl.inference-gateway.com/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 8080,
					},
					Language: schema.Language{
						Go: &schema.GoConfig{
							Module:  "github.com/example/test-agent",
							Version: "1.26.2",
						},
						TypeScript: &schema.TypeScriptConfig{
							PackageName: "test-agent",
							NodeVersion: "18",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "exactly one programming language must be defined for code generation, found 2",
		},
	}

	gen := New(Config{
		Version: "test-version",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.validateADL(tt.adl)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateADL() expected error, got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateADL() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateADL() unexpected error = %v", err)
			}
		})
	}
}

func TestGenerator_generateADLIgnoreFile(t *testing.T) {
	goADL := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-agent",
			Description: "Test agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Language: schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-agent",
					Version: "1.26.2",
				},
			},
			Tools: []schema.Tool{
				{
					ID:          "search_docs",
					Name:        "search_docs",
					Description: "Search documentation",
					Tags:        []string{"search"},
					Schema: map[string]interface{}{
						"type": "object",
					},
				},
			},
		},
	}

	rustADL := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "rust-agent",
			Description: "Rust test agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Agent: &schema.Agent{
				Provider: schema.AgentProviderDeepseek,
				Model:    "deepseek-v4-flash",
			},
			Language: schema.Language{
				Rust: &schema.RustConfig{
					PackageName: "rust-agent",
					Version:     "1.70",
					Edition:     "2021",
				},
			},
			Tools: []schema.Tool{
				{
					ID:          "process_data",
					Name:        "process_data",
					Description: "Process data",
					Tags:        []string{"processing"},
					Schema: map[string]interface{}{
						"type": "object",
					},
				},
			},
		},
	}

	tests := []struct {
		name         string
		templateName string
		adl          *schema.ADL
		wantContent  string
	}{
		{
			name:         "minimal template creates specific tool file for Go",
			templateName: "minimal",
			adl:          goADL,
			wantContent:  "tools/search_docs.go",
		},
		{
			name:         "minimal template creates specific tool file for Rust",
			templateName: "minimal",
			adl:          rustADL,
			wantContent:  "src/tools/process_data.rs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "adl-ignore-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Failed to remove temp dir: %v", err)
				}
			}()

			gen := New(Config{
				Template: tt.templateName,
			})

			err = gen.generateADLIgnoreFile(tmpDir, tt.templateName, tt.adl)
			if err != nil {
				t.Fatalf("generateADLIgnoreFile() error = %v", err)
			}

			ignoreFilePath := filepath.Join(tmpDir, ".adl-ignore")
			content, err := os.ReadFile(ignoreFilePath)
			if err != nil {
				t.Fatalf("Failed to read .adl-ignore file: %v", err)
			}

			contentStr := string(content)
			if !containsPattern(contentStr, tt.wantContent) {
				t.Errorf("generateADLIgnoreFile() content does not contain expected pattern %q", tt.wantContent)
			}
		})
	}
}

// containsPattern checks if the content contains the expected pattern
func containsPattern(content, pattern string) bool {
	return len(content) > 0 && (content == pattern ||
		(len(content) > len(pattern) &&
			(content[:len(pattern)] == pattern ||
				content[len(content)-len(pattern):] == pattern ||
				containsSubstring(content, pattern))))
}

// containsSubstring is a simple substring check
func containsSubstring(content, pattern string) bool {
	for i := 0; i <= len(content)-len(pattern); i++ {
		if content[i:i+len(pattern)] == pattern {
			return true
		}
	}
	return false
}

func TestGenerator_generateCD(t *testing.T) {
	validADL := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-cd-agent",
			Description: "Test CD agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-cd-agent",
					Version: "1.26.2",
				},
			},
			SCM: &schema.SCM{
				Provider: schema.SCMProviderGithub,
				URL:      "https://github.com/example/test-cd-agent",
			},
		},
	}

	githubAppADL := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-github-app-agent",
			Description: "Test GitHub App agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-github-app-agent",
					Version: "1.26.2",
				},
			},
			SCM: &schema.SCM{
				Provider:  schema.SCMProviderGithub,
				URL:       "https://github.com/example/test-github-app-agent",
				GithubApp: true,
			},
		},
	}

	tests := []struct {
		name               string
		adl                *schema.ADL
		expectGithubAppCD  bool
		expectedTokenUsage string
	}{
		{
			name:               "regular CD workflow",
			adl:                validADL,
			expectGithubAppCD:  false,
			expectedTokenUsage: "secrets.GITHUB_TOKEN",
		},
		{
			name:               "GitHub App CD workflow",
			adl:                githubAppADL,
			expectGithubAppCD:  true,
			expectedTokenUsage: "steps.app-token.outputs.token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "adl-cd-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Failed to remove temp dir: %v", err)
				}
			}()

			gen := New(Config{
				Template:   "minimal",
				Overwrite:  true,
				Version:    "test-version",
				GenerateCD: true,
			})

			ignoreChecker, err := NewIgnoreChecker(tmpDir)
			if err != nil {
				t.Fatalf("Failed to create ignore checker: %v", err)
			}

			err = gen.generateCD(tt.adl, tmpDir, ignoreChecker)
			if err != nil {
				t.Fatalf("generateCD() error = %v", err)
			}

			releasercPath := filepath.Join(tmpDir, ".releaserc.yaml")
			if _, err := os.Stat(releasercPath); os.IsNotExist(err) {
				t.Errorf("expected .releaserc.yaml to be created")
			}

			cdWorkflowPath := filepath.Join(tmpDir, ".github/workflows/cd.yml")
			if _, err := os.Stat(cdWorkflowPath); os.IsNotExist(err) {
				t.Errorf("expected .github/workflows/cd.yml to be created")
			}

			releasercContent, err := os.ReadFile(releasercPath)
			if err != nil {
				t.Fatalf("failed to read .releaserc.yaml: %v", err)
			}
			if !containsSubstring(string(releasercContent), tt.adl.Spec.SCM.URL) {
				t.Errorf("expected .releaserc.yaml to contain repository URL")
			}
			if !containsSubstring(string(releasercContent), "@semantic-release/github") {
				t.Errorf("expected .releaserc.yaml to contain semantic-release plugins")
			}

			cdContent, err := os.ReadFile(cdWorkflowPath)
			if err != nil {
				t.Fatalf("failed to read CD workflow: %v", err)
			}
			if !containsSubstring(string(cdContent), "workflow_dispatch") {
				t.Errorf("expected CD workflow to contain workflow_dispatch trigger")
			}
			if !containsSubstring(string(cdContent), "ghcr.io") {
				t.Errorf("expected CD workflow to contain GitHub Container Registry")
			}
			if !containsSubstring(string(cdContent), "semantic-release") {
				t.Errorf("expected CD workflow to contain semantic-release")
			}
			if !containsSubstring(string(cdContent), tt.expectedTokenUsage) {
				t.Errorf("expected CD workflow to contain token usage: %s", tt.expectedTokenUsage)
			}

			if tt.expectGithubAppCD {
				if !containsSubstring(string(cdContent), "actions/create-github-app-token") {
					t.Errorf("expected GitHub App CD workflow to contain github-app-token action")
				}
				if !containsSubstring(string(cdContent), "BOT_GH_APP_ID") {
					t.Errorf("expected GitHub App CD workflow to contain BOT_GH_APP_ID secret")
				}
				if !containsSubstring(string(cdContent), "BOT_GH_APP_PRIVATE_KEY") {
					t.Errorf("expected GitHub App CD workflow to contain BOT_GH_APP_PRIVATE_KEY secret")
				}
				if !containsSubstring(string(cdContent), "Get GitHub App User ID") {
					t.Errorf("expected GitHub App CD workflow to contain user ID step")
				}
			} else {
				if containsSubstring(string(cdContent), "actions/create-github-app-token") {
					t.Errorf("expected regular CD workflow not to contain github-app-token action")
				}
			}
		})
	}
}

func TestGenerator_Dependabot(t *testing.T) {
	makeADL := func(name string, dependabot bool, lang schema.Language, sandbox *schema.SandboxConfig) *schema.ADL {
		spec := schema.Spec{
			Capabilities: schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port: 8080,
			},
			Language: lang,
			SCM: &schema.SCM{
				Provider:   schema.SCMProviderGithub,
				URL:        "https://github.com/example/" + name,
				Dependabot: dependabot,
			},
		}
		if sandbox != nil {
			spec.Development = &schema.DevelopmentConfig{Sandbox: sandbox}
		}
		return &schema.ADL{
			APIVersion: "adl.inference-gateway.com/v1",
			Kind:       "Agent",
			Metadata: schema.Metadata{
				Name:        name,
				Description: "Test agent",
				Version:     "1.0.0",
			},
			Spec: spec,
		}
	}

	goLang := schema.Language{
		Go: &schema.GoConfig{
			Module:  "github.com/example/test",
			Version: "1.26.2",
		},
	}
	rustLang := schema.Language{
		Rust: &schema.RustConfig{
			PackageName: "test",
			Version:     "1.94.1",
			Edition:     "2024",
		},
	}

	tests := []struct {
		name             string
		adl              *schema.ADL
		registryLang     string
		expectDependabot bool
		mustContain      []string
		mustNotContain   []string
	}{
		{
			name:             "dependabot disabled: no file mapped",
			adl:              makeADL("agent-off", false, goLang, nil),
			registryLang:     "go",
			expectDependabot: false,
		},
		{
			name:             "go agent with dependabot: gomod ecosystem present",
			adl:              makeADL("go-agent", true, goLang, nil),
			registryLang:     "go",
			expectDependabot: true,
			mustContain: []string{
				"package-ecosystem: gomod",
				"package-ecosystem: github-actions",
				"package-ecosystem: docker",
				"ignore:",
				"dependency-name: golang",
				`">1.26.2"`,
				"dependency-name: ubuntu",
				`">24.04"`,
			},
			mustNotContain: []string{"package-ecosystem: cargo", "package-ecosystem: npm", "package-ecosystem: devcontainers"},
		},
		{
			name:             "rust agent with dependabot: cargo ecosystem present",
			adl:              makeADL("rust-agent", true, rustLang, nil),
			registryLang:     "rust",
			expectDependabot: true,
			mustContain:      []string{"package-ecosystem: cargo", "package-ecosystem: github-actions", "package-ecosystem: docker"},
			mustNotContain:   []string{"package-ecosystem: gomod", "package-ecosystem: npm", "dependency-name: golang", "dependency-name: ubuntu"},
		},
		{
			name: "devcontainer enabled: devcontainers ecosystem included",
			adl: makeADL("dev-agent", true, goLang, &schema.SandboxConfig{
				DevContainer: &schema.DevContainerConfig{Enabled: true},
			}),
			registryLang:     "go",
			expectDependabot: true,
			mustContain:      []string{"package-ecosystem: devcontainers", "dependency-name: golang"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := templates.NewRegistry(tt.registryLang)
			if err != nil {
				t.Fatalf("failed to create template registry: %v", err)
			}

			files := registry.GetFiles(tt.adl)

			_, found := files[".github/dependabot.yml"]
			if tt.expectDependabot != found {
				t.Fatalf("dependabot file mapping: expected %v, got %v (files=%v)", tt.expectDependabot, found, files)
			}

			if !tt.expectDependabot {
				return
			}

			engine := templates.NewWithRegistry("", registry)
			content, err := engine.ExecuteTemplate("github/dependabot.yaml", templates.Context{ADL: tt.adl})
			if err != nil {
				t.Fatalf("failed to execute dependabot template: %v", err)
			}

			for _, expected := range tt.mustContain {
				if !containsSubstring(content, expected) {
					t.Errorf("expected dependabot output to contain %q, got:\n%s", expected, content)
				}
			}
			for _, unexpected := range tt.mustNotContain {
				if containsSubstring(content, unexpected) {
					t.Errorf("expected dependabot output NOT to contain %q, got:\n%s", unexpected, content)
				}
			}

			gitattrContent, err := engine.ExecuteTemplate("config/gitattributes", templates.Context{ADL: tt.adl, Language: tt.registryLang})
			if err != nil {
				t.Fatalf("failed to execute gitattributes template: %v", err)
			}
			if !containsSubstring(gitattrContent, ".github/dependabot.yml linguist-generated=true") {
				t.Errorf("expected .gitattributes to mark dependabot.yml as linguist-generated when enabled, got:\n%s", gitattrContent)
			}
		})
	}

	t.Run("gitattributes omits dependabot entry when disabled", func(t *testing.T) {
		adl := makeADL("agent-off", false, goLang, nil)
		registry, err := templates.NewRegistry("go")
		if err != nil {
			t.Fatalf("failed to create template registry: %v", err)
		}
		engine := templates.NewWithRegistry("", registry)
		gitattrContent, err := engine.ExecuteTemplate("config/gitattributes", templates.Context{ADL: adl, Language: "go"})
		if err != nil {
			t.Fatalf("failed to execute gitattributes template: %v", err)
		}
		if containsSubstring(gitattrContent, ".github/dependabot.yml") {
			t.Errorf("expected .gitattributes NOT to reference dependabot.yml when disabled, got:\n%s", gitattrContent)
		}
	})
}

func TestGenerator_buildGenerateCommand(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectedCmd string
	}{
		{
			name: "minimal config",
			config: Config{
				ADLFile:   "agent.yaml",
				OutputDir: ".",
			},
			expectedCmd: "adl generate --file agent.yaml --output .",
		},
		{
			name: "config with all flags rewrites paths to in-project canonical values",
			config: Config{
				ADLFile:            "my-agent.yaml",
				OutputDir:          "./output",
				Template:           "custom",
				Overwrite:          true,
				GenerateCI:         true,
				GenerateCD:         true,
				DeploymentType:     "kubernetes",
				EnableFlox:         true,
				EnableDevContainer: true,
			},
			expectedCmd: "adl generate --file agent.yaml --output . --template custom --overwrite",
		},
		{
			name: "config reproducing the issue scenario (#146): legacy CI/CD/sandbox flags must not leak into the Taskfile",
			config: Config{
				ADLFile:    "agent.yaml",
				OutputDir:  ".",
				Overwrite:  true,
				GenerateCI: true,
				GenerateCD: true,
				EnableFlox: true,
			},
			expectedCmd: "adl generate --file agent.yaml --output . --overwrite",
		},
		{
			name: "config with default template (should not include template flag)",
			config: Config{
				ADLFile:   "agent.yaml",
				OutputDir: ".",
				Template:  "minimal",
			},
			expectedCmd: "adl generate --file agent.yaml --output .",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New(tt.config)
			cmd := g.buildGenerateCommand()
			if cmd != tt.expectedCmd {
				t.Errorf("buildGenerateCommand() = %q, expected %q", cmd, tt.expectedCmd)
			}
		})
	}
}

func TestGenerator_IssueTemplates(t *testing.T) {
	adlWithIssueTemplates := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-agent",
			Description: "Test agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-agent",
					Version: "1.26.2",
				},
			},
			SCM: &schema.SCM{
				Provider:       schema.SCMProviderGithub,
				URL:            "https://github.com/example/test-agent",
				IssueTemplates: true,
			},
		},
	}

	adlWithoutIssueTemplates := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-agent-no-templates",
			Description: "Test agent without issue templates",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-agent-no-templates",
					Version: "1.26.2",
				},
			},
			SCM: &schema.SCM{
				Provider:       schema.SCMProviderGithub,
				URL:            "https://github.com/example/test-agent-no-templates",
				IssueTemplates: false,
			},
		},
	}

	tests := []struct {
		name            string
		adl             *schema.ADL
		expectTemplates bool
		expectedFiles   []string
	}{
		{
			name:            "with issue templates enabled",
			adl:             adlWithIssueTemplates,
			expectTemplates: true,
			expectedFiles: []string{
				".github/ISSUE_TEMPLATE/bug_report.md",
				".github/ISSUE_TEMPLATE/feature_request.md",
				".github/ISSUE_TEMPLATE/refactor_request.md",
			},
		},
		{
			name:            "with issue templates disabled",
			adl:             adlWithoutIssueTemplates,
			expectTemplates: false,
			expectedFiles:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := templates.NewRegistry("go")
			if err != nil {
				t.Fatalf("Failed to create template registry: %v", err)
			}

			files := registry.GetFiles(tt.adl)

			for _, expectedFile := range tt.expectedFiles {
				if _, found := files[expectedFile]; !found {
					t.Errorf("Expected file %s not found in generated files when issue templates are enabled", expectedFile)
				}
			}

			if !tt.expectTemplates {
				for _, templateFile := range []string{
					".github/ISSUE_TEMPLATE/bug_report.md",
					".github/ISSUE_TEMPLATE/feature_request.md",
					".github/ISSUE_TEMPLATE/refactor_request.md",
				} {
					if _, found := files[templateFile]; found {
						t.Errorf("Unexpected file %s found in generated files when issue templates are disabled", templateFile)
					}
				}
			}
		})
	}
}

// TestGenerator_VendorWiring exercises the go.mod / Cargo.toml templates
// end-to-end with vendor.deps / vendor.devdeps populated. The cases cover
// every state from acceptance criteria #6 in issue #152: empty/missing
// vendor, runtime-only, dev-only, both, dedup against built-ins, and
// dedup against a runtime entry (Rust's [dependencies] / [dev-dependencies]
// dual-section rule).
func TestGenerator_VendorWiring(t *testing.T) {
	makeGo := func(v *schema.VendorConfig) *schema.ADL {
		return &schema.ADL{
			APIVersion: "adl.inference-gateway.com/v1",
			Kind:       "Agent",
			Metadata:   schema.Metadata{Name: "agent", Description: "x", Version: "1.0.0"},
			Spec: schema.Spec{
				Capabilities: schema.Capabilities{},
				Server:       schema.Server{Port: 8080},
				Language: schema.Language{
					Go: &schema.GoConfig{
						Module:  "github.com/example/agent",
						Version: "1.26.2",
						Vendor:  v,
					},
				},
			},
		}
	}
	makeRust := func(v *schema.VendorConfig) *schema.ADL {
		return &schema.ADL{
			APIVersion: "adl.inference-gateway.com/v1",
			Kind:       "Agent",
			Metadata:   schema.Metadata{Name: "agent", Description: "x", Version: "1.0.0"},
			Spec: schema.Spec{
				Capabilities: schema.Capabilities{},
				Server:       schema.Server{Port: 8080},
				Language: schema.Language{
					Rust: &schema.RustConfig{
						PackageName: "agent",
						Version:     "1.94.1",
						Edition:     "2024",
						Vendor:      v,
					},
				},
			},
		}
	}

	render := func(t *testing.T, lang, tmplKey string, adl *schema.ADL) string {
		t.Helper()
		registry, err := templates.NewRegistry(lang)
		if err != nil {
			t.Fatalf("NewRegistry: %v", err)
		}
		engine := templates.NewWithRegistry("", registry)
		view, err := vendor.ResolveADL(adl)
		if err != nil {
			t.Fatalf("vendor.ResolveADL: %v", err)
		}
		out, err := engine.ExecuteTemplate(tmplKey, templates.Context{
			ADL:      adl,
			Language: lang,
			Vendor:   view,
		})
		if err != nil {
			t.Fatalf("ExecuteTemplate %s: %v", tmplKey, err)
		}
		return out
	}

	t.Run("go: no vendor block renders unchanged require list and no tool directive", func(t *testing.T) {
		got := render(t, "go", "go.mod", makeGo(nil))
		if !strings.Contains(got, "github.com/inference-gateway/adk v0.18.4") {
			t.Fatalf("expected built-in ADK in require, got:\n%s", got)
		}
		if strings.Contains(got, "stretchr/testify") {
			t.Fatalf("unexpected vendor entry leaked, got:\n%s", got)
		}
		if strings.Contains(got, "tool (") {
			t.Fatalf("expected no tool directive when vendor is empty, got:\n%s", got)
		}
	})

	t.Run("go: deps land in require sorted+deduped; devdeps populate require // indirect and tool block", func(t *testing.T) {
		got := render(t, "go", "go.mod", makeGo(&schema.VendorConfig{
			Deps:    []string{"github.com/google/uuid@v1.6.0", "github.com/google/uuid@v1.5.0"},
			Devdeps: []string{"golang.org/x/tools/cmd/stringer@v0.20.0"},
		}))
		if !strings.Contains(got, "github.com/google/uuid v1.6.0") {
			t.Fatalf("expected uuid v1.6.0 in require, got:\n%s", got)
		}
		if strings.Contains(got, "v1.5.0") {
			t.Fatalf("expected duplicate uuid v1.5.0 to be deduped (first-wins), got:\n%s", got)
		}
		if !strings.Contains(got, "golang.org/x/tools/cmd/stringer v0.20.0 // indirect") {
			t.Fatalf("expected stringer in require as // indirect, got:\n%s", got)
		}
		toolIdx := strings.Index(got, "tool (")
		if toolIdx == -1 {
			t.Fatalf("expected tool directive, got:\n%s", got)
		}
		toolSection := got[toolIdx:]
		if !strings.Contains(toolSection, "golang.org/x/tools/cmd/stringer") {
			t.Fatalf("expected stringer in tool block, got:\n%s", toolSection)
		}
		if strings.Contains(toolSection, "v0.20.0") {
			t.Fatalf("tool block must list bare package paths (no version), got:\n%s", toolSection)
		}
	})

	t.Run("go: dev-only vendor still emits a tool directive without deps", func(t *testing.T) {
		got := render(t, "go", "go.mod", makeGo(&schema.VendorConfig{
			Devdeps: []string{"github.com/golang/mock/mockgen@v1.6.0"},
		}))
		if !strings.Contains(got, "github.com/golang/mock/mockgen v1.6.0 // indirect") {
			t.Fatalf("expected mockgen as // indirect require, got:\n%s", got)
		}
		if !strings.Contains(got, "tool (") {
			t.Fatalf("expected tool directive when only devdeps set, got:\n%s", got)
		}
	})

	t.Run("go: built-in conflict is dropped before the template renders", func(t *testing.T) {
		got := render(t, "go", "go.mod", makeGo(&schema.VendorConfig{
			Deps: []string{"github.com/inference-gateway/adk@v0.0.1"},
		}))
		if !strings.Contains(got, "github.com/inference-gateway/adk v0.18.4") {
			t.Fatalf("expected built-in version preserved, got:\n%s", got)
		}
		if strings.Contains(got, "v0.0.1") {
			t.Fatalf("expected conflicting vendor entry to be dropped, got:\n%s", got)
		}
	})

	t.Run("rust: deps go to [dependencies], devdeps go to [dev-dependencies]", func(t *testing.T) {
		got := render(t, "rust", "Cargo.toml", makeRust(&schema.VendorConfig{
			Deps:    []string{"regex@1.10.0"},
			Devdeps: []string{"mockall@0.12.1", "pretty_assertions@1.4.0"},
		}))
		depsSection := got[strings.Index(got, "[dependencies]"):strings.Index(got, "[dev-dependencies]")]
		devSection := got[strings.Index(got, "[dev-dependencies]"):]
		if !strings.Contains(depsSection, `regex = "1.10.0"`) {
			t.Fatalf("expected regex in [dependencies], got:\n%s", depsSection)
		}
		if !strings.Contains(devSection, `mockall = "0.12.1"`) || !strings.Contains(devSection, `pretty_assertions = "1.4.0"`) {
			t.Fatalf("expected mockall + pretty_assertions in [dev-dependencies], got:\n%s", devSection)
		}
		if strings.Index(devSection, "mockall") > strings.Index(devSection, "pretty_assertions") {
			t.Fatalf("expected dev-dependencies sorted, got:\n%s", devSection)
		}
	})

	t.Run("rust: dev-only vendor still emits the dev-dependencies section even without built-in tools", func(t *testing.T) {
		got := render(t, "rust", "Cargo.toml", makeRust(&schema.VendorConfig{
			Devdeps: []string{"mockall@0.12.1"},
		}))
		if !strings.Contains(got, "[dev-dependencies]") {
			t.Fatalf("expected [dev-dependencies] section, got:\n%s", got)
		}
		if !strings.Contains(got, `mockall = "0.12.1"`) {
			t.Fatalf("expected mockall in [dev-dependencies], got:\n%s", got)
		}
		if strings.Contains(got, "tempfile") {
			t.Fatalf("expected no tempfile when no built-in tools enabled, got:\n%s", got)
		}
	})

	t.Run("rust: built-in runtime crate is rejected from devdeps too (dual-section guard)", func(t *testing.T) {
		got := render(t, "rust", "Cargo.toml", makeRust(&schema.VendorConfig{
			Devdeps: []string{"tokio@0.1.0"},
		}))
		if strings.Contains(got, `tokio = "0.1.0"`) {
			t.Fatalf("expected conflicting tokio vendor entry to be dropped, got:\n%s", got)
		}
		if !strings.Contains(got, `tokio = { version = "1"`) {
			t.Fatalf("expected built-in tokio preserved, got:\n%s", got)
		}
	})

	t.Run("rust: empty vendor block + no built-in tools omits [dev-dependencies]", func(t *testing.T) {
		got := render(t, "rust", "Cargo.toml", makeRust(nil))
		if strings.Contains(got, "[dev-dependencies]") {
			t.Fatalf("expected no [dev-dependencies] section, got:\n%s", got)
		}
	})
}
