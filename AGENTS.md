# AGENTS.md

`adl` is a Go CLI that scaffolds complete A2A (Agent-to-Agent) projects from YAML Agent Definition Language (ADL) manifests. Three commands: `init` (interactive wizard that writes `agent.yaml`), `generate` (`agent.yaml` → Go/Rust/TypeScript project), `validate` (schema-check a manifest). The canonical schema lives upstream in `inference-gateway/adl`; this repo vendors it.

## Commands

All workflows go through Task (`task <name>`); plain `go` works if Task is missing.

| Task | Purpose |
| --- | --- |
| `task build` | Build CLI to `bin/adl` with `-X main.Version` |
| `task test` | `go test -v ./...` |
| `task fmt` / `task vet` / `task lint` | `go fmt ./...`, `go vet ./...`, `golangci-lint run` |
| `task lint:md` | `markdownlint '**/*.md' --ignore CHANGELOG.md` |
| `task ci` | fmt → lint → test → build → verify-schema (what CI runs) |
| `task examples:test` | Validates every YAML in `examples/` |
| `task examples:generate` | Generates every example into `test-output/` — the full smoke test |
| `task fetch-schema` / `verify-schema` / `generate-types` | Sync / drift-check / regenerate the vendored schema + types |

Single test: `go test -v ./internal/generator -run TestGenerate_Go`.

## Architecture (read before editing)

- **Single binary.** All `.tmpl` files (`internal/templates/`) and the JSON schema are embedded via `go:embed` — editing a template requires a rebuild; there is no runtime file lookup.
- **Pipeline:** `cmd/generate.go` → `internal/generator` → `internal/templates`. `internal/templates/registry.go` `GetFiles(adl)` maps output path → template key — wire new output files there. `engine.go` executes templates with Sprig + custom case-conversion funcs.
- **Never hand-edit generated code.** `internal/schema/types.go` (go-jsonschema output, DO NOT EDIT header) and `internal/schema/schema.json` (vendored upstream). `ADL_SCHEMA_VERSION` in `Taskfile.yml` is the single source of truth; `task verify-schema` fails CI on drift. Bump: `task fetch-schema && task generate-types`.
- **Version pins live in `internal/vendor/vendor.go`** (SDK deps, toolchains, GitHub Actions, semantic-release). Templates read them via `{{ pin "<group>" "<name>" }}`; an unknown key fails generation. Bump there, never in a `.tmpl`.
- **`.adl-ignore` protects user code.** The generator writes gitignore-style globs for TODO-placeholder files (tool implementations, custom services, bare skills); re-generate (even `--overwrite`) skips matched paths. Add new user-owned files to it.
- **Flag/manifest reconciliation is OR semantics** — a CLI flag (`--ci`, `--cd`, `--deployment`, …) wins over the manifest field.
- **Tools vs skills.** `spec.tools[]` → one generated file per tool (`tools/<id>.go`); `spec.skills[]` → markdown playbooks resolved from a registry/GitHub (`source:`) or scaffolded locally (`bare: true`). Five reserved tool IDs (`read`/`bash`/`write`/`edit`/`fetch`) render from `languages/<lang>/builtin/`.
- **Telemetry** (`spec.telemetry.enabled`, manifest-only): Go and TypeScript only; Rust ignores it.

## Conventions & gotchas

- Go 1.26.x; Flox sandbox (`.flox/env/manifest.toml`) pins `go`, `golangci-lint`, `go-task`, `prettier`, `markdownlint-cli` — `flox activate` to enter.
- Tabs in Go, 2-space in YAML/JSON/Markdown.
- Conventional commits (`feat:`, `fix:`, …) — semantic-release derives versions from them.
- **`examples/` is the regression suite.** When adding a feature, add/update an example and wire it into both lists (`examples:test` + `examples:generate`) in `Taskfile.yml`.
- Skills resolution hits the network (registry/GitHub, cached under `~/.adl/skills-cache`); `--offline` skips all network access.
- `internal/schema/validator.go::checkLegacySpecFields` rejects pre-orchestrators manifest shapes with migration hints — keep it in sync whenever the schema shape changes (JSON Schema `additionalProperties:true` would otherwise silently drop legacy fields).
