package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

// Cache is a simple disk-backed cache for fetched skill markdown blobs.
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

// Path returns the on-disk path that would be used for a skill id+version.
// version may be empty to denote the registry default.
func (c *Cache) Path(id, version string) string {
	v := version
	if v == "" {
		v = "latest"
	}
	return filepath.Join(c.dir, fmt.Sprintf("%s@%s.md", id, v))
}

// Get returns the cached bytes for (id, version) if present.
// The boolean is false when the entry is absent.
func (c *Cache) Get(id, version string) ([]byte, bool, error) {
	path := c.Path(id, version)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to read skill cache entry %s: %w", path, err)
	}
	return data, true, nil
}

// Put writes the bytes for (id, version) into the cache, creating the
// cache directory as needed.
func (c *Cache) Put(id, version string, data []byte) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return fmt.Errorf("failed to create skill cache directory: %w", err)
	}
	path := c.Path(id, version)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write skill cache entry %s: %w", path, err)
	}
	return nil
}
