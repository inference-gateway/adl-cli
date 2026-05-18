package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the canonical skill registry endpoint used when no
// per-skill `source:` overrides it.
const DefaultBaseURL = "https://registry.inference-gateway.com/skills/"

// EnvBaseURL is the env var that overrides DefaultBaseURL.
const EnvBaseURL = "ADL_SKILLS_REGISTRY"

// Client fetches SKILL.md from the default registry by id.
//
// `source:` overrides do NOT flow through this client; they go through
// the GitHub Installer instead so they can pull a whole skill directory
// (scripts, resources, …).
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient returns a Client with sensible defaults.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchByID retrieves a skill's SKILL.md by id (and optional version)
// from the configured BaseURL. Version "" resolves to the registry's
// default version of that skill.
func (c *Client) FetchByID(ctx context.Context, id, version string) ([]byte, error) {
	if id == "" {
		return nil, fmt.Errorf("skill id is required")
	}

	path := id
	if version != "" {
		path = id + "/" + version
	}
	path += ".md"

	target, err := joinURL(c.BaseURL, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request for %s: %w", target, err)
	}
	req.Header.Set("Accept", "text/markdown, text/plain; q=0.9, */*; q=0.1")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("skill not found at %s", target)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("registry returned %s for %s", resp.Status, target)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", target, err)
	}
	return body, nil
}

func joinURL(base, path string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid base URL %q: %w", base, err)
	}
	ref, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid path %q: %w", path, err)
	}
	return u.ResolveReference(ref).String(), nil
}
