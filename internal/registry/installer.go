package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

const (
	githubAPIBase     = "https://api.github.com"
	githubRawBase     = "https://raw.githubusercontent.com"
	installerTimeout  = 30 * time.Second
	installerUA       = "inference-gateway-adl-cli"
	treePartsExpected = 5

	defaultSkillsOrg    = "inference-gateway"
	defaultSkillsRepo   = "skills"
	defaultSkillsRef    = "main"
	defaultSkillsSubdir = "skills"
)

// ExpandShorthand turns shorthand install targets into full GitHub
// tree URLs. Accepted forms (an optional `@<tag>` suffix pins a branch
// or tag; without it, the default ref "main" is used):
//
//   - "<skill>[@<tag>]"
//     → https://github.com/inference-gateway/skills/tree/<tag>/skills/<skill>
//   - "<owner>/<repo>/<skill>[@<tag>]"
//     → https://github.com/<owner>/<repo>/tree/<tag>/skills/<skill>
//   - any http(s):// URL → returned unchanged
//
// Anything else (two slash-separated segments, four-plus segments,
// empty segments, etc.) is returned unchanged so ParseGitHubTreeURL
// produces a clear error.
func ExpandShorthand(input string) string {
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return input
	}
	trimmed := strings.Trim(input, "/")
	if trimmed == "" {
		return input
	}

	ref := defaultSkillsRef
	if at := strings.LastIndex(trimmed, "@"); at >= 0 {
		tag := trimmed[at+1:]
		body := trimmed[:at]
		if tag == "" || body == "" {
			return input
		}
		trimmed = body
		ref = tag
	}

	parts := strings.Split(trimmed, "/")
	if slices.Contains(parts, "") {
		return input
	}
	switch len(parts) {
	case 1:
		return fmt.Sprintf("https://github.com/%s/%s/tree/%s/%s/%s",
			defaultSkillsOrg, defaultSkillsRepo, ref, defaultSkillsSubdir, parts[0])
	case 3:
		return fmt.Sprintf("https://github.com/%s/%s/tree/%s/%s/%s",
			parts[0], parts[1], ref, defaultSkillsSubdir, parts[2])
	default:
		return input
	}
}

// GitHubLocation identifies a directory inside a public GitHub repository,
// parsed out of a /tree/<ref>/<path> URL.
type GitHubLocation struct {
	Owner string
	Repo  string
	Ref   string
	Path  string
}

// ParseGitHubTreeURL accepts URLs of the form
//
//	https://github.com/<owner>/<repo>/tree/<ref>/<path-to-skill>
//
// Refs containing a literal "/" (e.g. "feature/foo" branches) are not
// supported; pass the URL of a tag or single-segment branch instead.
func ParseGitHubTreeURL(rawURL string) (*GitHubLocation, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("URL must use http(s): got %q", u.Scheme)
	}
	if u.Host != "github.com" {
		return nil, fmt.Errorf("only github.com URLs are supported, got %q", u.Host)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 3 && parts[2] == "blob" {
		return nil, fmt.Errorf("URL points to a file (/blob/) - pass the directory URL (/tree/) instead: %s", rawURL)
	}
	if len(parts) < treePartsExpected || parts[2] != "tree" {
		return nil, fmt.Errorf("URL must be of the form https://github.com/<owner>/<repo>/tree/<ref>/<path>: got %s", rawURL)
	}

	return &GitHubLocation{
		Owner: parts[0],
		Repo:  parts[1],
		Ref:   parts[3],
		Path:  strings.Join(parts[4:], "/"),
	}, nil
}

type treeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type treeResponse struct {
	Tree      []treeEntry `json:"tree"`
	Truncated bool        `json:"truncated"`
}

// Installer downloads a skill folder from a public GitHub repo and
// returns its contents as a relative-path → bytes map. Tests substitute
// APIBase / RawBase to point at httptest.Server.
type Installer struct {
	Client  *http.Client
	APIBase string
	RawBase string
}

// NewInstaller returns an Installer pointed at github.com with a 30s HTTP
// timeout.
func NewInstaller() *Installer {
	return &Installer{
		Client:  &http.Client{Timeout: installerTimeout},
		APIBase: githubAPIBase,
		RawBase: githubRawBase,
	}
}

// Fetch downloads every blob under loc.Path at loc.Ref and returns a
// map keyed by the blob's path *relative to loc.Path*. The map is
// guaranteed to contain at least one entry; callers are responsible for
// verifying that "SKILL.md" is present.
func (i *Installer) Fetch(ctx context.Context, loc *GitHubLocation) (map[string][]byte, error) {
	tree, err := i.fetchTree(ctx, loc)
	if err != nil {
		return nil, err
	}

	prefix := loc.Path + "/"
	var blobs []treeEntry
	for _, e := range tree.Tree {
		if e.Type != "blob" {
			continue
		}
		if !strings.HasPrefix(e.Path, prefix) {
			continue
		}
		blobs = append(blobs, e)
	}
	if len(blobs) == 0 {
		return nil, fmt.Errorf("no files found under %s/%s/%s @ %s - check the URL", loc.Owner, loc.Repo, loc.Path, loc.Ref)
	}

	files := make(map[string][]byte, len(blobs))
	for _, b := range blobs {
		rel := strings.TrimPrefix(b.Path, prefix)
		data, err := i.downloadBlob(ctx, loc, b.Path)
		if err != nil {
			return nil, err
		}
		files[rel] = data
	}
	return files, nil
}

func (i *Installer) fetchTree(ctx context.Context, loc *GitHubLocation) (*treeResponse, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/git/trees/%s?recursive=1", i.APIBase, loc.Owner, loc.Repo, loc.Ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create tree request: %w", err)
	}
	req.Header.Set("User-Agent", installerUA)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := i.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tree: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("repository or ref not found: %s/%s @ %s", loc.Owner, loc.Repo, loc.Ref)
	case http.StatusForbidden:
		return nil, fmt.Errorf("GitHub API rate limit exceeded (60 req/hour for unauthenticated requests); try again later")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tree treeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, fmt.Errorf("failed to parse tree response: %w", err)
	}
	if tree.Truncated {
		return nil, fmt.Errorf("repository tree was truncated by GitHub (repo too large) - cannot reliably install")
	}
	return &tree, nil
}

func (i *Installer) downloadBlob(ctx context.Context, loc *GitHubLocation, repoPath string) ([]byte, error) {
	rawURL := fmt.Sprintf("%s/%s/%s/%s/%s", i.RawBase, loc.Owner, loc.Repo, loc.Ref, repoPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", repoPath, err)
	}
	req.Header.Set("User-Agent", installerUA)

	resp, err := i.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", repoPath, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s: status %d", repoPath, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", repoPath, err)
	}
	return data, nil
}
