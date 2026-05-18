package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// schemaBytes holds the canonical ADL JSON Schema, vendored from
// github.com/inference-gateway/adl at the version pinned in Taskfile.yml
// (ADL_SCHEMA_VERSION). Refresh with `task fetch-schema`.
//
//go:embed schema.json
var schemaBytes []byte

// Validator validates ADL files against the schema
type Validator struct {
	schema *gojsonschema.Schema
}

// NewValidator creates a new validator with the embedded schema
func NewValidator() *Validator {
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		panic(fmt.Sprintf("failed to load ADL schema: %v", err))
	}

	return &Validator{
		schema: schema,
	}
}

// ValidateFile validates an ADL file
func (v *Validator) ValidateFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var yamlData any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonData)
	result, err := v.schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return fmt.Errorf("validation failed:\n- %s", fmt.Sprintf("\n- %s", errors))
	}

	// Additional validation: check that injected services are defined
	var adl ADL
	if err := yaml.Unmarshal(data, &adl); err != nil {
		return fmt.Errorf("failed to parse ADL for service validation: %w", err)
	}

	if err := v.validateServiceReferences(&adl); err != nil {
		return fmt.Errorf("service validation failed: %w", err)
	}

	return nil
}

// validateServiceReferences checks that all injected services are defined in the spec
func (v *Validator) validateServiceReferences(adl *ADL) error {
	definedServices := make(map[string]bool)
	for serviceName := range adl.Spec.Services {
		definedServices[serviceName] = true
	}

	definedServices["logger"] = true
	definedServices["config"] = true

	definedConfigSections := make(map[string]bool)
	for configSection := range adl.Spec.Config {
		definedConfigSections[configSection] = true
	}

	for _, skill := range adl.Spec.Skills {
		for _, injectedService := range skill.Inject {
			if len(injectedService) > 7 && injectedService[:7] == "config." {
				configSection := injectedService[7:]
				if !definedConfigSections[configSection] {
					return fmt.Errorf("skill '%s' injects config section '%s' that is not defined in spec.config", skill.ID, configSection)
				}
			} else if !definedServices[injectedService] {
				return fmt.Errorf("skill '%s' injects service '%s' that is not defined in spec.services", skill.ID, injectedService)
			}
		}
	}

	return nil
}
