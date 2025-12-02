// ABOUTME: Skill tool implementation for invoking skills from Claude
// ABOUTME: Loads and returns skill content to extend Claude's context

package skills

import (
	"context"
	"fmt"
	"strings"
)

// ToolResult represents the output of a skill tool execution
type ToolResult struct {
	ToolName string
	Success  bool
	Output   string
	Error    string
	Metadata map[string]interface{}
}

// SkillTool allows Claude to invoke skills
type SkillTool struct {
	registry *Registry
}

// NewSkillTool creates a new skill tool with the given registry
func NewSkillTool(registry *Registry) *SkillTool {
	return &SkillTool{
		registry: registry,
	}
}

// Name returns the tool name
func (t *SkillTool) Name() string {
	return "Skill"
}

// Description returns the tool description
func (t *SkillTool) Description() string {
	return "Load and invoke a skill to extend Claude's knowledge and capabilities. Skills provide specialized knowledge, processes, and best practices for specific domains. Parameters: command (required) - skill name to invoke"
}

// RequiresApproval returns false - skills are safe to load
func (t *SkillTool) RequiresApproval(_ map[string]interface{}) bool {
	// Skills are just knowledge/documentation, safe to load without approval
	return false
}

// Execute loads and returns a skill's content
func (t *SkillTool) Execute(_ context.Context, params map[string]interface{}) (*ToolResult, error) {
	// Validate command parameter
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return &ToolResult{
			ToolName: "Skill",
			Success:  false,
			Error:    "missing or invalid 'command' parameter (must be non-empty string with skill name)",
		}, nil
	}

	// Get skill from registry
	skill, err := t.registry.Get(command)
	if err != nil {
		// Skill not found - provide helpful error with suggestions
		suggestions := t.findSimilarSkills(command)
		errorMsg := fmt.Sprintf("Skill '%s' not found", command)
		if len(suggestions) > 0 {
			errorMsg += fmt.Sprintf(". Did you mean: %s?", strings.Join(suggestions, ", "))
		}
		errorMsg += fmt.Sprintf("\n\nAvailable skills: %s", strings.Join(t.registry.List(), ", "))

		return &ToolResult{
			ToolName: "Skill",
			Success:  false,
			Error:    errorMsg,
			Metadata: map[string]interface{}{
				"available_skills": t.registry.List(),
				"suggestions":      suggestions,
			},
		}, nil
	}

	// Format skill output for Claude
	output := formatSkillForClaude(skill)

	return &ToolResult{
		ToolName: "Skill",
		Success:  true,
		Output:   output,
		Metadata: map[string]interface{}{
			"skill_name":   skill.Name,
			"skill_source": skill.Source,
			"priority":     skill.Priority,
			"tags":         skill.Tags,
			"version":      skill.Version,
		},
	}, nil
}

// findSimilarSkills finds skills with similar names (simple fuzzy matching)
func (t *SkillTool) findSimilarSkills(query string) []string {
	allSkills := t.registry.List()
	var similar []string

	lowerQuery := strings.ToLower(query)
	for _, name := range allSkills {
		lowerName := strings.ToLower(name)
		// Bidirectional substring matching
		if strings.Contains(lowerName, lowerQuery) || strings.Contains(lowerQuery, lowerName) {
			similar = append(similar, name)
		}
	}

	return similar
}

// formatSkillForClaude formats skill content for Claude's context
func formatSkillForClaude(skill *Skill) string {
	var sb strings.Builder

	// XML-style wrapper for clarity
	sb.WriteString(fmt.Sprintf("<skill name=%q>\n", skill.Name))

	// Add metadata header
	sb.WriteString("## Skill Metadata\n")
	sb.WriteString(fmt.Sprintf("- **Name**: %s\n", skill.Name))
	sb.WriteString(fmt.Sprintf("- **Description**: %s\n", skill.Description))
	sb.WriteString(fmt.Sprintf("- **Source**: %s\n", skill.Source))
	if skill.Priority != 5 {
		sb.WriteString(fmt.Sprintf("- **Priority**: %d\n", skill.Priority))
	}
	if len(skill.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("- **Tags**: %s\n", strings.Join(skill.Tags, ", ")))
	}
	if skill.Version != "" {
		sb.WriteString(fmt.Sprintf("- **Version**: %s\n", skill.Version))
	}
	sb.WriteString("\n")

	// Add skill content
	sb.WriteString("## Skill Content\n\n")
	sb.WriteString(skill.Content)
	sb.WriteString("\n\n")

	// Close wrapper
	sb.WriteString("</skill>\n")

	return sb.String()
}
