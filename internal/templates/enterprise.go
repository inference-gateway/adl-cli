package templates

// getEnterpriseTemplate returns the enterprise template files
func getEnterpriseTemplate() map[string]string {
	files := getAIPoweredTemplate() // Start with ai-powered template

	// Add enterprise-specific files
	files["middleware.go"] = middlewareGoTemplate
	files["metrics.go"] = metricsGoTemplate
	files["logging.go"] = loggingGoTemplate
	files["auth.go"] = authGoTemplate
	files["tool_metrics.go"] = toolMetricsGoTemplate
	files["docker-compose.yml"] = dockerComposeTemplate
	files["k8s/a2a-server.yaml"] = a2aOperatorTemplate
	files["k8s/namespace.yaml"] = namespaceTemplate

	// Override main.go for enterprise features
	files["main.go"] = enterpriseMainGoTemplate
	files["config.go"] = enterpriseConfigGoTemplate
	files["README.md"] = enterpriseReadmeTemplate

	return files
}

const enterpriseMainGoTemplate = `package main

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
	// Initialize logging
	logger := NewStructuredLogger()
	
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize metrics
	metrics := NewMetrics()
	metrics.Start()

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
	// Add tools with metrics
	toolbox := adk.NewToolbox()
	{{- range .ADL.Spec.Tools }}
	toolbox.AddTool("{{ .Name }}", WithMetrics({{ .Name | title }}Tool, metrics))
	{{- end }}
	agentBuilder.SetToolbox(toolbox)
	{{- end }}

	agent, err := agentBuilder.Build()
	if err != nil {
		logger.Fatal("Failed to build agent", "error", err)
	}

	serverBuilder.SetAgent(agent)
	{{- end }}

	// Configure server with enterprise features
	serverBuilder.SetPort(config.Port)
	serverBuilder.SetDebug(config.Debug)

	// Add middleware
	serverBuilder.AddMiddleware(LoggingMiddleware(logger))
	serverBuilder.AddMiddleware(MetricsMiddleware(metrics))
	{{- if .ADL.Spec.Server.Auth.Enabled }}
	serverBuilder.AddMiddleware(AuthMiddleware(config.Auth))
	{{- end }}
	serverBuilder.AddMiddleware(CORSMiddleware())
	serverBuilder.AddMiddleware(RateLimitMiddleware(config.RateLimit))

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

	// Add health check endpoint
	serverBuilder.SetCustomHandler("/health", NewHealthHandler(metrics))
	serverBuilder.SetCustomHandler("/metrics", metrics.Handler())

	// Build server
	srv, err := serverBuilder.Build()
	if err != nil {
		logger.Fatal("Failed to build server", "error", err)
	}

	// Start server
	logger.Info("Starting {{ .ADL.Metadata.Name }} agent", 
		"port", config.Port,
		"version", "{{ .ADL.Metadata.Version }}",
		"env", config.Environment)
		
	if err := srv.Run(ctx); err != nil {
		logger.Fatal("Server failed", "error", err)
	}

	logger.Info("{{ .ADL.Metadata.Name }} agent stopped gracefully")
}
`

const middlewareGoTemplate = `package main

import (
	"net/http"
	"time"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(wrapped, r)
			
			logger.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", time.Since(start),
				"user_agent", r.UserAgent(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// MetricsMiddleware collects request metrics
func MetricsMiddleware(metrics *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)
			
			duration := time.Since(start)
			metrics.RecordRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// AuthMiddleware handles authentication
func AuthMiddleware(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health and metrics endpoints
			if r.URL.Path == "/health" || r.URL.Path == "/metrics" || r.URL.Path == "/.well-known/agent.json" {
				next.ServeHTTP(w, r)
				return
			}

			// TODO: Implement authentication logic
			// Example: Bearer token validation
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// TODO: Validate token
			// if !validateToken(authHeader) {
			//     http.Error(w, "Invalid token", http.StatusUnauthorized)
			//     return
			// }

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(config RateLimitConfig) func(http.Handler) http.Handler {
	// TODO: Implement rate limiting
	// This is a placeholder implementation
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement rate limiting logic
			next.ServeHTTP(w, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
`

const metricsGoTemplate = `package main

import (
	"net/http"
	"time"
)

// Metrics handles application metrics
type Metrics struct {
	// TODO: Add metrics implementation
	// Example: Prometheus metrics
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// Start initializes metrics collection
func (m *Metrics) Start() {
	// TODO: Initialize metrics
}

// RecordRequest records HTTP request metrics
func (m *Metrics) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	// TODO: Record request metrics
}

// Handler returns the metrics HTTP handler
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Expose metrics endpoint
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("# TODO: Implement metrics endpoint\n"))
	})
}

// HealthHandler provides health check endpoint
type HealthHandler struct {
	metrics *Metrics
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(metrics *Metrics) *HealthHandler {
	return &HealthHandler{metrics: metrics}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement health checks
	health := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
		"version": "{{ .ADL.Metadata.Version }}",
	}
	
	w.Header().Set("Content-Type", "application/json")
	// TODO: Marshal and write health response
	w.Write([]byte(` + "`" + `{"status":"healthy"}` + "`" + `))
}
`

const loggingGoTemplate = `package main

import (
	"log/slog"
	"os"
)

// StructuredLogger provides structured logging
type StructuredLogger struct {
	*slog.Logger
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger() *StructuredLogger {
	// TODO: Configure structured logging
	// Example: JSON output, different log levels
	
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	
	logger := slog.New(handler)
	
	return &StructuredLogger{
		Logger: logger,
	}
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(msg string, args ...interface{}) {
	l.Error(msg, args...)
	os.Exit(1)
}
`

const authGoTemplate = `package main

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled    bool   ` + "`" + `env:"AUTH_ENABLED,default=false"` + "`" + `
	JWTSecret  string ` + "`" + `env:"JWT_SECRET"` + "`" + `
	APIKeyHeader string ` + "`" + `env:"API_KEY_HEADER,default=X-API-Key"` + "`" + `
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled bool ` + "`" + `env:"RATE_LIMIT_ENABLED,default=false"` + "`" + `
	RequestsPerMinute int ` + "`" + `env:"RATE_LIMIT_RPM,default=60"` + "`" + `
}

// TODO: Implement authentication functions
// func validateToken(authHeader string) bool { ... }
// func parseJWT(token string) (*Claims, error) { ... }
`

const dockerComposeTemplate = `version: '3.8'

services:
  {{ .ADL.Metadata.Name }}:
    build: .
    ports:
      - "{{ .ADL.Spec.Server.Port }}:{{ .ADL.Spec.Server.Port }}"
    environment:
      - PORT={{ .ADL.Spec.Server.Port }}
      - DEBUG={{ .ADL.Spec.Server.Debug }}
      - ENVIRONMENT=production
      {{- if .ADL.Spec.Agent }}
      {{- if eq .ADL.Spec.Agent.Provider "openai" }}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      {{- else if eq .ADL.Spec.Agent.Provider "anthropic" }}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      {{- end }}
      {{- end }}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:{{ .ADL.Spec.Server.Port }}/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana

volumes:
  grafana-storage:
`

const a2aOperatorTemplate = `---
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
      enabled: true
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
  env:
    - name: "ENVIRONMENT"
      value: "production"
    {{- if .ADL.Spec.Server.Debug }}
    - name: "DEBUG"
      value: "true"
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

const namespaceTemplate = `apiVersion: v1
kind: Namespace
metadata:
  name: {{ .ADL.Metadata.Name }}-ns
  labels:
    inference-gateway.com/managed: "true"
    app: {{ .ADL.Metadata.Name }}
`

const enterpriseConfigGoTemplate = `package main

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

// Config holds the application configuration
type Config struct {
	Port        int    ` + "`" + `env:"PORT,default={{ .ADL.Spec.Server.Port }}"` + "`" + `
	Debug       bool   ` + "`" + `env:"DEBUG,default={{ .ADL.Spec.Server.Debug }}"` + "`" + `
	Environment string ` + "`" + `env:"ENVIRONMENT,default=development"` + "`" + `

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

	// Enterprise features
	Auth      AuthConfig      ` + "`" + `env:",prefix=AUTH_"` + "`" + `
	RateLimit RateLimitConfig ` + "`" + `env:",prefix=RATE_LIMIT_"` + "`" + `
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

const toolMetricsGoTemplate = `package main

import (
	"context"
	"time"

	"github.com/inference-gateway/a2a/adk"
)

// WithMetrics wraps a tool function with metrics collection
func WithMetrics(toolFunc adk.ToolFunc, metrics *Metrics) adk.ToolFunc {
	return func(ctx context.Context, args map[string]interface{}) (string, error) {
		start := time.Now()
		
		result, err := toolFunc(ctx, args)
		
		duration := time.Since(start)
		// TODO: Record tool execution metrics
		_ = duration
		
		return result, err
	}
}
`

const enterpriseReadmeTemplate = `# {{ .ADL.Metadata.Name | title }} (Enterprise)

{{ .ADL.Metadata.Description }}

**Version:** {{ .ADL.Metadata.Version }}

## Overview

This enterprise-grade A2A agent was generated using the A2A CLI from an Agent Definition Language (ADL) file. This template includes production-ready features like structured logging, metrics, authentication, and monitoring.

### Enterprise Features

- ✅ **Structured Logging** - JSON structured logs with contextual information
- ✅ **Metrics & Monitoring** - Request metrics and health checks
- ✅ **Authentication** - Configurable API key/JWT authentication
- ✅ **Rate Limiting** - Request rate limiting and throttling
- ✅ **CORS Support** - Cross-origin request handling
- ✅ **Health Checks** - Kubernetes-ready health and readiness probes
- ✅ **Docker & K8s** - Production deployment configurations
- ✅ **Observability** - Prometheus metrics and Grafana dashboards

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
- Docker & Docker Compose (for containerized deployment)
- Kubernetes (for production deployment)

### Environment Variables

#### Core Configuration
- ` + "`PORT`" + `: Server port (default: {{ .ADL.Spec.Server.Port }})
- ` + "`DEBUG`" + `: Enable debug mode (default: {{ .ADL.Spec.Server.Debug }})
- ` + "`ENVIRONMENT`" + `: Environment name (development/staging/production)

{{- if .ADL.Spec.Agent }}
#### AI Provider Configuration
{{- if eq .ADL.Spec.Agent.Provider "openai" }}
- ` + "`OPENAI_API_KEY`" + `: Your OpenAI API key (required)
{{- else if eq .ADL.Spec.Agent.Provider "anthropic" }}
- ` + "`ANTHROPIC_API_KEY`" + `: Your Anthropic API key (required)
{{- else if eq .ADL.Spec.Agent.Provider "azure" }}
- ` + "`AZURE_OPENAI_ENDPOINT`" + `: Your Azure OpenAI endpoint (required)
- ` + "`AZURE_OPENAI_API_KEY`" + `: Your Azure OpenAI API key (required)
{{- else if eq .ADL.Spec.Agent.Provider "ollama" }}
- ` + "`OLLAMA_BASE_URL`" + `: Ollama server URL (default: http://localhost:11434)
{{- end }}
{{- end }}

#### Enterprise Features
- ` + "`AUTH_ENABLED`" + `: Enable authentication (default: false)
- ` + "`AUTH_JWT_SECRET`" + `: JWT secret for token validation
- ` + "`AUTH_API_KEY_HEADER`" + `: API key header name (default: X-API-Key)
- ` + "`RATE_LIMIT_ENABLED`" + `: Enable rate limiting (default: false)
- ` + "`RATE_LIMIT_RPM`" + `: Requests per minute limit (default: 60)

### Local Development

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

# Run with live reload
task dev
` + "```" + `

#### Using Go directly

` + "```bash" + `
# Run directly
go run .

# Build and run
go build -o bin/{{ .ADL.Metadata.Name }} .
./bin/{{ .ADL.Metadata.Name }}
` + "```" + `

### Docker Deployment

#### Single Container

` + "```bash" + `
# Build Docker image
task docker-build

# Run container
task docker-run
` + "```" + `

#### Docker Compose (with monitoring)

` + "```bash" + `
# Start the full stack
docker-compose up -d

# View logs
docker-compose logs -f {{ .ADL.Metadata.Name }}

# Stop the stack
docker-compose down
` + "```" + `

This includes:
- {{ .ADL.Metadata.Name }} agent
- Prometheus metrics collection
- Grafana dashboards

Access:
- Agent: http://localhost:{{ .ADL.Spec.Server.Port }}
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

### Kubernetes Deployment

This project includes manifests for the [Inference Gateway Operator](https://github.com/inference-gateway/operator) for production-ready Kubernetes deployment.

` + "```bash" + `
# Install the Inference Gateway Operator (if not already installed)
kubectl apply -f https://github.com/inference-gateway/operator/releases/latest/download/install.yaml

# Apply A2A Custom Resource
kubectl apply -f k8s/a2a-server.yaml

# Check A2A status
kubectl get a2a {{ .ADL.Metadata.Name }} -n {{ .ADL.Metadata.Name }}-ns

# View operator-managed deployment
kubectl get pods -n {{ .ADL.Metadata.Name }}-ns

# View logs
kubectl logs -l app={{ .ADL.Metadata.Name }} -n {{ .ADL.Metadata.Name }}-ns -f

# Port forward for testing
kubectl port-forward svc/{{ .ADL.Metadata.Name }} {{ .ADL.Spec.Server.Port }}:80 -n {{ .ADL.Metadata.Name }}-ns
` + "```" + `

The operator automatically manages:
- Deployment scaling and updates
- Service mesh integration
- TLS certificate management
- Health checks and monitoring
- Configuration management

## Development

### TODO: Enterprise Implementation

1. **Authentication** (` + "`auth.go`" + `):
   - Implement JWT token validation
   - Add API key authentication
   - Configure role-based access control

2. **Metrics** (` + "`metrics.go`" + `):
   - Add Prometheus metrics collection
   - Implement custom business metrics
   - Configure alerting rules

3. **Monitoring**:
   - Set up Grafana dashboards
   - Configure Prometheus alerting
   - Add distributed tracing

{{- if .ADL.Spec.Tools }}
4. **Tools** (` + "`tools.go`" + `):
{{- range .ADL.Spec.Tools }}
   - **{{ .Name | title }}**: {{ .Description }}
{{- end }}
{{- end }}

### Project Structure

` + "```" + `
.
├── main.go              # Main server with enterprise features
├── tools.go             # Tool implementations (⚠️ TODO)
├── config.go            # Enterprise configuration
├── middleware.go        # HTTP middleware (auth, metrics, etc.)
├── metrics.go           # Metrics collection
├── logging.go           # Structured logging
├── auth.go              # Authentication logic
├── go.mod               # Go module definition
├── Taskfile.yml         # Task definitions
├── Dockerfile           # Container configuration
├── docker-compose.yml   # Full stack deployment
├── k8s/                 # Kubernetes manifests
│   ├── deployment.yaml
│   ├── service.yaml
│   └── configmap.yaml
├── .well-known/
│   └── agent.json       # Agent capabilities (auto-generated)
└── README.md            # This file
` + "```" + `

### Monitoring & Observability

#### Health Checks
- ` + "`GET /health`" + ` - Application health status
- ` + "`GET /metrics`" + ` - Prometheus metrics endpoint

#### Logs
Structured JSON logs include:
- Request/response logging
- Error tracking
- Performance metrics
- Business event logging

#### Metrics
Available Prometheus metrics:
- HTTP request duration
- Request count by status code
- Tool execution metrics
- Custom business metrics

### Testing

` + "```bash" + `
# Run unit tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run integration tests
task test:integration

# Load testing
task test:load
` + "```" + `

### Security

- ✅ Input validation and sanitization
- ✅ Rate limiting to prevent abuse
- ✅ Authentication middleware
- ✅ CORS configuration
- ✅ Security headers
- ⚠️ TODO: Implement TLS/HTTPS
- ⚠️ TODO: Add input validation
- ⚠️ TODO: Configure security scanning

## API Endpoints

### Core Endpoints
- ` + "`GET /.well-known/agent.json`" + ` - Agent capabilities and metadata
- ` + "`POST /chat`" + ` - Chat with the agent
{{- if .ADL.Spec.Capabilities.Streaming }}
- ` + "`POST /stream`" + ` - Streaming chat endpoint
{{- end }}

### Enterprise Endpoints
- ` + "`GET /health`" + ` - Health check endpoint
- ` + "`GET /metrics`" + ` - Prometheus metrics
- ` + "`GET /ready`" + ` - Readiness probe

### Example Requests

` + "```bash" + `
# Health check
curl http://localhost:{{ .ADL.Spec.Server.Port }}/health

# Chat with authentication
curl -X POST http://localhost:{{ .ADL.Spec.Server.Port }}/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{"message": "Hello, world!"}'

# Get metrics
curl http://localhost:{{ .ADL.Spec.Server.Port }}/metrics
` + "```" + `

## Production Deployment

### Environment Setup
1. Configure environment variables
2. Set up SSL/TLS certificates
3. Configure load balancer
4. Set up monitoring and alerting
5. Configure log aggregation

### Scaling
- Horizontal scaling via Kubernetes
- Load balancing across instances
- Auto-scaling based on metrics
- Database connection pooling

### Monitoring
- Application performance monitoring
- Error tracking and alerting
- Resource utilization monitoring
- Business metrics dashboards

## Generated Files

This project was generated with:
- **CLI Version:** {{ .Metadata.CLIVersion }}
- **Template:** {{ .Metadata.Template }} (Enterprise)
- **Generated At:** {{ .Metadata.GeneratedAt.Format "2006-01-02 15:04:05" }}

To regenerate or sync with ADL changes:
` + "```bash" + `
a2a sync --file agent.yaml
` + "```" + `

---

> 🏢 Enterprise-grade agent powered by the [A2A (Agent-to-Agent) framework](https://github.com/inference-gateway/a2a)
`
