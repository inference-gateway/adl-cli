package templates

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/inference-gateway/a2a-cli/internal/schema"
)

// Engine handles template execution
type Engine struct {
	templateName string
}

// Context provides data for template execution
type Context struct {
	ADL      *schema.ADL
	Metadata schema.GeneratedMetadata
}

// New creates a new template engine
func New(templateName string) *Engine {
	return &Engine{
		templateName: templateName,
	}
}

// Execute executes a template with the given context
func (e *Engine) Execute(templateContent string, ctx Context) (string, error) {
	tmpl, err := template.New("template").Funcs(sprig.TxtFuncMap()).Parse(templateContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecuteWithHeader executes a template with the given context and adds a header if needed
func (e *Engine) ExecuteWithHeader(templateContent string, ctx Context, fileName string) (string, error) {
	content, err := e.Execute(templateContent, ctx)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	baseName := strings.ToLower(filepath.Base(fileName))

	var fileType string
	switch {
	case ext == ".go":
		fileType = "go"
	case ext == ".yaml" || ext == ".yml":
		fileType = "yaml"
	case baseName == "dockerfile":
		fileType = "dockerfile"
	case baseName == "taskfile.yml":
		fileType = "taskfile"
	default:
		return content, nil
	}

	header := getGeneratedFileHeader(fileType, ctx.Metadata.CLIVersion, ctx.Metadata.GeneratedAt)

	return header + content, nil
}

// GetFiles returns the template files for the current template with ADL context
func (e *Engine) GetFiles(adl *schema.ADL) map[string]string {
	return GetMinimalTemplate(adl)
}

// GetTemplate returns the template name
func (e *Engine) GetTemplate() string {
	return e.templateName
}
