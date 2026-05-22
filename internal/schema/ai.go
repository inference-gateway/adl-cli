package schema

// AIAgentToggles reports which per-agent AI assistant toggles in
// spec.development.ai are enabled. The fields mirror AIConfig's
// per-agent subsections (claudecode, codex, gemini, opencode, infer)
// introduced in ADL v0.8.0.
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
// is shared by codex, opencode, and infer — generate it once if any of
// those three are enabled.
func (t AIAgentToggles) AnyAgentsMD() bool {
	return t.Codex || t.OpenCode || t.Infer
}

// ResolveAIAgentToggles inspects spec.development.ai (the v0.8.0 shape)
// and returns the effective set of per-agent toggles. The pre-v0.8.0
// single-flag `spec.development.ai.enabled` shape is rejected by the
// validator (see checkLegacySpecFields), so this function only needs to
// honour the modern per-agent shape.
func ResolveAIAgentToggles(ai *AIConfig) AIAgentToggles {
	var t AIAgentToggles
	if ai == nil {
		return t
	}
	if ai.Claudecode != nil && ai.Claudecode.Enabled {
		t.ClaudeCode = true
	}
	if ai.Codex != nil && ai.Codex.Enabled {
		t.Codex = true
	}
	if ai.Gemini != nil && ai.Gemini.Enabled {
		t.Gemini = true
	}
	if ai.Opencode != nil && ai.Opencode.Enabled {
		t.OpenCode = true
	}
	if ai.Infer != nil && ai.Infer.Enabled {
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
