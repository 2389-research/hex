// ABOUTME: Tests for templates command functionality
// ABOUTME: Validates template loading and command execution

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplatesCommand_Registered(t *testing.T) {
	// Verify command is registered
	cmd, _, err := rootCmd.Find([]string{"templates"})
	require.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "templates", cmd.Name())
}

func TestTemplatesListCommand_Registered(t *testing.T) {
	// Verify subcommand is registered
	cmd, _, err := rootCmd.Find([]string{"templates", "list"})
	require.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Name())
}

func TestLoadTemplateByName_Success(t *testing.T) {
	// Create temp templates directory
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, ".clem", "templates")
	err := os.MkdirAll(templatesDir, 0750)
	require.NoError(t, err)

	// Create a test template
	templateContent := `name: test-template
description: A test template
system_prompt: You are helpful
`
	templatePath := filepath.Join(templatesDir, "test-template.yaml")
	err = os.WriteFile(templatePath, []byte(templateContent), 0600)
	require.NoError(t, err)

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Load template
	template, err := loadTemplateByName("test-template")
	require.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "test-template", template.Name)
	assert.Equal(t, "A test template", template.Description)
	assert.Equal(t, "You are helpful", template.SystemPrompt)
}

func TestLoadTemplateByName_NotFound(t *testing.T) {
	// Create temp templates directory
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, ".clem", "templates")
	err := os.MkdirAll(templatesDir, 0750)
	require.NoError(t, err)

	// Create a different template
	templateContent := `name: other-template
description: Different template
`
	templatePath := filepath.Join(templatesDir, "other.yaml")
	err = os.WriteFile(templatePath, []byte(templateContent), 0600)
	require.NoError(t, err)

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Try to load non-existent template
	_, err = loadTemplateByName("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "other-template") // Should suggest available templates
}

func TestLoadTemplateByName_EmptyDirectory(t *testing.T) {
	// Create empty templates directory
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, ".clem", "templates")
	err := os.MkdirAll(templatesDir, 0750)
	require.NoError(t, err)

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Try to load template from empty directory
	_, err = loadTemplateByName("any-template")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "No templates available")
}

func TestCreateExampleTemplates(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, ".clem", "templates")

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Create example templates
	err := createExampleTemplates()
	require.NoError(t, err)

	// Verify templates were created
	expectedTemplates := []string{
		"code-review.yaml",
		"debug-session.yaml",
		"refactor.yaml",
	}

	for _, templateFile := range expectedTemplates {
		path := filepath.Join(templatesDir, templateFile)
		_, err := os.Stat(path)
		assert.NoError(t, err, "Template file should exist: %s", templateFile)

		// Verify it's valid YAML by reading it
		//nolint:gosec // G304: Test file reads/writes are safe
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.NotEmpty(t, content)
		assert.Contains(t, string(content), "name:")
		assert.Contains(t, string(content), "description:")
	}
}

func TestCreateExampleTemplates_DoesNotOverwrite(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, ".clem", "templates")
	err := os.MkdirAll(templatesDir, 0750)
	require.NoError(t, err)

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Create existing template with custom content
	customContent := "name: custom\ndescription: my custom template\n"
	customPath := filepath.Join(templatesDir, "code-review.yaml")
	err = os.WriteFile(customPath, []byte(customContent), 0600)
	require.NoError(t, err)

	// Call createExampleTemplates
	err = createExampleTemplates()
	require.NoError(t, err)

	//nolint:gosec // G304: Test file reads/writes are safe
	// Verify custom template was not overwritten
	content, err := os.ReadFile(customPath) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err)
	assert.Equal(t, customContent, string(content))
}

func TestTemplatesCommand_Help(t *testing.T) {
	// Test that help text is present
	assert.Contains(t, templatesCmd.Long, "Templates are loaded from")
	assert.Contains(t, templatesCmd.Long, "System prompts")
	assert.Contains(t, templatesCmd.Long, "--template")
}

func TestTemplatesListCommand_Help(t *testing.T) {
	// Test that help text is present
	assert.Contains(t, templatesListCmd.Short, "List available templates")
	assert.Contains(t, templatesListCmd.Long, "session templates")
}
