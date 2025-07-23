<div align="center">

# A2A CLI

*A command-line interface for generating production-ready A2A (Agent-to-Agent) servers from Agent Definition Language (ADL) files.*

> ⚠️ **Early Development Warning**: This project is in its early stages of development. Breaking changes are expected and acceptable until we reach a stable version. Use with caution in production environments.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg?style=flat-square)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/inference-gateway/a2a-cli/ci.yml?style=flat-square&logo=github)](https://github.com/inference-gateway/a2a-cli/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/inference-gateway/a2a-cli?style=flat-square)](https://goreportcard.com/report/github.com/inference-gateway/a2a-cli)
[![Release](https://img.shields.io/github/v/release/inference-gateway/a2a-cli?style=flat-square&logo=github)](https://github.com/inference-gateway/a2a-cli/releases)

</div>

## Overview

The A2A CLI helps you build production-ready A2A agents quickly by generating complete project scaffolding from YAML-based Agent Definition Language (ADL) files. It eliminates boilerplate code and ensures consistent patterns across your agent implementations.

### Key Features

- 🚀 **Rapid Development** - Generate complete projects in seconds
- 📋 **Schema-Driven** - Use YAML ADL files to define your agents
- 🎯 **Production Ready** - Single unified template with AI integration and enterprise features
- � **Smart Ignore** - Protect your implementations with .a2a-ignore files
- ✅ **Validation** - Built-in ADL schema validation
- 🛠️ **Interactive Setup** - Guided project initialization
- 📦 **Production Ready** - Includes Docker, Kubernetes, and monitoring configs

## Installation

### Quick Install (Recommended)

Use our install script to automatically download and install the latest binary:

```bash
curl -fsSL https://raw.githubusercontent.com/inference-gateway/a2a-cli/main/install.sh | bash
```

Or download and run the script manually:

```bash
wget https://raw.githubusercontent.com/inference-gateway/a2a-cli/main/install.sh
chmod +x install.sh
./install.sh
```

**Install Options:**

- Install specific version: `./install.sh --version v1.0.0`
- Custom install directory: `INSTALL_DIR=~/bin ./install.sh`
- Show help: `./install.sh --help`


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

### Generate Command

```bash
# Generate project from ADL file
a2a generate --file agent.yaml --output ./my-agent

# Overwrite existing files (respects .a2a-ignore)
a2a generate --file agent.yaml --output ./my-agent --overwrite

# Generate with CI/CD workflow configuration
a2a generate --file agent.yaml --output ./my-agent --ci
```

#### Generate Flags

| Flag | Description |
|------|-------------|
| `--file`, `-f` | ADL file to generate from (default: "agent.yaml") |
| `--output`, `-o` | Output directory for generated code (default: ".") |
| `--template`, `-t` | Template to use (default: "minimal") |
| `--overwrite` | Overwrite existing files (respects .a2a-ignore) |
| `--devcontainer` | Generate VS Code devcontainer configuration |
| `--ci` | Generate CI/CD workflow configuration |


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
          city:
            - type: string
              description: "City name"
          country:
            - type: string
              description: "Country code"
        required:
          - city
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
├── .a2a-ignore          # Files to protect from regeneration
├── .well-known/
│   └── agent.json       # Agent capabilities (auto-generated)
└── README.md            # Project documentation
```

### CI/CD Integration

When using the `--ci` flag, the A2A CLI generates GitHub Actions workflows for your project:

```bash
# Generate project with CI/CD workflow
a2a generate --file agent.yaml --output ./my-agent --ci
```

This creates a comprehensive GitHub Actions workflow (`.github/workflows/ci.yml`) that includes:

- **Automated Testing**: Runs all tests on every push and pull request
- **Code Quality**: Format checking and linting
- **Multi-Environment**: Supports main and develop branches
- **Build Artifacts**: Automatically builds and uploads binaries
- **Caching**: Go module caching for faster builds
- **Task Integration**: Uses the generated Taskfile for consistent build steps

The generated workflow automatically detects your Go version from the ADL file and configures the appropriate environment.


## Examples

The CLI includes example ADL files in the `examples/` directory:

```bash
# Validate example
a2a validate examples/minimal-agent.yaml

# Generate from example
a2a generate --file examples/minimal-agent.yaml --output ./minimal-agent
```

## Customizing Generation with .a2a-ignore

The A2A CLI automatically creates a `.a2a-ignore` file during project generation to protect files containing TODO implementations. This file works similar to `.gitignore` and prevents important implementation files from being overwritten during subsequent generations.

### Automatically Protected Files

When you generate a project, implementation files are automatically added to `.a2a-ignore` to protect your business logic from being overwritten during regeneration.

You can control which additional files are generated or updated by editing the `.a2a-ignore` file:

```bash
# .a2a-ignore
# Skip Docker-related files if you have custom containerization
Dockerfile
docker-compose.yml

# Skip Kubernetes manifests if you use different deployment tools
k8s/

# Skip specific generated files you want to customize
middleware.go
auth.go

# Skip build configuration if you have custom setup
Taskfile.yml
```

### .a2a-ignore Patterns

- Use `#` for comments
- Use `/` at the end to match directories
- Use `*` for wildcards
- Exact file paths or glob patterns
- Protects files during all `generate` operations

### Common Use Cases

- **Custom Deployment**: Skip `Dockerfile`, `k8s/`, `docker-compose.yml`
- **Custom Build**: Skip `Taskfile.yml`, `Makefile`
- **Custom Auth**: Skip `auth.go`, `middleware.go`
- **Custom Documentation**: Skip `README.md`

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

## Roadmap

### Language Support

The A2A CLI currently supports Go, with plans to expand to additional programming languages:

#### ✅ Currently Supported
- **Go** - Full support with unified template

#### 🚧 Planned Support
- **TypeScript/Node.js** - Complete A2A agent generation with Express.js framework
  - AI-powered agents with OpenAI/Anthropic integration
  - Enterprise features (auth, metrics, logging)
  - Docker and Kubernetes deployment configs
  
- **Rust** - High-performance A2A agents with async support
  - Tokio-based async runtime
  - Enterprise-grade performance and safety
  - WebAssembly (WASM) compilation support

- **Python** - Rapid prototyping and AI-first development
  - FastAPI-based server generation
  - Rich AI ecosystem integration
  - Jupyter notebook support for development

#### 🔮 Future Considerations
- **Java/Kotlin** - Enterprise JVM support
- **C#/.NET** - Microsoft ecosystem integration
- **Swift** - Apple ecosystem and server-side Swift

### Template Enhancements

- **Multi-language projects** - Generate polyglot agents with language-specific microservices
- **Custom templates** - User-defined project templates and scaffolding
- **Plugin system** - Extensible architecture for custom generators
- **Cloud-native templates** - Serverless (AWS Lambda, Vercel) and edge deployment support

### Contribute to the Roadmap

We welcome community input on our roadmap! Please:
- 💡 Suggest new languages or frameworks via [Issues](https://github.com/inference-gateway/a2a-cli/issues)
- 🤝 Contribute implementations for new languages (see [Contributing Guide](CONTRIBUTING.md))

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://docs.inference-gateway.com)
- 💬 [Discussions](https://github.com/inference-gateway/a2a-cli/discussions)
- 🐛 [Issues](https://github.com/inference-gateway/a2a-cli/issues)

---

> 🤖 Powered by the [Inference Gateway framework](https://github.com/inference-gateway/)
