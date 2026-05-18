package registry

import (
	"path/filepath"
	"testing"
)

func TestCache_GetPut(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	if got, ok, err := cache.Get("missing", ""); err != nil || ok || got != nil {
		t.Fatalf("expected empty cache miss, got ok=%v err=%v data=%v", ok, err, got)
	}

	payload := []byte("---\nname: x\ndescription: y\n---\nbody\n")
	if err := cache.Put("my-skill", "1.0.0", payload); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, ok, err := cache.Get("my-skill", "1.0.0")
	if err != nil || !ok {
		t.Fatalf("expected cache hit, got ok=%v err=%v", ok, err)
	}
	if string(got) != string(payload) {
		t.Errorf("payload mismatch: got %q want %q", got, payload)
	}

	wantPath := filepath.Join(dir, "my-skill@1.0.0.md")
	if cache.Path("my-skill", "1.0.0") != wantPath {
		t.Errorf("Path = %q, want %q", cache.Path("my-skill", "1.0.0"), wantPath)
	}

	if cache.Path("foo", "") != filepath.Join(dir, "foo@latest.md") {
		t.Errorf("Path for empty version should use 'latest', got %q", cache.Path("foo", ""))
	}
}
