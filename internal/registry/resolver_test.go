package registry

import (
	"context"
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
		Client: NewClient(srv.URL + "/skills/"),
		Cache:  cache,
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
	if len(resolved.Body) != 0 {
		t.Errorf("bare skill should not have Body, got %q", resolved.Body)
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

func TestResolver_SourceOverride(t *testing.T) {
	var seenPath string
	resolver, closer := newTestResolver(t, func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		_, _ = w.Write([]byte("---\nname: external\ndescription: from override\n---\nbody\n"))
	})
	defer closer()

	skill := schema.Skill{
		ID:     "external-skill",
		Source: resolver.Client.BaseURL + "../custom/path/external.md",
	}

	resolved, err := resolver.Resolve(context.Background(), skill)
	if err != nil {
		t.Fatalf("Resolve with source override: %v", err)
	}
	if resolved.Name != "external" {
		t.Errorf("unexpected name: %q", resolved.Name)
	}
	if !strings.Contains(seenPath, "external.md") {
		t.Errorf("expected fetch through source override path, got %s", seenPath)
	}
}
