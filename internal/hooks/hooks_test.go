// ABOUTME: Comprehensive tests for hooks system
// ABOUTME: Tests configuration, execution, filtering, and integration

package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestEventDataToEnvVars tests that event data correctly converts to environment variables
func TestEventDataToEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		data     EventData
		expected map[string]string
	}{
		{
			name: "SessionStartData",
			data: SessionStartData{
				ProjectPath: "/test/project",
				ModelID:     "claude-sonnet-4",
			},
			expected: map[string]string{
				"CLAUDE_PROJECT_PATH": "/test/project",
				"CLAUDE_MODEL_ID":     "claude-sonnet-4",
			},
		},
		{
			name: "ToolUseData",
			data: ToolUseData{
				ToolName:   "Edit",
				FilePath:   "/test/file.go",
				Success:    true,
				IsSubagent: false,
			},
			expected: map[string]string{
				"CLAUDE_TOOL_NAME":      "Edit",
				"CLAUDE_TOOL_FILE_PATH": "/test/file.go",
				"CLAUDE_TOOL_SUCCESS":   "true",
				"CLAUDE_IS_SUBAGENT":    "false",
			},
		},
		{
			name: "UserPromptSubmitData",
			data: UserPromptSubmitData{
				MessageText:   "test message",
				MessageLength: 12,
			},
			expected: map[string]string{
				"CLAUDE_MESSAGE_TEXT":   "test message",
				"CLAUDE_MESSAGE_LENGTH": "12",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.ToEnvVars()
			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("expected %s=%s, got %s=%s", key, expectedValue, key, result[key])
				}
			}
		})
	}
}

// TestConfigLoading tests loading configuration from JSON files
func TestConfigLoading(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "settings.json")

	// Create test configuration
	settingsJSON := `{
		"hooks": {
			"PostToolUse": {
				"command": "echo 'test'",
				"description": "Test hook",
				"timeout": 10000,
				"match": {
					"toolName": "Edit",
					"filePattern": ".*\\.go$"
				}
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(settingsJSON), 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load configuration
	config := NewConfig()
	if err := config.loadFromFile(configPath); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify hooks were loaded
	hooks := config.GetHooks(EventPostToolUse)
	if len(hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hooks))
	}

	hook := hooks[0]
	if hook.Command != "echo 'test'" {
		t.Errorf("expected command 'echo 'test'', got '%s'", hook.Command)
	}
	if hook.Description != "Test hook" {
		t.Errorf("expected description 'Test hook', got '%s'", hook.Description)
	}
	if hook.Timeout != 10000 {
		t.Errorf("expected timeout 10000, got %d", hook.Timeout)
	}
	if hook.Match.ToolName != "Edit" {
		t.Errorf("expected toolName 'Edit', got '%s'", hook.Match.ToolName)
	}
}

// TestMultipleHooks tests loading multiple hooks for the same event
func TestMultipleHooks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "settings.json")

	settingsJSON := `{
		"hooks": {
			"PostToolUse": [
				{
					"command": "echo 'hook1'",
					"match": {"toolName": "Edit"}
				},
				{
					"command": "echo 'hook2'",
					"match": {"toolName": "Write"}
				}
			]
		}
	}`

	if err := os.WriteFile(configPath, []byte(settingsJSON), 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("failed to write test config: %v", err)
	}

	config := NewConfig()
	if err := config.loadFromFile(configPath); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	hooks := config.GetHooks(EventPostToolUse)
	if len(hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(hooks))
	}

	if hooks[0].Match.ToolName != "Edit" {
		t.Errorf("expected first hook toolName 'Edit', got '%s'", hooks[0].Match.ToolName)
	}
	if hooks[1].Match.ToolName != "Write" {
		t.Errorf("expected second hook toolName 'Write', got '%s'", hooks[1].Match.ToolName)
	}
}

// TestShouldExecute tests hook matching logic
func TestShouldExecute(t *testing.T) {
	tests := []struct {
		name     string
		hook     HookConfig
		event    *Event
		expected bool
	}{
		{
			name: "no matcher - always executes",
			hook: HookConfig{
				Command: "echo 'test'",
			},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{ToolName: "Edit"},
			},
			expected: true,
		},
		{
			name: "tool name match - success",
			hook: HookConfig{
				Command: "echo 'test'",
				Match:   &MatchConfig{ToolName: "Edit"},
			},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{ToolName: "Edit"},
			},
			expected: true,
		},
		{
			name: "tool name match - failure",
			hook: HookConfig{
				Command: "echo 'test'",
				Match:   &MatchConfig{ToolName: "Write"},
			},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{ToolName: "Edit"},
			},
			expected: false,
		},
		{
			name: "file pattern match - success",
			hook: HookConfig{
				Command: "echo 'test'",
				Match:   &MatchConfig{FilePattern: `.*\.go$`},
			},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{FilePath: "/test/file.go"},
			},
			expected: true,
		},
		{
			name: "file pattern match - failure",
			hook: HookConfig{
				Command: "echo 'test'",
				Match:   &MatchConfig{FilePattern: `.*\.go$`},
			},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{FilePath: "/test/file.ts"},
			},
			expected: false,
		},
		{
			name: "combined match - success",
			hook: HookConfig{
				Command: "echo 'test'",
				Match: &MatchConfig{
					ToolName:    "Edit",
					FilePattern: `.*\.go$`,
				},
			},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{
					ToolName: "Edit",
					FilePath: "/test/file.go",
				},
			},
			expected: true,
		},
		{
			name: "combined match - partial failure",
			hook: HookConfig{
				Command: "echo 'test'",
				Match: &MatchConfig{
					ToolName:    "Edit",
					FilePattern: `.*\.ts$`,
				},
			},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{
					ToolName: "Edit",
					FilePath: "/test/file.go",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.hook.ShouldExecute(tt.event)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsSubagentMatching tests subagent filtering
func TestIsSubagentMatching(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		match    *MatchConfig
		event    *Event
		expected bool
	}{
		{
			name:  "no subagent filter",
			match: &MatchConfig{},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{IsSubagent: true},
			},
			expected: true,
		},
		{
			name:  "subagent true - matches",
			match: &MatchConfig{IsSubagent: &trueVal},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{IsSubagent: true},
			},
			expected: true,
		},
		{
			name:  "subagent true - doesn't match",
			match: &MatchConfig{IsSubagent: &trueVal},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{IsSubagent: false},
			},
			expected: false,
		},
		{
			name:  "subagent false - matches",
			match: &MatchConfig{IsSubagent: &falseVal},
			event: &Event{
				Type: EventPostToolUse,
				Data: ToolUseData{IsSubagent: false},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := HookConfig{
				Command: "test",
				Match:   tt.match,
			}
			result := hook.ShouldExecute(tt.event)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestExecutor tests command execution
func TestExecutor(t *testing.T) {
	executor := NewExecutor("/test/project", "claude-sonnet-4")

	hook := &HookConfig{
		Command: "echo 'hello world'",
		Timeout: 5000,
	}

	event := &Event{
		Type:      EventSessionStart,
		Timestamp: time.Now(),
		Data: SessionStartData{
			ProjectPath: "/test/project",
			ModelID:     "claude-sonnet-4",
		},
	}

	result := executor.Execute(hook, event)

	if !result.Success {
		t.Errorf("expected success, got error: %v", result.Error)
	}

	if result.Stdout != "hello world\n" {
		t.Errorf("expected 'hello world\\n', got '%s'", result.Stdout)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

// TestExecutorTimeout tests command timeout handling
func TestExecutorTimeout(t *testing.T) {
	executor := NewExecutor("/test/project", "claude-sonnet-4")

	hook := &HookConfig{
		Command: "sleep 2",
		Timeout: 100, // 100ms timeout
	}

	event := &Event{
		Type:      EventSessionStart,
		Timestamp: time.Now(),
		Data:      SessionStartData{},
	}

	result := executor.Execute(hook, event)

	if result.Success {
		t.Error("expected timeout failure, got success")
	}

	if result.Error == nil {
		t.Error("expected error for timeout, got nil")
	}
}

// TestEngine tests the hook engine
func TestEngine(t *testing.T) {
	// Create a test configuration
	config := NewConfig()
	config.hooks[EventPostToolUse] = []HookConfig{
		{
			Command: "echo 'test hook executed'",
			Match:   &MatchConfig{ToolName: "Edit"},
		},
	}

	engine := NewEngineWithConfig(config, "/test/project", "claude-sonnet-4")

	// Fire the event
	err := engine.FirePostToolUse("Edit", "/test/file.go", true, "", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestEngineWithNoMatchingHooks tests that engine handles no matching hooks gracefully
func TestEngineWithNoMatchingHooks(t *testing.T) {
	config := NewConfig()
	config.hooks[EventPostToolUse] = []HookConfig{
		{
			Command: "echo 'test'",
			Match:   &MatchConfig{ToolName: "Write"},
		},
	}

	engine := NewEngineWithConfig(config, "/test/project", "claude-sonnet-4")

	// Fire event that doesn't match
	err := engine.FirePostToolUse("Edit", "/test/file.go", true, "", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestGetTimeout tests default timeout behavior
func TestGetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected int
	}{
		{"default timeout", 0, 5000},
		{"negative timeout", -100, 5000},
		{"custom timeout", 10000, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := HookConfig{Timeout: tt.timeout}
			result := hook.GetTimeout()
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestEnvVarBuilding tests that environment variables are correctly built
func TestEnvVarBuilding(t *testing.T) {
	executor := NewExecutor("/test/project", "claude-sonnet-4")

	hook := &HookConfig{
		Command: "env",
		Env: map[string]string{
			"CUSTOM_VAR": "custom_value",
		},
	}

	event := &Event{
		Type:      EventPostToolUse,
		Timestamp: time.Now(),
		Data: ToolUseData{
			ToolName: "Edit",
			FilePath: "/test/file.go",
		},
	}

	env := executor.buildEnv(hook, event)

	// Check for expected environment variables
	expectedVars := []string{
		"CLAUDE_EVENT=PostToolUse",
		"CLAUDE_TOOL_NAME=Edit",
		"CLAUDE_TOOL_FILE_PATH=/test/file.go",
		"CUSTOM_VAR=custom_value",
	}

	for _, expectedVar := range expectedVars {
		found := false
		for _, envVar := range env {
			if envVar == expectedVar {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected environment variable '%s' not found", expectedVar)
		}
	}
}

// TestJSONMarshaling tests that event types can be marshaled to JSON
func TestJSONMarshaling(t *testing.T) {
	settings := Settings{
		Hooks: map[EventType]interface{}{
			EventPostToolUse: HookConfig{
				Command:     "echo 'test'",
				Description: "Test hook",
			},
		},
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("failed to marshal settings: %v", err)
	}

	var unmarshaled Settings
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal settings: %v", err)
	}

	if unmarshaled.Hooks[EventPostToolUse] == nil {
		t.Error("expected PostToolUse hook to be present after unmarshaling")
	}
}

// TestDisableEnable tests engine enable/disable functionality
func TestDisableEnable(t *testing.T) {
	config := NewConfig()
	config.hooks[EventSessionStart] = []HookConfig{
		{Command: "echo 'test'"},
	}

	engine := NewEngineWithConfig(config, "/test", "claude-sonnet-4")

	// Disable engine
	engine.Disable()

	// Fire should do nothing when disabled
	err := engine.FireSessionStart("/test", "claude-sonnet-4")
	if err != nil {
		t.Errorf("unexpected error when disabled: %v", err)
	}

	// Re-enable
	engine.Enable()

	// Now it should execute
	err = engine.FireSessionStart("/test", "claude-sonnet-4")
	if err != nil {
		t.Errorf("unexpected error when enabled: %v", err)
	}
}
