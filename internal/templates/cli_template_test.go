package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// TestGoMainTemplate_IsCobraCLI verifies that the generated Go main.go is
// wired up as a cobra CLI with a `start` subcommand and a root that
// exposes `--version`. Locks in the contract behind issue #140.
func TestGoMainTemplate_IsCobraCLI(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	tmpl, err := r.GetTemplate("main.go")
	if err != nil {
		t.Fatalf("GetTemplate(main.go): %v", err)
	}

	engine := NewWithRegistry("main.go", r)
	ctx := Context{
		ADL:      minimalGoADL(),
		Language: "go",
	}
	rendered, err := engine.Execute(tmpl, ctx)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	for _, want := range []string{
		`cobra "github.com/spf13/cobra"`,
		"func newRootCmd() *cobra.Command",
		"func newStartCmd() *cobra.Command",
		`Use:   "start"`,
		"Version:       Version",
		"root.AddCommand(newStartCmd())",
		"func runStart(ctx context.Context) error",
		"newRootCmd().ExecuteContext(ctx)",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("generated main.go missing %q\n--- rendered ---\n%s", want, rendered)
		}
	}

	if !strings.Contains(rendered, "func main() {") {
		t.Errorf("generated main.go is missing func main()")
	}
}

// TestGoModTemplate_DeclaresCobra verifies the cobra dependency lands in
// the rendered go.mod alongside the existing ADK deps.
func TestGoModTemplate_DeclaresCobra(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tmpl, err := r.GetTemplate("go.mod")
	if err != nil {
		t.Fatalf("GetTemplate(go.mod): %v", err)
	}
	engine := NewWithRegistry("go.mod", r)
	rendered, err := engine.Execute(tmpl, Context{ADL: minimalGoADL(), Language: "go"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(rendered, "github.com/spf13/cobra") {
		t.Errorf("go.mod does not declare cobra dependency:\n%s", rendered)
	}
}

// TestRustMainTemplate_IsClapCLI verifies the generated Rust main.rs is a
// clap-derive CLI with a `Start` subcommand. The agent must require an
// explicit subcommand to boot.
func TestRustMainTemplate_IsClapCLI(t *testing.T) {
	r, err := NewRegistry("rust")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tmpl, err := r.GetTemplate("main.rs")
	if err != nil {
		t.Fatalf("GetTemplate(main.rs): %v", err)
	}

	adl := &schema.ADL{
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

	engine := NewWithRegistry("main.rs", r)
	rendered, err := engine.Execute(tmpl, Context{ADL: adl, Language: "rust"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	for _, want := range []string{
		"use clap::{Parser, Subcommand}",
		"struct Cli {",
		"enum Commands {",
		"/// Start the A2A server",
		"Start,",
		"let cli = Cli::parse();",
		"Commands::Start => run_start().await,",
		"async fn run_start()",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("generated main.rs missing %q\n--- rendered ---\n%s", want, rendered)
		}
	}
}

// TestRustCargoToml_DeclaresClap verifies the clap dependency lands in
// the rendered Cargo.toml.
func TestRustCargoToml_DeclaresClap(t *testing.T) {
	r, err := NewRegistry("rust")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tmpl, err := r.GetTemplate("Cargo.toml")
	if err != nil {
		t.Fatalf("GetTemplate(Cargo.toml): %v", err)
	}
	adl := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:    "rust-agent",
			Version: "0.1.0",
		},
		Spec: schema.Spec{
			Server: schema.Server{Port: 8080},
			Language: schema.Language{
				Rust: &schema.RustConfig{
					PackageName: "rust-agent",
					Version:     "1.94.1",
					Edition:     "2024",
				},
			},
		},
	}
	engine := NewWithRegistry("Cargo.toml", r)
	rendered, err := engine.Execute(tmpl, Context{ADL: adl, Language: "rust"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(rendered, `clap = { version = "4", features = ["derive"] }`) {
		t.Errorf("Cargo.toml does not declare clap with derive feature:\n%s", rendered)
	}
}

// TestDockerfileGo_InvokesStart verifies the generated Go Dockerfile CMD
// invokes the `start` subcommand rather than running the bare binary.
// (Previously `CMD ["./main"]` - now the binary is a CLI.)
func TestDockerfileGo_InvokesStart(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tmpl, err := r.GetTemplate("docker/dockerfile.go")
	if err != nil {
		t.Fatalf("GetTemplate(docker/dockerfile.go): %v", err)
	}
	engine := NewWithRegistry("docker/dockerfile.go", r)
	rendered, err := engine.Execute(tmpl, Context{ADL: minimalGoADL(), Language: "go"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(rendered, `CMD ["./main", "start"]`) {
		t.Errorf("Dockerfile CMD does not invoke the start subcommand:\n%s", rendered)
	}
}

// TestDockerfileRust_InvokesStart mirrors the Go test for Rust.
func TestDockerfileRust_InvokesStart(t *testing.T) {
	r, err := NewRegistry("rust")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tmpl, err := r.GetTemplate("docker/dockerfile.rust")
	if err != nil {
		t.Fatalf("GetTemplate(docker/dockerfile.rust): %v", err)
	}
	adl := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:    "rust-agent",
			Version: "0.1.0",
		},
		Spec: schema.Spec{
			Server: schema.Server{Port: 8080},
			Language: schema.Language{
				Rust: &schema.RustConfig{
					PackageName: "rust-agent",
					Version:     "1.94.1",
					Edition:     "2024",
				},
			},
		},
	}
	engine := NewWithRegistry("docker/dockerfile.rust", r)
	rendered, err := engine.Execute(tmpl, Context{ADL: adl, Language: "rust"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(rendered, `CMD ["./rust-agent", "start"]`) {
		t.Errorf("Rust Dockerfile CMD does not invoke the start subcommand:\n%s", rendered)
	}
}
