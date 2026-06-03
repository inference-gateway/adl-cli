package cmd

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// renderADL is a small helper that builds the manifest from answers and encodes
// it the same way writeADLFile does, so golden assertions see the exact bytes
// that land in agent.yaml.
func renderADL(t *testing.T, ans answers) string {
	t.Helper()
	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(buildADL(ans)); err != nil {
		t.Fatalf("encode: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return buf.String()
}

// TestBuildADLGolden locks the manifest shape produced by buildADL for a typical
// Go agent, independent of how the answers were collected (wizard or flags). It
// is the single guard that the empty-list extension points and orchestrator
// toggles keep rendering, so neither input path can silently drop them.
func TestBuildADLGolden(t *testing.T) {
	ans := answers{
		Name:        "weather-agent",
		Description: "Provides weather information",
		Version:     "0.1.0",
		AgentType:   "ai-powered",
		Streaming:   true,
		Port:        8080,
		Scheme:      "http",
		Language:    "go",
		GoModule:    "github.com/example/weather-agent",
		GoVersion:   "1.26.2",
		ScmProvider: "github",
		ScmURL:      "https://github.com/example/weather-agent",
		GithubApp:   true,
	}

	got := renderADL(t, ans)

	wantContains := []string{
		"apiVersion: adl.inference-gateway.com/v1",
		"kind: Agent",
		"name: weather-agent",
		"provider: \"\"",
		"streaming: true",
		"port: 8080",
		"module: github.com/example/weather-agent",
		"vendor:",
		"deps: []",
		"devdeps: []",
		"development:",
		"sandbox:",
		"flox:",
		"devcontainer:",
		"dockerCompose:",
		"deps: []",
		"ai:",
		"orchestrators:",
		"claudecode:",
		"codex:",
		"gemini:",
		"opencode:",
		"infer:",
		"provider: github",
		"github_app: true",
		"issue_templates: false",
		"dependabot: false",
		"ci: false",
		"cd: false",
	}
	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Errorf("manifest missing %q, got:\n%s", want, got)
		}
	}

	// Every orchestrator must default to disabled.
	adl := buildADL(ans)
	if o := adl.Spec.Development.AI.Orchestrators; o.Claudecode.Enabled ||
		o.Codex.Enabled || o.Gemini.Enabled || o.Opencode.Enabled || o.Infer.Enabled {
		t.Errorf("expected all orchestrators disabled by default")
	}
}

// TestBuildADLLanguages asserts that exactly the chosen language block is
// emitted, with its vendor extension points rendered as empty lists.
func TestBuildADLLanguages(t *testing.T) {
	cases := []struct {
		name     string
		ans      answers
		wantKey  string
		wrongKey string
	}{
		{
			name:     "rust",
			ans:      answers{Name: "a", Language: "rust", RustPackageName: "a", RustVersion: "1.89.0", RustEdition: "2024"},
			wantKey:  "rust:",
			wrongKey: "go:",
		},
		{
			name:     "typescript",
			ans:      answers{Name: "a", Language: "typescript", TSPackageName: "a"},
			wantKey:  "typescript:",
			wrongKey: "rust:",
		},
		{
			name:     "go default for unknown",
			ans:      answers{Name: "a", Language: "elixir", GoModule: "github.com/example/a", GoVersion: "1.26.2"},
			wantKey:  "go:",
			wrongKey: "rust:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := renderADL(t, tc.ans)
			if !strings.Contains(got, tc.wantKey) {
				t.Errorf("expected %q in manifest, got:\n%s", tc.wantKey, got)
			}
			if strings.Contains(got, tc.wrongKey) {
				t.Errorf("did not expect %q in manifest, got:\n%s", tc.wrongKey, got)
			}
			if !strings.Contains(got, "deps: []") || !strings.Contains(got, "devdeps: []") {
				t.Errorf("expected empty vendor lists, got:\n%s", got)
			}
		})
	}
}

// TestBuildADLMinimalOmitsAgent verifies that a minimal (non-AI) agent does not
// emit an agent block, and that toggled options round-trip into the document.
func TestBuildADLMinimalOmitsAgent(t *testing.T) {
	ans := answers{
		Name:                 "mini",
		Language:             "go",
		AgentType:            "minimal",
		ArtifactsEnabled:     true,
		AuthEnabled:          true,
		FloxEnabled:          true,
		DockerComposeEnabled: true,
		DeploymentType:       "kubernetes",
		Claudecode:           true,
		Gemini:               true,
		ScmProvider:          "github",
		CI:                   true,
	}

	adl := buildADL(ans)
	if adl.Spec.Agent != nil {
		t.Errorf("minimal agent should not emit spec.agent")
	}
	if adl.Spec.Artifacts == nil || !adl.Spec.Artifacts.Enabled {
		t.Errorf("expected artifacts enabled")
	}
	if adl.Spec.Server.Auth == nil || !adl.Spec.Server.Auth.Enabled {
		t.Errorf("expected auth enabled")
	}
	if adl.Spec.Deployment == nil || adl.Spec.Deployment.Type != "kubernetes" {
		t.Errorf("expected kubernetes deployment")
	}
	if !adl.Spec.Development.Sandbox.Flox.Enabled || !adl.Spec.Development.Sandbox.DockerCompose.Enabled {
		t.Errorf("expected flox + docker compose enabled")
	}
	if !adl.Spec.Development.AI.Orchestrators.Claudecode.Enabled {
		t.Errorf("expected claudecode enabled")
	}
	if !adl.Spec.Development.AI.Orchestrators.Gemini.Enabled {
		t.Errorf("expected gemini enabled")
	}
	if adl.Spec.Development.AI.Orchestrators.Codex.Enabled {
		t.Errorf("expected codex disabled")
	}
	if !adl.Spec.SCM.CI {
		t.Errorf("expected ci true")
	}

	got := renderADL(t, ans)
	if strings.Contains(got, "agent:") {
		t.Errorf("minimal manifest should not contain an agent block, got:\n%s", got)
	}
}

// TestBuildADLToolsAndSkills verifies tool schema generation and skill
// pass-through, including service injection.
func TestBuildADLToolsAndSkills(t *testing.T) {
	ans := answers{
		Name:     "svc-agent",
		Language: "go",
		Services: []string{"logger", "database"},
		Tools: []toolAnswer{
			{ID: "get_weather", Name: "get_weather", Description: "Fetch weather", Tags: []string{"weather"}, Inject: []string{"logger"}},
		},
		Skills: []skillAnswer{
			{ID: "data-analysis", Bare: true, Name: "Data Analysis", Description: "Analyze data", Tags: []string{"data"}},
			{ID: "summarize", Version: "0.1.0"},
		},
	}

	adl := buildADL(ans)
	if len(adl.Spec.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(adl.Spec.Tools))
	}
	tool := adl.Spec.Tools[0]
	if tool.Schema["type"] != "object" {
		t.Errorf("expected generated tool schema with type=object, got %v", tool.Schema["type"])
	}
	if len(tool.Inject) != 1 || tool.Inject[0] != "logger" {
		t.Errorf("expected logger injected, got %v", tool.Inject)
	}

	got := renderADL(t, ans)
	for _, want := range []string{"services:", "- logger", "- database", "id: get_weather", "bare: true", "id: summarize", "version: 0.1.0"} {
		if !strings.Contains(got, want) {
			t.Errorf("manifest missing %q, got:\n%s", want, got)
		}
	}
}
