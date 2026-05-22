package schema

// AIAgentToggles reports which per-agent AI assistant toggles in
// spec.development.ai are enabled. The fields mirror AIConfig's
// per-agent subsections (claudecode, codex, gemini, opencode, infer).
//
// AIConfig in ADL v0.8.0 replaces the single `enabled` flag with
// independent per-agent toggles. ResolveAIAgentToggles centralises that
// logic so the generator and templates don't each re-implement the
// branching, and bakes in the legacy `spec.development.ai.enabled`
// fallback (pre-v0.8.0 manifests still validate because the schema
// allows additional properties — see ResolveAIAgentToggles).
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
// and an optional legacyEnabled hint to produce the effective set of
// per-agent toggles.
//
// legacyEnabled captures two pre-v0.8.0 sources that flowed through a
// single `enabled` flag:
//
//   - The deprecated spec.development.ai.enabled field. The schema in
//     v0.8.0 dropped this property, but it's still permitted by JSON
//     Schema's default additionalProperties:true so old manifests
//     continue to parse. The caller is responsible for surfacing it
//     (the generator reads the raw YAML up front).
//   - The CLI's --ai flag, which historically toggled
//     CLAUDE.md + AGENTS.md generation.
//
// When legacyEnabled is true and no per-agent toggle is set, this
// resolves to claudecode + infer so the generator emits CLAUDE.md and
// AGENTS.md — the exact set of files the pre-v0.8.0 path produced. If
// any per-agent toggle is already set the legacy flag is ignored: an
// explicit modern manifest always wins.
func ResolveAIAgentToggles(ai *AIConfig, legacyEnabled bool) AIAgentToggles {
	var t AIAgentToggles
	if ai != nil {
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
	}

	if !t.Any() && legacyEnabled {
		t.ClaudeCode = true
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
