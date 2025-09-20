package templates

import (
	"bytes"
	"encoding/json"
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

// getDefaultAcronyms returns the default acronyms map
func getDefaultAcronyms() map[string]string {
	return map[string]string{
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
		"mb":    "MB",
		"gb":    "GB",
		"kb":    "KB",
	}
}

// buildAcronymsMap builds the acronyms map from default + custom acronyms
func buildAcronymsMap(customAcronyms []string) map[string]string {
	acronyms := getDefaultAcronyms()

	for _, acronym := range customAcronyms {
		lowerAcronym := strings.ToLower(acronym)
		upperAcronym := strings.ToUpper(acronym)
		acronyms[lowerAcronym] = upperAcronym
	}

	return acronyms
}

// toPascalCaseWithAcronyms converts snake_case, dash-case, or camelCase to PascalCase with custom acronyms
func toPascalCaseWithAcronyms(s string, acronyms map[string]string) string {
	if !strings.Contains(s, "_") && !strings.Contains(s, "-") {
		s = camelToSnakeCase(s)
	}

	s = strings.ReplaceAll(s, "-", "_")
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

// toPascalCase converts snake_case, dash-case, or camelCase to PascalCase with default acronyms (backward compatibility)
func toPascalCase(s string) string {
	return toPascalCaseWithAcronyms(s, getDefaultAcronyms())
}

// toCamelCase converts snake_case to camelCase with special handling for acronyms
func toCamelCase(s string) string {
	pascalCase := toPascalCase(s)
	if len(pascalCase) == 0 {
		return pascalCase
	}
	runes := []rune(pascalCase)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// camelToSnakeCase converts camelCase to snake_case with proper acronym handling
func camelToSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}

	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			needsUnderscore := false

			if i > 1 && unicode.IsUpper(runes[i-1]) {
				if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					needsUnderscore = true
				}
			} else {
				needsUnderscore = true
			}

			if needsUnderscore {
				result.WriteRune('_')
			}
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// toSnakeCase converts dash-case to snake_case
func toSnakeCase(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

// toUpperSnakeCase converts camelCase, dash-case, or snake_case to UPPER_SNAKE_CASE
func toUpperSnakeCase(s string) string {
	if !strings.Contains(s, "_") && !strings.Contains(s, "-") {
		s = camelToSnakeCase(s)
	}

	s = strings.ReplaceAll(s, "-", "_")

	return strings.ToUpper(s)
}

// toUpperSnakeCaseWithAcronyms converts camelCase, dash-case, or snake_case to UPPER_SNAKE_CASE with acronym support
func toUpperSnakeCaseWithAcronyms(s string, acronyms map[string]string) string {
	if !strings.Contains(s, "_") && !strings.Contains(s, "-") {
		s = camelToSnakeCase(s)
	}
	s = strings.ReplaceAll(s, "-", "_")

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
			result[i] = strings.ToUpper(word)
		}
	}
	return strings.Join(result, "_")
}

// Context provides data for template execution
type Context struct {
	ADL             *schema.ADL
	Metadata        schema.GeneratedMetadata
	Language        string
	GenerateCI      bool
	GenerateCD      bool
	EnableAI        bool
	GenerateCommand string
	customAcronyms  map[string]string
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

// toJson converts a value to JSON string representation
func toJson(v interface{}) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

// toGoMap converts a value to Go map literal string representation
func toGoMap(v interface{}) string {
	return convertToGoMapLiteral(v)
}

// convertToGoMapLiteral recursively converts values to Go map literal format
func convertToGoMapLiteral(v interface{}) string {
	switch val := v.(type) {
	case map[string]interface{}:
		if len(val) == 0 {
			return "map[string]any{}"
		}
		result := "map[string]any{"
		first := true
		for k, v := range val {
			if !first {
				result += ", "
			}
			first = false
			result += fmt.Sprintf(`"%s": %s`, k, convertToGoMapLiteral(v))
		}
		result += "}"
		return result
	case []interface{}:
		if len(val) == 0 {
			return "[]string{}"
		}

		allStrings := true
		for _, item := range val {
			if _, ok := item.(string); !ok {
				allStrings = false
				break
			}
		}
		if allStrings {
			result := "[]string{"
			for i, item := range val {
				if i > 0 {
					result += ", "
				}
				result += fmt.Sprintf(`"%s"`, item.(string))
			}
			result += "}"
			return result
		}

		jsonBytes, _ := json.Marshal(val)
		return string(jsonBytes)
	case string:
		return fmt.Sprintf(`"%s"`, val)
	default:
		jsonBytes, _ := json.Marshal(val)
		return string(jsonBytes)
	}
}

// findDependencyByID finds a dependency by name in the dependencies slice
func findDependencyByID(id string, deps []string) int {
	for i, dep := range deps {
		if dep == id {
			return i
		}
	}
	return -1
}

// customFuncMap returns a function map with Sprig functions plus custom functions
func customFuncMap() template.FuncMap {
	funcMap := sprig.TxtFuncMap()
	funcMap["toPascalCase"] = toPascalCase
	funcMap["toCamelCase"] = toCamelCase
	funcMap["toSnakeCase"] = toSnakeCase
	funcMap["toUpperCase"] = strings.ToUpper
	funcMap["toUpperSnakeCase"] = toUpperSnakeCase
	funcMap["toJson"] = toJson
	funcMap["toGoMap"] = toGoMap
	funcMap["findDependencyByID"] = findDependencyByID
	return funcMap
}

// customFuncMapWithAcronyms returns a function map with context-aware acronym functions
func customFuncMapWithAcronyms(acronyms map[string]string) template.FuncMap {
	funcMap := sprig.TxtFuncMap()

	funcMap["toPascalCase"] = func(s string) string {
		return toPascalCaseWithAcronyms(s, acronyms)
	}
	funcMap["toCamelCase"] = func(s string) string {
		pascalCase := toPascalCaseWithAcronyms(s, acronyms)
		if len(pascalCase) == 0 {
			return pascalCase
		}
		runes := []rune(pascalCase)
		runes[0] = unicode.ToLower(runes[0])
		return string(runes)
	}

	funcMap["toSnakeCase"] = toSnakeCase
	funcMap["toUpperCase"] = strings.ToUpper
	funcMap["toUpperSnakeCase"] = func(s string) string {
		return toUpperSnakeCaseWithAcronyms(s, acronyms)
	}
	funcMap["toJson"] = toJson
	funcMap["toGoMap"] = toGoMap
	funcMap["findDependencyByID"] = findDependencyByID
	return funcMap
}

// prepareContext prepares the context with custom acronyms
func (e *Engine) prepareContext(ctx Context) Context {
	var customAcronyms []string
	if ctx.ADL != nil && ctx.ADL.Spec.Acronyms != nil {
		customAcronyms = ctx.ADL.Spec.Acronyms
	}

	ctx.customAcronyms = buildAcronymsMap(customAcronyms)
	return ctx
}

// Execute executes a template with the given context
func (e *Engine) Execute(templateContent string, ctx Context) (string, error) {
	ctx = e.prepareContext(ctx)

	tmpl, err := template.New("template").Funcs(customFuncMapWithAcronyms(ctx.customAcronyms)).Parse(templateContent)
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

// ExecuteToolTemplate executes a skill template with skill-specific data
func (e *Engine) ExecuteToolTemplate(templateKey string, skillData any) (string, error) {
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
	if err := tmpl.Execute(&buf, skillData); err != nil {
		return "", err
	}

	result := buf.String()

	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result, nil
}

// ExecuteToolTemplateWithContext executes a skill template with ADL context for custom acronyms
func (e *Engine) ExecuteToolTemplateWithContext(templateKey string, skillData any, ctx Context) (string, error) {
	if e.registry == nil {
		return "", fmt.Errorf("no registry configured")
	}

	templateContent, err := e.registry.GetTemplate(templateKey)
	if err != nil {
		return "", fmt.Errorf("failed to get template %s: %w", templateKey, err)
	}

	ctx = e.prepareContext(ctx)

	tmpl, err := template.New("template").Funcs(customFuncMapWithAcronyms(ctx.customAcronyms)).Parse(templateContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, skillData); err != nil {
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
