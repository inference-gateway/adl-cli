package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestInitCommand(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}

	err := runInit(cmd, []string{"test-agent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	adlPath := filepath.Join(outputPath, "agent.yaml")
	if _, err := os.Stat(adlPath); os.IsNotExist(err) {
		t.Errorf("expected ADL file to be created at %s", adlPath)
	}

	content, err := os.ReadFile(adlPath)
	if err != nil {
		t.Errorf("failed to read ADL file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "apiVersion: adl.dev/v1") {
		t.Errorf("ADL file missing apiVersion")
	}
	if !strings.Contains(contentStr, "kind: Agent") {
		t.Errorf("ADL file missing kind")
	}
}

func TestInitCommandIncludesSCMDefaults(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}

	err := runInit(cmd, []string{"test-agent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	adlPath := filepath.Join(outputPath, "agent.yaml")
	content, err := os.ReadFile(adlPath)
	if err != nil {
		t.Fatalf("failed to read ADL file: %v", err)
	}

	contentStr := string(content)
	
	if !strings.Contains(contentStr, "scm:") {
		t.Errorf("ADL file missing SCM configuration")
	}
	if !strings.Contains(contentStr, "provider: github") {
		t.Errorf("ADL file missing SCM provider default")
	}
	if !strings.Contains(contentStr, "github_app: true") {
		t.Errorf("ADL file missing SCM github_app default")
	}
	if !strings.Contains(contentStr, "issue_templates: true") {
		t.Errorf("ADL file missing SCM issue_templates default")
	}
	
	t.Logf("Generated ADL content:\n%s", contentStr)
}

func TestInitIssueTemplatesDefault(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}

	err := runInit(cmd, []string{"test-agent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	adlPath := filepath.Join(outputPath, "agent.yaml")
	content, err := os.ReadFile(adlPath)
	if err != nil {
		t.Fatalf("failed to read ADL file: %v", err)
	}

	var adl adlData
	if err := yaml.Unmarshal(content, &adl); err != nil {
		t.Fatalf("failed to parse ADL YAML: %v", err)
	}

	if adl.Spec.SCM == nil {
		t.Fatalf("expected SCM configuration to be present")
	}
	if adl.Spec.SCM.Provider != "github" {
		t.Errorf("expected SCM provider to be 'github', got: %s", adl.Spec.SCM.Provider)
	}
	if !adl.Spec.SCM.IssueTemplates {
		t.Errorf("expected IssueTemplates to be true by default")
	}
	if !adl.Spec.SCM.GithubApp {
		t.Errorf("expected GithubApp to be true by default")
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "issue_templates: true") {
		t.Errorf("ADL file should contain 'issue_templates: true'")
	}
}

func TestInitDoesNotGenerateCode(t *testing.T) {
	tempDir := t.TempDir()

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", tempDir); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("language", "go"); err != nil {
		t.Fatal(err)
	}

	err := runInit(cmd, []string{"test-agent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	adlPath := filepath.Join(tempDir, "agent.yaml")
	if _, err := os.Stat(adlPath); os.IsNotExist(err) {
		t.Errorf("expected ADL file to be created")
	}

	goModPath := filepath.Join(tempDir, "go.mod")
	if _, err := os.Stat(goModPath); !os.IsNotExist(err) {
		t.Errorf("init command should not generate go.mod file")
	}

	mainGoPath := filepath.Join(tempDir, "main.go")
	if _, err := os.Stat(mainGoPath); !os.IsNotExist(err) {
		t.Errorf("init command should not generate main.go file")
	}
}

func TestInitDefaultsVendorNeutral(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("type", "ai-powered"); err != nil {
		t.Fatal(err)
	}

	err := runInit(cmd, []string{"test-agent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	adlPath := filepath.Join(outputPath, "agent.yaml")
	content, err := os.ReadFile(adlPath)
	if err != nil {
		t.Fatalf("failed to read ADL file: %v", err)
	}

	var adl adlData
	if err := yaml.Unmarshal(content, &adl); err != nil {
		t.Fatalf("failed to parse ADL YAML: %v", err)
	}

	if adl.Spec.Agent == nil {
		t.Fatalf("expected agent spec to be present for ai-powered agent")
	}

	if adl.Spec.Agent.Provider != "" {
		t.Errorf("expected provider to be empty string for vendor neutrality, got: %s", adl.Spec.Agent.Provider)
	}

	if adl.Spec.Agent.Model != "" {
		t.Errorf("expected model to be empty string for vendor neutrality, got: %s", adl.Spec.Agent.Model)
	}
}
