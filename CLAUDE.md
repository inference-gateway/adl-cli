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
    ├── registry/            # Skills registry client (fetch + cache + frontmatter)
    │   ├── client.go        # HTTP client for the skills registry
    │   ├── cache.go         # Local on-disk cache (~/.adl/skills-cache/)
    │   ├── frontmatter.go   # YAML frontmatter parser for skill markdown
    │   └── resolver.go      # Coordinates fetch/cache/scaffold for skills
    ├── schema/              # ADL schema definitions
    │   ├── types.go         # All ADL type definitions (ADL, Spec, Tool, Skill, etc.)
    │   └── validator.go     # JSON Schema validation
    ├── prompt/              # Interactive prompts for `init` command
    └── templates/           # Template system
        ├── engine.go        # Template rendering with Sprig v3 functions
        ├── registry.go      # Template loading and file mapping per language
        ├── headers.go       # Generated file headers
        ├── common/          # Universal templates (config, docs, CI/CD, skills/)
        ├── languages/       # Language-specific templates (go/, rust/, typescript/)
        └── sandbox/         # Dev environment templates (flox/, devcontainer/)
```

### Generation Flow

1. **Parse & Validate**: Load ADL YAML, validate against schema, check required fields
2. **Resolve skills**: Pre-fetch every `spec.skills[]` entry from the registry (or cache); scaffold bare skills from their inline metadata
3. **Template Selection**: Detect language from `spec.language`, build file mapping via `registry.go`
4. **Generate**: Render templates with ADL context (including resolved skill views), respect `.adl-ignore`, write files; non-bare skill markdown bodies are written verbatim
5. **Post-Process**: Run formatters (`go fmt`, `cargo fmt`), execute custom hooks

### Key Types (internal/schema/types.go)

- `ADL` - Root structure with `apiVersion`, `kind`, `metadata`, `spec`
- `Spec` - Contains `capabilities`, `agent`, `tools`, `skills`, `services`, `server`, `language`, `deployment`, `development`
- `DevelopmentConfig` - Local development experience under `spec.development`: `sandbox` (flox/devcontainer/dockerCompose) and `ai` (CLAUDE.md/AGENTS.md generation). Introduced in ADL v0.6.0 - previously these sat directly under `spec`.
- `Tool` - Function-call entrypoint with `id` (only `id` is required since v0.4.0). User-defined tools also set `name`, `description`, `tags`, `schema`, optional `inject`. Reserved IDs (`read`, `bash`, `write`, `edit`) take `id` alone - the generator owns metadata. See `internal/schema/builtin_config.go` for the reserved-ID set.
- `Skill` - Markdown playbook with `id` plus optional `version`/`source`/`bare`/`name`/`description`/`tags`. Generated as `skills/<id>/SKILL.md`. **At runtime, only the frontmatter is consumed** - the body lives on disk and the model reads it on demand via the `Read` built-in.
- `Service` - Injectable service with `interface`, `factory`, `description`

### Tools vs Skills

- **Tools** are functions the agent can invoke. Defined under `spec.tools`. Two kinds:
  - **User tools** - full entry (`id`, `name`, `description`, `tags`, `schema`, optional `inject`). Generated as `tools/<id>.<ext>` from `tool.<ext>.tmpl`.
  - **Reserved built-ins** - `id` only. Currently `read`, `bash`, `write`, `edit`. Generated from `builtin/<id>.<ext>.tmpl`. **All four default to `enabled: false`**; activate via `spec.config.tools.<id>.enabled: true`. The Read built-in is what loads SKILL.md bodies on demand - `spec.skills` non-empty REQUIRES `- id: read` listed AND enabled (validator enforces).
- **Skills** are markdown documents (YAML frontmatter + body). The frontmatter (`name`/`description`) is consumed at runtime to build an `AVAILABLE SKILLS:` block appended to the system prompt; the body is fetched on demand by the model via Read. Each skill goes to `skills/<id>/SKILL.md`. Resolution paths unchanged:
  - **`bare: true`** → scaffolded from inline `name`/`description`/`tags`; the whole `skills/<id>/` directory is listed in `.adl-ignore`.
  - **`source:` set** → must be a `github.com` `/tree/<ref>/<path>` URL or one of the shorthand forms below; the full directory is fetched via the GitHub trees API + `raw.githubusercontent.com`. Implemented in `internal/registry/installer.go`.
  - **Otherwise** → fetch `registry.inference-gateway.com/skills/<id>[/<version>].md`. Registry-by-id ships SKILL.md only.

  `source:` shorthand (optional `@<tag>` pins branch/tag/SHA, default `main`): `<skill>[@<tag>]` → `inference-gateway/skills`; `<owner>/<repo>/<skill>[@<tag>]` → arbitrary repo; full URL passes through. Non-`github.com` URLs are rejected. `adl generate --offline` skips network; every non-bare skill must be cached at `~/.adl/skills-cache/<id>@<ref>/`.

### Reserved tool config (`spec.config.tools.<id>`)

Built-in config lives under the reserved `spec.config.tools` namespace (not `spec.config.<id>` - that name is reserved). The validator decodes each block into a typed struct in `internal/schema/builtin_config.go` (`ReadBuiltinConfig`, `BashBuiltinConfig`, `WriteBuiltinConfig`, `EditBuiltinConfig`) with `ErrorUnused: true`, so typos surface immediately as `spec.config.tools.bash.tymeout_seconds`. Values flow into the generated runtime as compile-time literals in `main.<ext>`'s tool-registration block - they do NOT flow through `config/config.go` (which explicitly skips the `tools` section).

Bash env-var runtime overrides (read inside `tools/bash.<ext>`): `A2A_BASH_DISABLED=1` (kill switch), `A2A_BASH_WHITELIST=ls,cat,grep`. Resolution precedence: **env > compile-time literal > built-in default (disabled)**.

### Service Injection

Tools can inject services via the `inject` field. The `logger` service is built-in:

```yaml
spec:
  services:
    database:
      type: service
      interface: DatabaseService
      factory: NewDatabaseService
  tools:
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
      version: "1.94.1"
      edition: "2024"
      features:
        - redis
```

Runtime configuration (`A2A_QUEUE_PROVIDER`, `A2A_QUEUE_URL`, `A2A_QUEUE_NAMESPACE`)
is documented in the generated `.env.example` - not baked into `main.rs`. When
`spec.development.sandbox.dockerCompose.enabled: true` is also set, a working
`docker-compose.yaml` with a Redis service is produced alongside the agent.

## Adding a New Language

1. Create `internal/templates/languages/<lang>/` with templates
2. Add file mapping method in `registry.go` (e.g., `getRustFiles`)
3. Add language detection in `DetectLanguageFromADL` and `detectLanguage`
4. Add the language config to `schema/v1/schema.json` in the `adl` repo (e.g., `RustConfig` under `$defs`), then `task fetch-schema` + `task generate-types` here
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
task fetch-schema     # refresh internal/schema/schema.json from upstream
task generate-types   # regenerate internal/schema/types.go from the schema
task verify-schema    # CI gate: confirm committed schema matches upstream
```

`task verify-schema` runs as part of `task ci`.

`internal/schema/types.go` is **generated** by `atombender/go-jsonschema`
(see `task generate-types`) and **must not be edited by hand**. The schema is
authoritative; bump the schema in the `adl` repo, refresh, and regenerate.

`task generate-types` first runs `internal/schema/annotate` against the
committed schema. The annotator emits a transient copy with
`goJSONSchema: {pointer: false}` injected on every optional scalar / enum
property, so `go-jsonschema` emits value types (e.g. `string`, `bool`,
`SCMProvider`) instead of `*T` for them. Nested optional struct fields
(`*DeploymentConfig`, `*Card`, `*Agent`, …) keep their pointers so callers
can still nil-check for "section absent." The committed
`internal/schema/schema.json` is never written to, so `task verify-schema`
keeps passing.

Hand-written companion (only one):

- `internal/schema/metadata.go` - `GeneratedMetadata` struct used by the
  templating layer (not part of the ADL spec).

## Important Notes

- Go 1.26.2+ required
- Templates use Go `text/template` with Sprig v3 functions
- ADL schema version: `adl.inference-gateway.com/v1`
- Supports Go, Rust (TypeScript planned)
- Use table-driven tests with isolated mocks
