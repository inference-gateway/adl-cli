package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/inference-gateway/adl-cli/internal/schema"
	"github.com/inference-gateway/adl-cli/internal/templates"
	"gopkg.in/yaml.v3"
)

// Generator generates A2A agent projects from ADL files
type Generator struct {
	config Config
}

// Config holds generator configuration
type Config struct {
	Template       string
	Overwrite      bool
	Version        string
	GenerateCI     bool
	DeploymentType string
}

// New creates a new generator
func New(config Config) *Generator {
	return &Generator{
		config: config,
	}
}

// Generate generates an A2A agent project from an ADL file
func (g *Generator) Generate(adlFile, outputDir string) error {
	adl, err := g.parseADL(adlFile)
	if err != nil {
		return fmt.Errorf("failed to parse ADL file: %w", err)
	}

	if err := g.validateADL(adl); err != nil {
		return fmt.Errorf("ADL validation failed: %w", err)
	}

	// Override deployment configuration if specified via CLI flag
	if g.config.DeploymentType != "" {
		if adl.Spec.Deployment == nil {
			adl.Spec.Deployment = &schema.DeploymentConfig{}
		}
		adl.Spec.Deployment.Type = g.config.DeploymentType
	}

	template := g.config.Template
	if template == "" {
		template = g.detectTemplate(adl)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	language := templates.DetectLanguageFromADL(adl)

	registry, err := templates.NewRegistry(language)
	if err != nil {
		return fmt.Errorf("failed to create template registry: %w", err)
	}

	templateEngine := templates.NewWithRegistry(template, registry)

	if err := g.generateProject(templateEngine, adl, outputDir); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	return nil
}

// parseADL parses an ADL file
func (g *Generator) parseADL(adlFile string) (*schema.ADL, error) {
	data, err := os.ReadFile(adlFile)
	if err != nil {
		return nil, err
	}

	var adl schema.ADL
	if err := yaml.Unmarshal(data, &adl); err != nil {
		return nil, err
	}

	return &adl, nil
}

// validateADL validates the ADL structure for code generation requirements
func (g *Generator) validateADL(adl *schema.ADL) error {
	if adl.APIVersion != "adl.dev/v1" {
		return fmt.Errorf("unsupported API version: %s", adl.APIVersion)
	}
	if adl.Kind != "Agent" {
		return fmt.Errorf("unsupported kind: %s", adl.Kind)
	}

	if adl.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if adl.Metadata.Description == "" {
		return fmt.Errorf("metadata.description is required")
	}
	if adl.Metadata.Version == "" {
		return fmt.Errorf("metadata.version is required")
	}

	if adl.Spec.Server.Port == 0 {
		return fmt.Errorf("spec.server.port is required and must be greater than 0")
	}
	if adl.Spec.Server.Port < 1 || adl.Spec.Server.Port > 65535 {
		return fmt.Errorf("spec.server.port must be between 1 and 65535")
	}

	if adl.Spec.Capabilities == nil {
		return fmt.Errorf("spec.capabilities is required")
	}

	if adl.Spec.Language == nil {
		return fmt.Errorf("spec.language is required for code generation")
	}

	languageCount := 0
	if adl.Spec.Language.Go != nil {
		languageCount++
		if adl.Spec.Language.Go.Module == "" {
			return fmt.Errorf("spec.language.go.module is required")
		}
		if adl.Spec.Language.Go.Version == "" {
			return fmt.Errorf("spec.language.go.version is required")
		}
	}
	if adl.Spec.Language.TypeScript != nil {
		languageCount++
		if adl.Spec.Language.TypeScript.PackageName == "" {
			return fmt.Errorf("spec.language.typescript.packageName is required")
		}
		if adl.Spec.Language.TypeScript.NodeVersion == "" {
			return fmt.Errorf("spec.language.typescript.nodeVersion is required")
		}
	}
	if adl.Spec.Language.Rust != nil {
		languageCount++
		if adl.Spec.Language.Rust.PackageName == "" {
			return fmt.Errorf("spec.language.rust.packageName is required")
		}
		if adl.Spec.Language.Rust.Version == "" {
			return fmt.Errorf("spec.language.rust.version is required")
		}
		if adl.Spec.Language.Rust.Edition == "" {
			return fmt.Errorf("spec.language.rust.edition is required")
		}
	}

	if languageCount == 0 {
		return fmt.Errorf("at least one programming language must be defined in spec.language")
	}
	if languageCount > 1 {
		return fmt.Errorf("exactly one programming language must be defined for code generation, found %d", languageCount)
	}

	if adl.Spec.Agent != nil {
		if adl.Spec.Agent.Provider == "" {
			return fmt.Errorf("spec.agent.provider is required when agent configuration is specified")
		}
	}

	for i, tool := range adl.Spec.Tools {
		if tool.Name == "" {
			return fmt.Errorf("spec.tools[%d].name is required", i)
		}
		if tool.Description == "" {
			return fmt.Errorf("spec.tools[%d].description is required", i)
		}
		if tool.Schema == nil {
			return fmt.Errorf("spec.tools[%d].schema is required", i)
		}
	}

	return nil
}

// detectTemplate detects the appropriate template based on the ADL
func (g *Generator) detectTemplate(adl *schema.ADL) string {
	return "minimal"
}

// generateProject generates the complete project structure
func (g *Generator) generateProject(templateEngine *templates.Engine, adl *schema.ADL, outputDir string) error {
	ctx := templates.Context{
		ADL: adl,
		Metadata: schema.GeneratedMetadata{
			GeneratedAt: time.Now(),
			CLIVersion:  g.getVersion(),
			Template:    g.config.Template,
		},
		Language: templates.DetectLanguageFromADL(adl),
	}

	ignoreChecker, err := NewIgnoreChecker(outputDir)
	if err != nil {
		return fmt.Errorf("failed to initialize ignore checker: %w", err)
	}

	files := templateEngine.GetFiles(adl)
	for fileName, templateKey := range files {
		fileName = g.replacePlaceholders(fileName, adl)

		if ignoreChecker.ShouldIgnore(fileName) {
			fmt.Printf("🚫 Ignoring file (matches .adl-ignore): %s\n", fileName)
			continue
		}

		content, err := templateEngine.ExecuteTemplate(templateKey, ctx)
		if err != nil {
			return fmt.Errorf("failed to execute template %s: %w", templateKey, err)
		}

		ext := strings.ToLower(filepath.Ext(fileName))
		baseName := strings.ToLower(filepath.Base(fileName))

		var fileType string
		switch {
		case ext == ".go":
			fileType = "go"
		case ext == ".rs":
			fileType = "rust"
		case ext == ".yaml" || ext == ".yml":
			fileType = "yaml"
		case baseName == "dockerfile":
			fileType = "dockerfile"
		case baseName == "taskfile.yml":
			fileType = "taskfile"
		}

		if fileType != "" {
			header := templates.GetGeneratedFileHeader(fileType, ctx.Metadata.CLIVersion, ctx.Metadata.GeneratedAt)
			content = header + content
		}

		filePath := filepath.Join(outputDir, fileName)
		if err := g.writeFile(filePath, content); err != nil {
			return fmt.Errorf("failed to write %s: %w", fileName, err)
		}
	}

	if err := g.generateAgentJSON(adl, outputDir, ignoreChecker); err != nil {
		return fmt.Errorf("failed to generate agent.json: %w", err)
	}

	if err := g.generateADLIgnoreFile(outputDir, templateEngine.GetTemplate(), adl); err != nil {
		return fmt.Errorf("failed to generate .adl-ignore file: %w", err)
	}

	if g.config.GenerateCI {
		if err := g.generateCI(adl, outputDir, ignoreChecker); err != nil {
			return fmt.Errorf("failed to generate CI configuration: %w", err)
		}
	}

	return nil
}

// replacePlaceholders replaces placeholders in file names
func (g *Generator) replacePlaceholders(fileName string, adl *schema.ADL) string {
	replacements := map[string]string{
		"{{.Name}}": adl.Metadata.Name,
	}

	for placeholder, replacement := range replacements {
		fileName = strings.ReplaceAll(fileName, placeholder, replacement)
	}

	return fileName
}

// writeFile writes content to a file, creating directories as needed
func (g *Generator) writeFile(filePath, content string) error {
	if !g.config.Overwrite {
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("⚠️  Skipping existing file: %s\n", filePath)
			return nil
		}
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Generated: %s\n", filePath)
	return nil
}

// generateAgentJSON generates the .well-known/agent.json file
func (g *Generator) generateAgentJSON(adl *schema.ADL, outputDir string, ignoreChecker *IgnoreChecker) error {
	if ignoreChecker.ShouldIgnore(".well-known/agent.json") {
		fmt.Printf("🚫 Ignoring file (matches .adl-ignore): .well-known/agent.json\n")
		return nil
	}

	agentCard := map[string]interface{}{
		"name":         adl.Metadata.Name,
		"description":  adl.Metadata.Description,
		"version":      adl.Metadata.Version,
		"capabilities": adl.Spec.Capabilities,
		"_generated": map[string]interface{}{
			"by":        "A2A CLI",
			"version":   g.getVersion(),
			"timestamp": time.Now().Format(time.RFC3339),
			"warning":   "This file was automatically generated. DO NOT EDIT.",
		},
	}

	if len(adl.Spec.Tools) > 0 {
		tools := make([]map[string]interface{}, len(adl.Spec.Tools))
		for i, tool := range adl.Spec.Tools {
			tools[i] = map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"schema":      tool.Schema,
			}
		}
		agentCard["tools"] = tools
	}

	jsonData, err := g.formatJSONWithIndentation(agentCard)
	if err != nil {
		return err
	}

	wellKnownDir := filepath.Join(outputDir, ".well-known")
	if err := os.MkdirAll(wellKnownDir, 0755); err != nil {
		return err
	}

	agentJSONPath := filepath.Join(wellKnownDir, "agent.json")
	return g.writeFile(agentJSONPath, jsonData)
}

// getVersion returns the CLI version from config or default
func (g *Generator) getVersion() string {
	if g.config.Version != "" {
		return g.config.Version
	}
	return "dev"
}

// generateADLIgnoreFile creates a .adl-ignore file with files that contain TODOs
func (g *Generator) generateADLIgnoreFile(outputDir, templateName string, adl *schema.ADL) error {
	ignoreFilePath := filepath.Join(outputDir, ".adl-ignore")

	if _, err := os.Stat(ignoreFilePath); err == nil {
		fmt.Printf("📄 .adl-ignore file already exists, skipping creation\n")
		return nil
	}

	var filesToIgnore []string
	language := g.detectLanguage(adl)

	switch templateName {
	case "minimal":
		switch language {
		case "go":
			filesToIgnore = []string{
				"tools/*",
			}
		case "rust":
			filesToIgnore = []string{
				"src/tools/*",
			}
		case "typescript":
			filesToIgnore = []string{
				"src/tools/*",
			}
		}
	}

	if len(filesToIgnore) == 0 {
		return nil
	}

	content := generateA2aIgnoreContent(filesToIgnore)

	if err := os.WriteFile(ignoreFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .adl-ignore file: %w", err)
	}

	fmt.Printf("✅ Generated: .adl-ignore\n")
	fmt.Printf("🔒 Files with TODO implementations will be preserved on future generations\n")

	return nil
}

// generateA2aIgnoreContent generates the content for .adl-ignore file
func generateA2aIgnoreContent(filesToIgnore []string) string {
	content := `# .adl-ignore file
# This file specifies which files should not be overwritten during generation operations.
# Files listed here typically contain implementations that users have completed.
#
# Patterns supported:
# - Exact file names: tools/my_tool.go
# - Wildcards: *.go
# - Directory patterns: tools/*
# - Directories: build/
# - Comments: lines starting with #

`

	for _, file := range filesToIgnore {
		content += file + "\n"
	}

	content += `
# Add your own files to ignore here:
# my-custom-file.go
# config/secrets.yaml
`

	return content
}

// formatJSONWithIndentation formats JSON with proper indentation for nested objects
func (g *Generator) formatJSONWithIndentation(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// generateCI generates CI workflow configuration based on the programming language and SCM provider
func (g *Generator) generateCI(adl *schema.ADL, outputDir string, ignoreChecker *IgnoreChecker) error {
	scmProvider := g.detectSCMProvider(adl)

	switch scmProvider {
	case "github":
		return g.generateGitHubActionsWorkflow(adl, outputDir, ignoreChecker)
	case "gitlab":
		return g.generateGitLabCIWorkflow(adl, outputDir, ignoreChecker)
	default:
		fmt.Printf("⚠️  No SCM provider specified, defaulting to GitHub Actions\n")
		return g.generateGitHubActionsWorkflow(adl, outputDir, ignoreChecker)
	}
}

// detectLanguage detects the programming language from ADL
func (g *Generator) detectLanguage(adl *schema.ADL) string {
	if adl.Spec.Language.Go != nil {
		return "go"
	}
	if adl.Spec.Language.TypeScript != nil {
		return "typescript"
	}
	if adl.Spec.Language.Rust != nil {
		return "rust"
	}
	return "unknown"
}

// detectSCMProvider detects the SCM provider from ADL
func (g *Generator) detectSCMProvider(adl *schema.ADL) string {
	if adl.Spec.SCM != nil && adl.Spec.SCM.Provider != "" {
		return adl.Spec.SCM.Provider
	}
	return "github"
}

// generateGitHubActionsWorkflow generates a GitHub Actions workflow for projects
func (g *Generator) generateGitHubActionsWorkflow(adl *schema.ADL, outputDir string, ignoreChecker *IgnoreChecker) error {
	workflowPath := ".github/workflows/ci.yml"

	if ignoreChecker.ShouldIgnore(workflowPath) {
		fmt.Printf("🚫 Ignoring file (matches .adl-ignore): %s\n", workflowPath)
		return nil
	}

	language := g.detectLanguage(adl)
	var workflowContent string

	switch language {
	case "go":
		workflowContent = g.generateGoWorkflowContent(adl)
	case "typescript":
		workflowContent = g.generateTypeScriptWorkflowContent(adl)
	case "rust":
		workflowContent = g.generateRustWorkflowContent(adl)
	default:
		return fmt.Errorf("GitHub Actions generation not supported for language: %s", language)
	}

	fullWorkflowPath := filepath.Join(outputDir, workflowPath)
	if err := g.writeFile(fullWorkflowPath, workflowContent); err != nil {
		return fmt.Errorf("failed to write GitHub Actions workflow: %w", err)
	}

	fmt.Println("✅ CI workflow generated successfully!")
	fmt.Printf("📁 GitHub Actions workflow: %s\n", workflowPath)

	return nil
}

// generateGoWorkflowContent generates the GitHub Actions workflow content for Go projects
func (g *Generator) generateGoWorkflowContent(adl *schema.ADL) string {
	goVersion := "1.24"
	if adl.Spec.Language.Go != nil && adl.Spec.Language.Go.Version != "" {
		goVersion = adl.Spec.Language.Go.Version
	}

	return fmt.Sprintf(`name: CI

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  test:
    runs-on: ubuntu-24.04
    
    steps:
    - uses: actions/checkout@v4.2.2
    
    - name: Set up Go
      uses: actions/setup-go@v5.5.0
      with:
        go-version: %s
        cache: true

    - name: Install golangci-lint
      run: |
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v2.1.6

    - name: Install Task
      uses: arduino/setup-task@v2
      with:
        version: 3.x
    
    - name: Download dependencies
      run: go mod download
    
    - name: Format check
      run: task fmt
    
    - name: Lint
      run: task lint
    
    - name: Run tests
      run: task test
    
    - name: Build
      run: task build
`, goVersion)
}

// generateTypeScriptWorkflowContent generates the GitHub Actions workflow content for TypeScript projects
func (g *Generator) generateTypeScriptWorkflowContent(adl *schema.ADL) string {
	nodeVersion := "18"
	if adl.Spec.Language.TypeScript != nil && adl.Spec.Language.TypeScript.NodeVersion != "" {
		nodeVersion = adl.Spec.Language.TypeScript.NodeVersion
	}

	return fmt.Sprintf(`name: CI

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  test:
    runs-on: ubuntu-24.04
    
    steps:
    - uses: actions/checkout@v4.2.2
    
    - name: Set up Node.js
      uses: actions/setup-node@v4.1.0
      with:
        node-version: %s
        cache: 'npm'
    
    - name: Install Task
      uses: arduino/setup-task@v2
      with:
        version: 3.x
    
    - name: Install dependencies
      run: npm ci
    
    - name: Lint
      run: task lint
    
    - name: Run tests
      run: task test
    
    - name: Build
      run: task build
`, nodeVersion)
}

// generateRustWorkflowContent generates the GitHub Actions workflow content for Rust projects
func (g *Generator) generateRustWorkflowContent(adl *schema.ADL) string {
	rustVersion := "1.70"
	if adl.Spec.Language.Rust != nil && adl.Spec.Language.Rust.Version != "" {
		rustVersion = adl.Spec.Language.Rust.Version
	}

	return fmt.Sprintf(`name: CI

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  test:
    runs-on: ubuntu-24.04
    
    steps:
    - uses: actions/checkout@v4.2.2
    
    - name: Set up Rust
      uses: actions-rs/toolchain@v1
      with:
        toolchain: %s
        override: true
        components: rustfmt, clippy

    - name: Install Task
      uses: arduino/setup-task@v2
      with:
        version: 3.x
    
    - name: Cache cargo dependencies
      uses: actions/cache@v4
      with:
        path: |
          ~/.cargo/bin/
          ~/.cargo/registry/index/
          ~/.cargo/registry/cache/
          ~/.cargo/git/db/
          target/
        key: ${{ runner.os }}-cargo-${{ hashFiles('**/Cargo.lock') }}
    
    - name: Format check
      run: task fmt
    
    - name: Lint
      run: task lint
    
    - name: Run tests
      run: task test
    
    - name: Build
      run: task build
`, rustVersion)
}

// generateGitLabCIWorkflow generates a GitLab CI workflow
func (g *Generator) generateGitLabCIWorkflow(adl *schema.ADL, outputDir string, ignoreChecker *IgnoreChecker) error {
	// TODO: Implement GitLab CI workflow generation
	// This should generate .gitlab-ci.yml based on the programming language
	// and follow similar patterns to the GitHub Actions implementation
	fmt.Printf("⚠️  GitLab CI generation is not yet implemented\n")
	fmt.Printf("� This is a planned feature - contributions welcome!\n")
	return nil
}
