# AGENTS.md

`adl` is a Go CLI that scaffolds complete A2A (Agent-to-Agent) projects from YAML Agent Definition Language (ADL) manifests. Three commands: `init` (interactive wizard that writes `agent.yaml`), `generate` (`agent.yaml` → Go/Rust/TypeScript project), `validate` (schema-check a manifest). The canonical schema lives upstream in `inference-gateway/adl`; this repo vendors it.

## Commands

All workflows go through Task (`task <name>`); plain `go` works if Task is missing.

| Task | Purpose |
| --- | --- |
| `task build` | Build CLI to `bin/adl` with `-X main.Version` |
| `task test` | `go test -v ./...` |
| `task fmt` / `task vet` / `task lint` | `go fmt ./...`, `go vet ./...`, `golangci-lint run` |
| `task lint:md` / `task lint:md:fix` | `markdownlint '**/*.md' --ignore CHANGELOG.md` (+ `--fix`) |
| `task ci` | Local pipeline: fmt → lint → test → build → verify-schema |
| `task examples:test` | Validates every YAML in `examples/` |
| `task examples:generate` | Generates every example into `test-output/` — the full smoke test |
| `task fetch-schema` / `verify-schema` / `generate-types` | Sync / drift-check / regenerate the vendored schema + types |

Single generator test: `go test -v ./internal/generator -run TestGenerator_Generate`. GitHub CI (`.github/workflows/ci.yml`) runs markdownlint, a `go mod tidy` drift check, lint, build and test — **it does not run `verify-schema`**, so run that locally whenever the schema changes.

## Architecture (read before editing)

- **Single binary.** All `.tmpl` files (`internal/templates/`) and the JSON schema are embedded via `go:embed` — editing a template requires a rebuild; there is no runtime file lookup.
- **Pipeline:** `cmd/generate.go` → `internal/generator` → `internal/templates`. Stages: parse (with card migration) → generator-level validation (invariants the JSON Schema can't express) → flag/manifest reconciliation → resolve skills, vendor pins and sandbox deps into `templates.Context` → render each file → post-generation shell. `internal/templates/registry.go` `GetFiles(adl)` maps output path → template key — wire new output files there. `engine.go` executes templates with Sprig + custom case-conversion funcs.
- **Never hand-edit generated code.** `internal/schema/types.go` (go-jsonschema output, DO NOT EDIT header) and `internal/schema/schema.json` (vendored upstream) regenerate via `task fetch-schema && task generate-types`; hand-written companions live beside them (`validator.go`, `card_migration.go`, `builtin_config.go`, `metadata.go`, `ai.go`). `ADL_SCHEMA_VERSION` in `Taskfile.yml` pins the upstream version; bumps also arrive as automated PRs from `sync-schema.yml`.
- **Version pins live in `internal/vendor/vendor.go`** (SDK deps, toolchains, GitHub Actions, semantic-release). Templates read them via `{{ pin "<group>" "<name>" }}` — an unknown key fails generation. Built-in pins win over `spec.vendor` entries (collisions warn and drop the vendor entry). Bump there, never in a `.tmpl`.
- **`.adl-ignore` protects user code.** The generator writes `.adl-ignore` into the *output project*, listing the TODO-placeholder files it just scaffolded (tool implementations, custom services, authz, bare skills); re-generate (even `--overwrite`) skips matched paths (`internal/generator/ignore.go`), and the file is only created once — afterwards it is user-owned. When adding a new placeholder template, register the file in `generateADLIgnoreFile`/`generateA2aIgnoreContent` (`internal/generator/generator.go`).
- **Flag/manifest reconciliation is OR semantics** — a CLI flag (`--ci`, `--cd`, `--deployment`, `--flox`, `--devcontainer`) wins over the manifest field; runtime config follows the same pattern: manifest values render as defaults into the generated `.env.example` (`A2A_*` vars) and env vars override them. `init` applies one precedence everywhere: flag/env (viper, `ADL_` prefix) > prompt > default.
- **Tools vs skills.** `spec.tools[]` → one generated file per tool (`tools/<id>.go`, `src/tools/<id>.ts`); five reserved IDs (`read`/`bash`/`write`/`edit`/`fetch`) render from `languages/<lang>/builtin/` (Go and Rust — Go adds `*_test.go`; TypeScript's ADK supplies its own) and can carry `spec.config.tools.<id>` config. `spec.skills[]` → markdown playbooks resolved from the skills registry (override env `ADL_SKILLS_REGISTRY`) or GitHub (`source:`) via `internal/registry`, cached under `~/.adl/skills-cache` — or scaffolded locally (`bare: true`).
- **AI orchestrators.** `spec.development.ai.orchestrators` (claude-code, codex, gemini, opencode, infer — init's `--ai` maps to claude-code only) toggle `ai-*.yaml` GitHub workflows plus AGENTS.md/gemini-settings scaffolding (GitHub SCM only; opencode is docs-only). The CI/CD/AI templates build their own minimal `templates.Context` — a new `Context` field must be added there too.
- **Sandbox.** `--flox` / `--devcontainer` (or `spec.development.sandbox.*.enabled`) render the Flox/devcontainer scaffolds; `spec.development.deps` is deduped by `internal/sandbox` with collision warnings; the docker-compose sandbox emits `docker-compose.yaml` + `.env.example`.
- **Regeneration protects user work.** Existing files are skipped unless `--overwrite`; `go.mod` `// indirect` requires are merged back in (dependabot bumps survive); generated-file headers are skipped for skills/tools/services; post-generation runs `spec.hooks.post` if set, else vendor-tool-aware `go get`/`tidy`/`fmt` (Go) or `cargo fmt` (Rust).
- **Telemetry** (`spec.telemetry.enabled`, manifest-only): Go and TypeScript only; Rust ignores it.

## Conventions & gotchas

- Go 1.27.x; Flox sandbox (`.flox/env/manifest.toml`) pins `go`, `golangci-lint` (v2), `go-task`, `prettier`, `markdownlint-cli`, `goreleaser` — `flox activate` to enter.
- Stale `a2a-cli` URLs and `git remote` values still work via GitHub redirect — the canonical repo is `inference-gateway/adl-cli`, matching the Go module and npm package names.
- The CLI version lives in four files — `Taskfile.yml` (`VERSION`), `npm/package.json`, `flake.nix`, `.flox/env/manifest.toml` — kept in sync by release automation (`release.yml`: semantic-release + goreleaser + npm wrapper that downloads the native binary via `checksums.txt`).
- Import order is enforced by the `gci` formatter (see `.golangci.yml`): standard library, `github.com/stretchr/testify`, third-party, `github.com/inference-gateway/*`, then this module. Every non-standard-library import must be named after its last path element (`yaml "gopkg.in/yaml.v3"`), enforced by `importas`. Fix locally with `golangci-lint fmt` and `golangci-lint run --fix`.
- Tabs in Go, 2-space in YAML/JSON/Markdown (`task format` runs prettier; `task lint:md:fix` autofixes markdownlint).
- Conventional commits (`feat:`, `fix:`, …) — semantic-release derives versions from them.
- **`examples/` is the regression suite.** When adding a feature, add/update an example and wire it into both lists (`examples:test` + `examples:generate`) in `Taskfile.yml`.
- `--offline` makes generate/init skip all network access (skill resolution, registry fetches).
- `cmd/init.go` hand-mirrors the manifest shape to write `agent.yaml`; its two front-ends (huh wizard and `--defaults`) converge on a pure `buildADL`, golden-tested in `cmd/buildadl_test.go` — schema bumps touch it plus `internal/schema/types.go`.
- `internal/schema/validator.go::checkLegacySpecFields` rejects pre-orchestrators manifest shapes with migration hints — keep it in sync whenever the schema shape changes (JSON Schema `additionalProperties:true` would otherwise silently drop legacy fields).

## Code Readability

- Write self-explanatory code: clear names and small, single-purpose functions carry the intent.
  If a block needs a comment to be understood, extract it into a well-named function or variable.
- No inline comments inside function bodies.
- Doc comments on functions and types are at most 5 lines: what it does and why, not how.
- No comments above modules, packages, or files.
- Tool directives are not comments and stay where the tool needs them (lint suppressions, build
  tags, compiler pragmas, code generation markers).
- No semicolons in documentation prose (Markdown files, doc comments): split the sentence or use
  a dash instead.
