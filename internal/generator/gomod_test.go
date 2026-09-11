package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreserveIndirectRequires(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	old := `module example.com/agent

go 1.26

require (
	github.com/inference-gateway/adk v0.26.3
	go.uber.org/zap v1.28.0
)

require (
	google.golang.org/grpc v1.83.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
)
`
	if err := os.WriteFile(path, []byte(old), 0644); err != nil {
		t.Fatal(err)
	}

	generated := "module example.com/agent\n\ngo 1.26\n\nrequire (\n\tgithub.com/inference-gateway/adk v0.26.4\n\tgo.uber.org/multierr v1.12.0 // indirect\n)\n"
	got := preserveIndirectRequires(path, generated)

	if !strings.Contains(got, "\tgoogle.golang.org/grpc v1.83.1 // indirect\n") {
		t.Fatalf("indirect requirement dropped:\n%s", got)
	}
	if strings.Count(got, "go.uber.org/multierr") != 1 || !strings.Contains(got, "multierr v1.12.0") {
		t.Fatalf("generated pin must win over old indirect:\n%s", got)
	}
	if strings.Contains(got, "adk v0.26.3") || !strings.Contains(got, "adk v0.26.4") {
		t.Fatalf("direct pin must come from template:\n%s", got)
	}

	if preserveIndirectRequires(filepath.Join(dir, "missing"), generated) != generated {
		t.Fatal("missing go.mod must return content unchanged")
	}
}
