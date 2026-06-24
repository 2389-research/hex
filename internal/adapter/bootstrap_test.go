// ABOUTME: Tests for agent bootstrap functions.
// ABOUTME: Verifies root and subagent creation with proper tool filtering.
package adapter

import (
	"context"
	"os"
	"testing"

	"github.com/2389-research/hex/internal/skills"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/mux/llm"
)

type stubLLMClient struct{}

func (stubLLMClient) CreateMessage(context.Context, *llm.Request) (*llm.Response, error) {
	return &llm.Response{}, nil
}

func (stubLLMClient) CreateMessageStream(context.Context, *llm.Request) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent)
	close(ch)
	return ch, nil
}

func TestParseCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"Read", []string{"Read"}},
		{"Read,Grep,Glob", []string{"Read", "Grep", "Glob"}},
		{"Read, Grep, Glob", []string{"Read", "Grep", "Glob"}},
	}

	for _, tc := range tests {
		result := parseCSV(tc.input)
		if len(result) != len(tc.expected) {
			t.Errorf("parseCSV(%q): expected %v, got %v", tc.input, tc.expected, result)
			continue
		}
		for i := range result {
			if result[i] != tc.expected[i] {
				t.Errorf("parseCSV(%q)[%d]: expected %q, got %q", tc.input, i, tc.expected[i], result[i])
			}
		}
	}
}

func TestIsSubagent(t *testing.T) {
	// Clean env
	if err := os.Unsetenv("HEX_SUBAGENT_TYPE"); err != nil {
		t.Fatalf("failed to unset env: %v", err)
	}

	if IsSubagent() {
		t.Error("expected IsSubagent() to return false when env not set")
	}

	if err := os.Setenv("HEX_SUBAGENT_TYPE", "Explore"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("HEX_SUBAGENT_TYPE"); err != nil {
			t.Errorf("failed to unset env in defer: %v", err)
		}
	}()

	if !IsSubagent() {
		t.Error("expected IsSubagent() to return true when env is set")
	}
}

func TestAssistantTextFallsBackToTextBlocks(t *testing.T) {
	message := llm.Message{
		Role: llm.RoleAssistant,
		Blocks: []llm.ContentBlock{
			{Type: llm.ContentTypeThinking, Thinking: "internal reasoning"},
			{Type: llm.ContentTypeText, Text: "first"},
			{Type: llm.ContentTypeText, Text: " second"},
		},
	}

	if got := assistantText(message); got != "first second" {
		t.Fatalf("assistantText() = %q, want %q", got, "first second")
	}
}

func TestFilterToolsByAllowedMatchesAliasesAndSkill(t *testing.T) {
	available := []tools.Tool{
		tools.NewReadTool(),
		tools.NewGrepTool(),
		tools.NewGlobTool(),
		skills.NewToolAdapter(skills.NewRegistry()),
	}

	filtered := filterToolsByAllowed(available, []string{"Read", "Grep", "Glob", "Skill"})

	names := make(map[string]bool)
	for _, tool := range filtered {
		names[tool.Name()] = true
	}

	for _, name := range []string{"read_file", "grep", "glob", "Skill"} {
		if !names[name] {
			t.Fatalf("filterToolsByAllowed() missing %q from filtered names %v", name, names)
		}
	}
}

func TestNewRootAgentCarriesMaxIterations(t *testing.T) {
	root := NewRootAgent(Config{
		Model:         "test-model",
		LLMClient:     stubLLMClient{},
		MaxIterations: 7,
	})

	if got := root.Config().MaxIterations; got != 7 {
		t.Fatalf("root agent MaxIterations = %d, want 7", got)
	}
}

func TestNewSubagentCarriesMaxIterations(t *testing.T) {
	subagent := NewSubagent(Config{
		Model:         "test-model",
		LLMClient:     stubLLMClient{},
		MaxIterations: 3,
	})

	if got := subagent.Config().MaxIterations; got != 3 {
		t.Fatalf("subagent MaxIterations = %d, want 3", got)
	}
}
