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

// ValidateFile validates an ADL file. Returns any non-fatal warnings
// that callers should surface to the user (e.g., a skills-using agent
// that hasn't enabled the Read built-in). A nil error means the manifest
// is structurally valid; warnings may still be present.
func (v *Validator) ValidateFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var yamlData any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Reject pre-v0.6.0 manifests up front with a clear migration hint.
	// JSON Schema lets unknown properties slip through Spec, so without
	// this guard old `spec.sandbox` / `spec.ai` blocks would be silently
	// dropped at unmarshal time and the agent would generate without
	// sandboxes or AI docs.
	if err := checkLegacySpecFields(yamlData); err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to JSON: %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonData)
	result, err := v.schema.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return nil, fmt.Errorf("validation failed:\n- %s", fmt.Sprintf("\n- %s", errors))
	}

	// Additional validation: check that injected services are defined
	var adl ADL
	if err := yaml.Unmarshal(data, &adl); err != nil {
		return nil, fmt.Errorf("failed to parse ADL for service validation: %w", err)
	}

	if err := v.validateTools(&adl); err != nil {
		return nil, fmt.Errorf("tool validation failed: %w", err)
	}

	warnings, err := v.validateSkills(&adl)
	if err != nil {
		return nil, fmt.Errorf("skill validation failed: %w", err)
	}

	return warnings, nil
}

// checkLegacySpecFields rejects manifests that still use the pre-v0.6.0
// shape where `sandbox` and `ai` sat directly under `spec` (in v0.6.0
// they were grouped under `spec.development.{sandbox,ai}`) and the
// pre-v0.8.0 single-flag AI shape `spec.development.ai.enabled: true`
// (replaced by per-agent toggles claudecode/codex/gemini/opencode/infer).
// JSON Schema's default additionalProperties:true would otherwise silently
// drop these unknown fields, so we surface them as errors before they
// confuse the user.
func checkLegacySpecFields(yamlData any) error {
	root, ok := yamlData.(map[string]any)
	if !ok {
		return nil
	}
	spec, ok := root["spec"].(map[string]any)
	if !ok {
		return nil
	}

	var legacyV6 []string
	if _, exists := spec["sandbox"]; exists {
		legacyV6 = append(legacyV6, "spec.sandbox -> spec.development.sandbox")
	}
	if _, exists := spec["ai"]; exists {
		legacyV6 = append(legacyV6, "spec.ai -> spec.development.ai")
	}
	if len(legacyV6) > 0 {
		return fmt.Errorf("manifest uses pre-v0.6.0 schema fields; move them under spec.development:\n  - %s\nThe ADL schema is pinned at v0.6.0+ (see https://github.com/inference-gateway/adl/releases/tag/v0.6.0)",
			joinWithIndent(legacyV6, "\n  - "))
	}

	if dev, ok := spec["development"].(map[string]any); ok {
		if ai, ok := dev["ai"].(map[string]any); ok {
			if _, exists := ai["enabled"]; exists {
				return fmt.Errorf("manifest uses the pre-v0.8.0 single-flag AI shape `spec.development.ai.enabled`; this field was removed in ADL v0.8.0. Move it to a per-agent toggle under spec.development.ai, e.g.:\n\n  spec:\n    development:\n      ai:\n        claudecode:\n          enabled: true   # generates CLAUDE.md + .github/workflows/claude.yml\n        # codex / gemini / opencode / infer are independent toggles\n\nSee https://github.com/inference-gateway/adl/releases/tag/v0.8.0 for the full per-agent matrix")
			}
		}
	}

	return nil
}

// joinWithIndent is a tiny helper so we don't pull in strings just for one
// formatted error.
func joinWithIndent(items []string, sep string) string {
	out := ""
	for i, item := range items {
		if i > 0 {
			out += sep
		}
		out += item
	}
	return out
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

// validateSkills enforces bare-skill metadata and surfaces non-fatal
// warnings about the skills-need-read contract: a skills-using agent
// should list `- id: read` AND enable `spec.config.tools.read.enabled:
// true`, otherwise it can't load SKILL.md bodies at runtime. Returns
// warnings (not errors) for the read-tool case so partially configured
// manifests still pass validation - the warning prompts the user to fix
// it.
func (v *Validator) validateSkills(adl *ADL) ([]string, error) {
	for _, skill := range adl.Spec.Skills {
		if skill.Bare {
			if skill.Name == "" {
				return nil, fmt.Errorf("skill '%s' has bare: true but is missing name", skill.ID)
			}
			if skill.Description == "" {
				return nil, fmt.Errorf("skill '%s' has bare: true but is missing description", skill.ID)
			}
		}
	}

	if len(adl.Spec.Skills) == 0 || adl.Spec.Agent == nil {
		return nil, nil
	}

	hasReadTool := false
	for _, tool := range adl.Spec.Tools {
		if tool.ID == string(ReservedToolRead) {
			hasReadTool = true
			break
		}
	}
	if !hasReadTool {
		return []string{
			"spec.skills is non-empty but spec.tools is missing '- id: read'; the AVAILABLE SKILLS manifest will be added to the system prompt but the agent has no Read built-in to load SKILL.md bodies. Add '- id: read' to spec.tools and set spec.config.tools.read.enabled: true.",
		}, nil
	}

	toolsCfg := adl.Spec.Config[reservedConfigSection]
	readRaw, ok := toolsCfg[string(ReservedToolRead)]
	if !ok {
		return []string{
			"spec.skills is non-empty and '- id: read' is listed, but spec.config.tools.read is missing; the Read built-in will register in a disabled state and fail at runtime. Set spec.config.tools.read.enabled: true.",
		}, nil
	}
	decoded, err := DecodeBuiltinToolConfig(string(ReservedToolRead), readRaw)
	if err != nil {
		return nil, err
	}
	readCfg, ok := decoded.(*ReadBuiltinConfig)
	if !ok || !readCfg.Enabled {
		return []string{
			"spec.skills is non-empty but spec.config.tools.read.enabled is not true; the Read built-in will register in a disabled state and fail at runtime. Set spec.config.tools.read.enabled: true so the agent can load SKILL.md bodies.",
		}, nil
	}

	return nil, nil
}
