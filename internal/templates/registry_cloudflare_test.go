package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// cloudflareTSADL returns a minimal TypeScript ADL wired for Cloudflare Workers
// deployment. Individual tests tweak the returned Cloudflare block to exercise
// edge cases.
func cloudflareTSADL() *schema.ADL {
	return &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "cloudflare-agent",
			Description: "test",
			Version:     "0.1.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{Streaming: true},
			Server:       schema.Server{Port: 8080},
			Language: schema.Language{
				TypeScript: &schema.TypeScriptConfig{
					PackageName: "cloudflare-agent",
					NodeVersion: "24",
				},
			},
			Deployment: &schema.DeploymentConfig{
				Type: schema.DeploymentConfigTypeCloudflare,
				Cloudflare: &schema.CloudflareConfig{
					Name:               "support-agent",
					AccountID:          "${CLOUDFLARE_ACCOUNT_ID}",
					CompatibilityDate:  "2024-11-01",
					CompatibilityFlags: []string{"nodejs_compat", "streams_enable_constructors"},
					Routes:             []string{"agent.example.com/*"},
					WorkersDev:         false,
					Environment:        schema.CloudflareConfigEnvironment{"LOG_LEVEL": "info"},
				},
			},
		},
	}
}

// TestRegistry_CloudflareFiles_Mapping verifies that a cloudflare deployment
// maps the wrangler config and the TypeScript Worker entrypoint onto the
// project layout.
func TestRegistry_CloudflareFiles_Mapping(t *testing.T) {
	r, err := NewRegistry("typescript")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	files := r.GetFiles(cloudflareTSADL())

	if got := files["wrangler.toml"]; got != "cloudflare/wrangler.toml" {
		t.Errorf("wrangler.toml mapping = %q, want cloudflare/wrangler.toml", got)
	}
	if got := files["src/worker.ts"]; got != "worker.ts" {
		t.Errorf("src/worker.ts mapping = %q, want worker.ts", got)
	}
}

// TestRegistry_CloudflareFiles_OnlyWhenCloudflare ensures non-cloudflare
// deployments do not pull in the wrangler/Worker artifacts.
func TestRegistry_CloudflareFiles_OnlyWhenCloudflare(t *testing.T) {
	r, err := NewRegistry("typescript")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	adl := cloudflareTSADL()
	adl.Spec.Deployment = &schema.DeploymentConfig{Type: schema.DeploymentConfigTypeKubernetes}

	files := r.GetFiles(adl)
	if _, ok := files["wrangler.toml"]; ok {
		t.Errorf("wrangler.toml should not be generated for a kubernetes deployment")
	}
	if _, ok := files["src/worker.ts"]; ok {
		t.Errorf("src/worker.ts should not be generated for a kubernetes deployment")
	}
}

// TestRegistry_CloudflareFiles_GoOmitsWorker confirms the Worker entrypoint is
// TypeScript-only: a Go agent targeting cloudflare gets no wrangler/Worker
// artifacts (Workers run TS on the edge runtime, not a Go binary).
func TestRegistry_CloudflareFiles_GoOmitsWorker(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	adl := cloudflareTSADL()
	adl.Spec.Language = schema.Language{Go: &schema.GoConfig{Module: "example.com/agent", Version: "1.26.2"}}

	files := r.GetFiles(adl)
	if _, ok := files["wrangler.toml"]; ok {
		t.Errorf("wrangler.toml should not be generated for a Go cloudflare deployment")
	}
	if _, ok := files["src/worker.ts"]; ok {
		t.Errorf("src/worker.ts should not be generated for a Go cloudflare deployment")
	}
}

func renderWranglerTOML(t *testing.T, adl *schema.ADL) string {
	t.Helper()
	r, err := NewRegistry("typescript")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	out, err := NewWithRegistry("minimal", r).ExecuteTemplate("cloudflare/wrangler.toml", Context{ADL: adl, Language: "typescript"})
	if err != nil {
		t.Fatalf("ExecuteTemplate(cloudflare/wrangler.toml): %v", err)
	}
	return out
}

// TestRegistry_WranglerTOML_FullFieldSet verifies the full field set renders,
// including the ${VAR} account-id placeholder, custom compatibility flags,
// custom routes, the [vars] table, and workers_dev = false when the Worker is
// served exclusively via routes.
func TestRegistry_WranglerTOML_FullFieldSet(t *testing.T) {
	out := renderWranglerTOML(t, cloudflareTSADL())

	for _, want := range []string{
		`name = "support-agent"`,
		`main = "src/worker.ts"`,
		`compatibility_date = "2024-11-01"`,
		`compatibility_flags = ["nodejs_compat", "streams_enable_constructors"]`,
		`account_id = "${CLOUDFLARE_ACCOUNT_ID}"`,
		`workers_dev = false`,
		`routes = ["agent.example.com/*"]`,
		"[vars]",
		`LOG_LEVEL = "info"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("wrangler.toml missing %q\n---\n%s", want, out)
		}
	}
}

// TestRegistry_WranglerTOML_NameFallsBackToMetadata verifies the wrangler name
// falls back to metadata.name when the cloudflare block omits an explicit name.
func TestRegistry_WranglerTOML_NameFallsBackToMetadata(t *testing.T) {
	adl := cloudflareTSADL()
	adl.Spec.Deployment.Cloudflare.Name = ""

	out := renderWranglerTOML(t, adl)

	if !strings.Contains(out, `name = "cloudflare-agent"`) {
		t.Errorf("expected name to fall back to metadata.name\n---\n%s", out)
	}
}

// TestRegistry_WranglerTOML_DefaultsCompatibilityDate verifies the generator
// supplies a default compatibility_date when the manifest omits it.
func TestRegistry_WranglerTOML_DefaultsCompatibilityDate(t *testing.T) {
	adl := cloudflareTSADL()
	adl.Spec.Deployment.Cloudflare.CompatibilityDate = ""

	out := renderWranglerTOML(t, adl)

	if !strings.Contains(out, `compatibility_date = "2025-01-01"`) {
		t.Errorf("expected a default compatibility_date\n---\n%s", out)
	}
}

// TestRegistry_WranglerTOML_DefaultsCompatibilityFlags verifies nodejs_compat
// is supplied by default when the manifest declares no flags - the scaffold
// relies on Node.js API compatibility.
func TestRegistry_WranglerTOML_DefaultsCompatibilityFlags(t *testing.T) {
	adl := cloudflareTSADL()
	adl.Spec.Deployment.Cloudflare.CompatibilityFlags = nil

	out := renderWranglerTOML(t, adl)

	if !strings.Contains(out, `compatibility_flags = ["nodejs_compat"]`) {
		t.Errorf("expected default nodejs_compat flag\n---\n%s", out)
	}
}

// TestRegistry_WranglerTOML_WorkersDevTrue verifies workers_dev = true renders
// when the manifest opts into the *.workers.dev subdomain.
func TestRegistry_WranglerTOML_WorkersDevTrue(t *testing.T) {
	adl := cloudflareTSADL()
	adl.Spec.Deployment.Cloudflare.WorkersDev = true

	out := renderWranglerTOML(t, adl)

	if !strings.Contains(out, "workers_dev = true") {
		t.Errorf("expected workers_dev = true\n---\n%s", out)
	}
}

// TestRegistry_WranglerTOML_WorkersDevOmittedWithoutRoutes verifies that a bare
// block (workersDev false, no routes) omits workers_dev entirely so wrangler's
// own default keeps the Worker reachable on *.workers.dev.
func TestRegistry_WranglerTOML_WorkersDevOmittedWithoutRoutes(t *testing.T) {
	adl := cloudflareTSADL()
	adl.Spec.Deployment.Cloudflare.WorkersDev = false
	adl.Spec.Deployment.Cloudflare.Routes = nil

	out := renderWranglerTOML(t, adl)

	if strings.Contains(out, "workers_dev") {
		t.Errorf("workers_dev should be omitted when false and no routes are set\n---\n%s", out)
	}
	if strings.Contains(out, "routes =") {
		t.Errorf("routes should be omitted when none are set\n---\n%s", out)
	}
}

// TestRegistry_WranglerTOML_NilBlockRendersDefaults verifies a manifest that
// sets type: cloudflare without a cloudflare block still renders a valid
// wrangler.toml with sensible defaults rather than panicking.
func TestRegistry_WranglerTOML_NilBlockRendersDefaults(t *testing.T) {
	adl := cloudflareTSADL()
	adl.Spec.Deployment.Cloudflare = nil

	out := renderWranglerTOML(t, adl)

	for _, want := range []string{
		`name = "cloudflare-agent"`,
		`main = "src/worker.ts"`,
		`compatibility_date = "2025-01-01"`,
		`compatibility_flags = ["nodejs_compat"]`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("wrangler.toml missing %q\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, "[vars]") {
		t.Errorf("no [vars] table expected for a nil cloudflare block\n---\n%s", out)
	}
}
