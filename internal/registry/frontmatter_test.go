package registry

import (
	"strings"
	"testing"
)

func TestParseSkillDocument(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantErr  string
		wantName string
		wantTags []string
		wantBody string
	}{
		{
			name: "well-formed frontmatter and body",
			input: `---
name: data-analysis
description: Analyze tabular data
tags:
  - analytics
  - data
---

# Data analysis

Use this skill when the user asks about trends.
`,
			wantName: "data-analysis",
			wantTags: []string{"analytics", "data"},
			wantBody: "\n# Data analysis\n\nUse this skill when the user asks about trends.\n",
		},
		{
			name:     "frontmatter with BOM is accepted",
			input:    "\ufeff---\nname: report-writing\ndescription: Write business reports\n---\nbody\n",
			wantName: "report-writing",
			wantBody: "body\n",
		},
		{
			name:    "missing frontmatter",
			input:   "# just a heading\nno frontmatter here",
			wantErr: "missing a YAML frontmatter block",
		},
		{
			name: "unclosed frontmatter",
			input: `---
name: oops
description: never closed
body that should have been after the fence
`,
			wantErr: "frontmatter is not closed",
		},
		{
			name: "missing name field",
			input: `---
description: no name
---
body
`,
			wantErr: "missing required field 'name'",
		},
		{
			name: "missing description field",
			input: `---
name: no-desc
---
body
`,
			wantErr: "missing required field 'description'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := ParseSkillDocument([]byte(tc.input))
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if doc.Frontmatter.Name != tc.wantName {
				t.Errorf("name = %q, want %q", doc.Frontmatter.Name, tc.wantName)
			}
			if tc.wantTags != nil {
				if len(doc.Frontmatter.Tags) != len(tc.wantTags) {
					t.Errorf("tags = %v, want %v", doc.Frontmatter.Tags, tc.wantTags)
				} else {
					for i, tag := range tc.wantTags {
						if doc.Frontmatter.Tags[i] != tag {
							t.Errorf("tags[%d] = %q, want %q", i, doc.Frontmatter.Tags[i], tag)
						}
					}
				}
			}
			if tc.wantBody != "" && doc.Body != tc.wantBody {
				t.Errorf("body = %q, want %q", doc.Body, tc.wantBody)
			}
		})
	}
}
