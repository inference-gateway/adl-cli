package registry

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter holds the parsed YAML frontmatter of a skill markdown file.
type Frontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags,omitempty"`
	Version     string   `yaml:"version,omitempty"`
	License     string   `yaml:"license,omitempty"`
}

// SkillDocument is a parsed skill markdown: frontmatter plus body.
type SkillDocument struct {
	Frontmatter Frontmatter
	Body        string
	Raw         []byte
}

// ParseSkillDocument parses a skill markdown blob. The frontmatter must be
// the first content in the file, delimited by lines of "---". A skill
// markdown without a frontmatter block is rejected.
func ParseSkillDocument(content []byte) (*SkillDocument, error) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	trimmed := bytes.TrimPrefix(content, bom)
	trimmed = bytes.TrimLeft(trimmed, "\r\n\t ")
	if !bytes.HasPrefix(trimmed, []byte("---")) {
		return nil, fmt.Errorf("skill markdown is missing a YAML frontmatter block")
	}

	rest := trimmed[3:]
	rest = bytes.TrimLeft(rest, "\r\n")

	closeIdx := indexClosingFence(rest)
	if closeIdx < 0 {
		return nil, fmt.Errorf("skill markdown frontmatter is not closed (expected a line of '---')")
	}

	frontmatterBytes := rest[:closeIdx]
	body := rest[closeIdx:]
	body = trimAfterFence(body)

	var fm Frontmatter
	if err := yaml.Unmarshal(frontmatterBytes, &fm); err != nil {
		return nil, fmt.Errorf("failed to parse skill frontmatter: %w", err)
	}

	if fm.Name == "" {
		return nil, fmt.Errorf("skill frontmatter is missing required field 'name'")
	}
	if fm.Description == "" {
		return nil, fmt.Errorf("skill frontmatter is missing required field 'description'")
	}

	return &SkillDocument{
		Frontmatter: fm,
		Body:        string(body),
		Raw:         content,
	}, nil
}

// indexClosingFence returns the offset (in s) of the start of the line
// containing the closing "---" fence, or -1 if no such line exists.
func indexClosingFence(s []byte) int {
	lines := strings.Split(string(s), "\n")
	offset := 0
	for _, line := range lines {
		stripped := strings.TrimRight(line, "\r")
		if stripped == "---" || stripped == "..." {
			return offset
		}
		offset += len(line) + 1
	}
	return -1
}

// trimAfterFence skips the closing fence line and any single trailing newline.
func trimAfterFence(s []byte) []byte {
	if i := bytes.IndexByte(s, '\n'); i >= 0 {
		return s[i+1:]
	}
	return nil
}
