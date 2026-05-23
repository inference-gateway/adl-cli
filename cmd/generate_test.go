package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCommand(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: test-agent
  description: Test agent
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/agent
      version: "1.26.2"
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
	}()

	adlFile = adlPath
	outputDir = outputPath

	err := runGenerate(generateCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("expected output directory to be created")
	}

	mainGoPath := filepath.Join(outputPath, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Errorf("expected main.go to be generated")
	}

	goModPath := filepath.Join(outputPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Errorf("expected go.mod to be generated")
	}
}

func TestGenerateWithoutInit(t *testing.T) {
	tempDir := t.TempDir()

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: standalone-agent
  description: Standalone test agent
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  agent:
    provider: deepseek
    model: deepseek-v4-flash
  config:
    tools:
      read:
        enabled: true
  tools:
    - id: read
    - id: test_tool_id
      name: test_tool
      description: A test tool
      tags:
        - test
      schema:
        type: object
        properties:
          input:
            type: string
            description: Test input
        required:
          - input
  skills:
    - id: test-skill
      bare: true
      name: test-skill
      description: A test bare skill
      tags:
        - test
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/standalone
      version: "1.26.2"
`

	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	outputPath := filepath.Join(tempDir, "output")

	originalADLFile := adlFile
	originalOutputDir := outputDir
	originalOffline := offlineMode
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
		offlineMode = originalOffline
	}()

	adlFile = adlPath
	outputDir = outputPath
	offlineMode = true

	err := runGenerate(generateCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mainGoPath := filepath.Join(outputPath, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Errorf("expected main.go to be generated")
	}

	toolsDir := filepath.Join(outputPath, "tools")
	if _, err := os.Stat(toolsDir); os.IsNotExist(err) {
		t.Errorf("expected tools directory to be generated")
	}

	testToolPath := filepath.Join(toolsDir, "test_tool_id.go")
	if _, err := os.Stat(testToolPath); os.IsNotExist(err) {
		t.Errorf("expected test_tool_id.go to be generated")
	}

	testSkillPath := filepath.Join(outputPath, "skills", "test-skill", "SKILL.md")
	if _, err := os.Stat(testSkillPath); os.IsNotExist(err) {
		t.Errorf("expected skills/test-skill/SKILL.md to be scaffolded")
	}

	dockerfilePath := filepath.Join(outputPath, "Dockerfile")
	dockerfileBytes, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("failed to read generated Dockerfile: %v", err)
	}
	dockerfileContent := string(dockerfileBytes)
	if !strings.Contains(dockerfileContent, "COPY --from=builder /app/skills ./skills") {
		t.Errorf("expected Dockerfile to COPY skills/ when spec.skills is non-empty, got:\n%s", dockerfileContent)
	}
}

func TestGenerateDockerfileOmitsSkillsCopyWhenAbsent(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "no-skills-output")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: no-skills-agent
  description: Agent without skills
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/no-skills
      version: "1.26.2"
`

	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
	}()

	adlFile = adlPath
	outputDir = outputPath

	if err := runGenerate(generateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dockerfilePath := filepath.Join(outputPath, "Dockerfile")
	dockerfileBytes, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("failed to read generated Dockerfile: %v", err)
	}
	dockerfileContent := string(dockerfileBytes)
	if strings.Contains(dockerfileContent, "/app/skills") {
		t.Errorf("expected Dockerfile to omit skills COPY when spec.skills is empty, got:\n%s", dockerfileContent)
	}
}

func TestGenerateWithCD(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-cd-output")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: test-cd-agent
  description: Test CD agent
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/cd-agent
      version: "1.26.2"
  scm:
    provider: github
    url: https://github.com/test/cd-agent
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	originalGenerateCD := generateCD
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
		generateCD = originalGenerateCD
	}()

	adlFile = adlPath
	outputDir = outputPath
	generateCD = true

	err := runGenerate(generateCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mainGoPath := filepath.Join(outputPath, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Errorf("expected main.go to be generated")
	}

	releasercPath := filepath.Join(outputPath, ".releaserc.yaml")
	if _, err := os.Stat(releasercPath); os.IsNotExist(err) {
		t.Errorf("expected .releaserc.yaml to be generated")
	}

	cdWorkflowPath := filepath.Join(outputPath, ".github/workflows/cd.yml")
	if _, err := os.Stat(cdWorkflowPath); os.IsNotExist(err) {
		t.Errorf("expected .github/workflows/cd.yml to be generated")
	}

	releasercContent, err := os.ReadFile(releasercPath)
	if err != nil {
		t.Fatalf("failed to read .releaserc.yaml: %v", err)
	}
	if !containsString(string(releasercContent), "https://github.com/test/cd-agent") {
		t.Errorf("expected .releaserc.yaml to contain repository URL")
	}
	if !containsString(string(releasercContent), "@semantic-release/github") {
		t.Errorf("expected .releaserc.yaml to contain semantic-release plugins")
	}

	cdContent, err := os.ReadFile(cdWorkflowPath)
	if err != nil {
		t.Fatalf("failed to read CD workflow: %v", err)
	}
	if !containsString(string(cdContent), "workflow_dispatch") {
		t.Errorf("expected CD workflow to contain workflow_dispatch trigger")
	}
	if !containsString(string(cdContent), "ghcr.io") {
		t.Errorf("expected CD workflow to contain GitHub Container Registry")
	}
	if !containsString(string(cdContent), "semantic-release") {
		t.Errorf("expected CD workflow to contain semantic-release")
	}
}

// TestGenerateWithAIFromManifest verifies that declaring per-agent AI toggles
// in the manifest activates the matching docs + sandbox extensions. No CLI
// flag is involved - AI assistants are entirely manifest-driven post-v0.8.0.
func TestGenerateWithAIFromManifest(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-ai-output")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: test-ai-agent
  description: Test AI agent with sandbox environments
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/ai-agent
      version: "1.26.2"
  development:
    sandbox:
      flox:
        enabled: true
      devcontainer:
        enabled: true
    ai:
      claudecode:
        enabled: true
      codex:
        enabled: true
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
	}()

	adlFile = adlPath
	outputDir = outputPath

	if err := runGenerate(generateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claudeMdPath := filepath.Join(outputPath, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Errorf("expected CLAUDE.md to be generated when spec.development.ai.claudecode.enabled is true")
	}

	claudeMdContent, err := os.ReadFile(claudeMdPath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}
	if !containsString(string(claudeMdContent), "test-ai-agent") {
		t.Errorf("expected CLAUDE.md to contain agent name")
	}
	if !containsString(string(claudeMdContent), "Test AI agent with sandbox environments") {
		t.Errorf("expected CLAUDE.md to contain agent description")
	}

	devcontainerPath := filepath.Join(outputPath, ".devcontainer/devcontainer.json")
	if _, err := os.Stat(devcontainerPath); os.IsNotExist(err) {
		t.Errorf("expected .devcontainer/devcontainer.json to be generated")
	}

	devcontainerContent, err := os.ReadFile(devcontainerPath)
	if err != nil {
		t.Fatalf("failed to read devcontainer.json: %v", err)
	}
	if !containsString(string(devcontainerContent), "anthropic.claude-code") {
		t.Errorf("expected devcontainer.json to contain claude-code extension when claudecode is enabled")
	}

	floxManifestPath := filepath.Join(outputPath, ".flox/env/manifest.toml")
	if _, err := os.Stat(floxManifestPath); os.IsNotExist(err) {
		t.Errorf("expected .flox/env/manifest.toml to be generated")
	}

	floxContent, err := os.ReadFile(floxManifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest.toml: %v", err)
	}
	if !containsString(string(floxContent), "claude-code.pkg-path") {
		t.Errorf("expected manifest.toml to contain claude-code package when claudecode is enabled")
	}

	gitattributesPath := filepath.Join(outputPath, ".gitattributes")
	gitattributesBytes, err := os.ReadFile(gitattributesPath)
	if err != nil {
		t.Fatalf("failed to read .gitattributes: %v", err)
	}
	gitattributesContent := string(gitattributesBytes)
	if !containsString(gitattributesContent, "CLAUDE.md linguist-generated=true") {
		t.Errorf("expected .gitattributes to mark CLAUDE.md as linguist-generated, got:\n%s", gitattributesContent)
	}
	if !containsString(gitattributesContent, "AGENTS.md linguist-generated=true") {
		t.Errorf("expected .gitattributes to mark AGENTS.md as linguist-generated, got:\n%s", gitattributesContent)
	}
}

// TestGenerateRejectsLegacyAIEnabled locks in the v0.8.0 migration:
// `spec.development.ai.enabled: true` must be rejected by the validator
// with a hint pointing users at the per-agent toggles.
func TestGenerateRejectsLegacyAIEnabled(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "legacy-ai-enabled")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: legacy-ai-agent
  description: Pre-v0.8.0 manifest using the single-flag AI shape
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/legacy-ai
      version: "1.26.2"
  development:
    ai:
      enabled: true
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
	}()

	adlFile = adlPath
	outputDir = outputPath

	err := runGenerate(generateCmd, []string{})
	if err == nil {
		t.Fatalf("expected runGenerate to fail when spec.development.ai.enabled is set")
	}
	if !containsString(err.Error(), "spec.development.ai.enabled") {
		t.Errorf("expected error to mention spec.development.ai.enabled, got: %v", err)
	}
	if !containsString(err.Error(), "claudecode") {
		t.Errorf("expected error to point at the per-agent toggles (e.g. claudecode), got: %v", err)
	}
}

// TestGenerateWithCIFromManifest verifies that declaring spec.scm.ci: true
// in the manifest activates CI workflow generation without needing --ci.
func TestGenerateWithCIFromManifest(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "ci-from-manifest")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: manifest-ci-agent
  description: CI enabled declaratively
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/manifest-ci
      version: "1.26.2"
  scm:
    provider: github
    url: https://github.com/test/manifest-ci
    ci: true
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	originalGenerateCI := generateCI
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
		generateCI = originalGenerateCI
	}()

	adlFile = adlPath
	outputDir = outputPath
	generateCI = false

	if err := runGenerate(generateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ciWorkflowPath := filepath.Join(outputPath, ".github/workflows/ci.yml")
	if _, err := os.Stat(ciWorkflowPath); os.IsNotExist(err) {
		t.Errorf("expected .github/workflows/ci.yml to be generated when spec.scm.ci is true (without --ci flag)")
	}
}

// TestGenerateWithCDFromManifest verifies that declaring spec.scm.cd: true
// in the manifest activates CD pipeline generation without needing --cd.
func TestGenerateWithCDFromManifest(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "cd-from-manifest")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: manifest-cd-agent
  description: CD enabled declaratively
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/manifest-cd
      version: "1.26.2"
  scm:
    provider: github
    url: https://github.com/test/manifest-cd
    cd: true
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	originalGenerateCD := generateCD
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
		generateCD = originalGenerateCD
	}()

	adlFile = adlPath
	outputDir = outputPath
	generateCD = false

	if err := runGenerate(generateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cdWorkflowPath := filepath.Join(outputPath, ".github/workflows/cd.yml")
	if _, err := os.Stat(cdWorkflowPath); os.IsNotExist(err) {
		t.Errorf("expected .github/workflows/cd.yml to be generated when spec.scm.cd is true (without --cd flag)")
	}

	releasercPath := filepath.Join(outputPath, ".releaserc.yaml")
	if _, err := os.Stat(releasercPath); os.IsNotExist(err) {
		t.Errorf("expected .releaserc.yaml to be generated when spec.scm.cd is true (without --cd flag)")
	}
}

// TestGenerateCLIFlagOverridesManifest verifies that the --ci/--cd flags
// OR on top of the manifest value: setting the flag wins even if the manifest
func TestGenerateCLIFlagOverridesManifest(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "cli-overrides")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: cli-overrides-agent
  description: Manifest leaves CI/CD off but flags turn them on
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/cli-overrides
      version: "1.26.2"
  scm:
    provider: github
    url: https://github.com/test/cli-overrides
    ci: false
    cd: false
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	originalGenerateCI := generateCI
	originalGenerateCD := generateCD
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
		generateCI = originalGenerateCI
		generateCD = originalGenerateCD
	}()

	adlFile = adlPath
	outputDir = outputPath
	generateCI = true
	generateCD = true

	if err := runGenerate(generateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		".github/workflows/ci.yml",
		".github/workflows/cd.yml",
	} {
		if _, err := os.Stat(filepath.Join(outputPath, want)); os.IsNotExist(err) {
			t.Errorf("expected %s to be generated when CLI flag is set even though manifest has it off", want)
		}
	}
}

func containsString(content, substr string) bool {
	for i := 0; i <= len(content)-len(substr); i++ {
		if content[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestGenerateAIDocsRespectSandboxEnabledFlags verifies that AGENTS.md and
// CLAUDE.md only mention sandbox environments that are actually enabled. This
// guards against the regression in issue #129 where `init` populates all three
// sandbox pointers (flox/devcontainer/dockerCompose), causing the templates to
// document every environment regardless of its `enabled` flag.
func TestGenerateAIDocsRespectSandboxEnabledFlags(t *testing.T) {
	tests := []struct {
		name              string
		floxEnabled       bool
		devContainerOn    bool
		dockerComposeOn   bool
		wantFlox          bool
		wantDevContainer  bool
		wantDockerCompose bool
	}{
		{
			name:        "only flox enabled",
			floxEnabled: true,
			wantFlox:    true,
		},
		{
			name:             "only devcontainer enabled",
			devContainerOn:   true,
			wantDevContainer: true,
		},
		{
			name:              "only docker-compose enabled",
			dockerComposeOn:   true,
			wantDockerCompose: true,
		},
		{
			name:              "all three enabled",
			floxEnabled:       true,
			devContainerOn:    true,
			dockerComposeOn:   true,
			wantFlox:          true,
			wantDevContainer:  true,
			wantDockerCompose: true,
		},
		{
			name: "all three pointers present but disabled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, "out")

			adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: sandbox-docs-agent
  description: Agent for verifying sandbox doc rendering
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/sandbox-docs
      version: "1.26.2"
  development:
    ai:
      claudecode:
        enabled: true
      codex:
        enabled: true
    sandbox:
      flox:
        enabled: ` + boolStr(tc.floxEnabled) + `
      devcontainer:
        enabled: ` + boolStr(tc.devContainerOn) + `
      dockerCompose:
        enabled: ` + boolStr(tc.dockerComposeOn) + `
`
			adlPath := filepath.Join(tempDir, "agent.yaml")
			if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
				t.Fatalf("failed to write ADL file: %v", err)
			}

			originalADLFile := adlFile
			originalOutputDir := outputDir
			defer func() {
				adlFile = originalADLFile
				outputDir = originalOutputDir
			}()

			adlFile = adlPath
			outputDir = outputPath

			if err := runGenerate(generateCmd, []string{}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, doc := range []string{"AGENTS.md", "CLAUDE.md"} {
				path := filepath.Join(outputPath, doc)
				bytes, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read %s: %v", doc, err)
				}
				content := string(bytes)

				hasFlox := strings.Contains(content, "Flox Environment")
				hasDevContainer := strings.Contains(content, "DevContainer")
				hasDockerCompose := strings.Contains(content, "Docker Compose")

				if hasFlox != tc.wantFlox {
					t.Errorf("%s: Flox section present=%v, want=%v", doc, hasFlox, tc.wantFlox)
				}
				if hasDevContainer != tc.wantDevContainer {
					t.Errorf("%s: DevContainer section present=%v, want=%v", doc, hasDevContainer, tc.wantDevContainer)
				}
				if hasDockerCompose != tc.wantDockerCompose {
					t.Errorf("%s: Docker Compose section present=%v, want=%v", doc, hasDockerCompose, tc.wantDockerCompose)
				}

				noneEnabled := !tc.wantFlox && !tc.wantDevContainer && !tc.wantDockerCompose
				hasFallback := strings.Contains(content, "No sandboxed environments are configured")
				if noneEnabled && !hasFallback {
					t.Errorf("%s: expected fallback message when no sandbox environments are enabled, got:\n%s", doc, content)
				}
				if !noneEnabled && hasFallback {
					t.Errorf("%s: unexpected fallback message when at least one sandbox is enabled", doc)
				}
			}
		})
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// TestGenerateDockerComposeFileEmitted asserts that running `generate` on a
// Go ADL with `spec.development.sandbox.dockerCompose.enabled: true`
// actually writes a docker-compose.yaml to the output directory and that
// the file wires up the gateway, the agent (built from source), the infer
// CLI, and the a2a-debugger. Regression guard for issue #148, where the
// flag was advertised in CLAUDE.md but the file generator never produced
// the compose file for Go projects.
func TestGenerateDockerComposeFileEmitted(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "out")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: compose-agent
  description: Agent for verifying docker-compose generation
  version: 1.0.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  server:
    port: 8443
    debug: false
  agent:
    provider: openai
    model: gpt-5-mini
    systemPrompt: hello
  language:
    go:
      module: github.com/test/compose-agent
      version: "1.26.2"
  development:
    sandbox:
      dockerCompose:
        enabled: true
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
	}()

	adlFile = adlPath
	outputDir = outputPath

	if err := runGenerate(generateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	composePath := filepath.Join(outputPath, "docker-compose.yaml")
	bytes, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("docker-compose.yaml was not generated (issue #148 regression): %v", err)
	}
	content := string(bytes)

	wantFragments := []string{
		"image: ghcr.io/inference-gateway/inference-gateway:latest",
		"build: .",
		"image: ghcr.io/inference-gateway/cli:latest",
		"image: ghcr.io/inference-gateway/a2a-debugger:latest",
		`profiles: ["cli"]`,
		`profiles: ["debugger"]`,
		"A2A_AGENT_CLIENT_BASE_URL",
		"compose-agent:",
	}
	for _, frag := range wantFragments {
		if !strings.Contains(content, frag) {
			t.Errorf("docker-compose.yaml missing %q\n---\n%s", frag, content)
		}
	}
}
