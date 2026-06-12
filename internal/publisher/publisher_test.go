package publisher

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNormalizeRepoURL(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "https with .git", in: "https://github.com/acme/cool-agent.git", want: "https://github.com/acme/cool-agent"},
		{name: "https without .git", in: "https://github.com/acme/cool-agent", want: "https://github.com/acme/cool-agent"},
		{name: "https trailing slash", in: "https://github.com/acme/cool-agent/", want: "https://github.com/acme/cool-agent"},
		{name: "scp ssh", in: "git@github.com:acme/cool-agent.git", want: "https://github.com/acme/cool-agent"},
		{name: "ssh url", in: "ssh://git@github.com/acme/cool-agent.git", want: "https://github.com/acme/cool-agent"},
		{name: "host path", in: "github.com/acme/cool-agent", want: "https://github.com/acme/cool-agent"},
		{name: "bare owner/repo", in: "acme/cool-agent", want: "https://github.com/acme/cool-agent"},
		{name: "credentials in url", in: "https://user:token@github.com/acme/cool-agent.git", want: "https://github.com/acme/cool-agent"},
		{name: "whitespace", in: "  https://github.com/acme/cool-agent\n", want: "https://github.com/acme/cool-agent"},
		{name: "empty", in: "", wantErr: true},
		{name: "non-github host", in: "https://gitlab.com/acme/cool-agent", wantErr: true},
		{name: "too few segments", in: "https://github.com/acme", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeRepoURL(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %q", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeRepoURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestAppendEntry(t *testing.T) {
	existing := "# Catalog of agents\n- url: https://github.com/foo/bar\n  ref: main\n"
	got, err := AppendEntry([]byte(existing), "https://github.com/acme/cool-agent", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := existing + "- url: https://github.com/acme/cool-agent\n  ref: main\n"
	if string(got) != want {
		t.Fatalf("AppendEntry mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestAppendEntryPreservesLeadingComment(t *testing.T) {
	existing := "# Leading comment block\n# stays intact\n"
	got, err := AppendEntry([]byte(existing), "https://github.com/acme/cool-agent", "v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(string(got), existing) {
		t.Fatalf("leading comment not preserved: %q", got)
	}
	if !strings.HasSuffix(string(got), "- url: https://github.com/acme/cool-agent\n  ref: v1.2.3\n") {
		t.Fatalf("entry not appended correctly: %q", got)
	}
}

func TestAppendEntryAddsMissingTrailingNewline(t *testing.T) {
	existing := "- url: https://github.com/foo/bar\n  ref: main" // no trailing newline
	got, err := AppendEntry([]byte(existing), "https://github.com/acme/cool-agent", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "- url: https://github.com/foo/bar\n  ref: main\n- url: https://github.com/acme/cool-agent\n  ref: main\n"
	if string(got) != want {
		t.Fatalf("AppendEntry mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestAppendEntryDuplicate(t *testing.T) {
	existing := "- url: https://github.com/acme/cool-agent\n  ref: main\n"
	if _, err := AppendEntry([]byte(existing), "https://github.com/acme/cool-agent/", "main"); err == nil {
		t.Fatal("expected duplicate error, got nil")
	}
}

func TestBuildTitle(t *testing.T) {
	if got := BuildTitle("cool-agent"); got != "Add cool-agent to the agents catalog" {
		t.Fatalf("BuildTitle = %q", got)
	}
}

func TestBuildBody(t *testing.T) {
	body := BuildBody(Metadata{
		Name:        "cool-agent",
		Description: "Does cool things",
		Version:     "0.1.0",
	}, "https://github.com/acme/cool-agent", "main")

	for _, want := range []string{
		"Add `cool-agent` to the agents catalog",
		"Does cool things",
		"metadata.description",
		"https://github.com/acme/cool-agent",
		"(`main`)",
		"<!--",
		"-->",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q:\n%s", want, body)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Cool Agent":     "cool-agent",
		"weather_bot v2": "weather-bot-v2",
		"  --weird--  ":  "weird",
		"":               "agent",
		"!!!":            "agent",
	}
	for in, want := range tests {
		if got := slugify(in); got != want {
			t.Fatalf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

// recordingRunner captures every command invocation and returns canned
// responses keyed by a substring match of the joined command line.
type recordingRunner struct {
	calls   []recordedCall
	respond func(joined string, stdin []byte) ([]byte, error)
}

type recordedCall struct {
	args  []string
	stdin []byte
}

func (r *recordingRunner) Run(_ context.Context, stdin []byte, name string, args ...string) ([]byte, error) {
	full := append([]string{name}, args...)
	r.calls = append(r.calls, recordedCall{args: full, stdin: stdin})
	return r.respond(strings.Join(full, " "), stdin)
}

func (r *recordingRunner) find(substr string) (recordedCall, bool) {
	for _, c := range r.calls {
		if strings.Contains(strings.Join(c.args, " "), substr) {
			return c, true
		}
	}
	return recordedCall{}, false
}

func TestPublishDryRun(t *testing.T) {
	var out strings.Builder
	runner := &recordingRunner{respond: func(joined string, _ []byte) ([]byte, error) {
		t.Fatalf("dry run must not run commands, got: %s", joined)
		return nil, nil
	}}
	p := &Publisher{Runner: runner, Out: &out}

	url, err := p.Publish(context.Background(), Metadata{Name: "cool-agent", Description: "Does cool things"}, Options{
		RepoURL: "https://github.com/acme/cool-agent",
		Ref:     "main",
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "" {
		t.Fatalf("dry run should not return a URL, got %q", url)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("dry run should not invoke any commands, got %d", len(runner.calls))
	}
	for _, want := range []string{"Dry run", "- url: https://github.com/acme/cool-agent", "Add cool-agent to the agents catalog"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, out.String())
		}
	}
}

func TestPublishFullFlow(t *testing.T) {
	catalog := "# Inference Gateway agents catalog\n- url: https://github.com/foo/bar\n  ref: main\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(catalog))

	runner := &recordingRunner{}
	runner.respond = func(joined string, _ []byte) ([]byte, error) {
		switch {
		case joined == "gh auth status":
			return nil, nil
		case joined == "gh api user --jq .login":
			return []byte("octocat\n"), nil
		case strings.HasPrefix(joined, "gh repo fork "):
			return nil, nil
		case strings.HasPrefix(joined, "gh repo sync "):
			return nil, nil
		case strings.Contains(joined, "--jq .default_branch"):
			return []byte("main\n"), nil
		case strings.Contains(joined, "contents/agents.yaml?ref=main"):
			resp, _ := json.Marshal(map[string]string{"content": encoded, "encoding": "base64", "sha": "filesha123"})
			return resp, nil
		case strings.Contains(joined, "git/refs/heads/main --jq .object.sha"):
			return []byte("basesha456\n"), nil
		case strings.Contains(joined, "git/refs -X POST"):
			return []byte("{}"), nil
		case strings.Contains(joined, "contents/agents.yaml -X PUT"):
			return []byte("{}"), nil
		case strings.HasPrefix(joined, "gh pr create"):
			return []byte("https://github.com/inference-gateway/agents/pull/42\n"), nil
		}
		t.Fatalf("unexpected command: %s", joined)
		return nil, nil
	}

	var out strings.Builder
	p := &Publisher{Runner: runner, Out: &out, Sleep: func(_ time.Duration) {}}

	url, err := p.Publish(context.Background(), Metadata{
		Name:        "cool-agent",
		Description: "Does cool things",
		Version:     "0.1.0",
	}, Options{RepoURL: "https://github.com/acme/cool-agent", Ref: "v1.0.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://github.com/inference-gateway/agents/pull/42" {
		t.Fatalf("unexpected PR URL: %q", url)
	}

	if _, ok := runner.find("gh repo fork inference-gateway/agents"); !ok {
		t.Fatal("expected fork of inference-gateway/agents")
	}
	put, ok := runner.find("repos/octocat/agents/contents/agents.yaml -X PUT")
	if !ok {
		t.Fatal("expected PUT to the fork's agents.yaml")
	}
	var putBody struct {
		Content string `json:"content"`
		Branch  string `json:"branch"`
		SHA     string `json:"sha"`
	}
	if err := json.Unmarshal(put.stdin, &putBody); err != nil {
		t.Fatalf("could not parse PUT body: %v", err)
	}
	if putBody.SHA != "filesha123" {
		t.Fatalf("PUT used wrong file sha: %q", putBody.SHA)
	}
	decoded, err := base64.StdEncoding.DecodeString(putBody.Content)
	if err != nil {
		t.Fatalf("could not decode PUT content: %v", err)
	}
	if !strings.Contains(string(decoded), "- url: https://github.com/acme/cool-agent\n  ref: v1.0.0") {
		t.Fatalf("committed content missing new entry with agent ref:\n%s", decoded)
	}
	if !strings.HasPrefix(putBody.Branch, "adl-publish-cool-agent-") {
		t.Fatalf("unexpected branch name: %q", putBody.Branch)
	}

	pr, ok := runner.find("gh pr create")
	if !ok {
		t.Fatal("expected gh pr create call")
	}
	prArgs := strings.Join(pr.args, " ")
	if !strings.Contains(prArgs, "--head octocat:adl-publish-cool-agent-") {
		t.Fatalf("PR head not pointed at fork branch: %s", prArgs)
	}
	if !strings.Contains(prArgs, "--repo inference-gateway/agents") {
		t.Fatalf("PR not targeted at upstream: %s", prArgs)
	}
	if !strings.Contains(prArgs, "--base main") {
		t.Fatalf("PR base must be the catalog repo's default branch: %s", prArgs)
	}
	if !strings.Contains(string(pr.stdin), "(`v1.0.0`)") {
		t.Fatalf("PR body missing agent ref: %s", pr.stdin)
	}
}

func TestPublishDuplicateAborts(t *testing.T) {
	catalog := "- url: https://github.com/acme/cool-agent\n  ref: main\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(catalog))

	runner := &recordingRunner{}
	runner.respond = func(joined string, _ []byte) ([]byte, error) {
		switch {
		case joined == "gh auth status":
			return nil, nil
		case joined == "gh api user --jq .login":
			return []byte("octocat\n"), nil
		case strings.HasPrefix(joined, "gh repo fork "):
			return nil, nil
		case strings.HasPrefix(joined, "gh repo sync "):
			return nil, nil
		case strings.Contains(joined, "--jq .default_branch"):
			return []byte("main\n"), nil
		case strings.Contains(joined, "contents/agents.yaml?ref="):
			resp, _ := json.Marshal(map[string]string{"content": encoded, "encoding": "base64", "sha": "filesha"})
			return resp, nil
		}
		t.Fatalf("duplicate should abort before %s", joined)
		return nil, nil
	}

	p := &Publisher{Runner: runner, Out: &strings.Builder{}, Sleep: func(_ time.Duration) {}}
	_, err := p.Publish(context.Background(), Metadata{Name: "cool-agent", Description: "x"}, Options{
		RepoURL: "https://github.com/acme/cool-agent",
		Ref:     "main",
	})
	if err == nil || !strings.Contains(err.Error(), "already in the catalog") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestResolveRepoURLFromGit(t *testing.T) {
	runner := &recordingRunner{respond: func(joined string, _ []byte) ([]byte, error) {
		if strings.Contains(joined, "git -C . remote get-url origin") {
			return []byte("git@github.com:acme/cool-agent.git\n"), nil
		}
		return nil, nil
	}}
	got, err := ResolveRepoURLFromGit(context.Background(), runner, ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://github.com/acme/cool-agent" {
		t.Fatalf("got %q", got)
	}
}
