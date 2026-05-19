package schema

import (
	"os"
	"strings"
	"testing"
)

func TestValidator_ValidateFile(t *testing.T) {
	validADL := `apiVersion: adl.inference-gateway.com/v1
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
	validADLWithAgent := `apiVersion: adl.inference-gateway.com/v1
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
			adl: `apiVersion: adl.inference-gateway.com/v1
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
      version: "1.94.1"
      edition: "2024"
      features:
        - redis
`,
			wantErr: false,
		},
		{
			name: "rust language without features validates",
			adl: `apiVersion: adl.inference-gateway.com/v1
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
      version: "1.94.1"
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

func TestValidator_SkillsAndTools(t *testing.T) {
	cases := []struct {
		name    string
		adl     string
		wantErr bool
		errSub  string
	}{
		{
			name: "tool and bare skill both valid with read built-in",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: split-agent
  description: "skills and tools split"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  config:
    tools:
      read:
        enabled: true
  tools:
    - id: read
    - id: ping
      name: ping
      description: "Ping a host"
      tags: [network]
      schema:
        type: object
        properties:
          host:
            type: string
        required: [host]
  skills:
    - id: incident-response
      bare: true
      name: incident-response
      description: "How to triage incidents"
      tags: [ops]
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/split"
      version: "1.26.2"
`,
			wantErr: false,
		},
		{
			name: "skills present but no read tool is rejected",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: no-read
  description: "skills without read"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  agent:
    provider: deepseek
    model: deepseek-v4-flash
  skills:
    - id: x
      bare: true
      name: x
      description: "x"
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/x"
      version: "1.26.2"
`,
			wantErr: true,
			errSub:  "missing '- id: read'",
		},
		{
			name: "skills with read listed but not enabled is rejected",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: read-disabled
  description: "skills with disabled read"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  agent:
    provider: deepseek
    model: deepseek-v4-flash
  tools:
    - id: read
  skills:
    - id: x
      bare: true
      name: x
      description: "x"
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/x"
      version: "1.26.2"
`,
			wantErr: true,
			errSub:  "spec.config.tools.read.enabled",
		},
		{
			name: "skills without agent are allowed even without read",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: no-agent-with-skills
  description: "skills used as documentation, no agent loop"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  skills:
    - id: x
      bare: true
      name: x
      description: "x"
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/x"
      version: "1.26.2"
`,
			wantErr: false,
		},
		{
			name: "reserved tool with user-supplied name is rejected",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: reserved-with-name
  description: "reserved id must be id-only"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  tools:
    - id: bash
      name: MyBash
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/x"
      version: "1.26.2"
`,
			wantErr: true,
			errSub:  "reserved tool 'bash' must not set 'name'",
		},
		{
			name: "reserved tool config with unknown key is rejected",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: typo-config
  description: "unknown config key under bash"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  config:
    tools:
      bash:
        enabled: true
        tymeout_seconds: 30
  tools:
    - id: bash
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/x"
      version: "1.26.2"
`,
			wantErr: true,
			errSub:  "spec.config.tools.bash",
		},
		{
			name: "inject config.tools is rejected",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: inject-reserved
  description: "user tool tries to inject the reserved namespace"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  tools:
    - id: nope
      name: nope
      description: "Tries to grab reserved config"
      tags: [bad]
      inject:
        - config.tools
      schema:
        type: object
        properties:
          x:
            type: string
        required: [x]
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/x"
      version: "1.26.2"
`,
			wantErr: true,
			errSub:  "reserved namespace 'config.tools'",
		},
		{
			name: "reserved tool id-only entry is valid",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: minimal-bash
  description: "minimal reserved entry"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  config:
    tools:
      bash:
        enabled: true
        whitelist: [ls, cat]
        timeout_seconds: 30
  tools:
    - id: bash
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/x"
      version: "1.26.2"
`,
			wantErr: false,
		},
		{
			name: "bare skill missing description is rejected",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: bare-missing
  description: "bare must include description"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  skills:
    - id: incomplete
      bare: true
      name: incomplete
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/incomplete"
      version: "1.26.2"
`,
			wantErr: true,
			errSub:  "missing description",
		},
		{
			name: "tool injecting undefined service is rejected",
			adl: `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: missing-svc
  description: "tool injects unknown service"
  version: "0.1.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  tools:
    - id: ask
      name: ask
      description: "Ask"
      tags: [test]
      inject:
        - mystery
      schema:
        type: object
        properties:
          q:
            type: string
        required: [q]
  server:
    port: 8080
  language:
    go:
      module: "github.com/example/m"
      version: "1.26.2"
`,
			wantErr: true,
			errSub:  "injects service 'mystery'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "adl-skills-tools-*.yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer func() {
				if err := os.Remove(tmpFile.Name()); err != nil {
					t.Logf("Failed to remove temp file: %v", err)
				}
			}()
			if _, err := tmpFile.WriteString(tc.adl); err != nil {
				t.Fatalf("Failed to write: %v", err)
			}
			if err := tmpFile.Close(); err != nil {
				t.Fatalf("Failed to close: %v", err)
			}

			err = NewValidator().ValidateFile(tmpFile.Name())
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected validation error, got nil")
				}
				if tc.errSub != "" && !strings.Contains(err.Error(), tc.errSub) {
					t.Fatalf("expected error containing %q, got %v", tc.errSub, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}
