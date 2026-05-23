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
	"github.com/inference-gateway/adk":  "v0.18.4",
	"github.com/sethvargo/go-envconfig": "v1.3.0",
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
	"inference-gateway-adk": "0.4.3",
	"inference-gateway-sdk": "0.13.3",
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

// NpmBuiltinDeps / NpmBuiltinDevDeps are placeholders for the TypeScript
// generator. There are no TS templates today (only `.gitkeep`), but the
// schema accepts vendor entries for TypeScript so we still parse and
// dedupe them when the generator eventually wires them through. Keep the
// maps empty until a `package.json` template lands.
var NpmBuiltinDeps = map[string]string{}
var NpmBuiltinDevDeps = map[string]string{}

// View is the resolved vendor data injected into the template Context.
// Each field is a sorted, deduped slice ready to be rendered by the
// language-specific template.
type View struct {
	// GoRequires is the merged deps+devdeps set for Go (go.mod has no
	// notion of a separate dev section; see README for context).
	GoRequires []Entry

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

	if lang.Go != nil && lang.Go.Vendor != nil {
		merged := mergeRaw(lang.Go.Vendor.Deps, lang.Go.Vendor.Devdeps)
		entries, conflicts, err := Resolve(merged, GoBuiltins, "deps+devdeps")
		if err != nil {
			return View{}, fmt.Errorf("spec.language.go.vendor: %w", err)
		}
		view.GoRequires = entries
		view.Conflicts = append(view.Conflicts, conflicts...)
	}

	if lang.Rust != nil && lang.Rust.Vendor != nil {
		deps, conflicts, err := Resolve(lang.Rust.Vendor.Deps, CargoBuiltinDeps, "deps")
		if err != nil {
			return View{}, fmt.Errorf("spec.language.rust.vendor.deps: %w", err)
		}
		view.CargoDeps = deps
		view.Conflicts = append(view.Conflicts, conflicts...)

		// When checking dev-dependencies, treat as built-in: real dev
		// built-ins (tempfile), the runtime built-ins (tokio, etc.), AND
		// any vendor entry already accepted for `[dependencies]`. Cargo
		// rejects a crate appearing in both sections of the manifest, so
		// devdeps must be deduped against every prior dep we plan to emit.
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

		// Same dedupe-against-everything-prior reasoning as the Rust
		// branch above: a package in `dependencies` cannot also live in
		// `devDependencies`.
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

func mergeRaw(a, b []string) []string {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := make([]string, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

func cloneMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
