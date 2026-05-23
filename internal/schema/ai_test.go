package schema

import "testing"

func TestResolveAIAgentToggles(t *testing.T) {
	tests := []struct {
		name string
		ai   *AIConfig
		want AIAgentToggles
	}{
		{
			name: "nil AIConfig is all-off",
			ai:   nil,
			want: AIAgentToggles{},
		},
		{
			name: "claudecode only",
			ai: &AIConfig{
				Claudecode: &ClaudeCodeConfig{Enabled: true},
			},
			want: AIAgentToggles{ClaudeCode: true},
		},
		{
			name: "every agent enabled",
			ai: &AIConfig{
				Claudecode: &ClaudeCodeConfig{Enabled: true},
				Codex:      &CodexConfig{Enabled: true},
				Gemini:     &GeminiConfig{Enabled: true},
				Opencode:   &OpenCodeConfig{Enabled: true},
				Infer:      &InferConfig{Enabled: true},
			},
			want: AIAgentToggles{
				ClaudeCode: true,
				Codex:      true,
				Gemini:     true,
				OpenCode:   true,
				Infer:      true,
			},
		},
		{
			name: "agent block present but disabled is treated as off",
			ai: &AIConfig{
				Claudecode: &ClaudeCodeConfig{Enabled: false},
				Codex:      &CodexConfig{Enabled: false},
			},
			want: AIAgentToggles{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveAIAgentToggles(tt.ai)
			if got != tt.want {
				t.Fatalf("ResolveAIAgentToggles() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestAIAgentToggles_Any(t *testing.T) {
	if (AIAgentToggles{}).Any() {
		t.Fatal("empty toggles should report Any()=false")
	}
	if !(AIAgentToggles{OpenCode: true}).Any() {
		t.Fatal("opencode toggle should report Any()=true")
	}
}

func TestAIAgentToggles_AnyAgentsMD(t *testing.T) {
	tests := map[string]struct {
		toggles AIAgentToggles
		want    bool
	}{
		"claudecode alone is not enough":       {AIAgentToggles{ClaudeCode: true}, false},
		"gemini alone is not enough":           {AIAgentToggles{Gemini: true}, false},
		"codex flips it on":                    {AIAgentToggles{Codex: true}, true},
		"opencode flips it on":                 {AIAgentToggles{OpenCode: true}, true},
		"infer flips it on":                    {AIAgentToggles{Infer: true}, true},
		"claudecode + infer combo flips it on": {AIAgentToggles{ClaudeCode: true, Infer: true}, true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.toggles.AnyAgentsMD(); got != tc.want {
				t.Fatalf("AnyAgentsMD() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAIHasOfficialAction(t *testing.T) {
	cases := map[string]bool{
		"claudecode": true,
		"codex":      true,
		"gemini":     true,
		"opencode":   false,
		"infer":      false,
		"unknown":    false,
	}
	for agent, want := range cases {
		if got := AIHasOfficialAction(agent); got != want {
			t.Errorf("AIHasOfficialAction(%q) = %v, want %v", agent, got, want)
		}
	}
}
