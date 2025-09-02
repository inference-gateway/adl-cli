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
}

// Embed template files at compile time
//
//go:embed languages/*/*.tmpl common/*/*.tmpl sandbox/*/*.tmpl
var templateFS embed.FS

// NewRegistry creates a new template registry for the specified language
func NewRegistry(language string) (*Registry, error) {
	r := &Registry{
		templates: make(map[string]string),
		language:  language,
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
		"main.go":                "main.go",
		"go.mod":                 "go.mod",
		".well-known/agent.json": "config/agent.json",
		"Taskfile.yml":           "ci/taskfile.yml",
		"Dockerfile":             "docker/dockerfile.go",
		".gitignore":             "config/gitignore",
		".gitattributes":         "config/gitattributes",
		".editorconfig":          "config/editorconfig",
		"README.md":              "docs/README.md",
	}

	if adl.Spec.Deployment != nil && adl.Spec.Deployment.Type != "" {
		switch adl.Spec.Deployment.Type {
		case "kubernetes":
			files["k8s/deployment.yaml"] = "kubernetes/deployment.yaml"
		}
	}

	for _, skill := range adl.Spec.Skills {
		files[fmt.Sprintf("tools/%s.go", skill.Name)] = "tools.go"
	}

	r.addSandboxFiles(adl, files)

	return files
}

// getRustFiles returns the file mapping for Rust projects
func (r *Registry) getRustFiles(adl *schema.ADL) map[string]string {
	files := map[string]string{
		"src/main.rs":            "main.rs",
		"Cargo.toml":             "Cargo.toml",
		".well-known/agent.json": "config/agent.json",
		"Taskfile.yml":           "ci/taskfile.yml",
		"Dockerfile":             "docker/dockerfile.rust",
		".gitignore":             "config/gitignore",
		".gitattributes":         "config/gitattributes",
		".editorconfig":          "config/editorconfig",
		"README.md":              "docs/README.md",
	}

	if adl.Spec.Deployment != nil && adl.Spec.Deployment.Type != "" {
		switch adl.Spec.Deployment.Type {
		case "kubernetes":
			files["k8s/deployment.yaml"] = "kubernetes/deployment.yaml"
		}
	}

	for _, skill := range adl.Spec.Skills {
		files[fmt.Sprintf("src/tools/%s.rs", skill.Name)] = "tools.rs"
	}

	if len(adl.Spec.Skills) > 0 {
		files["src/tools/mod.rs"] = "tools.mod.rs"
	}

	r.addSandboxFiles(adl, files)

	return files
}

// getTypeScriptFiles returns the file mapping for TypeScript projects
func (r *Registry) getTypeScriptFiles(adl *schema.ADL) map[string]string {
	files := map[string]string{
		"src/index.ts":           "index.ts",
		"package.json":           "package.json",
		"tsconfig.json":          "tsconfig.json",
		".well-known/agent.json": "config/agent.json",
		"Taskfile.yml":           "ci/taskfile.yml",
		"Dockerfile":             "docker/dockerfile.ts",
		".gitignore":             "config/gitignore",
		".gitattributes":         "config/gitattributes",
		".editorconfig":          "config/editorconfig",
		"README.md":              "docs/README.md",
	}

	if adl.Spec.Deployment != nil && adl.Spec.Deployment.Type != "" {
		switch adl.Spec.Deployment.Type {
		case "kubernetes":
			files["k8s/deployment.yaml"] = "kubernetes/deployment.yaml"
		}
	}

	for _, skill := range adl.Spec.Skills {
		files[fmt.Sprintf("src/tools/%s.ts", skill.Name)] = "tools.ts"
	}

	r.addSandboxFiles(adl, files)

	return files
}

// addSandboxFiles adds sandbox-related files to the file mapping
func (r *Registry) addSandboxFiles(adl *schema.ADL, files map[string]string) {
	if adl.Spec.Sandbox == nil {
		return
	}

	if adl.Spec.Sandbox.Flox != nil && adl.Spec.Sandbox.Flox.Enabled {
		files[".flox/env/manifest.toml"] = "flox/manifest.toml"
		files[".flox/env.json"] = "flox/env.json"
		files[".flox/.gitignore"] = "flox/gitignore"
		files[".flox/.gitattributes"] = "flox/gitattributes"
	}

	if adl.Spec.Sandbox.DevContainer != nil && adl.Spec.Sandbox.DevContainer.Enabled {
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
