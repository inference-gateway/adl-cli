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

func TestGenerateWithAI(t *testing.T) {
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
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	originalEnableAI := enableAI
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
		enableAI = originalEnableAI
	}()

	adlFile = adlPath
	outputDir = outputPath
	enableAI = true

	err := runGenerate(generateCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claudeMdPath := filepath.Join(outputPath, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Errorf("expected CLAUDE.md to be generated when --ai flag is enabled")
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
		t.Errorf("expected devcontainer.json to contain claude-code extension when --ai flag is enabled")
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
		t.Errorf("expected manifest.toml to contain claude-code package when --ai flag is enabled")
	}
}

// TestGenerateWithAIFromManifest verifies that declaring spec.development.ai.enabled: true
// in the manifest activates AI assistant docs without needing the --ai CLI flag.
func TestGenerateWithAIFromManifest(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "ai-from-manifest")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: manifest-ai-agent
  description: AI enabled declaratively
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
      module: github.com/test/manifest-ai
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
	originalEnableAI := enableAI
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
		enableAI = originalEnableAI
	}()

	adlFile = adlPath
	outputDir = outputPath
	enableAI = false

	if err := runGenerate(generateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claudeMdPath := filepath.Join(outputPath, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Errorf("expected CLAUDE.md to be generated when spec.development.ai.enabled is true (without --ai flag)")
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

// TestGenerateCLIFlagOverridesManifest verifies that the --ai/--ci/--cd flags
// OR on top of the manifest value: setting the flag wins even if the manifest
// has the field unset or false.
func TestGenerateCLIFlagOverridesManifest(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "cli-overrides")

	adlContent := `apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: cli-overrides-agent
  description: Manifest leaves AI/CI/CD off but flags turn them on
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
  development:
    ai:
      enabled: false
`
	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	originalADLFile := adlFile
	originalOutputDir := outputDir
	originalEnableAI := enableAI
	originalGenerateCI := generateCI
	originalGenerateCD := generateCD
	defer func() {
		adlFile = originalADLFile
		outputDir = originalOutputDir
		enableAI = originalEnableAI
		generateCI = originalGenerateCI
		generateCD = originalGenerateCD
	}()

	adlFile = adlPath
	outputDir = outputPath
	enableAI = true
	generateCI = true
	generateCD = true

	if err := runGenerate(generateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		"CLAUDE.md",
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
