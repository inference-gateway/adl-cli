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

	if err := v.validateTools(&adl); err != nil {
		return fmt.Errorf("tool validation failed: %w", err)
	}

	if err := v.validateSkills(&adl); err != nil {
		return fmt.Errorf("skill validation failed: %w", err)
	}

	return nil
}

// reservedConfigSection is the namespace inside spec.config dedicated to
// built-in tool config (spec.config.tools.<id>). User-defined services
// cannot inject from it.
const reservedConfigSection = "tools"

// validateTools enforces the contract for both user-defined and reserved
// (built-in) tool IDs and the integrity of spec.config.tools.<id>.
func (v *Validator) validateTools(adl *ADL) error {
	definedServices := map[string]bool{"logger": true, "config": true}
	for serviceName := range adl.Spec.Services {
		definedServices[serviceName] = true
	}

	definedConfigSections := make(map[string]bool)
	for configSection := range adl.Spec.Config {
		definedConfigSections[configSection] = true
	}

	toolsCfg := adl.Spec.Config[reservedConfigSection]

	for _, tool := range adl.Spec.Tools {
		if IsReservedToolID(tool.ID) {
			if tool.Name != "" {
				return fmt.Errorf("reserved tool '%s' must not set 'name' (the generator supplies it)", tool.ID)
			}
			if tool.Description != "" {
				return fmt.Errorf("reserved tool '%s' must not set 'description' (the generator supplies it)", tool.ID)
			}
			if len(tool.Schema) > 0 {
				return fmt.Errorf("reserved tool '%s' must not set 'schema' (the generator supplies it)", tool.ID)
			}
			if len(tool.Inject) > 0 {
				return fmt.Errorf("reserved tool '%s' must not set 'inject' (built-ins do not use service injection)", tool.ID)
			}
			if raw, ok := toolsCfg[tool.ID]; ok {
				if _, err := DecodeBuiltinToolConfig(tool.ID, raw); err != nil {
					return err
				}
			}
			continue
		}

		if tool.Name == "" {
			return fmt.Errorf("tool '%s' must set 'name' (only reserved built-in IDs may omit it)", tool.ID)
		}
		if tool.Description == "" {
			return fmt.Errorf("tool '%s' must set 'description'", tool.ID)
		}
		if len(tool.Tags) == 0 {
			return fmt.Errorf("tool '%s' must set 'tags' (at least one)", tool.ID)
		}
		if len(tool.Schema) == 0 {
			return fmt.Errorf("tool '%s' must set 'schema'", tool.ID)
		}

		for _, injectedService := range tool.Inject {
			if len(injectedService) > 7 && injectedService[:7] == "config." {
				configSection := injectedService[7:]
				if configSection == reservedConfigSection {
					return fmt.Errorf("tool '%s' injects reserved namespace 'config.tools'; this namespace is owned by built-in tools and cannot be inject-referenced", tool.ID)
				}
				if !definedConfigSections[configSection] {
					return fmt.Errorf("tool '%s' injects config section '%s' that is not defined in spec.config", tool.ID, configSection)
				}
			} else if !definedServices[injectedService] {
				return fmt.Errorf("tool '%s' injects service '%s' that is not defined in spec.services", tool.ID, injectedService)
			}
		}
	}

	return nil
}

// validateSkills enforces bare-skill metadata and the skills-need-read
// contract: any skills-using agent must list `- id: read` AND enable
// `spec.config.tools.read.enabled: true`, otherwise the agent has no way
// to load SKILL.md bodies at runtime.
func (v *Validator) validateSkills(adl *ADL) error {
	for _, skill := range adl.Spec.Skills {
		if skill.Bare {
			if skill.Name == "" {
				return fmt.Errorf("skill '%s' has bare: true but is missing name", skill.ID)
			}
			if skill.Description == "" {
				return fmt.Errorf("skill '%s' has bare: true but is missing description", skill.ID)
			}
		}
	}

	if len(adl.Spec.Skills) == 0 || adl.Spec.Agent == nil {
		return nil
	}

	hasReadTool := false
	for _, tool := range adl.Spec.Tools {
		if tool.ID == string(ReservedToolRead) {
			hasReadTool = true
			break
		}
	}
	if !hasReadTool {
		return fmt.Errorf("spec.skills is non-empty but spec.tools is missing '- id: read'; the agent needs the Read built-in to load SKILL.md bodies on demand")
	}

	toolsCfg := adl.Spec.Config[reservedConfigSection]
	readRaw, ok := toolsCfg[string(ReservedToolRead)]
	if !ok {
		return fmt.Errorf("spec.skills is non-empty and '- id: read' is listed, but spec.config.tools.read is missing; set spec.config.tools.read.enabled: true")
	}
	decoded, err := DecodeBuiltinToolConfig(string(ReservedToolRead), readRaw)
	if err != nil {
		return err
	}
	readCfg, ok := decoded.(*ReadBuiltinConfig)
	if !ok || !readCfg.Enabled {
		return fmt.Errorf("spec.skills is non-empty but spec.config.tools.read.enabled is not true; set spec.config.tools.read.enabled: true so the agent can load SKILL.md bodies")
	}

	return nil
}
