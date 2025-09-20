package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCommand(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	adlContent := `apiVersion: adl.dev/v1
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
      version: "1.25"
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

	adlContent := `apiVersion: adl.dev/v1
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
    provider: openai
    model: gpt-4o-mini
  skills:
    - id: test_skill_id
      name: test_skill
      description: A test skill
      tags: ["test"]
      schema:
        type: object
        properties:
          input:
            type: string
            description: Test input
        required: [input]
  server:
    port: 8080
    debug: false
  language:
    go:
      module: github.com/test/standalone
      version: "1.25"
`

	adlPath := filepath.Join(tempDir, "agent.yaml")
	if err := os.WriteFile(adlPath, []byte(adlContent), 0644); err != nil {
		t.Fatalf("failed to write ADL file: %v", err)
	}

	outputPath := filepath.Join(tempDir, "output")

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

	mainGoPath := filepath.Join(outputPath, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Errorf("expected main.go to be generated")
	}

	skillsDir := filepath.Join(outputPath, "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Errorf("expected skills directory to be generated")
	}

	testSkillPath := filepath.Join(skillsDir, "test_skill_id.go")
	if _, err := os.Stat(testSkillPath); os.IsNotExist(err) {
		t.Errorf("expected test_skill_id.go to be generated")
	}
}

func TestGenerateWithCD(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-cd-output")

	adlContent := `apiVersion: adl.dev/v1
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
      version: "1.25"
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

	adlContent := `apiVersion: adl.dev/v1
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
      version: "1.25"
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

func containsString(content, substr string) bool {
	for i := 0; i <= len(content)-len(substr); i++ {
		if content[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
