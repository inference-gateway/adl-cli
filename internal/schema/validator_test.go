package schema

import (
	"os"
	"testing"
)

func TestValidator_ValidateFile(t *testing.T) {
	// Create a temporary valid ADL file
	validADL := `apiVersion: a2a.dev/v1
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
      goVersion: "1.21"
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

func TestValidator_ValidateFile_Invalid(t *testing.T) {
	// Create a temporary invalid ADL file
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
