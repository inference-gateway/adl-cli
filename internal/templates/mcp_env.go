package templates

import (
	"strconv"
	"strings"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

// MCPEnvVar is a single KEY=VALUE default emitted into the generated .env.example
// for the ADK's built-in MCP client. Description is the human-readable label the
// README env-var table renders alongside it.
type MCPEnvVar struct {
	Key         string
	Value       string
	Description string
}

// mcpEnvVars maps a manifest's spec.agent.mcp block to the ADK's A2A_MCP_*
// environment variables. Each field's manifest value (or the schema default when
// omitted) becomes the emitted default; the matching env var overrides it at
// runtime. A2A_MCP_SERVERS is derived from the servers' base URLs - only the
// 'http' transport is included, since the ADK MCP client is streamable-HTTP-only
// (stdio/sse servers cannot be reached and are dropped; the validator warns).
//
// Returns nil unless MCP is enabled on a Go agent: the ADK MCP client exists only
// in the Go ADK, so TypeScript/Rust agents get no A2A_MCP_* block.
func mcpEnvVars(adl *schema.ADL) []MCPEnvVar {
	if adl == nil || adl.Spec.Agent == nil || adl.Spec.Agent.Mcp == nil || !adl.Spec.Agent.Mcp.Enabled {
		return nil
	}
	if DetectLanguageFromADL(adl) != "go" {
		return nil
	}
	mcp := adl.Spec.Agent.Mcp

	out := []MCPEnvVar{
		{Key: "A2A_MCP_ENABLED", Value: "true", Description: "Enable the MCP client"},
		{Key: "A2A_MCP_SERVERS", Value: strings.Join(mcpServerURLs(mcp), ","),
			Description: "MCP server base URLs (comma-separated)"},
	}
	out = appendMCPDefault(out, "A2A_MCP_ENDPOINT", mcp.Endpoint, "/mcp",
		"HTTP path appended to each server URL for the streamable MCP endpoint")
	out = appendMCPDefault(out, "A2A_MCP_REFRESH_INTERVAL", mcp.RefreshInterval, "5m",
		"How often to refresh the tool catalog from each server")
	out = appendMCPDefault(out, "A2A_MCP_DIAL_TIMEOUT", mcp.DialTimeout, "30s",
		"Timeout for initializing/listing tools on a server")
	out = appendMCPDefault(out, "A2A_MCP_CALL_TIMEOUT", mcp.CallTimeout, "30s",
		"Timeout for a single tool invocation")
	out = append(out, MCPEnvVar{Key: "A2A_MCP_MAX_RETRIES", Value: strconv.Itoa(mcp.MaxRetries),
		Description: "Max initial connection attempts per server (0 = retry forever)"})
	out = appendMCPDefault(out, "A2A_MCP_RETRY_INTERVAL", mcp.RetryInterval, "2s",
		"Initial backoff between retries (doubles up to the max)")
	out = appendMCPDefault(out, "A2A_MCP_RETRY_MAX_INTERVAL", mcp.RetryMaxInterval, "30s",
		"Maximum backoff between retries")
	return out
}

// mcpServerURLs returns the base URLs of the http-transport servers, in manifest
// order. stdio/sse servers and http servers without a url are skipped.
func mcpServerURLs(mcp *schema.MCP) []string {
	var urls []string
	for _, s := range mcp.Servers {
		if s.Transport != schema.MCPServerTransportHttp {
			continue
		}
		if u := strings.TrimSpace(s.URL); u != "" {
			urls = append(urls, u)
		}
	}
	return urls
}

// appendMCPDefault emits key=value, falling back to the schema default when the
// manifest omits the field.
func appendMCPDefault(out []MCPEnvVar, key, value, fallback, desc string) []MCPEnvVar {
	if value == "" {
		value = fallback
	}
	return append(out, MCPEnvVar{Key: key, Value: value, Description: desc})
}
