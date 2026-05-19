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

	files := map[string][]byte{
		"SKILL.md":          []byte("---\nname: x\ndescription: y\n---\nbody\n"),
		"scripts/helper.sh": []byte("#!/bin/sh\necho hi\n"),
	}
	if err := cache.Put("my-skill", "1.0.0", files); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, ok, err := cache.Get("my-skill", "1.0.0")
	if err != nil || !ok {
		t.Fatalf("expected cache hit, got ok=%v err=%v", ok, err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 files, got %d", len(got))
	}
	if string(got["SKILL.md"]) != string(files["SKILL.md"]) {
		t.Errorf("SKILL.md mismatch: got %q", got["SKILL.md"])
	}
	if string(got["scripts/helper.sh"]) != string(files["scripts/helper.sh"]) {
		t.Errorf("scripts/helper.sh mismatch: got %q", got["scripts/helper.sh"])
	}

	wantDir := filepath.Join(dir, "my-skill@1.0.0")
	if cache.SkillDir("my-skill", "1.0.0") != wantDir {
		t.Errorf("SkillDir = %q, want %q", cache.SkillDir("my-skill", "1.0.0"), wantDir)
	}
	if cache.SkillDir("foo", "") != filepath.Join(dir, "foo@latest") {
		t.Errorf("SkillDir for empty version should use 'latest', got %q", cache.SkillDir("foo", ""))
	}
}

func TestCache_PutClearsStaleFiles(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	if err := cache.Put("a", "main", map[string][]byte{
		"SKILL.md":         []byte("---\nname: a\ndescription: v1\n---\n"),
		"scripts/v1.sh":    []byte("v1"),
		"resources/old.md": []byte("old"),
	}); err != nil {
		t.Fatalf("first Put: %v", err)
	}
	if err := cache.Put("a", "main", map[string][]byte{
		"SKILL.md":      []byte("---\nname: a\ndescription: v2\n---\n"),
		"scripts/v2.sh": []byte("v2"),
	}); err != nil {
		t.Fatalf("second Put: %v", err)
	}

	got, ok, err := cache.Get("a", "main")
	if err != nil || !ok {
		t.Fatalf("Get after second Put: ok=%v err=%v", ok, err)
	}
	if _, exists := got["scripts/v1.sh"]; exists {
		t.Error("expected stale scripts/v1.sh to be cleared, but it is still present")
	}
	if _, exists := got["resources/old.md"]; exists {
		t.Error("expected stale resources/old.md to be cleared, but it is still present")
	}
	if string(got["scripts/v2.sh"]) != "v2" {
		t.Errorf("expected fresh scripts/v2.sh, got %q", got["scripts/v2.sh"])
	}
}
