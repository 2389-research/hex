// ABOUTME: MCP client implementation using JSON-RPC 2.0 over stdio transport
// ABOUTME: Handles initialization, tool listing, and tool execution with MCP servers

// Package mcp provides Model Context Protocol client and server management.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Client represents an MCP client connected to a server via stdio
type Client struct {
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	nextID      atomic.Int64
	mu          sync.Mutex
	pending     map[int64]chan *jsonrpcResponse
	initialized bool
	serverInfo  ServerInfo
	reader      *bufio.Scanner
	done        chan struct{}
}

// ServerInfo contains information about the connected MCP server
type ServerInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// jsonrpcRequest represents a JSON-RPC 2.0 request
type jsonrpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// jsonrpcResponse represents a JSON-RPC 2.0 response
type jsonrpcResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int64                  `json:"id"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *jsonrpcError          `json:"error,omitempty"`
}

// jsonrpcError represents a JSON-RPC 2.0 error object
type jsonrpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewClient creates a new MCP client that will connect to the specified command
func NewClient(command string, args ...string) (*Client, error) {
	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	client := &Client{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		pending: make(map[int64]chan *jsonrpcResponse),
		done:    make(chan struct{}),
	}

	// Start the server process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Start reading responses
	client.reader = bufio.NewScanner(stdout)
	go client.readLoop()

	// Start reading stderr for logging
	go client.stderrLoop()

	return client, nil
}

// Initialize performs the MCP initialization handshake
func (c *Client) Initialize(ctx context.Context, clientName, clientVersion, protocolVersion string) error {
	params := map[string]interface{}{
		"protocolVersion": protocolVersion,
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    clientName,
			"version": clientVersion,
		},
	}

	resp, err := c.sendRequest(ctx, "initialize", params)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s (code %d)", resp.Error.Message, resp.Error.Code)
	}

	// Extract server info
	if serverInfo, ok := resp.Result["serverInfo"].(map[string]interface{}); ok {
		if name, ok := serverInfo["name"].(string); ok {
			c.serverInfo.Name = name
		}
		if version, ok := serverInfo["version"].(string); ok {
			c.serverInfo.Version = version
		}
	}

	if capabilities, ok := resp.Result["capabilities"].(map[string]interface{}); ok {
		c.serverInfo.Capabilities = capabilities
	}

	// Send initialized notification
	if err := c.sendNotification("notifications/initialized", nil); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	c.mu.Lock()
	c.initialized = true
	c.mu.Unlock()

	return nil
}

// ListTools retrieves the list of available tools from the server
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	if !c.isInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	resp, err := c.sendRequest(ctx, "tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("tools/list request failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list error: %s (code %d)", resp.Error.Message, resp.Error.Code)
	}

	// Parse tools from response
	toolsData, ok := resp.Result["tools"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tools/list response: missing tools array")
	}

	var tools []Tool
	for _, toolData := range toolsData {
		toolMap, ok := toolData.(map[string]interface{})
		if !ok {
			continue
		}

		tool := Tool{
			Name:        getString(toolMap, "name"),
			Description: getString(toolMap, "description"),
		}

		if schema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
			tool.InputSchema = schema
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

// CallTool invokes a tool on the server with the given arguments
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (map[string]interface{}, error) {
	if !c.isInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}

	resp, err := c.sendRequest(ctx, "tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("tools/call request failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tool '%s' error: %s (code %d)", name, resp.Error.Message, resp.Error.Code)
	}

	return resp.Result, nil
}

// Shutdown gracefully shuts down the client connection
func (c *Client) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Return early if already shut down
	if !c.initialized {
		return nil
	}

	c.initialized = false

	// Close stdin to signal server to exit
	if c.stdin != nil {
		_ = c.stdin.Close()
	}

	// Wait for command to exit (unlock first to avoid deadlock)
	c.mu.Unlock()
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Wait()
	}
	c.mu.Lock()

	// Close done channel safely (only once)
	select {
	case <-c.done:
		// Already closed, do nothing
	default:
		close(c.done)
	}

	return nil
}

// sendRequest sends a JSON-RPC request and waits for the response
func (c *Client) sendRequest(ctx context.Context, method string, params interface{}) (*jsonrpcResponse, error) {
	id := c.nextID.Add(1)

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Create response channel
	respChan := make(chan *jsonrpcResponse, 1)
	c.mu.Lock()
	c.pending[id] = respChan
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	// Send request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	data = append(data, '\n')

	if _, err := c.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respChan:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, fmt.Errorf("client shutdown")
	}
}

// sendNotification sends a JSON-RPC notification (no response expected)
func (c *Client) sendNotification(method string, params interface{}) error {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
	}

	if params != nil {
		req["params"] = params
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	data = append(data, '\n')

	if _, err := c.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return nil
}

// readLoop reads and dispatches JSON-RPC responses
func (c *Client) readLoop() {
	for c.reader.Scan() {
		line := c.reader.Bytes()
		if len(line) == 0 {
			continue
		}

		var resp jsonrpcResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			// Log parse error but continue
			continue
		}

		// Dispatch to waiting request
		c.mu.Lock()
		if ch, ok := c.pending[resp.ID]; ok {
			ch <- &resp
		}
		c.mu.Unlock()
	}
}

// stderrLoop reads and logs stderr output from the server
func (c *Client) stderrLoop() {
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		// In production, this would go to a proper logger
		// For now, we just consume it to prevent blocking
		_ = scanner.Text()
	}
}

// isInitialized checks if the client has been initialized
func (c *Client) isInitialized() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.initialized
}

// getString safely extracts a string from a map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
