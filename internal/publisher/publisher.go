// Package publisher submits a generated agent to the public catalog at
// registry.inference-gateway.com by opening a pull request against
// inference-gateway/agents that appends the agent's repository as a new
// { url, ref } entry in agents.yaml.
//
// The catalog repo stores only pointers to agent repositories; an agent's
// name, description, and version come from its own agent.yaml (metadata.*)
// when the catalog is built. This package therefore seeds the PR title and
// body from the local manifest and leaves the agents.yaml entry as the bare
// { url, ref } pair.
package publisher

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultCatalogOwner / DefaultCatalogRepo / DefaultCatalogFile identify
	// the catalog source that backs registry.inference-gateway.com.
	DefaultCatalogOwner = "inference-gateway"
	DefaultCatalogRepo  = "agents"
	DefaultCatalogFile  = "agents.yaml"

	// DefaultRef is the ref recorded in the catalog entry when none is given.
	DefaultRef = "main"

	// forkPropagationDelay is how long to wait between attempts to create a
	// branch on a freshly created fork, which GitHub provisions asynchronously.
	forkPropagationDelay = 2 * time.Second
	forkPropagationTries = 5
)

// Metadata is the subset of an agent's manifest used to seed the PR.
type Metadata struct {
	Name        string
	Description string
	Version     string
}

// Options configures a single publish run.
type Options struct {
	// RepoURL is the canonical agent repository URL
	// (https://github.com/<owner>/<repo>).
	RepoURL string
	// Ref is the branch, tag, or SHA recorded in the catalog entry.
	Ref string
	// DryRun prints the resolved entry, PR title, and body without opening a PR.
	DryRun bool

	// Catalog target overrides (default to the inference-gateway/agents
	// constants when empty); primarily a testing seam.
	CatalogOwner string
	CatalogRepo  string
	CatalogFile  string
}

// Runner executes an external command, optionally feeding it stdin, and
// returns its stdout. It is the single seam that lets tests avoid shelling
// out to git/gh.
type Runner interface {
	Run(ctx context.Context, stdin []byte, name string, args ...string) ([]byte, error)
}

// ExecRunner is the production Runner backed by os/exec.
type ExecRunner struct{}

// Run executes name with args, writing stdin if non-nil, and returns stdout.
// On failure the returned error carries the command's stderr (or stdout).
func (ExecRunner) Run(ctx context.Context, stdin []byte, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg != "" {
			return stdout.Bytes(), fmt.Errorf("%s: %w", msg, err)
		}
		return stdout.Bytes(), err
	}
	return stdout.Bytes(), nil
}

// Publisher orchestrates the catalog contribution flow.
type Publisher struct {
	Runner Runner
	Out    io.Writer
	// Sleep is used to back off while a fork propagates; overridable in tests.
	Sleep func(time.Duration)
}

// New returns a Publisher wired to the real git/gh executables.
func New(out io.Writer) *Publisher {
	return &Publisher{
		Runner: ExecRunner{},
		Out:    out,
		Sleep:  time.Sleep,
	}
}

// Publish opens (or, in dry-run mode, previews) a pull request that appends
// opts.RepoURL to the catalog. On success it returns the new PR's URL.
func (p *Publisher) Publish(ctx context.Context, meta Metadata, opts Options) (string, error) {
	owner := firstNonEmpty(opts.CatalogOwner, DefaultCatalogOwner)
	repo := firstNonEmpty(opts.CatalogRepo, DefaultCatalogRepo)
	file := firstNonEmpty(opts.CatalogFile, DefaultCatalogFile)
	ref := firstNonEmpty(opts.Ref, DefaultRef)

	title := BuildTitle(meta.Name)
	body := BuildBody(meta, opts.RepoURL, ref)

	if opts.DryRun {
		p.printf("Dry run - no pull request will be created.\n\n")
		p.printf("Catalog repository: %s/%s (%s)\n", owner, repo, file)
		p.printf("Agent repository:   %s\n", opts.RepoURL)
		p.printf("Ref:                %s\n\n", ref)
		p.printf("PR title:\n  %s\n\n", title)
		p.printf("PR body:\n%s\n\n", indentLines(body, "  "))
		p.printf("Catalog entry to append:\n  - url: %s\n    ref: %s\n", opts.RepoURL, ref)
		return "", nil
	}

	if _, err := p.Runner.Run(ctx, nil, "gh", "auth", "status"); err != nil {
		return "", fmt.Errorf("the GitHub CLI 'gh' must be installed (https://cli.github.com) and authenticated (run 'gh auth login'): %w", err)
	}

	login, err := p.ghAPIString(ctx, "user", ".login")
	if err != nil || login == "" {
		return "", fmt.Errorf("could not determine your GitHub username via 'gh api user': %w", err)
	}

	p.printf("Forking %s/%s...\n", owner, repo)
	if _, err := p.Runner.Run(ctx, nil, "gh", "repo", "fork", owner+"/"+repo, "--clone=false"); err != nil {
		return "", fmt.Errorf("failed to fork %s/%s: %w", owner, repo, err)
	}
	// Best-effort: bring an existing fork up to date so the upstream base SHA
	// is guaranteed to exist in the fork before we branch from it.
	_, _ = p.Runner.Run(ctx, nil, "gh", "repo", "sync", login+"/"+repo)

	// The catalog repo's own default branch is the PR base, distinct from the
	// agent's ref (which is only recorded in the catalog entry).
	baseBranch, err := p.ghAPIString(ctx, fmt.Sprintf("repos/%s/%s", owner, repo), ".default_branch")
	if err != nil || baseBranch == "" {
		baseBranch = "main"
	}

	content, fileSHA, err := p.readFile(ctx, owner, repo, file, baseBranch)
	if err != nil {
		return "", err
	}

	updated, err := AppendEntry(content, opts.RepoURL, ref)
	if err != nil {
		return "", err
	}

	baseSHA, err := p.ghAPIString(ctx, fmt.Sprintf("repos/%s/%s/git/refs/heads/%s", owner, repo, baseBranch), ".object.sha")
	if err != nil || baseSHA == "" {
		return "", fmt.Errorf("could not resolve the %q branch on %s/%s: %w", baseBranch, owner, repo, err)
	}

	branch := fmt.Sprintf("adl-publish-%s-%d", slugify(meta.Name), time.Now().Unix())
	if err := p.createBranch(ctx, login, repo, branch, baseSHA); err != nil {
		return "", err
	}

	if err := p.putFile(ctx, login, repo, file, branch, title, updated, fileSHA); err != nil {
		return "", err
	}

	p.printf("Opening pull request against %s/%s...\n", owner, repo)
	prURL, err := p.createPR(ctx, owner, repo, baseBranch, login, branch, title, body)
	if err != nil {
		return "", err
	}
	return prURL, nil
}

// ResolveRepoURLFromGit derives the canonical repository URL from the
// 'origin' remote of the git repository at dir.
func ResolveRepoURLFromGit(ctx context.Context, runner Runner, dir string) (string, error) {
	out, err := runner.Run(ctx, nil, "git", "-C", dir, "remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("could not read git remote 'origin': %w", err)
	}
	return NormalizeRepoURL(string(out))
}

// NormalizeRepoURL converts the common git remote forms (https, scp-like ssh,
// ssh://, github.com/owner/repo, or a bare owner/repo) into the canonical
// https://github.com/<owner>/<repo> shape with no .git suffix.
func NormalizeRepoURL(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("empty repository URL")
	}

	host := "github.com"
	path := s

	switch {
	case strings.HasPrefix(s, "git@"):
		rest := strings.TrimPrefix(s, "git@")
		parts := strings.SplitN(rest, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("unrecognized SSH remote %q", raw)
		}
		host, path = parts[0], parts[1]
	case strings.Contains(s, "://"):
		u, err := url.Parse(s)
		if err != nil {
			return "", fmt.Errorf("invalid repository URL %q: %w", raw, err)
		}
		host, path = u.Host, u.Path
	case strings.HasPrefix(s, "github.com/"):
		path = strings.TrimPrefix(s, "github.com/")
	}

	if host != "github.com" {
		return "", fmt.Errorf("only github.com repositories are supported, got host %q (use --url to set it explicitly)", host)
	}

	path = strings.Trim(path, "/")
	path = strings.TrimSuffix(path, ".git")
	segs := strings.Split(path, "/")
	if len(segs) != 2 || segs[0] == "" || segs[1] == "" {
		return "", fmt.Errorf("could not parse owner/repo from %q", raw)
	}
	return fmt.Sprintf("https://github.com/%s/%s", segs[0], segs[1]), nil
}

// BuildTitle returns the default PR title for an agent.
func BuildTitle(name string) string {
	return fmt.Sprintf("Add %s to the agents catalog", name)
}

// BuildBody returns the default PR body, seeded from the manifest. It echoes
// metadata.description as the proposed summary and embeds commented guidance
// noting that the live catalog description is pulled from the agent's own
// repository.
func BuildBody(meta Metadata, repoURL, ref string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Add `%s` to the agents catalog\n\n", meta.Name)
	b.WriteString("<!--\n")
	b.WriteString("  Thanks for contributing to the Inference Gateway agents catalog!\n\n")
	b.WriteString("  The entry added by this PR is only a pointer to your repository\n")
	b.WriteString("  ({ url, ref }). The name, description, and version shown in the live\n")
	b.WriteString("  catalog at https://registry.inference-gateway.com are pulled from your\n")
	b.WriteString("  repository's agent.yaml (metadata.*) when the catalog is built - not\n")
	b.WriteString("  from this pull request.\n\n")
	b.WriteString("  Feel free to refine the summary below so maintainers can review your\n")
	b.WriteString("  submission. To change how your agent is described in the catalog, edit\n")
	b.WriteString("  metadata.description in your repository's agent.yaml.\n")
	b.WriteString("-->\n\n")
	if v := strings.TrimSpace(meta.Version); v != "" {
		fmt.Fprintf(&b, "**Agent:** `%s` (v%s)\n\n", meta.Name, strings.TrimPrefix(v, "v"))
	} else {
		fmt.Fprintf(&b, "**Agent:** `%s`\n\n", meta.Name)
	}
	desc := strings.TrimSpace(meta.Description)
	if desc == "" {
		desc = "_No description provided in agent.yaml (metadata.description)._"
	}
	fmt.Fprintf(&b, "**Summary:** %s\n\n", desc)
	fmt.Fprintf(&b, "**Repository:** %s (`%s`)\n", repoURL, ref)
	return b.String()
}

type catalogEntry struct {
	URL string `yaml:"url"`
	Ref string `yaml:"ref"`
}

// AppendEntry appends a single { url, ref } entry to the raw agents.yaml
// content, preserving the existing leading comment block and formatting by
// appending textually. It returns an error if repoURL is already listed.
func AppendEntry(existing []byte, repoURL, ref string) ([]byte, error) {
	if strings.TrimSpace(string(existing)) != "" {
		var entries []catalogEntry
		if err := yaml.Unmarshal(existing, &entries); err != nil {
			return nil, fmt.Errorf("could not parse existing catalog file: %w", err)
		}
		want := canonicalURL(repoURL)
		for _, e := range entries {
			if canonicalURL(e.URL) == want {
				return nil, fmt.Errorf("%s is already in the catalog", repoURL)
			}
		}
	}

	out := existing
	if len(out) > 0 && !bytes.HasSuffix(out, []byte("\n")) {
		out = append(out, '\n')
	}
	out = append(out, []byte(fmt.Sprintf("- url: %s\n  ref: %s\n", repoURL, ref))...)
	return out, nil
}

func canonicalURL(s string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(s), "/"))
}

func (p *Publisher) ghAPIString(ctx context.Context, endpoint, jq string) (string, error) {
	out, err := p.Runner.Run(ctx, nil, "gh", "api", endpoint, "--jq", jq)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (p *Publisher) readFile(ctx context.Context, owner, repo, file, ref string) ([]byte, string, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", owner, repo, file, url.QueryEscape(ref))
	out, err := p.Runner.Run(ctx, nil, "gh", "api", endpoint)
	if err != nil {
		return nil, "", fmt.Errorf("could not read %s from %s/%s@%s: %w", file, owner, repo, ref, err)
	}
	var resp struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
		SHA      string `json:"sha"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, "", fmt.Errorf("could not parse the contents response for %s: %w", file, err)
	}
	if resp.Encoding != "base64" {
		return nil, "", fmt.Errorf("unexpected encoding %q for %s", resp.Encoding, file)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(resp.Content, "\n", ""))
	if err != nil {
		return nil, "", fmt.Errorf("could not decode %s: %w", file, err)
	}
	return decoded, resp.SHA, nil
}

func (p *Publisher) createBranch(ctx context.Context, owner, repo, branch, sha string) error {
	payload, err := json.Marshal(map[string]string{
		"ref": "refs/heads/" + branch,
		"sha": sha,
	})
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("repos/%s/%s/git/refs", owner, repo)
	var lastErr error
	for attempt := 0; attempt < forkPropagationTries; attempt++ {
		if attempt > 0 {
			p.sleep(forkPropagationDelay)
		}
		if _, err := p.Runner.Run(ctx, payload, "gh", "api", endpoint, "-X", "POST", "--input", "-"); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return fmt.Errorf("could not create branch %q on %s/%s (the fork may still be provisioning): %w", branch, owner, repo, lastErr)
}

func (p *Publisher) putFile(ctx context.Context, owner, repo, file, branch, message string, content []byte, sha string) error {
	payload, err := json.Marshal(map[string]string{
		"message": message,
		"content": base64.StdEncoding.EncodeToString(content),
		"branch":  branch,
		"sha":     sha,
	})
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, file)
	if _, err := p.Runner.Run(ctx, payload, "gh", "api", endpoint, "-X", "PUT", "--input", "-"); err != nil {
		return fmt.Errorf("could not commit %s to %s/%s@%s: %w", file, owner, repo, branch, err)
	}
	return nil
}

func (p *Publisher) createPR(ctx context.Context, owner, repo, base, headOwner, branch, title, body string) (string, error) {
	out, err := p.Runner.Run(ctx, []byte(body), "gh", "pr", "create",
		"--repo", owner+"/"+repo,
		"--base", base,
		"--head", headOwner+":"+branch,
		"--title", title,
		"--body-file", "-",
	)
	if err != nil {
		return "", fmt.Errorf("failed to create the pull request: %w", err)
	}
	prURL := lastLine(string(out))
	if prURL == "" {
		return "", fmt.Errorf("pull request created but 'gh' returned no URL")
	}
	return prURL, nil
}

func (p *Publisher) printf(format string, args ...any) {
	if p.Out == nil {
		return
	}
	_, _ = fmt.Fprintf(p.Out, format, args...)
}

func (p *Publisher) sleep(d time.Duration) {
	if p.Sleep != nil {
		p.Sleep(d)
		return
	}
	time.Sleep(d)
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = slugRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "agent"
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func indentLines(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if ln != "" {
			lines[i] = prefix + ln
		}
	}
	return strings.Join(lines, "\n")
}

func lastLine(s string) string {
	s = strings.TrimRight(s, "\n")
	if i := strings.LastIndex(s, "\n"); i >= 0 {
		return strings.TrimSpace(s[i+1:])
	}
	return strings.TrimSpace(s)
}
