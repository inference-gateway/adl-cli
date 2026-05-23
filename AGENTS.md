# AGENTS.md

This file provides guidance to AI agents working with the **ADL CLI** repository. It is complementary to CLAUDE.md and focuses on actionable information for autonomous coding agents.

---

## Project Overview

**ADL CLI** (`adl`) is a Go-based command-line tool that generates enterprise-ready A2A (Agent-to-Agent) server projects from YAML-based Agent Definition Language (ADL) files. It helps developers rapidly scaffold complete A2A agent projects with CI/CD, sandbox environments, skill registries, and built-in tools.

- **Repository**: `github.com/inference-gateway/adl-cli`
- **Language**: Go 1.26.2+
- **License**: Apache-2.0
- **Current Version**: 0.35.0

### Key Technologies

| Technology | Purpose |
|---|---|
| Go 1.26.2+ | Primary language |
| [Cobra](https://github.com/spf13/cobra) | CLI framework (`github.com/spf13/cobra v1.10.2`) |
| [Viper](https://github.com/spf13/viper) | Configuration management (`github.com/spf13/viper v1.21.0`) |
| [Sprig v3](https://github.com/Masterminds/sprig) | Template function library (`github.com/Masterminds/sprig/v3 v3.3.0`) |
| [go-jsonschema](https://github.com/atombender/go-jsonschema) | Go type generation from JSON Schema |
| [gojsonschema](https://github.com/xeipuuv/gojsonschema) | JSON Schema validation (`github.com/xeipuuv/gojsonschema v1.2.0`) |
| [mapstructure](https://github.com/go-viper/mapstructure) | Struct decoding for built-in tool configs |
| [Task](https://taskfile.dev/) | Task runner (build, test, lint) |
| [golangci-lint](https://golangci-lint.run/) | Linting |
| [GoReleaser](https://goreleaser.com/) | Release building |
| [semantic-release](https://semantic-release.gitbook.io/) | Automated releases |
| [Nix Flake](https://nixos.wiki/wiki/Flakes) | Development shell |
| [Flox](https://flox.dev) | Development environment |

### Generated Language Targets

| Language | Status | Key Templates |
|---|---|---|
| Go | Full support | `internal/templates/languages/go/` |
| Rust | Full support | `internal/templates/languages/rust/` |
| TypeScript | Template structure exists, implementation planned | `internal/templates/languages/typescript/` |

---

## Architecture

### Component Map

```text
main.go                       # Entry point, injects build version
├── cmd/                      # CLI commands (Cobra framework)
│   ├── root.go              # Root command, Viper init, signal handling
│   ├── init.go              # Interactive ADL manifest creation wizard (~1000 lines)
│   ├── generate.go          # Project generation from ADL (orchestrates generator)
│   └── validate.go          # ADL schema validation
├── internal/
│   ├── generator/           # Code generation engine
│   │   ├── generator.go     # Main generation logic, CI/CD gen, post-hooks (~1100 lines)
│   │   └── ignore.go        # .adl-ignore pattern matching
│   ├── registry/            # Skills registry (fetch + cache + frontmatter + installer)
│   │   ├── client.go        # HTTP client for default registry API
│   │   ├── cache.go         # On-disk cache at ~/.adl/skills-cache/
│   │   ├── frontmatter.go   # YAML frontmatter parser for SKILL.md
│   │   ├── installer.go     # GitHub-source skill installer (Trees API + raw content)
│   │   └── resolver.go      # Coordinates fetch/cache/scaffold for skills[]
│   ├── schema/              # ADL schema definitions and validation
│   │   ├── types.go         # Generated Go types from JSON Schema (DO NOT EDIT BY HAND)
│   │   ├── validator.go     # JSON Schema validation + extra semantic checks
│   │   ├── metadata.go      # GeneratedMetadata struct (hand-written)
│   │   ├── builtin_config.go # Reserved tool IDs (read/bash/write/edit/fetch) and config structs
│   │   └── schema.json      # Vendored ADL JSON Schema (pinned upstream version)
│   ├── prompt/              # Interactive readline prompts for `init` command
│   └── templates/           # Go text/template system (embedded via //go:embed)
│       ├── engine.go        # Template rendering with Sprig + custom funcs
│       ├── registry.go      # Template loading + file mapping per language
│       ├── headers.go       # Generated file "DO NOT EDIT" headers
│       ├── common/          # Universal templates across all languages
│       │   ├── ai/          # CLAUDE.md.tmpl, AGENTS.md.tmpl
│       │   ├── config/      # agent.json, editorconfig, gitattributes, gitignore, releaserc
│       │   ├── docker/      # Dockerfile templates (Go, Rust)
│       │   ├── docs/        # README.md template
│       │   ├── github/      # CI/CD workflows, issue templates, dependabot
│       │   ├── kubernetes/  # k8s deployment.yaml
│       │   ├── skills/      # SKILL.md template for bare skills
│       │   └── taskfile/    # Taskfile.yml template
│       ├── languages/       # Language-specific templates
│       │   ├── go/          # main.go, go.mod, config.go, logger.go, service.go, tool.go, builtin/*
│       │   ├── rust/        # main.rs, Cargo.toml, tool.rs, tool.mod.rs, env.example, builtin/*
│       │   └── typescript/  # Template structure only (not yet implemented)
│       └── sandbox/         # Environment templates (flox/, devcontainer/)
```

### Generation Flow

The `adl generate` command follows this pipeline:

1. **Parse & Validate**: Read ADL YAML file, validate against JSON Schema (`internal/schema/schema.json`), check semantic rules (service injection integrity, skill-read contract)
2. **Reconcile CLI Flags with Manifest**: Merge `--ai/--ci/--cd/--deployment/--flox/--devcontainer` flags with ADL manifest values (CLI flag OR's on top)
3. **Resolve Skills**: For each `spec.skills[]` entry:
   - `bare: true` → scaffold SKILL.md from inline metadata
   - `source:` set → fetch full directory from GitHub (Trees API + raw.githubusercontent.com)
   - otherwise → fetch from skills registry (`registry.inference-gateway.com/skills/<id>.md`)
   - Cache everything at `~/.adl/skills-cache/`; `--offline` requires pre-cached skills
4. **Template Selection**: Detect language from `spec.language`, build file mapping via `registry.go`
5. **Generate**: Render each template with the rich Context object, resolve services/tools/skills per file, respect `.adl-ignore` patterns
6. **Post-Process**: Run formatters (`go fmt`, `go mod tidy`, `cargo fmt`) and custom hooks from `spec.hooks.post`

### Key Types (`internal/schema/types.go`)

- **`ADL`** - Root structure: `apiVersion`, `kind`, `metadata`, `spec`
- **`Spec`** - Contains `capabilities`, `agent`, `tools`, `skills`, `services`, `server`, `language`, `config`, `deployment`, `development`, `scm`, `acronyms`, `hooks`
- **`Tool`** - Function-call entrypoint. Can be:
  - **User tool**: Full entry (`id`, `name`, `description`, `tags`, `schema`, `inject`)
  - **Reserved built-in**: `id` only (generator owns metadata). IDs: `read`, `bash`, `write`, `edit`, `fetch`
- **`Skill`** - Markdown playbook with `id` + optional `version`/`source`/`bare`/`name`/`description`/`tags`/`license`
- **`Service`** - Injectable service with `type`, `interface`, `factory`, `description`
- **`DevelopmentConfig`** - Contains `sandbox` (flox/devcontainer/dockerCompose) and `ai` (CLAUDE.md/AGENTS.md generation)

### Reserved Built-in Tools (`internal/schema/builtin_config.go`)

These tools have framework-supplied implementations generated when listed in `spec.tools`:

| ID | Generated as | Config namespace | Runtime overrides |
|---|---|---|---|
| `read` | `tools/read.go` | `spec.config.tools.read` | None |
| `bash` | `tools/bash.go` | `spec.config.tools.bash` | `A2A_BASH_DISABLED`, `A2A_BASH_WHITELIST` |
| `write` | `tools/write.go` | `spec.config.tools.write` | None |
| `edit` | `tools/edit.go` | `spec.config.tools.edit` | None |
| `fetch` | `tools/fetch.go` | `spec.config.tools.fetch` | Go: `TOOLS_FETCH_*` / Rust: `A2A_FETCH_*` |

**All five default to `enabled: false`.** Activate via `spec.config.tools.<id>.enabled: true`.

**Critical constraint**: If `spec.skills` is non-empty, the agent NEEDS `- id: read` listed AND `spec.config.tools.read.enabled: true` to load SKILL.md bodies at runtime. The validator enforces this with warnings.

---

## Development Environment Setup

### Prerequisites

- Go 1.26.2+
- [Task](https://taskfile.dev/) (recommended for development commands)
- [golangci-lint](https://golangci-lint.run/) (for linting)
- [markdownlint](https://github.com/DavidAnson/markdownlint-cli2) (for markdown linting)
- Git

### Quick Start

```bash
# Clone and enter
git clone https://github.com/inference-gateway/adl-cli.git
cd adl-cli

# Install dependencies
go mod download

# Build
task build

# Run tests
task test

# Run the CLI in development mode
task dev -- validate examples/go-agent.yaml

# Or run directly
./bin/adl validate examples/go-agent.yaml
```

### Nix Flox Development

```bash
# Enter development shell with all tools
nix develop github:inference-gateway/adl-cli

# Or using Flox (already configured in .flox/)
flox activate
```

---

## Key Commands

### Build & Run

| Command | Description |
|---|---|
| `task build` | Build binary to `bin/adl` |
| `task install` | Install to `$GOPATH/bin/adl` |
| `task dev -- <args>` | Build + run with CLI args (e.g., `task dev -- init my-agent`) |
| `task clean` | Remove `bin/` and `dist/` |

### Testing

| Command | Description |
|---|---|
| `task test` | Run all tests (`go test -v ./...`) |
| `task test:coverage` | Run tests with coverage (`go test -v -cover ./...`) |
| `go test -v ./cmd -run TestInit` | Run specific test by name |
| `go test -v ./internal/generator` | Test specific package |
| `task examples:test` | Validate all example ADL files |
| `task examples:generate` | Generate projects from all examples |

### Code Quality

| Command | Description |
|---|---|
| `task fmt` | Format Go code (`go fmt ./...`) |
| `task format` | Run Prettier formatter |
| `task vet` | Run `go vet ./...` |
| `task lint` | Run `golangci-lint run` |
| `task lint:md` | Run markdownlint |
| `task lint:md:fix` | Run markdownlint with auto-fix |
| `task ci` | Full CI pipeline: fmt, lint, test, build, verify-schema |

### Schema & Code Generation

| Command | Description |
|---|---|
| `task fetch-schema` | Refresh vendored `internal/schema/schema.json` from upstream ADL repo |
| `task generate-types` | Regenerate `internal/schema/types.go` from vendored schema |
| `task verify-schema` | CI gate: confirm committed schema matches upstream pinned version |

### Release

| Command | Description |
|---|---|
| `task release` | Build release binaries for multiple platforms (GoReleaser snapshot) |
| `task docker:build` | Build Docker image |

### CLI Usage (generated project workflow)

```bash
# Interactive project creation
task dev -- init my-agent

# Generate project code
task dev -- generate --file agent.yaml --output ./my-agent

# Generate with CI/CD/AI/Deployment
task dev -- generate --file agent.yaml --output ./my-agent --ci --ai --deployment cloudrun

# Validate ADL file
task dev -- validate examples/go-agent.yaml

# Generate offline (no network for skills)
task dev -- generate --file agent.yaml --output ./my-agent --offline
```

---

## Testing Instructions

### Test Patterns

This project uses standard Go testing with table-driven tests. Tests are located alongside the code they test (`*_test.go` files).

**Key patterns to follow:**

1. **Table-driven tests** - Define test cases as struct slices with `name`, input, expected output, and `wantErr` fields
2. **Temp directories** - Use `t.TempDir()` for temporary output directories (auto-cleaned)
3. **Global state save/restore** - Tests that modify global state (e.g., `adlFile`, `outputDir` variables) save originals in `defer` closures:

```go
originalADLFile := adlFile
originalOutputDir := outputDir
defer func() {
    adlFile = originalADLFile
    outputDir = originalOutputDir
}()
adlFile = adlPath
outputDir = outputPath
```

1. **Inline ADL content** - Tests define ADL content as raw YAML string literals, write them to temp files, then invoke `runGenerate()`
2. **File existence checks** - Use `os.Stat()` + `os.IsNotExist()` to verify generated files
3. **Content assertions** - Read generated files and check for expected substrings
4. **Isolated mocks** - Mock external dependencies (registry client, GitHub API)

### Test Files

| File | Tests |
|---|---|
| `cmd/generate_test.go` | Generation command with various flags (CI/CD/AI/overrides/skills) |
| `cmd/init_test.go` | Init command interactive flow |
| `cmd/validate_test.go` | Validation command (if exists) |
| `internal/generator/generator_test.go` | Core generation logic |
| `internal/schema/validator_test.go` | Schema validation |
| `internal/schema/builtin_config_test.go` | Built-in tool config decoding |
| `internal/registry/*_test.go` | Registry client, cache, installer, resolver |
| `internal/templates/*_test.go` | Template engine and registry |
| `internal/templates/cli_template_test.go` | CLI template rendering |

---

## Project Conventions

### Coding Standards

- **Go Style**: Standard Go conventions (`gofmt`, `golint`). Code must pass `go fmt ./...` and `golangci-lint run`.
- **Naming**: Use meaningful names. Follow Go idioms (e.g., `NewXxx()` for constructors).
- **Error Handling**: Return errors from `RunE` functions in Cobra commands. Use `fmt.Errorf("context: %w", err)` for wrapping.
- **Comments**: Document exported functions. Use Go-style doc comments.
- **Organization**: Keep functions focused and reasonably sized. Use early returns to reduce nesting.

### Template System Conventions

- Templates use Go `text/template` with Sprig v3 functions
- Template files have `.tmpl` extension
- Embedded at compile time via `//go:embed` in `internal/templates/registry.go`
- Template functions available: `toPascalCase`, `toCamelCase`, `toSnakeCase`, `toUpperSnakeCase`, `toJson`, `toGoMap`, `isBuiltinToolID`, `builtinToolMeta`
- Context object (`templates.Context`) provides full ADL data to all templates
- Generated files get "DO NOT EDIT" headers (controlled by `headers.go`)

### Acronym Handling

The template engine supports configurable acronyms for proper code generation. Default acronyms include: `id`, `api`, `url`, `uri`, `json`, `xml`, `html`, `http`, `https`, `sql`, `css`, `js`, `ui`, `uuid`, `tcp`, `udp`, `ip`, `dns`, `tls`, `ssl`, `cpu`, `gpu`, `ram`, `io`, `os`, `db`.

Custom acronyms can be added via `spec.acronyms` in the ADL file. They extend the defaults and take precedence.

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```text
feat: add TypeScript template support

- Add TypeScript template with Express.js framework
- Include enterprise features (auth, metrics, logging)

Fixes #123
```

Types: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`

### .adl-ignore System

The `.adl-ignore` file protects user-implemented files from being overwritten during regeneration. It works like `.gitignore`:

- Exact file paths: `tools/my_tool.go`
- Wildcards: `*.go`
- Directories: `skills/my-skill/`
- Comments start with `#`

The `internal/generator/ignore.go` `IgnoreChecker` handles pattern matching. User tool files, service files, and bare skill directories are automatically listed.

---

## Important Files

### Configuration & Build

| File | Purpose |
|---|---|
| `main.go` | Entry point, sets `Version` via ldflags |
| `go.mod` | Go module definition (`github.com/inference-gateway/adl-cli`) |
| `Taskfile.yml` | Task runner configuration (build, test, lint, ci, etc.) |
| `flake.nix` / `flake.lock` | Nix flake for development shell |
| `.flox/env/manifest.toml` | Flox environment configuration |
| `.goreleaser.yaml` | GoReleaser release configuration (if exists) |

### Schema & Core Types

| File | Purpose |
|---|---|
| `internal/schema/types.go` | **Generated** Go types from JSON Schema (DO NOT EDIT BY HAND) |
| `internal/schema/schema.json` | Vendored ADL JSON Schema (refresh via `task fetch-schema`) |
| `internal/schema/metadata.go` | Hand-written `GeneratedMetadata` struct |
| `internal/schema/builtin_config.go` | Reserved tool IDs and typed config structs |
| `internal/schema/validator.go` | Schema validation + semantic checks |

### Templates

| Directory | Purpose |
|---|---|
| `internal/templates/languages/go/` | Go project templates |
| `internal/templates/languages/go/builtin/` | Built-in tool Go templates (read, bash, write, edit, fetch) |
| `internal/templates/languages/rust/` | Rust project templates |
| `internal/templates/languages/rust/builtin/` | Built-in tool Rust templates |
| `internal/templates/common/` | Universal templates (CI/CD, Docker, docs, config, skills) |
| `internal/templates/sandbox/` | Development environment templates (flox, devcontainer) |

### Registry

| File | Purpose |
|---|---|
| `internal/registry/client.go` | HTTP client for default skills registry |
| `internal/registry/cache.go` | On-disk cache at `~/.adl/skills-cache/` |
| `internal/registry/installer.go` | GitHub-source skill installer |
| `internal/registry/resolver.go` | Coordinates fetch/cache/scaffold for skills |
| `internal/registry/frontmatter.go` | YAML frontmatter parser for SKILL.md |

### Examples

| File | Purpose |
|---|---|
| `examples/go-agent.yaml` | Basic Go agent example |
| `examples/rust-agent.yaml` | Basic Rust agent example |
| `examples/go-agent-builtin-tools.yaml` | Go agent with built-in tools |
| `examples/go-agent-artifacts-filesystem.yaml` | Go agent with filesystem artifacts |
| `examples/go-agent-artifacts-minio.yaml` | Go agent with MinIO artifacts |
| `examples/cloudrun-agent.yaml` | CloudRun deployment with GCR |
| `examples/cloudrun-ghcr-agent.yaml` | CloudRun deployment with GHCR |
| `examples/kubernetes-agent.yaml` | Kubernetes deployment |
| `examples/rust-agent-ai.yaml` | Rust AI agent |
| `examples/rust-agent-redis.yaml` | Rust agent with Redis |

---

## Adding a New Language

1. Create `internal/templates/languages/<lang>/` with `.tmpl` template files
2. Add file mapping method in `internal/templates/registry.go` (e.g., `getRustFiles()`)
3. Add language detection in `DetectLanguageFromADL()` and `detectLanguage()` in `generator.go`
4. Add language config to the upstream `adl` repo's `schema/v1/schema.json`, then `task fetch-schema` + `task generate-types` here
5. Add language config handling in `internal/generator/generator.go` (`validateADL`, `.adl-ignore` generation, post-generation hooks)
6. Add example ADL file in `examples/`
7. Add test cases for the new language

## Adding a New Command

1. Create `cmd/<command>.go` with a Cobra command struct
2. Register in `init()` via `rootCmd.AddCommand(<command>Cmd)`
3. Add flag definitions in `init()`
4. Add tests in `cmd/<command>_test.go`

## ADL Schema Maintenance

The canonical ADL JSON Schema lives in the separate [`inference-gateway/adl`](https://github.com/inference-gateway/adl) repository. This CLI vendors a pinned copy at `internal/schema/schema.json`.

Update flow:
```bash
# 1. Bump ADL_SCHEMA_VERSION in Taskfile.yml
# 2. Refresh vendored schema
task fetch-schema
# 3. Regenerate Go types
task generate-types
# 4. Verify everything matches
task verify-schema
```

The `task ci` command includes `verify-schema` so CI catches drift.

**IMPORTANT**: `internal/schema/types.go` is generated by `atombender/go-jsonschema`. The `task generate-types` command first runs `internal/schema/annotate` to inject `goJSONSchema: {pointer: false}` on optional scalar/enum properties so they generate as value types instead of `*T`. Never edit `types.go` by hand - modify the schema and regenerate.

---

## Common Agent Tasks

### Running the CLI in Development

```bash
# Build and run with args
go run . validate examples/go-agent.yaml

# Or use the task shortcut
task dev -- validate examples/go-agent.yaml

# Quick generate test
task dev -- generate --file examples/go-agent.yaml --output /tmp/test-agent --overwrite
```

### Schema Validation Flow

```go
validator := schema.NewValidator()
warnings, err := validator.ValidateFile("agent.yaml")
// err != nil → structural validation failure
// warnings != nil → semantic concerns (e.g., skills without Read tool)
```

### Generator Flow

```go
gen := generator.New(generator.Config{
    Template:   "minimal",
    Overwrite:  true,
    Version:    "0.35.0",
    GenerateCI: true,
    EnableAI:   true,
})
err := gen.Generate(adlFilePath, outputDir)
```

### Registry Resolver Flow

```go
resolver, _ := registry.NewDefaultResolver()
resolved, err := resolver.ResolveAll(ctx, adl.Spec.Skills)
// resolved[].Files contains fetched files for non-bare skills
// resolved[].Files is empty for bare skills (template scaffolds SKILL.md)
```

---

## Important Constraints

- **Go 1.26.2+ required** - The project uses modern Go features
- **Never edit `types.go` by hand** - It's auto-generated from the JSON Schema
- **Never edit vendored `schema.json`** - It must match the upstream pinned version
- **ADL Schema version**: `adl.inference-gateway.com/v1` (pinned via `ADL_SCHEMA_VERSION` in `Taskfile.yml`)
- **Templates use Go `text/template`** (not `html/template`) with Sprig v3
- **CLI flag OR's on top of manifest** - `--ai`/`--ci`/`--cd` flags override any manifest value
- **Generated files get headers** - YAML, Go, Rust, Dockerfile, and Taskfile outputs include `@generated` markers
- **AI-generated files are marked** - `CLAUDE.md` and `AGENTS.md` get `linguist-generated=true` in `.gitattributes` so they collapse in PR diffs
- **Skills + Read tool contract** - Skills require the Read built-in to be listed AND enabled
- **Template paths** - All embedded template paths use forward slashes, filepath operations use `filepath.ToSlash()` for normalization
