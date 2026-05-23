// Package sandbox resolves the optional `spec.development.deps` field on
// an ADL manifest into a deterministic, deduped list of sandbox-level
// extra packages that the templates can render directly.
//
// `spec.development.deps` is the cross-cutting equivalent of
// `spec.language.<lang>.vendor.deps`: each entry is a `<package>@<version>`
// literal (validated by the JSON schema as `^\S+@\S+$`) that the user
// wants installed inside whichever sandbox backend they've enabled
// (Flox, devcontainer, or docker-compose dev image - the last is out of
// scope for now). Unlike the language-vendor blocks, the generator has
// no built-ins for the sandbox layer beyond the per-language toolchain
// already baked into each template (Go/Cargo/Node + git + docker +
// go-task), so the merge policy is straightforward additive: built-in
// template entries always render; user `deps` are appended on top.
// Duplicates inside the user list are kept once (first occurrence
// wins); collisions with the built-in template list are flagged via
// the Conflicts slice so the caller can surface a warning, but the
// user entry is still rendered (the template lays the built-in down at
// a fixed version, and the user-pinned entry replaces nothing - they
// would coexist in the manifest and Flox/devcontainer would resolve the
// later definition).
//
// We deliberately reuse the per-language `<package>@<version>` shape
// rather than inventing a backend-specific syntax (e.g. Nix attribute
// paths for Flox) because the schema is shared across consumers and
// each backend can translate the literal in its own template. Flox
// turns `<pkg>@<ver>` into `pkg.pkg-path = "pkg"` / `pkg.version = "ver"`;
// devcontainer renders `pkg=ver` apt-style entries.
package sandbox

import (
	"fmt"
	"sort"
	"strings"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

// Entry is a parsed `<package>@<version>` tuple for a sandbox-level
// extra dependency.
type Entry struct {
	Name    string
	Version string
}

// Conflict describes a user `deps` entry whose package name collides
// with one of the generator's built-in sandbox packages (the toolchain
// the per-language template always installs - `go`, `cargo`, `git`,
// `docker`, `go-task`, ...). The conflict is reported so the caller
// can warn, but the user entry is NOT dropped: the user explicitly
// asked for a specific version, and we honour that.
type Conflict struct {
	Entry   Entry
	Builtin string // the built-in package name the conflict was against
}

// Parse splits a `<package>@<version>` literal into an Entry. The schema
// validates the pattern up front, but we still defend against malformed
// input here for callers that bypass the validator (e.g. generator
// tests that construct an ADL by hand).
func Parse(raw string) (Entry, error) {
	idx := strings.LastIndex(raw, "@")
	if idx <= 0 || idx == len(raw)-1 {
		return Entry{}, fmt.Errorf("invalid sandbox dep %q: expected '<package>@<version>'", raw)
	}
	name := strings.TrimSpace(raw[:idx])
	version := strings.TrimSpace(raw[idx+1:])
	if name == "" || version == "" {
		return Entry{}, fmt.Errorf("invalid sandbox dep %q: package and version must be non-empty", raw)
	}
	if strings.ContainsAny(name, " \t") || strings.ContainsAny(version, " \t") {
		return Entry{}, fmt.Errorf("invalid sandbox dep %q: name and version must not contain whitespace", raw)
	}
	return Entry{Name: name, Version: version}, nil
}

// floxBuiltinPackages enumerates the packages that the Flox manifest
// template always installs (see
// `internal/templates/sandbox/flox/manifest.toml.tmpl`). Used to flag -
// but not drop - collisions with user `spec.development.deps`.
//
// Keep this in sync with the manifest template. The order doesn't
// matter; it's only used for membership lookup.
var floxBuiltinPackages = map[string]struct{}{
	"go":            {},
	"golangci-lint": {},
	"cargo":         {},
	"rustc":         {},
	"clippy":        {},
	"rustfmt":       {},
	"rust-analyzer": {},
	"nodejs_24":     {},
	"go-task":       {},
	"git":           {},
	"docker":        {},
	"claude-code":   {},
	"adl":           {},
}

// devcontainerBuiltinPackages enumerates the well-known features the
// devcontainer template always provisions. The Devcontainer Features
// model is feature-id-keyed (not package-name-keyed), so collisions are
// rare; we only flag the most common shorthand matches here so users
// who type `git@2.53.0` get a helpful warning that git is already
// installed by the base image.
//
// Keep this in sync with `internal/templates/sandbox/devcontainer/devcontainer.json.tmpl`.
var devcontainerBuiltinPackages = map[string]struct{}{
	"git":              {},
	"docker":           {},
	"docker-in-docker": {},
}

// View is the resolved sandbox deps data injected into the template
// Context. The slice is deduped (first occurrence wins) and sorted by
// Name so output is deterministic across runs.
type View struct {
	// Deps holds the user-declared extra sandbox packages, parsed and
	// sorted. Always non-nil for the templates' `range` clause; empty
	// when `spec.development.deps` is absent.
	Deps []Entry

	// FloxConflicts collects entries whose package name collides with
	// the Flox manifest template's built-in package list. Each entry
	// is still rendered (user entries win on version conflict); the
	// generator surfaces a warning so the user knows the template
	// pin and the user pin will both appear in the generated
	// manifest.toml.
	FloxConflicts []Conflict

	// DevContainerConflicts mirrors FloxConflicts for the devcontainer
	// backend. Only the most common shorthand collisions (git, docker)
	// are flagged - the Devcontainer Features model is feature-id-keyed
	// so most user entries won't clash.
	DevContainerConflicts []Conflict
}

// HasDeps reports whether the view carries any user-declared deps.
// Templates use this to decide whether to render the optional block.
func (v View) HasDeps() bool {
	return len(v.Deps) > 0
}

// Resolve walks `adl.Spec.Development.Deps`, parses each entry,
// dedupes by package name (first wins), sorts the result, and flags
// collisions with the per-backend built-in package lists. Returns an
// empty View (no error) when the ADL or its development block is nil
// or when `deps` is empty.
func Resolve(adl *schema.ADL) (View, error) {
	view := View{}
	if adl == nil || adl.Spec.Development == nil || len(adl.Spec.Development.Deps) == 0 {
		return view, nil
	}

	var kept []Entry
	seen := make(map[string]struct{}, len(adl.Spec.Development.Deps))

	for _, raw := range adl.Spec.Development.Deps {
		entry, err := Parse(raw)
		if err != nil {
			return View{}, fmt.Errorf("spec.development.deps: %w", err)
		}
		if _, dup := seen[entry.Name]; dup {
			continue
		}
		seen[entry.Name] = struct{}{}
		kept = append(kept, entry)
	}

	sort.Slice(kept, func(i, j int) bool { return kept[i].Name < kept[j].Name })
	view.Deps = kept

	for _, e := range kept {
		if _, clash := floxBuiltinPackages[e.Name]; clash {
			view.FloxConflicts = append(view.FloxConflicts, Conflict{Entry: e, Builtin: e.Name})
		}
		if _, clash := devcontainerBuiltinPackages[e.Name]; clash {
			view.DevContainerConflicts = append(view.DevContainerConflicts, Conflict{Entry: e, Builtin: e.Name})
		}
	}

	return view, nil
}
