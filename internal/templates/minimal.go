package templates

import (
	"fmt"

	"github.com/inference-gateway/a2a-cli/internal/schema"
)

// GetMinimalTemplate returns the minimal agent template files
func GetMinimalTemplate(adl *schema.ADL) map[string]string {
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

const minimalMainGoTemplate = `package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/inference-gateway/a2a/adk/server"
	"github.com/inference-gateway/a2a/adk/server/config"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"

	"{{ .ADL.Spec.Language.Go.Module }}/tools"
)

type Config struct {
	A2A config.Config ` + "`" + `env:",prefix=A2A_"` + "`" + `
}

var (
	Version          = "{{ .ADL.Metadata.Version }}"
	AgentName        = "{{ .ADL.Metadata.Name }}"
	AgentDescription = "{{ .ADL.Metadata.Description }}"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment first
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatal("failed to load config:", err)
	}

	// Initialize logger based on DEBUG environment variable
	var logger *zap.Logger
	var err error
	if cfg.A2A.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatal("failed to initialize logger:", err)
	}
	defer logger.Sync()

	logger.Debug("loaded configuration", zap.Any("config", cfg))

	// Create toolbox
	toolBox := server.NewDefaultToolBox()

	{{- range .ADL.Spec.Tools }}
	// Add {{ .Name }} tool
	{{ .Name }}Tool := tools.New{{ .Name | title }}Tool()
	toolBox.AddTool({{ .Name }}Tool)
	{{- end }}

	// Create A2A server with agent
	agent, err := server.NewAgentBuilder(logger).
		WithConfig(&cfg.A2A.AgentConfig).
		WithToolBox(toolBox).
		WithSystemPrompt(` + "`" + `{{- if .ADL.Spec.Agent.SystemPrompt }}{{ .ADL.Spec.Agent.SystemPrompt }}{{- else }}You are a helpful AI assistant.{{- end }}` + "`" + `).
		Build()
	if err != nil {
		log.Fatal("failed to create agent:", err)
	}

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
		log.Fatal("failed to create A2A server:", err)
	}

	// Start server
	go func() {
		if err := a2aServer.Start(ctx); err != nil {
			log.Fatal("server failed to start:", err)
		}
	}()

	logger.Info("{{ .ADL.Metadata.Name }} agent running", zap.String("port", cfg.A2A.ServerConfig.Port))

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	a2aServer.Stop(ctx)
}
`

const goModTemplate = `module {{ .ADL.Spec.Language.Go.Module }}

go {{ .ADL.Spec.Language.Go.Version }}

require (
	github.com/inference-gateway/a2a v0.7.3
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

  test-cover:
    desc: Run tests with coverage
    cmd: go test -v -cover ./...

  lint:
    desc: Run linter
    cmd: |
      go fmt ./...
      go vet ./...

  clean:
    desc: Clean build artifacts
    cmd: rm -rf bin/

  docker-build:
    desc: Build Docker image
    cmd: docker build -t {{` + "`{{.APP_NAME}}`" + `}}:{{` + "`{{.VERSION}}`" + `}} .

  docker-run:
    desc: Run Docker container
    cmd: |
      docker run -d \
        --name {{` + "`{{.APP_NAME}}`" + `}} \
        -p {{ .ADL.Spec.Server.Port | default 8080 }}:{{ .ADL.Spec.Server.Port | default 8080 }} \
        {{` + "`{{.APP_NAME}}`" + `}}:{{` + "`{{.VERSION}}`" + `}}

  k8s-deploy:
    desc: Deploy to Kubernetes using operator
    cmd: kubectl apply -f k8s/a2a-server.yaml

  k8s-delete:
    desc: Delete from Kubernetes
    cmd: kubectl delete -f k8s/a2a-server.yaml

  dev:
    desc: Development mode with auto-reload
    deps: [build]
    cmd: |
      while true; do
        ./bin/{{` + "`{{.APP_NAME}}`" + `}} &
        PID=$!
        inotifywait -e modify -r . --exclude='bin/.*'
        kill $PID 2>/dev/null || true
        sleep 1
      done
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

const minimalReadmeTemplate = `# {{ .ADL.Metadata.Name | title }}

{{ .ADL.Metadata.Description }}

**Version:** {{ .ADL.Metadata.Version }}

## Overview

This A2A (Agent-to-Agent) server was generated using the A2A CLI from an Agent Definition Language (ADL) file. This agent uses the inference gateway A2A framework with AI-powered capabilities.

### Capabilities

{{- if .ADL.Spec.Capabilities }}
- **Streaming:** {{ .ADL.Spec.Capabilities.Streaming }}
- **Push Notifications:** {{ .ADL.Spec.Capabilities.PushNotifications }}
- **State Transition History:** {{ .ADL.Spec.Capabilities.StateTransitionHistory }}
{{- end }}

### Available Tools

{{- range .ADL.Spec.Tools }}
- **{{ .Name }}** - {{ .Description }}
{{- end }}

## Getting Started

### Prerequisites

- Go {{ .ADL.Spec.Language.Go.Version }}+
- [Task](https://taskfile.dev/) (optional, for using Taskfile commands)

### Environment Variables

Common A2A configuration:
- ` + "`A2A_SERVER_PORT`" + `: Server port (default: {{ .ADL.Spec.Server.Port | default 8080 }})
- ` + "`A2A_DEBUG`" + `: Enable debug mode (default: false)
- ` + "`A2A_AGENT_SYSTEM_PROMPT`" + `: Override the system prompt
- ` + "`A2A_AGENT_MODEL`" + `: AI model to use

### Running the Server

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
kubectl port-forward svc/{{ .ADL.Metadata.Name }} {{ .ADL.Spec.Server.Port | default 8080 }}:80 -n {{ .ADL.Metadata.Name }}-ns
` + "```" + `

The operator automatically manages deployment, scaling, health checks, and configuration.

## Development

### Project Structure

` + "```" + `
.
├── main.go              # Main server setup
├── tools/               # Tool implementations
{{- range .ADL.Spec.Tools }}
│   ├── {{ .Name }}.go   # {{ .Description }}
{{- end }}
├── go.mod               # Go module definition
├── .well-known/
│   └── agent.json       # Agent capabilities
├── Taskfile.yml         # Task definitions
├── Dockerfile           # Container configuration
├── k8s/
│   └── a2a-server.yaml  # Kubernetes deployment
└── README.md            # This file
` + "```" + `

### Implementing Tools

Each tool is implemented in its own file under the ` + "`tools/`" + ` directory. Tools follow this pattern:

` + "```go" + `
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

Add tests for your tool implementations:

` + "```bash" + `
# Run tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...
` + "```" + `

### Example Usage

` + "```bash" + `
# Test the agent capabilities
curl http://localhost:{{ .ADL.Spec.Server.Port | default 8080 }}/.well-known/agent.json
` + "```" + `

## API Endpoints

Once running, the server exposes:

- ` + "`GET /.well-known/agent.json`" + ` - Agent capabilities and metadata

## Generated Files

This project was generated with:
- **CLI Version:** {{ .Metadata.CLIVersion }}
- **Template:** {{ .Metadata.Template }}
- **Generated At:** {{ .Metadata.GeneratedAt.Format "2006-01-02 15:04:05" }}

To regenerate with ADL changes:
` + "```bash" + `
a2a generate --file agent.yaml --output . --overwrite
` + "```" + `

---

> 🤖 This server is powered by the [A2A (Agent-to-Agent) framework](https://github.com/inference-gateway/a2a)
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

	"github.com/inference-gateway/a2a/adk/server"
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
`
