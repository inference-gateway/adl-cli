package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/inference-gateway/adl-cli/internal/schema"
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
			wantContent:  "tools/*",
		},
		{
			name:         "minimal template creates tools ignore for Rust",
			templateName: "minimal",
			adl:          rustADL,
			wantContent:  "src/tools/*",
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
