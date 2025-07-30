package schema

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// Validator validates ADL files against the schema
type Validator struct {
	schema *gojsonschema.Schema
}

// NewValidator creates a new validator with the embedded schema
func NewValidator() *Validator {
	schemaLoader := gojsonschema.NewStringLoader(adlSchema)
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

	var yamlData interface{}
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

	return nil
}

const adlSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Agent Definition Language (ADL)",
  "type": "object",
  "required": ["apiVersion", "kind", "metadata", "spec"],
  "properties": {
    "apiVersion": {
      "type": "string",
      "const": "adl.dev/v1"
    },
    "kind": {
      "type": "string",
      "const": "Agent"
    },
    "metadata": {
      "type": "object",
      "required": ["name", "description", "version"],
      "properties": {
        "name": {
          "type": "string",
          "pattern": "^[a-z0-9-]+$"
        },
        "description": {
          "type": "string"
        },
        "version": {
          "type": "string",
          "pattern": "^\\d+\\.\\d+\\.\\d+$"
        }
      }
    },
    "spec": {
      "type": "object",
      "required": ["server", "capabilities", "language"],
      "properties": {
        "capabilities": {
          "type": "object",
          "required": ["streaming", "pushNotifications", "stateTransitionHistory"],
          "properties": {
            "streaming": {
              "type": "boolean"
            },
            "pushNotifications": {
              "type": "boolean"
            },
            "stateTransitionHistory": {
              "type": "boolean"
            }
          }
        },
        "agent": {
          "type": "object",
          "required": ["provider"],
          "properties": {
            "provider": {
              "type": "string",
              "enum": ["openai", "anthropic", "ollama", "azure", "deepseek", "none"]
            },
            "model": {
              "type": "string"
            },
            "systemPrompt": {
              "type": "string"
            },
            "maxTokens": {
              "type": "integer",
              "minimum": 1
            },
            "temperature": {
              "type": "number",
              "minimum": 0,
              "maximum": 2
            }
          }
        },
        "tools": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["name", "description", "schema"],
            "properties": {
              "name": {
                "type": "string",
                "pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
              },
              "description": {
                "type": "string"
              },
              "schema": {
                "type": "object"
              },
              "implementation": {
                "type": "string"
              }
            }
          }
        },
        "server": {
          "type": "object",
          "required": ["port"],
          "properties": {
            "port": {
              "type": "integer",
              "minimum": 1,
              "maximum": 65535
            },
            "debug": {
              "type": "boolean"
            },
            "auth": {
              "type": "object",
              "properties": {
                "enabled": {
                  "type": "boolean"
                }
              }
            }
          }
        },
        "language": {
          "type": "object",
          "properties": {
            "go": {
              "type": "object",
              "required": ["module", "version"],
              "properties": {
                "module": {
                  "type": "string"
                },
                "version": {
                  "type": "string"
                }
              }
            },
            "typescript": {
              "type": "object",
              "required": ["packageName", "nodeVersion"],
              "properties": {
                "packageName": {
                  "type": "string"
                },
                "nodeVersion": {
                  "type": "string"
                }
              }
            }
          },
          "minProperties": 1,
          "maxProperties": 1
        }
      }
    }
  }
}`
