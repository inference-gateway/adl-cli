package templates

import (
	"encoding/json"
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// TestAgentCardTemplate_EscapesQuotesInDescriptions is the regression test
// for issue #150. The previous template inlined description strings into the
// JSON output with raw `"{{ .Description }}"`, so any inner double quote
// (e.g. a skill description that says: Triggers on phrases like "PromQL")
// terminated the JSON string early and produced an invalid `.well-known/agent-card.json`.
//
// The fix routes every string field through `toJson`, which delegates to
// `encoding/json` and escapes `"`, `\`, and control characters per RFC 8259.
// This test asserts the rendered output parses cleanly and the original
// strings survive round-tripping with their inner quotes intact.
func TestAgentCardTemplate_EscapesQuotesInDescriptions(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tmpl, err := r.GetTemplate("config/agent.json")
	if err != nil {
		t.Fatalf("GetTemplate(config/agent.json): %v", err)
	}

	adl := minimalGoADL()
	adl.Metadata.Description = `An agent that handles "PromQL" and "sum by" queries.`
	adl.Spec.Card = &schema.Card{
		ProtocolVersion:    "0.3.0",
		URL:                "https://example.com/agent",
		PreferredTransport: "JSONRPC",
		DocumentationURL:   "https://example.com/docs",
		IconURL:            "https://example.com/icon.png",
	}

	promqlDesc := `Generate Prometheus query expressions. Triggers on phrases like "PromQL", "Prometheus query", and "sum by".`
	dashboardDesc := `Author Grafana dashboards. Handles phrases such as "stat panel" and "time series".`

	engine := NewWithRegistry("config/agent.json", r)
	ctx := Context{
		ADL:      adl,
		Language: "go",
		Skills: []SkillView{
			{
				ID:          "promql",
				Name:        "PromQL",
				Description: promqlDesc,
				Tags:        []string{"prometheus", "metrics"},
				Version:     "1.0.0",
			},
			{
				ID:          "dashboarding",
				Name:        "Dashboarding",
				Description: dashboardDesc,
				Tags:        []string{"grafana"},
			},
		},
	}

	rendered, err := engine.Execute(tmpl, ctx)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var card struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Skills      []struct {
			ID          string   `json:"id"`
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Tags        []string `json:"tags"`
			Version     string   `json:"version,omitempty"`
		} `json:"skills"`
	}
	if err := json.Unmarshal([]byte(rendered), &card); err != nil {
		t.Fatalf("rendered agent-card.json is not valid JSON: %v\n--- rendered ---\n%s", err, rendered)
	}

	if card.Description != adl.Metadata.Description {
		t.Errorf("metadata description was not preserved through JSON encoding\nwant: %q\ngot:  %q", adl.Metadata.Description, card.Description)
	}

	if len(card.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(card.Skills))
	}
	if card.Skills[0].Description != promqlDesc {
		t.Errorf("promql skill description not preserved\nwant: %q\ngot:  %q", promqlDesc, card.Skills[0].Description)
	}
	if card.Skills[1].Description != dashboardDesc {
		t.Errorf("dashboarding skill description not preserved\nwant: %q\ngot:  %q", dashboardDesc, card.Skills[1].Description)
	}
	if card.Skills[0].Version != "1.0.0" {
		t.Errorf("promql skill version not preserved\nwant: %q\ngot:  %q", "1.0.0", card.Skills[0].Version)
	}
}

// TestAgentCardTemplate_EscapesBackslashesAndControlChars covers the rest of
// the RFC 8259 must-escape set so a careless future edit that re-introduces
// raw interpolation gets caught.
func TestAgentCardTemplate_EscapesBackslashesAndControlChars(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tmpl, err := r.GetTemplate("config/agent.json")
	if err != nil {
		t.Fatalf("GetTemplate(config/agent.json): %v", err)
	}

	adl := minimalGoADL()
	adl.Metadata.Description = "line one\nline two with a \\ backslash and a \"quote\""

	engine := NewWithRegistry("config/agent.json", r)
	rendered, err := engine.Execute(tmpl, Context{ADL: adl, Language: "go"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var card struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(rendered), &card); err != nil {
		t.Fatalf("rendered agent-card.json is not valid JSON: %v\n--- rendered ---\n%s", err, rendered)
	}
	if card.Description != adl.Metadata.Description {
		t.Errorf("description not preserved\nwant: %q\ngot:  %q", adl.Metadata.Description, card.Description)
	}
}

// TestGoMainTemplate_EscapesQuotesInDescription is the sibling regression for
// issue #150: the Go main.go template also embeds metadata.description as a
// raw Go string literal (the AgentDescription var). Inner double quotes used
// to produce invalid Go that would fail `go build`. The fix routes the value
// through `toJson`, which emits a JSON-encoded string that doubles as a valid
// Go interpreted-string literal (same escape semantics for `"`, `\`, `\n`).
func TestGoMainTemplate_EscapesQuotesInDescription(t *testing.T) {
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tmpl, err := r.GetTemplate("main.go")
	if err != nil {
		t.Fatalf("GetTemplate(main.go): %v", err)
	}

	adl := minimalGoADL()
	adl.Metadata.Description = `Agent for "PromQL" and "sum by" queries`

	engine := NewWithRegistry("main.go", r)
	rendered, err := engine.Execute(tmpl, Context{ADL: adl, Language: "go"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := `AgentDescription = "Agent for \"PromQL\" and \"sum by\" queries"`
	if !strings.Contains(rendered, want) {
		t.Errorf("main.go did not contain properly escaped AgentDescription literal\nwant substring: %s\n--- rendered ---\n%s", want, rendered)
	}

	if strings.Contains(rendered, `"Agent for "PromQL"`) {
		t.Errorf("main.go still contains raw (unescaped) quotes inside the description literal\n--- rendered ---\n%s", rendered)
	}
}

// TestRustMainTemplate_EscapesQuotesInDescription is the sibling regression
// for issue #150 in the Rust generator. `name`, `about`, and `long_about` on
// the #[command(...)] attribute are Rust string literals; a raw `"` in the
// description used to break compilation.
func TestRustMainTemplate_EscapesQuotesInDescription(t *testing.T) {
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
			Description: `Agent for "PromQL" and "sum by" queries`,
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

	wantAbout := `about = "Agent for \"PromQL\" and \"sum by\" queries"`
	if !strings.Contains(rendered, wantAbout) {
		t.Errorf("main.rs did not contain properly escaped about literal\nwant substring: %s\n--- rendered ---\n%s", wantAbout, rendered)
	}

	// long_about must concatenate the description with the trailing boilerplate
	// AND keep inner quotes escaped.
	wantLongAbout := `long_about = "Agent for \"PromQL\" and \"sum by\" queries\n\nThis is an A2A (Agent-to-Agent) protocol server. Use the ` + "`start`" + ` subcommand to run it."`
	if !strings.Contains(rendered, wantLongAbout) {
		t.Errorf("main.rs did not contain properly escaped long_about literal\nwant substring: %s\n--- rendered ---\n%s", wantLongAbout, rendered)
	}

	if strings.Contains(rendered, `about = "Agent for "PromQL"`) {
		t.Errorf("main.rs still contains raw (unescaped) quotes inside the about literal\n--- rendered ---\n%s", rendered)
	}
}
