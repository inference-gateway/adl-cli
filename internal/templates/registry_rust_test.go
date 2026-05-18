package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

func minimalRustADL() *schema.ADL {
	return &schema.ADL{
		APIVersion: "adl.dev/v1",
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
					Version:     "1.88",
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

	adl.Spec.Sandbox = &schema.SandboxConfig{
		DockerCompose: &schema.DockerComposeConfig{Enabled: true},
	}
	if _, ok := r.getRustFiles(adl)["docker-compose.yaml"]; !ok {
		t.Fatalf("docker-compose.yaml missing when sandbox.dockerCompose.enabled=true")
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
