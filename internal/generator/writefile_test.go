package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFile_FormatsGo(t *testing.T) {
	dir := t.TempDir()
	g := &Generator{}
	path := filepath.Join(dir, "main.go")

	unformatted := "package main\n\t\nimport (\n\t\"os\"\n\t\"fmt\"\n)\n\nfunc main() {\n\tfmt.Println( os.Args )\n}// trailing\n"
	if err := g.writeFile(path, unformatted); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "\t\n") || !strings.Contains(string(got), "\t\"fmt\"\n\t\"os\"\n") || !strings.Contains(string(got), "fmt.Println(os.Args)\n} // trailing") {
		t.Fatalf("not gofmt-clean:\n%s", got)
	}

	if err := g.writeFile(filepath.Join(dir, "bad.go"), "package main\nfunc {"); err == nil {
		t.Fatal("expected error for invalid Go")
	}

	md := "#  Title\t\n"
	if err := g.writeFile(filepath.Join(dir, "README.md"), md); err != nil {
		t.Fatal(err)
	}
	got, _ = os.ReadFile(filepath.Join(dir, "README.md"))
	if string(got) != md {
		t.Fatal("non-Go file was modified")
	}
}
