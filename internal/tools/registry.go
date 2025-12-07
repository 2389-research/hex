// ABOUTME: Tool registry for managing available tools
// ABOUTME: Thread-safe registration, retrieval, and listing of tools

package tools

import (
	"fmt"
	"sort"
	"sync"

	"github.com/2389-research/hex/internal/core"
)

// Registry manages available tools
type Registry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name())
	}

	r.tools[tool.Name()] = tool
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// List returns all registered tool names sorted alphabetically
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetDefinitions returns tool definitions for all registered tools in API format
func (r *Registry) GetDefinitions() []core.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]core.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		def := core.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: getToolSchema(tool.Name()),
		}
		defs = append(defs, def)
	}
	return defs
}

// getToolSchema returns the JSON Schema for a specific tool's input parameters
func getToolSchema(toolName string) map[string]interface{} {
	switch toolName {
	case "read_file":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Optional offset to start reading from (default: 0)",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Optional maximum number of bytes to read (default: entire file)",
				},
			},
			"required": []string{"path"},
		}
	case "write_file":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to write",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write to the file",
				},
				"mode": map[string]interface{}{
					"type":        "string",
					"description": "Write mode: 'create' (fail if exists), 'overwrite', or 'append' (default: create)",
					"enum":        []string{"create", "overwrite", "append"},
				},
			},
			"required": []string{"path", "content"},
		}
	case "bash":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Shell command to execute",
				},
				"timeout": map[string]interface{}{
					"type":        "number",
					"description": "Optional timeout in seconds (default: 30, max: 300)",
				},
				"working_dir": map[string]interface{}{
					"type":        "string",
					"description": "Optional working directory for the command",
				},
				"run_in_background": map[string]interface{}{
					"type":        "boolean",
					"description": "Run command in background and return immediately with bash_id",
				},
			},
			"required": []string{"command"},
		}
	case "Skill":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Name of the skill to invoke (e.g., 'test-driven-development', 'systematic-debugging')",
				},
			},
			"required": []string{"command"},
		}
	case "grep":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The regular expression pattern to search for in file contents",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File or directory to search in (rg PATH). Defaults to current working directory.",
				},
				"output_mode": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"content", "files_with_matches", "count"},
					"description": "Output mode: \"content\" shows matching lines, \"files_with_matches\" shows file paths (default), \"count\" shows match counts",
				},
				"-i": map[string]interface{}{
					"type":        "boolean",
					"description": "Case insensitive search (rg -i)",
				},
				"-A": map[string]interface{}{
					"type":        "number",
					"description": "Number of lines to show after each match (rg -A). Requires output_mode: \"content\", ignored otherwise.",
				},
				"-B": map[string]interface{}{
					"type":        "number",
					"description": "Number of lines to show before each match (rg -B). Requires output_mode: \"content\", ignored otherwise.",
				},
				"-C": map[string]interface{}{
					"type":        "number",
					"description": "Number of lines to show before and after each match (rg -C). Requires output_mode: \"content\", ignored otherwise.",
				},
				"glob": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern to filter files (e.g. \"*.js\", \"*.{ts,tsx}\") - maps to rg --glob",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "File type to search (rg --type). Common types: js, py, rust, go, java, etc.",
				},
			},
			"required": []string{"pattern"},
		}
	case "glob":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The glob pattern to match files against",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The directory to search in. If not specified, the current working directory will be used.",
				},
			},
			"required": []string{"pattern"},
		}
	case "edit":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The absolute path to the file to modify",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "The exact text to replace (must be unique in file unless replace_all is true)",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "The text to replace it with (must be different from old_string)",
				},
				"replace_all": map[string]interface{}{
					"type":        "boolean",
					"description": "Replace all occurrences of old_string (default false). Use for renaming across file.",
				},
			},
			"required": []string{"file_path", "old_string", "new_string"},
		}
	default:
		// For other tools, return minimal schema
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}
}
