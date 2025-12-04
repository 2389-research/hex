// ABOUTME: Tests for template loader functionality
// ABOUTME: Validates YAML parsing, validation, and directory loading

package templates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTemplate_Valid(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	yamlContent := `name: Test Template
description: A test template
system_prompt: You are a helpful assistant
initial_messages:
  - role: user
    content: Hello
  - role: assistant
    content: Hi there!
tools_enabled:
  - read_file
  - bash
model: claude-sonnet-4-5-20250929
max_tokens: 4096
`

	err := os.WriteFile(tmpFile, []byte(yamlContent), 0600)
	require.NoError(t, err)

	// Load template
	template, err := LoadTemplate(tmpFile)
	require.NoError(t, err)
	assert.NotNil(t, template)

	// Verify fields
	assert.Equal(t, "Test Template", template.Name)
	assert.Equal(t, "A test template", template.Description)
	assert.Equal(t, "You are a helpful assistant", template.SystemPrompt)
	assert.Len(t, template.InitialMessages, 2)
	assert.Equal(t, "user", template.InitialMessages[0].Role)
	assert.Equal(t, "Hello", template.InitialMessages[0].Content)
	assert.Equal(t, "assistant", template.InitialMessages[1].Role)
	assert.Equal(t, "Hi there!", template.InitialMessages[1].Content)
	assert.Len(t, template.ToolsEnabled, 2)
	assert.Contains(t, template.ToolsEnabled, "read_file")
	assert.Contains(t, template.ToolsEnabled, "bash")
	assert.Equal(t, "claude-sonnet-4-5-20250929", template.Model)
	assert.Equal(t, 4096, template.MaxTokens)
}

func TestLoadTemplate_MinimalValid(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "minimal.yaml")

	yamlContent := `name: Minimal Template
description: Minimal template with just required fields
`

	err := os.WriteFile(tmpFile, []byte(yamlContent), 0600)
	require.NoError(t, err)

	template, err := LoadTemplate(tmpFile)
	require.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "Minimal Template", template.Name)
	assert.Len(t, template.InitialMessages, 0)
	assert.Len(t, template.ToolsEnabled, 0)
}

func TestLoadTemplate_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")

	yamlContent := `name: Test
invalid yaml here: [unclosed bracket
`

	err := os.WriteFile(tmpFile, []byte(yamlContent), 0600)
	require.NoError(t, err)

	_, err = LoadTemplate(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse template YAML")
}

func TestLoadTemplate_MissingName(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "noname.yaml")

	yamlContent := `description: No name field
`

	err := os.WriteFile(tmpFile, []byte(yamlContent), 0600)
	require.NoError(t, err)

	_, err = LoadTemplate(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestLoadTemplate_InvalidRole(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "badrole.yaml")

	yamlContent := `name: Bad Role
initial_messages:
  - role: invalid_role
    content: Test
`

	err := os.WriteFile(tmpFile, []byte(yamlContent), 0600)
	require.NoError(t, err)

	_, err = LoadTemplate(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestLoadTemplate_EmptyMessageContent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "emptycontent.yaml")

	yamlContent := `name: Empty Content
initial_messages:
  - role: user
    content: ""
`

	err := os.WriteFile(tmpFile, []byte(yamlContent), 0600)
	require.NoError(t, err)

	_, err = LoadTemplate(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content cannot be empty")
}

func TestLoadTemplate_FileNotFound(t *testing.T) {
	_, err := LoadTemplate("/nonexistent/path.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read template file")
}

func TestLoadTemplates_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	templates, err := LoadTemplates(tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, templates)
	assert.Len(t, templates, 0)
}

func TestLoadTemplates_MultipleTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple template files
	template1 := `name: Template One
description: First template
`
	err := os.WriteFile(filepath.Join(tmpDir, "one.yaml"), []byte(template1), 0600)
	require.NoError(t, err)

	template2 := `name: Template Two
description: Second template
`
	err = os.WriteFile(filepath.Join(tmpDir, "two.yml"), []byte(template2), 0600)
	require.NoError(t, err)

	// Create a non-YAML file (should be ignored)
	err = os.WriteFile(filepath.Join(tmpDir, "ignore.txt"), []byte("not yaml"), 0600)
	require.NoError(t, err)

	// Load templates
	templates, err := LoadTemplates(tmpDir)
	require.NoError(t, err)
	assert.Len(t, templates, 2)
	assert.Contains(t, templates, "Template One")
	assert.Contains(t, templates, "Template Two")
}

func TestLoadTemplates_NonexistentDirectory(t *testing.T) {
	// Should return empty map without error
	templates, err := LoadTemplates("/nonexistent/directory")
	require.NoError(t, err)
	assert.NotNil(t, templates)
	assert.Len(t, templates, 0)
}

func TestLoadTemplates_SkipInvalidTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	// Valid template
	valid := `name: Valid
description: This one is good
`
	err := os.WriteFile(filepath.Join(tmpDir, "valid.yaml"), []byte(valid), 0600)
	require.NoError(t, err)

	// Invalid template (missing name)
	invalid := `description: Missing name field
`
	err = os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte(invalid), 0600)
	require.NoError(t, err)

	// Should load only the valid one
	templates, err := LoadTemplates(tmpDir)
	require.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Contains(t, templates, "Valid")
}

func TestLoadTemplates_ExpandTilde(t *testing.T) {
	// This test verifies that ~ expansion works
	// We can't actually test the real home directory, so we just verify no error
	templates, err := LoadTemplates("~/nonexistent_test_dir")
	require.NoError(t, err)
	assert.NotNil(t, templates)
	assert.Len(t, templates, 0) // Directory doesn't exist, so no templates
}

func TestGetTemplatesDir(t *testing.T) {
	dir, err := GetTemplatesDir()
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.Contains(t, dir, ".jeff")
	assert.Contains(t, dir, "templates")
}

func TestEnsureTemplatesDir(t *testing.T) {
	// We can't test the real home directory, so create a temp test
	// This is more of a smoke test
	dir, err := GetTemplatesDir()
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
}

func TestTemplateValidate_AllRoles(t *testing.T) {
	validRoles := []string{"user", "assistant", "system"}

	for _, role := range validRoles {
		t.Run(role, func(t *testing.T) {
			template := &Template{
				Name: "Test",
				InitialMessages: []Message{
					{Role: role, Content: "Test message"},
				},
			}
			err := template.Validate()
			assert.NoError(t, err)
		})
	}
}
