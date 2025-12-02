package components

import (
	"strings"
	"testing"
)

func TestNewProgress(t *testing.T) {
	tests := []struct {
		name         string
		progressType ProgressType
		width        int
	}{
		{"streaming", ProgressTypeStreaming, 80},
		{"tool execution", ProgressTypeToolExecution, 60},
		{"token usage", ProgressTypeTokenUsage, 40},
		{"batch", ProgressTypeBatch, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := NewProgress(tt.progressType, tt.width)

			if prog == nil {
				t.Fatal("NewProgress returned nil")
			}

			if prog.width != tt.width {
				t.Errorf("Expected width %d, got %d", tt.width, prog.width)
			}

			if prog.progressType != tt.progressType {
				t.Errorf("Expected type %v, got %v", tt.progressType, prog.progressType)
			}

			if prog.theme == nil {
				t.Error("Theme not initialized")
			}
		})
	}
}

func TestProgressSetLabel(t *testing.T) {
	prog := NewProgress(ProgressTypeStreaming, 80)
	label := "Test Label"

	prog.SetLabel(label)

	if prog.label != label {
		t.Errorf("Expected label '%s', got '%s'", label, prog.label)
	}

	view := prog.View()
	if !strings.Contains(view, label) {
		t.Error("Label not rendered in view")
	}
}

func TestProgressSetProgress(t *testing.T) {
	prog := NewProgress(ProgressTypeStreaming, 80)

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},
		{0.0, 0.0},
		{1.0, 1.0},
		{-0.1, 0.0}, // Clamp to 0
		{1.5, 1.0},  // Clamp to 1
	}

	for _, tt := range tests {
		prog.SetProgress(tt.input)
		if prog.value != tt.expected {
			t.Errorf("SetProgress(%f): expected %f, got %f", tt.input, tt.expected, prog.value)
		}
	}
}

func TestProgressSetProgressValues(t *testing.T) {
	prog := NewProgress(ProgressTypeStreaming, 80)

	prog.SetProgressValues(50, 100)

	// SetProgressValues stores the raw values
	if prog.value != 0.5 {
		t.Errorf("Expected value 0.5 (calculated from 50/100), got %f", prog.value)
	}

	if prog.total != 100 {
		t.Errorf("Expected total 100, got %f", prog.total)
	}

	// Should calculate progress as 50%
	if prog.GetProgress() != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", prog.GetProgress())
	}
}

func TestProgressIncrementProgress(t *testing.T) {
	prog := NewProgress(ProgressTypeStreaming, 80)

	prog.SetProgress(0.3)
	prog.IncrementProgress(0.2)

	if prog.value != 0.5 {
		t.Errorf("Expected value 0.5 after increment, got %f", prog.value)
	}

	// Test clamping
	prog.SetProgress(0.9)
	prog.IncrementProgress(0.5)

	if prog.value != 1.0 {
		t.Errorf("Expected value clamped to 1.0, got %f", prog.value)
	}
}

func TestProgressSetWidth(t *testing.T) {
	prog := NewProgress(ProgressTypeStreaming, 80)

	prog.SetWidth(100)

	if prog.width != 100 {
		t.Errorf("Expected width 100, got %d", prog.width)
	}

	if prog.progress.Width != 100 {
		t.Errorf("Expected underlying progress width 100, got %d", prog.progress.Width)
	}
}

func TestProgressIsComplete(t *testing.T) {
	prog := NewProgress(ProgressTypeStreaming, 80)

	if prog.IsComplete() {
		t.Error("Progress should not be complete at initialization")
	}

	prog.SetProgress(0.5)
	if prog.IsComplete() {
		t.Error("Progress should not be complete at 50%")
	}

	prog.SetProgress(1.0)
	if !prog.IsComplete() {
		t.Error("Progress should be complete at 100%")
	}
}

func TestProgressReset(t *testing.T) {
	prog := NewProgress(ProgressTypeStreaming, 80)

	prog.SetProgress(0.7)
	prog.SetLabel("Test")

	prog.Reset()

	if prog.value != 0 {
		t.Errorf("Expected value 0 after reset, got %f", prog.value)
	}

	if prog.label != "" {
		t.Errorf("Expected empty label after reset, got '%s'", prog.label)
	}
}

func TestProgressView(t *testing.T) {
	prog := NewProgress(ProgressTypeStreaming, 80)
	prog.SetLabel("Testing")
	prog.SetProgress(0.5)

	view := prog.View()

	if view == "" {
		t.Error("View returned empty string")
	}

	if !strings.Contains(view, "Testing") {
		t.Error("View does not contain label")
	}

	if !strings.Contains(view, "50%") {
		t.Error("View does not contain percentage")
	}
}

func TestProgressViewCompact(t *testing.T) {
	prog := NewProgress(ProgressTypeStreaming, 80)
	prog.SetLabel("Testing")
	prog.SetProgress(0.5)

	compact := prog.ViewCompact()

	if compact == "" {
		t.Error("ViewCompact returned empty string")
	}

	// Compact view should not contain label
	if strings.Contains(compact, "Testing") {
		t.Error("ViewCompact should not contain label")
	}
}

func TestStreamingProgress(t *testing.T) {
	prog := StreamingProgress(50, 100, 80)

	if prog == nil {
		t.Fatal("StreamingProgress returned nil")
	}

	if prog.progressType != ProgressTypeStreaming {
		t.Error("Progress type should be Streaming")
	}

	if prog.label == "" {
		t.Error("StreamingProgress should have a label")
	}

	// Should be at 50%
	if prog.GetProgress() != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", prog.GetProgress())
	}
}

func TestStreamingProgressIndeterminate(t *testing.T) {
	prog := StreamingProgress(10, 0, 80)

	if prog == nil {
		t.Fatal("StreamingProgress returned nil")
	}

	// For indeterminate progress, should be at 50%
	if prog.GetProgress() != 0.5 {
		t.Errorf("Expected indeterminate progress 0.5, got %f", prog.GetProgress())
	}
}

func TestToolExecutionProgress(t *testing.T) {
	toolName := "bash"
	prog := ToolExecutionProgress(toolName, 80)

	if prog == nil {
		t.Fatal("ToolExecutionProgress returned nil")
	}

	if prog.progressType != ProgressTypeToolExecution {
		t.Error("Progress type should be ToolExecution")
	}

	if !strings.Contains(prog.label, toolName) {
		t.Errorf("Label should contain tool name '%s'", toolName)
	}

	// Should be indeterminate (0.5)
	if prog.GetProgress() != 0.5 {
		t.Errorf("Expected indeterminate progress 0.5, got %f", prog.GetProgress())
	}
}

func TestTokenUsageProgress(t *testing.T) {
	currentTokens := 1500
	maxTokens := 4096
	prog := TokenUsageProgress(currentTokens, maxTokens, 80)

	if prog == nil {
		t.Fatal("TokenUsageProgress returned nil")
	}

	if prog.progressType != ProgressTypeTokenUsage {
		t.Error("Progress type should be TokenUsage")
	}

	// Should show context window usage
	expectedProgress := float64(currentTokens) / float64(maxTokens)
	if prog.GetProgress() != expectedProgress {
		t.Errorf("Expected progress %f, got %f", expectedProgress, prog.GetProgress())
	}

	view := prog.View()
	if !strings.Contains(view, "tokens") {
		t.Error("Token usage view should contain 'tokens'")
	}
}

func TestBatchProgress(t *testing.T) {
	current := 3
	total := 10
	operation := "Processing files"
	prog := BatchProgress(current, total, operation, 80)

	if prog == nil {
		t.Fatal("BatchProgress returned nil")
	}

	if prog.progressType != ProgressTypeBatch {
		t.Error("Progress type should be Batch")
	}

	if !strings.Contains(prog.label, operation) {
		t.Error("Label should contain operation name")
	}

	if !strings.Contains(prog.label, "3/10") {
		t.Error("Label should contain progress count")
	}

	expectedProgress := float64(current) / float64(total)
	if prog.GetProgress() != expectedProgress {
		t.Errorf("Expected progress %f, got %f", expectedProgress, prog.GetProgress())
	}
}

func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar(80)

	if pb == nil {
		t.Fatal("NewProgressBar returned nil")
	}

	if pb.Progress == nil {
		t.Fatal("ProgressBar.Progress is nil")
	}

	if pb.width != 80 {
		t.Errorf("Expected width 80, got %d", pb.width)
	}
}
