package schema

import (
	"os"
	"testing"
)

func TestValidator_ValidateFile(t *testing.T) {
	validADL := `apiVersion: adl.dev/v1
kind: Agent
metadata:
  name: test-agent
  description: "Test agent"
  version: "1.0.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: "github.com/example/test-agent"
      version: "1.26.2"
`

	tmpFile, err := os.CreateTemp("", "test-adl-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpFile.WriteString(validADL); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	validator := NewValidator()
	if err := validator.ValidateFile(tmpFile.Name()); err != nil {
		t.Errorf("Validation failed for valid ADL: %v", err)
	}
}

func TestValidator_ValidateFile_AgentWithoutProvider(t *testing.T) {
	validADLWithAgent := `apiVersion: adl.dev/v1
kind: Agent
metadata:
  name: test-agent
  description: "Test agent"
  version: "1.0.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  agent:
    model: "gpt-3.5-turbo"
    systemPrompt: "You are a helpful assistant"
    maxTokens: 1000
    temperature: 0.7
  server:
    port: 8080
    debug: false
  language:
    go:
      module: "github.com/example/test-agent"
      version: "1.26.2"
`

	tmpFile, err := os.CreateTemp("", "test-adl-agent-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpFile.WriteString(validADLWithAgent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	validator := NewValidator()
	if err := validator.ValidateFile(tmpFile.Name()); err != nil {
		t.Errorf("Validation failed for ADL with agent section but no provider: %v", err)
	}
}

func TestValidator_RustFeatures(t *testing.T) {
	cases := []struct {
		name    string
		adl     string
		wantErr bool
	}{
		{
			name: "rust language with redis feature validates",
			adl: `apiVersion: adl.dev/v1
kind: Agent
metadata:
  name: redis-agent
  description: "with redis cargo feature"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
  language:
    rust:
      packageName: "agent"
      version: "1.88"
      edition: "2024"
      features:
        - redis
`,
			wantErr: false,
		},
		{
			name: "rust language without features validates",
			adl: `apiVersion: adl.dev/v1
kind: Agent
metadata:
  name: plain-rust
  description: "no features"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
  language:
    rust:
      packageName: "agent"
      version: "1.88"
      edition: "2024"
`,
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "adl-rust-features-*.yaml")
			if err != nil {
				t.Fatalf("temp file: %v", err)
			}
			defer func() {
				if rmErr := os.Remove(tmpFile.Name()); rmErr != nil {
					t.Logf("cleanup: %v", rmErr)
				}
			}()
			if _, err := tmpFile.WriteString(tc.adl); err != nil {
				t.Fatalf("write: %v", err)
			}
			if err := tmpFile.Close(); err != nil {
				t.Fatalf("close: %v", err)
			}

			err = NewValidator().ValidateFile(tmpFile.Name())
			if tc.wantErr && err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestValidator_ValidateFile_Invalid(t *testing.T) {
	invalidADL := `apiVersion: invalid
kind: Agent
metadata:
  name: test-agent
`

	tmpFile, err := os.CreateTemp("", "test-adl-invalid-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpFile.WriteString(invalidADL); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	validator := NewValidator()
	if err := validator.ValidateFile(tmpFile.Name()); err == nil {
		t.Error("Expected validation to fail for invalid ADL")
	}
}
