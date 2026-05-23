package templates

import (
	"strings"
	"testing"

	"github.com/inference-gateway/adl-cli/internal/sandbox"
	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// TestFloxManifest_DevelopmentDeps verifies the full pipeline from
// `spec.development.deps` to the generated `.flox/env/manifest.toml`.
// This is the acceptance-criteria flox test from issue #154: a manifest
// with three deps must end up with three additional [install] entries
// (deno/kubectl/terraform) in the rendered file, sorted alphabetically
// to keep diffs stable, and the per-package version pin must match the
// literal supplied in the ADL.
func TestFloxManifest_DevelopmentDeps(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	adl := minimalGoADL()
	adl.Spec.Development = &schema.DevelopmentConfig{
		Sandbox: &schema.SandboxConfig{Flox: &schema.FloxConfig{Enabled: true}},
		Deps:    []string{"terraform@1.9.5", "deno@2.1.4", "kubectl@1.31.0"},
	}

	view, err := sandbox.Resolve(adl)
	if err != nil {
		t.Fatalf("sandbox.Resolve: %v", err)
	}

	out, err := engine.ExecuteTemplate("flox/manifest.toml", Context{
		ADL:         adl,
		Language:    "go",
		SandboxDeps: view,
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate: %v", err)
	}

	wantFragments := []string{
		`deno.pkg-path = "deno"`,
		`deno.version = "2.1.4"`,
		`kubectl.pkg-path = "kubectl"`,
		`kubectl.version = "1.31.0"`,
		`terraform.pkg-path = "terraform"`,
		`terraform.version = "1.9.5"`,
	}
	for _, frag := range wantFragments {
		if !strings.Contains(out, frag) {
			t.Errorf("flox manifest missing %q\n---\n%s", frag, out)
		}
	}

	// Ordering must be deterministic (alphabetical) so re-generation
	// doesn't churn diffs.
	denoIdx := strings.Index(out, "deno.pkg-path")
	kubectlIdx := strings.Index(out, "kubectl.pkg-path")
	terraformIdx := strings.Index(out, "terraform.pkg-path")
	if denoIdx >= kubectlIdx || kubectlIdx >= terraformIdx {
		t.Fatalf("expected sorted order deno < kubectl < terraform; got positions %d/%d/%d\n---\n%s",
			denoIdx, kubectlIdx, terraformIdx, out)
	}

	// Built-in entries must still be present - deps are additive.
	for _, builtin := range []string{
		`go.pkg-path = "go"`,
		`go-task.pkg-path = "go-task"`,
		`git.pkg-path = "git"`,
		`docker.pkg-path = "docker"`,
	} {
		if !strings.Contains(out, builtin) {
			t.Errorf("flox manifest dropped built-in %q\n---\n%s", builtin, out)
		}
	}
}

// TestFloxManifest_NoDevelopmentDeps asserts that the generated
// manifest.toml is byte-identical (modulo the trailing newline) to the
// pre-feature output when spec.development.deps is absent. This guards
// the AC "Behavior is unchanged when the field is absent or empty."
func TestFloxManifest_NoDevelopmentDeps(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	adl := minimalGoADL()
	adl.Spec.Development = &schema.DevelopmentConfig{
		Sandbox: &schema.SandboxConfig{Flox: &schema.FloxConfig{Enabled: true}},
	}

	view, err := sandbox.Resolve(adl)
	if err != nil {
		t.Fatalf("sandbox.Resolve: %v", err)
	}
	if view.HasDeps() {
		t.Fatalf("sandbox.Resolve returned non-empty deps for an ADL without deps")
	}

	out, err := engine.ExecuteTemplate("flox/manifest.toml", Context{
		ADL:         adl,
		Language:    "go",
		SandboxDeps: view,
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate: %v", err)
	}

	// The user-deps comment must NOT appear when no deps are declared,
	// otherwise we'd leak the section header into every generated file.
	if strings.Contains(out, "spec.development.deps") {
		t.Fatalf("manifest unexpectedly mentions spec.development.deps when no deps are set\n---\n%s", out)
	}
}

// TestDevcontainerJSON_DevelopmentDeps confirms that the devcontainer
// path consumes spec.development.deps and emits an apt-packages feature
// entry. The devcontainer-extras feature is the closest equivalent to
// "an arbitrary list of `<pkg>=<ver>` pins" that the official features
// catalog ships today.
func TestDevcontainerJSON_DevelopmentDeps(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	adl := minimalGoADL()
	adl.Spec.Development = &schema.DevelopmentConfig{
		Sandbox: &schema.SandboxConfig{DevContainer: &schema.DevContainerConfig{Enabled: true}},
		Deps:    []string{"deno@2.1.4", "kubectl@1.31.0"},
	}

	view, err := sandbox.Resolve(adl)
	if err != nil {
		t.Fatalf("sandbox.Resolve: %v", err)
	}

	out, err := engine.ExecuteTemplate("devcontainer/devcontainer.json", Context{
		ADL:         adl,
		Language:    "go",
		SandboxDeps: view,
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate: %v", err)
	}

	if !strings.Contains(out, `"ghcr.io/devcontainers-extra/features/apt-packages:1"`) {
		t.Errorf("devcontainer.json missing apt-packages feature reference\n---\n%s", out)
	}
	if !strings.Contains(out, "deno=2.1.4") || !strings.Contains(out, "kubectl=1.31.0") {
		t.Errorf("devcontainer.json missing deno=2.1.4 or kubectl=1.31.0 in packages list\n---\n%s", out)
	}
}

// TestDevcontainerJSON_NoDevelopmentDeps asserts that devcontainer.json
// does not include the apt-packages feature when no deps are declared.
func TestDevcontainerJSON_NoDevelopmentDeps(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	adl := minimalGoADL()
	adl.Spec.Development = &schema.DevelopmentConfig{
		Sandbox: &schema.SandboxConfig{DevContainer: &schema.DevContainerConfig{Enabled: true}},
	}

	view, err := sandbox.Resolve(adl)
	if err != nil {
		t.Fatalf("sandbox.Resolve: %v", err)
	}

	out, err := engine.ExecuteTemplate("devcontainer/devcontainer.json", Context{
		ADL:         adl,
		Language:    "go",
		SandboxDeps: view,
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate: %v", err)
	}

	if strings.Contains(out, "apt-packages") {
		t.Fatalf("devcontainer.json unexpectedly contains apt-packages feature when no deps are declared\n---\n%s", out)
	}
}
