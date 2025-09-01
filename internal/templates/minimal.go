package templates

import (
	"fmt"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

// GetMinimalTemplate returns the minimal agent template files
func GetMinimalTemplate(adl *schema.ADL) map[string]string {
	// Detect language
	language := detectLanguageFromADL(adl)

	switch language {
	case "go":
		return getGoMinimalTemplate(adl)
	case "rust":
		return getRustMinimalTemplate(adl)
	default:
		// Default to Go for backward compatibility
		return getGoMinimalTemplate(adl)
	}
}

// detectLanguageFromADL detects the programming language from ADL
func detectLanguageFromADL(adl *schema.ADL) string {
	if adl.Spec.Language.Go != nil {
		return "go"
	}
	if adl.Spec.Language.Rust != nil {
		return "rust"
	}
	if adl.Spec.Language.TypeScript != nil {
		return "typescript"
	}
	return "go" // default
}

// getGoMinimalTemplate returns the minimal Go agent template files
func getGoMinimalTemplate(adl *schema.ADL) map[string]string {
	files := map[string]string{
		"main.go":                minimalMainGoTemplate,
		"go.mod":                 goModTemplate,
		".well-known/agent.json": cardJSONTemplate,
		"Taskfile.yml":           taskfileTemplate,
		"Dockerfile":             dockerfileTemplate,
		".gitignore":             gitignoreTemplate,
		".gitattributes":         gitattributesTemplate,
		".editorconfig":          editorconfigTemplate,
		"README.md":              minimalReadmeTemplate,
		"k8s/a2a-server.yaml":    minimalOperatorTemplate,
	}

	for _, tool := range adl.Spec.Tools {
		files["tools/"+tool.Name+".go"] = generateToolTemplate(tool)
	}

	return files
}

// getRustMinimalTemplate returns the minimal Rust agent template files
func getRustMinimalTemplate(adl *schema.ADL) map[string]string {
	files := map[string]string{
		"src/main.rs":            minimalMainRustTemplate,
		"Cargo.toml":             cargoTomlTemplate,
		".well-known/agent.json": cardJSONTemplate,
		"Taskfile.yml":           rustTaskfileTemplate,
		"Dockerfile":             rustDockerfileTemplate,
		".gitignore":             rustGitignoreTemplate,
		".gitattributes":         gitattributesTemplate,
		".editorconfig":          editorconfigTemplate,
		"README.md":              minimalRustReadmeTemplate,
		"k8s/a2a-server.yaml":    minimalOperatorTemplate,
	}

	for _, tool := range adl.Spec.Tools {
		files["src/tools/"+tool.Name+".rs"] = generateRustToolTemplate(tool)
	}

	// Generate tools module file
	if len(adl.Spec.Tools) > 0 {
		files["src/tools/mod.rs"] = generateRustToolsModTemplate(adl)
	}

	return files
}

const minimalMainGoTemplate = `package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/inference-gateway/adk/server"
	"github.com/inference-gateway/adk/server/config"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"

	"{{ .ADL.Spec.Language.Go.Module }}/tools"
)

// Config represents the application configuration
type Config struct {
	// Core application settings
	Environment string ` + "`" + `env:"ENVIRONMENT"` + "`" + `
	
	// A2A framework configuration (all A2A_ prefixed vars)
	A2A config.Config ` + "`" + `env:",prefix=A2A_"` + "`" + `
}

var (
	Version          = "{{ .ADL.Metadata.Version }}"
	AgentName        = "{{ .ADL.Metadata.Name }}"
	AgentDescription = "{{ .ADL.Metadata.Description }}"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment variables
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatal("failed to load config:", err)
	}

	// Initialize logger with simple configuration
	var logger *zap.Logger
	var err error
	if cfg.A2A.Debug || cfg.Environment == "dev" || cfg.Environment == "development" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatal("failed to initialize logger:", err)
	}
	defer logger.Sync()

	logger.Info("starting {{ .ADL.Metadata.Name }} agent", 
		zap.String("version", Version),
		zap.String("environment", cfg.Environment),
	)
	logger.Debug("loaded configuration")

	// Create toolbox and register tools
	toolBox := server.NewDefaultToolBox()

	{{- range .ADL.Spec.Tools }}
	// Register {{ .Name }} tool
	{{ .Name }}Tool := tools.New{{ .Name | title }}Tool()
	toolBox.AddTool({{ .Name }}Tool)
	logger.Info("registered tool", zap.String("tool", "{{ .Name }}"), zap.String("description", "{{ .Description }}"))
	{{- end }}

	// Create A2A agent with configuration
	agent, err := server.NewAgentBuilder(logger).
		WithConfig(&cfg.A2A.AgentConfig).
		WithToolBox(toolBox).
		WithSystemPrompt(` + "`" + `{{- if .ADL.Spec.Agent.SystemPrompt }}{{ .ADL.Spec.Agent.SystemPrompt }}{{- else }}You are a helpful AI assistant.{{- end }}` + "`" + `).
		Build()
	if err != nil {
		logger.Fatal("failed to create agent", zap.Error(err))
	}

	// Create A2A server with agent and configuration
	a2aServer, err := server.NewA2AServerBuilder(cfg.A2A, logger).
		WithAgent(agent).
		WithAgentCardFromFile("./.well-known/agent.json", map[string]interface{}{
			"name":        AgentName,
			"version":     Version,
			"description": AgentDescription,
			"url":         cfg.A2A.AgentURL,
		}).
		Build()
	if err != nil {
		logger.Fatal("failed to create A2A server", zap.Error(err))
	}

	// Start server in background
	go func() {
		logger.Info("starting A2A server", 
			zap.String("port", cfg.A2A.ServerConfig.Port),
			zap.String("host", cfg.A2A.ServerConfig.Host),
		)
		if err := a2aServer.Start(ctx); err != nil {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	logger.Info("{{ .ADL.Metadata.Name }} agent running successfully", 
		zap.String("port", cfg.A2A.ServerConfig.Port),
		zap.String("environment", cfg.Environment),
	)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received, gracefully stopping server...")
	a2aServer.Stop(ctx)
	logger.Info("{{ .ADL.Metadata.Name }} agent stopped")
}
`

const goModTemplate = `module {{ .ADL.Spec.Language.Go.Module }}

go {{ .ADL.Spec.Language.Go.Version }}

require (
	github.com/inference-gateway/adk v0.9.0
	github.com/sethvargo/go-envconfig v1.2.0
	go.uber.org/zap v1.27.0
)
`

const cardJSONTemplate = `{
	"schemaVersion": "0.1.0",
	"name": "{{ .ADL.Metadata.Name }}",
	"version": "{{ .ADL.Metadata.Version }}",
	"description": "{{ .ADL.Metadata.Description }}",
	"capabilities": {
		{{- if .ADL.Spec.Capabilities }}
		"streaming": {{ .ADL.Spec.Capabilities.Streaming }},
		"pushNotifications": {{ .ADL.Spec.Capabilities.PushNotifications }},
		"stateTransitionHistory": {{ .ADL.Spec.Capabilities.StateTransitionHistory }}
		{{- else }}
		"streaming": false,
		"pushNotifications": false,
		"stateTransitionHistory": false
		{{- end }}
	},
	"tools": [
		{{- range $index, $tool := .ADL.Spec.Tools }}
		{{- if $index }},{{ end }}
		{
			"name": "{{ $tool.Name }}",
			"description": "{{ $tool.Description }}",
			"schema": {{ $tool.Schema | toJson }}
		}
		{{- end }}
	]
}
`

const taskfileTemplate = `version: '3'

vars:
  APP_NAME: {{ .ADL.Metadata.Name }}
  VERSION: {{ .ADL.Metadata.Version }}

tasks:
  generate:
    desc: Generate code from ADL
    cmd: a2a generate --file agent.yaml --output .

  build:
    desc: Build the application
    cmd: go build -o bin/{{` + "`{{.APP_NAME}}`" + `}} .

  run:
    desc: Run the application in development mode
    cmd: go run .
    env:
      A2A_DEBUG: true
      A2A_SERVER_PORT: {{ .ADL.Spec.Server.Port | default 8080 }}

  test:
    desc: Run tests
    cmd: go test -v ./...

  test:cover:
    desc: Run tests with coverage
    cmd: go test -v -cover ./...

  fmt:
    desc: Format and vet code
    cmd: go fmt ./...

  vet:
    desc: Run go vet
    cmd: go vet ./...

  lint:
    desc: Run linter
    cmd: golangci-lint run

  clean:
    desc: Clean build artifacts
    cmd: rm -rf bin/

  docker:build:
    desc: Build Docker image
    cmd: docker build -t {{` + "`{{.APP_NAME}}`" + `}}:{{` + "`{{.VERSION}}`" + `}} .
`

const dockerfileTemplate = `FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy agent card
COPY --from=builder /app/.well-known ./.well-known

# Expose port
EXPOSE {{ .ADL.Spec.Server.Port | default 8080 }}

# Set environment variables
ENV A2A_SERVER_PORT={{ .ADL.Spec.Server.Port | default 8080 }}
ENV A2A_SERVER_HOST=0.0.0.0

# Run the application
CMD ["./main"]
`

const gitignoreTemplate = `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with ` + "`" + `go test -c` + "`" + `
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

# Build output
bin/
dist/

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Environment files
.env
.env.local
.env.*.local

# Log files
*.log

# Temporary files
tmp/
temp/

# Coverage reports
coverage.txt
coverage.html
`

const minimalReadmeTemplate = `<div align="center">

# {{ .ADL.Metadata.Name | title }}

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-{{ .ADL.Spec.Language.Go.Version }}+-00ADD8?style=flat&logo=go)](https://golang.org)
[![A2A Protocol](https://img.shields.io/badge/A2A-Protocol-blue?style=flat)](https://github.com/inference-gateway/a2a)

**{{ .ADL.Metadata.Description }}**

A production-ready [Agent-to-Agent (A2A)](https://github.com/inference-gateway/a2a) server that provides AI-powered capabilities through a standardized protocol. Built with Go for high performance and reliability.

</div>

## Quick Start

` + "```bash" + `
# Run the agent
go run .

# Or with Docker
docker build -t {{ .ADL.Metadata.Name }} .
docker run -p {{ .ADL.Spec.Server.Port | default 8080 }}:{{ .ADL.Spec.Server.Port | default 8080 }} {{ .ADL.Metadata.Name }}
` + "```" + `

## Features

- ✅ A2A protocol compliant
- ✅ AI-powered capabilities{{- if .ADL.Spec.Capabilities }}{{- if .ADL.Spec.Capabilities.Streaming }}
- ✅ Streaming support{{- end }}{{- if .ADL.Spec.Capabilities.PushNotifications }}
- ✅ Push notifications{{- end }}{{- if .ADL.Spec.Capabilities.StateTransitionHistory }}
- ✅ State transition history{{- end }}{{- end }}
- ✅ Production ready
- ✅ Minimal dependencies
- ✅ Built with Go for performance

## Endpoints

- ` + "`GET /.well-known/agent.json`" + ` - Agent metadata and capabilities
- ` + "`GET /health`" + ` - Health check endpoint
- ` + "`POST /a2a`" + ` - A2A protocol endpoint

## Available Tools

{{- range .ADL.Spec.Tools }}
- **{{ .Name }}** - {{ .Description }}
{{- end }}

## Configuration

Configure the agent via environment variables:

### Core Application Settings

- ` + "`ENVIRONMENT`" + ` - Deployment environment

### A2A Agent Configuration

#### Server Configuration

- ` + "`A2A_SERVER_PORT`" + ` - Server port (default: ` + "`{{ .ADL.Spec.Server.Port | default 8080 }}`" + `)
- ` + "`A2A_SERVER_READ_TIMEOUT`" + ` - Maximum duration for reading requests (default: ` + "`120s`" + `)
- ` + "`A2A_SERVER_WRITE_TIMEOUT`" + ` - Maximum duration for writing responses (default: ` + "`120s`" + `)
- ` + "`A2A_SERVER_IDLE_TIMEOUT`" + ` - Maximum time to wait for next request (default: ` + "`120s`" + `)
- ` + "`A2A_SERVER_DISABLE_HEALTHCHECK_LOG`" + ` - Disable logging for health check requests (default: ` + "`true`" + `)

#### LLM Client Configuration

- ` + "`A2A_AGENT_CLIENT_PROVIDER`" + ` - LLM provider: ` + "`openai`" + `, ` + "`anthropic`" + `, ` + "`groq`" + `, ` + "`ollama`" + `, ` + "`deepseek`" + `, ` + "`cohere`" + `, ` + "`cloudflare`" + `
- ` + "`A2A_AGENT_CLIENT_MODEL`" + ` - Model to use
- ` + "`A2A_AGENT_CLIENT_API_KEY`" + ` - API key for LLM provider
- ` + "`A2A_AGENT_CLIENT_BASE_URL`" + ` - Custom LLM API endpoint
- ` + "`A2A_AGENT_CLIENT_TIMEOUT`" + ` - Timeout for LLM requests (default: ` + "`30s`" + `)
- ` + "`A2A_AGENT_CLIENT_MAX_RETRIES`" + ` - Maximum retries for LLM requests (default: ` + "`3`" + `)
- ` + "`A2A_AGENT_CLIENT_MAX_CHAT_COMPLETION_ITERATIONS`" + ` - Maximum chat completion iterations (default: ` + "`10`" + `)
- ` + "`A2A_AGENT_CLIENT_MAX_TOKENS`" + ` - Maximum tokens for LLM responses (default: ` + "`4096`" + `)
- ` + "`A2A_AGENT_CLIENT_TEMPERATURE`" + ` - Controls randomness of LLM output (default: ` + "`0.7`" + `)
- ` + "`A2A_AGENT_CLIENT_TOP_P`" + ` - Top-p sampling parameter (default: ` + "`1.0`" + `)
- ` + "`A2A_AGENT_CLIENT_FREQUENCY_PENALTY`" + ` - Frequency penalty (default: ` + "`0.0`" + `)
- ` + "`A2A_AGENT_CLIENT_PRESENCE_PENALTY`" + ` - Presence penalty (default: ` + "`0.0`" + `)
- ` + "`A2A_AGENT_CLIENT_SYSTEM_PROMPT`" + ` - System prompt to guide the LLM{{- if .ADL.Spec.Agent.SystemPrompt }} (default: ` + "`{{ .ADL.Spec.Agent.SystemPrompt }}`" + `){{- else }} (default: ` + "`You are a helpful AI assistant.`" + `){{- end }}
- ` + "`A2A_AGENT_CLIENT_MAX_CONVERSATION_HISTORY`" + ` - Maximum conversation history per context (default: ` + "`20`" + `)
- ` + "`A2A_AGENT_CLIENT_USER_AGENT`" + ` - User agent string (default: ` + "`a2a-agent/1.0`" + `)

#### Capabilities Configuration

{{- if .ADL.Spec.Capabilities }}
- ` + "`A2A_CAPABILITIES_STREAMING`" + ` - Enable streaming support (default: ` + "`{{ .ADL.Spec.Capabilities.Streaming }}`" + `)
- ` + "`A2A_CAPABILITIES_PUSH_NOTIFICATIONS`" + ` - Enable push notifications (default: ` + "`{{ .ADL.Spec.Capabilities.PushNotifications }}`" + `)
- ` + "`A2A_CAPABILITIES_STATE_TRANSITION_HISTORY`" + ` - Enable state transition history (default: ` + "`{{ .ADL.Spec.Capabilities.StateTransitionHistory }}`" + `)
{{- else }}
- ` + "`A2A_CAPABILITIES_STREAMING`" + ` - Enable streaming support (default: ` + "`false`" + `)
- ` + "`A2A_CAPABILITIES_PUSH_NOTIFICATIONS`" + ` - Enable push notifications (default: ` + "`false`" + `)
- ` + "`A2A_CAPABILITIES_STATE_TRANSITION_HISTORY`" + ` - Enable state transition history (default: ` + "`false`" + `)
{{- end }}

#### Authentication Configuration

- ` + "`A2A_AUTH_ENABLE`" + ` - Enable OIDC authentication (default: ` + "`false`" + `)
- ` + "`A2A_AUTH_ISSUER_URL`" + ` - OIDC issuer URL (default: ` + "`http://keycloak:8080/realms/inference-gateway-realm`" + `)
- ` + "`A2A_AUTH_CLIENT_ID`" + ` - OIDC client ID (default: ` + "`inference-gateway-client`" + `)
- ` + "`A2A_AUTH_CLIENT_SECRET`" + ` - OIDC client secret

#### TLS Configuration

- ` + "`A2A_SERVER_TLS_ENABLE`" + ` - Enable TLS (default: ` + "`false`" + `)
- ` + "`A2A_SERVER_TLS_CERT_PATH`" + ` - Path to TLS certificate file
- ` + "`A2A_SERVER_TLS_KEY_PATH`" + ` - Path to TLS private key file

#### Queue Configuration

- ` + "`A2A_QUEUE_MAX_SIZE`" + ` - Queue maximum size (default: ` + "`100`" + `)
- ` + "`A2A_QUEUE_CLEANUP_INTERVAL`" + ` - Queue cleanup interval (default: ` + "`30s`" + `)

#### Telemetry Configuration

- ` + "`A2A_TELEMETRY_ENABLE`" + ` - Enable OpenTelemetry metrics collection (default: ` + "`false`" + `)
- ` + "`A2A_TELEMETRY_METRICS_PORT`" + ` - Metrics server port (default: ` + "`9090`" + `)
- ` + "`A2A_TELEMETRY_METRICS_HOST`" + ` - Metrics server host
- ` + "`A2A_TELEMETRY_METRICS_READ_TIMEOUT`" + ` - Metrics server read timeout (default: ` + "`30s`" + `)
- ` + "`A2A_TELEMETRY_METRICS_WRITE_TIMEOUT`" + ` - Metrics server write timeout (default: ` + "`30s`" + `)
- ` + "`A2A_TELEMETRY_METRICS_IDLE_TIMEOUT`" + ` - Metrics server idle timeout (default: ` + "`60s`" + `)

### Logging Configuration

- ` + "`A2A_DEBUG`" + ` - Enable debug mode (default: ` + "`false`" + `)
- ` + "`LOG_LEVEL`" + ` - Log level: ` + "`debug`" + `, ` + "`info`" + `, ` + "`warn`" + `, ` + "`error`" + ` (default: ` + "`info`" + `)
- ` + "`LOG_FORMAT`" + ` - Log format: ` + "`json`" + `, ` + "`console`" + ` (default: ` + "`json`" + `)

## Getting Started

### Prerequisites

- Go {{ .ADL.Spec.Language.Go.Version }}+
- [Task](https://taskfile.dev/) (optional, for using Taskfile commands)

### Installation & Running

#### Using Task (recommended)

` + "```bash" + `
# Install dependencies
go mod tidy

# Generate code from ADL (if needed)
task generate

# Run in development mode with debug logging
task run

# Build the binary
task build

# Run tests
task test

# Run linting
task lint
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
docker build -t {{ .ADL.Metadata.Name }}:{{ .ADL.Metadata.Version }} .

# Run container
docker run -p {{ .ADL.Spec.Server.Port | default 8080 }}:{{ .ADL.Spec.Server.Port | default 8080 }} {{ .ADL.Metadata.Name }}:{{ .ADL.Metadata.Version }}
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
kubectl port-forward svc/{{ .ADL.Metadata.Name }} {{ .ADL.Spec.Server.Port | default 8080 }}:80 -n {{ .ADL.Metadata.Name }}-ns
` + "```" + `

The operator automatically manages deployment, scaling, health checks, and configuration.

## Example Usage

` + "```bash" + `
# Test the agent capabilities
curl http://localhost:{{ .ADL.Spec.Server.Port | default 8080 }}/.well-known/agent.json

# Health check
curl http://localhost:{{ .ADL.Spec.Server.Port | default 8080 }}/health

# Send a message to the agent
curl -X POST http://localhost:{{ .ADL.Spec.Server.Port | default 8080 }}/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "message/send",
    "params": {
      "message": {
        "role": "user",
        "content": "Hello! What can you help me with?"
      }
    },
    "id": 1
  }'
` + "```" + `

## Development

### Project Structure

` + "```" + `
.
├── main.go                    # Main server setup and configuration
├── tools/                     # Tool implementations
{{- range .ADL.Spec.Tools }}
│   ├── {{ .Name }}.go         # {{ .Description }}
{{- end }}
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── .well-known/
│   └── agent.json             # Agent capabilities metadata
├── Taskfile.yml               # Task runner definitions
├── Dockerfile                 # Container configuration
├── .gitignore                 # Git ignore patterns
├── .gitattributes             # Git attributes
├── .editorconfig              # Editor configuration
├── k8s/
│   └── a2a-server.yaml        # Kubernetes A2A Custom Resource
└── README.md                  # This documentation
` + "```" + `

### Implementing Tools

Each tool is implemented in its own file under the ` + "`tools/`" + ` directory. Tools follow this pattern:

` + "```go" + `
package tools

import (
    "context"
    "github.com/inference-gateway/adk/server"
)

func NewToolNameTool() server.Tool {
    return server.NewBasicTool(
        "tool_name",
        "Tool description",
        schema,
        toolNameHandler,
    )
}

func toolNameHandler(ctx context.Context, args map[string]interface{}) (string, error) {
    // Your implementation here
    return result, nil
}
` + "```" + `

### Testing

Add comprehensive tests for your tool implementations:

` + "```bash" + `
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run tests with race detection
go test -v -race ./...

# Run specific tool tests
go test -v ./tools/
` + "```" + `

### Code Quality

` + "```bash" + `
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Or use Task for convenience
task lint
` + "```" + `

## Version Information

This project was generated with:
- **CLI Version:** {{ .Metadata.CLIVersion }}
- **Template:** {{ .Metadata.Template }}
- **Generated At:** {{ .Metadata.GeneratedAt.Format "2006-01-02 15:04:05" }}
- **A2A Version:** {{ .ADL.Metadata.Version }}

To regenerate with ADL changes:
` + "```bash" + `
a2a generate --file agent.yaml --output . --overwrite
` + "```" + `

## License

MIT

---

<div align="center">

🤖 **This server is powered by the [Inference Gateway A2A (Agent-to-Agent) framework](https://github.com/inference-gateway/a2a)**

[![A2A Documentation](https://img.shields.io/badge/📚-A2A_Docs-blue?style=for-the-badge)](https://github.com/inference-gateway/a2a)
[![Inference Gateway](https://img.shields.io/badge/🚀-Inference_Gateway-green?style=for-the-badge)](https://docs.inference-gateway.com)

</div>
`

const minimalOperatorTemplate = `---
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
  replicas: 1
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
  agent:
    enabled: true
    systemPrompt: |
{{- if .ADL.Spec.Agent.SystemPrompt }}{{ .ADL.Spec.Agent.SystemPrompt | nindent 6 }}{{- else }}      You are a helpful AI assistant.{{- end }}
  resources:
    limits:
      memory: "256Mi"
      cpu: "200m"
    requests:
      memory: "128Mi"
      cpu: "100m"
`

// generateToolTemplate creates a template for an individual tool
func generateToolTemplate(tool schema.Tool) string {
	return `package tools

import (
	"context"
	"fmt"

	"github.com/inference-gateway/adk/server"
)

// New` + titleCase(tool.Name) + `Tool creates a new ` + tool.Name + ` tool
func New` + titleCase(tool.Name) + `Tool() server.Tool {
	return server.NewBasicTool(
		"` + tool.Name + `",
		"` + tool.Description + `",
		` + generateSchemaString(tool.Schema) + `,
		` + tool.Name + `Handler,
	)
}

// ` + tool.Name + `Handler handles the ` + tool.Name + ` tool execution
func ` + tool.Name + `Handler(ctx context.Context, args map[string]interface{}) (string, error) {
	// TODO: Implement ` + tool.Name + ` logic
	// ` + tool.Description + `
	
	` + generateToolLogic(tool) + `
	
	return fmt.Sprintf(` + "`" + `{"result": "TODO: Implement ` + tool.Name + ` logic", "input": %+v}` + "`" + `, args), nil
}
`
}

// generateSchemaString converts a schema map to a Go string representation
func generateSchemaString(schema map[string]interface{}) string {
	if schema == nil {
		return `map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		}`
	}

	// For now, return a basic schema structure
	// TODO: Implement proper schema serialization
	return `map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			// TODO: Add proper schema properties
		},
	}`
}

// generateToolLogic generates basic parameter extraction logic
func generateToolLogic(tool schema.Tool) string {
	if tool.Schema == nil {
		return "// No parameters defined"
	}

	logic := "// Extract parameters from args\n"
	if properties, ok := tool.Schema["properties"].(map[string]interface{}); ok {
		for paramName, prop := range properties {
			if propMap, ok := prop.(map[string]interface{}); ok {
				if propType, ok := propMap["type"].(string); ok {
					switch propType {
					case "string":
						logic += fmt.Sprintf("\t// %s := args[\"%s\"].(string)\n", paramName, paramName)
					case "number":
						logic += fmt.Sprintf("\t// %s := args[\"%s\"].(float64)\n", paramName, paramName)
					case "boolean":
						logic += fmt.Sprintf("\t// %s := args[\"%s\"].(bool)\n", paramName, paramName)
					default:
						logic += fmt.Sprintf("\t// %s := args[\"%s\"]\n", paramName, paramName)
					}
				}
			}
		}
	}

	return logic
}

// titleCase capitalizes the first letter of a string
func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}

const editorconfigTemplate = `# EditorConfig is awesome: https://EditorConfig.org

# top-most EditorConfig file
root = true

# All files
[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

# Go files
[*.go]
indent_style = tab
indent_size = 4

# YAML files
[*.{yml,yaml}]
indent_style = space
indent_size = 2

# JSON files
[*.json]
indent_style = space
indent_size = 2

# Markdown files
[*.md]
indent_style = space
indent_size = 2

# Dockerfile
[Dockerfile]
indent_style = space
indent_size = 2

# Shell scripts
[*.sh]
indent_style = space
indent_size = 2

# Rust files
[*.rs]
indent_style = space
indent_size = 4
`

// Rust Templates

const minimalMainRustTemplate = `use std::sync::Arc;
use tokio::signal;
use tracing::{info, warn, error};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use rust_adk::{
    agent::{Agent, AgentBuilder},
    server::{Server, ServerConfig},
    tools::ToolBox,
};

mod tools;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize tracing
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "{{ .ADL.Metadata.Name | replace "-" "_" }}=info".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    info!("Starting {{ .ADL.Metadata.Name }} agent v{{ .ADL.Metadata.Version }}");

    // Create toolbox and register tools
    let mut toolbox = ToolBox::new();
    {{- range .ADL.Spec.Tools }}
    
    // Register {{ .Name }} tool
    let {{ .Name }}_tool = tools::{{ .Name }}::{{ .Name | title }}Tool::new();
    toolbox.add_tool("{{ .Name }}", Box::new({{ .Name }}_tool));
    info!("Registered tool: {{ .Name }} - {{ .Description }}");
    {{- end }}

    // Build agent
    let agent = AgentBuilder::new()
        .with_system_prompt("{{- if .ADL.Spec.Agent.SystemPrompt }}{{ .ADL.Spec.Agent.SystemPrompt }}{{- else }}You are a helpful AI assistant.{{- end }}")
        .with_toolbox(toolbox)
        .build()
        .await?;

    // Server configuration
    let server_config = ServerConfig {
        host: "0.0.0.0".to_string(),
        port: {{ .ADL.Spec.Server.Port | default 8080 }},
        debug: {{ .ADL.Spec.Server.Debug | default false }},
    };

    // Create and start server
    let server = Server::new(server_config, Arc::new(agent));
    
    // Start server in background
    let server_handle = tokio::spawn(async move {
        if let Err(e) = server.run().await {
            error!("Server error: {}", e);
        }
    });

    info!("{{ .ADL.Metadata.Name }} agent running on port {{ .ADL.Spec.Server.Port | default 8080 }}");

    // Wait for shutdown signal
    signal::ctrl_c().await?;
    warn!("Shutdown signal received, gracefully stopping...");

    // Cancel server
    server_handle.abort();
    
    info!("{{ .ADL.Metadata.Name }} agent stopped");
    Ok(())
}
`

const cargoTomlTemplate = `[package]
name = "{{ .ADL.Spec.Language.Rust.PackageName }}"
version = "{{ .ADL.Metadata.Version }}"
edition = "{{ .ADL.Spec.Language.Rust.Edition }}"
rust-version = "{{ .ADL.Spec.Language.Rust.Version }}"
description = "{{ .ADL.Metadata.Description }}"

[dependencies]
rust-adk = { git = "https://github.com/inference-gateway/rust-adk" }
tokio = { version = "1.0", features = ["full"] }
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
anyhow = "1.0"

[profile.release]
opt-level = 3
lto = true
codegen-units = 1
`

const rustTaskfileTemplate = `version: '3'

vars:
  APP_NAME: {{ .ADL.Metadata.Name }}
  VERSION: {{ .ADL.Metadata.Version }}

tasks:
  generate:
    desc: Generate code from ADL
    cmd: adl generate --file agent.yaml --output .

  build:
    desc: Build the application
    cmd: cargo build --release
    
  run:
    desc: Run the application in development mode
    cmd: cargo run
    env:
      RUST_LOG: debug

  test:
    desc: Run tests
    cmd: cargo test --verbose

  test:coverage:
    desc: Run tests with coverage
    cmd: cargo tarpaulin --verbose --all-features --workspace --timeout 120

  fmt:
    desc: Format code
    cmd: cargo fmt

  fmt:check:
    desc: Check code formatting
    cmd: cargo fmt --check

  lint:
    desc: Run clippy linter
    cmd: cargo clippy --all-targets --all-features -- -D warnings

  clean:
    desc: Clean build artifacts
    cmd: cargo clean

  docker:build:
    desc: Build Docker image
    cmd: docker build -t {{` + "`{{.APP_NAME}}`" + `}}:{{` + "`{{.VERSION}}`" + `}} .
`

const rustDockerfileTemplate = `FROM rust:{{ .ADL.Spec.Language.Rust.Version }}-slim as builder

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy manifest files
COPY Cargo.toml Cargo.lock ./

# Copy source code
COPY src ./src

# Build the application
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/target/release/{{ .ADL.Spec.Language.Rust.PackageName }} ./{{ .ADL.Metadata.Name }}

# Copy agent card
COPY --from=builder /app/.well-known ./.well-known

# Create non-root user
RUN groupadd -r appuser && useradd -r -g appuser appuser
RUN chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE {{ .ADL.Spec.Server.Port | default 8080 }}

# Set environment variables
ENV RUST_LOG=info

# Run the application
CMD ["./{{ .ADL.Metadata.Name }}"]
`

const rustGitignoreTemplate = `# Rust build artifacts
/target/
Cargo.lock

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Environment files
.env
.env.local
.env.*.local

# Log files
*.log

# Temporary files
tmp/
temp/

# Coverage reports
tarpaulin-report.html
lcov.info
`

const minimalRustReadmeTemplate = `<div align="center">

# {{ .ADL.Metadata.Name | title }}

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Rust Version](https://img.shields.io/badge/Rust-{{ .ADL.Spec.Language.Rust.Version }}+-orange?style=flat&logo=rust)](https://www.rust-lang.org)
[![A2A Protocol](https://img.shields.io/badge/A2A-Protocol-blue?style=flat)](https://github.com/inference-gateway/rust-adk)

**{{ .ADL.Metadata.Description }}**

A production-ready [Agent-to-Agent (A2A)](https://github.com/inference-gateway/rust-adk) server that provides AI-powered capabilities through a standardized protocol. Built with Rust for performance and safety.

</div>

## Quick Start

` + "```bash" + `
# Run the agent
cargo run

# Or with Docker
docker build -t {{ .ADL.Metadata.Name }} .
docker run -p {{ .ADL.Spec.Server.Port | default 8080 }}:{{ .ADL.Spec.Server.Port | default 8080 }} {{ .ADL.Metadata.Name }}
` + "```" + `

## Features

- ✅ A2A protocol compliant
- ✅ AI-powered capabilities{{- if .ADL.Spec.Capabilities }}{{- if .ADL.Spec.Capabilities.Streaming }}
- ✅ Streaming support{{- end }}{{- if .ADL.Spec.Capabilities.PushNotifications }}
- ✅ Push notifications{{- end }}{{- if .ADL.Spec.Capabilities.StateTransitionHistory }}
- ✅ State transition history{{- end }}{{- end }}
- ✅ Production ready
- ✅ Memory safe with Rust
- ✅ High performance async runtime

## Endpoints

- ` + "`GET /.well-known/agent.json`" + ` - Agent metadata and capabilities
- ` + "`GET /health`" + ` - Health check endpoint
- ` + "`POST /a2a`" + ` - A2A protocol endpoint

## Available Tools

{{- range .ADL.Spec.Tools }}
- **{{ .Name }}** - {{ .Description }}
{{- end }}

## Configuration

Configure the agent via environment variables:

### Logging Configuration

- ` + "`RUST_LOG`" + ` - Log level: ` + "`trace`" + `, ` + "`debug`" + `, ` + "`info`" + `, ` + "`warn`" + `, ` + "`error`" + ` (default: ` + "`info`" + `)

## Getting Started

### Prerequisites

- Rust {{ .ADL.Spec.Language.Rust.Version }}+
- [Task](https://taskfile.dev/) (optional, for using Taskfile commands)

### Installation & Running

#### Using Cargo (recommended)

` + "```bash" + `
# Run in development mode with debug logging
RUST_LOG=debug cargo run

# Build release binary
cargo build --release

# Run release binary
./target/release/{{ .ADL.Spec.Language.Rust.PackageName }}
` + "```" + `

#### Using Task

` + "```bash" + `
# Run in development mode with debug logging
task run

# Build the binary
task build

# Run tests
task test

# Run linting
task lint

# Format code
task fmt
` + "```" + `

#### Using Docker

` + "```bash" + `
# Build Docker image
docker build -t {{ .ADL.Metadata.Name }}:{{ .ADL.Metadata.Version }} .

# Run container
docker run -p {{ .ADL.Spec.Server.Port | default 8080 }}:{{ .ADL.Spec.Server.Port | default 8080 }} {{ .ADL.Metadata.Name }}:{{ .ADL.Metadata.Version }}
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
kubectl port-forward svc/{{ .ADL.Metadata.Name }} {{ .ADL.Spec.Server.Port | default 8080 }}:80 -n {{ .ADL.Metadata.Name }}-ns
` + "```" + `

The operator automatically manages deployment, scaling, health checks, and configuration.

## Example Usage

` + "```bash" + `
# Test the agent capabilities
curl http://localhost:{{ .ADL.Spec.Server.Port | default 8080 }}/.well-known/agent.json

# Health check
curl http://localhost:{{ .ADL.Spec.Server.Port | default 8080 }}/health

# Send a message to the agent
curl -X POST http://localhost:{{ .ADL.Spec.Server.Port | default 8080 }}/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "message/send",
    "params": {
      "message": {
        "role": "user",
        "content": "Hello! What can you help me with?"
      }
    },
    "id": 1
  }'
` + "```" + `

## Development

### Project Structure

` + "```" + `
.
├── src/
│   ├── main.rs                # Main server setup and configuration
│   └── tools/                 # Tool implementations
{{- range .ADL.Spec.Tools }}
│       ├── {{ .Name }}.rs     # {{ .Description }}
{{- end }}
│       └── mod.rs             # Tools module declaration
├── Cargo.toml                 # Rust package manifest
├── .well-known/
│   └── agent.json             # Agent capabilities metadata
├── Taskfile.yml               # Task runner definitions
├── Dockerfile                 # Container configuration
├── .gitignore                 # Git ignore patterns
├── .gitattributes             # Git attributes
├── .editorconfig              # Editor configuration
├── k8s/
│   └── a2a-server.yaml        # Kubernetes A2A Custom Resource
└── README.md                  # This documentation
` + "```" + `

### Implementing Tools

Each tool is implemented in its own file under the ` + "`src/tools/`" + ` directory. Tools follow this pattern:

` + "```rust" + `
use async_trait::async_trait;
use rust_adk::tools::Tool;
use serde_json::Value;

pub struct ToolNameTool;

impl ToolNameTool {
    pub fn new() -> Self {
        Self
    }
}

#[async_trait]
impl Tool for ToolNameTool {
    fn name(&self) -> &str {
        "tool_name"
    }

    fn description(&self) -> &str {
        "Tool description"
    }

    async fn execute(&self, args: Value) -> Result<Value, Box<dyn std::error::Error + Send + Sync>> {
        // Your implementation here
        Ok(serde_json::json!({"result": "success"}))
    }
}
` + "```" + `

### Testing

Add comprehensive tests for your tool implementations:

` + "```bash" + `
# Run all tests
cargo test --verbose

# Run tests with coverage
cargo tarpaulin --verbose --all-features --workspace --timeout 120

# Run specific tool tests
cargo test tools::
` + "```" + `

### Code Quality

` + "```bash" + `
# Format code
cargo fmt

# Check formatting
cargo fmt --check

# Run linter
cargo clippy --all-targets --all-features -- -D warnings

# Or use Task for convenience
task lint
` + "```" + `

## Version Information

This project was generated with:
- **CLI Version:** {{ .Metadata.CLIVersion }}
- **Template:** {{ .Metadata.Template }}
- **Generated At:** {{ .Metadata.GeneratedAt.Format "2006-01-02 15:04:05" }}
- **A2A Version:** {{ .ADL.Metadata.Version }}

To regenerate with ADL changes:
` + "```bash" + `
adl generate --file agent.yaml --output . --overwrite
` + "```" + `

## License

MIT

---

<div align="center">

🦀 **This server is powered by the [Rust ADK (Agent Development Kit)](https://github.com/inference-gateway/rust-adk)**

[![Rust ADK Documentation](https://img.shields.io/badge/📚-Rust_ADK_Docs-blue?style=for-the-badge)](https://github.com/inference-gateway/rust-adk)
[![Inference Gateway](https://img.shields.io/badge/🚀-Inference_Gateway-green?style=for-the-badge)](https://docs.inference-gateway.com)

</div>
`

// generateRustToolTemplate creates a Rust template for an individual tool
func generateRustToolTemplate(tool schema.Tool) string {
	return `use async_trait::async_trait;
use rust_adk::tools::Tool;
use serde_json::Value;
use tracing::{debug, error};

pub struct ` + titleCase(tool.Name) + `Tool;

impl ` + titleCase(tool.Name) + `Tool {
    pub fn new() -> Self {
        Self
    }
}

#[async_trait]
impl Tool for ` + titleCase(tool.Name) + `Tool {
    fn name(&self) -> &str {
        "` + tool.Name + `"
    }

    fn description(&self) -> &str {
        "` + tool.Description + `"
    }

    async fn execute(&self, args: Value) -> Result<Value, Box<dyn std::error::Error + Send + Sync>> {
        debug!("Executing ` + tool.Name + ` tool with args: {:?}", args);
        
        // TODO: Implement ` + tool.Name + ` logic
        // ` + tool.Description + `
        
        ` + generateRustToolLogic(tool) + `
        
        // Placeholder implementation
        let result = serde_json::json!({
            "result": "TODO: Implement ` + tool.Name + ` logic",
            "input": args
        });
        
        debug!("` + titleCase(tool.Name) + ` tool result: {:?}", result);
        Ok(result)
    }
}
`
}

// generateRustToolLogic generates basic parameter extraction logic for Rust tools
func generateRustToolLogic(tool schema.Tool) string {
	if tool.Schema == nil {
		return "// No parameters defined"
	}

	logic := "// Extract parameters from args\n"
	if properties, ok := tool.Schema["properties"].(map[string]interface{}); ok {
		for paramName, prop := range properties {
			if propMap, ok := prop.(map[string]interface{}); ok {
				if propType, ok := propMap["type"].(string); ok {
					switch propType {
					case "string":
						logic += fmt.Sprintf("        // let %s = args[\"%s\"].as_str().unwrap_or_default();\n", paramName, paramName)
					case "number":
						logic += fmt.Sprintf("        // let %s = args[\"%s\"].as_f64().unwrap_or(0.0);\n", paramName, paramName)
					case "boolean":
						logic += fmt.Sprintf("        // let %s = args[\"%s\"].as_bool().unwrap_or(false);\n", paramName, paramName)
					default:
						logic += fmt.Sprintf("        // let %s = &args[\"%s\"];\n", paramName, paramName)
					}
				}
			}
		}
	}

	return logic
}

// generateRustToolsModTemplate creates a module declaration file for Rust tools
func generateRustToolsModTemplate(adl *schema.ADL) string {
	content := "// Module declarations for tools\n\n"

	for _, tool := range adl.Spec.Tools {
		content += fmt.Sprintf("pub mod %s;\n", tool.Name)
	}

	return content
}
