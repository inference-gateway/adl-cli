package generator

import (
	"fmt"
	"os"
	"os/exec"
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
	Template           string
	Overwrite          bool
	Version            string
	GenerateCI         bool
	GenerateCD         bool
	DeploymentType     string
	EnableFlox         bool
	EnableDevContainer bool
	EnableAI           bool
	ADLFile            string
	OutputDir          string
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

	if g.config.DeploymentType != "" {
		if adl.Spec.Deployment == nil {
			adl.Spec.Deployment = &schema.DeploymentConfig{}
		}
		adl.Spec.Deployment.Type = g.config.DeploymentType

		if g.config.DeploymentType == "cloudrun" && adl.Spec.Deployment.CloudRun == nil {
			adl.Spec.Deployment.CloudRun = &schema.CloudRunConfig{
				Resources: &schema.ResourcesConfig{
					CPU:    "1",
					Memory: "512Mi",
				},
				Scaling: &schema.ScalingConfig{
					MinInstances: 0,
					MaxInstances: 10,
					Concurrency:  1000,
				},
				Service: &schema.ServiceConfig{
					Timeout:              3600,
					AllowUnauthenticated: true,
					ExecutionEnvironment: "gen2",
				},
			}
		}
	}

	if g.config.EnableFlox || g.config.EnableDevContainer {
		if adl.Spec.Sandbox == nil {
			adl.Spec.Sandbox = &schema.SandboxConfig{}
		}
		if g.config.EnableFlox && (adl.Spec.Sandbox.Flox == nil || !adl.Spec.Sandbox.Flox.Enabled) {
			adl.Spec.Sandbox.Flox = &schema.FloxConfig{Enabled: true}
		}
		if g.config.EnableDevContainer && (adl.Spec.Sandbox.DevContainer == nil || !adl.Spec.Sandbox.DevContainer.Enabled) {
			adl.Spec.Sandbox.DevContainer = &schema.DevContainerConfig{Enabled: true}
		}
	}

	template := g.config.Template
	if template == "" {
		template = g.detectTemplate(adl)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	language := templates.DetectLanguageFromADL(adl)

	registry, err := templates.NewRegistryWithOptions(templates.RegistryOptions{
		Language: language,
		EnableAI: g.config.EnableAI,
	})
	if err != nil {
		return fmt.Errorf("failed to create template registry: %w", err)
	}

	templateEngine := templates.NewWithRegistry(template, registry)

	if err := g.generateProject(templateEngine, adl, outputDir); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	if err := g.runPostGenerationSteps(adl, outputDir, language); err != nil {
		return fmt.Errorf("post-generation steps failed: %w", err)
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

	for i, skill := range adl.Spec.Skills {
		if skill.ID == "" {
			return fmt.Errorf("spec.skills[%d].id is required", i)
		}
		if skill.Name == "" {
			return fmt.Errorf("spec.skills[%d].name is required", i)
		}
		if skill.Description == "" {
			return fmt.Errorf("spec.skills[%d].description is required", i)
		}
		if len(skill.Tags) == 0 {
			return fmt.Errorf("spec.skills[%d].tags is required and must have at least one tag", i)
		}
		if skill.Schema == nil {
			return fmt.Errorf("spec.skills[%d].schema is required", i)
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
		Language:        templates.DetectLanguageFromADL(adl),
		GenerateCI:      g.config.GenerateCI,
		GenerateCD:      g.config.GenerateCD,
		EnableAI:        g.config.EnableAI,
		GenerateCommand: g.buildGenerateCommand(),
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

		var content string
		var err error

		if templateKey == "dependency.go" && strings.Contains(fileName, "internal/") && !strings.Contains(fileName, "internal/dependencies/") {
			parts := strings.Split(fileName, "/")
			if len(parts) >= 3 {
				dependencyFileName := parts[len(parts)-1]
				dependencyName := strings.TrimSuffix(dependencyFileName, filepath.Ext(dependencyFileName))

				var foundDependency string
				for depName := range adl.Spec.Dependencies {
					snakeCaseDependencyID := strings.ReplaceAll(depName, "-", "_")
					if snakeCaseDependencyID == dependencyName {
						foundDependency = depName
						break
					}
				}

				if foundDependency != "" {
					if foundDependency == "logger" {
						content, err = templateEngine.ExecuteToolTemplate("logger.go", ctx)
						if err != nil {
							return fmt.Errorf("failed to execute logger template: %w", err)
						}
					} else {
						dep := adl.Spec.Dependencies[foundDependency]
						depContext := map[string]interface{}{
							"Name":        dep.Interface,
							"ID":          foundDependency,
							"Interface":   dep.Interface,
							"Factory":     dep.Factory,
							"Type":        dep.Type,
							"Description": dep.Description,
							"Config":      adl.Spec.Config,
						}

						if adl.Spec.Language.Go != nil {
							depContext["GoModule"] = adl.Spec.Language.Go.Module
						}
						content, err = templateEngine.ExecuteToolTemplateWithContext(templateKey, depContext, ctx)
						if err != nil {
							return fmt.Errorf("failed to execute template %s for dependency %s: %w", templateKey, dependencyName, err)
						}
					}
				} else {
					return fmt.Errorf("dependency %s not found in ADL spec", dependencyName)
				}
			}
		} else if (templateKey == "skill.go" || templateKey == "skill.rs" || templateKey == "skill.ts") && strings.Contains(fileName, "/") {
			parts := strings.Split(fileName, "/")
			if len(parts) >= 2 {
				toolFileName := parts[len(parts)-1]
				toolName := strings.TrimSuffix(toolFileName, filepath.Ext(toolFileName))

				var foundSkill *schema.Skill
				for _, skill := range adl.Spec.Skills {
					snakeCaseSkillID := strings.ReplaceAll(skill.ID, "-", "_")
					if snakeCaseSkillID == toolName {
						foundSkill = &skill
						break
					}
				}

				if foundSkill != nil {
					skillContext := map[string]interface{}{
						"ID":             foundSkill.ID,
						"Name":           foundSkill.Name,
						"Description":    foundSkill.Description,
						"Tags":           foundSkill.Tags,
						"Examples":       foundSkill.Examples,
						"InputModes":     foundSkill.InputModes,
						"OutputModes":    foundSkill.OutputModes,
						"Schema":         foundSkill.Schema,
						"Implementation": foundSkill.Implementation,
						"Inject":         foundSkill.Inject,
					}

					if adl.Spec.Language.Go != nil {
						skillContext["GoModule"] = adl.Spec.Language.Go.Module
					}

					dependencyMap := make(map[string]interface{})
					for depName, dep := range adl.Spec.Dependencies {
						dependencyMap[depName] = map[string]interface{}{
							"ID":   depName,
							"Name": titleCase(depName),
							"Type": dep.Type,
							"Interface": dep.Interface,
							"Factory": dep.Factory,
							"Description": dep.Description,
						}
					}
					skillContext["DependencyMap"] = dependencyMap

					content, err = templateEngine.ExecuteToolTemplateWithContext(templateKey, skillContext, ctx)
					if err != nil {
						return fmt.Errorf("failed to execute template %s for skill %s: %w", templateKey, toolName, err)
					}
				} else {
					return fmt.Errorf("skill %s not found in ADL spec", toolName)
				}
			}
		} else if templateKey == "dependency.go" {
			return fmt.Errorf("dependency template reached fallback case - this should not happen")
		} else {
			content, err = templateEngine.ExecuteTemplate(templateKey, ctx)
			if err != nil {
				return fmt.Errorf("failed to execute template %s: %w", templateKey, err)
			}
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

		isSkillFile := (templateKey == "skill.go" || templateKey == "skill.rs" || templateKey == "skill.ts") ||
			(strings.Contains(fileName, "/skills/") && (ext == ".go" || ext == ".rs" || ext == ".ts"))

		if fileType != "" && !isSkillFile {
			header := templates.GetGeneratedFileHeader(fileType, ctx.Metadata.CLIVersion, ctx.Metadata.GeneratedAt)
			content = header + content
		}

		filePath := filepath.Join(outputDir, fileName)
		if err := g.writeFile(filePath, content); err != nil {
			return fmt.Errorf("failed to write %s: %w", fileName, err)
		}
	}

	// The agent.json is now generated via the template in the files map above
	// No need for separate generateAgentJSON function

	if err := g.generateADLIgnoreFile(outputDir, templateEngine.GetTemplate(), adl); err != nil {
		return fmt.Errorf("failed to generate .adl-ignore file: %w", err)
	}

	if g.config.GenerateCI {
		if err := g.generateCI(adl, outputDir, ignoreChecker); err != nil {
			return fmt.Errorf("failed to generate CI configuration: %w", err)
		}
	}

	if g.config.GenerateCD {
		if err := g.generateCD(adl, outputDir, ignoreChecker); err != nil {
			return fmt.Errorf("failed to generate CD configuration: %w", err)
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

// getVersion returns the CLI version from config or default
func (g *Generator) getVersion() string {
	if g.config.Version != "" {
		return g.config.Version
	}
	return "dev"
}

// buildGenerateCommand constructs the original adl generate command from the config
func (g *Generator) buildGenerateCommand() string {
	var parts []string

	parts = append(parts, "adl", "generate")

	if g.config.ADLFile != "" {
		parts = append(parts, "--file", g.config.ADLFile)
	} else {
		parts = append(parts, "--file", "agent.yaml")
	}

	if g.config.OutputDir != "" {
		parts = append(parts, "--output", g.config.OutputDir)
	} else {
		parts = append(parts, "--output", ".")
	}

	if g.config.Template != "" && g.config.Template != "minimal" {
		parts = append(parts, "--template", g.config.Template)
	}

	if g.config.Overwrite {
		parts = append(parts, "--overwrite")
	}

	if g.config.GenerateCI {
		parts = append(parts, "--ci")
	}

	if g.config.GenerateCD {
		parts = append(parts, "--cd")
	}

	if g.config.DeploymentType != "" {
		parts = append(parts, "--deployment", g.config.DeploymentType)
	}

	if g.config.EnableFlox {
		parts = append(parts, "--flox")
	}

	if g.config.EnableDevContainer {
		parts = append(parts, "--devcontainer")
	}

	if g.config.EnableAI {
		parts = append(parts, "--ai")
	}

	return strings.Join(parts, " ")
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
			for _, skill := range adl.Spec.Skills {
				snakeCaseName := strings.ReplaceAll(skill.ID, "-", "_")
				filesToIgnore = append(filesToIgnore, fmt.Sprintf("skills/%s.go", snakeCaseName))
			}

			for depName := range adl.Spec.Dependencies {
				snakeCaseName := strings.ReplaceAll(depName, "-", "_")
				filesToIgnore = append(filesToIgnore, fmt.Sprintf("internal/%s/%s.go", snakeCaseName, snakeCaseName))
			}
		case "rust":
			for _, skill := range adl.Spec.Skills {
				snakeCaseName := strings.ReplaceAll(skill.ID, "-", "_")
				filesToIgnore = append(filesToIgnore, fmt.Sprintf("src/skills/%s.rs", snakeCaseName))
			}

			for depName := range adl.Spec.Dependencies {
				snakeCaseName := strings.ReplaceAll(depName, "-", "_")
				filesToIgnore = append(filesToIgnore, fmt.Sprintf("src/dependencies/%s.rs", snakeCaseName))
			}
		case "typescript":
			for _, skill := range adl.Spec.Skills {
				snakeCaseName := strings.ReplaceAll(skill.ID, "-", "_")
				filesToIgnore = append(filesToIgnore, fmt.Sprintf("src/skills/%s.ts", snakeCaseName))
			}

			for depName := range adl.Spec.Dependencies {
				snakeCaseName := strings.ReplaceAll(depName, "-", "_")
				filesToIgnore = append(filesToIgnore, fmt.Sprintf("src/dependencies/%s.ts", snakeCaseName))
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
# - Exact file names: skills/agent_skill.go
# - Wildcards: *.go
# - Directory patterns: skills/*
# - Directories: build/
# - Comments: lines starting with #

`

	for _, file := range filesToIgnore {
		content += file + "\n"
	}

	content += `
# Go dependency files
go.sum

# Add your own files to ignore here:
# my-custom-file.go
# config/secrets.yaml
`

	return content
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

// generateGitHubActionsWorkflow generates a GitHub Actions workflow for projects using templates
func (g *Generator) generateGitHubActionsWorkflow(adl *schema.ADL, outputDir string, ignoreChecker *IgnoreChecker) error {
	workflowPath := ".github/workflows/ci.yml"

	if ignoreChecker.ShouldIgnore(workflowPath) {
		fmt.Printf("🚫 Ignoring file (matches .adl-ignore): %s\n", workflowPath)
		return nil
	}

	language := g.detectLanguage(adl)
	templateEngine, err := templates.NewRegistry(language)
	if err != nil {
		return fmt.Errorf("failed to create template registry: %w", err)
	}

	ctx := templates.Context{
		ADL: adl,
		Metadata: schema.GeneratedMetadata{
			CLIVersion:  g.config.Version,
			GeneratedAt: time.Now(),
			ADLFile:     g.config.ADLFile,
			Template:    g.config.Template,
		},
		Language:        language,
		GenerateCI:      g.config.GenerateCI,
		GenerateCD:      g.config.GenerateCD,
		EnableAI:        g.config.EnableAI,
		GenerateCommand: g.buildGenerateCommand(),
	}

	templateKey := fmt.Sprintf("github/workflows/ci.%s.yaml", language)
	workflowContent, err := templates.NewWithRegistry("", templateEngine).ExecuteTemplate(templateKey, ctx)
	if err != nil {
		return fmt.Errorf("failed to execute CI workflow template: %w", err)
	}

	// Add header to the workflow content
	header := templates.GetGeneratedFileHeader("yaml", ctx.Metadata.CLIVersion, ctx.Metadata.GeneratedAt)
	workflowContent = header + workflowContent

	fullWorkflowPath := filepath.Join(outputDir, workflowPath)
	if err := g.writeFile(fullWorkflowPath, workflowContent); err != nil {
		return fmt.Errorf("failed to write GitHub Actions workflow: %w", err)
	}

	fmt.Println("✅ CI workflow generated successfully!")
	fmt.Printf("📁 GitHub Actions workflow: %s\n", workflowPath)

	return nil
}

// runPostGenerationSteps runs language-specific post-generation steps
func (g *Generator) runPostGenerationSteps(adl *schema.ADL, outputDir, language string) error {
	var commands []string

	if adl.Spec.Hooks != nil && len(adl.Spec.Hooks.Post) > 0 {
		commands = adl.Spec.Hooks.Post
		fmt.Println("🔧 Running custom post-generation hooks...")
	} else {
		switch language {
		case "go":
			commands = []string{"go mod tidy", "go fmt ./..."}
			fmt.Println("🔧 Running default Go post-generation commands...")
		case "rust":
			commands = []string{"cargo fmt", "cargo check"}
		case "typescript":
			// Default TypeScript commands could be added here
			// commands = []string{"npm install", "npm run format"}
			return nil
		default:
			return nil
		}
	}

	for _, cmdStr := range commands {
		fmt.Printf("  ▶ Running: %s\n", cmdStr)

		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = outputDir
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("    ⚠️  Warning: command failed: %v\n", err)
			if len(output) > 0 {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if line != "" {
						fmt.Printf("       %s\n", line)
					}
				}
			}
			fmt.Printf("       You can run '%s' manually later\n", cmdStr)
			continue
		}

		fmt.Printf("    ✅ Successfully completed\n")
		if len(output) > 0 && strings.TrimSpace(string(output)) != "" {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("       %s\n", line)
				}
			}
		}
	}

	return nil
}

// generateGitLabCIWorkflow generates a GitLab CI workflow
func (g *Generator) generateGitLabCIWorkflow(adl *schema.ADL, outputDir string, ignoreChecker *IgnoreChecker) error {
	// TODO: Implement GitLab CI workflow generation
	// This should generate .gitlab-ci.yml based on the programming language
	// and follow similar patterns to the GitHub Actions implementation
	fmt.Printf("⚠️  GitLab CI generation is not yet implemented\n")
	fmt.Printf("This is a planned feature - contributions welcome!\n")
	return nil
}

// generateCD generates CD configuration files based on the programming language and SCM provider
func (g *Generator) generateCD(adl *schema.ADL, outputDir string, ignoreChecker *IgnoreChecker) error {
	scmProvider := g.detectSCMProvider(adl)

	switch scmProvider {
	case "github":
		return g.generateGitHubCDWorkflow(adl, outputDir, ignoreChecker)
	case "gitlab":
		return g.generateGitLabCDWorkflow(adl, outputDir, ignoreChecker)
	default:
		fmt.Printf("⚠️  No SCM provider specified, defaulting to GitHub Actions\n")
		return g.generateGitHubCDWorkflow(adl, outputDir, ignoreChecker)
	}
}

// generateGitHubCDWorkflow generates GitHub CD workflow and semantic-release configuration
func (g *Generator) generateGitHubCDWorkflow(adl *schema.ADL, outputDir string, ignoreChecker *IgnoreChecker) error {
	language := g.detectLanguage(adl)
	template := g.detectTemplate(adl)

	registry, err := templates.NewRegistry(language)
	if err != nil {
		return fmt.Errorf("failed to create template registry: %w", err)
	}

	templateEngine := templates.NewWithRegistry(template, registry)

	ctx := templates.Context{
		ADL: adl,
		Metadata: schema.GeneratedMetadata{
			GeneratedAt: time.Now(),
			CLIVersion:  g.getVersion(),
			Template:    g.config.Template,
		},
		Language:        language,
		GenerateCI:      g.config.GenerateCI,
		GenerateCD:      g.config.GenerateCD,
		EnableAI:        g.config.EnableAI,
		GenerateCommand: g.buildGenerateCommand(),
	}

	if err := g.generateReleaseRC(templateEngine, ctx, outputDir, ignoreChecker); err != nil {
		return fmt.Errorf("failed to generate .releaserc.yaml: %w", err)
	}

	workflowPath := ".github/workflows/cd.yml"

	if ignoreChecker.ShouldIgnore(workflowPath) {
		fmt.Printf("🚫 Ignoring file (matches .adl-ignore): %s\n", workflowPath)
		return nil
	}

	templateKey := "github/workflows/cd.yaml"

	workflowContent, err := templateEngine.ExecuteTemplate(templateKey, ctx)
	if err != nil {
		return fmt.Errorf("failed to execute CD workflow template: %w", err)
	}

	// Add header to the CD workflow content
	header := templates.GetGeneratedFileHeader("yaml", ctx.Metadata.CLIVersion, ctx.Metadata.GeneratedAt)
	workflowContent = header + workflowContent

	fullWorkflowPath := filepath.Join(outputDir, workflowPath)
	if err := g.writeFile(fullWorkflowPath, workflowContent); err != nil {
		return fmt.Errorf("failed to write GitHub CD workflow: %w", err)
	}

	fmt.Println("✅ CD pipeline generated successfully!")
	fmt.Printf("📁 GitHub CD workflow: %s\n", workflowPath)
	fmt.Printf("📁 Semantic release config: .releaserc.yaml\n")

	return nil
}

// generateReleaseRC generates the .releaserc.yaml configuration file
func (g *Generator) generateReleaseRC(templateEngine *templates.Engine, ctx templates.Context, outputDir string, ignoreChecker *IgnoreChecker) error {
	releasercPath := ".releaserc.yaml"

	if ignoreChecker.ShouldIgnore(releasercPath) {
		fmt.Printf("🚫 Ignoring file (matches .adl-ignore): %s\n", releasercPath)
		return nil
	}

	releasercContent, err := templateEngine.ExecuteTemplate("config/releaserc.yaml", ctx)
	if err != nil {
		return fmt.Errorf("failed to execute releaserc template: %w", err)
	}

	// Add header to the .releaserc.yaml content
	header := templates.GetGeneratedFileHeader("yaml", ctx.Metadata.CLIVersion, ctx.Metadata.GeneratedAt)
	releasercContent = header + releasercContent

	fullReleasercPath := filepath.Join(outputDir, releasercPath)
	if err := g.writeFile(fullReleasercPath, releasercContent); err != nil {
		return fmt.Errorf("failed to write .releaserc.yaml: %w", err)
	}

	return nil
}

// generateGitLabCDWorkflow generates a GitLab CD workflow
func (g *Generator) generateGitLabCDWorkflow(adl *schema.ADL, outputDir string, ignoreChecker *IgnoreChecker) error {
	fmt.Printf("⚠️  GitLab CD generation is not yet implemented\n")
	fmt.Printf("This is a planned feature - contributions welcome!\n")
	return nil
}

// titleCase converts a string to title case (first letter uppercase)
func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
