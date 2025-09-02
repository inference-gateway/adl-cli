package templates

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/inference-gateway/adl-cli/internal/schema"
)

// Engine handles template execution
type Engine struct {
	templateName string
	registry     *Registry
}

// toPascalCase converts snake_case to PascalCase with special handling for acronyms
func toPascalCase(s string) string {
	acronyms := map[string]string{
		"id":    "ID",
		"api":   "API",
		"url":   "URL",
		"uri":   "URI",
		"http":  "HTTP",
		"https": "HTTPS",
		"json":  "JSON",
		"xml":   "XML",
		"sql":   "SQL",
		"html":  "HTML",
		"css":   "CSS",
		"js":    "JS",
		"ui":    "UI",
		"uuid":  "UUID",
		"tcp":   "TCP",
		"udp":   "UDP",
		"ip":    "IP",
		"dns":   "DNS",
		"tls":   "TLS",
		"ssl":   "SSL",
		"cpu":   "CPU",
		"gpu":   "GPU",
		"ram":   "RAM",
		"io":    "IO",
		"os":    "OS",
		"db":    "DB",
	}

	words := strings.Split(s, "_")
	result := make([]string, len(words))

	for i, word := range words {
		if len(word) == 0 {
			continue
		}

		lowerWord := strings.ToLower(word)
		if acronym, exists := acronyms[lowerWord]; exists {
			result[i] = acronym
		} else {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			result[i] = string(runes)
		}
	}

	return strings.Join(result, "")
}

// Context provides data for template execution
type Context struct {
	ADL        *schema.ADL
	Metadata   schema.GeneratedMetadata
	Language   string
	GenerateCI bool
}

// New creates a new template engine
func New(templateName string) *Engine {
	return &Engine{
		templateName: templateName,
	}
}

// NewWithRegistry creates a new template engine with a registry
func NewWithRegistry(templateName string, registry *Registry) *Engine {
	return &Engine{
		templateName: templateName,
		registry:     registry,
	}
}

// customFuncMap returns a function map with Sprig functions plus custom functions
func customFuncMap() template.FuncMap {
	funcMap := sprig.TxtFuncMap()
	funcMap["toPascalCase"] = toPascalCase
	return funcMap
}

// Execute executes a template with the given context
func (e *Engine) Execute(templateContent string, ctx Context) (string, error) {
	tmpl, err := template.New("template").Funcs(customFuncMap()).Parse(templateContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	result := buf.String()

	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result, nil
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
	case ext == ".rs":
		fileType = "rust"
	case ext == ".yaml" || ext == ".yml":
		fileType = "yaml"
	case baseName == "dockerfile":
		fileType = "dockerfile"
	case baseName == "taskfile.yml":
		fileType = "taskfile"
	default:
		return content, nil
	}

	header := GetGeneratedFileHeader(fileType, ctx.Metadata.CLIVersion, ctx.Metadata.GeneratedAt)

	return header + content, nil
}

// GetFiles returns the template files for the current template with ADL context
func (e *Engine) GetFiles(adl *schema.ADL) map[string]string {
	if e.registry == nil {
		return make(map[string]string)
	}
	return e.registry.GetFiles(adl)
}

// ExecuteTemplate executes a template from the registry with the given context
func (e *Engine) ExecuteTemplate(templateKey string, ctx Context) (string, error) {
	if e.registry == nil {
		return "", fmt.Errorf("no registry configured")
	}

	templateContent, err := e.registry.GetTemplate(templateKey)
	if err != nil {
		return "", fmt.Errorf("failed to get template %s: %w", templateKey, err)
	}

	return e.Execute(templateContent, ctx)
}

// ExecuteToolTemplate executes a tool template with tool-specific data
func (e *Engine) ExecuteToolTemplate(templateKey string, toolData any) (string, error) {
	if e.registry == nil {
		return "", fmt.Errorf("no registry configured")
	}

	templateContent, err := e.registry.GetTemplate(templateKey)
	if err != nil {
		return "", fmt.Errorf("failed to get template %s: %w", templateKey, err)
	}

	tmpl, err := template.New("template").Funcs(customFuncMap()).Parse(templateContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, toolData); err != nil {
		return "", err
	}

	result := buf.String()

	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result, nil
}

// GetTemplate returns the template name
func (e *Engine) GetTemplate() string {
	return e.templateName
}
