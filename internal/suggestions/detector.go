// ABOUTME: Pattern-based tool suggestion system
// ABOUTME: Analyzes user input to suggest relevant tools based on context patterns

package suggestions

import (
	"regexp"
	"strings"
)

// Suggestion represents a tool suggestion with confidence score
type Suggestion struct {
	ToolName   string  // Name of the suggested tool
	Confidence float64 // Confidence score (0.0 - 1.0)
	Reason     string  // Human-readable reason for the suggestion
	Action     string  // Suggested action text (e.g., ":read /path/to/file")
}

// Detector analyzes user input to suggest appropriate tools
type Detector struct {
	patterns []pattern
}

// pattern represents a detection pattern for a specific tool
type pattern struct {
	name       string
	regex      *regexp.Regexp
	toolName   string
	confidence float64
	reason     string
	actionFunc func(matches []string) string
}

// NewDetector creates a new suggestion detector with default patterns
func NewDetector() *Detector {
	d := &Detector{}
	d.registerDefaultPatterns()
	return d
}

// registerDefaultPatterns registers all built-in detection patterns
func (d *Detector) registerDefaultPatterns() {
	// File path patterns - suggest read_file
	d.addPattern(pattern{
		name:       "absolute_path",
		regex:      regexp.MustCompile(`(/[a-zA-Z0-9_\-./]+(?:\.[a-zA-Z0-9]+)?)`),
		toolName:   "read_file",
		confidence: 0.85,
		reason:     "Detected absolute file path",
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
		reason:     "Detected relative file path",
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
		reason:     "Detected home directory path",
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
		reason:     "Detected HTTP/HTTPS URL",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				return ":web_fetch " + matches[1]
			}
			return ":web_fetch"
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
				// Clean up the query - remove common words
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
		reason:     "Detected command execution intent",
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
		reason:     "Input looks like a shell command",
		actionFunc: func(matches []string) string {
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
		reason:     "Detected file write intent",
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
		reason:     "Detected file edit intent",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				return ":edit " + matches[1]
			}
			return ":edit"
		},
	})

	// Web search patterns - suggest web_search
	d.addPattern(pattern{
		name:       "web_search_intent",
		regex:      regexp.MustCompile(`(?i)\b(?:google|search for|look up|web search)\s+["']?([^"'\n]+)["']?`),
		toolName:   "web_search",
		confidence: 0.80,
		reason:     "Detected web search intent",
		actionFunc: func(matches []string) string {
			if len(matches) > 1 {
				query := strings.TrimSpace(matches[1])
				return ":web_search " + query
			}
			return ":web_search"
		},
	})
}

// addPattern adds a pattern to the detector
func (d *Detector) addPattern(p pattern) {
	d.patterns = append(d.patterns, p)
}

// AnalyzeInput analyzes user input and returns suggestions
func (d *Detector) AnalyzeInput(input string) []Suggestion {
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

	suggestions := make([]Suggestion, 0)
	seen := make(map[string]bool) // Deduplicate by tool name

	for _, p := range d.patterns {
		matches := p.regex.FindStringSubmatch(input)
		if matches != nil {
			// Only add if we haven't already suggested this tool
			if !seen[p.toolName] {
				suggestions = append(suggestions, Suggestion{
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
	filtered := make([]Suggestion, 0)
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

// sortByConfidence sorts suggestions by confidence in descending order
func sortByConfidence(suggestions []Suggestion) {
	// Simple bubble sort since we have few items
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Confidence > suggestions[i].Confidence {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}
}
