package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// TestRegistry_Dockerignore_AllLanguages verifies that every supported
// language generates a .dockerignore alongside its Dockerfile, unconditionally.
// The file pairs with the Dockerfile (always generated) - it is NOT gated on
// the docker-compose sandbox toggle.
func TestRegistry_Dockerignore_AllLanguages(t *testing.T) {
	cases := []struct {
		name     string
		language string
		makeADL  func() *schema.ADL
	}{
		{name: "go", language: "go", makeADL: minimalGoADL},
		{name: "rust", language: "rust", makeADL: minimalRustADL},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRegistry(tc.language)
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			files := r.GetFiles(tc.makeADL())
			tmplKey, ok := files[".dockerignore"]
			if !ok {
				t.Fatalf(".dockerignore missing from generated files for %s", tc.language)
			}
			if tmplKey != "config/dockerignore" {
				t.Fatalf(".dockerignore mapped to %q, want %q", tmplKey, "config/dockerignore")
			}
			if _, err := r.GetTemplate(tmplKey); err != nil {
				t.Fatalf("template %q not loaded: %v", tmplKey, err)
			}
		})
	}
}

// TestDockerignoreTemplate_ContainsExpectedExclusions checks the rendered
// .dockerignore body covers the high-value categories - sensitive .env files,
// the compose stack itself, common build outputs, and VCS metadata. These
// are the items most likely to leak into image layers or break cache hits.
func TestDockerignoreTemplate_ContainsExpectedExclusions(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)
	out, err := engine.ExecuteTemplate("config/dockerignore", Context{ADL: minimalGoADL()})
	if err != nil {
		t.Fatalf("ExecuteTemplate: %v", err)
	}

	wantLines := []string{
		".git",
		".env",
		"!.env.example",
		"docker-compose.yaml",
		"Dockerfile",
		".dockerignore",
		"bin/",
		"target/",
		"node_modules/",
		"README.md",
	}
	for _, line := range wantLines {
		if !strings.Contains(out, line) {
			t.Errorf("dockerignore missing %q\n---\n%s", line, out)
		}
	}
}

// TestRegistry_EnvExample_GatedOnDockerCompose confirms the .env.example
// file is emitted only when the dockerCompose sandbox is enabled, for every
// supported language. The template is shared (common/config/env.example).
func TestRegistry_EnvExample_GatedOnDockerCompose(t *testing.T) {
	cases := []struct {
		name     string
		language string
		makeADL  func() *schema.ADL
	}{
		{name: "go", language: "go", makeADL: minimalGoADL},
		{name: "rust", language: "rust", makeADL: minimalRustADL},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/disabled", func(t *testing.T) {
			r, err := NewRegistry(tc.language)
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			if _, ok := r.GetFiles(tc.makeADL())[".env.example"]; ok {
				t.Fatalf(".env.example unexpectedly emitted when dockerCompose disabled (%s)", tc.language)
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
			tmplKey, ok := r.GetFiles(adl)[".env.example"]
			if !ok {
				t.Fatalf(".env.example missing when dockerCompose enabled (%s)", tc.language)
			}
			if tmplKey != "config/env.example" {
				t.Fatalf(".env.example mapped to %q, want %q", tmplKey, "config/env.example")
			}
			if _, err := r.GetTemplate(tmplKey); err != nil {
				t.Fatalf("template %q not loaded: %v", tmplKey, err)
			}
		})
	}
}

// TestEnvExampleTemplate_Essentials verifies the rendered .env.example
// is the minimal essentials-only form: agent client block from spec.agent,
// the full provider-key list, and the redis queue block only when the Rust
// `redis` Cargo feature is enabled. No explanatory comments are emitted.
func TestEnvExampleTemplate_Essentials(t *testing.T) {
	providerKeys := []string{
		"ANTHROPIC_API_KEY=",
		"CLOUDFLARE_API_KEY=",
		"COHERE_API_KEY=",
		"GROQ_API_KEY=",
		"OLLAMA_API_KEY=",
		"OLLAMA_CLOUD_API_KEY=",
		"OPENAI_API_KEY=",
		"DEEPSEEK_API_KEY=",
		"GOOGLE_API_KEY=",
		"MISTRAL_API_KEY=",
		"MOONSHOT_API_KEY=",
	}

	t.Run("go agent renders all provider keys and agent block", func(t *testing.T) {
		registry, err := NewRegistry("go")
		if err != nil {
			t.Fatalf("NewRegistry: %v", err)
		}
		engine := NewWithRegistry("minimal", registry)
		adl := minimalGoADL()
		adl.Spec.Agent = &schema.Agent{Provider: "openai", Model: "gpt-5-mini"}

		out, err := engine.ExecuteTemplate("config/env.example", Context{ADL: adl, Language: "go"})
		if err != nil {
			t.Fatalf("ExecuteTemplate: %v", err)
		}
		if !strings.Contains(out, "A2A_AGENT_CLIENT_PROVIDER=openai") {
			t.Errorf("env.example missing agent provider\n---\n%s", out)
		}
		if !strings.Contains(out, "CLI_PROVIDER=openai") || !strings.Contains(out, "CLI_MODEL=gpt-5-mini") {
			t.Errorf("env.example missing CLI defaults derived from spec.agent\n---\n%s", out)
		}
		for _, key := range providerKeys {
			if !strings.Contains(out, key) {
				t.Errorf("env.example missing provider key %q\n---\n%s", key, out)
			}
		}
		for _, section := range []string{"# Gateway", "# A2A Agent Server / LLM client", "# CLI"} {
			if !strings.Contains(out, section) {
				t.Errorf("env.example missing section header %q\n---\n%s", section, out)
			}
		}
		if strings.Contains(out, "RUST_LOG") {
			t.Errorf("env.example unexpectedly contains RUST_LOG\n---\n%s", out)
		}
	})

	t.Run("rust agent with redis includes queue vars", func(t *testing.T) {
		registry, err := NewRegistry("rust")
		if err != nil {
			t.Fatalf("NewRegistry: %v", err)
		}
		engine := NewWithRegistry("minimal", registry)
		adl := minimalRustADL()
		adl.Spec.Agent = &schema.Agent{Provider: "openai", Model: "gpt-5-mini"}
		adl.Spec.Language.Rust.Features = []string{"redis"}

		out, err := engine.ExecuteTemplate("config/env.example", Context{ADL: adl, Language: "rust"})
		if err != nil {
			t.Fatalf("ExecuteTemplate: %v", err)
		}
		if !strings.Contains(out, "A2A_QUEUE_PROVIDER=redis") {
			t.Errorf("rust env.example missing redis queue vars\n---\n%s", out)
		}
	})

	t.Run("rust agent without redis omits queue vars", func(t *testing.T) {
		registry, err := NewRegistry("rust")
		if err != nil {
			t.Fatalf("NewRegistry: %v", err)
		}
		engine := NewWithRegistry("minimal", registry)
		adl := minimalRustADL()
		adl.Spec.Agent = &schema.Agent{Provider: "openai", Model: "gpt-5-mini"}

		out, err := engine.ExecuteTemplate("config/env.example", Context{ADL: adl, Language: "rust"})
		if err != nil {
			t.Fatalf("ExecuteTemplate: %v", err)
		}
		if strings.Contains(out, "A2A_QUEUE_PROVIDER") {
			t.Errorf("rust env.example unexpectedly contains queue vars without redis feature\n---\n%s", out)
		}
	})
}
