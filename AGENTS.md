# AGENTS.md

This file provides comprehensive guidance for AI agents working with the ADL CLI project. It covers project architecture, development workflow, testing strategies, and conventions.

## Project Overview

**ADL CLI** (Agent Definition Language Command Line Interface) is a Go-based tool for generating enterprise-ready A2A (Agent-to-Agent) servers from YAML-based Agent Definition Language (ADL) files. It eliminates boilerplate code and ensures consistent patterns across agent implementations.

### Key Technologies
- **Language**: Go 1.25+
- **Framework**: Cobra CLI framework
- **Templating**: Go templates with Sprig functions
- **Build System**: Taskfile for automation
- **CI/CD**: GitHub Actions with semantic-release
- **Containerization**: Docker with multi-platform support

## Architecture and Structure

### Project Layout
```
adl-cli/
├── cmd/                    # CLI command implementations
│   ├── root.go            # Main CLI setup
│   ├── generate.go        # Generate command
│   ├── init.go            # Init command
│   ├── validate.go        # Validate command
│   └── *_test.go          # Command tests
├── internal/              # Internal packages
│   ├── generator/         # Code generation engine
│   ├── schema/           # ADL schema definitions and validation
│   ├── templates/        # Template system
│   │   ├── common/       # Cross-language templates
│   │   ├── languages/    # Language-specific templates
│   │   └── sandbox/      # Environment templates
│   └── prompt/           # Interactive prompt system
├── examples/              # Example ADL files
├── .github/workflows/     # CI/CD workflows
├── Taskfile.yml          # Build automation
├── go.mod                # Go dependencies
└── main.go               # Application entry point
```

### Core Components

1. **CLI Interface** (`cmd/`): Cobra-based command structure with `init`, `generate`, and `validate` commands
2. **Template Engine** (`internal/templates/`): Multi-language template system with support for Go, Rust, and TypeScript
3. **Schema System** (`internal/schema/`): YAML schema validation using JSON Schema
4. **Generator** (`internal/generator/`): File generation engine with `.adl-ignore` support
5. **Prompt System** (`internal/prompt/`): Interactive CLI prompts for project initialization

## Development Environment Setup

### Prerequisites
- Go 1.25+ (as specified in `go.mod`)
- [Task](https://taskfile.dev/) for build automation
- Docker (for container operations)
- Git for version control

### Quick Setup
```bash
# Clone the repository
git clone https://github.com/inference-gateway/adl-cli.git
cd adl-cli

# Install dependencies
task mod

# Build the CLI
task build

# Run tests
task test
```

### Development Tools
- **Task Runner**: Use `task <command>` for all development operations
- **Linting**: `golangci-lint` configured in CI workflow
- **Formatting**: `go fmt` and `prettier` for consistent code style
- **Testing**: Go's built-in testing framework with coverage support

## Key Commands

### Build and Development
```bash
# Build the ADL CLI binary
task build                    # Build to bin/adl
task install                  # Install to GOPATH/bin

# Development mode
task dev -- <args>           # Run CLI with arguments
task dev -- init my-agent    # Interactive project setup
```

### Code Quality
```bash
# Format code
task fmt                     # Format Go code
task format                  # Run prettier formatter

# Lint and vet
task lint                    # Run golangci-lint
task vet                     # Run go vet

# Complete CI pipeline
task ci                      # fmt → lint → test → build
```

### Testing
```bash
# Run tests
task test                    # Run all tests
task test:coverage           # Tests with coverage

# Example validation
task examples:test           # Validate all example ADL files
task examples:generate       # Generate projects from examples
```

### Release and Distribution
```bash
# Build release binaries
task release                 # Multi-platform builds via goreleaser

# Docker operations
task docker:build           # Build Docker image
```

## Testing Instructions

### Test Structure
- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test command execution and file generation
- **Example Tests**: Validate example ADL files and generated output

### Running Tests
```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./cmd
go test -v ./internal/generator
go test -v ./internal/schema

# Run with coverage
go test -v -cover ./...
```

### Test Patterns
1. **Table-driven tests**: Use for comprehensive test coverage
2. **Golden files**: Compare generated output with expected results
3. **Mock file system**: Test file operations without touching disk
4. **Command testing**: Test CLI commands with different arguments

### Example Testing
```bash
# Validate example ADL files
adl validate examples/go-agent.yaml
adl validate examples/rust-agent.yaml
adl validate examples/cloudrun-agent.yaml

# Generate and test example projects
rm -rf test-output
mkdir -p test-output
adl generate --file examples/go-agent.yaml --output test-output/go-agent --overwrite --ci
```

## Project Conventions and Coding Standards

### Go Code Style
- Follow standard Go conventions and `gofmt` formatting
- Use `golangci-lint` for linting with project-specific configuration
- Prefer early returns to avoid deep nesting
- Use table-driven tests for comprehensive coverage
- Always add a new line at the end of files
- Use lowercase log messages
- Code to interfaces for easier testing

### Commit Messages
Use conventional commits format:
```
type(scope): Description

feat(generator): Add CloudRun deployment support
fix(schema): Resolve validation issue with nested config
chore(deps): Update Go dependencies
```

**Types**: `feat`, `fix`, `refactor`, `perf`, `ci`, `docs`, `style`, `test`, `build`, `chore`, `security`

### File Organization
- **Generated Files**: Mark with "DO NOT EDIT" headers
- **Internal Packages**: Use `internal/` for non-public code
- **Test Files**: Place `*_test.go` files alongside implementation
- **Configuration**: Centralize in root directory files

### Template Development
When modifying templates:
1. Update templates in `internal/templates/`
2. Add corresponding tests
3. Update example files if needed
4. Test generation with example ADL files
5. Ensure backward compatibility or document breaking changes

## Important Files and Configurations

### Configuration Files
- **`Taskfile.yml`**: Build automation and development tasks
- **`.goreleaser.yaml`**: Multi-platform binary distribution
- **`.releaserc.yaml`**: Semantic-release configuration
- **`.github/workflows/`**: CI/CD pipeline definitions
- **`go.mod`**: Go module dependencies

### Key Source Files
- **`cmd/root.go`**: Main CLI setup and version management
- **`internal/schema/types.go`**: ADL type definitions
- **`internal/templates/engine.go`**: Template execution engine
- **`internal/generator/generator.go`**: File generation logic

### Template Structure
```
internal/templates/
├── common/              # Cross-language templates
│   ├── ai/             # AI assistant instructions
│   ├── config/         # Configuration templates
│   ├── docker/         # Docker configurations
│   ├── docs/           # Documentation templates
│   ├── github/         # GitHub workflows
│   ├── kubernetes/     # K8s deployment
│   └── taskfile/       # Taskfile templates
├── languages/          # Language-specific templates
│   ├── go/            # Go project templates
│   ├── rust/          # Rust project templates
│   └── typescript/    # TypeScript templates (planned)
└── sandbox/           # Development environments
    ├── devcontainer/  # VS Code DevContainer
    └── flox/          # Flox environment
```

## Development Workflow

### 1. Making Changes
```bash
# Create feature branch
git checkout -b feat/feature-name

# Make changes and test
task ci                     # Run full CI pipeline

# Commit with conventional commit
git commit -m "feat(scope): Description"

# Push and create PR
git push origin feat/feature-name
```

### 2. Testing Changes
```bash
# Test command changes
task dev -- init test-agent
task dev -- generate --file examples/go-agent.yaml --output test-output

# Test template changes
rm -rf test-output
task examples:generate
ls -la test-output/go-agent/

# Run all tests
task test
```

### 3. Release Process
1. Merge changes to `main` branch
2. Run `task release` to build binaries
3. Use semantic-release for versioning and changelog
4. GitHub Actions automatically creates releases

## ADL Schema and Generation

### ADL File Structure
```yaml
apiVersion: adl.dev/v1
kind: Agent
metadata:
  name: agent-name
  description: Agent description
  version: "1.0.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: false
  agent:
    provider: "openai"
    model: "gpt-4o-mini"
    systemPrompt: "You are a helpful assistant."
    maxTokens: 4096
    temperature: 0.7
  skills:
    - name: get_weather
      description: "Get current weather"
      schema:
        type: object
        properties:
          city:
            type: string
            description: "City name"
        required:
          - city
  server:
    port: 8080
    debug: false
  language:
    go:
      module: "github.com/user/agent"
      version: "1.25"
```

### Generation Features
- **Multi-language support**: Go, Rust, TypeScript (planned)
- **CI/CD integration**: GitHub Actions workflows
- **Deployment options**: Kubernetes, CloudRun
- **Sandbox environments**: Flox, DevContainer
- **Service injection**: Type-safe dependency injection
- **Configuration management**: Environment variable mapping
- **Artifacts support**: Filesystem and MinIO storage

## Error Handling and Debugging

### Common Issues
1. **Template errors**: Check template syntax and context variables
2. **Schema validation**: Use `adl validate` to debug ADL files
3. **File permissions**: Ensure write access to output directories
4. **Dependency issues**: Run `task mod` to update dependencies

### Debugging Commands
```bash
# Validate ADL schema
adl validate agent.yaml

# Debug generation
adl generate --file agent.yaml --output ./test --verbose

# Check generated structure
tree test-output/go-agent/

# Test with minimal example
adl init test-minimal --defaults
adl generate --file agent.yaml --output test-minimal
```

### Logging and Diagnostics
- Use `--verbose` flag for detailed output
- Check `test-output/` directory for generation results
- Review GitHub Actions logs for CI/CD issues
- Use Go's built-in testing `-v` flag for test output

## Security Considerations

### Protected Paths
- `.infer/config.yaml`: Inference configuration
- `.infer/*.db`: Database files
- `.infer/shortcuts/`: Shortcut configurations
- `.infer/agents.yaml`: Agent configurations
- `.infer/mcp.yaml`: MCP configurations
- `.git/`: Git repository data
- `*.env`: Environment files

### Security Best Practices
1. Never commit secrets or API keys
2. Use environment variables for sensitive configuration
3. Validate all user input in ADL files
4. Follow principle of least privilege for file operations
5. Use `.adl-ignore` to protect custom implementations

## Performance Optimization

### Build Optimization
- Use `-trimpath` flag for reproducible builds
- Enable Go module caching in CI
- Parallelize test execution where possible
- Use `--skip=publish` for local release testing

### Generation Optimization
- Batch file operations when possible
- Use template caching for repeated generations
- Minimize file I/O operations
- Validate ADL before generation to avoid wasted work

## Contributing Guidelines

### Adding New Features
1. **Language Support**: Add templates to `internal/templates/languages/`
2. **New Commands**: Add to `cmd/` with tests
3. **Template Features**: Update engine and add tests
4. **Schema Extensions**: Update `internal/schema/types.go`

### Testing Requirements
- Add unit tests for new functionality
- Update example files if schema changes
- Test generation with example ADL files
- Ensure backward compatibility or document breaking changes

### Documentation Updates
- Update README.md for user-facing changes
- Update AGENTS.md for agent workflow changes
- Update example files for new features
- Add changelog entries for significant changes

## Resources and References

### Documentation
- [README.md](./README.md): User documentation and examples
- [CLAUDE.md](./CLAUDE.md): AI assistant instructions
- [CONTRIBUTING.md](./CONTRIBUTING.md): Contribution guidelines
- [CHANGELOG.md](./CHANGELOG.md): Version history

### External Resources
- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Taskfile Documentation](https://taskfile.dev/)
- [A2A Protocol](https://github.com/inference-gateway/adk)

### Support
- GitHub Issues: Bug reports and feature requests
- GitHub Discussions: Questions and community support
- Documentation: Project documentation and examples

---

**Last Updated**: Based on project analysis on January 27, 2026  
**Project Version**: 0.26.4 (from Taskfile.yml)  
**Go Version**: 1.25.5 (from go.mod)  
**Main Branch**: main
