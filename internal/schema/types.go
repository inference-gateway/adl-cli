package schema

import "time"

// ADL represents the complete Agent Definition Language structure
type ADL struct {
	APIVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
	Spec       Spec     `yaml:"spec" json:"spec"`
}

// Metadata contains agent metadata
type Metadata struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Version     string `yaml:"version" json:"version"`
}

// Spec contains the agent specification
type Spec struct {
	Capabilities *Capabilities `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Agent        *Agent        `yaml:"agent,omitempty" json:"agent,omitempty"`
	Tools        []Tool        `yaml:"tools,omitempty" json:"tools,omitempty"`
	Server       Server        `yaml:"server" json:"server"`
	Language     *Language     `yaml:"language,omitempty" json:"language,omitempty"`
}

// Capabilities defines what the agent can do
type Capabilities struct {
	Streaming              bool `yaml:"streaming" json:"streaming"`
	PushNotifications      bool `yaml:"pushNotifications" json:"pushNotifications"`
	StateTransitionHistory bool `yaml:"stateTransitionHistory" json:"stateTransitionHistory"`
}

// Agent configuration for AI providers
type Agent struct {
	Provider     string  `yaml:"provider" json:"provider"`
	Model        string  `yaml:"model" json:"model"`
	SystemPrompt string  `yaml:"systemPrompt" json:"systemPrompt"`
	MaxTokens    int     `yaml:"maxTokens" json:"maxTokens"`
	Temperature  float64 `yaml:"temperature" json:"temperature"`
}

// Tool represents a function the agent can call
type Tool struct {
	Name           string                 `yaml:"name" json:"name"`
	Description    string                 `yaml:"description" json:"description"`
	Schema         map[string]interface{} `yaml:"schema" json:"schema"`
	Implementation string                 `yaml:"implementation,omitempty" json:"implementation,omitempty"`
}

// Server configuration
type Server struct {
	Port  int         `yaml:"port" json:"port"`
	Debug bool        `yaml:"debug" json:"debug"`
	Auth  *AuthConfig `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// AuthConfig for server authentication
type AuthConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// GoConfig for Go-specific settings
type GoConfig struct {
	Module  string `yaml:"module" json:"module"`
	Version string `yaml:"version" json:"version"`
}

// TypeScriptConfig for TypeScript-specific settings
type TypeScriptConfig struct {
	PackageName string `yaml:"packageName" json:"packageName"`
	NodeVersion string `yaml:"nodeVersion" json:"nodeVersion"`
}

// Language configuration for different programming languages
type Language struct {
	Go         *GoConfig         `yaml:"go,omitempty" json:"go,omitempty"`
	TypeScript *TypeScriptConfig `yaml:"typescript,omitempty" json:"typescript,omitempty"`
}

// GeneratedMetadata contains information about the generation
type GeneratedMetadata struct {
	GeneratedAt time.Time `json:"generatedAt"`
	CLIVersion  string    `json:"cliVersion"`
	Template    string    `json:"template"`
	ADLFile     string    `json:"adlFile"`
}
