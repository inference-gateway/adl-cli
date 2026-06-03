package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

func minimalTypeScriptADL() *schema.ADL {
	return &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "typescript-agent",
			Description: "test",
			Version:     "0.1.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{Streaming: true},
			Server:       schema.Server{Port: 8080},
			Language: schema.Language{
				TypeScript: &schema.TypeScriptConfig{
					PackageName: "typescript-agent",
					NodeVersion: "24",
				},
			},
		},
	}
}

// TestRegistry_Gitignore_AllLanguages verifies every supported language maps
// .gitignore to the shared config/gitignore template and loads it.
func TestRegistry_Gitignore_AllLanguages(t *testing.T) {
	cases := []struct {
		name     string
		language string
		makeADL  func() *schema.ADL
	}{
		{name: "go", language: "go", makeADL: minimalGoADL},
		{name: "rust", language: "rust", makeADL: minimalRustADL},
		{name: "typescript", language: "typescript", makeADL: minimalTypeScriptADL},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRegistry(tc.language)
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			tmplKey, ok := r.GetFiles(tc.makeADL())[".gitignore"]
			if !ok {
				t.Fatalf(".gitignore missing from generated files for %s", tc.language)
			}
			if tmplKey != "config/gitignore" {
				t.Fatalf(".gitignore mapped to %q, want %q", tmplKey, "config/gitignore")
			}
			if _, err := r.GetTemplate(tmplKey); err != nil {
				t.Fatalf("template %q not loaded: %v", tmplKey, err)
			}
		})
	}
}

func renderGitignore(t *testing.T, language string, adl *schema.ADL) string {
	t.Helper()
	registry, err := NewRegistry(language)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	out, err := NewWithRegistry("minimal", registry).ExecuteTemplate("config/gitignore", Context{
		ADL:      adl,
		Language: language,
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate: %v", err)
	}
	return out
}

// TestGitignoreTemplate_TypeScript locks issue #197: a generated TypeScript
// project must ignore node_modules/ and the common Node/TS build artifacts,
// and must NOT carry the Go-centric entries.
func TestGitignoreTemplate_TypeScript(t *testing.T) {
	out := renderGitignore(t, "typescript", minimalTypeScriptADL())

	if !strings.HasPrefix(out, "# Node.js dependencies") {
		t.Errorf("typescript .gitignore should start with the Node section (no leading blank line)\n---\n%s", out)
	}

	for _, want := range []string{
		"node_modules/",
		"*.tsbuildinfo",
		"npm-debug.log*",
		"yarn-error.log",
		".pnpm-debug.log*",
		// shared blocks still present
		"bin/",
		"dist/",
		".env*",
		"!.env*.example",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("typescript .gitignore missing %q\n---\n%s", want, out)
		}
	}

	for _, banned := range []string{"go.work", "*.test"} {
		if strings.Contains(out, banned) {
			t.Errorf("typescript .gitignore should not contain Go entry %q\n---\n%s", banned, out)
		}
	}
}

// TestGitignoreTemplate_Go guards against regressing the Go ignore list while
// making the template language-aware - the Go-centric entries must remain and
// node_modules/ must NOT leak in.
func TestGitignoreTemplate_Go(t *testing.T) {
	out := renderGitignore(t, "go", minimalGoADL())

	if !strings.HasPrefix(out, "# Binaries for programs and plugins") {
		t.Errorf("go .gitignore should start with the binaries section (no leading blank line)\n---\n%s", out)
	}

	for _, want := range []string{
		"*.test",
		"go.work",
		"*.exe",
		"bin/",
		"dist/",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("go .gitignore missing %q\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, "node_modules/") {
		t.Errorf("go .gitignore should not contain node_modules/\n---\n%s", out)
	}
}

// TestGitignoreTemplate_Rust verifies Rust no longer inherits the Go-centric
// entries and gets neither node_modules/ nor go.work.
func TestGitignoreTemplate_Rust(t *testing.T) {
	out := renderGitignore(t, "rust", minimalRustADL())

	if !strings.Contains(out, "target/") {
		t.Errorf("rust .gitignore missing target/\n---\n%s", out)
	}
	for _, banned := range []string{"node_modules/", "go.work", "*.test"} {
		if strings.Contains(out, banned) {
			t.Errorf("rust .gitignore should not contain %q\n---\n%s", banned, out)
		}
	}
}
