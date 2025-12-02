// ABOUTME: Tests for SlashCommand tool execution
// ABOUTME: Validates command invocation, argument handling, and error cases

package commands

import (
	"context"
	"strings"
	"testing"
)

func TestSlashCommandToolName(t *testing.T) {
	registry := NewRegistry()
	tool := NewSlashCommandTool(registry)

	if tool.Name() != "SlashCommand" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "SlashCommand")
	}
}

func TestSlashCommandToolDescription(t *testing.T) {
	registry := NewRegistry()
	tool := NewSlashCommandTool(registry)

	desc := tool.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}

	// Should mention parameters
	if !strings.Contains(desc, "command") {
		t.Error("Description should mention 'command' parameter")
	}
}

func TestSlashCommandToolRequiresApproval(t *testing.T) {
	registry := NewRegistry()
	tool := NewSlashCommandTool(registry)

	// Slash commands should not require approval
	if tool.RequiresApproval(nil) {
		t.Error("RequiresApproval() = true, want false")
	}

	if tool.RequiresApproval(map[string]interface{}{"command": "test"}) {
		t.Error("RequiresApproval() = true, want false")
	}
}

func TestExecuteMissingCommand(t *testing.T) {
	registry := NewRegistry()
	tool := NewSlashCommandTool(registry)
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Success {
		t.Error("Result.Success = true, want false")
	}

	if !strings.Contains(result.Error, "missing or invalid") {
		t.Errorf("Error = %q, should mention missing parameter", result.Error)
	}
}

func TestExecuteCommandNotFound(t *testing.T) {
	registry := NewRegistry()
	tool := NewSlashCommandTool(registry)
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "nonexistent",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Success {
		t.Error("Result.Success = true, want false")
	}

	if !strings.Contains(result.Error, "not found") {
		t.Errorf("Error = %q, should mention command not found", result.Error)
	}

	// Should list available commands
	if !strings.Contains(result.Error, "Available commands") {
		t.Error("Error should list available commands")
	}
}

func TestExecuteSuccess(t *testing.T) {
	registry := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Description: "Test command",
		Content:     "This is a test command.",
	}
	_ = registry.Register(cmd) //nolint:errcheck // test setup

	tool := NewSlashCommandTool(registry)
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "test",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Result.Success = false, error = %q", result.Error)
	}

	// Should contain command-message indicator
	if !strings.Contains(result.Output, "<command-message>") {
		t.Error("Output should contain <command-message> indicator")
	}

	// Should contain command content
	if !strings.Contains(result.Output, "This is a test command") {
		t.Error("Output should contain command content")
	}

	// Check metadata
	if metadata, ok := result.Metadata["command"].(string); !ok || metadata != "test" {
		t.Errorf("Metadata[command] = %v, want %q", result.Metadata["command"], "test")
	}
}

func TestExecuteWithArgs(t *testing.T) {
	registry := NewRegistry()

	cmd := &Command{
		Name:        "greet",
		Description: "Greet someone",
		Content:     "Hello, {{.name}}!",
		Args: map[string]string{
			"name": "Name to greet",
		},
	}
	_ = registry.Register(cmd) //nolint:errcheck // test setup

	tool := NewSlashCommandTool(registry)
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "greet",
		"args": map[string]interface{}{
			"name": "World",
		},
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Result.Success = false, error = %q", result.Error)
	}

	// Should contain expanded template
	if !strings.Contains(result.Output, "Hello, World!") {
		t.Errorf("Output = %q, should contain expanded template", result.Output)
	}
}

func TestExecuteWithSlashPrefix(t *testing.T) {
	registry := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Description: "Test command",
		Content:     "Test content",
	}
	_ = registry.Register(cmd) //nolint:errcheck // test setup

	tool := NewSlashCommandTool(registry)
	ctx := context.Background()

	// Command name with leading slash should work
	params := map[string]interface{}{
		"command": "/test",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Result.Success = false, error = %q", result.Error)
	}
}

func TestExecuteTemplateError(t *testing.T) {
	registry := NewRegistry()

	cmd := &Command{
		Name:        "broken",
		Description: "Broken template",
		Content:     "{{.invalid syntax",
	}
	_ = registry.Register(cmd) //nolint:errcheck // test setup

	tool := NewSlashCommandTool(registry)
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "broken",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Success {
		t.Error("Result.Success = true, want false for template error")
	}

	if !strings.Contains(result.Error, "template") {
		t.Errorf("Error = %q, should mention template error", result.Error)
	}
}

func TestListCommands(t *testing.T) {
	registry := NewRegistry()

	commands := []*Command{
		{Name: "cmd1", Description: "First command"},
		{Name: "cmd2", Description: "Second command", Args: map[string]string{"file": "File to process"}},
	}
	_ = registry.RegisterAll(commands) //nolint:errcheck // test setup

	tool := NewSlashCommandTool(registry)
	ctx := context.Background()

	params := map[string]interface{}{
		"list": true,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Result.Success = false, error = %q", result.Error)
	}

	// Should contain both command names
	if !strings.Contains(result.Output, "cmd1") || !strings.Contains(result.Output, "cmd2") {
		t.Error("Output should list both commands")
	}

	// Should show descriptions
	if !strings.Contains(result.Output, "First command") {
		t.Error("Output should show command descriptions")
	}

	// Should show args for cmd2
	if !strings.Contains(result.Output, "file") {
		t.Error("Output should show command arguments")
	}

	// Check metadata
	if count, ok := result.Metadata["command_count"].(int); !ok || count != 2 {
		t.Errorf("Metadata[command_count] = %v, want 2", result.Metadata["command_count"])
	}
}

func TestFindSimilarCommands(t *testing.T) {
	registry := NewRegistry()

	commands := []*Command{
		{Name: "review", Description: "Code review"},
		{Name: "review-security", Description: "Security review"},
		{Name: "test", Description: "Run tests"},
	}
	_ = registry.RegisterAll(commands) //nolint:errcheck // test setup

	tool := NewSlashCommandTool(registry)

	tests := []struct {
		query string
		want  []string
	}{
		{
			query: "rev",
			want:  []string{"review", "review-security"},
		},
		{
			query: "security",
			want:  []string{"review-security"},
		},
		{
			query: "test",
			want:  []string{"test"},
		},
		{
			query: "nonexistent",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			similar := tool.findSimilarCommands(tt.query)

			if len(similar) != len(tt.want) {
				t.Errorf("findSimilarCommands(%q) returned %d items, want %d", tt.query, len(similar), len(tt.want))
				return
			}

			// Convert to map for easier checking
			gotMap := make(map[string]bool)
			for _, s := range similar {
				gotMap[s] = true
			}

			for _, want := range tt.want {
				if !gotMap[want] {
					t.Errorf("findSimilarCommands(%q) missing %q", tt.query, want)
				}
			}
		})
	}
}

func TestExecuteWithSuggestions(t *testing.T) {
	registry := NewRegistry()

	commands := []*Command{
		{Name: "review", Description: "Code review"},
		{Name: "review-security", Description: "Security review"},
	}
	_ = registry.RegisterAll(commands) //nolint:errcheck // test setup

	tool := NewSlashCommandTool(registry)
	ctx := context.Background()

	params := map[string]interface{}{
		"command": "rev",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Success {
		t.Error("Result.Success = true, want false")
	}

	// Should show suggestions
	if !strings.Contains(result.Error, "Did you mean") {
		t.Error("Error should provide suggestions")
	}

	// Check metadata has suggestions
	if suggestions, ok := result.Metadata["suggestions"].([]string); !ok || len(suggestions) == 0 {
		t.Error("Metadata should contain suggestions")
	}
}
