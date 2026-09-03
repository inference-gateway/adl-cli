// Package vendor resolves the optional `spec.language.<lang>.vendor.deps`
// and `spec.language.<lang>.vendor.devdeps` fields on an ADL manifest into
// per-language dependency lists that the templates can render directly.
//
// The schema validates the raw entries up front (each must match
// `^\S+@\S+$`), but here we re-split them into name/version, dedupe each
// list against the generator's built-in dependency set for that language,
// and sort the result so the output is deterministic.
//
// Conflict policy: built-in dependencies always win. If a user lists
// `github.com/inference-gateway/adk@v0.0.1` in `vendor.deps`, it is
// dropped and reported via Resolve's `Dropped` slice so the caller can
// surface a warning. This prevents accidental downgrades of the core
// runtime SDK or related plumbing the generator depends on.
package vendor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

// Entry is a parsed `<package>@<version>` tuple.
type Entry struct {
	Name    string
	Version string
}

// Conflict describes a vendor entry that was dropped because its package
// name collided with one of the generator's built-in dependencies.
type Conflict struct {
	Entry    Entry
	Builtin  string // the built-in package name the conflict was against
	DepGroup string // "deps" or "devdeps"
}

// Parse splits a `<package>@<version>` literal into an Entry. The schema
// validates the pattern up front, but we still defend against malformed
// input here for callers that bypass the validator (e.g. generator tests
// that construct an ADL by hand).
func Parse(raw string) (Entry, error) {
	idx := strings.LastIndex(raw, "@")
	if idx <= 0 || idx == len(raw)-1 {
		return Entry{}, fmt.Errorf("invalid vendor entry %q: expected '<package>@<version>'", raw)
	}
	name := strings.TrimSpace(raw[:idx])
	version := strings.TrimSpace(raw[idx+1:])
	if name == "" || version == "" {
		return Entry{}, fmt.Errorf("invalid vendor entry %q: package and version must be non-empty", raw)
	}
	if strings.ContainsAny(name, " \t") || strings.ContainsAny(version, " \t") {
		return Entry{}, fmt.Errorf("invalid vendor entry %q: name and version must not contain whitespace", raw)
	}
	return Entry{Name: name, Version: version}, nil
}

// Resolve parses each raw entry, drops duplicates and built-in collisions,
// and returns the surviving entries sorted by Name. If two vendor entries
// declare the same package, the first one wins.
//
// `depGroup` is propagated into any Conflict reported by the second
// return value so callers can mention "deps" vs "devdeps" in warnings.
func Resolve(raws []string, builtins map[string]string, depGroup string) ([]Entry, []Conflict, error) {
	if len(raws) == 0 {
		return nil, nil, nil
	}

	var kept []Entry
	var conflicts []Conflict
	seen := make(map[string]struct{}, len(raws))

	for _, raw := range raws {
		entry, err := Parse(raw)
		if err != nil {
			return nil, nil, err
		}
		if _, dup := seen[entry.Name]; dup {
			continue
		}
		if builtinVersion, clash := builtins[entry.Name]; clash {
			conflicts = append(conflicts, Conflict{
				Entry:    entry,
				Builtin:  fmt.Sprintf("%s@%s", entry.Name, builtinVersion),
				DepGroup: depGroup,
			})
			continue
		}
		seen[entry.Name] = struct{}{}
		kept = append(kept, entry)
	}

	sort.Slice(kept, func(i, j int) bool { return kept[i].Name < kept[j].Name })
	return kept, conflicts, nil
}

// GoBuiltins enumerates the modules that the generator always writes to
// the generated `go.mod` require block (see
// `internal/templates/languages/go/go.mod.tmpl`). Vendor entries that
// match one of these are dropped to keep the generator in charge of its
// runtime SDK pins.
//
// Keep this in sync with the go.mod template; the generator's TestVendor*
// tests will fail loudly if a built-in is added there without being
// mirrored here.
var GoBuiltins = map[string]string{
	"github.com/inference-gateway/adk":  "v0.26.4",
	"github.com/sethvargo/go-envconfig": "v1.4.3",
	"github.com/spf13/cobra":            "v1.10.2",
	"go.uber.org/zap":                   "v1.28.0",
	"gopkg.in/yaml.v3":                  "v3.0.1",
}

// CargoBuiltinDeps enumerates the crates the Cargo.toml template always
// writes to the `[dependencies]` section. Some are conditional on
// features (e.g. `reqwest` when the `fetch` built-in is enabled); list
// them all so users can't shadow them regardless of which features they
// activate.
var CargoBuiltinDeps = map[string]string{
	"inference-gateway-adk": "0.11.2",
	"inference-gateway-sdk": "0.19.0",
	"tokio":                 "1",
	"tracing":               "0.1",
	"tracing-subscriber":    "0.3",
	"clap":                  "4",
	"serde":                 "1",
	"serde_json":            "1",
	"serde_yaml":            "0.9",
	"anyhow":                "1",
	"async-trait":           "0.1",
	"uuid":                  "1",
	"chrono":                "0.4",
	"dotenvy":               "0.15.7",
	"envy":                  "0.4.2",
	"reqwest":               "0.12",
}

// CargoBuiltinDevDeps mirrors CargoBuiltinDeps for the `[dev-dependencies]`
// section. Currently only `tempfile` is emitted (and only when a built-in
// tool is enabled), but listing it here lets us catch shadowing
// regardless of features.
var CargoBuiltinDevDeps = map[string]string{
	"tempfile": "3",
}

// GoTelemetryDeps are the extra go.mod requires emitted only when
// `spec.telemetry.enabled` is set. They still count as built-ins for
// vendor conflict checks so users can't downgrade them.
var GoTelemetryDeps = map[string]string{
	"go.opentelemetry.io/otel":       "v1.46.0",
	"go.opentelemetry.io/otel/sdk":   "v1.46.0",
	"go.opentelemetry.io/otel/trace": "v1.46.0",
}

// NpmBuiltinDeps / NpmBuiltinDevDeps mirror the package.json template.
// `@inference-gateway/adl-cli` is intentionally absent: it tracks the CLI
// version at generation time rather than a static pin.
var NpmBuiltinDeps = map[string]string{
	"@inference-gateway/adk": "0.15.1",
}
var NpmBuiltinDevDeps = map[string]string{
	"@types/node": "^24.1.0",
	"prettier":    "^3.8.3",
	"tsx":         "^4.19.2",
	"typescript":  "^6.0.3",
}

// Tools pins the toolchain / sandbox package versions the generated Flox
// manifest, devcontainer and CI workflows install. Bare numbers: templates
// add their own `^` / `v` prefix as the target format requires.
var Tools = map[string]string{
	"flox-schema":   "1.13.0",
	"golangci-lint": "2.12.2",
	"go-task":       "3.48.0",
	"rust":          "1.94.1",
	"rust-analyzer": "2026-04-27",
	"nodejs":        "24.15.0",
	"pnpm":          "11.8.0",
	"git":           "2.53.0",
	"docker":        "29.5.1",
	"claude-code":   "2.1.201",
	"infer":         "0.154.0",
}

// Actions pins the GitHub Actions referenced by the generated workflows
// (`uses: <name>@<version>`).
var Actions = map[string]string{
	"actions/cache":                        "v5.0.5",
	"actions/checkout":                     "v7.0.1",
	"actions/create-github-app-token":      "v3.2.0",
	"actions/setup-go":                     "v7.0.0",
	"actions/setup-node":                   "v7.0.0",
	"anthropics/claude-code-action":        "v1.0.214",
	"arduino/setup-task":                   "v3.0.0",
	"azure/setup-kubectl":                  "v4.0.1",
	"docker/login-action":                  "v4.6.0",
	"docker/setup-buildx-action":           "v4.3.0",
	"docker/setup-qemu-action":             "v4.2.0",
	"golangci/golangci-lint-action":        "v9.3.0",
	"google-github-actions/auth":           "v2.1.13",
	"google-github-actions/run-gemini-cli": "v0.1.22",
	"google-github-actions/setup-gcloud":   "v3.0.1",
	"inference-gateway/infer-action":       "v0.51.2",
	"openai/codex-action":                  "v1.8",
	"oven-sh/setup-bun":                    "v2.2.0",
	"peter-evans/create-pull-request":      "v8.1.1",
}

// Release pins the semantic-release npm packages the generated CD
// workflow installs into a throwaway package.json.
var Release = map[string]string{
	"semantic-release":                           "25.0.5",
	"@semantic-release/commit-analyzer":          "13.0.1",
	"@semantic-release/release-notes-generator":  "14.1.1",
	"@semantic-release/changelog":                "6.0.3",
	"@semantic-release/exec":                     "7.1.0",
	"@semantic-release/git":                      "10.0.1",
	"@semantic-release/github":                   "12.0.9",
	"conventional-changelog-conventionalcommits": "10.2.0",
	"conventional-changelog-writer":              "^9.1.0",
}

// Pin returns the pinned version for `name` in `group`. It is exposed to
// templates as the `pin` func; an unknown group or name fails template
// execution instead of silently rendering an empty string.
func Pin(group, name string) (string, error) {
	groups := map[string]map[string]string{
		"go":           GoBuiltins,
		"go-telemetry": GoTelemetryDeps,
		"cargo":        CargoBuiltinDeps,
		"cargo-dev":    CargoBuiltinDevDeps,
		"npm":          NpmBuiltinDeps,
		"npm-dev":      NpmBuiltinDevDeps,
		"tool":         Tools,
		"action":       Actions,
		"release":      Release,
	}
	m, ok := groups[group]
	if !ok {
		return "", fmt.Errorf("pin: unknown group %q", group)
	}
	v, ok := m[name]
	if !ok {
		return "", fmt.Errorf("pin: no %q version pinned in group %q", name, group)
	}
	return v, nil
}

// View is the resolved vendor data injected into the template Context.
// Each field is a sorted, deduped slice ready to be rendered by the
// language-specific template.
type View struct {
	// GoRequires holds Go runtime dependencies (`vendor.deps`) that the
	// `go.mod` template appends to the built-in `require` block. Test
	// libraries that are `import`-ed by `*_test.go` belong here too:
	// Go has no separate test-dependency notion.
	GoRequires []Entry

	// GoBuiltinEntries mirrors GoBuiltins as a sorted slice so the
	// go.mod template can iterate over it as a single source of truth.
	// When bumping a built-in version, update GoBuiltins (the map) and
	// the template picks it up automatically.
	GoBuiltinEntries []Entry

	// GoTools holds Go executable dev tools (`vendor.devdeps`). Each
	// entry is rendered both as a `// indirect` line in `require` (so
	// the module is downloadable) and as a bare package path inside the
	// `tool ( ... )` block introduced in Go 1.24. Users supply the full
	// tool package path (e.g. `golang.org/x/tools/cmd/stringer`); a
	// post-generation `go mod tidy` normalises the require entry to the
	// owning module root.
	GoTools []Entry

	// CargoDeps / CargoDevDeps map to the matching Cargo.toml sections.
	// CargoDevDeps is additionally deduped against CargoDeps so we never
	// emit the same crate in both sections.
	CargoDeps    []Entry
	CargoDevDeps []Entry

	// NpmDeps / NpmDevDeps map to package.json's dependencies /
	// devDependencies. NpmDevDeps is deduped against NpmDeps for the same
	// reason as Cargo.
	NpmDeps    []Entry
	NpmDevDeps []Entry

	// Conflicts collects every entry that was dropped because of a
	// built-in collision so the caller can surface warnings to the user.
	Conflicts []Conflict
}

// Resolve walks the language-specific vendor blocks on the ADL manifest
// and produces a View suitable for templating. Languages whose vendor
// block is absent contribute empty slices. Errors only surface for
// malformed entries (which should never happen post-validation, but we
// re-check here so generator tests can pass hand-built ADLs).
func ResolveADL(adl *schema.ADL) (View, error) {
	view := View{}
	if adl == nil {
		return view, nil
	}

	lang := adl.Spec.Language

	view.GoBuiltinEntries = goBuiltinEntries()

	if lang.Go != nil && lang.Go.Vendor != nil {
		goBuiltins := cloneMap(GoBuiltins)
		for k, v := range GoTelemetryDeps {
			goBuiltins[k] = v
		}
		deps, depConflicts, err := Resolve(lang.Go.Vendor.Deps, goBuiltins, "deps")
		if err != nil {
			return View{}, fmt.Errorf("spec.language.go.vendor.deps: %w", err)
		}
		view.GoRequires = deps
		view.Conflicts = append(view.Conflicts, depConflicts...)

		toolEffectiveBuiltins := goBuiltins
		for _, e := range deps {
			toolEffectiveBuiltins[e.Name] = e.Version
		}
		tools, toolConflicts, err := Resolve(lang.Go.Vendor.Devdeps, toolEffectiveBuiltins, "devdeps")
		if err != nil {
			return View{}, fmt.Errorf("spec.language.go.vendor.devdeps: %w", err)
		}
		view.GoTools = tools
		view.Conflicts = append(view.Conflicts, toolConflicts...)
	}

	if lang.Rust != nil && lang.Rust.Vendor != nil {
		deps, conflicts, err := Resolve(lang.Rust.Vendor.Deps, CargoBuiltinDeps, "deps")
		if err != nil {
			return View{}, fmt.Errorf("spec.language.rust.vendor.deps: %w", err)
		}
		view.CargoDeps = deps
		view.Conflicts = append(view.Conflicts, conflicts...)

		devEffectiveBuiltins := cloneMap(CargoBuiltinDevDeps)
		for k, v := range CargoBuiltinDeps {
			if _, set := devEffectiveBuiltins[k]; !set {
				devEffectiveBuiltins[k] = v
			}
		}
		for _, e := range deps {
			devEffectiveBuiltins[e.Name] = e.Version
		}
		devdeps, devConflicts, err := Resolve(lang.Rust.Vendor.Devdeps, devEffectiveBuiltins, "devdeps")
		if err != nil {
			return View{}, fmt.Errorf("spec.language.rust.vendor.devdeps: %w", err)
		}
		view.CargoDevDeps = devdeps
		view.Conflicts = append(view.Conflicts, devConflicts...)
	}

	if lang.TypeScript != nil && lang.TypeScript.Vendor != nil {
		deps, conflicts, err := Resolve(lang.TypeScript.Vendor.Deps, NpmBuiltinDeps, "deps")
		if err != nil {
			return View{}, fmt.Errorf("spec.language.typescript.vendor.deps: %w", err)
		}
		view.NpmDeps = deps
		view.Conflicts = append(view.Conflicts, conflicts...)

		devEffectiveBuiltins := cloneMap(NpmBuiltinDevDeps)
		for k, v := range NpmBuiltinDeps {
			if _, set := devEffectiveBuiltins[k]; !set {
				devEffectiveBuiltins[k] = v
			}
		}
		for _, e := range deps {
			devEffectiveBuiltins[e.Name] = e.Version
		}
		devdeps, devConflicts, err := Resolve(lang.TypeScript.Vendor.Devdeps, devEffectiveBuiltins, "devdeps")
		if err != nil {
			return View{}, fmt.Errorf("spec.language.typescript.vendor.devdeps: %w", err)
		}
		view.NpmDevDeps = devdeps
		view.Conflicts = append(view.Conflicts, devConflicts...)
	}

	return view, nil
}

func cloneMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// goBuiltinEntries converts the GoBuiltins map to a sorted slice of Entry
// so the go.mod template can iterate over it deterministically.
func goBuiltinEntries() []Entry {
	entries := make([]Entry, 0, len(GoBuiltins))
	for name, version := range GoBuiltins {
		entries = append(entries, Entry{Name: name, Version: version})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}
