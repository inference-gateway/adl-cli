<div align="center">

# ADL CLI

_A command-line interface for generating enterprise-ready A2A (Agent-to-Agent) servers from Agent Definition Language (ADL) files._

> ⚠️ **Early Development Warning**: This project is in its early stages of development. Breaking changes are expected and acceptable until we reach a stable version. Use with caution in production environments.

[![Go Version](https://img.shields.io/github/go-mod/go-version/inference-gateway/adl-cli?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg?style=flat-square)](LICENSE)
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
- [Artifacts Support](#artifacts-support)
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

The ADL CLI helps you build enterprise-ready A2A agents quickly by generating complete project scaffolding from YAML-based Agent Definition Language (ADL) files. It eliminates boilerplate code and ensures consistent patterns across your agent implementations.

### Key Features

- 🚀 **Rapid Development** - Generate complete projects in seconds
- 📋 **Schema-Driven** - Use YAML Agent Definition Language files (ADL) to define your agents
- 🎯 **Enterprise Ready** - Single unified template with AI integration and enterprise features
- 🔐 **Enterprise Features** - Authentication, SCM integration, and audit logging
- 🛠️ **Smart Ignore** - Protect your implementations with .adl-ignore files
- ✅ **Validation** - Built-in ADL schema validation
- 🛠️ **Interactive Setup** - Guided project initialization with extensive CLI options
- 🔗 **Structured Services** - Type-safe dependency injection with interfaces and factory functions
- ⚙️ **Configuration Management** - Automatic environment variable mapping with proper naming conventions
- 🔧 **CI/CD Generation** - Automatic GitHub Actions workflows with semantic-release CD pipelines
- 🏗️ **Sandbox Environments** - Flox and DevContainer support for isolated development
- 🎣 **Post-Generation Hooks** - Customize build, format, and test commands after generation
- 🤖 **Multi-Provider AI** - OpenAI, Anthropic, DeepSeek, Ollama, Google, Mistral, and Groq support
- 📁 **Artifacts Support** - Integrated filesystem and MinIO object storage for artifact management

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

### Nix Flake

Run the latest version directly without installing:

```bash
nix run github:inference-gateway/adl-cli
```

Or pin a specific version:

```bash
nix run github:inference-gateway/adl-cli/v0.27.13
```

Build and add it to your profile:

```bash
nix profile install github:inference-gateway/adl-cli/v0.27.13
```

Enter a development shell with `go`, `go-task`, `golangci-lint`, `gopls`, and
`goreleaser` available:

```bash
nix develop github:inference-gateway/adl-cli
```

### Flox

Pin `adl` to a specific version inside a [Flox](https://flox.dev) environment by
adding it to your `.flox/env/manifest.toml`:

```toml
[install]
adl.flake = "github:inference-gateway/adl-cli/v0.31.0"
```

Then activate the environment:

```bash
flox activate
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

| Command               | Description                                                        |
| --------------------- | ------------------------------------------------------------------ |
| `adl init [name]`     | Create ADL manifest file interactively with options                |
| `adl generate`        | Generate project code from ADL file with CI/CD and sandbox support |
| `adl validate [file]` | Validate an ADL file against the complete schema                   |

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
  --provider deepseek \
  --model deepseek-v4-flash \
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
- `--go-version` - Go version (e.g., `1.26.2`)

**Rust Options:**

- `--rust-package-name` - Rust package name
- `--rust-version` - Rust version (e.g., `1.94`)
- `--rust-edition` - Rust edition (e.g., `2024`)

**TypeScript Options:**

- `--typescript-name` - TypeScript package name

**Environment Options:**

- `--flox` - Enable Flox environment
- `--devcontainer` - Enable DevContainer environment

**Pipeline / AI Options (declarative, written into the manifest as `false` by default):**

- `--ai` - Shortcut for the init wizard: writes
  `spec.development.ai.claudecode.enabled: true` into the generated `agent.yaml`.
  Every other per-agent toggle (`codex`, `gemini`, `opencode`, `infer`) stays
  off; edit `agent.yaml` after init to enable additional agents
  (see [Per-agent AI assistants](#per-agent-ai-assistants)).
- `--ci` - Sets `spec.scm.ci: true` (generate CI workflow on `adl generate`)
- `--cd` - Sets `spec.scm.cd: true` (generate CD pipeline + semantic-release on `adl generate`)

### Generate Command

```bash
# Generate project from ADL file
adl generate --file agent.yaml --output ./test-my-agent

# Overwrite existing files (respects .adl-ignore)
adl generate --file agent.yaml --output ./test-my-agent --overwrite

# Generate with CI workflow configuration
adl generate --file agent.yaml --output ./test-my-agent --ci

# Generate with CloudRun deployment configuration
adl generate --file agent.yaml --output ./test-my-agent --deployment cloudrun

# Generate with CloudRun deployment and CD pipeline
adl generate --file agent.yaml --output ./test-my-agent --deployment cloudrun --cd
```

#### Generate Flags

| Flag               | Description                                                                                |
| ------------------ | ------------------------------------------------------------------------------------------ |
| `--file`, `-f`     | ADL file to generate from (default: "agent.yaml")                                          |
| `--output`, `-o`   | Output directory for generated code (default: ".")                                         |
| `--template`, `-t` | Template to use (default: "minimal")                                                       |
| `--overwrite`      | Overwrite existing files (respects .adl-ignore)                                            |
| `--ci`             | Generate CI workflow configuration (GitHub Actions). Overrides `spec.scm.ci`.              |
| `--cd`             | Generate CD pipeline configuration with semantic-release. Overrides `spec.scm.cd`.         |
| `--deployment`     | Generate deployment configuration (`kubernetes`, `cloudrun`)                               |

> **Declarative equivalents:** `--ci` and `--cd` are mirrored by `spec.scm.ci`
> and `spec.scm.cd`. The CLI flag is OR'd on top of the manifest value (passing
> the flag wins; omitting it falls back to the manifest). AI assistants are
> entirely manifest-driven via the per-agent toggles in `spec.development.ai`
> - see the matrix below.
> `adl init` writes all toggles as `false` by default - they're opt-in. Generated files
> (`CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `.github/workflows/ci.yml`,
> `.github/workflows/cd.yml`, `.github/workflows/claude.yml`,
> `.github/workflows/codex.yml`, `.github/workflows/gemini.yml`,
> `.releaserc.yaml`) are tagged `linguist-generated=true` in `.gitattributes`
> so they collapse in pull request diffs.

**CI Generation Features:**

- **Automatic Provider Detection**: Detects GitHub from ADL `spec.scm.provider` (GitLab support planned)
- **Language-Specific Workflows**: Tailored CI configurations for Go, Rust, and TypeScript
- **Version Integration**: Uses language versions from ADL configuration
- **Task Integration**: Leverages generated Taskfile for consistent build processes
- **Caching**: Includes service caching for faster builds

**CD Generation Features:**

- **Semantic Release Integration**: Automatic versioning based on conventional commits
- **Multi-Language Support**: Builds and tests for Go, Rust, and TypeScript projects
- **Container Publishing**: Builds and pushes Docker images to GitHub Container Registry
- **Manual Dispatch**: CD workflow triggered manually via GitHub Actions
- **Changelog Generation**: Automatic CHANGELOG.md generation with release notes
- **GitHub Releases**: Creates GitHub releases with appropriate tagging
- **Deployment Integration**: Supports automatic deployment to Kubernetes and Cloud Run after successful releases

**AI Integration Features:**

The ADL CLI honours the per-agent toggles in `spec.development.ai` (introduced
in ADL schema v0.8.0). Each entry is independent and defaults to `false`:

```yaml
spec:
  development:
    ai:
      claudecode:
        enabled: true   # generates CLAUDE.md + .github/workflows/claude.yml
      codex:
        enabled: false  # would generate AGENTS.md + .github/workflows/codex.yml
      gemini:
        enabled: false  # would generate GEMINI.md + .github/workflows/gemini.yml
      opencode:
        enabled: false  # would generate AGENTS.md (no upstream action yet)
      infer:
        enabled: false  # would generate AGENTS.md (no upstream action yet)
```

#### Per-agent AI assistants

| Agent toggle  | Docs file the agent reads | GitHub Actions workflow generated? |
|---------------|---------------------------|------------------------------------|
| `claudecode`  | `CLAUDE.md`               | yes (`.github/workflows/claude.yml`, uses `anthropics/claude-code-action`) |
| `codex`       | `AGENTS.md` (shared)      | yes (`.github/workflows/codex.yml`, uses `openai/codex-action`) |
| `gemini`      | `GEMINI.md`               | yes (`.github/workflows/gemini.yml`, uses `google-github-actions/run-gemini-cli`) |
| `opencode`    | `AGENTS.md` (shared)      | no upstream action yet - docs only |
| `infer`       | `AGENTS.md` (shared)      | no workflow scaffolded yet - docs only |

- `AGENTS.md` is generated **once** and is shared by every enabled agent that
  reads from it (`codex`, `opencode`, `infer`); the file's contents are
  agent-agnostic.
- `CLAUDE.md` and `GEMINI.md` are agent-specific and only appear when the
  matching toggle is on.
- If no toggles are enabled, no AI docs or workflows are emitted.
- Pre-v0.8.0 manifests using `spec.development.ai.enabled: true` are no longer
  accepted - `adl validate` and `adl generate` will fail with a migration hint
  pointing at the per-agent toggles. Move `enabled: true` to the specific agent
  you want (e.g. `claudecode.enabled: true`).
- When `claudecode` is enabled, sandbox environments (Flox, DevContainer)
  also gain the `claude-code` CLI / extension automatically.

**Deployment Generation Features:**
The `--deployment` flag generates platform-specific deployment configurations:

- **CloudRun Deployment**: Creates a `deploy` task in the root `Taskfile.yml` for gcloud deployment
  - Supports both Google Container Registry (GCR) and GitHub Container Registry (GHCR)
  - Configurable resources (CPU, memory), scaling (min/max instances), and service options
  - Uses direct gcloud commands for truly serverless deployment (no Kubernetes required)
  - Automatic container building with Docker or Cloud Build integration
- **Kubernetes Deployment**: Creates `k8s/deployment.yaml` with standard Kubernetes manifests
  - Enterprise-ready configurations with resource limits and health checks
  - ConfigMap and Secret integration for environment variables
  - Service and Ingress configurations for load balancing

## Agent Definition Language (ADL)

ADL files use YAML to define your agent's configuration, capabilities, and tools.

The canonical schema lives in the [inference-gateway/adl](https://github.com/inference-gateway/adl) repository - that repo is the single source of truth for the ADL specification. This CLI vendors a pinned copy at `internal/schema/schema.json` (refresh with `task fetch-schema`).

### Example ADL File

```yaml
apiVersion: adl.inference-gateway.com/v1
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
    provider: "" # Choose: openai, anthropic, deepseek, ollama, google, mistral, groq
    model: "" # Specify default model name for chosen provider
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
      version: "1.26.2"
  acronyms: # Optional: Custom acronyms for better code generation
    - api
    - json
    - xml
```

### ADL Schema

The complete ADL schema includes:

- **metadata**: Agent name, description, and version
- **capabilities**: Streaming, notifications, state history
- **config**: Structured configuration sections with environment variable mapping
- **services**: Service services with interfaces, factories, and type definitions
- **agent**: AI provider configuration (OpenAI, Anthropic, DeepSeek, Ollama, Google, Mistral, Groq)
- **tools**: Function-call definitions with JSON schemas, validation, and service injection support
- **skills**: Markdown playbooks (id + optional `bare`, version, source) pulled from the skills registry, fetched as a full directory from a GitHub repo (shorthand or URL), or scaffolded locally; advertised on the agent card and prepended to the system prompt at runtime
- **server**: HTTP server configuration with authentication support
- **language**: Programming language-specific settings (Go, Rust, TypeScript) and configurable acronyms
- **scm**: Source control management configuration (GitHub, GitLab)
- **sandbox**: Development environment configuration (Flox, DevContainer)
- **deployment**: Platform-specific deployment configuration (Kubernetes, Cloud Run)

### Complete ADL Example

```yaml
apiVersion: adl.inference-gateway.com/v1
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
    provider: deepseek
    model: deepseek-v4-flash
    systemPrompt: |
      You are a helpful assistant with enterprise capabilities.
      Always prioritize security and compliance.
    maxTokens: 8192
    temperature: 0.3
  config:
    database:
      connectionString: "postgresql://user:pass@localhost:5432/db"
      maxConnections: "10"
      timeout: "30s"
    notifications:
      slackWebhook: "https://hooks.slack.com/services/..."
      emailApiKey: "your-email-api-key"
      retryAttempts: "3"
  services:
    database:
      type: service
      interface: DatabaseService
      factory: NewDatabaseService
      description: PostgreSQL database service for persistent storage
    notifications:
      type: service
      interface: NotificationService
      factory: NewNotificationService
      description: Multi-channel notification service
  tools:
    - name: query_database
      description: "Execute database queries with validation"
      inject:
        - logger
        - database
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
        required:
          - query
          - table
    - name: send_notification
      description: "Send multi-channel notifications"
      inject:
        - logger
        - notifications
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
            enum:
              - low
              - medium
              - high
              - critical
          channel:
            type: string
            enum:
              - email
              - slack
              - teams
              - webhook
        required:
          - recipient
          - message
          - priority
          - channel
  server:
    port: 8443
    debug: false
    auth:
      enabled: true
  language:
    go:
      module: "github.com/company/advanced-agent"
      version: "1.26.2"
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
  development:
    sandbox:
      flox:
        enabled: true
```

### Extra dependencies (`spec.language.<lang>.vendor`)

Every language config block accepts an optional `vendor` section that lets
the manifest extend the generator's built-in dependency set. Use `deps`
for runtime/production dependencies and `devdeps` for development-only
ones. The exact meaning of `devdeps` depends on the language - see the
mapping table below. Each entry must be `<package>@<version>` using
the target language's native package and version syntax - the schema
validates the shape up front (`^\S+@\S+$`) and points at the offending
key if you mistype it (e.g. `spec.language.go.vendor.deps.0`).

**Conflict policy:** generator built-ins always win. If your manifest
lists a package that the generator already pins (e.g.
`github.com/inference-gateway/adk` for Go, `tokio` for Rust), the
vendor entry is silently dropped and a `⚠️  vendor … collides with
built-in …` warning is printed to stderr. This prevents accidental
downgrades of the core runtime SDK.

**Output mapping per language:**

| Language   | `deps` lands in               | `devdeps` lands in                 |
| ---------- | ----------------------------- | ---------------------------------- |
| Go         | `go.mod` `require` block      | `go.mod` [`tool` directive](https://go.dev/doc/modules/managing-dependencies#tools) (executable dev tools: code generators, linters, etc.) plus an `// indirect` entry in `require` so the module is downloadable. Test libraries that you `import` (testify, go-cmp, …) belong in `deps`, not `devdeps`. |
| Rust       | `Cargo.toml` `[dependencies]` | `Cargo.toml` `[dev-dependencies]`  |
| TypeScript | `package.json` `dependencies` | `package.json` `devDependencies` _(plumbed end-to-end once the TypeScript generator templates land - the schema and validator already accept the field)_ |

For Go, supply the **full tool package path** (the binary's `main` package,
e.g. `golang.org/x/tools/cmd/stringer`) with a version. After generation,
run `go mod tidy` so Go normalises the indirect `require` entry to the
actual module root.

Examples per language:

```yaml
# Go: uuid for runtime, stringer + mockgen as dev tools.
spec:
  language:
    go:
      module: github.com/example/agent
      version: "1.26.2"
      vendor:
        deps:
          - github.com/google/uuid@v1.6.0
          - github.com/stretchr/testify@v1.10.0  # imported by *_test.go
        devdeps:
          - golang.org/x/tools/cmd/stringer@v0.20.0
          - github.com/golang/mock/mockgen@v1.6.0
```

```yaml
# Rust: regex at runtime, mockall + pretty_assertions for tests.
spec:
  language:
    rust:
      packageName: agent
      version: "1.94.1"
      edition: "2024"
      vendor:
        deps:
          - regex@1.10.0
        devdeps:
          - mockall@0.12.1
          - pretty_assertions@1.4.0
```

```yaml
# TypeScript: axios at runtime, vitest + @types/node for tests.
spec:
  language:
    typescript:
      packageName: "@example/agent"
      nodeVersion: "20"
      vendor:
        deps:
          - axios@1.7.0
        devdeps:
          - "@types/node@20.11.0"
          - vitest@1.6.0
```

### Extra sandbox dependencies (`spec.development.deps`)

`spec.development.deps` is the cross-cutting equivalent of the per-language
`vendor.deps` block above: it lets you install tools into the development
sandbox (Flox, devcontainer) that don't belong to any single language's
package manager. Use it for things like `deno`, `kubectl`, `terraform`,
`awscli`, or any other CLI you want available inside the dev shell. Each
entry follows the `<package>@<version>` shape, validated up front by the
schema (`^\S+@\S+$`).

```yaml
spec:
  development:
    sandbox:
      flox:
        enabled: true
      devcontainer:
        enabled: true
    deps:
      - deno@2.1.4
      - kubectl@1.31.0
      - terraform@1.9.5
```

**Merge semantics:** additive. The per-language toolchain that each
sandbox template already installs (`go`/`cargo`/`nodejs`, plus `git`,
`docker`, `go-task`) is always emitted; `spec.development.deps` entries
are appended on top, sorted alphabetically so the generated file diffs
stay stable on re-run. Duplicate package names inside `deps` are deduped
(first occurrence wins on version). When a `deps` entry's package name
collides with one of the template's built-ins (e.g. `git@2.53.0`), the
generator prints a `⚠️  spec.development.deps … collides with a Flox
built-in …` warning to stderr but still renders the user entry - the
maintainer's pin wins on version conflict.

**Output mapping per backend:**

| Backend         | Generated file                  | Per-entry rendering                                                                                                                                                            |
| --------------- | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `flox`          | `.flox/env/manifest.toml`       | A pair of TOML lines under `[install]`: `<pkg>.pkg-path = "<pkg>"` / `<pkg>.version = "<version>"`. Resolved against Nixpkgs at activation time.                               |
| `devcontainer`  | `.devcontainer/devcontainer.json` | Added to the `features` block as `ghcr.io/devcontainers-extra/features/apt-packages:1` with the comma-joined `<pkg>=<version>` list. Resolution is best-effort against apt.    |
| `dockerCompose` | _(out of scope for this release; see issue #154)_ | n/a                                                                                                                                                                            |

If a package isn't published under its short attribute path in Nixpkgs,
or isn't available as an apt package on the devcontainer base image,
add the file to your project's `.adl-ignore` and hand-author the
override - this is the same escape hatch we use for any sandbox file
the template can't infer.

Worked example per backend:

```yaml
# Flox: pin deno + kubectl + terraform alongside the Go toolchain
spec:
  language:
    go: { module: github.com/example/agent, version: "1.26.2" }
  development:
    sandbox:
      flox:
        enabled: true
    deps:
      - deno@2.1.4
      - kubectl@1.31.0
      - terraform@1.9.5
```

```yaml
# Devcontainer: same deps, rendered as an apt-packages feature
spec:
  language:
    go: { module: github.com/example/agent, version: "1.26.2" }
  development:
    sandbox:
      devcontainer:
        enabled: true
    deps:
      - deno@2.1.4
      - kubectl@1.31.0
```

## Skills vs. Tools

The ADL spec distinguishes two complementary concepts:

- **Tools** (`spec.tools`) are function-call entrypoints with explicit JSON schemas. They are generated as code in the target language and registered with the agent's toolbox. The model invokes them by name with structured arguments.
- **Skills** (`spec.skills`) are markdown playbooks (with YAML frontmatter) that describe _when and how_ to use the tools. Each is written to its own directory at `skills/<id>/SKILL.md` in the generated project, advertised on the agent card so orchestrators can discover them, and prepended to the system prompt at runtime. The directory layout matches Anthropic's [agent skills convention](https://github.com/anthropics/skills) - bare skills can ship arbitrary scripts, templates, or reference material alongside `SKILL.md`.

A skill entry is small:

```yaml
spec:
  skills:
    - id: data-analysis          # pulled from registry.inference-gateway.com/skills/
      version: 0.1.0             # optional pin
    - id: report-writing         # pulled at the default version
    - id: company-policy         # scaffolded locally (not fetched)
      bare: true
      name: company-policy
      description: "Internal compliance rules to follow"
      license: Proprietary       # optional SPDX id or "Proprietary"
      tags: [policy, compliance]
    - id: pdf                    # pull a full skill directory from GitHub
      source: anthropics/skills/pdf
    - id: skill-creator
      source: skill-creator@v1.0 # pin to a tag/branch/sha
```

### Resolution rules

- **`bare: true`** → the CLI scaffolds `skills/<id>/SKILL.md` with frontmatter from the manifest and a TODO body that you author by hand. The whole `skills/<id>/` directory is listed in `.adl-ignore`, so any bundled scripts, templates, or resources you drop alongside `SKILL.md` are preserved on regeneration.
- **`source:` set** → the source must resolve to a public GitHub directory (a `/tree/<ref>/<path>` URL, or one of the shorthand forms below). The CLI pulls the _entire_ directory - `SKILL.md`, reference docs, bundled scripts, anything else - and writes it to `skills/<id>/`. Non-`github.com` URLs are rejected so the same code path always produces a complete skill bundle, not a stray markdown file.
- **Otherwise** → fetch `https://registry.inference-gateway.com/skills/<id>[/<version>].md` (becomes `skills/<id>/SKILL.md`). Override the registry with `ADL_SKILLS_REGISTRY`. Registry-by-id currently ships `SKILL.md` only; if you need bundled assets, use `source:` to point at a GitHub directory.

### Licensing

`license` is optional on every skill entry. When set, it must be one of the
SPDX identifiers enumerated in the schema (`MIT`, `Apache-2.0`, `BSD-2-Clause`,
`BSD-3-Clause`, `GPL-2.0`, `GPL-3.0`, `LGPL-2.1`, `LGPL-3.0`, `MPL-2.0`, `ISC`,
`CC0-1.0`, `CC-BY-4.0`, `CC-BY-SA-4.0`, `Unlicense`) or the literal string
`Proprietary` for closed-source skills. The resolver mirrors the value into the
generated `SKILL.md` frontmatter so the licence travels with the playbook -
shipping a separate `LICENSE` file alongside `SKILL.md` is optional. When the
ADL entry and the fetched frontmatter both set `license`, the value in the ADL
manifest wins.

Use `adl generate --offline` to skip network access - every non-bare skill must already be cached at `~/.adl/skills-cache/<id>@<ref>/` (where `<ref>` is the pinned tag/branch, or `latest` for an unpinned registry fetch).

### `source:` shorthand grammar

Every form below resolves to a GitHub `tree/<ref>/<path>` URL. An optional `@<tag>` suffix pins a branch, tag, or commit SHA; omit it to use the default `main` branch.

| Shorthand                        | Expands to                                                                   |
| -------------------------------- | ---------------------------------------------------------------------------- |
| `<skill>`                        | `https://github.com/inference-gateway/skills/tree/main/skills/<skill>`       |
| `<skill>@<tag>`                  | `https://github.com/inference-gateway/skills/tree/<tag>/skills/<skill>`      |
| `<owner>/<repo>/<skill>`         | `https://github.com/<owner>/<repo>/tree/main/skills/<skill>`                 |
| `<owner>/<repo>/<skill>@<tag>`   | `https://github.com/<owner>/<repo>/tree/<tag>/skills/<skill>`                |
| Full `https://github.com/...` URL | passed through unchanged                                                    |

Concrete examples:

```yaml
# Default inference-gateway/skills, latest main:
- id: skill-creator
  source: skill-creator

# Default inference-gateway/skills, pinned to a tag:
- id: skill-creator
  source: skill-creator@v1.0

# Different repo (Anthropic's official skill library):
- id: pdf
  source: anthropics/skills/pdf

# Different repo, pinned to a commit SHA:
- id: pdf
  source: anthropics/skills/pdf@abc1234

# Full URL for anything that doesn't fit the shorthand:
- id: custom
  source: https://github.com/my-org/my-repo/tree/release/path/to/skill
```

The 3-segment form assumes a `skills/<id>/` subdirectory inside the repo (the convention used by both `inference-gateway/skills` and `anthropics/skills`). If your repo lays skills out differently, pass the full URL.

### Runtime: AVAILABLE SKILLS manifest + on-demand Read

The generated agent advertises skills to the LLM via a frontmatter-only manifest, **not** by inlining SKILL.md bodies. At startup it walks first-level subdirectories under `skills/` (overridable with `A2A_SKILLS_DIR`), parses each `<id>/SKILL.md`'s YAML frontmatter, and appends an `AVAILABLE SKILLS:` block to the system prompt:

```text
AVAILABLE SKILLS:
Skills are reusable instructions for specific tasks. When a task matches a
skill's description, read the SKILL.md file at the listed path using the Read
tool, then follow its instructions.

- incident-response: Use this when the user reports a production incident...
  Path: skills/incident-response/SKILL.md
- pdf: Fill in PDF forms and extract structured data from PDFs.
  Path: skills/pdf/SKILL.md
```

The model loads each SKILL.md body on demand via the `Read` built-in tool, and executes any bundled scripts via `Bash` / `Write` / `Edit`. **A skills-using agent must therefore list `- id: read` in `spec.tools` and set `spec.config.tools.read.enabled: true`** - the validator enforces this; see [Reserved built-in tools](#reserved-built-in-tools).

### Reserved built-in tools

`spec.tools` accepts five reserved IDs that map to framework-supplied implementations:

| Reserved ID | Generated as            | Purpose                                                                              |
| ----------- | ----------------------- | ------------------------------------------------------------------------------------ |
| `read`      | `tools/read.go` etc.    | Read a file (`file_path`, optional `offset`/`limit`).                                |
| `bash`      | `tools/bash.go` etc.    | Execute a shell command (subject to whitelist + timeout).                            |
| `write`     | `tools/write.go` etc.   | Write content to a file (creates parent dirs).                                       |
| `edit`      | `tools/edit.go` etc.    | Replace a unique string in a file (`old_string` → `new_string`).                     |
| `fetch`     | `tools/fetch.go` etc.   | Fetch an http(s) URL (whitelist, max-bytes cap, optional save-to-disk inside `/tmp`).|

Opt in by listing the id alone - the generator owns `name`, `description`, and the JSON schema:

```yaml
spec:
  tools:
    - id: read
    - id: bash
    - id: query_database       # user tool: full entry still required
      name: query_database
      description: "..."
      schema: { type: object, ... }
```

**All five built-ins default to `enabled: false`.** Activate them via the reserved namespace `spec.config.tools.<id>`:

```yaml
spec:
  config:
    tools:
      read:
        enabled: true
        max_lines: 2000          # offset/limit default window
        allowed_roots: []        # empty = project-wide
      bash:
        enabled: true
        whitelist: [ls, cat, grep, jq]
        timeout_seconds: 30
      write:
        enabled: false           # listed but explicitly disabled
      edit:
        enabled: true
      fetch:
        enabled: true
        allowed_domains:         # whitelist of hosts (empty = unrestricted, discouraged)
          - example.com
          - .api.dev             # entries starting with "." match any subdomain
        max_bytes: 10485760      # 10 MiB cap on response body (default)
        timeout_seconds: 30      # total request timeout (default)
        download_dir: /tmp       # root for save_path writes (default /tmp)
        allow_downloads: false   # set true to allow writing response bodies to disk
```

Values are baked into the generated constructor as compile-time literals - there's no `ToolsConfig` struct in `config/config.go` because reserved-namespace sections are intentionally skipped. The validator decodes each `spec.config.tools.<id>` block into the built-in's typed shape and rejects unknown keys (typos like `tymeout_seconds` fail with `spec.config.tools.bash.tymeout_seconds`).

Runtime overrides for Bash (read inside `tools/bash.go`):

- `A2A_BASH_DISABLED=1` is a kill switch - overrides `enabled: true` back to false.
- `A2A_BASH_WHITELIST=ls,cat,grep` overrides the compile-time whitelist.

Runtime overrides for Fetch (resolution precedence: env > compile-time literal > default-disabled):

- Go: `TOOLS_FETCH_ENABLED`, `TOOLS_FETCH_ALLOWED_DOMAINS`, `TOOLS_FETCH_MAX_BYTES`, `TOOLS_FETCH_TIMEOUT_SECONDS`, `TOOLS_FETCH_DOWNLOAD_DIR`, `TOOLS_FETCH_ALLOW_DOWNLOADS` (envconfig-style; comma-separated for lists).
- Rust: `A2A_FETCH_DISABLED=1` (kill switch), `A2A_FETCH_ALLOWED_DOMAINS`, `A2A_FETCH_MAX_BYTES`, `A2A_FETCH_TIMEOUT_SECONDS`, `A2A_FETCH_DOWNLOAD_DIR`, `A2A_FETCH_ALLOW_DOWNLOADS`.

The Fetch tool supports `GET` and `HEAD` only. Optional `save_path` writes the response body to a path resolved under `download_dir` - absolute paths and parent-directory traversal (`..`) are rejected, and the request fails unless `allow_downloads: true`. Bodies (and on-disk files) are capped at `max_bytes`; oversized responses are truncated and the result payload sets `"truncated": true`. The Go template uses only the standard library (`net/http`); the Rust template adds `reqwest` (rustls-tls + json features) to `Cargo.toml` automatically when `- id: fetch` is present in `spec.tools`.

Resolution precedence at runtime: **env > compile-time literal > built-in default (disabled)**.

## Service Injection & Configuration Management

The ADL CLI provides a sophisticated service injection system with structured configuration management. This system improves testability, separation of concerns, and provides type-safe configuration with environment variable mapping.

### Structured Service System

Define services with explicit types, interfaces, and factory functions. The system supports both built-in services (like logger) and custom service services:

```yaml
spec:
  config:
    googleCalendar:
      scopes: "https://www.googleapis.com/auth/calendar"
      credentialsPath: "/secrets/credentials.json"
    cache:
      ttl: "3600"
      maxEntries: "1000"
  services:
    googleCalendar:
      type: service
      interface: CalendarService
      factory: NewCalendarService
      description: Google Calendar API service for managing calendar events
    cache:
      type: service
      interface: CacheRepository
      factory: NewCacheRepository
      description: High-performance caching layer for API responses
  tools:
    - name: create_event
      description: "Create a new calendar event"
      inject:
        - logger # Built-in, always available
        - googleCalendar # Custom service
        - cache # Custom service
      schema:
        type: object
        properties:
          title:
            type: string
            description: "Event title"
          start:
            type: string
            description: "Start time (ISO 8601)"
        required: [title, start]
```

### Configuration Management

The configuration system generates type-safe structs with automatic environment variable mapping:

**Generated Configuration (`config/config.go`):**

```go
type Config struct {
    // Core application settings
    Environment string `env:"ENVIRONMENT"`

    // A2A configuration
    A2A serverConfig.Config `env:",prefix=A2A_"`

    // Custom configuration sections
    Cache          CacheConfig          `env:",prefix=CACHE_"`
    GoogleCalendar GoogleCalendarConfig `env:",prefix=GOOGLE_CALENDAR_"`
}

type GoogleCalendarConfig struct {
    CredentialsPath string `env:"CREDENTIALS_PATH"`
    Scopes          string `env:"SCOPES"`
}

type CacheConfig struct {
    MaxEntries string `env:"MAX_ENTRIES"`
    Ttl        string `env:"TTL"`
}
```

**Environment Variables:**

- `GOOGLE_CALENDAR_CREDENTIALS_PATH="/secrets/google-creds.json"`
- `GOOGLE_CALENDAR_SCOPES="https://www.googleapis.com/auth/calendar"`
- `CACHE_MAX_ENTRIES="1000"`
- `CACHE_TTL="3600"`

### Config Subsection Injection

In addition to injecting entire configuration objects, you can inject specific config subsections directly into skills using dotted notation. This provides type-safe access to focused configuration scopes.

**Example ADL Configuration:**

```yaml
spec:
  config:
    database:
      connectionString: "postgresql://localhost:5432/db"
      maxConnections: "10"
      timeout: "30s"
    email:
      apiKey: ""
      fromAddress: "noreply@example.com"
      provider: "sendgrid"
  services:
    database:
      type: service
      interface: DatabaseService
      factory: NewDatabaseService
      description: PostgreSQL database service
  tools:
    - name: export_report
      description: "Export data and email report"
      inject:
        - logger
        - database
        - config.email  # Inject only the email config subsection
      schema:
        type: object
        properties:
          recipient:
            type: string
        required: [recipient]
```

**Generated Skill Code:**

```go
type ExportReportSkill struct {
    logger   *zap.Logger
    database database.DatabaseService
    email    *config.EmailConfig  // Type-safe access to email config only
}

func NewExportReportSkill(
    logger *zap.Logger,
    database database.DatabaseService,
    email *config.EmailConfig,
) server.Tool {
    skill := &ExportReportSkill{
        logger:   logger,
        database: database,
        email:    email,
    }
    // ...
}

func (s *ExportReportSkill) ExportReportHandler(ctx context.Context, args map[string]any) (string, error) {
    // Direct access to email config subsection
    apiKey := s.email.APIKey
    fromAddress := s.email.FromAddress
    provider := s.email.Provider

    // ... implementation
}
```

**Main Registration:**

```go
// In main.go - config subsection is passed directly
exportReportSkill := skills.NewExportReportSkill(l, databaseSvc, &cfg.Email)
toolBox.AddTool(exportReportSkill)
```

**Benefits of Config Subsection Injection:**

- **Scoped Access**: Skills only receive the configuration they need, following principle of least privilege
- **Type Safety**: Compile-time validation ensures config fields exist
- **Clear Dependencies**: Explicit declaration of which config sections each skill requires
- **Easier Testing**: Mock specific config subsections without full config object
- **Better Separation**: Skills don't have access to unrelated configuration
- **Auto-Validation**: ADL CLI validates that injected config sections exist in `spec.config`

**Injection Patterns:**

```yaml
inject:
  - logger                 # Built-in logger service
  - config                 # Entire config object (*config.Config)
  - config.database        # Database config subsection (*config.DatabaseConfig)
  - config.email           # Email config subsection (*config.EmailConfig)
  - myService              # Custom service from spec.services
```

### Service Architecture

The service injection system generates:

1. **Built-in Logger**: Automatically available as `*zap.Logger` without declaration
2. **Type-Safe Configuration**: Structured config with environment variable mapping
3. **Service Interfaces**: Custom service packages with interface definitions
4. **Factory Functions**: Constructor functions that receive logger and configuration
5. **Automatic Registration**: Services are automatically wired into skills
6. **File Protection**: Generated service files are automatically added to `.adl-ignore`

### Generated Structure

```text
my-agent/
├── config/
│   └── config.go                    # Type-safe configuration with env mapping
├── internal/
│   ├── logger/
│   │   └── logger.go               # Built-in logger factory
│   ├── googleCalendar/
│   │   └── googleCalendar.go       # Calendar service with interface
│   └── cache/
│       └── cache.go                # Cache service with interface
├── tools/
│   ├── create_event.go             # Function-call tools with injected services
│   └── list_events.go
├── skills/
│   ├── calendar-workflow/          # Markdown playbooks loaded into the system prompt
│   │   └── SKILL.md
│   └── meeting-summary/
│       └── SKILL.md
└── .adl-ignore                     # Protects custom implementations
```

### Generated Service Code

Each service generates a package with interface and factory:

**Example `internal/googleCalendar/googleCalendar.go`:**

```go
type CalendarService interface {
    // TODO: Define your CalendarService interface methods
    CreateEvent(ctx context.Context, event *Event) error
    ListEvents(ctx context.Context, query *Query) ([]*Event, error)
}

type calendarService struct {
    logger *zap.Logger
    config *config.Config
}

func NewCalendarService(logger *zap.Logger, cfg *config.Config) (CalendarService, error) {
    // TODO: Implement CalendarService initialization
    return &calendarService{
        logger: logger,
        config: cfg,
    }, nil
}
```

### Skill Integration

Skills automatically receive injected services as constructor parameters:

**Example `skills/create_event.go`:**

```go
type CreateEventSkill struct {
    logger    *zap.Logger
    calendar  googleCalendar.CalendarService
    cache     cache.CacheRepository
}

func NewCreateEventSkill(logger *zap.Logger, calendar googleCalendar.CalendarService, cache cache.CacheRepository) *CreateEventSkill {
    return &CreateEventSkill{
        logger:   logger,
        calendar: calendar,
        cache:    cache,
    }
}
```

### Benefits

- **Type Safety**: Structured configuration with compile-time validation
- **Environment Variables**: Automatic mapping with proper naming conventions
- **Interface-Based Design**: Testable services with clear contracts
- **Separation of Concerns**: Configuration separate from service definitions
- **Language Agnostic**: Works across Go, Rust, and planned TypeScript support
- **Hot Reload**: Configuration changes via environment variables
- **Security**: No secrets in code, environment-based configuration
- **Scalability**: Easy to add new services and configuration sections

### Best Practices

1. **Configuration**: Use environment variables for secrets and environment-specific values
2. **Interfaces**: Define clear interfaces for testability and modularity
3. **Factory Functions**: Initialize services with proper error handling
4. **Logging**: Use the injected logger for consistent log formatting
5. **Testing**: Create mock implementations of service interfaces
6. **Documentation**: Document interface methods and configuration options

## Generated Project Structure

The ADL CLI generates project scaffolding tailored to your chosen language:

### Go Project Structure

```text
my-go-agent/
├── main.go                    # Main server setup
├── go.mod                     # Go module definition
├── config/
│   └── config.go              # Centralized application configuration
├── internal/
│   └── logger/
│       └── logger.go          # Built-in logger factory
├── tools/                     # Function-call tool implementations
│   ├── query_database.go      # Individual tool files (TODO placeholders)
│   └── send_notification.go
├── skills/                    # Skill directories (SKILL.md + optional bundled assets)
│   ├── incident-response/     # Loaded into the system prompt at startup
│   │   └── SKILL.md
│   └── support-handoff/
│       └── SKILL.md
├── Taskfile.yml               # Development tasks (build, test, lint)
├── Dockerfile                 # Container configuration
├── .adl-ignore                # Files to protect from regeneration
├── .well-known/
│   └── agent-card.json        # Agent capabilities (auto-generated)
├── .github/                   # GitHub-specific configurations
│   ├── workflows/             # Generated when using --ci flag
│   │   ├── ci.yml             # GitHub Actions CI workflow
│   │   └── cd.yml             # GitHub Actions CD workflow (with --cd flag)
│   ├── dependabot.yml         # Generated when scm.dependabot: true
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
├── CLAUDE.md                  # AI assistant instructions (spec.development.ai.claudecode.enabled: true)
└── README.md                  # Project documentation with setup instructions
```

### Rust Project Structure

```text
my-rust-agent/
├── src/
│   ├── main.rs                # Main application entry point
│   └── tools/                 # Function-call tool implementations
│       ├── mod.rs             # Module declarations
│       ├── query_database.rs  # Individual tool implementations
│       └── send_notification.rs
├── skills/                    # Skill directories (SKILL.md + optional bundled assets)
│   ├── incident-response/
│   │   └── SKILL.md
│   └── support-handoff/
│       └── SKILL.md
├── Cargo.toml                 # Rust package configuration
├── Taskfile.yml               # Development tasks
├── Dockerfile                 # Rust-optimized container
├── .adl-ignore                # Protection configuration
├── .well-known/
│   └── agent-card.json        # Agent capabilities
├── .github/workflows/         # CI configuration (with --ci)
│   ├── ci.yml                 # Rust-specific CI workflow
│   └── cd.yml                 # GitHub Actions CD workflow (with --cd flag)
├── .releaserc.yaml            # Semantic-release configuration (with --cd flag)
├── k8s/
│   └── deployment.yaml        # Kubernetes deployment
├── cloudrun/
│   └── deploy.sh              # CloudRun deployment script (with --deployment cloudrun)
├── CLAUDE.md                  # AI assistant instructions (spec.development.ai.claudecode.enabled: true)
└── README.md                  # Documentation
```

### Universal Generated Files

All projects include these essential files regardless of language:

- **`.well-known/agent-card.json`** - A2A agent discovery and capabilities manifest
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
- **AI Assistant Instructions** - Per-agent toggles under `spec.development.ai`
  (see [Per-agent AI assistants](#per-agent-ai-assistants)):
  - **CLAUDE.md** when `spec.development.ai.claudecode.enabled: true`
  - **GEMINI.md** when `spec.development.ai.gemini.enabled: true`
  - **AGENTS.md** (shared) when any of `codex`, `opencode`, or `infer` is enabled

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

```text
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
        registry: gcr.io # gcr.io or ghcr.io
        repository: my-agent # Repository name
        tag: latest # Image tag
        useCloudBuild: true # Use Cloud Build or local Docker
      resources:
        cpu: "2" # CPU allocation (0.1 to 8)
        memory: 1Gi # Memory limit (128Mi to 32Gi)
      scaling:
        minInstances: 0 # Minimum instances (0 to 1000)
        maxInstances: 100 # Maximum instances (1 to 1000)
        concurrency: 1000 # Max concurrent requests per instance
      service:
        timeout: 3600 # Request timeout in seconds
        allowUnauthenticated: true # Allow public access
        serviceAccount: my-agent@PROJECT_ID.iam.gserviceaccount.com
        executionEnvironment: gen2 # gen1 or gen2
      environment: # Custom environment variables
        LOG_LEVEL: info
        ENVIRONMENT: production
```

### Container Registry Options

**Google Container Registry (GCR):**

```yaml
image:
  registry: gcr.io
  repository: my-project/my-agent
  useCloudBuild: true # Automatically build and push
```

**GitHub Container Registry (GHCR):**

```yaml
image:
  registry: ghcr.io
  repository: myorg/my-agent
  useCloudBuild: false # Skip Cloud Build, use pre-built image
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
  development:
    sandbox:
      flox:
        enabled: true
```

Generated files:

- `.flox/env/manifest.toml` - Flox environment manifest with language-specific services
- `.flox/env.json` - Environment configuration
- `.flox/.gitignore` - Flox-specific ignore patterns
- `.flox/.gitattributes` - Git attributes for Flox files

### DevContainer Environment

Configure DevContainer for your project:

```yaml
spec:
  development:
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
  development:
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
    provider: github # gitlab support planned
    url: "https://github.com/company/my-agent"
    github_app: false # optional: enable GitHub App for CD
    issue_templates: true # optional: generate GitHub issue templates
    dependabot: true # optional: generate Dependabot configuration
```

**Features:**

- **Automatic CI Detection** - Generates appropriate workflows based on SCM provider
- **Repository Integration** - Links generated projects to source control
- **Workflow Optimization** - SCM-specific optimizations and best practices
- **GitHub App Support** - Enhanced security for enterprise CD pipelines
- **Issue Templates** - Generate GitHub issue templates for standardized bug reports and feature requests
- **Dependabot Configuration** - Auto-generate `.github/dependabot.yml` for dependency upgrades

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

## Artifacts Support

Enable artifacts support to allow your agent to create, store, and manage files and resources:

```yaml
spec:
  artifacts:
    enabled: true
```

Configure storage via environment variables (see generated README for A2A_ARTIFACT_* variables). Supports both filesystem and MinIO/S3 storage backends.

**Examples:**
- `examples/go-agent-artifacts-filesystem.yaml` - Filesystem storage example
- `examples/go-agent-artifacts-minio.yaml` - MinIO storage example

## GitHub Issue Templates

The ADL CLI can automatically generate GitHub issue templates for your agent projects, providing standardized forms for bug reports, feature requests, and refactoring tasks:

```yaml
spec:
  scm:
    provider: github
    url: "https://github.com/company/my-agent"
    issue_templates: true # Enable issue template generation
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

## Dependabot Configuration

The ADL CLI can generate a `.github/dependabot.yml` manifest so generated agent
projects keep their dependencies up to date automatically. The feature defaults
to `false` and is opted into via `spec.scm.dependabot`:

```yaml
spec:
  scm:
    provider: github
    url: "https://github.com/company/my-agent"
    dependabot: true # Enable Dependabot configuration
```

When `dependabot: true` is set (and the SCM provider is GitHub), the generator
emits a weekly-schedule manifest covering the ecosystems present in your ADL:

- **Language ecosystem** - `gomod`, `cargo`, or `npm` (selected from `spec.language`)
- **`github-actions`** - Keeps `.github/workflows/` actions pinned
- **`docker`** - Tracks the base image in the generated `Dockerfile`
- **`devcontainers`** - Included when `spec.development.sandbox.devcontainer.enabled: true`

Each ecosystem groups all updates into a single PR per week to keep noise low.
The default of `false` keeps the existing behavior unchanged for projects that
manage dependency upgrades themselves.

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
- `tools/{toolname}.go` → Individual function-call tool implementations
- `skills/{skillid}/SKILL.md` → Markdown skill playbooks (loaded into system prompt at runtime)
- `go.mod` → Go module configuration
- Language-specific Dockerfile and CI configurations

**Rust Projects:**

- `src/main.rs` → Rust main application
- `src/tools/{toolname}.rs` → Tool descriptor + handler
- `src/tools/mod.rs` → Module declarations
- `skills/{skillid}/SKILL.md` → Markdown skill playbooks
- `Cargo.toml` → Rust package configuration

**Universal Files:**

- `Taskfile.yml` → Development task runner
- `.well-known/agent-card.json` → A2A capabilities manifest (skills are listed here, not tools)
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

- Go 1.26.2+
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

Define custom acronyms in your ADL file's `spec.acronyms` field. These acronyms will be properly capitalized when generating identifiers in your code.

### Configuration

```yaml
spec:
  language:
    go:
      module: "github.com/company/my-agent"
      version: "1.26.2"
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
apiVersion: adl.inference-gateway.com/v1
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
    - "go mod download" # Download dependencies first
    - "go generate ./..." # Generate code if needed
    - "gofumpt -l -w ." # Improved formatting
    - "golangci-lint run --fix" # Lint and auto-fix
    - "go test -race -short ./..." # Run tests
    - "go build -v ./..." # Verify build works
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

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://docs.inference-gateway.com)
- 💬 [Discussions](https://github.com/inference-gateway/adl-cli/discussions)
- 🐛 [Issues](https://github.com/inference-gateway/adl-cli/issues)

---

> 🤖 Powered by the [Inference Gateway ecosystem](https://github.com/inference-gateway/)
