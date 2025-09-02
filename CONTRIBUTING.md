# Contributing to ADL CLI

Thank you for your interest in contributing to the ADL CLI! This guide will help you get started with contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Submitting Changes](#submitting-changes)
- [Adding Language Support](#adding-language-support)
- [Testing](#testing)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- [Task](https://taskfile.dev/) (recommended for development tasks)
- Docker (for testing container builds)

### Development Setup

1. **Fork the repository**
   ```bash
   # Fork the repo on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/adl-cli.git
   cd adl-cli
   ```

2. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/inference-gateway/adl-cli.git
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Verify setup**
   ```bash
   # Run tests
   task test
   
   # Run linting
   task lint
   
   # Build the CLI
   task build
   ```

## Contributing Guidelines

### Types of Contributions

We welcome several types of contributions:

- **🐛 Bug fixes** - Fix issues in existing functionality
- **✨ New features** - Add new capabilities to the CLI
- **🌍 Language support** - Add support for new programming languages
- **📝 Documentation** - Improve docs, examples, or help content
- **🧪 Tests** - Add or improve test coverage
- **🔧 Tooling** - Improve development tools and processes

### Before You Start

1. **Check existing issues** - Look for existing issues or discussions about your idea
2. **Create an issue** - For new features or significant changes, create an issue first to discuss the approach
3. **Small changes** - For small bug fixes or documentation improvements, you can go directly to creating a PR

### Coding Standards

#### Go Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and reasonably sized
- Use early returns to reduce nesting

#### Code Organization

- Place new language templates in `internal/templates/`
- Add new commands in `cmd/`
- Use the existing error handling patterns
- Follow the current project structure

#### Commit Messages

Use clear, descriptive commit messages:

```
feat: add TypeScript template support

- Add TypeScript template with Express.js framework
- Include enterprise features (auth, metrics, logging)
- Add corresponding tests and documentation

Fixes #123
```

Format: `type: description`

Types:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Adding or updating tests
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

## Submitting Changes

### Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write code following our standards
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**
   ```bash
   # Run all tests
   task test
   
   # Run linting
   task lint
   
   # Test with examples
   task examples:test
   ```

4. **Commit and push**
   ```bash
   git add .
   git commit -m "feat: your descriptive commit message"
   git push origin feature/your-feature-name
   ```

5. **Create a Pull Request**
   - Use a clear title and description
   - Reference any related issues
   - Include screenshots for UI changes
   - Add reviewers if you know who should review

### Pull Request Requirements

- [ ] Code follows project conventions
- [ ] Tests pass (`task test`)
- [ ] Linting passes (`task lint`)
- [ ] Documentation updated (if applicable)
- [ ] Examples work (if adding new functionality)
- [ ] Commit messages are clear and descriptive

## Adding Language Support

Adding support for a new programming language is a significant contribution. Here's how to approach it:

### 1. Planning

- Create an issue to discuss the language addition
- Define the scope and features for the language template
- Identify the web framework and dependencies to use

### 2. Implementation Steps

#### A. Schema Updates

Update `internal/schema/types.go` to add language configuration:

```go
type LanguageSpec struct {
    Go         *GoConfig         `yaml:"go,omitempty"`
    TypeScript *TypeScriptConfig `yaml:"typescript,omitempty"`
    Rust       *RustConfig       `yaml:"rust,omitempty"` // New language
    // ... other languages
}

type RustConfig struct {
    Edition     string `yaml:"edition"`     // e.g., "2021"
    Framework   string `yaml:"framework"`   // e.g., "axum"
    AsyncRuntime string `yaml:"runtime"`    // e.g., "tokio"
}
```

#### B. Template Creation

Create templates in `internal/templates/`:

```go
// internal/templates/rust.go
package templates

func getRustTemplate() map[string]string {
    return map[string]string{
        "Cargo.toml":     rustCargoTemplate,
        "src/main.rs":    rustMainTemplate,
        "src/handlers.rs": rustHandlersTemplate,
        // ... other files
    }
}

const rustMainTemplate = `// Rust main template
use axum::{routing::post, Router};
// ... template content
`
```

#### C. Engine Updates

Update `internal/templates/engine.go` to handle the new language:

```go
func (e *Engine) GetFiles() map[string]string {
    switch e.templateName {
    case "rust":
        return getRustTemplate()
    // ... existing cases
    }
}
```

#### D. Generator Updates

Update `internal/generator/generator.go` for language detection and validation.

### 3. Testing

- Add tests for the new language
- Create example ADL files in `examples/`
- Test the template thoroughly
- Verify generated code compiles and runs

### 4. Documentation

- Update README.md roadmap
- Add language-specific documentation
- Include example ADL files
- Document any language-specific features

## Testing

### Running Tests

```bash
# Run all tests
task test

# Run tests with coverage
task test:coverage

# Run specific tests
go test ./internal/generator -v

# Test examples
task examples:test
```

### Writing Tests

- Use table-driven tests when appropriate
- Test both success and error cases
- Mock external dependencies
- Include integration tests for end-to-end flows

Example test structure:

```go
func TestGenerator_GenerateRust(t *testing.T) {
    tests := []struct {
        name        string
        adlFile     string
        wantFiles   []string
        wantErr     bool
    }{
        {
            name:      "rust template",
            adlFile:   "testdata/rust-agent.yaml",
            wantFiles: []string{"Cargo.toml", "src/main.rs"},
            wantErr:   false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Documentation

### Types of Documentation

1. **Code Comments** - Document exported functions and complex logic
2. **README.md** - Keep the main README updated
3. **Examples** - Add practical examples for new features
4. **ADL Documentation** - Document new ADL schema fields

### Documentation Standards

- Use clear, concise language
- Include code examples
- Keep examples up to date
- Test documentation examples

## Community

### Getting Help

- **Discussions**: Use [GitHub Discussions](https://github.com/inference-gateway/adl-cli/discussions) for questions
- **Issues**: Use [GitHub Issues](https://github.com/inference-gateway/adl-cli/issues) for bugs and feature requests

### Reviewing Process

1. **Maintainer Review** - A project maintainer will review your PR
2. **Community Feedback** - Other contributors may provide feedback
3. **CI Checks** - Automated tests must pass
4. **Approval** - At least one maintainer approval required for merge

### Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes for significant contributions
- Project documentation

## Release Process

The ADL CLI follows semantic versioning:

- **Patch** (1.0.1) - Bug fixes and small improvements
- **Minor** (1.1.0) - New features, backward compatible
- **Major** (2.0.0) - Breaking changes

Releases are managed by maintainers, but contributors can suggest features for upcoming releases.

## Questions?

If you have questions about contributing:

1. Check this guide and existing documentation
2. Search existing issues and discussions
3. Create a new discussion or issue
4. Reach out to maintainers

Thank you for contributing to ADL CLI! 🚀

---

> This guide is inspired by best practices from the Go community and other successful open-source projects.
