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
			name: "orchestrators present but empty is all-off",
			ai:   &AIConfig{Orchestrators: &OrchestratorsConfig{}},
			want: AIAgentToggles{},
		},
		{
			name: "claudecode only",
			ai: &AIConfig{
				Orchestrators: &OrchestratorsConfig{
					Claudecode: &ClaudeCodeConfig{Enabled: true},
				},
			},
			want: AIAgentToggles{ClaudeCode: true},
		},
		{
			name: "every agent enabled",
			ai: &AIConfig{
				Orchestrators: &OrchestratorsConfig{
					Claudecode: &ClaudeCodeConfig{Enabled: true},
					Codex:      &CodexConfig{Enabled: true},
					Gemini:     &GeminiConfig{Enabled: true},
					Opencode:   &OpenCodeConfig{Enabled: true},
					Infer:      &InferConfig{Enabled: true},
				},
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
				Orchestrators: &OrchestratorsConfig{
					Claudecode: &ClaudeCodeConfig{Enabled: false},
					Codex:      &CodexConfig{Enabled: false},
				},
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
	for name, toggles := range map[string]AIAgentToggles{
		"claudecode": {ClaudeCode: true},
		"codex":      {Codex: true},
		"gemini":     {Gemini: true},
		"opencode":   {OpenCode: true},
		"infer":      {Infer: true},
	} {
		if !toggles.Any() {
			t.Fatalf("%s toggle should report Any()=true", name)
		}
	}
}

func TestAIHasOfficialAction(t *testing.T) {
	cases := map[string]bool{
		"claudecode": true,
		"codex":      true,
		"gemini":     true,
		"infer":      true,
		"opencode":   false,
		"unknown":    false,
	}
	for agent, want := range cases {
		if got := AIHasOfficialAction(agent); got != want {
			t.Errorf("AIHasOfficialAction(%q) = %v, want %v", agent, got, want)
		}
	}
}
