// ABOUTME: Adapter to bridge SkillTool to tools.Tool interface
// ABOUTME: Converts between skills.ToolResult and tools.Result

package skills

import (
	"context"

	"github.com/harper/clem/internal/tools"
)

// ToolAdapter adapts SkillTool to implement tools.Tool interface
type ToolAdapter struct {
	skillTool *SkillTool
}

// NewToolAdapter creates a new skill tool adapter
func NewToolAdapter(registry *Registry) *ToolAdapter {
	return &ToolAdapter{
		skillTool: NewSkillTool(registry),
	}
}

// Name returns the tool name
func (a *ToolAdapter) Name() string {
	return a.skillTool.Name()
}

// Description returns the tool description
func (a *ToolAdapter) Description() string {
	return a.skillTool.Description()
}

// RequiresApproval returns whether the tool requires user approval
func (a *ToolAdapter) RequiresApproval(params map[string]interface{}) bool {
	return a.skillTool.RequiresApproval(params)
}

// Execute runs the skill tool and converts the result
func (a *ToolAdapter) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	skillResult, err := a.skillTool.Execute(ctx, params)
	if err != nil {
		return nil, err
	}

	// Convert SkillToolResult to tools.Result
	return &tools.Result{
		ToolName: skillResult.ToolName,
		Success:  skillResult.Success,
		Output:   skillResult.Output,
		Error:    skillResult.Error,
		Metadata: skillResult.Metadata,
	}, nil
}
