package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

func minimalGoADL() *schema.ADL {
	return &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "go-agent",
			Description: "test",
			Version:     "0.1.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{Streaming: true},
			Server:       schema.Server{Port: 8080},
			Language: schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/go-agent",
					Version: "1.26.2",
				},
			},
		},
	}
}

// TestRegistry_getGoFiles_ScaffoldsTestForEachBuiltin verifies that each
// reserved built-in tool emits both an implementation and a unit-test
// file - the second AC of #138.
func TestRegistry_getGoFiles_ScaffoldsTestForEachBuiltin(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	adl := minimalGoADL()
	adl.Spec.Tools = []schema.Tool{
		{ID: "read"},
		{ID: "bash"},
		{ID: "write"},
		{ID: "edit"},
		{ID: "fetch"},
	}

	files := r.getGoFiles(adl)

	for _, id := range []string{"read", "bash", "write", "edit", "fetch"} {
		implPath := "tools/" + id + ".go"
		testPath := "tools/" + id + "_test.go"

		implKey, ok := files[implPath]
		if !ok {
			t.Errorf("missing implementation file mapping for %s", implPath)
			continue
		}
		if implKey != "builtin/"+id+".go" {
			t.Errorf("expected impl template builtin/%s.go for %s, got %q", id, implPath, implKey)
		}

		testKey, ok := files[testPath]
		if !ok {
			t.Errorf("missing test file mapping for %s", testPath)
			continue
		}
		if testKey != "builtin/"+id+"_test.go" {
			t.Errorf("expected test template builtin/%s_test.go for %s, got %q", id, testPath, testKey)
		}
	}
}

// TestRegistry_getGoFiles_NoTestForCustomTools ensures the test scaffold
// only ships with built-in tools. Custom tools deliberately don't get a
// test file - users can crib from the built-in tests as an example.
func TestRegistry_getGoFiles_NoTestForCustomTools(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	adl := minimalGoADL()
	adl.Spec.Tools = []schema.Tool{
		{
			ID:          "weather",
			Name:        "get_weather",
			Description: "Get weather data",
			Tags:        []string{"weather"},
			Schema:      schema.ToolSchema{},
		},
	}

	files := r.getGoFiles(adl)

	if _, ok := files["tools/weather_test.go"]; ok {
		t.Errorf("custom tool should not receive a _test.go scaffold; got entry in files map")
	}
	if _, ok := files["tools/weather.go"]; !ok {
		t.Errorf("expected custom tool implementation tools/weather.go to be present")
	}
}

// TestRegistry_BuiltinTestTemplates_AreLoaded verifies that every
// builtin/<id>_test.go template ships with the binary and can be
// retrieved through the registry by templateKey.
func TestRegistry_BuiltinTestTemplates_AreLoaded(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	for _, id := range []string{"read", "bash", "write", "edit", "fetch"} {
		key := "builtin/" + id + "_test.go"
		tmpl, err := r.GetTemplate(key)
		if err != nil {
			t.Errorf("GetTemplate(%q): %v", key, err)
			continue
		}
		if !strings.Contains(tmpl, "package tools") {
			t.Errorf("template %q does not declare package tools", key)
		}
		if !strings.Contains(tmpl, "func Test") {
			t.Errorf("template %q has no Test* functions", key)
		}
	}
}

// TestRegistry_GoFetchTest_RendersFromTemplateEngine renders one of the
// built-in test templates end-to-end through the template engine, with a
// minimal ADL context. This guards against template-syntax regressions:
// the test files are deliberately static (no {{ ... }} variables), so
// any Go-template parsing error here means we accidentally introduced a
// template directive.
func TestRegistry_GoFetchTest_RendersFromTemplateEngine(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	adl := minimalGoADL()
	adl.Spec.Tools = []schema.Tool{{ID: "fetch"}}

	out, err := engine.ExecuteTemplate("builtin/fetch_test.go", Context{ADL: adl})
	if err != nil {
		t.Fatalf("ExecuteTemplate: %v", err)
	}
	if !strings.Contains(out, "package tools") {
		t.Fatalf("rendered template missing package declaration:\n%s", out)
	}
	if !strings.Contains(out, "func TestFetchTool_") {
		t.Fatalf("rendered template missing Fetch test functions:\n%s", out)
	}
}
