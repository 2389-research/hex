// ABOUTME: Tests for command parsing and template expansion
// ABOUTME: Validates YAML frontmatter parsing and template rendering

package commands

import (
	"strings"
	"testing"

	"github.com/2389-research/hex/internal/frontmatter"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(*testing.T, *Command)
	}{
		{
			name: "valid command with frontmatter",
			content: `---
name: test-command
description: A test command
args:
  file: File to test
---

This is the command content.
`,
			wantErr: false,
			check: func(t *testing.T, cmd *Command) {
				if cmd.Name != "test-command" {
					t.Errorf("Name = %q, want %q", cmd.Name, "test-command")
				}
				if cmd.Description != "A test command" {
					t.Errorf("Description = %q, want %q", cmd.Description, "A test command")
				}
				if len(cmd.Args) != 1 {
					t.Errorf("Args count = %d, want 1", len(cmd.Args))
				}
				if cmd.Args["file"] != "File to test" {
					t.Errorf("Args[file] = %q, want %q", cmd.Args["file"], "File to test")
				}
				if !strings.Contains(cmd.Content, "This is the command content") {
					t.Errorf("Content doesn't contain expected text")
				}
			},
		},
		{
			name: "command without frontmatter",
			content: `This is just plain content.
No frontmatter here.`,
			wantErr: true, // Should fail - missing name and description
		},
		{
			name: "missing name field",
			content: `---
description: A test command
---

Content here.
`,
			wantErr: true,
		},
		{
			name: "missing description field",
			content: `---
name: test-command
---

Content here.
`,
			wantErr: true,
		},
		{
			name: "unclosed frontmatter",
			content: `---
name: test-command
description: A test command

Content without closing delimiter.
`,
			wantErr: true,
		},
		{
			name: "empty args map",
			content: `---
name: test-command
description: A test command
args: {}
---

Content here.
`,
			wantErr: false,
			check: func(t *testing.T, cmd *Command) {
				if len(cmd.Args) != 0 {
					t.Errorf("Args count = %d, want 0", len(cmd.Args))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseBytes("test.md", []byte(tt.content))
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, cmd)
			}
		})
	}
}

func TestExpand(t *testing.T) {
	tests := []struct {
		name    string
		content string
		args    map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "simple template",
			content: "Hello {{.name}}!",
			args:    map[string]interface{}{"name": "World"},
			want:    "Hello World!",
			wantErr: false,
		},
		{
			name:    "conditional template",
			content: "{{if .feature}}Feature: {{.feature}}{{else}}No feature{{end}}",
			args:    map[string]interface{}{"feature": "auth"},
			want:    "Feature: auth",
			wantErr: false,
		},
		{
			name:    "conditional template - false",
			content: "{{if .feature}}Feature: {{.feature}}{{else}}No feature{{end}}",
			args:    map[string]interface{}{},
			want:    "No feature",
			wantErr: false,
		},
		{
			name:    "missing variable",
			content: "Hello {{.missing}}!",
			args:    map[string]interface{}{},
			want:    "Hello <no value>!",
			wantErr: false,
		},
		{
			name:    "invalid template syntax",
			content: "Hello {{.name",
			args:    map[string]interface{}{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "no template variables",
			content: "Static content only.",
			args:    map[string]interface{}{},
			want:    "Static content only.",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:    "test",
				Content: tt.content,
			}

			got, err := cmd.Expand(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasArgs(t *testing.T) {
	tests := []struct {
		name string
		args map[string]string
		want bool
	}{
		{
			name: "has args",
			args: map[string]string{"file": "File to test"},
			want: true,
		},
		{
			name: "no args",
			args: map[string]string{},
			want: false,
		},
		{
			name: "nil args",
			args: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{Args: tt.args}
			if got := cmd.HasArgs(); got != tt.want {
				t.Errorf("HasArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsageString(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		args    map[string]string
		want    string
	}{
		{
			name:    "command with args",
			cmdName: "review",
			args:    map[string]string{"file": "File to review"},
			want:    "/review <file>",
		},
		{
			name:    "command without args",
			cmdName: "plan",
			args:    map[string]string{},
			want:    "/plan",
		},
		{
			name:    "command with multiple args",
			cmdName: "test",
			args:    map[string]string{"target": "Target", "type": "Type"},
			// Note: map iteration order is not guaranteed, so we just check it contains both
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name: tt.cmdName,
				Args: tt.args,
			}
			got := cmd.UsageString()

			if tt.name != "command with multiple args" {
				if got != tt.want {
					t.Errorf("UsageString() = %q, want %q", got, tt.want)
				}
			} else {
				// For multiple args, just verify format
				if !strings.HasPrefix(got, "/"+tt.cmdName) {
					t.Errorf("UsageString() = %q, should start with /%s", got, tt.cmdName)
				}
				if !strings.Contains(got, "<target>") || !strings.Contains(got, "<type>") {
					t.Errorf("UsageString() = %q, should contain both <target> and <type>", got)
				}
			}
		})
	}
}

func TestSplitFrontmatter(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantFrontmatter string
		wantContent     string
		wantErr         bool
	}{
		{
			name: "valid frontmatter",
			input: `---
name: test
---
content here`,
			wantFrontmatter: "name: test",
			wantContent:     "content here",
			wantErr:         false,
		},
		{
			name:            "no frontmatter",
			input:           "just content",
			wantFrontmatter: "",
			wantContent:     "just content",
			wantErr:         false,
		},
		{
			name: "unclosed frontmatter",
			input: `---
name: test
content without closing`,
			wantErr: true,
		},
		{
			name: "empty frontmatter",
			input: `---
---
content here`,
			wantFrontmatter: "",
			wantContent:     "content here",
			wantErr:         false,
		},
		{
			name: "multiline content",
			input: `---
name: test
---
line 1
line 2
line 3`,
			wantFrontmatter: "name: test",
			wantContent:     "line 1\nline 2\nline 3",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, content, err := frontmatter.Split([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			gotFrontmatter := strings.TrimSpace(string(fm))
			gotContent := strings.TrimSpace(string(content))
			wantFrontmatter := strings.TrimSpace(tt.wantFrontmatter)
			wantContent := strings.TrimSpace(tt.wantContent)

			if gotFrontmatter != wantFrontmatter {
				t.Errorf("frontmatter = %q, want %q", gotFrontmatter, wantFrontmatter)
			}
			if gotContent != wantContent {
				t.Errorf("content = %q, want %q", gotContent, wantContent)
			}
		})
	}
}
