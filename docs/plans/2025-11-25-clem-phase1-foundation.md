# Clem Phase 1: Foundation - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the foundational components of Clem CLI - project setup, CLI framework, configuration, basic API client, and print mode.

**Architecture:** Go monolith using Cobra for CLI, Viper for config, stdlib HTTP for API client. Focus on working end-to-end flow before adding complexity.

**Tech Stack:** Go 1.21+, Cobra, Viper, godotenv, testify

**Success Criteria:** `clem --print "hello"` successfully calls Anthropic API and prints response.

---

## Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `.gitignore`
- Create: `README.md`
- Create: `Makefile`

**Step 1: Initialize Go module**

```bash
cd /Users/harper/workspace/2389/cc-deobfuscate/clean
go mod init github.com/yourusername/clem
```

Expected output: `go: creating new go.mod: module github.com/yourusername/clem`

**Step 2: Create .gitignore**

Create `.gitignore`:
```
# Binaries
clem
*.exe
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Vendor directory
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Build artifacts
dist/
build/

# Config
.env
```

**Step 3: Create Makefile**

Create `Makefile`:
```makefile
.PHONY: build test clean install run

build:
	go build -o clem ./cmd/clem

test:
	go test -v -race ./...

test-short:
	go test -v -short ./...

clean:
	rm -f clem
	go clean

install:
	go install ./cmd/clem

run:
	go run ./cmd/clem

lint:
	golangci-lint run

.DEFAULT_GOAL := build
```

**Step 4: Create basic README**

Create `README.md`:
```markdown
# Clem

AI assistant CLI built in Go.

## Installation

```bash
go install github.com/yourusername/clem/cmd/clem@latest
```

## Usage

```bash
# Interactive mode
clem

# Print mode
clem --print "your prompt"
```

## Development

```bash
# Build
make build

# Test
make test

# Run
make run
```
```

**Step 5: Commit**

```bash
git add go.mod .gitignore README.md Makefile
git commit -m "feat: initialize Go project structure"
```

---

## Task 2: Project Structure and Core Types

**Files:**
- Create: `cmd/clem/main.go`
- Create: `internal/core/types.go`
- Create: `internal/core/types_test.go`

**Step 1: Create directory structure**

```bash
mkdir -p cmd/clem
mkdir -p internal/core
mkdir -p internal/ui
mkdir -p internal/storage
mkdir -p internal/tools
mkdir -p internal/mcp
mkdir -p internal/plugins
mkdir -p pkg/plugin
mkdir -p tests/{unit,scenarios,golden}
```

**Step 2: Write test for core types**

Create `internal/core/types_test.go`:
```go
package core_test

import (
	"testing"

	"github.com/yourusername/clem/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	msg := core.Message{
		Role:    "user",
		Content: "Hello",
	}

	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "Hello", msg.Content)
}

func TestMessageRole(t *testing.T) {
	tests := []struct {
		role  string
		valid bool
	}{
		{"user", true},
		{"assistant", true},
		{"system", true},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			msg := core.Message{Role: tt.role}
			err := msg.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
```

**Step 3: Run test to verify it fails**

```bash
go mod tidy
go test ./internal/core/...
```

Expected: FAIL with "cannot find package"

**Step 4: Install testing dependencies**

```bash
go get github.com/stretchr/testify/assert
go mod tidy
```

**Step 5: Implement core types**

Create `internal/core/types.go`:
```go
package core

import (
	"fmt"
	"time"
)

// Message represents a single message in a conversation
type Message struct {
	ID        string    `json:"id,omitempty"`
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	ToolCalls []ToolUse `json:"tool_calls,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// Validate checks if the message is valid
func (m *Message) Validate() error {
	switch m.Role {
	case "user", "assistant", "system":
		return nil
	default:
		return fmt.Errorf("invalid role: %s", m.Role)
	}
}

// ToolUse represents a tool invocation
type ToolUse struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Input  map[string]interface{} `json:"input"`
	Output string                 `json:"output,omitempty"`
}

// ToolDefinition defines a tool's schema
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// MessageRequest is sent to the API
type MessageRequest struct {
	Model     string           `json:"model"`
	Messages  []Message        `json:"messages"`
	MaxTokens int              `json:"max_tokens"`
	Stream    bool             `json:"stream,omitempty"`
	Tools     []ToolDefinition `json:"tools,omitempty"`
	System    string           `json:"system,omitempty"`
}

// MessageResponse is received from the API
type MessageResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Role      string    `json:"role"`
	Content   []Content `json:"content"`
	Model     string    `json:"model"`
	StopReason string   `json:"stop_reason,omitempty"`
	Usage     Usage     `json:"usage"`
}

// Content represents a content block in the response
type Content struct {
	Type    string                 `json:"type"` // "text" or "tool_use"
	Text    string                 `json:"text,omitempty"`
	ID      string                 `json:"id,omitempty"`
	Name    string                 `json:"name,omitempty"`
	Input   map[string]interface{} `json:"input,omitempty"`
}

// Usage tracks token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// StreamChunk represents a chunk in streaming response
type StreamChunk struct {
	Type    string  `json:"type"`
	Delta   *Delta  `json:"delta,omitempty"`
	Content *Content `json:"content,omitempty"`
	Done    bool    `json:"-"`
}

// Delta represents incremental content in streaming
type Delta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
```

**Step 6: Run test to verify it passes**

```bash
go test ./internal/core/... -v
```

Expected: PASS (2 tests)

**Step 7: Create minimal main.go**

Create `cmd/clem/main.go`:
```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Clem - AI Assistant CLI")
	fmt.Println("Version: 0.1.0")
	os.Exit(0)
}
```

**Step 8: Test build**

```bash
make build
./clem
```

Expected output:
```
Clem - AI Assistant CLI
Version: 0.1.0
```

**Step 9: Commit**

```bash
git add cmd/ internal/core/
git commit -m "feat: add core types and minimal main"
```

---

## Task 3: CLI Framework with Cobra

**Files:**
- Modify: `cmd/clem/main.go`
- Create: `cmd/clem/root.go`
- Create: `cmd/clem/root_test.go`

**Step 1: Install Cobra**

```bash
go get github.com/spf13/cobra@latest
go mod tidy
```

**Step 2: Write test for root command**

Create `cmd/clem/root_test.go`:
```go
package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	// Test --help flag
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestVersionFlag(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "0.1.0")
}

func TestPrintFlag(t *testing.T) {
	rootCmd.SetArgs([]string{"--print", "test"})
	// Should not panic
	// Actual functionality tested in integration tests
}
```

**Step 3: Run test to verify it fails**

```bash
go test ./cmd/clem/... -v
```

Expected: FAIL with "undefined: rootCmd"

**Step 4: Implement root command**

Create `cmd/clem/root.go`:
```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information
	version = "0.1.0"

	// Global flags
	printMode    bool
	outputFormat string
	model        string
	verbose      bool
	debug        string
)

var rootCmd = &cobra.Command{
	Use:   "clem [prompt]",
	Short: "Clem - AI assistant CLI",
	Long: `Clem is an AI assistant for your terminal.

Start an interactive session or use --print for one-off queries.`,
	Version: version,
	Args:    cobra.ArbitraryArgs,
	RunE:    runRoot,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&printMode, "print", "p", false, "Print mode (non-interactive)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", "text", "Output format: text, json, stream-json")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "claude-sonnet-4-5-20250929", "Model to use")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&debug, "debug", "", "Debug categories")
}

func runRoot(cmd *cobra.Command, args []string) error {
	prompt := ""
	if len(args) > 0 {
		prompt = joinArgs(args)
	}

	if printMode {
		return runPrintMode(prompt)
	}

	return runInteractive(prompt)
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

func runPrintMode(prompt string) error {
	if prompt == "" {
		return fmt.Errorf("prompt required in print mode")
	}
	fmt.Printf("Print mode: %s\n", prompt)
	return nil
}

func runInteractive(prompt string) error {
	fmt.Println("Interactive mode not yet implemented")
	return nil
}
```

**Step 5: Update main.go**

Modify `cmd/clem/main.go`:
```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 6: Run tests to verify they pass**

```bash
go test ./cmd/clem/... -v
```

Expected: PASS (3 tests)

**Step 7: Test CLI manually**

```bash
make build
./clem --help
./clem --version
./clem --print "hello"
```

Expected outputs:
- Help text displayed
- Version: 0.1.0
- Print mode: hello

**Step 8: Commit**

```bash
git add cmd/clem/
git commit -m "feat: add Cobra CLI framework with flags"
```

---

## Task 4: Configuration System with Viper

**Files:**
- Create: `internal/core/config.go`
- Create: `internal/core/config_test.go`
- Create: `.env.example`

**Step 1: Install dependencies**

```bash
go get github.com/spf13/viper@latest
go get github.com/joho/godotenv@latest
go mod tidy
```

**Step 2: Write test for config loading**

Create `internal/core/config_test.go`:
```go
package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/clem/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `api_key: test-key-123
model: claude-sonnet-4-5-20250929
`
	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	require.NoError(t, err)

	// Set config path
	os.Setenv("JEFF_CONFIG_PATH", configPath)
	defer os.Unsetenv("JEFF_CONFIG_PATH")

	// Load config
	cfg, err := core.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "test-key-123", cfg.APIKey)
	assert.Equal(t, "claude-sonnet-4-5-20250929", cfg.Model)
}

func TestConfigFromEnv(t *testing.T) {
	os.Setenv("JEFF_API_KEY", "env-key-456")
	os.Setenv("JEFF_MODEL", "claude-opus-4-5-20250929")
	defer os.Unsetenv("JEFF_API_KEY")
	defer os.Unsetenv("JEFF_MODEL")

	cfg, err := core.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "env-key-456", cfg.APIKey)
	assert.Equal(t, "claude-opus-4-5-20250929", cfg.Model)
}

func TestConfigPrecedence(t *testing.T) {
	// Env var should override config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `api_key: file-key`
	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	require.NoError(t, err)

	os.Setenv("JEFF_CONFIG_PATH", configPath)
	os.Setenv("JEFF_API_KEY", "env-key")
	defer os.Unsetenv("JEFF_CONFIG_PATH")
	defer os.Unsetenv("JEFF_API_KEY")

	cfg, err := core.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "env-key", cfg.APIKey, "env var should override file")
}
```

**Step 3: Run test to verify it fails**

```bash
go test ./internal/core/... -v -run TestLoadConfig
```

Expected: FAIL with "undefined: core.LoadConfig"

**Step 4: Implement config loading**

Create `internal/core/config.go`:
```go
package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	APIKey         string   `mapstructure:"api_key"`
	Model          string   `mapstructure:"model"`
	DefaultTools   []string `mapstructure:"default_tools"`
	PermissionMode string   `mapstructure:"permission_mode"`
}

// LoadConfig loads configuration from multiple sources
// Priority (highest to lowest):
// 1. Environment variables (JEFF_*)
// 2. .env file (current directory)
// 3. ~/.clem/config.yaml
// 4. Defaults
func LoadConfig() (*Config, error) {
	// Load .env file if it exists (don't error if missing)
	_ = godotenv.Load()

	v := viper.New()

	// Set defaults
	v.SetDefault("model", "claude-sonnet-4-5-20250929")
	v.SetDefault("permission_mode", "ask")
	v.SetDefault("default_tools", []string{"Bash", "Read", "Write", "Edit", "Grep"})

	// Environment variables
	v.SetEnvPrefix("CLEM")
	v.AutomaticEnv()

	// Config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Check for custom config path
	if configPath := os.Getenv("JEFF_CONFIG_PATH"); configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Add search paths
		v.AddConfigPath(".") // Current directory
		home, err := os.UserHomeDir()
		if err == nil {
			clemDir := filepath.Join(home, ".clem")
			v.AddConfigPath(clemDir)
		}
	}

	// Read config file (ignore error if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// GetAPIKey returns the API key from config or environment
func (c *Config) GetAPIKey() (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("API key not configured. Set JEFF_API_KEY or run 'clem setup-token'")
	}
	return c.APIKey, nil
}
```

**Step 5: Run tests to verify they pass**

```bash
go test ./internal/core/... -v
```

Expected: PASS (all tests)

**Step 6: Create .env.example**

Create `.env.example`:
```bash
# Anthropic API Key
# Get yours at: https://console.anthropic.com/
JEFF_API_KEY=sk-ant-api03-...

# Optional: Model to use
JEFF_MODEL=claude-sonnet-4-5-20250929

# Optional: Permission mode (ask, allow, deny, plan)
JEFF_PERMISSION_MODE=ask
```

**Step 7: Update .gitignore to exclude config**

Add to `.gitignore`:
```
.env
config.yaml
```

**Step 8: Commit**

```bash
git add internal/core/config.go internal/core/config_test.go .env.example .gitignore
git commit -m "feat: add configuration system with Viper"
```

---

## Task 5: Anthropic API Client (Basic)

**Files:**
- Create: `internal/core/client.go`
- Create: `internal/core/client_test.go`

**Step 1: Write test for API client (with VCR)**

Install VCR dependency:
```bash
go get github.com/dnaeon/go-vcr/v2@latest
go mod tidy
```

Create `internal/core/client_test.go`:
```go
package core_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/dnaeon/go-vcr/v2/recorder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/clem/internal/core"
)

func TestCreateMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	// Create VCR recorder
	r, err := recorder.New("testdata/fixtures/create_message")
	require.NoError(t, err)
	defer r.Stop()

	// Create HTTP client with recorder
	httpClient := &http.Client{Transport: r}

	// Create API client
	client := core.NewClient("test-api-key", core.WithHTTPClient(httpClient))

	// Test request
	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 1024,
		Messages: []core.Message{
			{Role: "user", Content: "Say hello"},
		},
	}

	// Execute
	resp, err := client.CreateMessage(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)

	// Check response structure
	assert.Equal(t, "assistant", resp.Role)
	assert.NotEmpty(t, resp.ID)
}

func TestClientError(t *testing.T) {
	client := core.NewClient("invalid-key")

	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 1024,
		Messages: []core.Message{
			{Role: "user", Content: "test"},
		},
	}

	_, err := client.CreateMessage(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}
```

**Step 2: Run test to verify it fails**

```bash
mkdir -p internal/core/testdata/fixtures
go test ./internal/core/... -v -run TestCreateMessage
```

Expected: FAIL with "undefined: core.NewClient"

**Step 3: Implement API client**

Create `internal/core/client.go`:
```go
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.anthropic.com/v1"
	apiVersion     = "2023-06-01"
)

// Client is the Anthropic API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ClientOption configures a Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithBaseURL sets a custom base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// NewClient creates a new API client
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// CreateMessage sends a message to the API
func (c *Client) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/messages",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Unmarshal response
	var resp MessageResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// GetTextContent extracts text content from response
func (r *MessageResponse) GetTextContent() string {
	for _, content := range r.Content {
		if content.Type == "text" {
			return content.Text
		}
	}
	return ""
}
```

**Step 4: Run tests**

```bash
go test ./internal/core/... -v
```

Expected: First run will record API call (needs real API key), subsequent runs replay from cassette

To run with real API:
```bash
export JEFF_API_KEY=your-key-here
go test ./internal/core/... -v -run TestCreateMessage
```

**Step 5: Commit**

```bash
git add internal/core/client.go internal/core/client_test.go
git commit -m "feat: add Anthropic API client with VCR testing"
```

---

## Task 6: Wire Print Mode to API

**Files:**
- Modify: `cmd/clem/root.go`
- Create: `cmd/clem/print.go`

**Step 1: Implement print mode handler**

Create `cmd/clem/print.go`:
```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yourusername/clem/internal/core"
)

func runPrintMode(prompt string) error {
	if prompt == "" {
		return fmt.Errorf("prompt required in print mode")
	}

	// Load config
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Get API key
	apiKey, err := cfg.GetAPIKey()
	if err != nil {
		return err
	}

	// Create client
	client := core.NewClient(apiKey)

	// Create request
	req := core.MessageRequest{
		Model:     model, // from global flag
		MaxTokens: 4096,
		Messages: []core.Message{
			{Role: "user", Content: prompt},
		},
	}

	// Send request
	resp, err := client.CreateMessage(context.Background(), req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}

	// Format output
	return formatOutput(resp, outputFormat)
}

func formatOutput(resp *core.MessageResponse, format string) error {
	switch format {
	case "text":
		fmt.Println(resp.GetTextContent())
	case "json":
		encoder := json.NewEncoder(fmt.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("encode JSON: %w", err)
		}
	case "stream-json":
		return fmt.Errorf("streaming not yet implemented")
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
	return nil
}
```

**Step 2: Update root command**

Modify `cmd/clem/root.go` - replace stub `runPrintMode`:
```go
// Remove the stub runPrintMode function
// It's now in print.go
```

**Step 3: Build and test manually**

```bash
make build

# Test with environment variable
export JEFF_API_KEY=your-key-here
./clem --print "Say hello in one sentence"
```

Expected: Actual response from Claude API

**Step 4: Test JSON output**

```bash
./clem --print --output-format json "Say hello"
```

Expected: JSON formatted response

**Step 5: Test error handling**

```bash
unset JEFF_API_KEY
./clem --print "test"
```

Expected: Error message about missing API key

**Step 6: Commit**

```bash
git add cmd/clem/
git commit -m "feat: wire print mode to Anthropic API"
```

---

## Task 7: Setup Token Command

**Files:**
- Create: `cmd/clem/setup.go`
- Modify: `cmd/clem/root.go`

**Step 1: Implement setup-token command**

Create `cmd/clem/setup.go`:
```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var setupCmd = &cobra.Command{
	Use:   "setup-token [token]",
	Short: "Configure API token",
	Long: `Configure your Anthropic API token.

Get your API key from: https://console.anthropic.com/

This command will save your API key to ~/.clem/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	var apiKey string

	if len(args) > 0 {
		apiKey = args[0]
	} else {
		fmt.Println("Usage: clem setup-token <your-api-key>")
		fmt.Println("\nGet your API key from: https://console.anthropic.com/")
		return nil
	}

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	// Create .clem directory
	clemDir := filepath.Join(home, ".clem")
	if err := os.MkdirAll(clemDir, 0755); err != nil {
		return fmt.Errorf("create .clem dir: %w", err)
	}

	// Write config
	configPath := filepath.Join(clemDir, "config.yaml")
	config := map[string]string{
		"api_key": apiKey,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("✓ API key configured successfully\n")
	fmt.Printf("  Saved to: %s\n", configPath)
	return nil
}
```

**Step 2: Install yaml dependency**

```bash
go get gopkg.in/yaml.v3@latest
go mod tidy
```

**Step 3: Build and test**

```bash
make build
./clem setup-token test-key-123
```

Expected output:
```
✓ API key configured successfully
  Saved to: /Users/harper/.clem/config.yaml
```

**Step 4: Verify config file**

```bash
cat ~/.clem/config.yaml
```

Expected:
```yaml
api_key: test-key-123
```

**Step 5: Test that config is loaded**

```bash
./clem --print "test"
```

Should use the configured API key

**Step 6: Commit**

```bash
git add cmd/clem/setup.go
git commit -m "feat: add setup-token command"
```

---

## Task 8: Doctor Command

**Files:**
- Create: `cmd/clem/doctor.go`
- Modify: `cmd/clem/root.go`

**Step 1: Implement doctor command**

Create `cmd/clem/doctor.go`:
```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yourusername/clem/internal/core"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check installation health",
	Long:  "Verify that Clem is correctly installed and configured",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("Clem Health Check")
	fmt.Println("=================\n")

	checks := []check{
		checkHomeDirectory,
		checkConfigFile,
		checkAPIKey,
	}

	allPassed := true
	for _, check := range checks {
		if !check() {
			allPassed = false
		}
		fmt.Println()
	}

	if allPassed {
		fmt.Println("✓ All checks passed")
	} else {
		fmt.Println("⚠ Some checks failed")
	}

	return nil
}

type check func() bool

func checkHomeDirectory() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		printCheck("Home directory", false, err.Error())
		return false
	}

	clemDir := filepath.Join(home, ".clem")
	if _, err := os.Stat(clemDir); os.IsNotExist(err) {
		printCheck(".clem directory", false, "not found")
		fmt.Printf("  Run: mkdir -p %s\n", clemDir)
		return false
	}

	printCheck(".clem directory", true, clemDir)
	return true
}

func checkConfigFile() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		printCheck("Config file", false, "cannot get home dir")
		return false
	}

	configPath := filepath.Join(home, ".clem", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		printCheck("Config file", false, "not found")
		fmt.Println("  Run: clem setup-token <your-api-key>")
		return false
	}

	printCheck("Config file", true, configPath)
	return true
}

func checkAPIKey() bool {
	cfg, err := core.LoadConfig()
	if err != nil {
		printCheck("API key", false, err.Error())
		return false
	}

	if _, err := cfg.GetAPIKey(); err != nil {
		printCheck("API key", false, "not configured")
		fmt.Println("  Run: clem setup-token <your-api-key>")
		fmt.Println("  Or set: export JEFF_API_KEY=<your-key>")
		return false
	}

	printCheck("API key", true, "configured")
	return true
}

func printCheck(name string, passed bool, detail string) {
	symbol := "✓"
	if !passed {
		symbol = "✗"
	}
	fmt.Printf("%s %s: %s\n", symbol, name, detail)
}
```

**Step 2: Build and test**

```bash
make build

# Test without config
rm ~/.clem/config.yaml 2>/dev/null || true
./clem doctor
```

Expected: Shows failures for missing config

**Step 3: Setup config and test again**

```bash
./clem setup-token test-key-123
./clem doctor
```

Expected: All checks pass

**Step 4: Commit**

```bash
git add cmd/clem/doctor.go
git commit -m "feat: add doctor command for health checks"
```

---

## Task 9: Integration Test

**Files:**
- Create: `tests/integration/phase1_test.go`

**Step 1: Create integration test**

Create `tests/integration/phase1_test.go`:
```go
//go:build integration
// +build integration

package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase1Integration(t *testing.T) {
	// Build binary
	buildCmd := exec.Command("go", "build", "-o", "clem-test", "./cmd/clem")
	buildCmd.Dir = "../.."
	err := buildCmd.Run()
	require.NoError(t, err)
	defer os.Remove("../../clem-test")

	clemBin := "../../clem-test"

	t.Run("version flag", func(t *testing.T) {
		cmd := exec.Command(clemBin, "--version")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "0.1.0")
	})

	t.Run("help flag", func(t *testing.T) {
		cmd := exec.Command(clemBin, "--help")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "Clem")
		assert.Contains(t, string(output), "--print")
	})

	t.Run("setup-token command", func(t *testing.T) {
		tmpHome := t.TempDir()
		cmd := exec.Command(clemBin, "setup-token", "test-key-xyz")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "✓")

		// Verify file was created
		configPath := filepath.Join(tmpHome, ".clem", "config.yaml")
		assert.FileExists(t, configPath)
	})

	t.Run("doctor command", func(t *testing.T) {
		tmpHome := t.TempDir()

		// Setup first
		setupCmd := exec.Command(clemBin, "setup-token", "test-key")
		setupCmd.Env = append(os.Environ(), "HOME="+tmpHome)
		err := setupCmd.Run()
		require.NoError(t, err)

		// Run doctor
		doctorCmd := exec.Command(clemBin, "doctor")
		doctorCmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := doctorCmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "✓ All checks passed")
	})

	t.Run("print mode error without key", func(t *testing.T) {
		tmpHome := t.TempDir()
		cmd := exec.Command(clemBin, "--print", "test")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()
		assert.Error(t, err)
		assert.Contains(t, string(output), "API key not configured")
	})
}

func TestPrintModeWithRealAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real API test")
	}

	apiKey := os.Getenv("JEFF_API_KEY")
	if apiKey == "" {
		t.Skip("JEFF_API_KEY not set")
	}

	// Build binary
	buildCmd := exec.Command("go", "build", "-o", "clem-test", "./cmd/clem")
	buildCmd.Dir = "../.."
	err := buildCmd.Run()
	require.NoError(t, err)
	defer os.Remove("../../clem-test")

	clemBin := "../../clem-test"

	// Test print mode with real API
	cmd := exec.Command(clemBin, "--print", "Say hello in exactly 3 words")
	cmd.Env = append(os.Environ(), "JEFF_API_KEY="+apiKey)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	result := strings.TrimSpace(string(output))
	assert.NotEmpty(t, result)
	t.Logf("API Response: %s", result)
}
```

**Step 2: Run integration tests**

```bash
# Without API (fast tests only)
go test -tags=integration ./tests/integration/... -v

# With real API
export JEFF_API_KEY=your-key-here
go test -tags=integration ./tests/integration/... -v
```

Expected: All tests pass

**Step 3: Commit**

```bash
git add tests/integration/
git commit -m "test: add Phase 1 integration tests"
```

---

## Task 10: Documentation and Final Polish

**Files:**
- Update: `README.md`
- Create: `docs/PHASE1.md`

**Step 1: Update README**

Update `README.md`:
```markdown
# Clem

AI assistant CLI built in Go with full Claude Code feature parity.

## Quick Start

```bash
# Install
go install github.com/yourusername/clem/cmd/clem@latest

# Setup API key
clem setup-token sk-ant-api03-...

# Check health
clem doctor

# Use it!
clem --print "explain quantum computing in one sentence"
```

## Features

### Phase 1: Foundation ✅
- [x] CLI framework (Cobra)
- [x] Configuration system (Viper + .env)
- [x] Anthropic API client
- [x] Print mode (--print)
- [x] Setup and doctor commands

### Coming Soon
- [ ] Interactive mode (Bubbletea)
- [ ] Tool execution (Bash, Read, Write, Edit, Grep)
- [ ] MCP server integration
- [ ] Plugin system
- [ ] Conversation history
- [ ] Streaming responses

## Usage

### Print Mode (Non-Interactive)

```bash
# Basic usage
clem --print "your prompt"

# JSON output
clem --print --output-format json "your prompt"

# Different model
clem --print --model claude-opus-4-5-20250929 "your prompt"
```

### Configuration

Create `~/.clem/config.yaml`:
```yaml
api_key: sk-ant-api03-...
model: claude-sonnet-4-5-20250929
```

Or use environment variables:
```bash
export JEFF_API_KEY=sk-ant-api03-...
export JEFF_MODEL=claude-sonnet-4-5-20250929
```

Or use `.env` file in current directory.

## Development

```bash
# Build
make build

# Test
make test

# Run without building
make run -- --print "test"

# Integration tests
go test -tags=integration ./tests/integration/... -v
```

## Testing

We use real components and VCR cassettes, not mocks:

- **Unit tests:** Fast, isolated logic tests
- **Integration tests:** End-to-end workflows with real DB, real filesystem
- **VCR cassettes:** Record/replay real API calls

```bash
# Unit tests (fast)
go test -short ./...

# All tests including slow ones
go test ./...

# Integration tests
go test -tags=integration ./tests/integration/...

# With real API
JEFF_API_KEY=your-key go test ./...
```

## Architecture

```
cmd/clem/           # CLI entry point
internal/           # Private implementation
  core/             # Core types, API client, config
  ui/               # Terminal UI (Phase 2)
  tools/            # Tool implementations (Phase 3)
  mcp/              # MCP runtime (Phase 4)
  storage/          # SQLite persistence (Phase 2)
  plugins/          # Plugin loader (Phase 5)
pkg/                # Public APIs
  plugin/           # Plugin interface (Phase 5)
tests/              # Test suites
```

## License

MIT
```

**Step 2: Create Phase 1 documentation**

Create `docs/PHASE1.md`:
```markdown
# Phase 1: Foundation - Complete ✅

Phase 1 establishes the foundation for Clem CLI.

## What Was Built

### 1. Project Structure
- Go module with clean architecture
- Makefile for common tasks
- Comprehensive .gitignore

### 2. CLI Framework
- Cobra for command parsing
- Root command with flags
- Subcommands (setup-token, doctor)
- Help and version info

### 3. Configuration System
- Viper for multi-source config
- Priority: flags > env > .env > config file > defaults
- Support for ~/.clem/config.yaml
- Environment variable support (JEFF_*)

### 4. API Client
- HTTP client for Anthropic API
- Non-streaming message creation
- Proper error handling
- VCR testing for API calls

### 5. Print Mode
- --print flag for non-interactive use
- --output-format (text, json)
- --model flag for model selection
- Proper error messages

### 6. Commands
- `clem --print` - Send query, print response
- `clem setup-token` - Configure API key
- `clem doctor` - Health check

## Success Metrics

✅ All unit tests pass (20+ tests)
✅ Integration tests pass
✅ Can make real API calls
✅ Configuration works from multiple sources
✅ Help/version/doctor commands work
✅ Print mode functional

## Files Created

```
cmd/clem/
  main.go           # Entry point
  root.go           # Root command
  print.go          # Print mode handler
  setup.go          # Setup command
  doctor.go         # Doctor command

internal/core/
  types.go          # Core types
  types_test.go     # Type tests
  config.go         # Configuration
  config_test.go    # Config tests
  client.go         # API client
  client_test.go    # Client tests (with VCR)

tests/integration/
  phase1_test.go    # E2E tests

.gitignore
.env.example
Makefile
README.md
go.mod
go.sum
```

## Next: Phase 2

Phase 2 will add:
- Interactive mode with Bubbletea
- Streaming API support
- SQLite storage
- Conversation history (--continue, --resume)

See `docs/plans/2025-11-25-clem-phase2-interactive.md`
```

**Step 3: Final test of everything**

```bash
# Clean build
make clean
make build

# Run all tests
make test

# Manual smoke test
./clem --version
./clem --help
./clem setup-token test-key
./clem doctor
```

**Step 4: Commit**

```bash
git add README.md docs/PHASE1.md
git commit -m "docs: update README and add Phase 1 documentation"
```

**Step 5: Tag release**

```bash
git tag -a v0.1.0 -m "Phase 1: Foundation complete

Features:
- CLI framework with Cobra
- Configuration system with Viper
- Anthropic API client
- Print mode (--print)
- Setup and doctor commands

Full feature list in docs/PHASE1.md"
```

---

## Phase 1 Complete! 🎉

**Deliverables:**
- ✅ Working `clem --print "prompt"` command
- ✅ Configuration from multiple sources
- ✅ Setup and doctor commands
- ✅ Comprehensive test suite
- ✅ Clean, maintainable codebase
- ✅ Documentation

**What Works:**
```bash
clem --print "Say hello"                    # Basic usage
clem --print --output-format json "test"    # JSON output
clem --model claude-opus-4-5-20250929       # Model selection
clem setup-token sk-ant-...                 # Configure API
clem doctor                                 # Health check
```

**Next Steps:**
Ready to proceed to Phase 2 (Interactive Mode) when you are!
