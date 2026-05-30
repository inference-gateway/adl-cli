package schema

// AIAgentToggles reports which per-agent AI assistant toggles in
// spec.development.ai.orchestrators are enabled. The fields mirror
// OrchestratorsConfig's per-agent subsections (claudecode, codex,
// gemini, opencode, infer), nested under ai.orchestrators since the
// ADL schema's orchestrators refactor.
type AIAgentToggles struct {
	ClaudeCode bool
	Codex      bool
	Gemini     bool
	OpenCode   bool
	Infer      bool
}

// Any reports whether at least one agent toggle is enabled.
func (t AIAgentToggles) Any() bool {
	return t.ClaudeCode || t.Codex || t.Gemini || t.OpenCode || t.Infer
}

// AnyAgentsMD reports whether AGENTS.md should be generated. AGENTS.md
// is shared by codex, opencode, and infer - generate it once if any of
// those three are enabled.
func (t AIAgentToggles) AnyAgentsMD() bool {
	return t.Codex || t.OpenCode || t.Infer
}

// ResolveAIAgentToggles inspects spec.development.ai.orchestrators and
// returns the effective set of per-agent toggles. The legacy flat shapes
// (`spec.development.ai.enabled` and `spec.development.ai.<agent>`) are
// rejected by the validator (see checkLegacySpecFields), so this function
// only needs to honour the nested orchestrators shape.
func ResolveAIAgentToggles(ai *AIConfig) AIAgentToggles {
	var t AIAgentToggles
	if ai == nil || ai.Orchestrators == nil {
		return t
	}
	o := ai.Orchestrators
	if o.Claudecode != nil && o.Claudecode.Enabled {
		t.ClaudeCode = true
	}
	if o.Codex != nil && o.Codex.Enabled {
		t.Codex = true
	}
	if o.Gemini != nil && o.Gemini.Enabled {
		t.Gemini = true
	}
	if o.Opencode != nil && o.Opencode.Enabled {
		t.OpenCode = true
	}
	if o.Infer != nil && o.Infer.Enabled {
		t.Infer = true
	}
	return t
}

// AIHasOfficialAction reports whether the generator should emit a
// per-agent GitHub Actions workflow under .github/workflows/ for the
// given agent. Only claudecode, codex, and gemini ship a workflow
// today; opencode has no upstream action and the infer integration is
// still being scoped (see inference-gateway/adl-cli#142 for the
// acceptance matrix).
func AIHasOfficialAction(agent string) bool {
	switch agent {
	case "claudecode", "codex", "gemini":
		return true
	default:
		return false
	}
}
