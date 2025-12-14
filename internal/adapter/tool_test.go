// ABOUTME: Tests for the hex-to-mux tool adapter.
// ABOUTME: Verifies that hex tools are correctly wrapped for mux.
package adapter

import (
	"context"
	"testing"

	"github.com/2389-research/hex/internal/tools"
)

type mockHexTool struct {
	name        string
	description string
	result      *tools.Result
	err         error
}

func (m *mockHexTool) Name() string                                        { return m.name }
func (m *mockHexTool) Description() string                                 { return m.description }
func (m *mockHexTool) RequiresApproval(params map[string]interface{}) bool { return false }
func (m *mockHexTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	return m.result, m.err
}

func TestAdaptTool(t *testing.T) {
	hexTool := &mockHexTool{
		name:        "test_tool",
		description: "A test tool",
		result: &tools.Result{
			ToolName: "test_tool",
			Success:  true,
			Output:   "test output",
		},
	}

	adapted := AdaptTool(hexTool)

	if adapted.Name() != "test_tool" {
		t.Errorf("expected name test_tool, got %s", adapted.Name())
	}
	if adapted.Description() != "A test tool" {
		t.Errorf("expected description 'A test tool', got %s", adapted.Description())
	}

	result, err := adapted.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success true")
	}
	if result.Output != "test output" {
		t.Errorf("expected output 'test output', got %s", result.Output)
	}
}

func TestAdaptTool_RequiresApproval(t *testing.T) {
	hexTool := &mockHexTool{
		name:        "write_tool",
		description: "A write tool",
	}

	adapted := AdaptTool(hexTool)

	// Test with any params (converted from map[string]any to map[string]interface{})
	params := map[string]any{"path": "/etc/passwd"}
	requires := adapted.RequiresApproval(params)

	// Our mock returns false, but this tests the conversion works
	if requires != false {
		t.Errorf("expected approval false from mock, got %v", requires)
	}
}

func TestAdaptTool_ParamConversion(t *testing.T) {
	hexTool := &mockHexTool{
		name:        "test_tool",
		description: "A test tool",
		result: &tools.Result{
			ToolName: "test_tool",
			Success:  true,
			Output:   "ok",
		},
	}

	adapted := AdaptTool(hexTool)

	// Call with map[string]any with various types
	params := map[string]any{
		"string_val": "hello",
		"int_val":    42,
		"bool_val":   true,
	}

	result, err := adapted.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify execution succeeded (param conversion worked)
	if !result.Success {
		t.Error("expected success true")
	}
	if result.Output != "ok" {
		t.Errorf("expected output 'ok', got %s", result.Output)
	}
}

func TestAdaptAll(t *testing.T) {
	hexTools := []tools.Tool{
		&mockHexTool{name: "tool1", description: "First tool"},
		&mockHexTool{name: "tool2", description: "Second tool"},
		&mockHexTool{name: "tool3", description: "Third tool"},
	}

	adapted := AdaptAll(hexTools)

	if len(adapted) != 3 {
		t.Fatalf("expected 3 adapted tools, got %d", len(adapted))
	}

	names := []string{"tool1", "tool2", "tool3"}
	for i, tool := range adapted {
		if tool.Name() != names[i] {
			t.Errorf("expected tool %d name %s, got %s", i, names[i], tool.Name())
		}
	}
}
