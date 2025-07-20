package templates

// getAIPoweredTemplate returns the AI-powered agent template files
func getAIPoweredTemplate() map[string]string {
	return map[string]string{
		"main.go":             mainGoTemplate,
		"go.mod":              goModTemplate,
		"tools.go":            toolsGoTemplate,
		"config.go":           configGoTemplate,
		"Taskfile.yml":        taskfileTemplate,
		"Dockerfile":          dockerfileTemplate,
		".gitignore":          gitignoreTemplate,
		"README.md":           readmeTemplate,
		"k8s/a2a-server.yaml": aiPoweredOperatorTemplate,
	}
}

const mainGoTemplate = `package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/inference-gateway/a2a/adk"
	"github.com/inference-gateway/a2a/adk/server"
)

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context that cancels on interrupt
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Build A2A server
	serverBuilder := server.NewA2AServerBuilder()

	{{- if .ADL.Spec.Agent }}
	// Configure AI agent
	agentBuilder := adk.NewAgentBuilder()
	agentBuilder.SetProvider("{{ .ADL.Spec.Agent.Provider }}")
	agentBuilder.SetModel("{{ .ADL.Spec.Agent.Model }}")
	{{- if .ADL.Spec.Agent.SystemPrompt }}
	agentBuilder.SetSystemPrompt(` + "`" + `{{ .ADL.Spec.Agent.SystemPrompt }}` + "`" + `)
	{{- end }}
	{{- if .ADL.Spec.Agent.MaxTokens }}
	agentBuilder.SetMaxTokens({{ .ADL.Spec.Agent.MaxTokens }})
	{{- end }}
	{{- if .ADL.Spec.Agent.Temperature }}
	agentBuilder.SetTemperature({{ .ADL.Spec.Agent.Temperature }})
	{{- end }}

	{{- if .ADL.Spec.Tools }}
	// Add tools
	toolbox := adk.NewToolbox()
	{{- range .ADL.Spec.Tools }}
	toolbox.AddTool("{{ .Name }}", {{ .Name | title }}Tool)
	{{- end }}
	agentBuilder.SetToolbox(toolbox)
	{{- end }}

	agent, err := agentBuilder.Build()
	if err != nil {
		log.Fatalf("Failed to build agent: %v", err)
	}

	serverBuilder.SetAgent(agent)
	{{- end }}

	// Configure server
	serverBuilder.SetPort(config.Port)
	serverBuilder.SetDebug(config.Debug)

	{{- if .ADL.Spec.Capabilities }}
	// Configure capabilities
	{{- if .ADL.Spec.Capabilities.Streaming }}
	serverBuilder.EnableStreaming()
	{{- end }}
	{{- if .ADL.Spec.Capabilities.PushNotifications }}
	serverBuilder.EnablePushNotifications()
	{{- end }}
	{{- if .ADL.Spec.Capabilities.StateTransitionHistory }}
	serverBuilder.EnableStateTransitionHistory()
	{{- end }}
	{{- end }}

	// Build server
	srv, err := serverBuilder.Build()
	if err != nil {
		log.Fatalf("Failed to build server: %v", err)
	}

	// Start server
	fmt.Printf("🚀 Starting {{ .ADL.Metadata.Name }} agent on port %d\n", config.Port)
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	fmt.Println("👋 {{ .ADL.Metadata.Name }} agent stopped")
}
`

const goModTemplate = `module {{ .ADL.Spec.Language.Go.Module }}

go {{ .ADL.Spec.Language.Go.GoVersion }}

require (
	github.com/inference-gateway/a2a/adk v0.1.0
	github.com/sethvargo/go-envconfig v1.1.0
)
`

const toolsGoTemplate = `package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/inference-gateway/a2a/adk"
)

{{- range .ADL.Spec.Tools }}

// {{ .Name | title }}Tool implements the {{ .Name }} tool
// {{ .Description }}
func {{ .Name | title }}Tool(ctx context.Context, args map[string]interface{}) (string, error) {
	// TODO: Implement {{ .Name }} tool logic
	{{- if .Implementation }}
	{{ .Implementation }}
	{{- else }}
	// Parse arguments
	{{- range $key, $value := .Schema.properties }}
	{{- if $value.type }}
	// {{ $key }}: {{ $value.type }} - {{ $value.description }}
	{{- end }}
	{{- end }}

	// Example implementation - replace with your business logic
	result := map[string]interface{}{
		"tool": "{{ .Name }}",
		"args": args,
		"result": "TODO: Implement this tool",
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonResult), nil
	{{- end }}
}
{{- end }}
`

const configGoTemplate = `package main

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

// Config holds the application configuration
type Config struct {
	Port  int  ` + "`" + `env:"PORT,default={{ .ADL.Spec.Server.Port }}"` + "`" + `
	Debug bool ` + "`" + `env:"DEBUG,default={{ .ADL.Spec.Server.Debug }}"` + "`" + `

	{{- if .ADL.Spec.Agent }}
	// AI Provider Configuration
	{{- if eq .ADL.Spec.Agent.Provider "openai" }}
	OpenAIAPIKey string ` + "`" + `env:"OPENAI_API_KEY,required"` + "`" + `
	{{- else if eq .ADL.Spec.Agent.Provider "anthropic" }}
	AnthropicAPIKey string ` + "`" + `env:"ANTHROPIC_API_KEY,required"` + "`" + `
	{{- else if eq .ADL.Spec.Agent.Provider "azure" }}
	AzureOpenAIEndpoint string ` + "`" + `env:"AZURE_OPENAI_ENDPOINT,required"` + "`" + `
	AzureOpenAIAPIKey   string ` + "`" + `env:"AZURE_OPENAI_API_KEY,required"` + "`" + `
	{{- else if eq .ADL.Spec.Agent.Provider "ollama" }}
	OllamaBaseURL string ` + "`" + `env:"OLLAMA_BASE_URL,default=http://localhost:11434"` + "`" + `
	{{- end }}
	{{- end }}
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	ctx := context.Background()
	var config Config
	if err := envconfig.Process(ctx, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
`

const taskfileTemplate = `version: '3'

vars:
  APP_NAME: {{ .ADL.Metadata.Name }}
  VERSION: {{ .ADL.Metadata.Version | default "1.0.0" }}

tasks:
  build:
    desc: Build the {{ .ADL.Metadata.Name }} agent
    cmds:
      - go build -ldflags "-X main.Version=${VERSION}" -o bin/${APP_NAME} .

  run:
    desc: Run the {{ .ADL.Metadata.Name }} agent
    cmds:
      - go run .
    env:
      DEBUG: true

  test:
    desc: Run tests
    cmds:
      - go test -v ./...

  lint:
    desc: Run linter
    cmds:
      - golangci-lint run

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -rf bin/

  docker-build:
    desc: Build Docker image
    cmds:
      - docker build -t ${APP_NAME}:${VERSION} .

  docker-run:
    desc: Run Docker container
    cmds:
      - docker run --rm -p {{ .ADL.Spec.Server.Port }}:{{ .ADL.Spec.Server.Port }} ${APP_NAME}:${VERSION}

  dev:
    desc: Run in development mode with auto-reload
    deps: [build]
    cmds:
      - ./bin/${APP_NAME}
    watch: true
    sources:
      - "**/*.go"
      - "go.mod"
      - "go.sum"
`

const dockerfileTemplate = `# Build stage
FROM golang:{{ .ADL.Spec.Language.Go.GoVersion }}-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy the .well-known directory
COPY --from=builder /app/.well-known ./.well-known

# Expose port
EXPOSE {{ .ADL.Spec.Server.Port }}

# Run the binary
CMD ["./main"]
`

const gitignoreTemplate = `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/

# Test binary, built with ` + "`go test -c`" + `
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
vendor/

# Go workspace file
go.work
go.work.sum

# Environment variables
.env
.env.local

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Logs
*.log
`

const readmeTemplate = `# {{ .ADL.Metadata.Name | title }}

{{ .ADL.Metadata.Description }}

**Version:** {{ .ADL.Metadata.Version }}

## Overview

This A2A agent was generated using the A2A CLI from an Agent Definition Language (ADL) file.

### Capabilities

{{- if .ADL.Spec.Capabilities }}
- **Streaming:** {{ .ADL.Spec.Capabilities.Streaming }}
- **Push Notifications:** {{ .ADL.Spec.Capabilities.PushNotifications }}
- **State Transition History:** {{ .ADL.Spec.Capabilities.StateTransitionHistory }}
{{- end }}

{{- if .ADL.Spec.Agent }}

### AI Configuration

- **Provider:** {{ .ADL.Spec.Agent.Provider }}
- **Model:** {{ .ADL.Spec.Agent.Model }}
{{- if .ADL.Spec.Agent.MaxTokens }}
- **Max Tokens:** {{ .ADL.Spec.Agent.MaxTokens }}
{{- end }}
{{- if .ADL.Spec.Agent.Temperature }}
- **Temperature:** {{ .ADL.Spec.Agent.Temperature }}
{{- end }}
{{- end }}

{{- if .ADL.Spec.Tools }}

### Available Tools

{{- range .ADL.Spec.Tools }}
#### {{ .Name | title }}
{{ .Description }}

**Implementation Status:** ⚠️ TODO - Implement in ` + "`tools.go`" + `
{{- end }}
{{- end }}

## Getting Started

### Prerequisites

- Go {{ .ADL.Spec.Language.Go.GoVersion }}+
- [Task](https://taskfile.dev/) (optional, for using Taskfile commands)

### Environment Variables

{{- if .ADL.Spec.Agent }}
{{- if eq .ADL.Spec.Agent.Provider "openai" }}
- ` + "`OPENAI_API_KEY`" + `: Your OpenAI API key
{{- else if eq .ADL.Spec.Agent.Provider "anthropic" }}
- ` + "`ANTHROPIC_API_KEY`" + `: Your Anthropic API key
{{- else if eq .ADL.Spec.Agent.Provider "azure" }}
- ` + "`AZURE_OPENAI_ENDPOINT`" + `: Your Azure OpenAI endpoint
- ` + "`AZURE_OPENAI_API_KEY`" + `: Your Azure OpenAI API key
{{- else if eq .ADL.Spec.Agent.Provider "ollama" }}
- ` + "`OLLAMA_BASE_URL`" + `: Ollama server URL (default: http://localhost:11434)
{{- end }}
{{- end }}
- ` + "`PORT`" + `: Server port (default: {{ .ADL.Spec.Server.Port }})
- ` + "`DEBUG`" + `: Enable debug mode (default: {{ .ADL.Spec.Server.Debug }})

### Running the Agent

#### Using Task (recommended)

` + "```bash" + `
# Install dependencies
go mod tidy

# Run in development mode
task run

# Build the binary
task build

# Run tests
task test
` + "```" + `

#### Using Go directly

` + "```bash" + `
# Run directly
go run .

# Build and run
go build -o bin/{{ .ADL.Metadata.Name }} .
./bin/{{ .ADL.Metadata.Name }}
` + "```" + `

#### Using Docker

` + "```bash" + `
# Build Docker image
task docker-build

# Run container
task docker-run
` + "```" + `

#### Kubernetes Deployment

For production deployment using the [Inference Gateway Operator](https://github.com/inference-gateway/operator):

` + "```bash" + `
# Install the Inference Gateway Operator (if not already installed)
kubectl apply -f https://github.com/inference-gateway/operator/releases/latest/download/install.yaml

# Apply A2A Custom Resource
kubectl apply -f k8s/a2a-server.yaml

# Check A2A status
kubectl get a2a {{ .ADL.Metadata.Name }} -n {{ .ADL.Metadata.Name }}-ns

# View operator-managed deployment
kubectl get pods -n {{ .ADL.Metadata.Name }}-ns

# Port forward for testing
kubectl port-forward svc/{{ .ADL.Metadata.Name }} {{ .ADL.Spec.Server.Port }}:80 -n {{ .ADL.Metadata.Name }}-ns
` + "```" + `

The operator automatically manages deployment, scaling, health checks, and configuration.

## Development

### TODO: Implement Tools

{{- if .ADL.Spec.Tools }}
The following tools need implementation in ` + "`tools.go`" + `:

{{- range .ADL.Spec.Tools }}
- **{{ .Name | title }}**: {{ .Description }}
{{- end }}

Each tool function receives a ` + "`context.Context`" + ` and ` + "`map[string]interface{}`" + ` with the tool arguments, and should return a JSON string result.
{{- else }}
No tools defined in the ADL file.
{{- end }}

### Project Structure

` + "```" + `
.
├── main.go              # Main server setup
├── tools.go             # Tool implementations (⚠️ TODO)
├── config.go            # Configuration management
├── go.mod               # Go module definition
├── Taskfile.yml         # Task definitions
├── Dockerfile           # Container configuration
├── .well-known/
│   └── agent.json       # Agent capabilities (auto-generated)
└── README.md            # This file
` + "```" + `

### Testing

Add tests for your tool implementations:

` + "```bash" + `
# Run tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...
` + "```" + `

## API Endpoints

Once running, the agent exposes:

- ` + "`GET /.well-known/agent.json`" + ` - Agent capabilities and metadata
- ` + "`POST /chat`" + ` - Chat with the agent
{{- if .ADL.Spec.Capabilities.Streaming }}
- ` + "`POST /stream`" + ` - Streaming chat endpoint
{{- end }}

## Generated Files

This project was generated with:
- **CLI Version:** {{ .Metadata.CLIVersion }}
- **Template:** {{ .Metadata.Template }}
- **Generated At:** {{ .Metadata.GeneratedAt.Format "2006-01-02 15:04:05" }}

To regenerate or sync with ADL changes:
` + "```bash" + `
a2a sync --file agent.yaml
` + "```" + `

---

> 🤖 This agent is powered by the [A2A (Agent-to-Agent) framework](https://github.com/inference-gateway/a2a)
`

const aiPoweredOperatorTemplate = `---
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .ADL.Metadata.Name }}-ns
  labels:
    inference-gateway.com/managed: "true"
---
apiVersion: core.inference-gateway.com/v1alpha1
kind: A2A
metadata:
  name: {{ .ADL.Metadata.Name }}
  namespace: {{ .ADL.Metadata.Name }}-ns
spec:
  image: "{{ .ADL.Metadata.Name }}:{{ .ADL.Metadata.Version | default "latest" }}"
  timezone: "UTC"
  port: {{ .ADL.Spec.Server.Port | default 8080 }}
  host: "0.0.0.0"
  readTimeout: "30s"
  writeTimeout: "30s"
  idleTimeout: "60s"
  logging:
    level: "info"
    format: "json"
  telemetry:
    enabled: true
    metrics:
      enabled: true
      port: 9090
  queue:
    enabled: true
    maxSize: 1000
    cleanupInterval: "5m"
  tls:
    enabled: false
    secretRef: ""
  {{- if .ADL.Spec.Agent }}
  agent:
    enabled: true
    tls:
      enabled: false
      secretRef: ""
    maxConversationHistory: 10
    maxChatCompletionIterations: 5
    maxRetries: 3
    apiKey:
      secretRef: "{{ .ADL.Metadata.Name }}-api-key"
    llm:
      model: "{{ .ADL.Spec.Agent.Provider }}/{{ .ADL.Spec.Agent.Model }}"
      {{- if .ADL.Spec.Agent.MaxTokens }}
      maxTokens: {{ .ADL.Spec.Agent.MaxTokens }}
      {{- end }}
      {{- if .ADL.Spec.Agent.Temperature }}
      temperature: "{{ .ADL.Spec.Agent.Temperature }}"
      {{- end }}
      systemPrompt: {{ .ADL.Spec.Agent.SystemPrompt | quote }}
  {{- else }}
  agent:
    enabled: false
  {{- end }}
{{- if .ADL.Spec.Agent }}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ .ADL.Metadata.Name }}-api-key
  namespace: {{ .ADL.Metadata.Name }}-ns
type: Opaque
data:
  # TODO: Replace with base64 encoded API key
  api-key: "YOUR_BASE64_ENCODED_API_KEY_HERE"
{{- end }}
`
