package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

func newTestResolver(t *testing.T, handler http.HandlerFunc) (*Resolver, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	cache, err := NewCache(t.TempDir())
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	return &Resolver{
		Client:    NewClient(srv.URL + "/skills/"),
		Installer: NewInstaller(),
		Cache:     cache,
	}, srv.Close
}

func TestResolver_Bare(t *testing.T) {
	resolver, closer := newTestResolver(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("bare skill should not hit the registry, got request: %s", r.URL.Path)
	})
	defer closer()

	resolved, err := resolver.Resolve(context.Background(), schema.Skill{
		ID:          "company-policy",
		Bare:        true,
		Name:        "company-policy",
		Description: "Internal rules",
		Tags:        []string{"policy"},
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !resolved.Bare {
		t.Error("expected Bare=true")
	}
	if resolved.Name != "company-policy" || resolved.Description != "Internal rules" {
		t.Errorf("unexpected resolved metadata: %+v", resolved)
	}
	if len(resolved.Files) != 0 {
		t.Errorf("bare skill should not have Files, got %d entries", len(resolved.Files))
	}
}

func TestResolver_BareMissingMetadata(t *testing.T) {
	resolver, closer := newTestResolver(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not hit registry")
	})
	defer closer()

	_, err := resolver.Resolve(context.Background(), schema.Skill{
		ID:   "company-policy",
		Bare: true,
	})
	if err == nil || !strings.Contains(err.Error(), "requires both name and description") {
		t.Fatalf("expected bare-metadata error, got %v", err)
	}
}

func TestResolver_FetchAndCache(t *testing.T) {
	var calls int
	resolver, closer := newTestResolver(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte("---\nname: data-analysis\ndescription: from registry\ntags: [analytics]\n---\nbody\n"))
	})
	defer closer()

	skill := schema.Skill{ID: "data-analysis"}

	resolved, err := resolver.Resolve(context.Background(), skill)
	if err != nil {
		t.Fatalf("first Resolve: %v", err)
	}
	if resolved.Name != "data-analysis" || resolved.Description != "from registry" {
		t.Errorf("unexpected metadata: %+v", resolved)
	}
	if len(resolved.Tags) != 1 || resolved.Tags[0] != "analytics" {
		t.Errorf("unexpected tags: %v", resolved.Tags)
	}
	if _, ok := resolved.Files["SKILL.md"]; !ok {
		t.Errorf("expected resolved.Files to contain SKILL.md, got keys: %v", keysOf(resolved.Files))
	}
	if calls != 1 {
		t.Errorf("expected 1 registry call, got %d", calls)
	}

	if _, err := resolver.Resolve(context.Background(), skill); err != nil {
		t.Fatalf("second Resolve (cached): %v", err)
	}
	if calls != 1 {
		t.Errorf("expected cache hit on second resolve, got %d calls", calls)
	}
}

func TestResolver_OfflineMissingCache(t *testing.T) {
	resolver, closer := newTestResolver(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("offline resolver should not call registry")
	})
	defer closer()
	resolver.Offline = true

	_, err := resolver.Resolve(context.Background(), schema.Skill{ID: "not-cached"})
	if err == nil || !strings.Contains(err.Error(), "--offline") {
		t.Fatalf("expected offline error, got %v", err)
	}
}

func TestResolver_SourceRejectsNonGitHub(t *testing.T) {
	resolver, closer := newTestResolver(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not hit registry when source is invalid")
	})
	defer closer()

	_, err := resolver.Resolve(context.Background(), schema.Skill{
		ID:     "external",
		Source: "https://example.com/some/skill.md",
	})
	if err == nil || !strings.Contains(err.Error(), "github.com") {
		t.Fatalf("expected github.com URL error, got %v", err)
	}
}

func TestResolver_SourceFetchesGitHubDirectory(t *testing.T) {
	files := map[string]string{
		"skills/skill-creator/SKILL.md":         "---\nname: skill-creator\ndescription: from gh\n---\nbody\n",
		"skills/skill-creator/scripts/hello.sh": "#!/bin/sh\necho hi\n",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/skills/git/trees/main", func(w http.ResponseWriter, r *http.Request) {
		resp := treeResponse{
			Tree: []treeEntry{
				{Path: "skills/skill-creator", Type: "tree"},
				{Path: "skills/skill-creator/SKILL.md", Type: "blob"},
				{Path: "skills/skill-creator/scripts", Type: "tree"},
				{Path: "skills/skill-creator/scripts/hello.sh", Type: "blob"},
				{Path: "skills/other-skill/SKILL.md", Type: "blob"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var body string
		var ok bool
		for repoPath, content := range files {
			if strings.HasSuffix(r.URL.Path, "/"+repoPath) {
				body = content
				ok = true
				break
			}
		}
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(body))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cache, err := NewCache(t.TempDir())
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	resolver := &Resolver{
		Client: NewClient(""),
		Installer: &Installer{
			Client:  srv.Client(),
			APIBase: srv.URL,
			RawBase: srv.URL,
		},
		Cache: cache,
	}

	resolved, err := resolver.Resolve(context.Background(), schema.Skill{
		ID:     "skill-creator",
		Source: "https://github.com/acme/skills/tree/main/skills/skill-creator",
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if resolved.Name != "skill-creator" || resolved.Description != "from gh" {
		t.Errorf("unexpected metadata: %+v", resolved)
	}
	if len(resolved.Files) != 2 {
		t.Fatalf("expected 2 files (SKILL.md + scripts/hello.sh), got %d: %v", len(resolved.Files), keysOf(resolved.Files))
	}
	if _, ok := resolved.Files["SKILL.md"]; !ok {
		t.Errorf("missing SKILL.md, got keys: %v", keysOf(resolved.Files))
	}
	if _, ok := resolved.Files["scripts/hello.sh"]; !ok {
		t.Errorf("missing scripts/hello.sh, got keys: %v", keysOf(resolved.Files))
	}
	if resolved.Version != "main" {
		t.Errorf("expected Version derived from ref 'main', got %q", resolved.Version)
	}

	if _, ok, err := resolver.Cache.Get("skill-creator", "main"); err != nil || !ok {
		t.Errorf("expected cache to be populated after Resolve, got ok=%v err=%v", ok, err)
	}
}

func keysOf(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
