package skills

import (
	"context"
	"strings"
	"testing"
)

func TestNewSkillTool(t *testing.T) {
	registry := NewRegistry()
	tool := NewSkillTool(registry)

	if tool == nil {
		t.Fatal("NewSkillTool returned nil")
	}
	if tool.registry != registry {
		t.Error("Tool registry not set correctly")
	}
}

func TestSkillToolName(t *testing.T) {
	tool := NewSkillTool(NewRegistry())
	if tool.Name() != "Skill" {
		t.Errorf("Name() = %q; want %q", tool.Name(), "Skill")
	}
}

func TestSkillToolDescription(t *testing.T) {
	tool := NewSkillTool(NewRegistry())
	desc := tool.Description()
	if desc == "" {
		t.Error("Description is empty")
	}
	if !strings.Contains(desc, "skill") {
		t.Error("Description should mention 'skill'")
	}
}

func TestSkillToolRequiresApproval(t *testing.T) {
	tool := NewSkillTool(NewRegistry())
	if tool.RequiresApproval(nil) {
		t.Error("Skill tool should not require approval")
	}
}

func TestSkillToolExecute_MissingCommand(t *testing.T) {
	tool := NewSkillTool(NewRegistry())

	result, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected failure for missing command")
	}
	if result.Error == "" {
		t.Error("Expected error message for missing command")
	}
}

func TestSkillToolExecute_InvalidCommand(t *testing.T) {
	tool := NewSkillTool(NewRegistry())

	params := map[string]interface{}{
		"command": 123, // Not a string
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected failure for invalid command type")
	}
}

func TestSkillToolExecute_SkillNotFound(t *testing.T) {
	registry := NewRegistry()
	tool := NewSkillTool(registry)

	params := map[string]interface{}{
		"command": "nonexistent-skill",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected failure for nonexistent skill")
	}
	if !strings.Contains(result.Error, "not found") {
		t.Errorf("Error should mention 'not found', got: %s", result.Error)
	}
}

func TestSkillToolExecute_Success(t *testing.T) {
	registry := NewRegistry()
	skill := &Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Content:     "# Test Skill\n\nThis is test content.",
		Source:      "test",
		Priority:    5,
		Tags:        []string{"testing"},
		Version:     "1.0.0",
	}
	_ = registry.Register(skill)

	tool := NewSkillTool(registry)

	params := map[string]interface{}{
		"command": "test-skill",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}
	if result.Output == "" {
		t.Error("Output is empty")
	}
	if !strings.Contains(result.Output, "test-skill") {
		t.Error("Output should contain skill name")
	}
	if !strings.Contains(result.Output, "Test Skill") {
		t.Error("Output should contain skill content")
	}
}

func TestSkillToolExecute_Metadata(t *testing.T) {
	registry := NewRegistry()
	skill := &Skill{
		Name:        "test-skill",
		Description: "Test",
		Content:     "Content",
		Source:      "user",
		Priority:    7,
		Tags:        []string{"tag1", "tag2"},
		Version:     "2.0.0",
	}
	_ = registry.Register(skill)

	tool := NewSkillTool(registry)

	params := map[string]interface{}{
		"command": "test-skill",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	// Check metadata
	if result.Metadata == nil {
		t.Fatal("Metadata is nil")
	}

	if name, ok := result.Metadata["skill_name"].(string); !ok || name != "test-skill" {
		t.Errorf("Metadata skill_name = %v; want %q", result.Metadata["skill_name"], "test-skill")
	}
	if source, ok := result.Metadata["skill_source"].(string); !ok || source != "user" {
		t.Errorf("Metadata skill_source = %v; want %q", result.Metadata["skill_source"], "user")
	}
	if priority, ok := result.Metadata["priority"].(int); !ok || priority != 7 {
		t.Errorf("Metadata priority = %v; want 7", result.Metadata["priority"])
	}
	if version, ok := result.Metadata["version"].(string); !ok || version != "2.0.0" {
		t.Errorf("Metadata version = %v; want %q", result.Metadata["version"], "2.0.0")
	}
}

func TestSkillToolExecute_WithSuggestions(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Register(&Skill{Name: "test-driven-development", Description: "TDD", Content: "TDD content"})
	_ = registry.Register(&Skill{Name: "test-coverage", Description: "Coverage", Content: "Coverage content"})

	tool := NewSkillTool(registry)

	params := map[string]interface{}{
		"command": "test", // Partial match
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected failure for nonexistent exact skill")
	}

	// Should suggest similar skills
	if !strings.Contains(result.Error, "Did you mean") {
		t.Error("Error should include suggestions")
	}
	if !strings.Contains(result.Error, "test-driven-development") && !strings.Contains(result.Error, "test-coverage") {
		t.Error("Suggestions should include similar skill names")
	}
}

func TestFormatSkillForClaude(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "Test description",
		Content:     "# Skill Content\n\nDetails here.",
		Source:      "user",
		Priority:    8,
		Tags:        []string{"testing", "quality"},
		Version:     "1.2.3",
	}

	output := formatSkillForClaude(skill)

	// Check structure
	if !strings.Contains(output, "<skill name=\"test-skill\">") {
		t.Error("Output should start with skill XML tag")
	}
	if !strings.Contains(output, "</skill>") {
		t.Error("Output should end with closing skill tag")
	}

	// Check metadata
	if !strings.Contains(output, "test-skill") {
		t.Error("Output should contain skill name")
	}
	if !strings.Contains(output, "Test description") {
		t.Error("Output should contain description")
	}
	if !strings.Contains(output, "user") {
		t.Error("Output should contain source")
	}
	if !strings.Contains(output, "testing, quality") {
		t.Error("Output should contain tags")
	}
	if !strings.Contains(output, "1.2.3") {
		t.Error("Output should contain version")
	}

	// Check content
	if !strings.Contains(output, "Skill Content") {
		t.Error("Output should contain skill content")
	}
	if !strings.Contains(output, "Details here") {
		t.Error("Output should contain skill content details")
	}
}

func TestFormatSkillForClaude_DefaultPriority(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "Test",
		Content:     "Content",
		Priority:    5, // Default priority
	}

	output := formatSkillForClaude(skill)

	// Default priority should not be shown
	if strings.Contains(output, "Priority: 5") {
		t.Error("Default priority (5) should not be displayed")
	}
}

func TestFormatSkillForClaude_NoOptionalFields(t *testing.T) {
	skill := &Skill{
		Name:        "minimal-skill",
		Description: "Minimal",
		Content:     "Content only",
		Priority:    5,
	}

	output := formatSkillForClaude(skill)

	// Should still format correctly without optional fields
	if !strings.Contains(output, "minimal-skill") {
		t.Error("Output should contain skill name")
	}
	if !strings.Contains(output, "Content only") {
		t.Error("Output should contain content")
	}
}

func TestFindSimilarSkills(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Register(&Skill{Name: "test-driven-development", Description: "TDD"})
	_ = registry.Register(&Skill{Name: "testing-patterns", Description: "Patterns"})
	_ = registry.Register(&Skill{Name: "code-review", Description: "Review"})

	tool := NewSkillTool(registry)

	tests := []struct {
		query string
		want  int
	}{
		{"test", 2},        // Matches "test-driven-development" and "testing-patterns" (both contain "test")
		{"testing", 1},     // Matches "testing-patterns" only ("testing" is substring match)
		{"code", 1},        // Matches "code-review"
		{"review", 1},      // Matches "code-review"
		{"nonexistent", 0}, // No matches
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			similar := tool.findSimilarSkills(tt.query)
			if len(similar) != tt.want {
				t.Errorf("findSimilarSkills(%q) length = %d; want %d (got: %v)", tt.query, len(similar), tt.want, similar)
			}
		})
	}
}

func TestSkillToolExecute_ContextCancellation(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Register(&Skill{Name: "test-skill", Description: "Test", Content: "Content"})

	tool := NewSkillTool(registry)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	params := map[string]interface{}{
		"command": "test-skill",
	}

	// Should still work (skill loading is not async)
	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success even with cancelled context, got: %s", result.Error)
	}
}
