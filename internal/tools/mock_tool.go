// ABOUTME: Mock tool implementation for testing
// ABOUTME: Provides configurable tool behavior for unit tests

package tools

import "context"

// MockTool is a simple tool for testing
type MockTool struct {
	NameValue             string
	DescriptionValue      string
	RequiresApprovalValue bool
	ExecuteFunc           func(context.Context, map[string]interface{}) (*Result, error)
}

// Name returns the tool name
func (m *MockTool) Name() string {
	return m.NameValue
}

// Description returns the tool description
func (m *MockTool) Description() string {
	return m.DescriptionValue
}

// RequiresApproval returns whether approval is needed
func (m *MockTool) RequiresApproval(params map[string]interface{}) bool {
	return m.RequiresApprovalValue
}

// Execute runs the mock tool
func (m *MockTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, params)
	}

	// Default behavior: return success
	return &Result{
		ToolName: m.NameValue,
		Success:  true,
		Output:   "mock output",
	}, nil
}
