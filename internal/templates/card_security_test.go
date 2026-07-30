package templates

import (
	"encoding/json"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// TestCardSecuritySchemes_MapsFlatDSLToADKWrappers pins the mapping from the
// flat authoring DSL to the ADK AgentCard wrapper shape: 'type' selects the
// per-variant key and apiKey's 'in' becomes 'location'.
func TestCardSecuritySchemes_MapsFlatDSLToADKWrappers(t *testing.T) {
	card := &schema.Card{
		SecuritySchemes: schema.CardSecuritySchemes{
			"apiKey": {Type: schema.SecuritySchemeTypeAPIKey, Name: "X-API-Key", In: schema.SecuritySchemeInHeader, Description: "key"},
			"bearer": {Type: schema.SecuritySchemeTypeHttp, Scheme: "Bearer", BearerFormat: "JWT"},
			"mtls":   {Type: schema.SecuritySchemeTypeMutualTLS},
		},
	}

	got := roundTrip(t, cardSecuritySchemes(card))
	want := map[string]any{
		"apiKey": map[string]any{"apiKeySecurityScheme": map[string]any{"name": "X-API-Key", "location": "header", "description": "key"}},
		"bearer": map[string]any{"httpAuthSecurityScheme": map[string]any{"scheme": "Bearer", "bearerFormat": "JWT"}},
		"mtls":   map[string]any{"mtlsSecurityScheme": map[string]any{}},
	}
	assertJSONEqual(t, want, got)

	if cardSecuritySchemes(&schema.Card{}) != nil {
		t.Fatal("expected nil when no schemes declared")
	}
}

// TestCardSecurity_MapsRequirementsToSchemesList pins the 'security' mapping onto
// the ADK { schemes: { <name>: { list: [scopes] } } } shape, empty scopes -> [].
func TestCardSecurity_MapsRequirementsToSchemesList(t *testing.T) {
	card := &schema.Card{
		Security: []schema.CardSecurityElem{
			{"bearer": {"read", "write"}},
			{"apiKey": nil},
		},
	}

	got := roundTrip(t, cardSecurity(card))
	want := []any{
		map[string]any{"schemes": map[string]any{"bearer": map[string]any{"list": []any{"read", "write"}}}},
		map[string]any{"schemes": map[string]any{"apiKey": map[string]any{"list": []any{}}}},
	}
	assertJSONEqual(t, want, got)

	if cardSecurity(&schema.Card{}) != nil {
		t.Fatal("expected nil when no security declared")
	}
}

// roundTrip marshals through JSON so the assertion compares the exact wire shape
// the template renders via toJson.
func roundTrip(t *testing.T, v any) any {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return out
}

func assertJSONEqual(t *testing.T, want, got any) {
	t.Helper()
	w, _ := json.Marshal(want)
	g, _ := json.Marshal(got)
	// Re-marshal want through the same normalization so map key order matches.
	var wn any
	_ = json.Unmarshal(w, &wn)
	wnb, _ := json.Marshal(wn)
	if string(wnb) != string(g) {
		t.Fatalf("mismatch:\nwant %s\ngot  %s", wnb, g)
	}
}
