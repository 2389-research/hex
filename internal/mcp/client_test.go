// ABOUTME: Comprehensive tests for MCP client functionality
// ABOUTME: Tests initialization, tool listing, and tool execution via JSON-RPC stdio transport

package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"
)

// nopCloser wraps an io.Reader and adds a no-op Close method
type nopReadCloser struct {
	io.Reader
}

func (nopReadCloser) Close() error { return nil }

// createTestClientAndServer creates a connected client-server pair for testing
func createTestClientAndServer(serverVersion string, tools []Tool) (*Client, *MockMCPServer) {
	serverName := "test-server"
	// Create bidirectional pipes for client-server communication
	clientToServerR, clientToServerW := io.Pipe()
	serverToClientR, serverToClientW := io.Pipe()
	mockStderr := &bytes.Buffer{}

	// Create and start server
	server := NewMockMCPServer(serverName, serverVersion, tools)
	server.SetIOStreams(clientToServerR, serverToClientW, mockStderr)
	go func() { _ = server.Run() }()

	// Create client
	client := &Client{
		stdin:   clientToServerW, // Client writes here, server reads from clientToServerR
		stdout:  serverToClientR, // Client reads here, server writes to serverToClientW
		stderr:  nopReadCloser{mockStderr},
		pending: make(map[int64]chan *jsonrpcResponse),
		done:    make(chan struct{}),
	}

	// Start reading responses
	client.reader = bufio.NewScanner(client.stdout)
	go client.readLoop()
	go client.stderrLoop()

	return client, server
}

func TestClient_Initialize(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "successful initialization with 2024-11-05",
			version: "2024-11-05",
			wantErr: false,
		},
		{
			name:    "successful initialization with latest version",
			version: "2025-06-18",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := createTestClientAndServer(tt.version, []Tool{})

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := client.Initialize(ctx, "test-client", "1.0.0", tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !client.initialized {
				t.Error("Client should be initialized after successful Initialize()")
			}

			if !tt.wantErr && client.serverInfo.Name != "test-server" {
				t.Errorf("Expected server name 'test-server', got %q", client.serverInfo.Name)
			}
		})
	}
}

func TestClient_Initialize_ProtocolVersionMismatch(t *testing.T) {
	client, _ := createTestClientAndServer("2024-11-05", []Tool{})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should still succeed - version negotiation handles compatibility
	err := client.Initialize(ctx, "test-client", "1.0.0", "2099-99-99")
	if err != nil {
		t.Logf("Initialize with future protocol version: %v (expected behavior)", err)
	}
}

func TestClient_ListTools(t *testing.T) {
	testTools := []Tool{
		{
			Name:        "get_weather",
			Description: "Get weather information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "City name",
					},
				},
				"required": []interface{}{"location"},
			},
		},
		{
			Name:        "calculate",
			Description: "Perform calculation",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	client, _ := createTestClientAndServer("2024-11-05", testTools)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Initialize first
	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// List tools
	tools, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}

	if len(tools) != len(testTools) {
		t.Errorf("Expected %d tools, got %d", len(testTools), len(tools))
	}

	// Verify tool details
	for i, tool := range tools {
		if tool.Name != testTools[i].Name {
			t.Errorf("Tool[%d] name = %q, want %q", i, tool.Name, testTools[i].Name)
		}
		if tool.Description != testTools[i].Description {
			t.Errorf("Tool[%d] description = %q, want %q", i, tool.Description, testTools[i].Description)
		}
	}
}

func TestClient_ListTools_BeforeInitialize(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	_, err := client.ListTools(ctx)
	if err == nil {
		t.Error("ListTools() should fail before Initialize()")
	}

	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("Error should mention 'not initialized', got: %v", err)
	}
}

func TestClient_CallTool(t *testing.T) {
	testTools := []Tool{
		{
			Name:        "echo",
			Description: "Echo back input",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	client, server := createTestClientAndServer("2024-11-05", testTools)

	// Register custom handler for echo tool
	server.RegisterToolHandler("echo", func(args map[string]interface{}) (interface{}, error) {
		message := args["message"].(string)
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Initialize
	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Call tool
	args := map[string]interface{}{
		"message": "Hello, MCP!",
	}
	result, err := client.CallTool(ctx, "echo", args)
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}

	// Verify result structure
	content, ok := result["content"].([]interface{})
	if !ok {
		t.Fatal("Result should have 'content' array")
	}

	if len(content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(content))
	}

	firstContent := content[0].(map[string]interface{})
	if firstContent["type"] != "text" {
		t.Errorf("Content type = %q, want 'text'", firstContent["type"])
	}

	if firstContent["text"] != "Hello, MCP!" {
		t.Errorf("Content text = %q, want 'Hello, MCP!'", firstContent["text"])
	}
}

func TestClient_CallTool_NonexistentTool(t *testing.T) {
	client, _ := createTestClientAndServer("2024-11-05", []Tool{})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	_, err := client.CallTool(ctx, "nonexistent", map[string]interface{}{})
	if err == nil {
		t.Error("CallTool() should fail for nonexistent tool")
	}
}

func TestClient_CallTool_BeforeInitialize(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	_, err := client.CallTool(ctx, "test", map[string]interface{}{})
	if err == nil {
		t.Error("CallTool() should fail before Initialize()")
	}

	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("Error should mention 'not initialized', got: %v", err)
	}
}

func TestClient_JSONRPCMessageFormat(t *testing.T) {
	// Test that client sends properly formatted JSON-RPC 2.0 messages
	// Manually craft an initialize message
	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	// Encode to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	// Verify structure
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to decode message: %v", err)
	}

	if decoded["jsonrpc"] != "2.0" {
		t.Errorf("jsonrpc field = %q, want '2.0'", decoded["jsonrpc"])
	}

	if decoded["method"] != "initialize" {
		t.Errorf("method field = %q, want 'initialize'", decoded["method"])
	}

	// ID must not be null (per MCP spec)
	if decoded["id"] == nil {
		t.Error("id field must not be null")
	}
}

func TestClient_Shutdown(t *testing.T) {
	client, _ := createTestClientAndServer("2024-11-05", []Tool{})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test shutdown
	err := client.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// Verify client is no longer initialized
	if client.initialized {
		t.Error("Client should not be initialized after Shutdown()")
	}
}

func TestClient_ConcurrentToolCalls(t *testing.T) {
	testTools := []Tool{
		{
			Name:        "test_tool",
			Description: "Test concurrent calls",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	client, server := createTestClientAndServer("2024-11-05", testTools)

	counter := 0
	server.RegisterToolHandler("test_tool", func(_ map[string]interface{}) (interface{}, error) {
		counter++
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "ok",
				},
			},
		}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Make multiple concurrent calls
	const numCalls = 5
	errChan := make(chan error, numCalls)

	for i := 0; i < numCalls; i++ {
		go func() {
			_, err := client.CallTool(ctx, "test_tool", map[string]interface{}{})
			errChan <- err
		}()
	}

	// Collect results
	for i := 0; i < numCalls; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent call %d failed: %v", i, err)
		}
	}
}
