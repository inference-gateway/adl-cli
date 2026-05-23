package sandbox

import (
	"reflect"
	"testing"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Entry
		wantErr bool
	}{
		{
			name: "simple",
			raw:  "deno@2.1.4",
			want: Entry{Name: "deno", Version: "2.1.4"},
		},
		{
			name: "kubectl with x.y.z",
			raw:  "kubectl@1.31.0",
			want: Entry{Name: "kubectl", Version: "1.31.0"},
		},
		{
			name:    "missing version",
			raw:     "deno@",
			wantErr: true,
		},
		{
			name:    "missing name",
			raw:     "@1.0.0",
			wantErr: true,
		},
		{
			name:    "no at sign",
			raw:     "deno-2.1.4",
			wantErr: true,
		},
		{
			name:    "embedded whitespace in name",
			raw:     "foo bar@2.1.4",
			wantErr: true,
		},
		{
			name:    "embedded whitespace in version",
			raw:     "foo@2.1 4",
			wantErr: true,
		},
		{
			name: "version with at sign uses last",
			raw:  "@org/pkg@1.0.0",
			want: Entry{Name: "@org/pkg", Version: "1.0.0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Parse(%q) err=%v wantErr=%v", tc.raw, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Parse(%q) = %#v, want %#v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestResolve_NilOrEmpty(t *testing.T) {
	cases := []struct {
		name string
		adl  *schema.ADL
	}{
		{name: "nil adl", adl: nil},
		{name: "no development", adl: &schema.ADL{}},
		{
			name: "development without deps",
			adl: &schema.ADL{
				Spec: schema.Spec{Development: &schema.DevelopmentConfig{}},
			},
		},
		{
			name: "empty deps slice",
			adl: &schema.ADL{
				Spec: schema.Spec{Development: &schema.DevelopmentConfig{Deps: []string{}}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := Resolve(tc.adl)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if v.HasDeps() {
				t.Fatalf("expected empty deps, got %#v", v.Deps)
			}
			if len(v.FloxConflicts) != 0 || len(v.DevContainerConflicts) != 0 {
				t.Fatalf("expected no conflicts, got flox=%v devcontainer=%v", v.FloxConflicts, v.DevContainerConflicts)
			}
		})
	}
}

func TestResolve_DedupesAndSorts(t *testing.T) {
	adl := &schema.ADL{
		Spec: schema.Spec{
			Development: &schema.DevelopmentConfig{
				Deps: []string{
					"terraform@1.9.5",
					"deno@2.1.4",
					"kubectl@1.31.0",
					"deno@2.0.0", // duplicate package name - first wins
				},
			},
		},
	}

	v, err := Resolve(adl)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	want := []Entry{
		{Name: "deno", Version: "2.1.4"},
		{Name: "kubectl", Version: "1.31.0"},
		{Name: "terraform", Version: "1.9.5"},
	}
	if !reflect.DeepEqual(v.Deps, want) {
		t.Fatalf("Resolve deps = %#v, want %#v", v.Deps, want)
	}
}

func TestResolve_FloxConflict(t *testing.T) {
	adl := &schema.ADL{
		Spec: schema.Spec{
			Development: &schema.DevelopmentConfig{
				Deps: []string{
					"git@2.53.0",   // clashes with built-in
					"deno@2.1.4",   // safe
					"go-task@3.50", // clashes with built-in
				},
			},
		},
	}

	v, err := Resolve(adl)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if len(v.Deps) != 3 {
		t.Fatalf("expected all 3 deps preserved (conflicts warn but don't drop), got %d", len(v.Deps))
	}

	conflictNames := make(map[string]bool, len(v.FloxConflicts))
	for _, c := range v.FloxConflicts {
		conflictNames[c.Entry.Name] = true
	}
	if !conflictNames["git"] || !conflictNames["go-task"] {
		t.Fatalf("expected git and go-task to flag flox conflicts, got %v", conflictNames)
	}
	if conflictNames["deno"] {
		t.Fatalf("did not expect deno to flag a flox conflict")
	}
}

func TestResolve_DevContainerConflict(t *testing.T) {
	adl := &schema.ADL{
		Spec: schema.Spec{
			Development: &schema.DevelopmentConfig{
				Deps: []string{"git@2.53.0", "kubectl@1.31.0"},
			},
		},
	}

	v, err := Resolve(adl)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if len(v.DevContainerConflicts) != 1 {
		t.Fatalf("expected exactly one devcontainer conflict (git), got %d", len(v.DevContainerConflicts))
	}
	if v.DevContainerConflicts[0].Entry.Name != "git" {
		t.Fatalf("expected git devcontainer conflict, got %q", v.DevContainerConflicts[0].Entry.Name)
	}
}

func TestResolve_MalformedEntry(t *testing.T) {
	adl := &schema.ADL{
		Spec: schema.Spec{
			Development: &schema.DevelopmentConfig{
				Deps: []string{"valid@1.0", "broken-without-at"},
			},
		},
	}

	if _, err := Resolve(adl); err == nil {
		t.Fatalf("expected error for malformed entry, got nil")
	}
}
