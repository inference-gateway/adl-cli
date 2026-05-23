package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const baseManifest = `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: ai-toggle-agent
  description: Agent used to exercise the per-agent AI toggles.
  version: 1.0.0
spec:
  capabilities:
    streaming: true
  server:
    port: 8080
  language:
    go:
      module: github.com/example/ai-toggle-agent
      version: "1.26.2"
`

// writeManifest writes the canonical manifest plus an optional
// development.ai block to a temp dir and returns the manifest path.
func writeManifest(t *testing.T, dir, devAIYAML string) string {
	t.Helper()
	path := filepath.Join(dir, "agent.yaml")
	body := baseManifest + devAIYAML
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func mustGenerate(t *testing.T, manifestPath, outputDir string, config Config) {
	t.Helper()
	gen := New(config)
	if err := gen.Generate(manifestPath, outputDir); err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
}

// assertFile asserts that a path under outputDir does or doesn't exist.
func assertFile(t *testing.T, outputDir, rel string, wantExists bool) {
	t.Helper()
	full := filepath.Join(outputDir, rel)
	_, err := os.Stat(full)
	switch {
	case wantExists && err != nil:
		t.Errorf("expected %s to exist, got error: %v", rel, err)
	case !wantExists && err == nil:
		t.Errorf("expected %s NOT to exist", rel)
	}
}

func TestGenerator_AI_NoTogglesNoFiles(t *testing.T) {
	tmp := t.TempDir()
	manifest := writeManifest(t, tmp, "")
	out := filepath.Join(tmp, "out")

	mustGenerate(t, manifest, out, Config{Overwrite: true, Version: "test"})

	assertFile(t, out, "CLAUDE.md", false)
	assertFile(t, out, "AGENTS.md", false)
	assertFile(t, out, "GEMINI.md", false)
	assertFile(t, out, ".github/workflows/claude.yml", false)
	assertFile(t, out, ".github/workflows/codex.yml", false)
	assertFile(t, out, ".github/workflows/gemini.yml", false)
}

func TestGenerator_AI_ClaudeCodeOnly(t *testing.T) {
	tmp := t.TempDir()
	manifest := writeManifest(t, tmp, `  development:
    ai:
      claudecode:
        enabled: true
`)
	out := filepath.Join(tmp, "out")

	mustGenerate(t, manifest, out, Config{Overwrite: true, Version: "test"})

	assertFile(t, out, "CLAUDE.md", true)
	assertFile(t, out, ".github/workflows/claude.yml", true)

	assertFile(t, out, "AGENTS.md", false)
	assertFile(t, out, "GEMINI.md", false)
	assertFile(t, out, ".github/workflows/codex.yml", false)
	assertFile(t, out, ".github/workflows/gemini.yml", false)
}

func TestGenerator_AI_GeminiOnly(t *testing.T) {
	tmp := t.TempDir()
	manifest := writeManifest(t, tmp, `  development:
    ai:
      gemini:
        enabled: true
`)
	out := filepath.Join(tmp, "out")

	mustGenerate(t, manifest, out, Config{Overwrite: true, Version: "test"})

	assertFile(t, out, "GEMINI.md", true)
	assertFile(t, out, ".github/workflows/gemini.yml", true)

	assertFile(t, out, "CLAUDE.md", false)
	assertFile(t, out, "AGENTS.md", false)
	assertFile(t, out, ".github/workflows/claude.yml", false)
	assertFile(t, out, ".github/workflows/codex.yml", false)
}

func TestGenerator_AI_AgentsMDSharedAcrossCodexOpencodeInfer(t *testing.T) {
	tests := []struct {
		name              string
		devAI             string
		wantCodexWorkflow bool
	}{
		{
			name: "codex only",
			devAI: `  development:
    ai:
      codex:
        enabled: true
`,
			wantCodexWorkflow: true,
		},
		{
			name: "opencode only (no workflow)",
			devAI: `  development:
    ai:
      opencode:
        enabled: true
`,
		},
		{
			name: "infer only (no workflow yet)",
			devAI: `  development:
    ai:
      infer:
        enabled: true
`,
		},
		{
			name: "codex + opencode + infer share a single AGENTS.md",
			devAI: `  development:
    ai:
      codex:
        enabled: true
      opencode:
        enabled: true
      infer:
        enabled: true
`,
			wantCodexWorkflow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			manifest := writeManifest(t, tmp, tt.devAI)
			out := filepath.Join(tmp, "out")

			mustGenerate(t, manifest, out, Config{Overwrite: true, Version: "test"})

			assertFile(t, out, "AGENTS.md", true)
			assertFile(t, out, "CLAUDE.md", false)
			assertFile(t, out, "GEMINI.md", false)
			assertFile(t, out, ".github/workflows/codex.yml", tt.wantCodexWorkflow)
			assertFile(t, out, ".github/workflows/claude.yml", false)
			assertFile(t, out, ".github/workflows/gemini.yml", false)
		})
	}
}

func TestGenerator_AI_AllAgentsEnabled(t *testing.T) {
	tmp := t.TempDir()
	manifest := writeManifest(t, tmp, `  development:
    ai:
      claudecode:
        enabled: true
      codex:
        enabled: true
      gemini:
        enabled: true
      opencode:
        enabled: true
      infer:
        enabled: true
`)
	out := filepath.Join(tmp, "out")

	mustGenerate(t, manifest, out, Config{Overwrite: true, Version: "test"})

	assertFile(t, out, "CLAUDE.md", true)
	assertFile(t, out, "GEMINI.md", true)
	assertFile(t, out, "AGENTS.md", true)

	assertFile(t, out, ".github/workflows/claude.yml", true)
	assertFile(t, out, ".github/workflows/codex.yml", true)
	assertFile(t, out, ".github/workflows/gemini.yml", true)
}

func TestGenerator_AI_NoWorkflowsWhenSCMNotGithub(t *testing.T) {
	tmp := t.TempDir()
	manifest := writeManifest(t, tmp, `  scm:
    provider: gitlab
  development:
    ai:
      claudecode:
        enabled: true
`)
	out := filepath.Join(tmp, "out")

	mustGenerate(t, manifest, out, Config{Overwrite: true, Version: "test"})

	assertFile(t, out, "CLAUDE.md", true)
	assertFile(t, out, ".github/workflows/claude.yml", false)
}

// readGenerated reads a generated file under outputDir, failing the test
// if the file is missing.
func readGenerated(t *testing.T, outputDir, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(outputDir, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

// assertContains asserts that haystack contains needle, with a helpful
// message that names what was being checked.
func assertContains(t *testing.T, haystack, needle, what string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected %s to contain %q", what, needle)
	}
}

// assertNotContains is the inverse of assertContains.
func assertNotContains(t *testing.T, haystack, needle, what string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("expected %s NOT to contain %q", what, needle)
	}
}

func TestGenerator_AI_ClaudeWorkflowGoContent(t *testing.T) {
	tmp := t.TempDir()
	manifest := writeManifest(t, tmp, `  development:
    ai:
      claudecode:
        enabled: true
`)
	out := filepath.Join(tmp, "out")

	mustGenerate(t, manifest, out, Config{Overwrite: true, Version: "test"})

	body := readGenerated(t, out, ".github/workflows/claude.yml")

	assertNotContains(t, body, "claude-code.yml", "Claude Code workflow body")

	assertNotContains(t, body, "picks up CLAUDE.md", "Claude Code workflow body")

	assertContains(t, body, "issue_comment:\n    types:\n      - created", "Claude Code workflow body")
	assertContains(t, body, "issues:\n    types:\n      - opened\n      - assigned", "Claude Code workflow body")
	assertNotContains(t, body, "types: [created]", "Claude Code workflow body")

	assertContains(t, body, "actions: read", "Claude Code workflow body")

	assertContains(t, body, "Set up Go", "Claude Code workflow body (go)")
	assertContains(t, body, "actions/setup-go@v6.4.0", "Claude Code workflow body (go)")
	assertContains(t, body, "Install golangci-lint", "Claude Code workflow body (go)")
	assertNotContains(t, body, "Set up Rust", "Claude Code workflow body (go)")
	assertNotContains(t, body, "actions-rs/toolchain", "Claude Code workflow body (go)")

	assertContains(t, body, "arduino/setup-task@v2.0.0", "Claude Code workflow body")
	assertContains(t, body, "Install ADL skill", "Claude Code workflow body")
	assertContains(t, body, "raw.githubusercontent.com/inference-gateway/skills/main/skills/adl/SKILL.md", "Claude Code workflow body")

	assertContains(t, body, "anthropics/claude-code-action@v1.0.131", "Claude Code workflow body")
	assertContains(t, body, "claude_code_oauth_token:", "Claude Code workflow body")
	assertContains(t, body, "use_commit_signing: true", "Claude Code workflow body")
	assertContains(t, body, "branch_prefix: 'claude/'", "Claude Code workflow body")
	assertContains(t, body, "--model claude-opus-4-7", "Claude Code workflow body")

	assertContains(t, body, "Bash(go:*)", "Claude Code workflow body (go)")
	assertNotContains(t, body, "Bash(cargo:*)", "Claude Code workflow body (go)")
}

func TestGenerator_AI_ClaudeWorkflowRustContent(t *testing.T) {
	tmp := t.TempDir()
	manifest := filepath.Join(tmp, "agent.yaml")
	body := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: ai-toggle-agent
  description: Agent used to exercise the per-agent AI toggles.
  version: 1.0.0
spec:
  capabilities:
    streaming: true
  server:
    port: 8080
  language:
    rust:
      packageName: ai-toggle-agent
      version: "1.94.1"
      edition: "2024"
  development:
    ai:
      claudecode:
        enabled: true
`
	if err := os.WriteFile(manifest, []byte(body), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out := filepath.Join(tmp, "out")

	mustGenerate(t, manifest, out, Config{Overwrite: true, Version: "test"})

	wf := readGenerated(t, out, ".github/workflows/claude.yml")

	assertContains(t, wf, "Set up Rust", "Claude Code workflow body (rust)")
	assertContains(t, wf, "actions-rs/toolchain@v1", "Claude Code workflow body (rust)")
	assertContains(t, wf, "1.94.1", "Claude Code workflow body (rust)")
	assertNotContains(t, wf, "Set up Go", "Claude Code workflow body (rust)")
	assertNotContains(t, wf, "actions/setup-go", "Claude Code workflow body (rust)")
	assertNotContains(t, wf, "Install golangci-lint", "Claude Code workflow body (rust)")

	assertContains(t, wf, "Bash(cargo:*)", "Claude Code workflow body (rust)")
	assertNotContains(t, wf, "Bash(go:*)", "Claude Code workflow body (rust)")
}
