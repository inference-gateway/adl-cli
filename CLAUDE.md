# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The ADL CLI is a command-line tool for generating production-ready A2A (Agent-to-Agent) servers from YAML-based Agent Definition Language (ADL) files. It creates complete project scaffolding with business logic placeholders, allowing developers to focus on implementing agent functionality.

## Key Architecture Components

### Core Components
- **CLI Commands**: Located in `cmd/` directory with Cobra-based command structure
- **Generator**: `internal/generator/generator.go` - Main generation logic for creating projects from ADL files
- **Schema**: `internal/schema/types.go` - Go structs representing ADL YAML structure
- **Templates**: `internal/templates/` - Template engine and language-specific templates (Go templates with Sprig functions)
- **Validation**: `internal/schema/validator.go` - ADL file validation against JSON schema

### Template System
- **Engine**: `internal/templates/engine.go` - Template execution with Sprig functions and auto-generated headers
- **Template**: Single unified template that generates production-ready agents with AI integration and enterprise features
- **Context**: Template receives ADL data and metadata for rendering
- **Ignore System**: `.adl-ignore` files protect user implementations from regeneration

### ADL Structure
ADL files define agents with:
- **Metadata**: Name, description, version
- **Capabilities**: Streaming, notifications, state history
- **Agent**: AI provider configuration (OpenAI, Anthropic, etc.)
- **Tools**: Function definitions with JSON schemas  
- **Server**: HTTP server configuration
- **Language**: Programming language settings (currently Go, planned TypeScript/Rust/Python)

## Development Commands

### Build and Development
```bash
# Build the CLI
task build

# Install to GOPATH/bin
task install

# Run in development mode
task dev

# Clean build artifacts
task clean
```

### Testing and Quality
```bash
# Run tests
task test

# Run tests with coverage
task test:coverage

# Format code
task fmt

# Run linter (golangci-lint)
task lint

# Full CI pipeline (fmt, lint, test, build)
task ci
```

### Example Testing
```bash
# Validate all example ADL files
task examples:test

# Generate all example projects
task examples:generate
```

### Release
```bash
# Build release binaries for multiple platforms
task release

# Build Docker image
task docker:build
```

## Code Patterns and Conventions

### Go Code Style
- Use table-driven testing for all tests
- Prefer early returns to avoid deep nesting
- Use switch statements over if-else chains for multiple conditions
- Always use lowercase log messages
- Code to interfaces for easier mocking in tests
- Each test case should have isolated mock dependencies
- Always ensure a new line at the end of files
- Always use conventional commit messages for clarity (example: `feat: Add new agent type`, `fix: Resolve validation issue`)

### Template Development
- Templates use Go's text/template with Sprig functions
- Template files are embedded as Go constants in `internal/templates/`
- Context provides ADL data and generation metadata
- Generated files get automatic headers with generation info

### Error Handling
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Validate ADL structure early in generation process
- Provide clear error messages for validation failures

## Testing Strategy

### Required Testing
- Always run `task lint` before committing
- Always run `task test` to ensure all tests pass
- Test both success and error cases
- Use isolated mocks for each test case

### Test Organization
- Generator tests in `internal/generator/generator_test.go`
- Schema validation tests in `internal/schema/validator_test.go`
- Example ADL files serve as integration tests

## Language Support Architecture

### Current Support
- **Go**: Full support with unified template

### Language Detection
The generator detects target language from ADL `spec.language` section and validates exactly one language is specified.

### Adding New Languages
1. Update `internal/schema/types.go` with new language config struct
2. Add templates in `internal/templates/[language].go`
3. Update `internal/templates/engine.go` to handle new language templates
4. Add validation logic in `internal/generator/generator.go`
5. Create example ADL files in `examples/`

## Common Development Tasks

### Running CLI During Development
```bash
# After building
./bin/adl generate --file examples/go-agent.yaml --output ./test-output

# Using task dev
task dev -- generate --file examples/go-agent.yaml --output ./test-output
```

### Testing Generation
```bash
# Generate from example
./bin/adl generate --file examples/rust-agent.yaml --output ./test-output

# Validate ADL file
./bin/adl validate examples/rust-agent.yaml
```

### Debugging Generation Issues
- Check ADL validation errors first (detailed in `internal/generator/generator.go:validateADL`)
- Template execution errors include context about which template failed
- Generated files include headers showing generation metadata
- Use `.adl-ignore` to protect files during iterative development

## CLI Command Structure

- `adl init [name]` - Interactive project initialization
- `adl generate` - Generate project from ADL file (main command)
- `adl validate [file]` - Validate ADL file against schema

### Generate Command Options
- `--file` - ADL file path (required)
- `--output` - Output directory (required)  
- `--overwrite` - Overwrite existing files (respects .adl-ignore)
