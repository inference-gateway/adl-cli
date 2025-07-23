package devcontainer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/inference-gateway/a2a-cli/internal/schema"
)

func TestGenerator_Generate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "devcontainer-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	adlFile := filepath.Join(tempDir, "test-agent.yaml")
	adlContent := `apiVersion: a2a.dev/v1
kind: Agent
metadata:
  name: test-agent
  description: "Test agent"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: true
    stateTransitionHistory: true
  server:
    port: 8080
  language:
    go:
      module: "github.com/test/test-agent"
      version: "1.23"
`
	if err := os.WriteFile(adlFile, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	generator := New()
	if err := generator.Generate(adlFile, tempDir); err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	devcontainerDir := filepath.Join(tempDir, ".devcontainer")

	files := []string{
		"devcontainer.json",
		"Dockerfile",
	}

	for _, file := range files {
		filePath := filepath.Join(devcontainerDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", file)
		}
	}
}

func TestGenerator_validateADL(t *testing.T) {
	generator := New()

	tests := []struct {
		name        string
		adl         *schema.ADL
		expectError bool
	}{
		{
			name: "valid Go ADL",
			adl: &schema.ADL{
				Spec: schema.Spec{
					Language: &schema.Language{
						Go: &schema.GoConfig{
							Module:  "github.com/test/agent",
							Version: "1.23",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid TypeScript ADL",
			adl: &schema.ADL{
				Spec: schema.Spec{
					Language: &schema.Language{
						TypeScript: &schema.TypeScriptConfig{
							PackageName: "test-agent",
							NodeVersion: "18",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing language",
			adl: &schema.ADL{
				Spec: schema.Spec{
					Language: nil,
				},
			},
			expectError: true,
		},
		{
			name: "no language specified",
			adl: &schema.ADL{
				Spec: schema.Spec{
					Language: &schema.Language{},
				},
			},
			expectError: true,
		},
		{
			name: "multiple languages specified",
			adl: &schema.ADL{
				Spec: schema.Spec{
					Language: &schema.Language{
						Go: &schema.GoConfig{
							Module:  "github.com/test/agent",
							Version: "1.23",
						},
						TypeScript: &schema.TypeScriptConfig{
							PackageName: "test-agent",
							NodeVersion: "18",
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.validateADL(tt.adl)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
