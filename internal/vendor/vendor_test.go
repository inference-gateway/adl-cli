package vendor

import (
	"strings"
	"testing"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantName string
		wantVer  string
		wantErr  string
	}{
		{
			name:     "go module path",
			raw:      "github.com/stretchr/testify@v1.10.0",
			wantName: "github.com/stretchr/testify",
			wantVer:  "v1.10.0",
		},
		{
			name:     "npm scoped package keeps single @ on its name and splits on the version separator",
			raw:      "@types/node@20.11.0",
			wantName: "@types/node",
			wantVer:  "20.11.0",
		},
		{
			name:     "rust crate plain version",
			raw:      "tokio@1.36.0",
			wantName: "tokio",
			wantVer:  "1.36.0",
		},
		{
			name:    "missing version",
			raw:     "tokio@",
			wantErr: "expected '<package>@<version>'",
		},
		{
			name:    "missing name",
			raw:     "@v1.0.0",
			wantErr: "expected '<package>@<version>'",
		},
		{
			name:    "no separator",
			raw:     "tokio",
			wantErr: "expected '<package>@<version>'",
		},
		{
			name:    "internal whitespace rejected",
			raw:     "github.com/foo@v 1.0",
			wantErr: "must not contain whitespace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Parse(%q) succeeded; wanted error containing %q", tt.raw, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Parse(%q) error %v; wanted to contain %q", tt.raw, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tt.raw, err)
			}
			if got.Name != tt.wantName || got.Version != tt.wantVer {
				t.Fatalf("Parse(%q) = {%q %q}; want {%q %q}",
					tt.raw, got.Name, got.Version, tt.wantName, tt.wantVer)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	builtins := map[string]string{"github.com/inference-gateway/adk": "v0.18.4"}

	t.Run("empty input yields nothing", func(t *testing.T) {
		entries, conflicts, err := Resolve(nil, builtins, "deps")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entries != nil || conflicts != nil {
			t.Fatalf("expected nil/nil, got entries=%v conflicts=%v", entries, conflicts)
		}
	})

	t.Run("dedup against built-ins drops conflict and records it", func(t *testing.T) {
		entries, conflicts, err := Resolve(
			[]string{
				"github.com/stretchr/testify@v1.10.0",
				"github.com/inference-gateway/adk@v0.0.1",
			},
			builtins,
			"deps",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 1 || entries[0].Name != "github.com/stretchr/testify" {
			t.Fatalf("expected only testify to survive, got %+v", entries)
		}
		if len(conflicts) != 1 || conflicts[0].Entry.Name != "github.com/inference-gateway/adk" {
			t.Fatalf("expected adk conflict, got %+v", conflicts)
		}
		if conflicts[0].DepGroup != "deps" {
			t.Fatalf("conflict.DepGroup = %q, want %q", conflicts[0].DepGroup, "deps")
		}
	})

	t.Run("internal duplicates are collapsed first-wins", func(t *testing.T) {
		entries, _, err := Resolve(
			[]string{
				"github.com/stretchr/testify@v1.10.0",
				"github.com/stretchr/testify@v1.9.0",
			},
			builtins,
			"deps",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 1 || entries[0].Version != "v1.10.0" {
			t.Fatalf("expected first-wins dedup, got %+v", entries)
		}
	})

	t.Run("results are sorted by Name", func(t *testing.T) {
		entries, _, err := Resolve(
			[]string{
				"z-pkg@v1.0.0",
				"a-pkg@v1.0.0",
				"m-pkg@v1.0.0",
			},
			builtins,
			"deps",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := []string{entries[0].Name, entries[1].Name, entries[2].Name}
		want := []string{"a-pkg", "m-pkg", "z-pkg"}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("sort mismatch at %d: got %v, want %v", i, got, want)
			}
		}
	})

	t.Run("malformed entry surfaces an error", func(t *testing.T) {
		_, _, err := Resolve([]string{"not-a-valid-entry"}, builtins, "deps")
		if err == nil {
			t.Fatalf("expected error for malformed entry")
		}
	})
}

func goADLWithVendor(deps, devdeps []string) *schema.ADL {
	return &schema.ADL{
		Spec: schema.Spec{
			Language: schema.Language{
				Go: &schema.GoConfig{
					Module:  "example.com/agent",
					Version: "1.26.2",
					Vendor: &schema.VendorConfig{
						Deps:    deps,
						Devdeps: devdeps,
					},
				},
			},
		},
	}
}

func rustADLWithVendor(deps, devdeps []string) *schema.ADL {
	return &schema.ADL{
		Spec: schema.Spec{
			Language: schema.Language{
				Rust: &schema.RustConfig{
					PackageName: "agent",
					Version:     "1.94.1",
					Edition:     "2024",
					Vendor: &schema.VendorConfig{
						Deps:    deps,
						Devdeps: devdeps,
					},
				},
			},
		},
	}
}

func TestResolveADL_NoVendor(t *testing.T) {
	adl := &schema.ADL{
		Spec: schema.Spec{
			Language: schema.Language{
				Go: &schema.GoConfig{Module: "example.com/agent", Version: "1.26.2"},
			},
		},
	}
	view, err := ResolveADL(adl)
	if err != nil {
		t.Fatalf("ResolveADL: %v", err)
	}
	if len(view.GoRequires) != 0 || len(view.CargoDeps) != 0 || len(view.Conflicts) != 0 {
		t.Fatalf("expected empty View, got %+v", view)
	}
}

func TestResolveADL_GoSplitsDepsAndToolDevDeps(t *testing.T) {
	adl := goADLWithVendor(
		[]string{"github.com/stretchr/testify@v1.10.0"},
		[]string{"golang.org/x/tools/cmd/stringer@v0.20.0"},
	)
	view, err := ResolveADL(adl)
	if err != nil {
		t.Fatalf("ResolveADL: %v", err)
	}
	if len(view.GoRequires) != 1 || view.GoRequires[0].Name != "github.com/stretchr/testify" {
		t.Fatalf("expected testify in GoRequires, got %+v", view.GoRequires)
	}
	if len(view.GoTools) != 1 || view.GoTools[0].Name != "golang.org/x/tools/cmd/stringer" {
		t.Fatalf("expected stringer in GoTools, got %+v", view.GoTools)
	}
}

func TestResolveADL_GoToolDevDepDedupedAgainstDeps(t *testing.T) {
	adl := goADLWithVendor(
		[]string{"github.com/golang/mock/mockgen@v1.6.0"},
		[]string{"github.com/golang/mock/mockgen@v1.5.0"},
	)
	view, err := ResolveADL(adl)
	if err != nil {
		t.Fatalf("ResolveADL: %v", err)
	}
	if len(view.GoRequires) != 1 || view.GoRequires[0].Version != "v1.6.0" {
		t.Fatalf("expected deps to win at v1.6.0, got %+v", view.GoRequires)
	}
	if len(view.GoTools) != 0 {
		t.Fatalf("expected duplicate tool entry to be dropped, got %+v", view.GoTools)
	}
	if len(view.Conflicts) != 1 || view.Conflicts[0].DepGroup != "devdeps" {
		t.Fatalf("expected single devdeps conflict, got %+v", view.Conflicts)
	}
}

func TestResolveADL_GoBuiltinConflictDropsAndWarns(t *testing.T) {
	adl := goADLWithVendor(
		[]string{"github.com/inference-gateway/adk@v0.0.1"},
		nil,
	)
	view, err := ResolveADL(adl)
	if err != nil {
		t.Fatalf("ResolveADL: %v", err)
	}
	if len(view.GoRequires) != 0 {
		t.Fatalf("expected built-in conflict to be dropped, got %+v", view.GoRequires)
	}
	if len(view.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %+v", view.Conflicts)
	}
}

func TestResolveADL_RustDevDepsDedupedAgainstDeps(t *testing.T) {
	adl := rustADLWithVendor(
		[]string{"mockall@0.12.0"},
		[]string{"mockall@0.12.0", "pretty_assertions@1.4.0"},
	)
	view, err := ResolveADL(adl)
	if err != nil {
		t.Fatalf("ResolveADL: %v", err)
	}
	if len(view.CargoDeps) != 1 || view.CargoDeps[0].Name != "mockall" {
		t.Fatalf("expected mockall in deps, got %+v", view.CargoDeps)
	}
	if len(view.CargoDevDeps) != 1 || view.CargoDevDeps[0].Name != "pretty_assertions" {
		t.Fatalf("expected pretty_assertions in devdeps with mockall deduped, got %+v", view.CargoDevDeps)
	}
}

func TestResolveADL_RustBuiltinConflictDropped(t *testing.T) {
	adl := rustADLWithVendor(
		[]string{"tokio@0.1.0"},
		[]string{"tempfile@2"},
	)
	view, err := ResolveADL(adl)
	if err != nil {
		t.Fatalf("ResolveADL: %v", err)
	}
	if len(view.CargoDeps) != 0 {
		t.Fatalf("expected tokio to be dropped, got %+v", view.CargoDeps)
	}
	if len(view.CargoDevDeps) != 0 {
		t.Fatalf("expected tempfile to be dropped, got %+v", view.CargoDevDeps)
	}
	if len(view.Conflicts) != 2 {
		t.Fatalf("expected 2 conflicts, got %+v", view.Conflicts)
	}
}

func TestResolveADL_RustRuntimeBuiltinAlsoBlocksDevDeps(t *testing.T) {
	adl := rustADLWithVendor(nil, []string{"tokio@0.1.0"})
	view, err := ResolveADL(adl)
	if err != nil {
		t.Fatalf("ResolveADL: %v", err)
	}
	if len(view.CargoDevDeps) != 0 {
		t.Fatalf("expected tokio in devdeps to be dropped, got %+v", view.CargoDevDeps)
	}
	if len(view.Conflicts) != 1 || view.Conflicts[0].DepGroup != "devdeps" {
		t.Fatalf("expected single devdeps conflict, got %+v", view.Conflicts)
	}
}

func TestResolveADL_MalformedEntrySurfacesError(t *testing.T) {
	adl := goADLWithVendor([]string{"not-valid"}, nil)
	_, err := ResolveADL(adl)
	if err == nil {
		t.Fatal("expected error for malformed entry")
	}
	if !strings.Contains(err.Error(), "spec.language.go.vendor") {
		t.Fatalf("error should point at offending key, got %v", err)
	}
}
