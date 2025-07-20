# A2A CLI

A command-line interface for generating production-ready A2A (Agent-to-Agent) servers from Agent Definition Language (ADL) files.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Overview

The A2A CLI helps you build production-ready A2A agents quickly by generating complete project scaffolding from YAML-based Agent Definition Language (ADL) files. It eliminates boilerplate code and ensures consistent patterns across your agent implementations.

### Key Features

- 🚀 **Rapid Development** - Generate complete projects in seconds
- 📋 **Schema-Driven** - Use YAML ADL files to define your agents
- 🎯 **Multiple Templates** - Choose from minimal, AI-powered, or enterprise templates
- 🔄 **Smart Sync** - Update generated code while preserving your implementations
- ✅ **Validation** - Built-in ADL schema validation
- 🛠️ **Interactive Setup** - Guided project initialization
- 📦 **Production Ready** - Includes Docker, Kubernetes, and monitoring configs

## Installation

### From Source

```bash
git clone https://github.com/inference-gateway/a2a-cli.git
cd a2a-cli
go install .
```

### Using Go Install

```bash
go install github.com/inference-gateway/a2a-cli@latest
```

### Pre-built Binaries

Download pre-built binaries from the [releases page](https://github.com/inference-gateway/a2a-cli/releases).

## Quick Start

### 1. Initialize a New Project

```bash
# Interactive project setup
a2a init my-weather-agent

# Or generate from an existing ADL file
a2a generate --file agent.yaml --output ./my-agent
```

### 2. Implement Your Business Logic

The generated project includes TODO placeholders for your implementations:

```go
// TODO: Implement weather API logic
func GetWeatherTool(ctx context.Context, args map[string]interface{}) (string, error) {
    city := args["city"].(string)
    // TODO: Replace with actual weather API call
    return fmt.Sprintf(`{"city": "%s", "temp": "22°C"}`, city), nil
}
```

### 3. Build and Run

```bash
cd my-weather-agent
task build
task run
```

## Usage

### Commands

| Command | Description |
|---------|-------------|
| `a2a init [name]` | Initialize a new project interactively |
| `a2a generate` | Generate project from ADL file |
| `a2a validate [file]` | Validate an ADL file |
| `a2a sync` | Update generated code preserving implementations |

### Generate Command

```bash
# Generate with default template (ai-powered)
a2a generate --file agent.yaml --output ./my-agent

# Use specific template
a2a generate --file agent.yaml --output ./my-agent --template minimal

# Overwrite existing files
a2a generate --file agent.yaml --output ./my-agent --overwrite
```

### Available Templates

| Template | Description |
|----------|-------------|
| `minimal` | Basic HTTP server without AI capabilities |
| `ai-powered` | Full AI agent with LLM integration (default) |
| `enterprise` | Production-ready with auth, metrics, and monitoring |

## Agent Definition Language (ADL)

ADL files use YAML to define your agent's configuration, capabilities, and tools.

### Example ADL File

```yaml
apiVersion: a2a.dev/v1
kind: Agent
metadata:
  name: weather-agent
  description: "Provides weather information for cities worldwide"
  version: "1.0.0"

spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false

  agent:
    provider: openai
    model: gpt-4o-mini
    systemPrompt: "You are a helpful weather assistant."
    maxTokens: 4096
    temperature: 0.7

  tools:
    - name: get_weather
      description: "Get current weather for a city"
      schema:
        type: object
        properties:
          city: {type: string, description: "City name"}
          country: {type: string, description: "Country code"}
        required: [city]

  server:
    port: 8080
    debug: false

  language:
    go:
      module: "github.com/example/weather-agent"
      version: "1.24"
```

### ADL Schema

The complete ADL schema includes:

- **metadata**: Agent name, description, and version
- **capabilities**: Streaming, notifications, state history
- **agent**: AI provider configuration (OpenAI, Anthropic, etc.)
- **tools**: Function definitions with JSON schemas
- **server**: HTTP server configuration
- **language**: Programming language-specific settings (Go, TypeScript, etc.)

## Generated Project Structure

```
my-agent/
├── main.go              # Main server setup
├── tools.go             # Tool implementations (TODO placeholders)
├── config.go            # Configuration management
├── go.mod               # Go module definition
├── Taskfile.yml         # Development tasks
├── Dockerfile           # Container configuration
├── .well-known/
│   └── agent.json       # Agent capabilities (auto-generated)
└── README.md            # Project documentation
```

### Enterprise Template Additions

```
├── middleware.go        # HTTP middleware (auth, metrics, CORS)
├── metrics.go           # Prometheus metrics
├── logging.go           # Structured logging
├── auth.go              # Authentication logic
├── docker-compose.yml   # Full stack deployment
└── k8s/                 # Kubernetes manifests
    ├── deployment.yaml
    ├── service.yaml
    └── configmap.yaml
```

## Examples

The CLI includes example ADL files in the `examples/` directory:

```bash
# Validate examples
a2a validate examples/weather-agent.yaml
a2a validate examples/minimal-agent.yaml
a2a validate examples/enterprise-agent.yaml

# Generate from examples
a2a generate --file examples/weather-agent.yaml --output ./weather-agent
a2a generate --file examples/minimal-agent.yaml --output ./minimal-agent --template minimal
a2a generate --file examples/enterprise-agent.yaml --output ./enterprise-agent --template enterprise
```

## Development

### Prerequisites

- Go 1.21+
- [Task](https://taskfile.dev/) (optional, for using Taskfile commands)

### Building from Source

```bash
git clone https://github.com/inference-gateway/a2a-cli.git
cd a2a-cli

# Install dependencies
go mod download

# Build
task build

# Run tests
task test

# Format code
task fmt

# Lint
task lint
```

### Testing

```bash
# Run tests
task test

# Test with coverage
task test:coverage

# Test all examples
task examples:test

# Generate all examples
task examples:generate
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `task ci` to ensure everything passes
6. Submit a pull request

## Related Projects

- [A2A Framework](https://github.com/inference-gateway/a2a) - The core A2A framework
- [A2A Examples](https://github.com/inference-gateway/a2a-examples) - Example A2A agents

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://docs.a2a.dev)
- 💬 [Discussions](https://github.com/inference-gateway/a2a/discussions)
- 🐛 [Issues](https://github.com/inference-gateway/a2a-cli/issues)

---

> 🤖 Powered by the [A2A (Agent-to-Agent) framework](https://github.com/inference-gateway/a2a)
