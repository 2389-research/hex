// ABOUTME: Tool registry for managing available tools
// ABOUTME: Thread-safe registration, retrieval, and listing of tools

package tools

import (
	"fmt"
	"sort"
	"sync"

	"github.com/harper/jeff/internal/core"
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

	// Email tools
	case "send_email":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"to": map[string]interface{}{
					"type":        "string",
					"description": "Recipient email address(es), comma-separated",
				},
				"subject": map[string]interface{}{
					"type":        "string",
					"description": "Email subject line",
				},
				"body": map[string]interface{}{
					"type":        "string",
					"description": "Email body content",
				},
				"cc": map[string]interface{}{
					"type":        "string",
					"description": "CC recipients, comma-separated (optional)",
				},
				"bcc": map[string]interface{}{
					"type":        "string",
					"description": "BCC recipients, comma-separated (optional)",
				},
			},
			"required": []string{"to", "subject", "body"},
		}
	case "search_emails":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query string",
				},
				"from": map[string]interface{}{
					"type":        "string",
					"description": "Filter by sender email (optional)",
				},
				"to": map[string]interface{}{
					"type":        "string",
					"description": "Filter by recipient email (optional)",
				},
				"after": map[string]interface{}{
					"type":        "string",
					"description": "Filter emails after this date (YYYY-MM-DD format, optional)",
				},
				"before": map[string]interface{}{
					"type":        "string",
					"description": "Filter emails before this date (YYYY-MM-DD format, optional)",
				},
				"is_unread": map[string]interface{}{
					"type":        "boolean",
					"description": "Filter by unread status (optional)",
				},
			},
			"required": []string{},
		}
	case "read_email":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier of the email message",
				},
			},
			"required": []string{"message_id"},
		}

	// Calendar tools
	case "create_event":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Event title",
				},
				"start": map[string]interface{}{
					"type":        "string",
					"description": "Start time in ISO 8601 format (e.g., 2025-12-01T14:00:00Z)",
				},
				"end": map[string]interface{}{
					"type":        "string",
					"description": "End time in ISO 8601 format (e.g., 2025-12-01T15:00:00Z)",
				},
				"attendees": map[string]interface{}{
					"type":        "string",
					"description": "Comma-separated list of attendee emails (optional)",
				},
				"location": map[string]interface{}{
					"type":        "string",
					"description": "Event location (optional)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Event description (optional)",
				},
			},
			"required": []string{"title", "start", "end"},
		}
	case "list_events":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"start_date": map[string]interface{}{
					"type":        "string",
					"description": "Start of date range (YYYY-MM-DD)",
				},
				"end_date": map[string]interface{}{
					"type":        "string",
					"description": "End of date range (YYYY-MM-DD)",
				},
			},
			"required": []string{"start_date", "end_date"},
		}

	// Task tools
	case "create_task":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Task title",
				},
				"due_date": map[string]interface{}{
					"type":        "string",
					"description": "Due date in YYYY-MM-DD format (optional)",
				},
				"notes": map[string]interface{}{
					"type":        "string",
					"description": "Additional task notes (optional)",
				},
				"priority": map[string]interface{}{
					"type":        "string",
					"description": "Priority level: low, medium, or high (optional)",
					"enum":        []string{"low", "medium", "high"},
				},
			},
			"required": []string{"title"},
		}
	case "list_tasks":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filter": map[string]interface{}{
					"type":        "string",
					"description": "Filter tasks by status (optional)",
				},
				"due_before": map[string]interface{}{
					"type":        "string",
					"description": "Show tasks due before this date (YYYY-MM-DD, optional)",
				},
			},
			"required": []string{},
		}
	case "complete_task":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier of the task",
				},
			},
			"required": []string{"task_id"},
		}

	default:
		// For other tools, return minimal schema
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}
}
