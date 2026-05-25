# Repository Guidelines

## Project Structure & Module Organization

This repository contains the `adl` Go CLI for generating A2A agent projects from ADL YAML files. `main.go` wires the binary, and CLI commands live in `cmd/` (`init`, `generate`, `validate`). Shared code is under `internal/`: `generator/` writes projects, `templates/` owns embedded templates and registries, `schema/` validates ADL files, `registry/` resolves remote assets, and `sandbox/` manages sandbox dependencies. Example manifests are in `examples/`; generated smoke-test output goes to `test-output/`; binaries go to `bin/`.

## Build, Test, and Development Commands

- `task build`: compile `bin/adl` with the version from `Taskfile.yml`.
- `task dev -- <args>`: build and run the local CLI, for example `task dev -- validate examples/go-agent.yaml`.
- `task test`: run `go test -v ./...`.
- `task test:coverage`: run tests with package coverage.
- `task fmt`: run `go fmt ./...`.
- `task lint`: run `golangci-lint run`.
- `task examples:test`: validate every example manifest.
- `task examples:generate`: regenerate examples under `test-output/`.
- `task ci`: run the main local CI sequence.

## Coding Style & Naming Conventions

Follow standard Go conventions: tabs from `gofmt`, short package names, documented exported identifiers, and early returns for errors. Keep command code in `cmd/` thin and move reusable behavior into `internal/`. Template files use `.tmpl` suffixes and should follow existing registry patterns. When schema files change, regenerate Go types with `task generate-types`.

## Testing Guidelines

Tests use Go's standard `testing` package and live in `*_test.go` files next to the code under test. Prefer table-driven tests for validators, generators, and template registries. Run `task test` before submitting changes; use `task examples:test` or `task examples:generate` when modifying templates, schema behavior, or examples. Add regression tests for fixes that affect CLI output, validation, or generated files.

## Commit & Pull Request Guidelines

Recent history uses Conventional Commit-style subjects such as `chore: Generate CLAUDE.md file` and `chore(deps): Add codex`. Use `type: concise imperative summary`, optionally with a scope (`fix(schema): ...`, `docs: ...`, `test: ...`).

Pull requests should include a clear description, linked issue when applicable, and testing notes. Update docs and examples when behavior changes. For template or CLI changes, include representative command output or generated-file notes.

## Security & Configuration Tips

Do not commit secrets, local credentials, or generated environment files. Keep example configuration safe and generic. Schema updates are pinned in `Taskfile.yml`; use `task fetch-schema` only when intentionally syncing to the configured upstream version, then verify with `task verify-schema`.
