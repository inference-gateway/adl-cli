package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/inference-gateway/adl-cli/internal/schema"
	"github.com/inference-gateway/adl-cli/internal/templates"
)

func TestGenerator_Generate(t *testing.T) {
	validADL := &schema.ADL{
		APIVersion: "adl.dev/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-agent",
			Description: "Test agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: &schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: &schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-agent",
					Version: "1.24",
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
				APIVersion: "adl.dev/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: &schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 8080,
					},
					Language: &schema.Language{
						Go: &schema.GoConfig{
							Module:  "github.com/example/test-agent",
							Version: "1.24",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing capabilities",
			adl: &schema.ADL{
				APIVersion: "adl.dev/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Server: schema.Server{
						Port: 8080,
					},
					Language: &schema.Language{
						Go: &schema.GoConfig{
							Module:  "github.com/example/test-agent",
							Version: "1.24",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "spec.capabilities is required",
		},
		{
			name: "missing language",
			adl: &schema.ADL{
				APIVersion: "adl.dev/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: &schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 8080,
					},
				},
			},
			wantErr: true,
			errMsg:  "spec.language is required for code generation",
		},
		{
			name: "missing Go module",
			adl: &schema.ADL{
				APIVersion: "adl.dev/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: &schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 8080,
					},
					Language: &schema.Language{
						Go: &schema.GoConfig{
							Version: "1.24",
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
				APIVersion: "adl.dev/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: &schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 0,
					},
					Language: &schema.Language{
						Go: &schema.GoConfig{
							Module:  "github.com/example/test-agent",
							Version: "1.24",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "spec.server.port is required and must be greater than 0",
		},
		{
			name: "multiple languages specified",
			adl: &schema.ADL{
				APIVersion: "adl.dev/v1",
				Kind:       "Agent",
				Metadata: schema.Metadata{
					Name:        "test-agent",
					Description: "Test agent",
					Version:     "1.0.0",
				},
				Spec: schema.Spec{
					Capabilities: &schema.Capabilities{
						Streaming:              true,
						PushNotifications:      false,
						StateTransitionHistory: false,
					},
					Server: schema.Server{
						Port: 8080,
					},
					Language: &schema.Language{
						Go: &schema.GoConfig{
							Module:  "github.com/example/test-agent",
							Version: "1.24",
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
		APIVersion: "adl.dev/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-agent",
			Description: "Test agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Language: &schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-agent",
					Version: "1.24",
				},
			},
		},
	}

	rustADL := &schema.ADL{
		APIVersion: "adl.dev/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "rust-agent",
			Description: "Rust test agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Language: &schema.Language{
				Rust: &schema.RustConfig{
					PackageName: "rust-agent",
					Version:     "1.70",
					Edition:     "2021",
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
			name:         "minimal template creates tools ignore for Go",
			templateName: "minimal",
			adl:          goADL,
			wantContent:  "skills/*",
		},
		{
			name:         "minimal template creates tools ignore for Rust",
			templateName: "minimal",
			adl:          rustADL,
			wantContent:  "src/skills/*",
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
		APIVersion: "adl.dev/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-cd-agent",
			Description: "Test CD agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: &schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: &schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-cd-agent",
					Version: "1.24",
				},
			},
			SCM: &schema.SCM{
				Provider: "github",
				URL:      "https://github.com/example/test-cd-agent",
			},
		},
	}

	githubAppADL := &schema.ADL{
		APIVersion: "adl.dev/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-github-app-agent",
			Description: "Test GitHub App agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: &schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: &schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-github-app-agent",
					Version: "1.24",
				},
			},
			SCM: &schema.SCM{
				Provider:  "github",
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
			name: "config with all flags",
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
				EnableAI:           true,
			},
			expectedCmd: "adl generate --file my-agent.yaml --output ./output --template custom --overwrite --ci --cd --deployment kubernetes --flox --devcontainer --ai",
		},
		{
			name: "config reproducing the issue scenario",
			config: Config{
				ADLFile:    "agent.yaml",
				OutputDir:  ".",
				Overwrite:  true,
				GenerateCI: true,
				GenerateCD: true,
				EnableFlox: true,
			},
			expectedCmd: "adl generate --file agent.yaml --output . --overwrite --ci --cd --flox",
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
		APIVersion: "adl.dev/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-agent",
			Description: "Test agent",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: &schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: &schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-agent",
					Version: "1.24",
				},
			},
			SCM: &schema.SCM{
				Provider:       "github",
				URL:            "https://github.com/example/test-agent",
				IssueTemplates: true,
			},
		},
	}

	adlWithoutIssueTemplates := &schema.ADL{
		APIVersion: "adl.dev/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "test-agent-no-templates",
			Description: "Test agent without issue templates",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: &schema.Capabilities{
				Streaming:              true,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			Server: schema.Server{
				Port:  8080,
				Debug: false,
			},
			Language: &schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/test-agent-no-templates",
					Version: "1.24",
				},
			},
			SCM: &schema.SCM{
				Provider:       "github",
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
			// Test the registry file mapping logic
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
