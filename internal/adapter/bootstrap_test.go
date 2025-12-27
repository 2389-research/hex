// ABOUTME: Tests for agent bootstrap functions.
// ABOUTME: Verifies root and subagent creation with proper tool filtering.
package adapter

import (
	"os"
	"testing"
)

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
