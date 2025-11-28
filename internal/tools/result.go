// ABOUTME: Tool execution result types and structures
// ABOUTME: Represents success/failure, output, and metadata from tool execution

package tools

// Result represents the output of a tool execution
type Result struct {
	ToolName string                 // Tool that was executed
	Success  bool                   // Did it succeed?
	Output   string                 // Standard output/result
	Error    string                 // Error message if failed
	Metadata map[string]interface{} // Additional metadata (file path, exit code, etc.)
}
