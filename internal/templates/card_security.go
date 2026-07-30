package templates

import "github.com/inference-gateway/adl-cli/internal/schema"

// cardSecuritySchemes maps the manifest's flat security-scheme DSL
// (spec.card.securitySchemes) onto the ADK AgentCard wrapper shape that the
// generated .well-known/agent-card.json is unmarshalled into: each scheme is
// nested under a per-variant key ('type' selects the wrapper, 'in' becomes
// 'location'). Returns nil when none are declared so the template omits the
// field. OIDC/OAuth2 are intentionally not modelled here - the ADK derives them
// from runtime config (see cardSecurity and main.go's OIDC wiring).
func cardSecuritySchemes(card *schema.Card) map[string]any {
	if card == nil || len(card.SecuritySchemes) == 0 {
		return nil
	}
	out := make(map[string]any, len(card.SecuritySchemes))
	for name, s := range card.SecuritySchemes {
		out[name] = wrapSecurityScheme(s)
	}
	return out
}

// wrapSecurityScheme converts one flat DSL scheme into its ADK per-variant
// wrapper object.
func wrapSecurityScheme(s schema.SecurityScheme) map[string]any {
	inner := map[string]any{}
	if s.Description != "" {
		inner["description"] = s.Description
	}
	switch s.Type {
	case schema.SecuritySchemeTypeAPIKey:
		inner["name"] = s.Name
		inner["location"] = string(s.In)
		return map[string]any{"apiKeySecurityScheme": inner}
	case schema.SecuritySchemeTypeHttp:
		inner["scheme"] = s.Scheme
		if s.BearerFormat != "" {
			inner["bearerFormat"] = s.BearerFormat
		}
		return map[string]any{"httpAuthSecurityScheme": inner}
	case schema.SecuritySchemeTypeMutualTLS:
		return map[string]any{"mtlsSecurityScheme": inner}
	}
	return map[string]any{}
}

// cardSecurity maps the manifest's flat security requirements
// (spec.card.security) onto the ADK AgentCard 'security' shape:
// [{ schemes: { <name>: { list: [scopes] } } }]. Returns nil when none are
// declared.
func cardSecurity(card *schema.Card) []any {
	if card == nil || len(card.Security) == 0 {
		return nil
	}
	out := make([]any, 0, len(card.Security))
	for _, req := range card.Security {
		schemes := make(map[string]any, len(req))
		for name, scopes := range req {
			if scopes == nil {
				scopes = []string{}
			}
			schemes[name] = map[string]any{"list": scopes}
		}
		out = append(out, map[string]any{"schemes": schemes})
	}
	return out
}
