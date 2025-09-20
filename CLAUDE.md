# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
- `task build` - Build the ADL CLI binary to `bin/adl`
- `task test` - Run all tests
- `task test:coverage` - Run tests with coverage report
- `task lint` - Run golangci-lint (must be installed)
- `task fmt` - Format all Go code
- `task vet` - Run go vet for static analysis
- `task mod` - Download Go modules and tidy go.mod
- `task ci` - Run complete CI pipeline: fmt, lint, test, build

### Testing
- `go test -v ./...` - Run all tests with verbose output
- `go test -v ./cmd -run TestInit` - Run specific test
- `go test -v ./internal/generator` - Test specific package
- `task examples:test` - Validate all example ADL files
- `task examples:generate` - Generate projects from all examples

### Development Workflow
- `task dev -- init my-agent` - Run ADL CLI in development mode
- `task dev -- generate --file agent.yaml --output ./test` - Generate project
- `task dev -- validate examples/go-agent.yaml` - Validate ADL file

## Architecture

### Service Injection

The ADL CLI supports service injection to improve testability and maintainability of skills:

#### Configuration
Services are defined at the spec level and injected into specific skills. The `logger` service is built-in and doesn't need to be declared in the services list:

```yaml
spec:
  services:
    database:  # Custom services
      type: service
      interface: DatabaseService
      factory: NewDatabaseService
      description: Database service for data access
  skills:
    - id: query_database
      name: query_database
      inject:
        - logger    # Built-in, always available
        - database  # Must be declared in services
```

#### Implementation
- **Built-in Logger**: `logger` service is automatically available as `*zap.Logger` without declaration
- **Custom Services**: User-defined services create packages in `internal/` (e.g., `internal/database/`)
- **Constructor Functions**: Skills receive injected services as constructor parameters
- **Interface-Based**: Custom services use interfaces for better testability
- **Configuration Package**: Application configuration is centralized in `config/config.go`
- **Validation**: Build-time validation ensures injected services are defined in spec

#### Benefits
- Improved testability through service mocking
- Better separation of concerns
- Centralized service and configuration management
- Type-safe service contracts
- Simplified logging with direct zap.Logger integration

### Core Components

The ADL CLI follows a command-based architecture using Cobra framework:

```
main.go                       # Entry point, sets version
├── cmd/                      # CLI commands
│   ├── root.go              # Root command setup
│   ├── init.go              # Interactive ADL manifest creation
│   ├── generate.go          # Project generation from ADL
│   └── validate.go          # ADL schema validation
└── internal/                 # Core business logic
    ├── generator/            # Code generation engine
    │   ├── generator.go      # Main generation logic
    │   └── ignore.go         # .adl-ignore handling
    ├── schema/               # ADL schema definitions
    │   ├── types.go          # ADL type definitions
    │   └── validator.go      # Schema validation
    ├── prompt/               # Interactive prompt utilities
    │   └── prompt.go         # User input handling
    └── templates/            # Template system
        ├── engine.go         # Template rendering engine
        ├── registry.go       # Template registration
        ├── headers.go        # File header generation
        ├── common/           # Universal templates
        ├── languages/        # Language-specific templates
        │   ├── go/          # Go project templates
        │   ├── rust/        # Rust project templates
        │   └── typescript/  # TypeScript templates (planned)
        └── sandbox/          # Development environment templates
            ├── flox/        # Flox environment configs
            └── devcontainer/ # DevContainer configs
```

### Template System

The template system uses Go's `text/template` with Sprig functions:

1. **Template Registry** (`internal/templates/registry.go`):
   - Maps ADL configurations to template files
   - Handles language detection and file mapping
   - Supports conditional generation based on flags

2. **Template Engine** (`internal/templates/engine.go`):
   - Renders templates with ADL context
   - Manages file headers with metadata
   - Handles post-generation hooks

3. **Template Context**:
   - ADL configuration (complete spec)
   - Generation metadata (timestamp, version)
   - Language-specific helpers

### Generation Flow

1. **Validation Phase**:
   - Parse and validate ADL YAML against schema
   - Check for required fields and constraints
   - Validate skill schemas and configurations
   - Verify service injection references (built-in logger support)

2. **Template Selection**:
   - Detect target language from ADL spec
   - Build file mapping for selected language including config package
   - Include conditional files (CI, deployment, etc.)

3. **Generation Phase**:
   - Create output directory structure
   - Render templates with ADL context and service information
   - Handle .adl-ignore for file protection
   - Execute post-generation hooks

4. **Post-Processing**:
   - Run language-specific formatters
   - Execute custom hooks from ADL
   - Generate .adl-ignore for skill files

### Key Interfaces

- `schema.ADL` - Root ADL configuration structure
- `generator.Generator` - Main generation interface
- `templates.Engine` - Template rendering interface
- `prompt.Prompter` - Interactive input interface

## Testing Strategy

### Unit Tests
- Table-driven tests for all packages
- Mock interfaces for external services
- Isolated test cases with dedicated mocks
- Coverage target: >80%

### Integration Tests
- End-to-end generation tests
- ADL validation against examples
- Template rendering verification
- CI/CD generation validation

### Test Patterns
```go
// Table-driven test example
func TestGenerateProject(t *testing.T) {
    tests := []struct {
        name    string
        adl     *schema.ADL
        flags   GenerateFlags
        wantErr bool
    }{
        // Test cases here
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Code Style Guidelines

### Go Conventions
- Use early returns to reduce nesting
- Prefer switch over if-else chains
- Table-driven tests for comprehensive coverage
- Lowercase log messages for consistency
- Code to interfaces for testability
- Type safety over dynamic typing

### Error Handling
- Wrap errors with context using `fmt.Errorf`
- Return errors early
- Log errors at appropriate levels
- Provide actionable error messages

### Testing
- Each test case with isolated services
- Mock external services and file systems
- Use test fixtures in `testdata/` directories
- Clean up test artifacts

## Common Tasks

### Adding a New Language

1. Create language directory: `internal/templates/languages/<lang>/`
2. Add language detection in `internal/generator/generator.go`
3. Create file mapping in `internal/templates/registry.go`
4. Add language-specific templates
5. Update ADL schema for language config
6. Add example ADL file
7. Write integration tests

### Adding a New Command

1. Create command file in `cmd/<command>.go`
2. Register with root command in `cmd/root.go`
3. Add command logic and flags
4. Write unit tests in `cmd/<command>_test.go`
5. Update README documentation

### Modifying Templates

1. Locate template in `internal/templates/`
2. Update template syntax (Go text/template)
3. Test with example ADL files
4. Verify generated output
5. Update integration tests if needed

## CI/CD Pipeline

### GitHub Actions Workflows

- **ci.yml**: Runs on PR and main branch
  - Go 1.25 setup with caching
  - golangci-lint v2.4.0
  - Module tidying and dirty check
  - Full test suite

- **release.yml**: Manual dispatch for releases
  - Semantic versioning with conventional commits
  - GitHub App authentication for releases
  - Multi-platform binary builds via goreleaser
  - Docker image publishing

### Release Process

1. Trigger release workflow manually
2. Semantic-release analyzes commits
3. Version bump based on commit types
4. Build binaries for multiple platforms
5. Create GitHub release with artifacts
6. Update install script with new version

## Debugging Tips

- Use `task dev` for quick iteration
- Add debug logging with `log.Printf`
- Test templates with minimal ADL files
- Use `--overwrite` flag for regeneration
- Check `.adl-ignore` for protected files
- Validate ADL with `adl validate` first

## Important Notes

- Go 1.25 required (uses new Go features)
- Templates use Sprig v3 functions
- ADL schema version: `adl.dev/v1`
- Supports Go, Rust, TypeScript (planned)
- CI/CD generation for GitHub Actions
- Deployment support for Kubernetes and CloudRun