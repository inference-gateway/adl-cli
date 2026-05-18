# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build & Test

- `task build` - Build the ADL CLI binary to `bin/adl`
- `task test` - Run all tests
- `task test:coverage` - Run tests with coverage report
- `task lint` - Run golangci-lint
- `task fmt` - Format all Go code
- `task ci` - Run complete CI pipeline: fmt, lint, test, build

### Running Specific Tests

- `go test -v ./cmd -run TestInit` - Run specific test by name
- `go test -v ./internal/generator` - Test specific package

### Development Mode

- `task dev -- init my-agent` - Run CLI init command
- `task dev -- generate --file agent.yaml --output ./test` - Generate project
- `task dev -- validate examples/go-agent.yaml` - Validate ADL file
- `task examples:test` - Validate all example ADL files
- `task examples:generate` - Generate projects from all examples

## Architecture

The ADL CLI generates A2A (Agent-to-Agent) agent projects from YAML-based Agent Definition Language files.

### Core Components

```text
main.go                       # Entry point, sets version
├── cmd/                      # CLI commands (Cobra framework)
│   ├── root.go              # Root command setup
│   ├── init.go              # Interactive ADL manifest creation
│   ├── generate.go          # Project generation from ADL
│   └── validate.go          # ADL schema validation
└── internal/
    ├── generator/           # Code generation engine
    │   ├── generator.go     # Main generation logic, CI/CD generation
    │   └── ignore.go        # .adl-ignore handling
    ├── schema/              # ADL schema definitions
    │   ├── types.go         # All ADL type definitions (ADL, Spec, Skill, etc.)
    │   └── validator.go     # JSON Schema validation
    ├── prompt/              # Interactive prompts for `init` command
    └── templates/           # Template system
        ├── engine.go        # Template rendering with Sprig v3 functions
        ├── registry.go      # Template loading and file mapping per language
        ├── headers.go       # Generated file headers
        ├── common/          # Universal templates (config, docs, CI/CD)
        ├── languages/       # Language-specific templates (go/, rust/, typescript/)
        └── sandbox/         # Dev environment templates (flox/, devcontainer/)
```

### Generation Flow

1. **Parse & Validate**: Load ADL YAML, validate against schema, check required fields
2. **Template Selection**: Detect language from `spec.language`, build file mapping via `registry.go`
3. **Generate**: Render templates with ADL context, respect `.adl-ignore`, write files
4. **Post-Process**: Run formatters (`go fmt`, `cargo fmt`), execute custom hooks

### Key Types (internal/schema/types.go)

- `ADL` - Root structure with `apiVersion`, `kind`, `metadata`, `spec`
- `Spec` - Contains `capabilities`, `agent`, `skills`, `services`, `server`, `language`, `deployment`
- `Skill` - Agent capability with `id`, `name`, `schema`, `inject` (for service injection)
- `Service` - Injectable service with `interface`, `factory`, `description`

### Service Injection

Skills can inject services via the `inject` field. The `logger` service is built-in:

```yaml
spec:
  services:
    database:
      type: service
      interface: DatabaseService
      factory: NewDatabaseService
  skills:
    - id: query_database
      inject:
        - logger # logger is built-in
        - database
```

Generated services go to `internal/<service>/` with interface and factory function.
Service injection is Go-specific; the Rust generator does not consume `spec.services`.

### Rust Cargo Features

Opt-in to ADK Cargo features through `spec.language.rust.features`. Each entry is
forwarded to the `inference-gateway-adk` dependency in the generated `Cargo.toml`.
The `redis` feature enables the ADK's Redis-backed task queue:

```yaml
spec:
  language:
    rust:
      packageName: my-agent
      version: "1.88"
      edition: "2024"
      features:
        - redis
```

Runtime configuration (`A2A_QUEUE_PROVIDER`, `A2A_QUEUE_URL`, `A2A_QUEUE_NAMESPACE`)
is documented in the generated `.env.example` — not baked into `main.rs`. When
`spec.sandbox.dockerCompose.enabled: true` is also set, a working
`docker-compose.yaml` with a Redis service is produced alongside the agent.

## Adding a New Language

1. Create `internal/templates/languages/<lang>/` with templates
2. Add file mapping method in `registry.go` (e.g., `getRustFiles`)
3. Add language detection in `DetectLanguageFromADL` and `detectLanguage`
4. Add language config type in `schema/types.go` (e.g., `RustConfig`)
5. Add example ADL file in `examples/`

## Adding a New Command

1. Create `cmd/<command>.go` with Cobra command
2. Register in `cmd/root.go` via `rootCmd.AddCommand()`
3. Add tests in `cmd/<command>_test.go`

## ADL Schema Source of Truth

The canonical ADL JSON Schema lives in
[`inference-gateway/adl`](https://github.com/inference-gateway/adl) under
`schema/v1/schema.json`. This repo vendors it at `internal/schema/schema.json`
(embedded into the binary via `//go:embed` in `internal/schema/validator.go`).
The pinned upstream version is the `ADL_SCHEMA_VERSION` variable in
`Taskfile.yml`.

Update flow:

```bash
# bump the version in Taskfile.yml, then:
task fetch-schema     # writes internal/schema/schema.json
task verify-schema    # CI gate: confirms the committed file matches upstream
task generate-types   # (follow-up work) regenerates internal/schema/types.go
```

`task verify-schema` runs in `task ci` to catch accidental edits or drift.

Go types in `internal/schema/types.go` are currently hand-maintained and
**must be kept in sync with `schema/v1/schema.json`** when fields are added or
removed. A `generate-types` Taskfile target is wired up
(`atombender/go-jsonschema`) but the generated output uses pointer-for-optional
semantics, which would force a wider consumer refactor across templates,
`cmd/init.go`, and `internal/generator/generator.go`; that migration is tracked
as a follow-up.

## Important Notes

- Go 1.26.2+ required
- Templates use Go `text/template` with Sprig v3 functions
- ADL schema version: `adl.dev/v1`
- Supports Go, Rust (TypeScript planned)
- Use table-driven tests with isolated mocks
