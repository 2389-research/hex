// ABOUTME: Comprehensive tests for tool suggestion detector
// ABOUTME: Tests pattern matching, confidence scoring, and edge cases

package suggestions

import (
	"testing"
)

func TestDetector_AnalyzeInput_FilePaths(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name          string
		input         string
		expectedTool  string
		minConfidence float64
		shouldSuggest bool
	}{
		{
			name:          "absolute path",
			input:         "Can you read /etc/hosts for me?",
			expectedTool:  "read_file",
			minConfidence: 0.80,
			shouldSuggest: true,
		},
		{
			name:          "relative path",
			input:         "Check ./config/app.yaml",
			expectedTool:  "read_file",
			minConfidence: 0.75,
			shouldSuggest: true,
		},
		{
			name:          "home directory path",
			input:         "Look at ~/Documents/notes.txt",
			expectedTool:  "read_file",
			minConfidence: 0.80,
			shouldSuggest: true,
		},
		{
			name:          "multiple paths",
			input:         "Compare /etc/nginx/nginx.conf and ./config.yaml",
			expectedTool:  "read_file",
			minConfidence: 0.75,
			shouldSuggest: true,
		},
		{
			name:          "path with spaces in text",
			input:         "The file is located at /var/log/app.log and contains errors",
			expectedTool:  "read_file",
			minConfidence: 0.80,
			shouldSuggest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := detector.AnalyzeInput(tt.input)

			if tt.shouldSuggest {
				if len(suggestions) == 0 {
					t.Errorf("Expected suggestions but got none")
					return
				}

				found := false
				for _, s := range suggestions {
					if s.ToolName == tt.expectedTool {
						found = true
						if s.Confidence < tt.minConfidence {
							t.Errorf("Confidence too low: got %.2f, want >= %.2f", s.Confidence, tt.minConfidence)
						}
						if s.Action == "" {
							t.Errorf("Expected Action to be set")
						}
						if s.Reason == "" {
							t.Errorf("Expected Reason to be set")
						}
						break
					}
				}

				if !found {
					t.Errorf("Expected suggestion for %s but got suggestions for: %v", tt.expectedTool, getSuggestedTools(suggestions))
				}
			} else {
				if len(suggestions) > 0 {
					t.Errorf("Expected no suggestions but got %d", len(suggestions))
				}
			}
		})
	}
}

func TestDetector_AnalyzeInput_URLs(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name          string
		input         string
		shouldSuggest bool
	}{
		{
			name:          "https URL",
			input:         "Fetch https://api.example.com/data",
			shouldSuggest: true,
		},
		{
			name:          "http URL",
			input:         "Get http://example.com",
			shouldSuggest: true,
		},
		{
			name:          "URL with path",
			input:         "Check https://github.com/user/repo/blob/main/README.md",
			shouldSuggest: true,
		},
		{
			name:          "URL with query params",
			input:         "Load https://api.example.com/search?q=test&limit=10",
			shouldSuggest: true,
		},
		{
			name:          "incomplete URL",
			input:         "example.com",
			shouldSuggest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := detector.AnalyzeInput(tt.input)

			hasSuggestion := false
			for _, s := range suggestions {
				if s.ToolName == "web_fetch" {
					hasSuggestion = true
					if s.Confidence < 0.85 {
						t.Errorf("web_fetch confidence too low: %.2f", s.Confidence)
					}
					break
				}
			}

			if tt.shouldSuggest && !hasSuggestion {
				t.Errorf("Expected web_fetch suggestion but got none")
			}
			if !tt.shouldSuggest && hasSuggestion {
				t.Errorf("Did not expect web_fetch suggestion but got one")
			}
		})
	}
}

func TestDetector_AnalyzeInput_SearchIntent(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "search for pattern",
			input: "search for 'TODO' in the codebase",
		},
		{
			name:  "find pattern",
			input: "find all instances of getUserInfo",
		},
		{
			name:  "grep pattern",
			input: "grep for error messages",
		},
		{
			name:  "look for pattern",
			input: "look for configuration files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := detector.AnalyzeInput(tt.input)

			found := false
			for _, s := range suggestions {
				if s.ToolName == "grep" {
					found = true
					if s.Confidence < 0.70 {
						t.Errorf("grep confidence too low: %.2f", s.Confidence)
					}
					break
				}
			}

			if !found {
				t.Errorf("Expected grep suggestion but got none for input: %s", tt.input)
			}
		})
	}
}

func TestDetector_AnalyzeInput_CommandIntent(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "run command",
			input: "run npm test",
		},
		{
			name:  "execute command",
			input: "execute go build",
		},
		{
			name:  "exec command",
			input: "exec docker ps",
		},
		{
			name:  "direct shell command",
			input: "ls -la",
		},
		{
			name:  "git command",
			input: "git status",
		},
		{
			name:  "docker command",
			input: "docker ps -a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := detector.AnalyzeInput(tt.input)

			found := false
			for _, s := range suggestions {
				if s.ToolName == "bash" {
					found = true
					if s.Confidence < 0.70 {
						t.Errorf("bash confidence too low: %.2f", s.Confidence)
					}
					break
				}
			}

			if !found {
				t.Errorf("Expected bash suggestion but got none for input: %s", tt.input)
			}
		})
	}
}

func TestDetector_AnalyzeInput_GlobPatterns(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "double star pattern",
			input: "Find all **/*.go files",
		},
		{
			name:  "single star extension",
			input: "List *.txt files",
		},
		{
			name:  "star in path",
			input: "Show me src/**/*.test.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := detector.AnalyzeInput(tt.input)

			found := false
			for _, s := range suggestions {
				if s.ToolName == "glob" {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected glob suggestion for: %s", tt.input)
			}
		})
	}
}

func TestDetector_AnalyzeInput_WriteIntent(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name         string
		input        string
		expectedTool string
	}{
		{
			name:         "write file",
			input:        "write to config.yaml",
			expectedTool: "write_file",
		},
		{
			name:         "create file",
			input:        "create file output.txt",
			expectedTool: "write_file",
		},
		{
			name:         "save to file",
			input:        "save to /tmp/data.json",
			expectedTool: "write_file",
		},
		{
			name:         "edit file",
			input:        "edit src/main.go",
			expectedTool: "edit",
		},
		{
			name:         "modify file",
			input:        "modify config.json",
			expectedTool: "edit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := detector.AnalyzeInput(tt.input)

			found := false
			for _, s := range suggestions {
				if s.ToolName == tt.expectedTool {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected %s suggestion for: %s", tt.expectedTool, tt.input)
			}
		})
	}
}

func TestDetector_AnalyzeInput_WebSearch(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "google query",
			input: "google how to use channels in Go",
		},
		{
			name:  "search for query",
			input: "search for latest React documentation",
		},
		{
			name:  "look up query",
			input: "look up Docker best practices",
		},
		{
			name:  "web search query",
			input: "web search Kubernetes tutorials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := detector.AnalyzeInput(tt.input)

			found := false
			for _, s := range suggestions {
				if s.ToolName == "web_search" {
					found = true
					if s.Confidence < 0.70 {
						t.Errorf("web_search confidence too low: %.2f", s.Confidence)
					}
					break
				}
			}

			if !found {
				t.Errorf("Expected web_search suggestion for: %s", tt.input)
			}
		})
	}
}

func TestDetector_AnalyzeInput_NoSuggestions(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name   string
		input  string
		reason string
	}{
		{
			name:   "too short",
			input:  "hi",
			reason: "Input too short",
		},
		{
			name:   "empty",
			input:  "",
			reason: "Empty input",
		},
		{
			name:   "whitespace only",
			input:  "   ",
			reason: "Whitespace only",
		},
		{
			name:   "already tool command",
			input:  ":read /etc/hosts",
			reason: "Already a tool command",
		},
		{
			name:   "generic question",
			input:  "What is the weather like?",
			reason: "Generic question with no tool patterns",
		},
		{
			name:   "trailing space short",
			input:  "how ",
			reason: "Mid-sentence with trailing space",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := detector.AnalyzeInput(tt.input)

			if len(suggestions) > 0 {
				t.Errorf("Expected no suggestions for '%s' (%s), but got %d: %v",
					tt.input, tt.reason, len(suggestions), getSuggestedTools(suggestions))
			}
		})
	}
}

func TestDetector_AnalyzeInput_ConfidenceThreshold(t *testing.T) {
	detector := NewDetector()

	// All returned suggestions should have confidence >= 0.70
	inputs := []string{
		"/etc/hosts",
		"https://example.com",
		"search for pattern",
		"run npm test",
		"*.go",
	}

	for _, input := range inputs {
		suggestions := detector.AnalyzeInput(input)
		for _, s := range suggestions {
			if s.Confidence < 0.70 {
				t.Errorf("Suggestion for '%s' has confidence %.2f below threshold 0.70",
					input, s.Confidence)
			}
		}
	}
}

func TestDetector_AnalyzeInput_MaxSuggestions(t *testing.T) {
	detector := NewDetector()

	// Input that could match multiple patterns
	input := "search for files in /etc/config matching *.conf and run ls -la"

	suggestions := detector.AnalyzeInput(input)

	// Should limit to 3 suggestions max
	if len(suggestions) > 3 {
		t.Errorf("Expected max 3 suggestions, got %d", len(suggestions))
	}
}

func TestDetector_AnalyzeInput_Deduplication(t *testing.T) {
	detector := NewDetector()

	// Input with multiple paths (should not duplicate read_file suggestion)
	input := "Compare /etc/hosts and ./config.yaml"

	suggestions := detector.AnalyzeInput(input)

	// Count occurrences of each tool
	toolCounts := make(map[string]int)
	for _, s := range suggestions {
		toolCounts[s.ToolName]++
	}

	for tool, count := range toolCounts {
		if count > 1 {
			t.Errorf("Tool %s suggested %d times, expected only once", tool, count)
		}
	}
}

func TestDetector_AnalyzeInput_SortedByConfidence(t *testing.T) {
	detector := NewDetector()

	// Input that triggers multiple suggestions
	input := "search for pattern in /etc/hosts"

	suggestions := detector.AnalyzeInput(input)

	// Verify sorted by confidence (descending)
	for i := 1; i < len(suggestions); i++ {
		if suggestions[i].Confidence > suggestions[i-1].Confidence {
			t.Errorf("Suggestions not sorted by confidence: %.2f > %.2f at index %d",
				suggestions[i].Confidence, suggestions[i-1].Confidence, i)
		}
	}
}

func TestDetector_AnalyzeInput_ActionGeneration(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name           string
		input          string
		expectedTool   string
		expectedAction string
	}{
		{
			name:           "file path with action",
			input:          "Read /etc/hosts",
			expectedTool:   "read_file",
			expectedAction: ":read /etc/hosts",
		},
		{
			name:           "URL with action",
			input:          "Fetch https://example.com",
			expectedTool:   "web_fetch",
			expectedAction: ":web_fetch https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := detector.AnalyzeInput(tt.input)

			found := false
			for _, s := range suggestions {
				if s.ToolName == tt.expectedTool {
					found = true
					if s.Action != tt.expectedAction {
						t.Errorf("Expected action '%s', got '%s'", tt.expectedAction, s.Action)
					}
					break
				}
			}

			if !found {
				t.Errorf("Expected to find suggestion for %s", tt.expectedTool)
			}
		})
	}
}

// Helper function to get tool names from suggestions
func getSuggestedTools(suggestions []Suggestion) []string {
	tools := make([]string, len(suggestions))
	for i, s := range suggestions {
		tools[i] = s.ToolName
	}
	return tools
}
