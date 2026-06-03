package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/inference-gateway/adl-cli/internal/schema"
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
	if !strings.Contains(contentStr, "apiVersion: adl.inference-gateway.com/v1") {
		t.Errorf("ADL file missing apiVersion")
	}
	if !strings.Contains(contentStr, "kind: Agent") {
		t.Errorf("ADL file missing kind")
	}
}

// TestInitDefaultsManifestValidates guards issue #190: `adl init --defaults`
// adds a default `logger` service, which must be rendered as a schema-valid
// object (keyed by name) so the generated manifest passes `adl validate`.
func TestInitDefaultsManifestValidates(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}

	if err := runInit(cmd, []string{"test-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	adlPath := filepath.Join(outputPath, "agent.yaml")

	var adl adlData
	content, err := os.ReadFile(adlPath)
	if err != nil {
		t.Fatalf("failed to read ADL file: %v", err)
	}
	if err := yaml.Unmarshal(content, &adl); err != nil {
		t.Fatalf("failed to parse ADL YAML: %v", err)
	}
	logger, ok := adl.Spec.Services["logger"]
	if !ok {
		t.Fatalf("expected default logger service, got services=%v", adl.Spec.Services)
	}
	if logger.Type == "" || logger.Interface == "" || logger.Factory == "" {
		t.Errorf("default logger service missing required fields: %+v", logger)
	}

	if _, err := schema.NewValidator().ValidateFile(adlPath); err != nil {
		t.Fatalf("default manifest failed schema validation: %v", err)
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
	if !strings.Contains(contentStr, "issue_templates: false") {
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
	if adl.Spec.SCM.IssueTemplates {
		t.Errorf("expected IssueTemplates to be false by default")
	}
	if !adl.Spec.SCM.GithubApp {
		t.Errorf("expected GithubApp to be true by default")
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "issue_templates: false") {
		t.Errorf("ADL file should contain 'issue_templates: false'")
	}
}

func TestInitDependabotDefault(t *testing.T) {
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
	if adl.Spec.SCM.Dependabot {
		t.Errorf("expected Dependabot to be false by default, got true")
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "dependabot: false") {
		t.Errorf("ADL file should contain 'dependabot: false'")
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

// TestInitAICICDDefaults verifies that `adl init --defaults` writes
// ai/ci/cd as false (or omitted) by default - they should be opt-in.
func TestInitAICICDDefaults(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}

	if err := runInit(cmd, []string{"test-agent"}); err != nil {
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
	if adl.Spec.SCM.CI {
		t.Errorf("expected SCM.CI to default to false")
	}
	if adl.Spec.SCM.CD {
		t.Errorf("expected SCM.CD to default to false")
	}

	if adl.Spec.Development != nil && adl.Spec.Development.AI != nil && adl.Spec.Development.AI.Orchestrators != nil {
		if o := adl.Spec.Development.AI.Orchestrators; o.Claudecode != nil && o.Claudecode.Enabled ||
			o.Codex != nil && o.Codex.Enabled ||
			o.Gemini != nil && o.Gemini.Enabled ||
			o.Opencode != nil && o.Opencode.Enabled ||
			o.Infer != nil && o.Infer.Enabled {
			t.Errorf("expected every spec.development.ai.orchestrators.<agent>.enabled to default to false, got at least one true")
		}
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "ci: false") {
		t.Errorf("ADL file should contain 'ci: false', got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "cd: false") {
		t.Errorf("ADL file should contain 'cd: false', got:\n%s", contentStr)
	}
}

// TestInitAIFlag verifies that `--ai` flag at init time writes
// spec.development.ai.orchestrators.claudecode.enabled: true.
func TestInitAIFlag(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("ai", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Flags().Set("ai", "false") }()

	if err := runInit(cmd, []string{"test-agent"}); err != nil {
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

	if adl.Spec.Development == nil || adl.Spec.Development.AI == nil ||
		adl.Spec.Development.AI.Orchestrators == nil ||
		adl.Spec.Development.AI.Orchestrators.Claudecode == nil ||
		!adl.Spec.Development.AI.Orchestrators.Claudecode.Enabled {
		t.Errorf("expected spec.development.ai.orchestrators.claudecode.enabled to be true when --ai is passed to init")
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "development:") || !strings.Contains(contentStr, "orchestrators:") || !strings.Contains(contentStr, "claudecode:") || !strings.Contains(contentStr, "enabled: true") {
		t.Errorf("ADL file should contain development.ai.orchestrators.claudecode.enabled: true, got:\n%s", contentStr)
	}
}

// TestInitDevelopmentDefaultsEmitted verifies that `adl init --defaults`
// always writes spec.development.sandbox.{flox,devcontainer,dockerCompose}.enabled
// and spec.development.ai.orchestrators.<agent>.enabled as explicit `false`
// values, so the defaults are discoverable in the generated agent.yaml rather
// than silently omitted.
func TestInitDevelopmentDefaultsEmitted(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}

	if err := runInit(cmd, []string{"test-agent"}); err != nil {
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

	if adl.Spec.Development == nil {
		t.Fatalf("expected spec.development to be present by default")
	}
	if adl.Spec.Development.Sandbox == nil {
		t.Fatalf("expected spec.development.sandbox to be present by default")
	}
	if adl.Spec.Development.Sandbox.Flox == nil || adl.Spec.Development.Sandbox.Flox.Enabled {
		t.Errorf("expected spec.development.sandbox.flox.enabled to be false by default")
	}
	if adl.Spec.Development.Sandbox.DevContainer == nil || adl.Spec.Development.Sandbox.DevContainer.Enabled {
		t.Errorf("expected spec.development.sandbox.devcontainer.enabled to be false by default")
	}
	if adl.Spec.Development.Sandbox.DockerCompose == nil || adl.Spec.Development.Sandbox.DockerCompose.Enabled {
		t.Errorf("expected spec.development.sandbox.dockerCompose.enabled to be false by default")
	}
	if adl.Spec.Development.AI == nil || adl.Spec.Development.AI.Orchestrators == nil {
		t.Fatalf("expected spec.development.ai.orchestrators to be present by default")
	}
	if o := adl.Spec.Development.AI.Orchestrators; (o.Claudecode != nil && o.Claudecode.Enabled) ||
		(o.Codex != nil && o.Codex.Enabled) ||
		(o.Gemini != nil && o.Gemini.Enabled) ||
		(o.Opencode != nil && o.Opencode.Enabled) ||
		(o.Infer != nil && o.Infer.Enabled) {
		t.Errorf("expected every spec.development.ai.orchestrators.<agent>.enabled to default to false")
	}

	contentStr := string(content)
	for _, want := range []string{
		"development:",
		"sandbox:",
		"flox:",
		"devcontainer:",
		"dockerCompose:",
		"ai:",
		"claudecode:",
		"codex:",
		"gemini:",
		"opencode:",
		"infer:",
		// Empty-list extension points must be rendered explicitly so
		// first-time users see where to drop additional dependencies
		// without consulting the schema.
		"vendor:",
		"deps: []",
		"devdeps: []",
	} {
		if !strings.Contains(contentStr, want) {
			t.Errorf("ADL file should contain %q, got:\n%s", want, contentStr)
		}
	}

	t.Logf("Generated ADL content:\n%s", contentStr)
}

// TestInitDefaultsEmitsVendorAndSandboxDeps asserts that `adl init --defaults`
// writes the three list-shaped extension points (`spec.language.go.vendor.deps`,
// `spec.language.go.vendor.devdeps`, `spec.development.deps`) as explicit empty
// lists. These keys are how users discover where to add extra Go modules /
// dev tools / cross-cutting sandbox packages; omitting them on init forces
// users to consult the schema. See issue #156.
func TestInitDefaultsEmitsVendorAndSandboxDeps(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("language", "go"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Flags().Set("language", "") }()

	if err := runInit(cmd, []string{"test-agent"}); err != nil {
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

	if adl.Spec.Language == nil || adl.Spec.Language.Go == nil {
		t.Fatalf("expected spec.language.go to be present")
	}
	if adl.Spec.Language.Go.Vendor == nil {
		t.Fatalf("expected spec.language.go.vendor to be present by default")
	}
	if adl.Spec.Language.Go.Vendor.Deps == nil {
		t.Errorf("expected spec.language.go.vendor.deps to be an empty list (not nil)")
	}
	if len(adl.Spec.Language.Go.Vendor.Deps) != 0 {
		t.Errorf("expected spec.language.go.vendor.deps to be empty, got %v", adl.Spec.Language.Go.Vendor.Deps)
	}
	if adl.Spec.Language.Go.Vendor.Devdeps == nil {
		t.Errorf("expected spec.language.go.vendor.devdeps to be an empty list (not nil)")
	}
	if len(adl.Spec.Language.Go.Vendor.Devdeps) != 0 {
		t.Errorf("expected spec.language.go.vendor.devdeps to be empty, got %v", adl.Spec.Language.Go.Vendor.Devdeps)
	}

	if adl.Spec.Development == nil {
		t.Fatalf("expected spec.development to be present by default")
	}
	if adl.Spec.Development.Deps == nil {
		t.Errorf("expected spec.development.deps to be an empty list (not nil)")
	}
	if len(adl.Spec.Development.Deps) != 0 {
		t.Errorf("expected spec.development.deps to be empty, got %v", adl.Spec.Development.Deps)
	}

	contentStr := string(content)
	for _, want := range []string{
		"vendor:",
		"deps: []",
		"devdeps: []",
	} {
		if !strings.Contains(contentStr, want) {
			t.Errorf("ADL file should contain %q, got:\n%s", want, contentStr)
		}
	}
}

// TestInitDockerComposeFlag verifies that the --docker-compose flag at init
// time writes spec.development.sandbox.dockerCompose.enabled: true.
func TestInitDockerComposeFlag(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-output")

	cmd := initCmd
	if err := cmd.Flags().Set("defaults", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("path", outputPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("docker-compose", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Flags().Set("docker-compose", "false") }()

	if err := runInit(cmd, []string{"test-agent"}); err != nil {
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

	if adl.Spec.Development == nil ||
		adl.Spec.Development.Sandbox == nil ||
		adl.Spec.Development.Sandbox.DockerCompose == nil ||
		!adl.Spec.Development.Sandbox.DockerCompose.Enabled {
		t.Errorf("expected spec.development.sandbox.dockerCompose.enabled to be true when --docker-compose is passed to init")
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

// TestInitProviderModelFlags guards issue #191: `--provider` and `--model` must
// be honored in `--defaults` mode and written into the manifest, instead of
// being silently dropped in favor of the empty default.
func TestInitProviderModelFlags(t *testing.T) {
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
	defer func() { _ = cmd.Flags().Set("type", "") }()
	if err := cmd.Flags().Set("provider", "anthropic"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Flags().Set("provider", "") }()
	if err := cmd.Flags().Set("model", "claude-sonnet-4-5"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Flags().Set("model", "") }()

	if err := runInit(cmd, []string{"test-agent"}); err != nil {
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
	if adl.Spec.Agent.Provider != "anthropic" {
		t.Errorf("expected provider 'anthropic' from --provider flag, got: %q", adl.Spec.Agent.Provider)
	}
	if adl.Spec.Agent.Model != "claude-sonnet-4-5" {
		t.Errorf("expected model 'claude-sonnet-4-5' from --model flag, got: %q", adl.Spec.Agent.Model)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "provider: anthropic") {
		t.Errorf("ADL file should contain 'provider: anthropic', got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "model: claude-sonnet-4-5") {
		t.Errorf("ADL file should contain 'model: claude-sonnet-4-5', got:\n%s", contentStr)
	}
}
