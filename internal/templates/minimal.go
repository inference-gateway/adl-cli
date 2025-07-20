package templates

// getMinimalTemplate returns the minimal agent template files (non-AI)
func getMinimalTemplate() map[string]string {
	return map[string]string{
		"main.go":             minimalMainGoTemplate,
		"go.mod":              goModTemplate, // Reuse from ai-powered
		"handlers.go":         handlersGoTemplate,
		"config.go":           minimalConfigGoTemplate,
		"Taskfile.yml":        taskfileTemplate,   // Reuse from ai-powered
		"Dockerfile":          dockerfileTemplate, // Reuse from ai-powered
		".gitignore":          gitignoreTemplate,  // Reuse from ai-powered
		"README.md":           minimalReadmeTemplate,
		"k8s/a2a-server.yaml": minimalOperatorTemplate,
	}
}

const minimalMainGoTemplate = `package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// Build A2A server (minimal mode - no AI agent)
	serverBuilder := server.NewA2AServerBuilder()

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

	// Add custom handlers
	serverBuilder.SetCustomHandler("/chat", NewChatHandler())
	{{- range .ADL.Spec.Tools }}
	serverBuilder.SetCustomHandler("/{{ .Name }}", New{{ .Name | title }}Handler())
	{{- end }}

	// Build server
	srv, err := serverBuilder.Build()
	if err != nil {
		log.Fatalf("Failed to build server: %v", err)
	}

	// Start server
	fmt.Printf("🚀 Starting {{ .ADL.Metadata.Name }} server on port %d\n", config.Port)
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	fmt.Println("👋 {{ .ADL.Metadata.Name }} server stopped")
}
`

const handlersGoTemplate = `package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ChatHandler handles chat requests
type ChatHandler struct{}

// NewChatHandler creates a new chat handler
func NewChatHandler() *ChatHandler {
	return &ChatHandler{}
}

// ServeHTTP handles HTTP requests
func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement chat logic
	// Parse request body
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Process the chat request
	// This is where you implement your business logic

	// Example response
	response := map[string]interface{}{
		"message": "Hello! This is a minimal A2A agent. Implement your logic in handlers.go",
		"agent":   "{{ .ADL.Metadata.Name }}",
		"version": "{{ .ADL.Metadata.Version }}",
		"request": request,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

{{- range .ADL.Spec.Tools }}

// {{ .Name | title }}Handler handles {{ .Name }} requests
type {{ .Name | title }}Handler struct{}

// New{{ .Name | title }}Handler creates a new {{ .Name }} handler
func New{{ .Name | title }}Handler() *{{ .Name | title }}Handler {
	return &{{ .Name | title }}Handler{}
}

// ServeHTTP handles HTTP requests for {{ .Name }}
func (h *{{ .Name | title }}Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement {{ .Name }} logic
	// {{ .Description }}

	// Parse request body
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	{{- if .Implementation }}
	{{ .Implementation }}
	{{- else }}
	// TODO: Implement {{ .Name }} functionality
	// Expected parameters:
	{{- range $key, $value := .Schema.properties }}
	{{- if $value.type }}
	// - {{ $key }}: {{ $value.type }} - {{ $value.description }}
	{{- end }}
	{{- end }}

	// Example response
	response := map[string]interface{}{
		"tool":   "{{ .Name }}",
		"result": "TODO: Implement {{ .Name }} logic",
		"input":  request,
	}
	{{- end }}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
{{- end }}
`

const minimalConfigGoTemplate = `package main

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

// Config holds the application configuration
type Config struct {
	Port  int  ` + "`" + `env:"PORT,default={{ .ADL.Spec.Server.Port }}"` + "`" + `
	Debug bool ` + "`" + `env:"DEBUG,default={{ .ADL.Spec.Server.Debug }}"` + "`" + `
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

const minimalReadmeTemplate = `# {{ .ADL.Metadata.Name | title }}

{{ .ADL.Metadata.Description }}

**Version:** {{ .ADL.Metadata.Version }}

## Overview

This minimal A2A server was generated using the A2A CLI from an Agent Definition Language (ADL) file. This template provides a basic HTTP server without AI capabilities - perfect for implementing custom business logic.

### Capabilities

{{- if .ADL.Spec.Capabilities }}
- **Streaming:** {{ .ADL.Spec.Capabilities.Streaming }}
- **Push Notifications:** {{ .ADL.Spec.Capabilities.PushNotifications }}
- **State Transition History:** {{ .ADL.Spec.Capabilities.StateTransitionHistory }}
{{- end }}

### Available Endpoints

- ` + "`POST /chat`" + ` - Main chat endpoint
{{- range .ADL.Spec.Tools }}
- ` + "`POST /{{ .Name }}`" + ` - {{ .Description }}
{{- end }}

## Getting Started

### Prerequisites

- Go {{ .ADL.Spec.Language.Go.Version }}+
- [Task](https://taskfile.dev/) (optional, for using Taskfile commands)

### Environment Variables

- ` + "`PORT`" + `: Server port (default: {{ .ADL.Spec.Server.Port }})
- ` + "`DEBUG`" + `: Enable debug mode (default: {{ .ADL.Spec.Server.Debug }})

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
kubectl port-forward svc/{{ .ADL.Metadata.Name }} {{ .ADL.Spec.Server.Port }}:80 -n {{ .ADL.Metadata.Name }}-ns
` + "```" + `

The operator automatically manages deployment, scaling, health checks, and configuration.

## Development

### TODO: Implement Handlers

{{- if .ADL.Spec.Tools }}
The following handlers need implementation in ` + "`handlers.go`" + `:

{{- range .ADL.Spec.Tools }}
- **{{ .Name | title }}Handler**: {{ .Description }}
{{- end }}

Each handler receives HTTP requests and should implement your business logic.
{{- else }}
The main chat handler needs implementation in ` + "`handlers.go`" + `.
{{- end }}

### Project Structure

` + "```" + `
.
├── main.go              # Main server setup
├── handlers.go          # HTTP handlers (⚠️ TODO)
├── config.go            # Configuration management
├── go.mod               # Go module definition
├── Taskfile.yml         # Task definitions
├── Dockerfile           # Container configuration
├── .well-known/
│   └── agent.json       # Agent capabilities (auto-generated)
└── README.md            # This file
` + "```" + `

### Testing

Add tests for your handler implementations:

` + "```bash" + `
# Run tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...
` + "```" + `

### Example Usage

` + "```bash" + `
# Test the chat endpoint
curl -X POST http://localhost:{{ .ADL.Spec.Server.Port }}/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, world!"}'

{{- range .ADL.Spec.Tools }}
# Test the {{ .Name }} endpoint
curl -X POST http://localhost:{{ $.ADL.Spec.Server.Port }}/{{ .Name }} \
  -H "Content-Type: application/json" \
  -d '{}'
{{- end }}
` + "```" + `

## API Endpoints

Once running, the server exposes:

- ` + "`GET /.well-known/agent.json`" + ` - Agent capabilities and metadata
- ` + "`POST /chat`" + ` - Main chat endpoint
{{- range .ADL.Spec.Tools }}
- ` + "`POST /{{ .Name }}`" + ` - {{ .Description }}
{{- end }}
{{- if .ADL.Spec.Capabilities.Streaming }}
- ` + "`POST /stream`" + ` - Streaming endpoint (if implemented)
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
  # Minimal template - no AI agent
  agent:
    enabled: false
{{- if .ADL.Spec.Env }}
  env:
    {{- range .ADL.Spec.Env }}
    - name: "{{ .Name }}"
      {{- if .Value }}
      value: "{{ .Value }}"
      {{- else if .ValueFrom }}
      valueFrom:
        {{- if .ValueFrom.ConfigMapKeyRef }}
        configMapKeyRef:
          name: "{{ .ValueFrom.ConfigMapKeyRef.Name }}"
          key: "{{ .ValueFrom.ConfigMapKeyRef.Key }}"
        {{- else if .ValueFrom.SecretKeyRef }}
        secretKeyRef:
          name: "{{ .ValueFrom.SecretKeyRef.Name }}"
          key: "{{ .ValueFrom.SecretKeyRef.Key }}"
        {{- end }}
      {{- end }}
    {{- end }}
{{- end }}
{{- if .ADL.Spec.Env }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .ADL.Metadata.Name }}-config
  namespace: {{ .ADL.Metadata.Name }}-ns
data:
  {{- range .ADL.Spec.Env }}
  {{- if .Value }}
  {{ .Name }}: "{{ .Value }}"
  {{- end }}
  {{- end }}
{{- end }}
`
