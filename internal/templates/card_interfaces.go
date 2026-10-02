package templates

import schema "github.com/inference-gateway/adl-cli/internal/schema"

// cardSupportedInterfaces renders spec.card.supportedInterfaces for the emitted
// AgentCard. The A2A v1.0.1 card requires at least one interface, so a manifest
// that declares none falls back to a single JSONRPC entry whose URL the generated
// agent overrides at startup from A2A_AGENT_URL.
func cardSupportedInterfaces(card *schema.Card) []any {
	if card == nil || len(card.SupportedInterfaces) == 0 {
		return []any{map[string]any{
			"url":             "",
			"protocolBinding": schema.DefaultProtocolBinding,
			"protocolVersion": schema.DefaultProtocolVersion,
		}}
	}

	out := make([]any, 0, len(card.SupportedInterfaces))
	for _, iface := range card.SupportedInterfaces {
		out = append(out, map[string]any{
			"url":             iface.URL,
			"protocolBinding": iface.ProtocolBinding,
			"protocolVersion": iface.ProtocolVersion,
		})
	}
	return out
}
