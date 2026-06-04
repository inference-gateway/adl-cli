package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// vercelTSADL returns a minimal TypeScript ADL wired for Vercel deployment.
// Individual tests tweak the returned Vercel block to exercise edge cases.
func vercelTSADL() *schema.ADL {
	return &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "vercel-agent",
			Description: "test",
			Version:     "0.1.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{Streaming: true},
			Server:       schema.Server{Port: 8080},
			Language: schema.Language{
				TypeScript: &schema.TypeScriptConfig{
					PackageName: "vercel-agent",
					NodeVersion: "24",
				},
			},
			Deployment: &schema.DeploymentConfig{
				Type: schema.DeploymentConfigTypeVercel,
				Vercel: &schema.VercelConfig{
					Project:     "vercel-agent",
					Team:        "my-team",
					Framework:   "nextjs",
					Runtime:     schema.VercelConfigRuntimeEdge,
					Regions:     []string{"iad1"},
					Functions:   &schema.VercelConfigFunctions{Memory: 1024, MaxDuration: 60},
					Environment: schema.VercelConfigEnvironment{"LOG_LEVEL": "info"},
				},
			},
		},
	}
}

// TestRegistry_VercelFiles_Mapping verifies that a vercel deployment maps the
// two Vercel artifacts onto the project layout (AC: vercel.json + linkage).
func TestRegistry_VercelFiles_Mapping(t *testing.T) {
	r, err := NewRegistry("typescript")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	files := r.GetFiles(vercelTSADL())

	if got := files["vercel.json"]; got != "vercel/vercel.json" {
		t.Errorf("vercel.json mapping = %q, want vercel/vercel.json", got)
	}
	if got := files[".vercel/project.json"]; got != "vercel/project.json" {
		t.Errorf(".vercel/project.json mapping = %q, want vercel/project.json", got)
	}
}

// TestRegistry_VercelFiles_OnlyWhenVercel ensures non-vercel deployments do
// not pull in the Vercel artifacts.
func TestRegistry_VercelFiles_OnlyWhenVercel(t *testing.T) {
	r, err := NewRegistry("typescript")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	adl := vercelTSADL()
	adl.Spec.Deployment = &schema.DeploymentConfig{Type: schema.DeploymentConfigTypeKubernetes}

	files := r.GetFiles(adl)
	if _, ok := files["vercel.json"]; ok {
		t.Errorf("vercel.json should not be generated for a kubernetes deployment")
	}
}

func renderVercelJSON(t *testing.T, adl *schema.ADL) string {
	t.Helper()
	r, err := NewRegistry("typescript")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	out, err := NewWithRegistry("minimal", r).ExecuteTemplate("vercel/vercel.json", Context{ADL: adl, Language: "typescript"})
	if err != nil {
		t.Fatalf("ExecuteTemplate(vercel/vercel.json): %v", err)
	}
	return out
}

// TestRegistry_VercelJSON_EdgeRuntime verifies the full field set renders and
// that runtime: edge maps to the @vercel/edge function runtime (AC).
func TestRegistry_VercelJSON_EdgeRuntime(t *testing.T) {
	out := renderVercelJSON(t, vercelTSADL())

	for _, want := range []string{
		`"$schema"`,
		`"framework": "nextjs"`,
		`"iad1"`,
		`"memory": 1024`,
		`"maxDuration": 60`,
		`"runtime": "@vercel/edge"`,
		`"api/**/*.ts"`,
		`"LOG_LEVEL": "info"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("vercel.json missing %q\n---\n%s", want, out)
		}
	}
}

// TestRegistry_VercelJSON_NodeRuntime confirms the Node runtime does not emit
// the @vercel/edge marker - Vercel auto-selects the Node runtime.
func TestRegistry_VercelJSON_NodeRuntime(t *testing.T) {
	adl := vercelTSADL()
	adl.Spec.Deployment.Vercel.Runtime = schema.VercelConfigRuntimeNodejs

	out := renderVercelJSON(t, adl)

	if strings.Contains(out, "@vercel/edge") {
		t.Errorf("nodejs runtime should not emit @vercel/edge\n---\n%s", out)
	}
	// memory/maxDuration still render a functions block even without an
	// explicit runtime.
	if !strings.Contains(out, `"memory": 1024`) {
		t.Errorf("expected functions block to still render for nodejs\n---\n%s", out)
	}
}

// TestRegistry_VercelJSON_OmitsFrameworkWhenUnset verifies framework is left
// out so Vercel auto-detects (AC: "omitted for auto-detect").
func TestRegistry_VercelJSON_OmitsFrameworkWhenUnset(t *testing.T) {
	adl := vercelTSADL()
	adl.Spec.Deployment.Vercel.Framework = ""

	out := renderVercelJSON(t, adl)

	if strings.Contains(out, "framework") {
		t.Errorf("framework should be omitted when unset\n---\n%s", out)
	}
}

// TestRegistry_VercelProjectJSON verifies the linkage file carries the project
// and team from the manifest.
func TestRegistry_VercelProjectJSON(t *testing.T) {
	r, err := NewRegistry("typescript")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	out, err := NewWithRegistry("minimal", r).ExecuteTemplate("vercel/project.json", Context{ADL: vercelTSADL(), Language: "typescript"})
	if err != nil {
		t.Fatalf("ExecuteTemplate(vercel/project.json): %v", err)
	}

	for _, want := range []string{`"projectId": "vercel-agent"`, `"orgId": "my-team"`} {
		if !strings.Contains(out, want) {
			t.Errorf("project.json missing %q\n---\n%s", want, out)
		}
	}
}
