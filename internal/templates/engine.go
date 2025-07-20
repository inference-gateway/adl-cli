package templates

import (
	"bytes"
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

// GetFiles returns the template files for the current template
func (e *Engine) GetFiles() map[string]string {
	switch e.templateName {
	case "minimal":
		return getMinimalTemplate()
	case "ai-powered":
		return getAIPoweredTemplate()
	case "enterprise":
		return getEnterpriseTemplate()
	default:
		return getAIPoweredTemplate() // default to ai-powered
	}
}
