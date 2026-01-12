// ABOUTME: Pattern-based tool suggestion provider for tux integration
// ABOUTME: Implements tux.SuggestionProvider interface

package tui

import (
	"regexp"
	"strings"

	"github.com/2389-research/tux"
)

// Detector implements tux.SuggestionProvider to analyze user input
// and suggest relevant tools based on context patterns.
type Detector struct {
	patterns []pattern
}

// pattern represents a detection pattern for a specific tool.
type pattern struct {
	name       string
	regex      *regexp.Regexp
	toolName   string
	confidence float64
	reason     string
	actionFunc func(matches []string) string
}

// NewDetector creates a new suggestion detector with default patterns.
func NewDetector() *Detector {
	d := &Detector{}
	d.registerDefaultPatterns()
	return d
}

// registerDefaultPatterns registers all built-in detection patterns.
func (d *Detector) registerDefaultPatterns() {
	// File path patterns - suggest read_file
	d.addPattern(pattern{
		name:       "absolute_path",
		regex:      regexp.MustCompile(`(/[a-zA-Z0-9_\-./]+(?:\.[a-zA-Z0-9]+)?)`),
		toolName:   "read_file",
		confidence: 0.85,
		reason:     "Detected file path",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				return ":read " + matches[1]
			}
			return ":read"
		},
	})

	d.addPattern(pattern{
		name:       "relative_path",
		regex:      regexp.MustCompile(`(\./[a-zA-Z0-9_\-./]+(?:\.[a-zA-Z0-9]+)?)`),
		toolName:   "read_file",
		confidence: 0.80,
		reason:     "Detected relative path",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				return ":read " + matches[1]
			}
			return ":read"
		},
	})

	d.addPattern(pattern{
		name:       "home_path",
		regex:      regexp.MustCompile(`(~/[a-zA-Z0-9_\-./]+(?:\.[a-zA-Z0-9]+)?)`),
		toolName:   "read_file",
		confidence: 0.85,
		reason:     "Detected home path",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				return ":read " + matches[1]
			}
			return ":read"
		},
	})

	// URL patterns - suggest web_fetch
	d.addPattern(pattern{
		name:       "http_url",
		regex:      regexp.MustCompile(`\b(https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=%]+)`),
		toolName:   "web_fetch",
		confidence: 0.90,
		reason:     "Detected URL",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				return ":web " + matches[1]
			}
			return ":web"
		},
	})

	// Search patterns - suggest grep
	d.addPattern(pattern{
		name:       "search_intent",
		regex:      regexp.MustCompile(`(?i)\b(?:search|find|grep|look for)\s+(?:for\s+)?["']?([^"'\n]+)["']?`),
		toolName:   "grep",
		confidence: 0.75,
		reason:     "Detected search intent",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				query := strings.TrimSpace(matches[1])
				query = strings.TrimSuffix(query, " in")
				query = strings.TrimSuffix(query, " for")
				return ":grep " + query
			}
			return ":grep"
		},
	})

	// Command patterns - suggest bash
	d.addPattern(pattern{
		name:       "run_command",
		regex:      regexp.MustCompile(`(?i)\b(?:run|execute|exec)\s+["']?([^"'\n]+)["']?`),
		toolName:   "bash",
		confidence: 0.80,
		reason:     "Detected command",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				cmd := strings.TrimSpace(matches[1])
				return ":bash " + cmd
			}
			return ":bash"
		},
	})

	// Shell command pattern - looks like a shell command
	d.addPattern(pattern{
		name:       "shell_command_like",
		regex:      regexp.MustCompile(`^(ls|cd|pwd|cat|grep|find|git|npm|yarn|go|docker|kubectl)\s+`),
		toolName:   "bash",
		confidence: 0.85,
		reason:     "Looks like shell command",
		actionFunc: func(_ []string) string {
			return ":bash"
		},
	})

	// File glob patterns - suggest glob
	d.addPattern(pattern{
		name:       "glob_pattern",
		regex:      regexp.MustCompile(`\*\*?/[a-zA-Z0-9_\-*./]+|\*\.[a-zA-Z0-9]+`),
		toolName:   "glob",
		confidence: 0.75,
		reason:     "Detected glob pattern",
		actionFunc: func(matches []string) string {
			if len(matches) > 0 {
				return ":glob " + matches[0]
			}
			return ":glob"
		},
	})

	// Write/create file patterns - suggest write_file
	d.addPattern(pattern{
		name:       "write_intent",
		regex:      regexp.MustCompile(`(?i)\b(?:write|create|save)\s+(?:file|to)\s+([/~.]?[a-zA-Z0-9_\-./]+)`),
		toolName:   "write_file",
		confidence: 0.80,
		reason:     "Detected write intent",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				return ":write " + matches[1]
			}
			return ":write"
		},
	})

	// Edit file patterns - suggest edit
	d.addPattern(pattern{
		name:       "edit_intent",
		regex:      regexp.MustCompile(`(?i)\b(?:edit|modify|change|update)\s+(?:file\s+)?([/~.]?[a-zA-Z0-9_\-./]+)`),
		toolName:   "edit",
		confidence: 0.75,
		reason:     "Detected edit intent",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				return ":edit " + matches[1]
			}
			return ":edit"
		},
	})
}

// addPattern adds a pattern to the detector.
func (d *Detector) addPattern(p pattern) {
	d.patterns = append(d.patterns, p)
}

// Analyze implements tux.SuggestionProvider.
func (d *Detector) Analyze(input string) []tux.Suggestion {
	// Don't suggest if input is too short or empty
	if len(strings.TrimSpace(input)) < 3 {
		return nil
	}

	// Don't suggest if input already looks like a tool command
	if strings.HasPrefix(strings.TrimSpace(input), ":") {
		return nil
	}

	// Don't suggest if user is mid-sentence (trailing space or incomplete)
	if strings.HasSuffix(input, " ") && len(input) < 20 {
		return nil
	}

	suggestions := make([]tux.Suggestion, 0)
	seen := make(map[string]bool) // Deduplicate by tool name

	for _, p := range d.patterns {
		matches := p.regex.FindStringSubmatch(input)
		if matches != nil {
			// Only add if we haven't already suggested this tool
			if !seen[p.toolName] {
				suggestions = append(suggestions, tux.Suggestion{
					ToolName:   p.toolName,
					Confidence: p.confidence,
					Reason:     p.reason,
					Action:     p.actionFunc(matches),
				})
				seen[p.toolName] = true
			}
		}
	}

	// Sort by confidence (highest first)
	sortByConfidence(suggestions)

	// Only return high-confidence suggestions (>= 0.70)
	filtered := make([]tux.Suggestion, 0)
	for _, s := range suggestions {
		if s.Confidence >= 0.70 {
			filtered = append(filtered, s)
		}
	}

	// Limit to top 3 suggestions
	if len(filtered) > 3 {
		filtered = filtered[:3]
	}

	return filtered
}

// sortByConfidence sorts suggestions by confidence in descending order.
func sortByConfidence(suggestions []tux.Suggestion) {
	// Simple bubble sort since we have few items
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Confidence > suggestions[i].Confidence {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}
}
