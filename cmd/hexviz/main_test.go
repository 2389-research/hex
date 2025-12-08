// ABOUTME: main_test.go contains comprehensive tests for the hexviz visualization tool
// ABOUTME: Tests cover tree view, timeline view, cost view, filtering, and HTML export

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadEvents verifies loading events from JSON Lines file
func TestLoadEvents(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "test_events.jsonl")

	events := []map[string]interface{}{
		{
			"id":        "evt-1",
			"agent_id":  "root",
			"parent_id": "",
			"type":      "AgentStart",
			"timestamp": time.Now().Format(time.RFC3339),
			"data": map[string]interface{}{
				"task": "Main task",
			},
		},
		{
			"id":        "evt-2",
			"agent_id":  "root.1",
			"parent_id": "root",
			"type":      "AgentStart",
			"timestamp": time.Now().Format(time.RFC3339),
			"data": map[string]interface{}{
				"task": "Sub task",
			},
		},
	}

	// Write events to file
	f, err := os.Create(eventFile)
	require.NoError(t, err)
	for _, evt := range events {
		line, _ := json.Marshal(evt)
		_, _ = f.Write(append(line, '\n'))
	}
	_ = f.Close()

	// Test loading
	loaded, err := loadEvents(eventFile)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, "root", loaded[0].AgentID)
	assert.Equal(t, "root.1", loaded[1].AgentID)
}

// TestLoadEvents_NonExistent verifies error handling for missing files
func TestLoadEvents_NonExistent(t *testing.T) {
	_, err := loadEvents("/nonexistent/file.jsonl")
	assert.Error(t, err)
}

// TestLoadEvents_Malformed verifies graceful handling of malformed lines
func TestLoadEvents_Malformed(t *testing.T) {
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "bad_events.jsonl")

	// Write mixed good and bad lines
	f, err := os.Create(eventFile)
	require.NoError(t, err)
	_, _ = f.WriteString(`{"id":"evt-1","agent_id":"root","type":"AgentStart","timestamp":"2025-12-07T10:00:00Z","data":{}}` + "\n")
	_, _ = f.WriteString("INVALID JSON LINE\n")
	_, _ = f.WriteString(`{"id":"evt-2","agent_id":"root.1","type":"AgentStart","timestamp":"2025-12-07T10:01:00Z","data":{}}` + "\n")
	_ = f.Close()

	// Should skip malformed lines and return valid ones
	loaded, err := loadEvents(eventFile)
	require.NoError(t, err)
	assert.Len(t, loaded, 2) // Only valid lines
}

// TestBuildAgentTree verifies agent hierarchy construction
func TestBuildAgentTree(t *testing.T) {
	events := []Event{
		{
			AgentID:   "root",
			Type:      "AgentStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"task": "Main task",
			},
		},
		{
			AgentID:   "root.1",
			ParentID:  "root",
			Type:      "AgentStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"task": "Sub task 1",
			},
		},
		{
			AgentID:   "root.2",
			ParentID:  "root",
			Type:      "AgentStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"task": "Sub task 2",
			},
		},
		{
			AgentID:   "root.1.1",
			ParentID:  "root.1",
			Type:      "AgentStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"task": "Nested task",
			},
		},
		{
			AgentID:   "root",
			Type:      "AgentStop",
			Timestamp: time.Now().Add(2 * time.Minute),
			Data: map[string]interface{}{
				"usage": map[string]interface{}{
					"input_tokens":  1000.0,
					"output_tokens": 500.0,
				},
			},
		},
	}

	tree := buildAgentTree(events)

	// Verify root
	assert.Equal(t, "root", tree.ID)
	assert.Equal(t, "Main task", tree.Task)
	assert.Len(t, tree.Children, 2)
	assert.Greater(t, tree.Duration, time.Duration(0))

	// Verify children
	child1 := tree.Children[0]
	assert.Equal(t, "root.1", child1.ID)
	assert.Equal(t, "Sub task 1", child1.Task)
	assert.Len(t, child1.Children, 1)

	// Verify nested child
	nested := child1.Children[0]
	assert.Equal(t, "root.1.1", nested.ID)
	assert.Equal(t, "Nested task", nested.Task)
}

// TestRenderTreeView verifies tree visualization output
func TestRenderTreeView(t *testing.T) {
	events := []Event{
		{
			AgentID:   "root",
			Type:      "AgentStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"task": "Main task",
			},
		},
		{
			AgentID:   "root.1",
			ParentID:  "root",
			Type:      "AgentStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"task": "Sub task",
			},
		},
		{
			AgentID:   "root",
			Type:      "AgentStop",
			Timestamp: time.Now().Add(1 * time.Minute),
			Data:      map[string]interface{}{},
		},
	}

	output := renderTreeView(events)

	// Verify structure
	assert.Contains(t, output, "root")
	assert.Contains(t, output, "Main task")
	assert.Contains(t, output, "root.1")
	assert.Contains(t, output, "Sub task")

	// Verify tree characters present
	assert.True(t, strings.Contains(output, "├") || strings.Contains(output, "└"))
}

// TestRenderTimelineView verifies chronological event display
func TestRenderTimelineView(t *testing.T) {
	now := time.Now()
	events := []Event{
		{
			AgentID:   "root",
			Type:      "AgentStart",
			Timestamp: now,
			Data: map[string]interface{}{
				"task": "Main task",
			},
		},
		{
			AgentID:   "root",
			Type:      "StreamStart",
			Timestamp: now.Add(1 * time.Second),
			Data:      map[string]interface{}{},
		},
		{
			AgentID:   "root",
			Type:      "ToolCall",
			Timestamp: now.Add(2 * time.Second),
			Data: map[string]interface{}{
				"tool_name": "read_file",
				"input": map[string]interface{}{
					"path": "main.go",
				},
			},
		},
		{
			AgentID:   "root",
			Type:      "ToolResult",
			Timestamp: now.Add(3 * time.Second),
			Data: map[string]interface{}{
				"tool_name": "read_file",
				"success":   true,
			},
		},
	}

	output := renderTimelineView(events)

	// Verify all events present
	assert.Contains(t, output, "AgentStart")
	assert.Contains(t, output, "StreamStart")
	assert.Contains(t, output, "ToolCall")
	assert.Contains(t, output, "ToolResult")

	// Verify tool details
	assert.Contains(t, output, "read_file")
	assert.Contains(t, output, "main.go")

	// Verify chronological order (first event appears before last)
	startIdx := strings.Index(output, "AgentStart")
	resultIdx := strings.Index(output, "ToolResult")
	assert.Less(t, startIdx, resultIdx)
}

// TestRenderCostView verifies cost breakdown display
func TestRenderCostView(t *testing.T) {
	events := []Event{
		{
			AgentID:   "root",
			Type:      "AgentStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"model": "claude-sonnet-4-5-20250929",
			},
		},
		{
			AgentID:   "root",
			Type:      "AgentStop",
			Timestamp: time.Now().Add(1 * time.Minute),
			Data: map[string]interface{}{
				"usage": map[string]interface{}{
					"input_tokens":       12500.0,
					"output_tokens":      3200.0,
					"cache_read_tokens":  8000.0,
					"cache_write_tokens": 0.0,
				},
			},
		},
		{
			AgentID:   "root.1",
			ParentID:  "root",
			Type:      "AgentStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"model": "claude-sonnet-4-5-20250929",
			},
		},
		{
			AgentID:   "root.1",
			Type:      "AgentStop",
			Timestamp: time.Now().Add(30 * time.Second),
			Data: map[string]interface{}{
				"usage": map[string]interface{}{
					"input_tokens":  5000.0,
					"output_tokens": 1800.0,
				},
			},
		},
	}

	output := renderCostView(events)

	// Verify agent sections
	assert.Contains(t, output, "root")
	assert.Contains(t, output, "root.1")

	// Verify model names
	assert.Contains(t, output, "claude-sonnet-4-5-20250929")

	// Verify token counts with formatting
	assert.Contains(t, output, "12,500")
	assert.Contains(t, output, "3,200")
	assert.Contains(t, output, "8,000")
	assert.Contains(t, output, "5,000")
	assert.Contains(t, output, "1,800")

	// Verify cost calculations present
	assert.Contains(t, output, "$")
	assert.Contains(t, output, "Total Cost:")
}

// TestFilterByAgent verifies agent ID filtering
func TestFilterByAgent(t *testing.T) {
	events := []Event{
		{AgentID: "root", Type: "AgentStart"},
		{AgentID: "root.1", Type: "AgentStart"},
		{AgentID: "root.2", Type: "AgentStart"},
		{AgentID: "root.1.1", Type: "AgentStart"},
	}

	// Filter for root.1 and descendants
	filtered := filterByAgent(events, "root.1")

	assert.Len(t, filtered, 2)
	assert.Equal(t, "root.1", filtered[0].AgentID)
	assert.Equal(t, "root.1.1", filtered[1].AgentID)
}

// TestFilterByAgent_Exact verifies exact agent match
func TestFilterByAgent_Exact(t *testing.T) {
	events := []Event{
		{AgentID: "root", Type: "AgentStart"},
		{AgentID: "root.1", Type: "AgentStart"},
		{AgentID: "root.11", Type: "AgentStart"}, // Should NOT match root.1
	}

	filtered := filterByAgent(events, "root.1")

	assert.Len(t, filtered, 1)
	assert.Equal(t, "root.1", filtered[0].AgentID)
}

// TestFilterByType verifies event type filtering
func TestFilterByType(t *testing.T) {
	events := []Event{
		{Type: "ToolCall"},
		{Type: "ToolResult"},
		{Type: "ToolCall"},
		{Type: "StreamChunk"},
	}

	filtered := filterByType(events, "ToolCall")

	assert.Len(t, filtered, 2)
	for _, evt := range filtered {
		assert.Equal(t, EventType("ToolCall"), evt.Type)
	}
}

// TestExportHTML verifies HTML file generation
func TestExportHTML(t *testing.T) {
	events := []Event{
		{
			AgentID:   "root",
			Type:      "AgentStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"task": "Test task",
			},
		},
		{
			AgentID:   "root",
			Type:      "AgentStop",
			Timestamp: time.Now().Add(1 * time.Minute),
			Data:      map[string]interface{}{},
		},
	}

	tmpFile := filepath.Join(t.TempDir(), "output.html")
	err := exportHTML(events, tmpFile)

	require.NoError(t, err)
	assert.FileExists(t, tmpFile)

	// Verify HTML structure
	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	html := string(content)
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "<html>")
	assert.Contains(t, html, "</html>")
	assert.Contains(t, html, "root")
	assert.Contains(t, html, "Test task")
}

// TestFormatDuration verifies duration formatting
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{1*time.Minute + 15*time.Second, "1m 15s"},
		{2*time.Minute + 30*time.Second, "2m 30s"},
		{1 * time.Hour, "1h 0m 0s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		assert.Equal(t, tt.expected, result)
	}
}

// TestFormatTokens verifies token number formatting
func TestFormatTokens(t *testing.T) {
	tests := []struct {
		tokens   int64
		expected string
	}{
		{0, "0"},
		{100, "100"},
		{1000, "1,000"},
		{12500, "12,500"},
		{1234567, "1,234,567"},
	}

	for _, tt := range tests {
		result := formatTokens(tt.tokens)
		assert.Equal(t, tt.expected, result)
	}
}

// TestCalculateCost verifies cost calculation
func TestCalculateCost(t *testing.T) {
	// Claude Sonnet 4.5 pricing
	cost := calculateCost(1000000, 3.0) // 1M tokens at $3 per MTok
	assert.Equal(t, 3.0, cost)

	cost = calculateCost(500000, 15.0) // 500K tokens at $15 per MTok
	assert.Equal(t, 7.5, cost)
}

// TestExtractTaskFromData verifies task extraction
func TestExtractTaskFromData(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "task field present",
			data:     map[string]interface{}{"task": "Do something"},
			expected: "Do something",
		},
		{
			name:     "no task field",
			data:     map[string]interface{}{"other": "value"},
			expected: "",
		},
		{
			name:     "nil data",
			data:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTask(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractModelFromData verifies model extraction
func TestExtractModelFromData(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "model field present",
			data:     map[string]interface{}{"model": "claude-sonnet-4-5-20250929"},
			expected: "claude-sonnet-4-5-20250929",
		},
		{
			name:     "no model field",
			data:     map[string]interface{}{"other": "value"},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractModel(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractUsageFromData verifies usage extraction
func TestExtractUsageFromData(t *testing.T) {
	data := map[string]interface{}{
		"usage": map[string]interface{}{
			"input_tokens":       12500.0,
			"output_tokens":      3200.0,
			"cache_read_tokens":  8000.0,
			"cache_write_tokens": 1000.0,
		},
	}

	usage := extractUsage(data)

	assert.Equal(t, int64(12500), usage.InputTokens)
	assert.Equal(t, int64(3200), usage.OutputTokens)
	assert.Equal(t, int64(8000), usage.CacheReads)
	assert.Equal(t, int64(1000), usage.CacheWrites)
}

// TestExtractUsageFromData_Missing verifies handling of missing usage
func TestExtractUsageFromData_Missing(t *testing.T) {
	data := map[string]interface{}{"other": "value"}

	usage := extractUsage(data)

	assert.Equal(t, int64(0), usage.InputTokens)
	assert.Equal(t, int64(0), usage.OutputTokens)
}

// TestGetPricing verifies model pricing lookup
func TestGetPricing(t *testing.T) {
	pricing, err := getPricing("claude-sonnet-4-5-20250929")
	require.NoError(t, err)

	assert.Equal(t, 3.0, pricing.InputTokenPrice)
	assert.Equal(t, 15.0, pricing.OutputTokenPrice)
	assert.Equal(t, 0.30, pricing.CacheReadPrice)
	assert.Equal(t, 3.75, pricing.CacheWritePrice)
}

// TestGetPricing_UnknownModel verifies fallback for unknown models
func TestGetPricing_UnknownModel(t *testing.T) {
	pricing, err := getPricing("unknown-model")
	require.NoError(t, err)

	// Should use default pricing
	assert.Greater(t, pricing.InputTokenPrice, 0.0)
	assert.Greater(t, pricing.OutputTokenPrice, 0.0)
}
