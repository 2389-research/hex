// ABOUTME: Mock MCP server implementation for testing MCP client functionality
// ABOUTME: Implements JSON-RPC 2.0 over stdio transport for test scenarios

package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// MockMCPServer simulates an MCP server for testing
type MockMCPServer struct {
	Name         string
	Version      string
	Tools        []Tool
	stdin        io.Reader
	stdout       io.Writer
	stderr       io.Writer
	nextID       int
	mu           sync.Mutex
	initialized  bool
	toolHandlers map[string]func(map[string]interface{}) (interface{}, error)
}

// NewMockMCPServer creates a new mock server
func NewMockMCPServer(name, version string, tools []Tool) *MockMCPServer {
	return &MockMCPServer{
		Name:         name,
		Version:      version,
		Tools:        tools,
		stdin:        os.Stdin,
		stdout:       os.Stdout,
		stderr:       os.Stderr,
		toolHandlers: make(map[string]func(map[string]interface{}) (interface{}, error)),
	}
}

// SetIOStreams sets custom I/O streams for testing
func (s *MockMCPServer) SetIOStreams(stdin io.Reader, stdout, stderr io.Writer) {
	s.stdin = stdin
	s.stdout = stdout
	s.stderr = stderr
}

// RegisterToolHandler registers a custom handler for a tool
func (s *MockMCPServer) RegisterToolHandler(toolName string, handler func(map[string]interface{}) (interface{}, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolHandlers[toolName] = handler
}

// Run starts the mock server and processes JSON-RPC messages
func (s *MockMCPServer) Run() error {
	scanner := bufio.NewScanner(s.stdin)
	encoder := json.NewEncoder(s.stdout)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(line, &msg); err != nil {
			s.sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		// Handle request
		if method, ok := msg["method"].(string); ok {
			id := msg["id"]
			params, _ := msg["params"].(map[string]interface{})

			switch method {
			case "initialize":
				s.handleInitialize(id, params, encoder)
			case "notifications/initialized":
				s.handleInitialized()
			case "tools/list":
				s.handleToolsList(id, encoder)
			case "tools/call":
				s.handleToolsCall(id, params, encoder)
			default:
				s.sendError(id, -32601, "Method not found", method)
			}
		}
	}

	return scanner.Err()
}

func (s *MockMCPServer) handleInitialize(id interface{}, params map[string]interface{}, encoder *json.Encoder) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Extract protocol version
	protocolVersion, _ := params["protocolVersion"].(string)

	// Log to stderr (as real servers do)
	fmt.Fprintf(s.stderr, "Initializing with protocol version: %s\n", protocolVersion)

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"protocolVersion": s.Version,
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    s.Name,
				"version": "1.0.0",
			},
		},
	}

	encoder.Encode(response)
}

func (s *MockMCPServer) handleInitialized() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.initialized = true
	fmt.Fprintf(s.stderr, "Server initialized and ready\n")
}

func (s *MockMCPServer) handleToolsList(id interface{}, encoder *json.Encoder) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		s.sendErrorWithEncoder(encoder, id, -32002, "Server not initialized", "")
		return
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": s.Tools,
		},
	}

	encoder.Encode(response)
}

func (s *MockMCPServer) handleToolsCall(id interface{}, params map[string]interface{}, encoder *json.Encoder) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		s.sendErrorWithEncoder(encoder, id, -32002, "Server not initialized", "")
		return
	}

	toolName, _ := params["name"].(string)
	arguments, _ := params["arguments"].(map[string]interface{})

	// Check if tool exists
	found := false
	for _, tool := range s.Tools {
		if tool.Name == toolName {
			found = true
			break
		}
	}

	if !found {
		s.sendErrorWithEncoder(encoder, id, -32602, "Tool not found", toolName)
		return
	}

	// Call custom handler if registered
	var result interface{}
	if handler, ok := s.toolHandlers[toolName]; ok {
		res, err := handler(arguments)
		if err != nil {
			s.sendErrorWithEncoder(encoder, id, -32603, "Tool execution failed", err.Error())
			return
		}
		result = res
	} else {
		// Default response
		result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Mock response for %s with args: %v", toolName, arguments),
				},
			},
		}
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}

	encoder.Encode(response)
}

func (s *MockMCPServer) sendError(id interface{}, code int, message, data string) {
	encoder := json.NewEncoder(s.stdout)
	s.sendErrorWithEncoder(encoder, id, code, message, data)
}

func (s *MockMCPServer) sendErrorWithEncoder(encoder *json.Encoder, id interface{}, code int, message, data string) {
	errResp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	if data != "" {
		errResp["error"].(map[string]interface{})["data"] = data
	}

	encoder.Encode(errResp)
}

// MockServerCommand represents a test helper to launch mock servers
type MockServerCommand struct {
	Name    string
	Version string
	Tools   []Tool
}

// BuildMockServer creates a mock server command for testing
func BuildMockServer(name, version string, tools []Tool) *MockServerCommand {
	return &MockServerCommand{
		Name:    name,
		Version: version,
		Tools:   tools,
	}
}

// AsCommand returns the command line that would launch this mock server
func (cmd *MockServerCommand) AsCommand() []string {
	return []string{"go", "run", "./testdata/mock_server.go"}
}
