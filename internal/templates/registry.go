package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

// Registry manages template loading and lookup
type Registry struct {
	templates map[string]string
	language  string
	enableAI  bool
}

// Embed template files at compile time
//
//go:embed languages/*/*.tmpl languages/*/builtin/*.tmpl common/*/*.tmpl common/github/*/*.tmpl sandbox/*/*.tmpl
var templateFS embed.FS

// RegistryOptions holds options for creating a new registry
type RegistryOptions struct {
	Language string
	EnableAI bool
}

// NewRegistry creates a new template registry for the specified language
func NewRegistry(language string) (*Registry, error) {
	return NewRegistryWithOptions(RegistryOptions{
		Language: language,
		EnableAI: false,
	})
}

// NewRegistryWithOptions creates a new template registry with options
func NewRegistryWithOptions(opts RegistryOptions) (*Registry, error) {
	r := &Registry{
		templates: make(map[string]string),
		language:  opts.Language,
		enableAI:  opts.EnableAI,
	}

	if err := r.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return r, nil
}

// loadTemplates loads all templates from the embedded filesystem
func (r *Registry) loadTemplates() error {
	langPath := filepath.Join("languages", r.language)
	if err := r.loadTemplatesFromPath(langPath); err != nil {
		return fmt.Errorf("failed to load language templates: %w", err)
	}

	if err := r.loadTemplatesFromPath("common"); err != nil {
		return fmt.Errorf("failed to load common templates: %w", err)
	}

	if err := r.loadTemplatesFromPath("sandbox"); err != nil {
		return fmt.Errorf("failed to load sandbox templates: %w", err)
	}

	return nil
}

// loadTemplatesFromPath loads templates from a specific path
func (r *Registry) loadTemplatesFromPath(path string) error {
	return fs.WalkDir(templateFS, path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(filePath, ".tmpl") {
			return nil
		}

		content, err := templateFS.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", filePath, err)
		}

		key := r.createTemplateKey(filePath)
		r.templates[key] = string(content)

		return nil
	})
}

// createTemplateKey creates a template key from a file path
func (r *Registry) createTemplateKey(filePath string) string {
	key := strings.TrimSuffix(filePath, ".tmpl")

	langPrefix := fmt.Sprintf("languages/%s/", r.language)
	key = strings.TrimPrefix(key, langPrefix)

	key = strings.TrimPrefix(key, "common/")

	key = strings.TrimPrefix(key, "sandbox/")

	return key
}

// GetTemplate retrieves a template by key
func (r *Registry) GetTemplate(key string) (string, error) {
	// Try exact match first
	if tmpl, ok := r.templates[key]; ok {
		return tmpl, nil
	}

	langKey := fmt.Sprintf("%s.%s", key, r.language)
	if tmpl, ok := r.templates[langKey]; ok {
		return tmpl, nil
	}

	return "", fmt.Errorf("template not found: %s", key)
}

// GetFiles returns all files that should be generated for the current language
func (r *Registry) GetFiles(adl *schema.ADL) map[string]string {
	switch r.language {
	case "go":
		return r.getGoFiles(adl)
	case "rust":
		return r.getRustFiles(adl)
	case "typescript":
		return r.getTypeScriptFiles(adl)
	default:
		return r.getGoFiles(adl)
	}
}

// getGoFiles returns the file mapping for Go projects
func (r *Registry) getGoFiles(adl *schema.ADL) map[string]string {
	files := map[string]string{
		"main.go":                     "main.go",
		"go.mod":                      "go.mod",
		"config/config.go":            "config.go",
		".well-known/agent-card.json": "config/agent.json",
		"Taskfile.yml":                "taskfile/taskfile.yml",
		"Dockerfile":                  "docker/dockerfile.go",
		".gitignore":                  "config/gitignore",
		".gitattributes":              "config/gitattributes",
		".editorconfig":               "config/editorconfig",
		"README.md":                   "docs/README.md",
	}

	if adl.Spec.Deployment != nil {
		switch adl.Spec.Deployment.Type {
		case schema.DeploymentConfigTypeKubernetes:
			files["k8s/deployment.yaml"] = "kubernetes/deployment.yaml"
		case schema.DeploymentConfigTypeCloudRun:
			// CloudRun deployment is handled via Taskfile
		}
	}

	for _, tool := range adl.Spec.Tools {
		if schema.IsReservedToolID(tool.ID) {
			files[fmt.Sprintf("tools/%s.go", tool.ID)] = fmt.Sprintf("builtin/%s.go", tool.ID)
			continue
		}
		snakeCaseName := strings.ReplaceAll(tool.ID, "-", "_")
		files[fmt.Sprintf("tools/%s.go", snakeCaseName)] = "tool.go"
	}

	for _, skill := range adl.Spec.Skills {
		if skill.Bare {
			files[fmt.Sprintf("skills/%s/SKILL.md", skill.ID)] = "skills/skill.md"
		}
	}

	files["internal/logger/logger.go"] = "logger.go"

	for serviceName := range adl.Spec.Services {
		snakeCaseName := strings.ReplaceAll(serviceName, "-", "_")
		files[fmt.Sprintf("internal/%s/%s.go", snakeCaseName, snakeCaseName)] = "service.go"
	}

	r.addSandboxFiles(adl, files)
	r.addAIFiles(files)
	r.addIssueTemplateFiles(adl, files)
	r.addDependabotFiles(adl, files)

	return files
}

// getRustFiles returns the file mapping for Rust projects
func (r *Registry) getRustFiles(adl *schema.ADL) map[string]string {
	files := map[string]string{
		"src/main.rs":                 "main.rs",
		"Cargo.toml":                  "Cargo.toml",
		".well-known/agent-card.json": "config/agent.json",
		"Taskfile.yml":                "taskfile/taskfile.yml",
		"Dockerfile":                  "docker/dockerfile.rust",
		".gitignore":                  "config/gitignore",
		".gitattributes":              "config/gitattributes",
		".editorconfig":               "config/editorconfig",
		".env.example":                "env.example",
		"README.md":                   "docs/README.md",
	}

	if adl.Spec.Deployment != nil {
		switch adl.Spec.Deployment.Type {
		case schema.DeploymentConfigTypeKubernetes:
			files["k8s/deployment.yaml"] = "kubernetes/deployment.yaml"
		case schema.DeploymentConfigTypeCloudRun:
			// CloudRun deployment is handled via Taskfile
		}
	}

	if adl.Spec.Agent != nil {
		for _, tool := range adl.Spec.Tools {
			if schema.IsReservedToolID(tool.ID) {
				files[fmt.Sprintf("src/tools/%s.rs", tool.ID)] = fmt.Sprintf("builtin/%s.rs", tool.ID)
				continue
			}
			snakeCaseName := strings.ReplaceAll(tool.ID, "-", "_")
			files[fmt.Sprintf("src/tools/%s.rs", snakeCaseName)] = "tool.rs"
		}

		if len(adl.Spec.Tools) > 0 {
			files["src/tools/mod.rs"] = "tool.mod.rs"
		}
	}

	for _, skill := range adl.Spec.Skills {
		if skill.Bare {
			files[fmt.Sprintf("skills/%s/SKILL.md", skill.ID)] = "skills/skill.md"
		}
	}

	if adl.Spec.Development != nil &&
		adl.Spec.Development.Sandbox != nil &&
		adl.Spec.Development.Sandbox.DockerCompose != nil &&
		adl.Spec.Development.Sandbox.DockerCompose.Enabled {
		files["docker-compose.yaml"] = "docker-compose.yaml"
	}

	r.addSandboxFiles(adl, files)
	r.addAIFiles(files)
	r.addIssueTemplateFiles(adl, files)
	r.addDependabotFiles(adl, files)

	return files
}

// getTypeScriptFiles returns the file mapping for TypeScript projects
func (r *Registry) getTypeScriptFiles(adl *schema.ADL) map[string]string {
	files := map[string]string{
		"src/index.ts":                "index.ts",
		"package.json":                "package.json",
		"tsconfig.json":               "tsconfig.json",
		".well-known/agent-card.json": "config/agent.json",
		"Taskfile.yml":                "taskfile/taskfile.yml",
		"Dockerfile":                  "docker/dockerfile.ts",
		".gitignore":                  "config/gitignore",
		".gitattributes":              "config/gitattributes",
		".editorconfig":               "config/editorconfig",
		"README.md":                   "docs/README.md",
	}

	if adl.Spec.Deployment != nil {
		switch adl.Spec.Deployment.Type {
		case schema.DeploymentConfigTypeKubernetes:
			files["k8s/deployment.yaml"] = "kubernetes/deployment.yaml"
		case schema.DeploymentConfigTypeCloudRun:
			// CloudRun deployment is handled via Taskfile
		}
	}

	for _, tool := range adl.Spec.Tools {
		snakeCaseName := strings.ReplaceAll(tool.ID, "-", "_")
		files[fmt.Sprintf("src/tools/%s.ts", snakeCaseName)] = "tool.ts"
	}

	for _, skill := range adl.Spec.Skills {
		if skill.Bare {
			files[fmt.Sprintf("skills/%s/SKILL.md", skill.ID)] = "skills/skill.md"
		}
	}

	r.addSandboxFiles(adl, files)
	r.addAIFiles(files)
	r.addIssueTemplateFiles(adl, files)
	r.addDependabotFiles(adl, files)

	return files
}

// addAIFiles adds AI-related files to the file mapping when EnableAI is true
func (r *Registry) addAIFiles(files map[string]string) {
	if r.enableAI {
		files["CLAUDE.md"] = "ai/claude.md"
		files["AGENTS.md"] = "ai/agents.md"
	}
}

// addSandboxFiles adds sandbox-related files to the file mapping
func (r *Registry) addSandboxFiles(adl *schema.ADL, files map[string]string) {
	if adl.Spec.Development == nil || adl.Spec.Development.Sandbox == nil {
		return
	}

	sandbox := adl.Spec.Development.Sandbox

	if sandbox.Flox != nil && sandbox.Flox.Enabled {
		files[".flox/env/manifest.toml"] = "flox/manifest.toml"
		files[".flox/env.json"] = "flox/env.json"
		files[".flox/.gitignore"] = "flox/gitignore"
		files[".flox/.gitattributes"] = "flox/gitattributes"
	}

	if sandbox.DevContainer != nil && sandbox.DevContainer.Enabled {
		files[".devcontainer/devcontainer.json"] = "devcontainer/devcontainer.json"
	}
}

// ListTemplates returns a list of all loaded template keys
func (r *Registry) ListTemplates() []string {
	keys := make([]string, 0, len(r.templates))
	for k := range r.templates {
		keys = append(keys, k)
	}
	return keys
}

// addIssueTemplateFiles adds GitHub issue template files when enabled
func (r *Registry) addIssueTemplateFiles(adl *schema.ADL, files map[string]string) {
	if adl.Spec.SCM == nil || !adl.Spec.SCM.IssueTemplates {
		return
	}
	if adl.Spec.SCM.Provider == schema.SCMProviderGithub || adl.Spec.SCM.Provider == "" {
		files[".github/ISSUE_TEMPLATE/bug_report.md"] = "github/bug_report.md"
		files[".github/ISSUE_TEMPLATE/feature_request.md"] = "github/feature_request.md"
		files[".github/ISSUE_TEMPLATE/refactor_request.md"] = "github/refactor_request.md"
	}
}

// addDependabotFiles adds the GitHub Dependabot configuration when enabled.
// The generated manifest enumerates ecosystems based on the ADL spec
// (gomod/cargo/npm by language plus github-actions, docker, and
// devcontainers when applicable). Only emitted for the GitHub SCM
// provider - GitLab/Bitbucket equivalents are out of scope for now.
func (r *Registry) addDependabotFiles(adl *schema.ADL, files map[string]string) {
	if adl.Spec.SCM == nil || !adl.Spec.SCM.Dependabot {
		return
	}
	if adl.Spec.SCM.Provider == schema.SCMProviderGithub || adl.Spec.SCM.Provider == "" {
		files[".github/dependabot.yml"] = "github/dependabot.yaml"
	}
}

// DetectLanguageFromADL detects the programming language from ADL
func DetectLanguageFromADL(adl *schema.ADL) string {
	if adl.Spec.Language.Go != nil {
		return "go"
	}
	if adl.Spec.Language.Rust != nil {
		return "rust"
	}
	if adl.Spec.Language.TypeScript != nil {
		return "typescript"
	}
	return "go"
}
