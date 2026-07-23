package templates

import (
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// mcpEnvMap flattens the emitted MCP vars into a name -> value lookup.
func mcpEnvMap(vars []MCPEnvVar) map[string]string {
	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}
	return m
}

// TestMCPEnvVars_DerivesServersAndDefaults pins the core contract: A2A_MCP_SERVERS
// is the comma-joined base URLs of the http servers (stdio/sse dropped), and an
// omitted optional field falls back to its schema default.
func TestMCPEnvVars_DerivesServersAndDefaults(t *testing.T) {
	adl := minimalGoADL()
	adl.Spec.Agent = &schema.Agent{
		Mcp: &schema.MCP{
			Enabled:     true,
			CallTimeout: "45s",
			Servers: []schema.MCPServer{
				{Name: "tools", Transport: schema.MCPServerTransportHttp, URL: "http://mcp-tools:8080"},
				{Name: "local", Transport: schema.MCPServerTransportStdio, Command: "npx"},
				{Name: "search", Transport: schema.MCPServerTransportHttp, URL: "http://mcp-search:8080"},
				{Name: "legacy", Transport: schema.MCPServerTransportSse, URL: "http://sse:8080"},
			},
		},
	}

	got := mcpEnvMap(mcpEnvVars(adl))

	want := map[string]string{
		"A2A_MCP_ENABLE":             "true",
		"A2A_MCP_SERVERS":            "http://mcp-tools:8080,http://mcp-search:8080",
		"A2A_MCP_ENDPOINT":           "/mcp",
		"A2A_MCP_REFRESH_INTERVAL":   "5m",
		"A2A_MCP_DIAL_TIMEOUT":       "30s",
		"A2A_MCP_CALL_TIMEOUT":       "45s",
		"A2A_MCP_MAX_RETRIES":        "0",
		"A2A_MCP_RETRY_INTERVAL":     "2s",
		"A2A_MCP_RETRY_MAX_INTERVAL": "30s",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("MCP env %s = %q, want %q (all: %v)", k, got[k], v, got)
		}
	}
}

// TestMCPEnvVars_DisabledOrNonGo asserts nothing is emitted when MCP is disabled,
// absent, or the agent is not Go (the ADK MCP client is Go-only).
func TestMCPEnvVars_DisabledOrNonGo(t *testing.T) {
	disabled := minimalGoADL()
	disabled.Spec.Agent = &schema.Agent{Mcp: &schema.MCP{Enabled: false}}
	if got := mcpEnvVars(disabled); got != nil {
		t.Errorf("disabled MCP emitted %v, want nil", got)
	}

	absent := minimalGoADL()
	absent.Spec.Agent = &schema.Agent{}
	if got := mcpEnvVars(absent); got != nil {
		t.Errorf("absent MCP emitted %v, want nil", got)
	}

	ts := minimalTypeScriptADL()
	ts.Spec.Agent = &schema.Agent{Mcp: &schema.MCP{
		Enabled: true,
		Servers: []schema.MCPServer{{Name: "tools", Transport: schema.MCPServerTransportHttp, URL: "http://mcp:8080"}},
	}}
	if got := mcpEnvVars(ts); got != nil {
		t.Errorf("TypeScript MCP emitted %v, want nil", got)
	}
}
