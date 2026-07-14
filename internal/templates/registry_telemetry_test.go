package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// TestTelemetryEnabled covers the three states of spec.telemetry: absent
// (nil block), explicitly disabled, and enabled. Telemetry is off by default,
// so only the last case reports true.
func TestTelemetryEnabled(t *testing.T) {
	cases := []struct {
		name string
		tel  *schema.TelemetryConfig
		want bool
	}{
		{"nil block", nil, false},
		{"disabled", &schema.TelemetryConfig{Enabled: false}, false},
		{"enabled", &schema.TelemetryConfig{Enabled: true}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adl := minimalGoADL()
			adl.Spec.Telemetry = tc.tel
			if got := telemetryEnabled(adl); got != tc.want {
				t.Errorf("telemetryEnabled = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestHasBuiltinTool verifies the reserved-tool detector that gates the shared
// span helper: any of read/bash/write/edit/fetch counts, custom tools do not.
func TestHasBuiltinTool(t *testing.T) {
	cases := []struct {
		name  string
		tools []schema.Tool
		want  bool
	}{
		{"no tools", nil, false},
		{"only custom", []schema.Tool{{ID: "weather"}}, false},
		{"one builtin", []schema.Tool{{ID: "read"}}, true},
		{"custom and builtin", []schema.Tool{{ID: "weather"}, {ID: "bash"}}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adl := minimalGoADL()
			adl.Spec.Tools = tc.tools
			if got := hasBuiltinTool(adl); got != tc.want {
				t.Errorf("hasBuiltinTool = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestRegistry_getGoFiles_TelemetryFile checks that tools/telemetry.go is
// mapped only when telemetry is on AND at least one built-in tool exists - the
// helper is used exclusively by built-in tool handlers, so it would be dead
// code otherwise.
func TestRegistry_getGoFiles_TelemetryFile(t *testing.T) {
	cases := []struct {
		name      string
		telemetry bool
		tools     []schema.Tool
		want      bool
	}{
		{"on with builtin", true, []schema.Tool{{ID: "read"}}, true},
		{"on with only custom", true, []schema.Tool{{ID: "weather"}}, false},
		{"off with builtin", false, []schema.Tool{{ID: "read"}}, false},
		{"off no tools", false, nil, false},
	}

	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adl := minimalGoADL()
			adl.Spec.Tools = tc.tools
			if tc.telemetry {
				adl.Spec.Telemetry = &schema.TelemetryConfig{Enabled: true}
			}

			files := r.getGoFiles(adl)
			key, ok := files["tools/telemetry.go"]
			if ok != tc.want {
				t.Fatalf("tools/telemetry.go present = %v, want %v", ok, tc.want)
			}
			if tc.want && key != "telemetry.go" {
				t.Errorf("expected template key telemetry.go, got %q", key)
			}
		})
	}
}

// TestRegistry_TelemetryTemplate_Renders renders the span-helper template end
// to end and confirms it wires the module-scoped tracer name and exposes the
// startToolSpan entrypoint the built-in handlers call.
func TestRegistry_TelemetryTemplate_Renders(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	adl := minimalGoADL()
	adl.Spec.Telemetry = &schema.TelemetryConfig{Enabled: true}

	out, err := engine.ExecuteTemplate("telemetry.go", Context{ADL: adl})
	if err != nil {
		t.Fatalf("ExecuteTemplate: %v", err)
	}
	if !strings.Contains(out, "package tools") {
		t.Fatalf("rendered telemetry helper missing package declaration:\n%s", out)
	}
	if !strings.Contains(out, "func startToolSpan(") {
		t.Fatalf("rendered telemetry helper missing startToolSpan:\n%s", out)
	}
	if !strings.Contains(out, adl.Spec.Language.Go.Module+"/tools") {
		t.Fatalf("rendered telemetry helper missing module-scoped tracer name:\n%s", out)
	}
}

// TestBuiltinTool_TelemetrySpanGating renders the Bash built-in with telemetry
// on and off and asserts the span is emitted only when enabled - the .env flag
// controls runtime export, but the span call itself is a generation-time gate.
func TestBuiltinTool_TelemetrySpanGating(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	cfg, err := schema.DecodeBuiltinToolConfig("bash", nil)
	if err != nil {
		t.Fatalf("DecodeBuiltinToolConfig: %v", err)
	}

	render := func(enabled bool) string {
		out, execErr := engine.ExecuteToolTemplate("builtin/bash.go", map[string]any{
			"ID":               "bash",
			"Config":           cfg,
			"TelemetryEnabled": enabled,
		})
		if execErr != nil {
			t.Fatalf("ExecuteToolTemplate(enabled=%v): %v", enabled, execErr)
		}
		return out
	}

	on := render(true)
	if !strings.Contains(on, `startToolSpan(ctx, "bash")`) {
		t.Errorf("telemetry-on Bash handler missing span call:\n%s", on)
	}

	off := render(false)
	if strings.Contains(off, "startToolSpan") {
		t.Errorf("telemetry-off Bash handler should not reference startToolSpan:\n%s", off)
	}
}
