package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

func minimalRustADL() *schema.ADL {
	return &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "rust-agent",
			Description: "test",
			Version:     "0.1.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{Streaming: true},
			Server:       schema.Server{Port: 8080},
			Language: schema.Language{
				Rust: &schema.RustConfig{
					PackageName: "rust-agent",
					Version:     "1.94.1",
					Edition:     "2024",
				},
			},
		},
	}
}

func TestRegistry_getRustFiles_AlwaysIncludesEnvExample(t *testing.T) {
	r, err := NewRegistry("rust")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	files := r.getRustFiles(minimalRustADL())
	if _, ok := files[".env.example"]; !ok {
		t.Fatalf(".env.example missing from generated files: %v", files)
	}
}

func TestRegistry_getRustFiles_DockerComposeOnlyWhenEnabled(t *testing.T) {
	r, err := NewRegistry("rust")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	adl := minimalRustADL()
	if _, ok := r.getRustFiles(adl)["docker-compose.yaml"]; ok {
		t.Fatalf("docker-compose.yaml unexpectedly emitted when sandbox flag unset")
	}

	adl.Spec.Development = &schema.DevelopmentConfig{
		Sandbox: &schema.SandboxConfig{
			DockerCompose: &schema.DockerComposeConfig{Enabled: true},
		},
	}
	if _, ok := r.getRustFiles(adl)["docker-compose.yaml"]; !ok {
		t.Fatalf("docker-compose.yaml missing when sandbox.dockerCompose.enabled=true")
	}
}

// TestRustBuiltinTemplates_ContainTestModule verifies that each Rust
// built-in template ships an inline `#[cfg(test)] mod tests` block so
// the generated agent has unit tests next to each built-in tool.
func TestRustBuiltinTemplates_ContainTestModule(t *testing.T) {
	r, err := NewRegistry("rust")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	for _, id := range []string{"read", "bash", "write", "edit", "fetch"} {
		key := "builtin/" + id + ".rs"
		tmpl, err := r.GetTemplate(key)
		if err != nil {
			t.Errorf("GetTemplate(%q): %v", key, err)
			continue
		}
		if !strings.Contains(tmpl, "#[cfg(test)]") {
			t.Errorf("template %q missing #[cfg(test)] test module", key)
		}
		if !strings.Contains(tmpl, "mod tests") {
			t.Errorf("template %q missing mod tests block", key)
		}
	}
}

// TestRustCargoToml_DevDepsForBuiltins verifies that tempfile shows up
// as a dev-dependency when the spec includes any reserved built-in -
// the Rust unit tests rely on it for TempDir-based fixtures.
func TestRustCargoToml_DevDepsForBuiltins(t *testing.T) {
	cases := []struct {
		name    string
		toolIDs []string
		wantDev bool
	}{
		{name: "no tools -> no dev-deps", toolIDs: nil, wantDev: false},
		{name: "custom-only -> no dev-deps", toolIDs: []string{"my_custom"}, wantDev: false},
		{name: "any builtin -> dev-deps", toolIDs: []string{"read"}, wantDev: true},
		{name: "multiple builtins -> dev-deps", toolIDs: []string{"read", "bash", "edit"}, wantDev: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			registry, err := NewRegistry("rust")
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			engine := NewWithRegistry("minimal", registry)

			adl := minimalRustADL()
			for _, id := range tc.toolIDs {
				tool := schema.Tool{ID: id}
				if !schema.IsReservedToolID(id) {
					tool.Name = id
					tool.Description = "custom"
					tool.Tags = []string{"custom"}
					tool.Schema = schema.ToolSchema{}
				}
				adl.Spec.Tools = append(adl.Spec.Tools, tool)
			}

			out, err := engine.ExecuteTemplate("Cargo.toml", Context{ADL: adl})
			if err != nil {
				t.Fatalf("ExecuteTemplate: %v", err)
			}
			hasDev := strings.Contains(out, "[dev-dependencies]") && strings.Contains(out, `tempfile = "3"`)
			if hasDev != tc.wantDev {
				t.Fatalf("dev-deps present=%v, want=%v\nCargo.toml:\n%s", hasDev, tc.wantDev, out)
			}
		})
	}
}

func TestRustCargoToml_RedisFeatureFlag(t *testing.T) {
	cases := []struct {
		name     string
		features []string
		want     string
	}{
		{
			name:     "no features -> plain dep",
			features: nil,
			want:     `inference-gateway-adk = "0.4.3"`,
		},
		{
			name:     "redis feature -> feature flag",
			features: []string{"redis"},
			want:     `inference-gateway-adk = { version = "0.4.3", features = ["redis"] }`,
		},
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

			out, err := engine.ExecuteTemplate("Cargo.toml", Context{ADL: adl})
			if err != nil {
				t.Fatalf("ExecuteTemplate: %v", err)
			}
			if !strings.Contains(out, tc.want) {
				t.Fatalf("Cargo.toml missing %q\nfull output:\n%s", tc.want, out)
			}
		})
	}
}
