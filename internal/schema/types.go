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
	Capabilities  *Capabilities             `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Card          *Card                     `yaml:"card,omitempty" json:"card,omitempty"`
	Agent         *Agent                    `yaml:"agent,omitempty" json:"agent,omitempty"`
	Config        map[string]map[string]any `yaml:"config,omitempty" json:"config,omitempty"`
	Dependencies  map[string]Dependency     `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Skills        []Skill                   `yaml:"skills,omitempty" json:"skills,omitempty"`
	Server        Server                    `yaml:"server" json:"server"`
	Language      *Language                 `yaml:"language,omitempty" json:"language,omitempty"`
	Acronyms      []string                  `yaml:"acronyms,omitempty" json:"acronyms,omitempty"`
	SCM           *SCM                      `yaml:"scm,omitempty" json:"scm,omitempty"`
	Sandbox       *SandboxConfig            `yaml:"sandbox,omitempty" json:"sandbox,omitempty"`
	Deployment    *DeploymentConfig         `yaml:"deployment,omitempty" json:"deployment,omitempty"`
	Hooks         *Hooks                    `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

// Card represents the agent card configuration
type Card struct {
	ProtocolVersion    string   `yaml:"protocolVersion,omitempty" json:"protocolVersion,omitempty"`
	URL                string   `yaml:"url,omitempty" json:"url,omitempty"`
	PreferredTransport string   `yaml:"preferredTransport,omitempty" json:"preferredTransport,omitempty"`
	DefaultInputModes  []string `yaml:"defaultInputModes,omitempty" json:"defaultInputModes,omitempty"`
	DefaultOutputModes []string `yaml:"defaultOutputModes,omitempty" json:"defaultOutputModes,omitempty"`
	DocumentationURL   string   `yaml:"documentationUrl,omitempty" json:"documentationUrl,omitempty"`
	IconURL            string   `yaml:"iconUrl,omitempty" json:"iconUrl,omitempty"`
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

// Dependency represents a service dependency
type Dependency struct {
	Type        string `yaml:"type" json:"type"`
	Interface   string `yaml:"interface" json:"interface"`
	Factory     string `yaml:"factory" json:"factory"`
	Description string `yaml:"description" json:"description"`
}


// Skill represents a distinct capability or function that an agent can perform
type Skill struct {
	ID             string         `yaml:"id" json:"id"`
	Name           string         `yaml:"name" json:"name"`
	Description    string         `yaml:"description" json:"description"`
	Tags           []string       `yaml:"tags" json:"tags"`
	Examples       []string       `yaml:"examples,omitempty" json:"examples,omitempty"`
	InputModes     []string       `yaml:"inputModes,omitempty" json:"inputModes,omitempty"`
	OutputModes    []string       `yaml:"outputModes,omitempty" json:"outputModes,omitempty"`
	Schema         map[string]any `yaml:"schema" json:"schema"`
	Implementation string         `yaml:"implementation,omitempty" json:"implementation,omitempty"`
	Inject         []string       `yaml:"inject,omitempty" json:"inject,omitempty"`
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

// RustConfig for Rust-specific settings
type RustConfig struct {
	PackageName string `yaml:"packageName" json:"packageName"`
	Version     string `yaml:"version" json:"version"`
	Edition     string `yaml:"edition" json:"edition"`
}

// Language configuration for different programming languages
type Language struct {
	Go         *GoConfig         `yaml:"go,omitempty" json:"go,omitempty"`
	TypeScript *TypeScriptConfig `yaml:"typescript,omitempty" json:"typescript,omitempty"`
	Rust       *RustConfig       `yaml:"rust,omitempty" json:"rust,omitempty"`
}

// SCM contains source control management configuration
type SCM struct {
	Provider       string `yaml:"provider" json:"provider"`
	URL            string `yaml:"url,omitempty" json:"url,omitempty"`
	GithubApp      bool   `yaml:"github_app,omitempty" json:"github_app,omitempty"`
	IssueTemplates bool   `yaml:"issue_templates,omitempty" json:"issue_templates,omitempty"`
}

// SandboxConfig for sandbox environment settings
type SandboxConfig struct {
	Flox         *FloxConfig         `yaml:"flox,omitempty" json:"flox,omitempty"`
	DevContainer *DevContainerConfig `yaml:"devcontainer,omitempty" json:"devcontainer,omitempty"`
}

// FloxConfig for Flox environment settings
type FloxConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// DevContainerConfig for Dev Container environment settings
type DevContainerConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// DeploymentConfig for deployment platform settings
type DeploymentConfig struct {
	Type     string            `yaml:"type,omitempty" json:"type,omitempty"`
	CloudRun *CloudRunConfig  `yaml:"cloudrun,omitempty" json:"cloudrun,omitempty"`
}

// CloudRunConfig for Google Cloud Run specific deployment settings
type CloudRunConfig struct {
	Image        *ImageConfig        `yaml:"image,omitempty" json:"image,omitempty"`
	Resources    *ResourcesConfig    `yaml:"resources,omitempty" json:"resources,omitempty"`
	Scaling      *ScalingConfig      `yaml:"scaling,omitempty" json:"scaling,omitempty"`
	Service      *ServiceConfig      `yaml:"service,omitempty" json:"service,omitempty"`
	Environment  map[string]string   `yaml:"environment,omitempty" json:"environment,omitempty"`
}

// ImageConfig for container image settings
type ImageConfig struct {
	Registry    string `yaml:"registry,omitempty" json:"registry,omitempty"`
	Repository  string `yaml:"repository,omitempty" json:"repository,omitempty"`
	Tag         string `yaml:"tag,omitempty" json:"tag,omitempty"`
	UseCloudBuild bool `yaml:"useCloudBuild,omitempty" json:"useCloudBuild,omitempty"`
}

// ResourcesConfig for CloudRun resource allocation
type ResourcesConfig struct {
	CPU    string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`
}

// ScalingConfig for CloudRun scaling settings
type ScalingConfig struct {
	MinInstances int `yaml:"minInstances,omitempty" json:"minInstances,omitempty"`
	MaxInstances int `yaml:"maxInstances,omitempty" json:"maxInstances,omitempty"`
	Concurrency  int `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
}

// ServiceConfig for CloudRun service settings
type ServiceConfig struct {
	Timeout                int    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	AllowUnauthenticated  bool   `yaml:"allowUnauthenticated,omitempty" json:"allowUnauthenticated,omitempty"`
	ServiceAccount        string `yaml:"serviceAccount,omitempty" json:"serviceAccount,omitempty"`
	ExecutionEnvironment  string `yaml:"executionEnvironment,omitempty" json:"executionEnvironment,omitempty"`
}

// Hooks contains lifecycle hooks for the generation process
type Hooks struct {
	Post []string `yaml:"post,omitempty" json:"post,omitempty"`
}

// GeneratedMetadata contains information about the generation
type GeneratedMetadata struct {
	GeneratedAt time.Time `json:"generatedAt"`
	CLIVersion  string    `json:"cliVersion"`
	Template    string    `json:"template"`
	ADLFile     string    `json:"adlFile"`
}
