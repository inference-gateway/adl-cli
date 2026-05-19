package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExpandShorthand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single segment defaults to inference-gateway/skills@main", "skill-creator", "https://github.com/inference-gateway/skills/tree/main/skills/skill-creator"},
		{"single segment with @tag", "skill-creator@v1.0", "https://github.com/inference-gateway/skills/tree/v1.0/skills/skill-creator"},
		{"three segments pin owner/repo/skill", "anthropics/skills/pdf", "https://github.com/anthropics/skills/tree/main/skills/pdf"},
		{"three segments with @tag", "anthropics/skills/pdf@main", "https://github.com/anthropics/skills/tree/main/skills/pdf"},
		{"three segments with @sha", "anthropics/skills/pdf@abc123", "https://github.com/anthropics/skills/tree/abc123/skills/pdf"},
		{"https URL unchanged", "https://github.com/anthropics/skills/tree/main/skills/pdf", "https://github.com/anthropics/skills/tree/main/skills/pdf"},
		{"http URL unchanged", "http://example.com/x", "http://example.com/x"},
		{"empty unchanged", "", ""},
		{"two segments unchanged (no implicit repo=skills)", "anthropics/pdf", "anthropics/pdf"},
		{"four segments unchanged", "a/b/c/d", "a/b/c/d"},
		{"leading/trailing slashes trimmed", "/skill/", "https://github.com/inference-gateway/skills/tree/main/skills/skill"},
		{"empty middle segment unchanged", "a//b/c", "a//b/c"},
		{"empty tag unchanged", "skill@", "skill@"},
		{"empty body before @ unchanged", "@v1", "@v1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandShorthand(tt.input); got != tt.want {
				t.Errorf("ExpandShorthand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseGitHubTreeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *GitHubLocation
		wantErr string
	}{
		{
			name:  "happy path",
			input: "https://github.com/inference-gateway/skills/tree/main/skills/skill-creator",
			want:  &GitHubLocation{Owner: "inference-gateway", Repo: "skills", Ref: "main", Path: "skills/skill-creator"},
		},
		{
			name:    "blob URL rejected",
			input:   "https://github.com/foo/bar/blob/main/README.md",
			wantErr: "/blob/",
		},
		{
			name:    "non-github host rejected",
			input:   "https://gitlab.com/foo/bar/tree/main/skills/x",
			wantErr: "github.com",
		},
		{
			name:    "missing path segment rejected",
			input:   "https://github.com/foo/bar/tree/main",
			wantErr: "must be of the form",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGitHubTreeURL(tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if *got != *tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestInstaller_Fetch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/skills/git/trees/main", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(treeResponse{
			Tree: []treeEntry{
				{Path: "skills/foo/SKILL.md", Type: "blob"},
				{Path: "skills/foo/scripts/a.sh", Type: "blob"},
				{Path: "skills/foo/scripts", Type: "tree"},
				{Path: "skills/bar/SKILL.md", Type: "blob"},
			},
		})
	})
	mux.HandleFunc("/acme/skills/main/skills/foo/SKILL.md", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("skill body\n"))
	})
	mux.HandleFunc("/acme/skills/main/skills/foo/scripts/a.sh", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("#!/bin/sh\n"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	inst := &Installer{Client: srv.Client(), APIBase: srv.URL, RawBase: srv.URL}
	files, err := inst.Fetch(context.Background(), &GitHubLocation{Owner: "acme", Repo: "skills", Ref: "main", Path: "skills/foo"})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}
	if string(files["SKILL.md"]) != "skill body\n" {
		t.Errorf("SKILL.md mismatch: %q", files["SKILL.md"])
	}
	if string(files["scripts/a.sh"]) != "#!/bin/sh\n" {
		t.Errorf("scripts/a.sh mismatch: %q", files["scripts/a.sh"])
	}
}

func TestInstaller_FetchEmptyPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/skills/git/trees/main", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(treeResponse{
			Tree: []treeEntry{
				{Path: "skills/other/SKILL.md", Type: "blob"},
			},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	inst := &Installer{Client: srv.Client(), APIBase: srv.URL, RawBase: srv.URL}
	_, err := inst.Fetch(context.Background(), &GitHubLocation{Owner: "acme", Repo: "skills", Ref: "main", Path: "skills/missing"})
	if err == nil || !strings.Contains(err.Error(), "no files found") {
		t.Fatalf("expected no-files error, got %v", err)
	}
}
