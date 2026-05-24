package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// TestRegistry_DockerCompose_AllLanguages verifies that
// spec.development.sandbox.dockerCompose.enabled emits a docker-compose.yaml
// for every supported language, not just Rust. This guards the regression
// reported in issue #148 where Go projects silently skipped compose
// generation even though CLAUDE.md advertised the file. TypeScript is not
// covered yet - its template tree is still empty (the language is planned
// but not implemented). When TS templates land, add a "typescript" case
// here.
func TestRegistry_DockerCompose_AllLanguages(t *testing.T) {
	cases := []struct {
		name     string
		language string
		makeADL  func() *schema.ADL
	}{
		{
			name:     "go",
			language: "go",
			makeADL: func() *schema.ADL {
				return &schema.ADL{
					APIVersion: "adl.inference-gateway.com/v1",
					Kind:       "Agent",
					Metadata:   schema.Metadata{Name: "go-agent", Description: "x", Version: "1.0.0"},
					Spec: schema.Spec{
						Capabilities: schema.Capabilities{Streaming: true},
						Server:       schema.Server{Port: 8080},
						Language: schema.Language{
							Go: &schema.GoConfig{Module: "example.com/agent", Version: "1.26.2"},
						},
					},
				}
			},
		},
		{
			name:     "rust",
			language: "rust",
			makeADL:  minimalRustADL,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/disabled", func(t *testing.T) {
			r, err := NewRegistry(tc.language)
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			if _, ok := r.GetFiles(tc.makeADL())["docker-compose.yaml"]; ok {
				t.Fatalf("docker-compose.yaml unexpectedly emitted when sandbox flag unset")
			}
		})

		t.Run(tc.name+"/enabled", func(t *testing.T) {
			r, err := NewRegistry(tc.language)
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			adl := tc.makeADL()
			adl.Spec.Development = &schema.DevelopmentConfig{
				Sandbox: &schema.SandboxConfig{
					DockerCompose: &schema.DockerComposeConfig{Enabled: true},
				},
			}
			files := r.GetFiles(adl)
			tmplKey, ok := files["docker-compose.yaml"]
			if !ok {
				t.Fatalf("docker-compose.yaml missing when sandbox.dockerCompose.enabled=true (language=%s)", tc.language)
			}
			if tmplKey != "docker/docker-compose.yaml" {
				t.Fatalf("docker-compose.yaml mapped to %q, want %q", tmplKey, "docker/docker-compose.yaml")
			}
			if _, err := r.GetTemplate(tmplKey); err != nil {
				t.Fatalf("template %q not loaded: %v", tmplKey, err)
			}
		})
	}
}

// TestDockerComposeTemplate_ContainsRequiredServices verifies that the
// generated docker-compose.yaml is a working local stack with every service
// promised by the bug report: gateway, the agent built from source,
// the infer CLI, and the a2a-debugger.
func TestDockerComposeTemplate_ContainsRequiredServices(t *testing.T) {
	cases := []struct {
		name     string
		language string
		makeADL  func() *schema.ADL
	}{
		{
			name:     "go agent",
			language: "go",
			makeADL: func() *schema.ADL {
				return &schema.ADL{
					APIVersion: "adl.inference-gateway.com/v1",
					Kind:       "Agent",
					Metadata:   schema.Metadata{Name: "shipping-agent", Description: "x", Version: "1.0.0"},
					Spec: schema.Spec{
						Capabilities: schema.Capabilities{Streaming: true},
						Server:       schema.Server{Port: 8443},
						Agent: &schema.Agent{
							Provider:     "openai",
							Model:        "gpt-5-mini",
							SystemPrompt: "hello",
						},
						Language: schema.Language{
							Go: &schema.GoConfig{Module: "example.com/agent", Version: "1.26.2"},
						},
					},
				}
			},
		},
		{
			name:     "rust agent",
			language: "rust",
			makeADL:  minimalRustADL,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			registry, err := NewRegistry(tc.language)
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			engine := NewWithRegistry("minimal", registry)
			out, err := engine.ExecuteTemplate("docker/docker-compose.yaml", Context{ADL: tc.makeADL()})
			if err != nil {
				t.Fatalf("ExecuteTemplate: %v", err)
			}

			wantFragments := []string{
				"image: ghcr.io/inference-gateway/inference-gateway:latest",
				"image: ghcr.io/inference-gateway/cli:latest",
				"image: ghcr.io/inference-gateway/a2a-debugger:latest",
				"build: .",
				"profiles:\n      - cli",
				"profiles:\n      - debugger",
				"gateway:",
				"depends_on:",
				"condition: service_started",
			}
			for _, frag := range wantFragments {
				if !strings.Contains(out, frag) {
					t.Errorf("compose output missing %q\n---\n%s", frag, out)
				}
			}
		})
	}
}

// TestDockerComposeTemplate_ArtifactsWiring verifies that when
// spec.artifacts.enabled is true, the generated compose file pre-wires
// the artifacts server on the agent and the matching infer CLI plumbing
// so users can fetch artifacts end-to-end without manual edits. When the
// flag is unset/false, none of the wiring is emitted and the CLI's web
// fetch tool stays disabled (no behavior change for non-artifact agents).
func TestDockerComposeTemplate_ArtifactsWiring(t *testing.T) {
	cases := []struct {
		name             string
		artifactsEnabled bool
		wantPresent      []string
		wantAbsent       []string
	}{
		{
			name:             "artifacts disabled",
			artifactsEnabled: false,
			wantPresent: []string{
				"INFER_TOOLS_WEB_FETCH_ENABLED: false",
			},
			wantAbsent: []string{
				"A2A_ARTIFACTS_ENABLE",
				"A2A_ARTIFACTS_SERVER_HOST",
				"A2A_ARTIFACTS_SERVER_PORT",
				"INFER_TOOLS_WEB_FETCH_WHITELISTED_DOMAINS",
				"./tmp:/home/infer/.infer/tmp",
			},
		},
		{
			name:             "artifacts enabled",
			artifactsEnabled: true,
			wantPresent: []string{
				"A2A_ARTIFACTS_ENABLE: true",
				"A2A_ARTIFACTS_SERVER_HOST: browser-agent",
				"A2A_ARTIFACTS_SERVER_PORT: 8081",
				"INFER_TOOLS_WEB_FETCH_ENABLED: true",
				"INFER_TOOLS_WEB_FETCH_WHITELISTED_DOMAINS: |\n        - browser-agent",
				"./tmp:/home/infer/.infer/tmp",
			},
			wantAbsent: []string{
				"INFER_TOOLS_WEB_FETCH_ENABLED: false",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			registry, err := NewRegistry("go")
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			engine := NewWithRegistry("minimal", registry)
			adl := &schema.ADL{
				APIVersion: "adl.inference-gateway.com/v1",
				Kind:       "Agent",
				Metadata:   schema.Metadata{Name: "browser-agent", Description: "x", Version: "1.0.0"},
				Spec: schema.Spec{
					Capabilities: schema.Capabilities{Streaming: true},
					Server:       schema.Server{Port: 8080},
					Agent: &schema.Agent{
						Provider:     "openai",
						Model:        "gpt-5-mini",
						SystemPrompt: "hello",
					},
					Language: schema.Language{
						Go: &schema.GoConfig{Module: "example.com/agent", Version: "1.26.2"},
					},
				},
			}
			if tc.artifactsEnabled {
				adl.Spec.Artifacts = &schema.ArtifactsConfig{Enabled: true}
			}

			out, err := engine.ExecuteTemplate("docker/docker-compose.yaml", Context{ADL: adl})
			if err != nil {
				t.Fatalf("ExecuteTemplate: %v", err)
			}

			for _, want := range tc.wantPresent {
				if !strings.Contains(out, want) {
					t.Errorf("compose output missing %q\n---\n%s", want, out)
				}
			}
			for _, notWant := range tc.wantAbsent {
				if strings.Contains(out, notWant) {
					t.Errorf("compose output unexpectedly contains %q\n---\n%s", notWant, out)
				}
			}
		})
	}
}

// TestDockerComposeTemplate_RedisOnlyWithRustFeature confirms that the
// Redis service is added when (and only when) the Rust `redis` Cargo
// feature is enabled - the queue stack stays out of the way for Go and
// TypeScript agents that don't ship the feature.
func TestDockerComposeTemplate_RedisOnlyWithRustFeature(t *testing.T) {
	cases := []struct {
		name      string
		features  []string
		wantRedis bool
	}{
		{name: "no features", features: nil, wantRedis: false},
		{name: "redis feature", features: []string{"redis"}, wantRedis: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			registry, err := NewRegistry("rust")
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			engine := NewWithRegistry("minimal", registry)
			adl := minimalRustADL()
			adl.Spec.Language.Rust.Features = tc.features

			out, err := engine.ExecuteTemplate("docker/docker-compose.yaml", Context{ADL: adl})
			if err != nil {
				t.Fatalf("ExecuteTemplate: %v", err)
			}

			hasRedis := strings.Contains(out, "image: redis:8-alpine")
			if hasRedis != tc.wantRedis {
				t.Fatalf("redis service present=%v, want=%v\n---\n%s", hasRedis, tc.wantRedis, out)
			}

			hasQueueEnv := strings.Contains(out, "A2A_QUEUE_PROVIDER: redis")
			if hasQueueEnv != tc.wantRedis {
				t.Fatalf("A2A_QUEUE_PROVIDER env present=%v, want=%v", hasQueueEnv, tc.wantRedis)
			}
		})
	}
}
