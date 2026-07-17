package generator

import (
	"os"
	"path/filepath"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

func TestGenerator_seedExamples(t *testing.T) {
	dir := t.TempDir()
	g := &Generator{}
	adl := &schema.ADL{}
	adl.Spec.Examples = schema.Examples{
		{Title: "Basic Chat", Description: "A minimal conversation."},
	}

	if err := g.seedExamples(adl, dir); err != nil {
		t.Fatal(err)
	}
	stub := filepath.Join(dir, "examples", "basic-chat", "README.md")
	content, err := os.ReadFile(stub)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) == "" || !filepath.IsAbs(stub) {
		t.Fatal("empty stub")
	}

	if err := os.WriteFile(stub, []byte("user edit"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := g.seedExamples(adl, dir); err != nil {
		t.Fatal(err)
	}
	content, _ = os.ReadFile(stub)
	if string(content) != "user edit" {
		t.Fatalf("stub was overwritten: %q", content)
	}
}
