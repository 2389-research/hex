// ABOUTME: Tests for YAML frontmatter parsing
// ABOUTME: Verifies frontmatter extraction and safety limits

package frontmatter

import (
	"bytes"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantFrontmatter string
		wantContent     string
		wantErr         bool
	}{
		{
			name: "basic frontmatter",
			input: `---
name: test
description: test skill
---
# Content here
`,
			wantFrontmatter: "name: test\ndescription: test skill",
			wantContent:     "# Content here\n",
			wantErr:         false,
		},
		{
			name: "no frontmatter",
			input: `# Just content
No frontmatter here
`,
			wantFrontmatter: "",
			wantContent: `# Just content
No frontmatter here
`,
			wantErr: false,
		},
		{
			name: "empty frontmatter",
			input: `---
---
Content only
`,
			wantFrontmatter: "",
			wantContent:     "Content only\n",
			wantErr:         false,
		},
		{
			name: "unclosed frontmatter",
			input: `---
name: test
description: test
no closing delimiter
`,
			wantFrontmatter: "",
			wantContent:     "",
			wantErr:         true,
		},
		{
			name:            "frontmatter bomb protection",
			input:           "---\n" + strings.Repeat("line: value\n", 200) + "---\ncontent",
			wantFrontmatter: "",
			wantContent:     "",
			wantErr:         true,
		},
		{
			name:  "windows line endings",
			input: "---\r\nname: test\r\n---\r\nContent\r\n",
			// Note: After splitting by \n, \r will remain in the lines
			wantFrontmatter: "name: test\r",
			wantContent:     "Content\r\n",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter, content, err := Split([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Split() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotFrontmatter := string(frontmatter)
				gotContent := string(content)

				if gotFrontmatter != tt.wantFrontmatter {
					t.Errorf("Split() frontmatter = %q, want %q", gotFrontmatter, tt.wantFrontmatter)
				}

				if gotContent != tt.wantContent {
					t.Errorf("Split() content = %q, want %q", gotContent, tt.wantContent)
				}
			}
		})
	}
}

func TestSplitLargeButValidFrontmatter(t *testing.T) {
	// Test that frontmatter just under the limit works
	lines := []string{"---"}
	for i := 0; i < MaxFrontmatterLines-2; i++ {
		lines = append(lines, "key: value")
	}
	lines = append(lines, "---", "Content here")
	input := strings.Join(lines, "\n")

	_, content, err := Split([]byte(input))
	if err != nil {
		t.Errorf("Split() should succeed for frontmatter under limit, got error: %v", err)
	}

	if !bytes.Contains(content, []byte("Content here")) {
		t.Errorf("Split() should extract content correctly")
	}
}
