package generator

import (
	"os"
	"strings"
)

// preserveIndirectRequires carries the `// indirect` requirements of an
// existing go.mod at path into the freshly generated content. The template
// only emits the generator-owned direct pins; without this, `go mod tidy`
// re-resolves every indirect module to its MVS minimum and undoes any
// dependabot bump made in the project (see #402). Tidy still prunes
// entries that are no longer needed and raises ones below the minimum.
func preserveIndirectRequires(path, generated string) string {
	old, err := os.ReadFile(path)
	if err != nil {
		return generated
	}

	var keep []string
	for _, line := range strings.Split(string(old), "\n") {
		fields := strings.Fields(line)
		// ponytail: line scan instead of x/mod/modfile; go.mod require lines are `path version // indirect`
		if len(fields) < 4 || fields[2] != "//" || fields[3] != "indirect" {
			continue
		}
		if strings.Contains(generated, "\t"+fields[0]+" ") {
			continue
		}
		keep = append(keep, "\t"+fields[0]+" "+fields[1]+" // indirect")
	}
	if len(keep) == 0 {
		return generated
	}

	return strings.TrimRight(generated, "\n") + "\n\nrequire (\n" + strings.Join(keep, "\n") + "\n)\n"
}
