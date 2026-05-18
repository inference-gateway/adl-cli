package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Cache is a disk-backed cache for fetched skill directories. Each
// cached skill lives under <dir>/<id>@<ref>/ and may contain any number
// of files (SKILL.md plus optional scripts/resources).
type Cache struct {
	dir string
}

// NewCache returns a Cache rooted at the given directory. Pass an empty
// string to use the default ~/.adl/skills-cache location.
func NewCache(dir string) (*Cache, error) {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve user home directory: %w", err)
		}
		dir = filepath.Join(home, ".adl", "skills-cache")
	}
	return &Cache{dir: dir}, nil
}

// Dir returns the cache root directory.
func (c *Cache) Dir() string { return c.dir }

// SkillDir returns the on-disk directory for a skill at (id, ref). An
// empty ref resolves to "latest".
func (c *Cache) SkillDir(id, ref string) string {
	r := ref
	if r == "" {
		r = "latest"
	}
	return filepath.Join(c.dir, fmt.Sprintf("%s@%s", id, r))
}

// Get returns the cached file map for (id, ref) if the directory exists.
// The boolean is false when the entry is absent.
func (c *Cache) Get(id, ref string) (map[string][]byte, bool, error) {
	dir := c.SkillDir(id, ref)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to stat skill cache entry %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, false, fmt.Errorf("cache entry %s is not a directory", dir)
	}

	files := make(map[string][]byte)
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		files[filepath.ToSlash(rel)] = data
		return nil
	}); err != nil {
		return nil, false, fmt.Errorf("failed to read skill cache entry %s: %w", dir, err)
	}
	if len(files) == 0 {
		return nil, false, nil
	}
	return files, true, nil
}

// Put writes the file map into the cache directory for (id, ref). The
// directory is wiped first so stale entries from a previous version
// can't bleed through.
func (c *Cache) Put(id, ref string, files map[string][]byte) error {
	dir := c.SkillDir(id, ref)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to clear skill cache directory %s: %w", dir, err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create skill cache directory: %w", err)
	}
	for rel, data := range files {
		if strings.HasPrefix(rel, "/") || strings.Contains(rel, "..") {
			return fmt.Errorf("refusing to cache file with suspicious relative path: %q", rel)
		}
		outPath := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("failed to create dir for %s: %w", rel, err)
		}
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return fmt.Errorf("failed to write skill cache entry %s: %w", outPath, err)
		}
	}
	return nil
}
