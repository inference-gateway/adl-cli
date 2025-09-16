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

	// Additional validation: check that injected dependencies are defined
	var adl ADL
	if err := yaml.Unmarshal(data, &adl); err != nil {
		return fmt.Errorf("failed to parse ADL for dependency validation: %w", err)
	}

	if err := v.validateDependencyReferences(&adl); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	return nil
}

// validateDependencyReferences checks that all injected dependencies are defined in the spec
func (v *Validator) validateDependencyReferences(adl *ADL) error {
	definedDeps := make(map[string]bool)
	for depName := range adl.Spec.Dependencies {
		definedDeps[depName] = true
	}

	definedDeps["logger"] = true

	for _, skill := range adl.Spec.Skills {
		for _, injectedDep := range skill.Inject {
			if !definedDeps[injectedDep] {
				return fmt.Errorf("skill '%s' injects dependency '%s' that is not defined in spec.dependencies", skill.ID, injectedDep)
			}
		}
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
        "card": {
          "type": "object",
          "properties": {
            "protocolVersion": {
              "type": "string"
            },
            "url": {
              "type": "string"
            },
            "preferredTransport": {
              "type": "string"
            },
            "defaultInputModes": {
              "type": "array",
              "items": {
                "type": "string"
              }
            },
            "defaultOutputModes": {
              "type": "array",
              "items": {
                "type": "string"
              }
            },
            "documentationUrl": {
              "type": "string"
            },
            "iconUrl": {
              "type": "string"
            }
          }
        },
        "agent": {
          "type": "object",
          "properties": {
            "provider": {
              "type": "string",
              "enum": ["", "openai", "anthropic", "ollama", "deepseek", "google", "mistral", "groq"]
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
        "config": {
          "type": "object",
          "additionalProperties": {
            "type": "object",
            "additionalProperties": true
          }
        },
        "dependencies": {
          "type": "object",
          "additionalProperties": {
            "type": "object",
            "required": ["type", "interface", "factory", "description"],
            "properties": {
              "type": {
                "type": "string",
                "enum": ["service", "repository", "client", "middleware"]
              },
              "interface": {
                "type": "string",
                "pattern": "^[a-zA-Z][a-zA-Z0-9_]*$"
              },
              "factory": {
                "type": "string",
                "pattern": "^[a-zA-Z][a-zA-Z0-9_]*$"
              },
              "description": {
                "type": "string"
              }
            }
          }
        },
        "acronyms": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "skills": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["id", "name", "description", "tags", "schema"],
            "properties": {
              "id": {
                "type": "string",
                "pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
              },
              "name": {
                "type": "string"
              },
              "description": {
                "type": "string"
              },
              "tags": {
                "type": "array",
                "items": {
                  "type": "string"
                }
              },
              "examples": {
                "type": "array",
                "items": {
                  "type": "string"
                }
              },
              "inputModes": {
                "type": "array",
                "items": {
                  "type": "string"
                }
              },
              "outputModes": {
                "type": "array",
                "items": {
                  "type": "string"
                }
              },
              "schema": {
                "type": "object",
                "required": ["type"],
                "properties": {
                  "type": {
                    "type": "string",
                    "enum": ["object", "array", "string", "number", "integer", "boolean", "null"]
                  },
                  "properties": {
                    "type": "object",
                    "additionalProperties": {
                      "type": "object",
                      "properties": {
                        "type": {
                          "type": "string",
                          "enum": ["object", "array", "string", "number", "integer", "boolean", "null"]
                        },
                        "description": {
                          "type": "string"
                        },
                        "enum": {
                          "type": "array"
                        },
                        "format": {
                          "type": "string"
                        },
                        "minimum": {
                          "type": "number"
                        },
                        "maximum": {
                          "type": "number"
                        }
                      }
                    }
                  },
                  "required": {
                    "type": "array",
                    "items": {
                      "type": "string"
                    }
                  },
                  "additionalProperties": {
                    "type": "boolean"
                  }
                }
              },
              "implementation": {
                "type": "string"
              },
              "inject": {
                "type": "array",
                "items": {
                  "type": "string",
                  "pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
                }
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
            },
            "rust": {
              "type": "object",
              "required": ["packageName", "version", "edition"],
              "properties": {
                "packageName": {
                  "type": "string"
                },
                "version": {
                  "type": "string"
                },
                "edition": {
                  "type": "string"
                }
              }
            }
          },
          "minProperties": 1
        },
        "scm": {
          "type": "object",
          "properties": {
            "provider": {
              "type": "string",
              "enum": ["github", "gitlab", "bitbucket"]
            },
            "url": {
              "type": "string"
            }
          }
        },
        "sandbox": {
          "type": "object",
          "properties": {
            "type": {
              "type": "string",
              "enum": ["flox", "devcontainer"]
            },
            "flox": {
              "type": "object",
              "properties": {
                "enabled": {
                  "type": "boolean"
                }
              },
              "required": ["enabled"]
            },
            "devcontainer": {
              "type": "object",
              "properties": {
                "enabled": {
                  "type": "boolean"
                }
              },
              "required": ["enabled"]
            }
          }
        },
        "deployment": {
          "type": "object",
          "properties": {
            "type": {
              "type": "string",
              "enum": ["kubernetes"]
            }
          }
        }
      }
    }
  }
}`
