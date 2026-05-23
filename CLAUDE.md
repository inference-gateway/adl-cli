# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build & Test

- `task build` - Build the ADL CLI binary to `bin/adl`
- `task test` - Run all tests
- `task test:coverage` - Run tests with coverage report
- `task lint` - Run golangci-lint
- `task lint:md` / `task lint:md:fix` - Run markdownlint (auto-fix variant available)
- `task fmt` - Format all Go code (`go fmt`)
- `task vet` - Run `go vet`
- `task install` - Build and install the CLI to `$GOPATH/bin` as `adl`
- `task generate-types` - Regenerate `internal/schema/types.go` from the embedded schema
- `task ci` - Run complete CI pipeline: fmt, lint, test, build, verify-schema

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
    │   ├── installer.go     # GitHub-source skill installer (trees API + raw.githubusercontent.com)
    │   └── resolver.go      # Coordinates fetch/cache/scaffold for skills
    ├── sandbox/             # Resolver for spec.development.deps (sandbox-level extras)
    │   └── deps.go          # Parse/dedupe/sort `<pkg>@<ver>` entries, flag flox/devcontainer conflicts
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
- `DevelopmentConfig` - Local development experience under `spec.development`: `sandbox` (flox/devcontainer/dockerCompose), `ai` (CLAUDE.md/AGENTS.md generation), and `deps` (cross-cutting sandbox-level extras like `deno`, `kubectl`, `terraform`). Introduced in ADL v0.6.0 (sandbox+ai); `deps` added in ADL v0.10.0.
- `Tool` - Function-call entrypoint with `id` (only `id` is required since v0.4.0). User-defined tools also set `name`, `description`, `tags`, `schema`, optional `inject`. Reserved IDs (`read`, `bash`, `write`, `edit`) take `id` alone - the generator owns metadata. See `internal/schema/builtin_config.go` for the reserved-ID set.
- `Skill` - Markdown playbook with `id` plus optional `version`/`source`/`bare`/`name`/`description`/`tags`/`license`. Generated as `skills/<id>/SKILL.md`. **At runtime, only the frontmatter is consumed** - the body lives on disk and the model reads it on demand via the `Read` built-in. `license` accepts the SPDX identifiers enumerated in the schema (`MIT`, `Apache-2.0`, `BSD-2-Clause`, `BSD-3-Clause`, `GPL-2.0`, `GPL-3.0`, `LGPL-2.1`, `LGPL-3.0`, `MPL-2.0`, `ISC`, `CC0-1.0`, `CC-BY-4.0`, `CC-BY-SA-4.0`, `Unlicense`) plus `Proprietary` for closed-source skills. The resolver mirrors it into the generated `SKILL.md` frontmatter so the licence travels with the playbook; when both the ADL entry and the fetched frontmatter set `license`, the ADL value wins.
- `Service` - Injectable service with `interface`, `factory`, `description`

### Tools vs Skills

- **Tools** are functions the agent can invoke. Defined under `spec.tools`. Two kinds:
  - **User tools** - full entry (`id`, `name`, `description`, `tags`, `schema`, optional `inject`). Generated as `tools/<id>.<ext>` from `tool.<ext>.tmpl`.
  - **Reserved built-ins** - `id` only. Currently `read`, `bash`, `write`, `edit`, `fetch`. Generated from `builtin/<id>.<ext>.tmpl`. **All five default to `enabled: false`**; activate via `spec.config.tools.<id>.enabled: true`. The Read built-in is what loads SKILL.md bodies on demand - `spec.skills` non-empty REQUIRES `- id: read` listed AND enabled (validator enforces). The `fetch` built-in performs HTTP(S) GET/HEAD calls subject to an `allowed_domains` whitelist, a `max_bytes` cap, a request `timeout_seconds`, and (when `allow_downloads: true`) writes responses into `download_dir` (default `/tmp`). Go uses `net/http` (no extra dep); Rust adds `reqwest` (rustls-tls + json) only when `fetch` is listed.
- **Skills** are markdown documents (YAML frontmatter + body). The frontmatter (`name`/`description`) is consumed at runtime to build an `AVAILABLE SKILLS:` block appended to the system prompt; the body is fetched on demand by the model via Read. Each skill goes to `skills/<id>/SKILL.md`. Resolution paths unchanged:
  - **`bare: true`** → scaffolded from inline `name`/`description`/`tags`; the whole `skills/<id>/` directory is listed in `.adl-ignore`.
  - **`source:` set** → must be a `github.com` `/tree/<ref>/<path>` URL or one of the shorthand forms below; the full directory is fetched via the GitHub trees API + `raw.githubusercontent.com`. Implemented in `internal/registry/installer.go`.
  - **Otherwise** → fetch `registry.inference-gateway.com/skills/<id>[/<version>].md`. Registry-by-id ships SKILL.md only.

  `source:` shorthand (optional `@<tag>` pins branch/tag/SHA, default `main`): `<skill>[@<tag>]` → `inference-gateway/skills`; `<owner>/<repo>/<skill>[@<tag>]` → arbitrary repo; full URL passes through. Non-`github.com` URLs are rejected. `adl generate --offline` skips network; every non-bare skill must be cached at `~/.adl/skills-cache/<id>@<ref>/`.

### Reserved tool config (`spec.config.tools.<id>`)

Built-in config lives under the reserved `spec.config.tools` namespace (not `spec.config.<id>` - that name is reserved). The validator decodes each block into a typed struct in `internal/schema/builtin_config.go` (`ReadBuiltinConfig`, `BashBuiltinConfig`, `WriteBuiltinConfig`, `EditBuiltinConfig`, `FetchBuiltinConfig`) with `ErrorUnused: true`, so typos surface immediately as `spec.config.tools.bash.tymeout_seconds`. Values flow into the generated runtime as compile-time literals in `main.<ext>`'s tool-registration block - they do NOT flow through `config/config.go` (which explicitly skips the `tools` section).

Bash env-var runtime overrides (read inside `tools/bash.<ext>`): `A2A_BASH_DISABLED=1` (kill switch), `A2A_BASH_WHITELIST=ls,cat,grep`. Resolution precedence: **env > compile-time literal > built-in default (disabled)**.

Fetch env-var runtime overrides: Go uses `TOOLS_FETCH_ENABLED`, `TOOLS_FETCH_ALLOWED_DOMAINS`, `TOOLS_FETCH_MAX_BYTES`, `TOOLS_FETCH_TIMEOUT_SECONDS`, `TOOLS_FETCH_DOWNLOAD_DIR`, `TOOLS_FETCH_ALLOW_DOWNLOADS` (via `envconfig`); Rust uses `A2A_FETCH_DISABLED=1` (kill switch), `A2A_FETCH_ALLOWED_DOMAINS`, `A2A_FETCH_MAX_BYTES`, `A2A_FETCH_TIMEOUT_SECONDS`, `A2A_FETCH_DOWNLOAD_DIR`, `A2A_FETCH_ALLOW_DOWNLOADS`. An entry in `allowed_domains` that starts with `.` (e.g. `.example.com`) is a suffix match allowing any subdomain. `save_path` is rejected unless `allow_downloads: true`; absolute paths and parent-dir traversal are also rejected.

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

### Vendor (extra dependencies)

Every `spec.language.<lang>` config accepts an optional `vendor.{deps,devdeps}`
block whose entries follow the `<package>@<version>` form. They are resolved
by `internal/vendor` (parsed, deduped against the generator's built-in
dependency set, sorted), exposed on `templates.Context.Vendor` as
`GoRequires` / `GoTools` / `CargoDeps` / `CargoDevDeps` / `NpmDeps` /
`NpmDevDeps`, and rendered into `go.mod` / `Cargo.toml` directly.
Built-ins always win on conflict (`internal/vendor/vendor.go` lists them
as `GoBuiltins`, `CargoBuiltinDeps`, `CargoBuiltinDevDeps`); collisions
surface as stderr warnings from the generator. The TypeScript fields are
validated and resolved today but only land in `package.json` once the TS
templates exist (currently only `.gitkeep`).

**Go-specific semantics:** `vendor.deps` map straight into `go.mod`'s
`require` block (plus the built-ins). `vendor.devdeps` are treated as
executable tool dependencies and emitted via Go 1.24+'s
[`tool` directive](https://go.dev/doc/modules/managing-dependencies#tools);
each entry also lands in `require` as `// indirect` so the module is
downloadable. Users supply the full tool package path
(e.g. `golang.org/x/tools/cmd/stringer@v0.20.0`); running `go mod tidy`
after generation normalises the `require` entry to the module root.
Test libraries that are `import`-ed by `*_test.go` files (testify,
go-cmp, …) belong in `vendor.deps`, not `vendor.devdeps`.

When you update a built-in dependency in a language template, mirror the
new pin into the matching map in `internal/vendor/vendor.go` so vendor
conflicts continue to be caught.

### Sandbox extras (`spec.development.deps`)

`spec.development.deps` is the cross-cutting equivalent of
`vendor.deps` - same `<package>@<version>` shape, but resolved into the
sandbox manifests instead of `go.mod` / `Cargo.toml`. Entries are parsed
by `internal/sandbox`, deduped (first wins), sorted alphabetically, and
exposed on `templates.Context.SandboxDeps`. Two backends render from
this view today:

- **flox** - each entry appends `<pkg>.pkg-path = "<pkg>"` /
  `<pkg>.version = "<version>"` lines under `[install]` in
  `.flox/env/manifest.toml`. Resolved against Nixpkgs.
- **devcontainer** - all entries are joined into a single
  `ghcr.io/devcontainers-extra/features/apt-packages:1` feature entry
  with a comma-separated `<pkg>=<version>` list.

Merge policy is **additive**: the template's built-in toolchain
(`go`/`cargo`/`nodejs`, `git`, `docker`, `go-task`) is always emitted,
and user `deps` are appended on top. If a user entry's package name
collides with one of these built-ins (e.g. `git@2.53.0`), the
generator prints a `⚠️  spec.development.deps … collides with a Flox
built-in …` warning to stderr but still renders the user entry — the
maintainer's pin wins on version conflict. The conflict-detection
map lives in `internal/sandbox/deps.go` (`floxBuiltinPackages` /
`devcontainerBuiltinPackages`); keep them in sync if you add new
packages to the sandbox templates.

`dockerCompose` is out of scope for now — the docker-compose dev image
doesn't consume `spec.development.deps`. Tracked in issue #154.

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
