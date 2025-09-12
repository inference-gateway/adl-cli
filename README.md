<div align="center">

# ADL CLI

*A command-line interface for generating production-ready A2A (Agent-to-Agent) servers from Agent Definition Language (ADL) files.*

> ⚠️ **Early Development Warning**: This project is in its early stages of development. Breaking changes are expected and acceptable until we reach a stable version. Use with caution in production environments.

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg?style=flat-square)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/inference-gateway/adl-cli/ci.yml?style=flat-square&logo=github)](https://github.com/inference-gateway/adl-cli/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/inference-gateway/adl-cli?style=flat-square)](https://goreportcard.com/report/github.com/inference-gateway/adl-cli)
[![Release](https://img.shields.io/github/v/release/inference-gateway/adl-cli?style=flat-square&logo=github)](https://github.com/inference-gateway/adl-cli/releases)

</div>

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
  - [Quick Install (Recommended)](#quick-install-recommended)
  - [From Source](#from-source)
  - [Using Go Install](#using-go-install)
  - [Pre-built Binaries](#pre-built-binaries)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [Commands](#commands)
  - [Init Command](#init-command)
  - [Generate Command](#generate-command)
- [Agent Definition Language (ADL)](#agent-definition-language-adl)
- [Generated Project Structure](#generated-project-structure)
- [Sandbox Environments](#sandbox-environments)
- [Enterprise Features](#enterprise-features)
- [GitHub Issue Templates](#github-issue-templates)
- [Examples](#examples)
- [Template System & Architecture](#template-system--architecture)
- [Customizing Generation with .adl-ignore](#customizing-generation-with-adl-ignore)
- [Configurable Acronyms](#configurable-acronyms)
- [Post-Generation Hooks](#post-generation-hooks)
- [Development](#development)
- [Roadmap](#roadmap)
- [License](#license)
- [Support](#support)

## Overview

The ADL CLI helps you build production-ready A2A agents quickly by generating complete project scaffolding from YAML-based Agent Definition Language (ADL) files. It eliminates boilerplate code and ensures consistent patterns across your agent implementations.

### Key Features

- 🚀 **Rapid Development** - Generate complete projects in seconds
- 📋 **Schema-Driven** - Use YAML Agent Definition Language files (ADL) to define your agents
- 🎯 **Production Ready** - Single unified template with AI integration and enterprise features
- 🔐 **Enterprise Features** - Authentication, SCM integration, and audit logging
- 🛠️ **Smart Ignore** - Protect your implementations with .adl-ignore files
- ✅ **Validation** - Built-in ADL schema validation
- 🛠️ **Interactive Setup** - Guided project initialization with extensive CLI options
- 🔧 **CI/CD Generation** - Automatic GitHub Actions workflows with semantic-release CD pipelines
- 🏗️ **Sandbox Environments** - Flox and DevContainer support for isolated development
- 🎣 **Post-Generation Hooks** - Customize build, format, and test commands after generation
- 🤖 **Multi-Provider AI** - OpenAI, Anthropic, DeepSeek, Ollama, Google, Mistral, and Groq support

## Installation

### Quick Install (Recommended)

Use our install script to automatically download and install the latest binary:

```bash
curl -fsSL https://raw.githubusercontent.com/inference-gateway/adl-cli/main/install.sh | bash
```

Or download and run the script manually:

```bash
wget https://raw.githubusercontent.com/inference-gateway/adl-cli/main/install.sh
chmod +x install.sh
./install.sh
```

**Install Options:**

- Install specific version: `./install.sh --version v1.0.0`
- Custom install directory: `INSTALL_DIR=~/bin ./install.sh`
- Show help: `./install.sh --help`


### From Source

```bash
git clone https://github.com/inference-gateway/adl-cli.git
cd adl-cli
go install .
```

### Using Go Install

```bash
go install github.com/inference-gateway/adl-cli@latest
```

### Pre-built Binaries

Download pre-built binaries from the [releases page](https://github.com/inference-gateway/adl-cli/releases).

## Quick Start

### 1. Initialize a New Project

```bash
# Interactive project setup - creates ADL manifest
adl init my-weather-agent

# Generate project code from the manifest
adl generate --file agent.yaml --output ./test-my-agent
```

### 2. Implement Your Business Logic

The generated project includes TODO placeholders for your implementations:

```go
// TODO: Implement weather API logic
func GetWeatherTool(ctx context.Context, args map[string]any) (string, error) {
    city := args["city"].(string)
    // TODO: Replace with actual weather API call
    return fmt.Sprintf(`{"city": "%s", "temp": "22°C"}`, city), nil
}
```

### 3. Build and Run

```bash
cd test-weather-agent
task build
task run
```

## Usage

### Commands

| Command | Description |
|---------|-------------|
| `adl init [name]` | Create ADL manifest file interactively with options |
| `adl generate` | Generate project code from ADL file with CI/CD and sandbox support |
| `adl validate [file]` | Validate an ADL file against the complete schema |

### Init Command

The `adl init` command provides a interactive wizard for creating ADL manifest files:

```bash
# Interactive ADL manifest creation
adl init my-weather-agent

# Use defaults for all prompts
adl init my-agent --defaults

# Non-interactive with specific configuration
adl init my-agent \
  --name "Weather Agent" \
  --description "Provides weather information" \
  --provider openai \
  --model gpt-4o-mini \
  --language go \
  --flox
```

#### Init Command Options

The init command supports extensive configuration options:

**Project Settings:**
- `--defaults` - Use default values for all prompts
- `--path` - Project directory path
- `--name` - Agent name
- `--description` - Agent description  
- `--version` - Agent version

**Agent Configuration:**
- `--type` - Agent type (`ai-powered`/`minimal`)
- `--provider` - AI provider (`openai`/`anthropic`/`deepseek`/`ollama`/`google`/`mistral`/`groq`)
- `--model` - AI model name
- `--system-prompt` - System prompt for the agent
- `--max-tokens` - Maximum tokens (integer)
- `--temperature` - Temperature (0.0-2.0)

**Capabilities:**
- `--streaming` - Enable streaming responses
- `--notifications` - Enable push notifications
- `--history` - Enable state transition history

**Server Configuration:**
- `--port` - Server port (integer)
- `--debug` - Enable debug mode

**Language-Specific Options:**
- `--language` - Programming language (`go`/`rust`, TypeScript support planned)

**Go Options:**
- `--go-module` - Go module path (e.g., `github.com/user/project`)
- `--go-version` - Go version (e.g., `1.24`)

**Rust Options:**
- `--rust-package-name` - Rust package name
- `--rust-version` - Rust version (e.g., `1.88`)  
- `--rust-edition` - Rust edition (e.g., `2024`)

**TypeScript Options:**
- `--typescript-name` - TypeScript package name

**Environment Options:**
- `--flox` - Enable Flox environment
- `--devcontainer` - Enable DevContainer environment

### Generate Command

```bash
# Generate project from ADL file
adl generate --file agent.yaml --output ./test-my-agent

# Overwrite existing files (respects .adl-ignore)
adl generate --file agent.yaml --output ./test-my-agent --overwrite

# Generate with CI workflow configuration
adl generate --file agent.yaml --output ./test-my-agent --ci

# Generate with AI assistant instructions and claude-code integration
adl generate --file agent.yaml --output ./test-my-agent --ai

# Generate with CloudRun deployment configuration
adl generate --file agent.yaml --output ./test-my-agent --deployment cloudrun

# Generate with CloudRun deployment and CD pipeline
adl generate --file agent.yaml --output ./test-my-agent --deployment cloudrun --cd
```

#### Generate Flags

| Flag | Description |
|------|-------------|
| `--file`, `-f` | ADL file to generate from (default: "agent.yaml") |
| `--output`, `-o` | Output directory for generated code (default: ".") |
| `--template`, `-t` | Template to use (default: "minimal") |
| `--overwrite` | Overwrite existing files (respects .adl-ignore) |
| `--ci` | Generate CI workflow configuration (GitHub Actions) |
| `--cd` | Generate CD pipeline configuration with semantic-release |
| `--deployment` | Generate deployment configuration (`kubernetes`, `cloudrun`) |
| `--ai` | Generate AI assistant instructions (CLAUDE.md) and add claude-code to sandbox environments |

**CI Generation Features:**
- **Automatic Provider Detection**: Detects GitHub from ADL `spec.scm.provider` (GitLab support planned)
- **Language-Specific Workflows**: Tailored CI configurations for Go, Rust, and TypeScript
- **Version Integration**: Uses language versions from ADL configuration
- **Task Integration**: Leverages generated Taskfile for consistent build processes
- **Caching**: Includes dependency caching for faster builds

**CD Generation Features:**
- **Semantic Release Integration**: Automatic versioning based on conventional commits
- **Multi-Language Support**: Builds and tests for Go, Rust, and TypeScript projects
- **Container Publishing**: Builds and pushes Docker images to GitHub Container Registry
- **Manual Dispatch**: CD workflow triggered manually via GitHub Actions
- **Changelog Generation**: Automatic CHANGELOG.md generation with release notes
- **GitHub Releases**: Creates GitHub releases with appropriate tagging
- **Deployment Integration**: Supports automatic deployment to Kubernetes and Cloud Run after successful releases

**AI Integration Features:**
The `--ai` flag enables enhanced development experience with AI assistant capabilities:

- **CLAUDE.md Generation**: Creates AI assistant instructions tailored to your agent
  - Project-specific guidelines based on your ADL configuration
  - Language-specific development patterns and best practices
  - Skills implementation guidance with TODO placeholders context
  - Testing strategies and development workflow recommendations
- **Claude Code Integration**: Automatically adds claude-code to sandbox environments
  - DevContainer integration for seamless AI-assisted development
  - Flox environment integration with claude-code tooling
  - Improved development experience with AI pair programming capabilities

**Deployment Generation Features:**
The `--deployment` flag generates platform-specific deployment configurations:

- **CloudRun Deployment**: Creates a `deploy` task in the root `Taskfile.yml` for gcloud deployment
  - Supports both Google Container Registry (GCR) and GitHub Container Registry (GHCR)
  - Configurable resources (CPU, memory), scaling (min/max instances), and service options
  - Uses direct gcloud commands for truly serverless deployment (no Kubernetes required)
  - Automatic container building with Docker or Cloud Build integration
- **Kubernetes Deployment**: Creates `k8s/deployment.yaml` with standard Kubernetes manifests
  - Production-ready configurations with resource limits and health checks
  - ConfigMap and Secret integration for environment variables
  - Service and Ingress configurations for load balancing

## Agent Definition Language (ADL)

ADL files use YAML to define your agent's configuration, capabilities, and tools.

### Example ADL File

```yaml
apiVersion: adl.dev/v1
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
    provider: ""  # Choose: openai, anthropic, deepseek, ollama, google, mistral, groq
    model: ""     # Specify default model name for chosen provider
    systemPrompt: "You are a helpful weather assistant."
    maxTokens: 4096
    temperature: 0.7
  skills:
    - name: get_weather
      description: "Get current weather for a city"
      schema:
        type: object
        properties:
          city:
            type: string
            description: "City name"
          country:
            type: string
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
    acronyms: ["api", "json", "xml"] # Optional: Custom acronyms for better code generation
```

### ADL Schema

The complete ADL schema includes:

- **metadata**: Agent name, description, and version
- **capabilities**: Streaming, notifications, state history
- **agent**: AI provider configuration (OpenAI, Anthropic, DeepSeek, Ollama, Google, Mistral, Groq)
- **skills**: Function definitions with complex JSON schemas and validation
- **server**: HTTP server configuration with authentication support
- **language**: Programming language-specific settings (Go, Rust, TypeScript) and configurable acronyms
- **scm**: Source control management configuration (GitHub, GitLab) 
- **sandbox**: Development environment configuration (Flox, DevContainer)
- **deployment**: Platform-specific deployment configuration (Kubernetes, Cloud Run)

### Complete ADL Example

```yaml
apiVersion: adl.dev/v1
kind: Agent
metadata:
  name: advanced-agent
  description: "Enterprise agent with full feature set"
  version: "1.0.0"
spec:
  capabilities:
    streaming: true
    pushNotifications: true
    stateTransitionHistory: true
  agent:
    provider: openai
    model: gpt-4o-mini
    systemPrompt: |
      You are a helpful assistant with enterprise capabilities.
      Always prioritize security and compliance.
    maxTokens: 8192
    temperature: 0.3
  skills:
    - name: query_database
      description: "Execute database queries with validation"
      schema:
        type: object
        properties:
          query:
            type: string
            description: "SQL query to execute"
          table:
            type: string
            description: "Target table name"
          limit:
            type: integer
            description: "Result limit"
            maximum: 1000
        required: [query, table]
    - name: send_notification
      description: "Send multi-channel notifications"
      schema:
        type: object
        properties:
          recipient:
            type: string
            description: "Recipient identifier"
          message:
            type: string
            description: "Message content"
          priority:
            type: string
            enum: ["low", "medium", "high", "critical"]
          channel:
            type: string
            enum: ["email", "slack", "teams", "webhook"]
        required: [recipient, message, priority, channel]
  server:
    port: 8443
    debug: false
    auth:
      enabled: true
  language:
    go:
      module: "github.com/company/advanced-agent"
      version: "1.24"
  scm:
    provider: github
    url: "https://github.com/company/advanced-agent"
  deployment:
    type: cloudrun
    cloudrun:
      image:
        registry: gcr.io
        repository: advanced-agent
        tag: latest
        useCloudBuild: true
      resources:
        cpu: "2"
        memory: 1Gi
      scaling:
        minInstances: 1
        maxInstances: 100
        concurrency: 1000
      service:
        timeout: 3600
        allowUnauthenticated: false
        serviceAccount: advanced-agent@PROJECT_ID.iam.gserviceaccount.com
        executionEnvironment: gen2
      environment:
        LOG_LEVEL: info
        ENVIRONMENT: production
  sandbox:
    flox:
      enabled: true
```

## Generated Project Structure

The ADL CLI generates project scaffolding tailored to your chosen language:

### Go Project Structure
```
my-go-agent/
├── main.go                    # Main server setup
├── go.mod                     # Go module definition
├── skills/                    # Skill implementations directory
│   ├── query_database.go      # Individual skill files (TODO placeholders)
│   └── send_notification.go
├── Taskfile.yml               # Development tasks (build, test, lint)
├── Dockerfile                 # Container configuration
├── .adl-ignore                # Files to protect from regeneration
├── .well-known/
│   └── agent.json             # Agent capabilities (auto-generated)
├── .github/                   # GitHub-specific configurations
│   ├── workflows/             # Generated when using --ci flag
│   │   ├── ci.yml             # GitHub Actions CI workflow
│   │   └── cd.yml             # GitHub Actions CD workflow (with --cd flag)
│   └── ISSUE_TEMPLATE/        # Generated when issue_templates: true
│       ├── bug_report.md      # Bug report template
│       ├── feature_request.md # Feature request template
│       └── refactor_request.md # Refactoring request template
├── .releaserc.yaml            # Semantic-release configuration (with --cd flag)
├── k8s/
│   └── deployment.yaml        # Kubernetes deployment manifest
├── cloudrun/
│   └── deploy.sh              # CloudRun deployment script (with --deployment cloudrun)
├── .flox/                     # Generated when sandbox: flox
│   ├── env/manifest.toml
│   ├── env.json
│   ├── .gitignore
│   └── .gitattributes
├── .gitignore                 # Standard Git ignore patterns
├── .gitattributes             # Git attributes configuration
├── .editorconfig              # Editor configuration
├── CLAUDE.md                  # AI assistant instructions (generated with --ai flag)
└── README.md                  # Project documentation with setup instructions
```

### Rust Project Structure
```
my-rust-agent/
├── src/
│   ├── main.rs                # Main application entry point
│   └── skills/                # Skill implementations directory
│       ├── mod.rs             # Module declarations
│       ├── query_database.rs  # Individual skill implementations
│       └── send_notification.rs
├── Cargo.toml                 # Rust package configuration
├── Taskfile.yml               # Development tasks
├── Dockerfile                 # Rust-optimized container
├── .adl-ignore                # Protection configuration
├── .well-known/
│   └── agent.json             # Agent capabilities
├── .github/workflows/         # CI configuration (with --ci)
│   ├── ci.yml                 # Rust-specific CI workflow
│   └── cd.yml                 # GitHub Actions CD workflow (with --cd flag)
├── .releaserc.yaml            # Semantic-release configuration (with --cd flag)
├── k8s/
│   └── deployment.yaml        # Kubernetes deployment
├── cloudrun/
│   └── deploy.sh              # CloudRun deployment script (with --deployment cloudrun)
├── CLAUDE.md                  # AI assistant instructions (generated with --ai flag)
└── README.md                  # Documentation
```

### Universal Generated Files

All projects include these essential files regardless of language:

- **`.well-known/agent.json`** - A2A agent discovery and capabilities manifest
- **`Taskfile.yml`** - Unified task runner configuration for build, test, lint, run
- **`Dockerfile`** - Language-optimized container configuration  
- **`k8s/deployment.yaml`** - Kubernetes deployment manifest
- **`deploy` task in `Taskfile.yml`** - CloudRun deployment task (when using `--deployment cloudrun`)
- **`.adl-ignore`** - Protects user implementations from overwrite
- **CI Workflows** - When using `--ci` flag, generates GitHub Actions workflows:
  - **GitHub Actions**: `.github/workflows/ci.yml`
  - **GitLab CI**: `.gitlab-ci.yml` (planned, not yet implemented)
- **CD Workflows** - When using `--cd` flag, generates continuous deployment:
  - **GitHub Actions**: `.github/workflows/cd.yml`
  - **Semantic Release**: `.releaserc.yaml`
- **Development Environment** - Based on `sandbox` configuration:
  - **Flox**: `.flox/` directory with environment configuration when `sandbox.flox.enabled: true`
  - **DevContainer**: `.devcontainer/devcontainer.json` when `sandbox.devcontainer.enabled: true`
- **AI Assistant Instructions** - When using `--ai` flag:
  - **CLAUDE.md**: AI assistant instructions tailored to your agent configuration

### CI Integration

When using the `--ci` flag, the ADL CLI generates GitHub Actions workflows for your project:

```bash
# Generate project with CI workflow
adl generate --file agent.yaml --output ./test-my-agent --ci
```

This creates a GitHub Actions workflow (`.github/workflows/ci.yml`) that includes:

- **Automated Testing**: Runs all tests on every push and pull request
- **Code Quality**: Format checking and linting
- **Multi-Environment**: Supports main and develop branches
- **Caching**: Go module caching for faster builds
- **Task Integration**: Uses the generated Taskfile for consistent build steps

The generated workflow automatically detects your Go version from the ADL file and configures the appropriate environment.

### CD Integration

The ADL CLI can generate continuous deployment (CD) pipelines with semantic release automation:

```bash
# Generate project with CD pipeline
adl generate --file agent.yaml --output ./test-my-agent --cd
```

This creates a complete CD setup including:

- **`.releaserc.yaml`** - Semantic-release configuration with conventional commits
- **`.github/workflows/cd.yml`** - GitHub Actions CD workflow with manual dispatch

The generated CD pipeline includes:

- **Semantic Versioning**: Automatic version bumping based on conventional commit messages
- **Release Automation**: Creates GitHub releases with generated release notes
- **Container Publishing**: Builds and publishes Docker images to GitHub Container Registry
- **Multi-Platform Builds**: Supports both AMD64 and ARM64 architectures
- **Language Detection**: Automatically configures build steps based on your project language
- **Change Detection**: Only publishes releases when there are changes to release

#### CD Workflow Features

**Manual Trigger**: The CD workflow uses `workflow_dispatch` for controlled releases:
```bash
# Trigger via GitHub CLI
gh workflow run cd.yml

# Or trigger via GitHub Actions UI
```

**Conventional Commits Support**: The pipeline recognizes these commit types for versioning:
- `feat:` - Minor version bump (new features)
- `fix:` - Patch version bump (bug fixes)
- `refactor:`, `perf:`, `ci:`, `docs:`, `style:`, `test:`, `build:`, `chore:` - Patch version bump

**Container Registry**: Published images are available at:
```
ghcr.io/your-org/your-agent:latest
ghcr.io/your-org/your-agent:v1.0.0
ghcr.io/your-org/your-agent:1.0
```

## CloudRun Deployment

The ADL CLI provides native support for deploying A2A agents to Google Cloud Run, offering a truly serverless deployment experience without Kubernetes complexity.

### CloudRun Configuration

Configure CloudRun deployment in your ADL file:

```yaml
spec:
  deployment:
    type: cloudrun
    cloudrun:
      image:
        registry: gcr.io        # gcr.io or ghcr.io
        repository: my-agent    # Repository name
        tag: latest            # Image tag
        useCloudBuild: true    # Use Cloud Build or local Docker
      resources:
        cpu: "2"               # CPU allocation (0.1 to 8)
        memory: 1Gi           # Memory limit (128Mi to 32Gi)
      scaling:
        minInstances: 0        # Minimum instances (0 to 1000)
        maxInstances: 100     # Maximum instances (1 to 1000)
        concurrency: 1000     # Max concurrent requests per instance
      service:
        timeout: 3600         # Request timeout in seconds
        allowUnauthenticated: true    # Allow public access
        serviceAccount: my-agent@PROJECT_ID.iam.gserviceaccount.com
        executionEnvironment: gen2    # gen1 or gen2
      environment:            # Custom environment variables
        LOG_LEVEL: info
        ENVIRONMENT: production
```

### Container Registry Options

**Google Container Registry (GCR):**
```yaml
image:
  registry: gcr.io
  repository: my-project/my-agent
  useCloudBuild: true        # Automatically build and push
```

**GitHub Container Registry (GHCR):**
```yaml
image:
  registry: ghcr.io
  repository: myorg/my-agent
  useCloudBuild: false       # Skip Cloud Build, use pre-built image
```

### Generated Deployment Script

When using `--deployment cloudrun`, the ADL CLI generates a `deploy` task in the `Taskfile.yml` that:

- **Validates Environment**: Checks for required `PROJECT_ID` and `REGION` variables
- **Container Building**: Uses Docker locally or Cloud Build based on configuration
- **Direct gcloud Deployment**: Uses `gcloud run deploy` for serverless deployment
- **Configuration Summary**: Displays all deployment settings for verification

### CloudRun Deployment Workflow

```bash
# 1. Generate project with CloudRun deployment
adl generate --file agent.yaml --output ./my-agent --deployment cloudrun

# 2. Set required environment variables
export PROJECT_ID="my-gcp-project"
export REGION="us-central1"

# 3. Deploy to CloudRun
cd my-agent
task deploy
```

### CloudRun with CI/CD

Generate CloudRun deployment with continuous deployment:

```bash
adl generate --file agent.yaml --deployment cloudrun --cd
```

This creates:
- **CD Workflow**: Automatically deploys to CloudRun after releases
- **Environment Integration**: Uses GitHub secrets for GCP authentication
- **Multi-Environment Support**: Deploy to different regions/projects

**Required GitHub Secrets:**
- `GCP_SA_KEY`: Service account key JSON
- `GCP_PROJECT_ID`: Google Cloud project ID
- `GCP_REGION`: Deployment region (e.g., us-central1)

### CloudRun Benefits

- **Truly Serverless**: No Kubernetes clusters or infrastructure management
- **Auto-Scaling**: Scale to zero when idle, scale up automatically under load
- **Pay-per-Use**: Only pay for actual request processing time
- **Global Edge**: Deploy to multiple regions with traffic management
- **Integrated Monitoring**: Built-in logging, metrics, and tracing
- **Custom Domains**: HTTPS support with automatic SSL certificates

### Example ADL Files

The CLI includes CloudRun example files:

```bash
# Validate CloudRun examples
adl validate examples/cloudrun-agent.yaml
adl validate examples/cloudrun-ghcr-agent.yaml

# Generate CloudRun projects
adl generate --file examples/cloudrun-agent.yaml --output ./cloudrun-test
adl generate --file examples/cloudrun-ghcr-agent.yaml --output ./ghcr-test
```

## Sandbox Environments

The ADL CLI supports multiple development environments for isolated, reproducible development:

### Flox Environment

Configure Flox for your project by adding to your ADL file:

```yaml
spec:
  sandbox:
    flox:
      enabled: true
```

Generated files:
- `.flox/env/manifest.toml` - Flox environment manifest with language-specific dependencies
- `.flox/env.json` - Environment configuration
- `.flox/.gitignore` - Flox-specific ignore patterns  
- `.flox/.gitattributes` - Git attributes for Flox files

### DevContainer Environment  

Configure DevContainer for your project:

```yaml
spec:
  sandbox:
    devcontainer:
      enabled: true
```

Generated files:
- `.devcontainer/devcontainer.json` - VS Code DevContainer configuration with language support

### Multiple Environment Support

You can enable multiple sandbox environments simultaneously:

```yaml
spec:
  sandbox:
    flox:
      enabled: true
    devcontainer:
      enabled: true
```

This generates both Flox and DevContainer configurations, allowing developers to choose their preferred environment.

### Benefits of Sandbox Environments

- **Reproducible Development** - Consistent environments across team members
- **Isolated Dependencies** - No conflicts with system-wide installations
- **Language-Specific Tooling** - Pre-configured with appropriate development tools
- **CI/CD Integration** - Matches production environment characteristics

## Enterprise Features

### Authentication Configuration

Enable server authentication in your ADL file:

```yaml
spec:
  server:
    port: 8443
    debug: false  
    auth:
      enabled: true
```

This generates enterprise-ready authentication scaffolding in your project.

### SCM Integration

Configure source control management for automatic CI/CD provider detection:

```yaml
spec:
  scm:
    provider: github  # gitlab support planned
    url: "https://github.com/company/my-agent"
    github_app: false  # optional: enable GitHub App for CD
    issue_templates: true  # optional: generate GitHub issue templates
```

**Features:**
- **Automatic CI Detection** - Generates appropriate workflows based on SCM provider
- **Repository Integration** - Links generated projects to source control
- **Workflow Optimization** - SCM-specific optimizations and best practices
- **GitHub App Support** - Enhanced security for enterprise CD pipelines
- **Issue Templates** - Generate GitHub issue templates for standardized bug reports and feature requests

#### GitHub App Integration

For enterprise environments, you can enable GitHub App-based CD deployment for enhanced security:

```yaml
spec:
  scm:
    provider: github
    url: "https://github.com/company/my-agent"
    github_app: true
```

**GitHub App CD Benefits:**
- **Enhanced Security** - App tokens are automatically revoked after pipeline execution
- **Enterprise Compliance** - Keeps main branch protected from direct pushes
- **Bot Identity** - Release operations performed by dedicated bot account
- **Audit Trail** - Clear attribution of automated actions

**Required GitHub Secrets:**
- `BOT_GH_APP_ID` - Your GitHub App ID
- `BOT_GH_APP_PRIVATE_KEY` - Your GitHub App private key

When `github_app: true` is set, the generated CD pipeline will use GitHub App authentication instead of the default `GITHUB_TOKEN`, providing better security isolation for release management.

### AI Provider Support

The ADL CLI supports multiple AI providers including OpenAI, Anthropic, DeepSeek, Ollama (for local LLMs), Google AI, Mistral, and Groq. Each provider requires appropriate API keys to be configured as environment variables. See the ADL examples above for configuration details.

## GitHub Issue Templates

The ADL CLI can automatically generate GitHub issue templates for your agent projects, providing standardized forms for bug reports, feature requests, and refactoring tasks:

```yaml
spec:
  scm:
    provider: github
    url: "https://github.com/company/my-agent"
    issue_templates: true  # Enable issue template generation
```

When `issue_templates: true` is set, the following templates are generated in `.github/ISSUE_TEMPLATE/`:

- **`bug_report.md`** - Structured bug reporting with severity levels, reproduction steps, and environment details
- **`feature_request.md`** - Feature proposals with use case descriptions and acceptance criteria
- **`refactor_request.md`** - Code improvement requests with motivation and impact analysis

**Issue Template Features:**
- **Agent Context** - Templates include agent name and version from your ADL metadata
- **Structured Sections** - Consistent formatting for better issue triage and tracking
- **GitHub Integration** - Automatic labels and assignees configured in frontmatter
- **Severity Levels** - Priority classification for bug reports (critical, high, medium, low)
- **Environment Info** - Sections for capturing logs, system details, and configurations

## Examples

The CLI includes example ADL files in the `examples/` directory:

```bash
# Validate examples
adl validate examples/go-agent.yaml
adl validate examples/rust-agent.yaml  
adl validate examples/github-app-agent.yaml
adl validate examples/cloudrun-agent.yaml
adl validate examples/cloudrun-ghcr-agent.yaml

# Generate from examples
adl generate --file examples/go-agent.yaml --output ./test-go-agent
adl generate --file examples/rust-agent.yaml --output ./test-rust-agent
adl generate --file examples/github-app-agent.yaml --output ./test-github-app-agent --cd
adl generate --file examples/cloudrun-agent.yaml --output ./test-cloudrun-agent --deployment cloudrun
adl generate --file examples/cloudrun-ghcr-agent.yaml --output ./test-ghcr-agent --deployment cloudrun

# Generate with CI/CD pipeline
adl generate --file examples/github-app-agent.yaml --output ./enterprise-agent --ci --cd
adl generate --file examples/cloudrun-agent.yaml --output ./cloudrun-enterprise --deployment cloudrun --cd
```

**Example ADL Files:**
- `go-agent.yaml` - Basic Go agent with multiple skills and capabilities
- `rust-agent.yaml` - Rust agent with enterprise features
- `github-app-agent.yaml` - Enterprise agent with GitHub App CD integration
- `cloudrun-agent.yaml` - CloudRun deployment with Google Container Registry
- `cloudrun-ghcr-agent.yaml` - CloudRun deployment with GitHub Container Registry

## Template System & Architecture

The ADL CLI uses a sophisticated template system that generates language-specific projects:

### Language Detection

The generator automatically detects your target language from the ADL file:

```go
// Automatic detection based on spec.language configuration
func DetectLanguageFromADL(adl *schema.ADL) string {
    if adl.Spec.Language.Go != nil     { return "go" }
    if adl.Spec.Language.Rust != nil   { return "rust" }  
    if adl.Spec.Language.TypeScript != nil { return "typescript" }
    return "go" // default
}
```

### File Mapping System

Each language has its own file mapping that determines what gets generated:

**Go Projects:**
- `main.go` → Go main server setup
- `skills/{skillname}.go` → Individual skill implementations  
- `go.mod` → Go module configuration
- Language-specific Dockerfile and CI configurations

**Rust Projects:**  
- `src/main.rs` → Rust main application
- `src/skills/{skillname}.rs` → Skill implementations
- `src/skills/mod.rs` → Module declarations
- `Cargo.toml` → Rust package configuration

**Universal Files:**
- `Taskfile.yml` → Development task runner
- `.well-known/agent.json` → A2A capabilities manifest
- `k8s/deployment.yaml` → Kubernetes deployment
- CI workflows and sandbox configurations

### Template Context

All templates receive a rich context object:

```go
type Context struct {
    ADL      *schema.ADL           // Complete ADL configuration
    Metadata GeneratedMetadata     // Generation metadata
    Language string               // Detected language
}
```

This allows templates to access any ADL configuration and generate language-appropriate code.

## Customizing Generation with .adl-ignore

The ADL CLI automatically creates a `.adl-ignore` file during project generation to protect files containing TODO implementations. This file works similar to `.gitignore` and prevents important implementation files from being overwritten during subsequent generations.

### Automatically Protected Files

When you generate a project, implementation files are automatically added to `.adl-ignore` to protect your business logic from being overwritten during regeneration.

You can control which additional files are generated or updated by editing the `.adl-ignore` file:

```bash
# .adl-ignore
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

### .adl-ignore Patterns

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

- Go 1.24+
- [Task](https://taskfile.dev/) (optional, for using Taskfile commands)

### Building from Source

```bash
git clone https://github.com/inference-gateway/adl-cli.git
cd adl-cli

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

## Configurable Acronyms

The ADL CLI includes support for configurable acronyms to improve code generation readability. This feature helps generate more readable function and struct names by properly capitalizing acronyms in generated code.

### How It Works

Define custom acronyms in your ADL file's `spec.language.acronyms` field. These acronyms will be properly capitalized when generating identifiers in your code.

### Configuration

```yaml
spec:
  language:
    go:
      module: "github.com/company/my-agent"
      version: "1.24"
    acronyms: ["n8n", "xml", "mqtt", "iot", "uuid"]
```

### Generated Code Examples

**Without custom acronyms:**
- `get_n8n_docs` → `GetN8nDocsSkill` 
- `process_xml_data` → `ProcessXmlDataSkill`

**With custom acronyms:**
- `get_n8n_docs` → `GetN8NDocsSkill`
- `process_xml_data` → `ProcessXMLDataSkill`

### Default Acronyms

The following acronyms are recognized by default:
- **Common**: id, api, url, uri, json, xml, sql, html, css, js, ui, uuid
- **Network**: http, https, tcp, udp, ip, dns, tls, ssl  
- **Tech**: cpu, gpu, ram, io, os, db

Your custom acronyms extend these defaults and take precedence over them.

## Post-Generation Hooks

The ADL CLI supports custom post-generation hooks that run automatically after project generation. These hooks allow you to execute commands like formatting, linting, testing, or custom setup scripts.

### Default Hooks

Each language has sensible defaults:

**Go Projects:**
- `go fmt ./...` - Format all Go source files
- `go mod tidy` - Download dependencies and clean up go.mod

**Rust Projects:**
- `cargo fmt` - Format all Rust source files  
- `cargo check` - Check the project for errors

### Custom Hooks

You can customize or extend the default behavior by adding a `hooks` section to your ADL file:

```yaml
apiVersion: adl.dev/v1
kind: Agent
metadata:
  name: my-agent
spec:
  # ... other configuration ...
  
  # Custom post-generation hooks
  hooks:
    post:
      - "go fmt ./..."
      - "go mod tidy"
      - "go vet ./..."
      - "go test -short ./..."
      - "golangci-lint run --fix"
```

### Hooks Behavior

- **Override Defaults**: When you specify custom hooks, they completely replace the language defaults
- **Command Execution**: Commands run in the generated project directory
- **Error Handling**: Failed commands show warnings but don't stop generation
- **Sequential Execution**: Commands run in the order specified
- **Shell Support**: Commands are executed through the system shell

### Example Configurations

**Extended Go Development:**
```yaml
hooks:
  post:
    - "go mod download"             # Download dependencies first
    - "go generate ./..."           # Generate code if needed
    - "gofumpt -l -w ."             # Improved formatting
    - "golangci-lint run --fix"     # Lint and auto-fix
    - "go test -race -short ./..."  # Run tests
    - "go build -v ./..."           # Verify build works
```

**Rust with Additional Tools:**
```yaml
hooks:
  post:
    - "cargo fmt"
    - "cargo clippy --fix --allow-dirty"
    - "cargo check --all-targets"
    - "cargo test --lib"
```

**TypeScript/Node.js:**
```yaml
hooks:
  post:
    - "npm install"
    - "npm run format"
    - "npm run lint:fix"
    - "npm run type-check"
    - "npm test"
```

### Best Practices

- **Keep hooks fast** - Avoid long-running commands that slow down generation
- **Use error-tolerant commands** - Commands should gracefully handle missing tools
- **Order matters** - Place dependencies first (e.g., `npm install` before `npm run lint`)
- **Document requirements** - Note any required tools in your project README

## Roadmap

### Language Support

The ADL CLI currently supports Go and Rust, with plans to expand to additional programming languages:

#### ✅ Currently Supported
- **Go** - Full support with templates for main.go, go.mod, and tools
- **Rust** - Full support with templates for main.rs, Cargo.toml, and tools

#### 🚧 Planned Support
- **TypeScript/Node.js** - Template structure exists but templates not yet implemented
  - Complete A2A agent generation with Express.js framework planned
  - AI-powered agents with OpenAI/Anthropic integration
  - Enterprise features (auth, metrics, logging)
  - Docker and Kubernetes deployment configs

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
- 💡 Suggest new languages or frameworks via [Issues](https://github.com/inference-gateway/adl-cli/issues)
- 🤝 Contribute implementations for new languages (see [Contributing Guide](CONTRIBUTING.md))

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://docs.inference-gateway.com)
- 💬 [Discussions](https://github.com/inference-gateway/adl-cli/discussions)
- 🐛 [Issues](https://github.com/inference-gateway/adl-cli/issues)

---

> 🤖 Powered by the [Inference Gateway ecosystem](https://github.com/inference-gateway/)
