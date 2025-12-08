# Multi-Provider Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add support for OpenAI, Gemini, and OpenRouter alongside existing Anthropic integration.

**Architecture:** Create Provider abstraction layer. Refactor existing Anthropic code into providers/anthropic/. Implement new providers with unified message translation. Update config system to TOML with per-provider API keys.

**Tech Stack:** Go, Viper (config), golang-migrate (database), net/http (API clients)

---

## Task 1: Provider Interface and Factory

**Files:**
- Create: `internal/providers/provider.go`
- Create: `internal/providers/factory.go`
- Create: `internal/providers/provider_test.go`

**Step 1: Write the provider interface test**

```go
// internal/providers/provider_test.go
package providers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/harper/hex/internal/providers"
)

func TestFactoryRegistersProviders(t *testing.T) {
	factory := providers.NewFactory()

	// Should have no providers initially
	_, err := factory.GetProvider("nonexistent")
	assert.Error(t, err)
}

func TestFactoryCreatesProvider(t *testing.T) {
	factory := providers.NewFactory()

	// Register a test provider
	factory.Register("test", &mockProvider{name: "test"})

	provider, err := factory.GetProvider("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", provider.Name())
}

type mockProvider struct {
	name string
}

func (m *mockProvider) CreateStream(ctx context.Context, req *providers.MessageRequest) (providers.Stream, error) {
	return nil, nil
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) ValidateConfig(cfg providers.ProviderConfig) error {
	return nil
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/providers -v
```
Expected: FAIL with "package providers not found"

**Step 3: Write provider interface**

```go
// internal/providers/provider.go
// ABOUTME: Provider interface defines the contract for all LLM provider implementations
// ABOUTME: Each provider translates between hex's universal format and provider-specific APIs
package providers

import (
	"context"
	"io"
)

// Provider defines the interface all LLM providers must implement
type Provider interface {
	// CreateStream sends a message request and returns a streaming response
	CreateStream(ctx context.Context, req *MessageRequest) (Stream, error)

	// Name returns the provider's identifier (e.g., "anthropic", "openai")
	Name() string

	// ValidateConfig checks if the provider configuration is valid
	ValidateConfig(cfg ProviderConfig) error
}

// Stream represents a streaming response from a provider
type Stream interface {
	// Next returns the next chunk in the stream
	Next() (*StreamChunk, error)

	// Close closes the stream and releases resources
	Close() error
}

// MessageRequest represents a universal message request format
type MessageRequest struct {
	Model       string
	Messages    []Message
	Tools       []Tool
	MaxTokens   int
	Stream      bool
	Temperature float64
}

// Message represents a single message in the conversation
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// Tool represents a tool that can be called by the model
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Type    string // "message_start", "content_block_delta", "message_stop"
	Content string
	Done    bool
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	APIKey  string
	BaseURL string
}
```

**Step 4: Write factory implementation**

```go
// internal/providers/factory.go
// ABOUTME: Factory creates and manages provider instances
// ABOUTME: Provides registry for available providers and instantiation logic
package providers

import (
	"fmt"
	"sync"
)

// Factory creates provider instances
type Factory struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the factory
func (f *Factory) Register(name string, provider Provider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[name] = provider
}

// GetProvider returns a provider by name
func (f *Factory) GetProvider(name string) (Provider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, ok := f.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// ListProviders returns all registered provider names
func (f *Factory) ListProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}
```

**Step 5: Run tests to verify they pass**

```bash
go test ./internal/providers -v
```
Expected: PASS

**Step 6: Commit**

```bash
git add internal/providers/
git commit -m "feat: add provider interface and factory

Create abstraction layer for multi-provider support.
Defines Provider interface with CreateStream, Name, ValidateConfig.
Implements Factory for provider registration and retrieval.

Part of multi-provider support implementation."
```

---

## Task 2: Anthropic Provider Refactor

**Files:**
- Create: `internal/providers/anthropic/client.go`
- Create: `internal/providers/anthropic/stream.go`
- Create: `internal/providers/anthropic/types.go`
- Create: `internal/providers/anthropic/client_test.go`
- Modify: `internal/core/client.go` (will be refactored)

**Step 1: Write test for Anthropic provider**

```go
// internal/providers/anthropic/client_test.go
package anthropic_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/harper/hex/internal/providers"
	"github.com/harper/hex/internal/providers/anthropic"
)

func TestAnthropicProviderName(t *testing.T) {
	provider := anthropic.NewProvider(providers.ProviderConfig{
		APIKey: "test-key",
	})

	assert.Equal(t, "anthropic", provider.Name())
}

func TestAnthropicValidateConfig(t *testing.T) {
	provider := anthropic.NewProvider(providers.ProviderConfig{})

	tests := []struct{
		name    string
		cfg     providers.ProviderConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: providers.ProviderConfig{
				APIKey: "sk-ant-test",
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			cfg: providers.ProviderConfig{
				APIKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateConfig(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/providers/anthropic -v
```
Expected: FAIL with "package anthropic not found"

**Step 3: Copy and adapt existing Anthropic client code**

```go
// internal/providers/anthropic/client.go
// ABOUTME: Anthropic API client implementation of Provider interface
// ABOUTME: Handles Anthropic-specific message format and streaming
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/harper/hex/internal/providers"
)

const (
	defaultBaseURL = "https://api.anthropic.com/v1/messages"
	apiVersion     = "2023-06-01"
)

// Provider implements the Provider interface for Anthropic
type Provider struct {
	config     providers.ProviderConfig
	httpClient *http.Client
}

// NewProvider creates a new Anthropic provider
func NewProvider(config providers.ProviderConfig) *Provider {
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{
		config:     config,
		httpClient: &http.Client{},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "anthropic"
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig(cfg providers.ProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("anthropic: API key is required")
	}
	return nil
}

// CreateStream creates a streaming request to Anthropic
func (p *Provider) CreateStream(ctx context.Context, req *providers.MessageRequest) (providers.Stream, error) {
	// Translate to Anthropic format
	anthropicReq := translateRequest(req)

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return newStream(resp.Body), nil
}
```

**Step 4: Write types translation**

```go
// internal/providers/anthropic/types.go
// ABOUTME: Type translation between hex universal format and Anthropic API format
// ABOUTME: Handles message structure, tool definitions, and response parsing
package anthropic

import (
	"github.com/harper/hex/internal/providers"
)

// anthropicRequest is the Anthropic-specific request format
type anthropicRequest struct {
	Model       string              `json:"model"`
	Messages    []anthropicMessage  `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Stream      bool                `json:"stream"`
	Temperature float64             `json:"temperature,omitempty"`
	Tools       []anthropicTool     `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// translateRequest converts universal format to Anthropic format
func translateRequest(req *providers.MessageRequest) *anthropicRequest {
	messages := make([]anthropicMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = anthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	tools := make([]anthropicTool, len(req.Tools))
	for i, tool := range req.Tools {
		tools[i] = anthropicTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		}
	}

	return &anthropicRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		Tools:       tools,
	}
}
```

**Step 5: Write streaming implementation**

```go
// internal/providers/anthropic/stream.go
// ABOUTME: SSE stream parsing for Anthropic API responses
// ABOUTME: Handles message_start, content_block_delta, message_stop events
package anthropic

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/harper/hex/internal/providers"
)

type stream struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
}

func newStream(reader io.ReadCloser) *stream {
	return &stream{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
	}
}

func (s *stream) Next() (*providers.StreamChunk, error) {
	for s.scanner.Scan() {
		line := s.scanner.Text()

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return &providers.StreamChunk{
				Type: "message_stop",
				Done: true,
			}, nil
		}

		var event anthropicEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil, fmt.Errorf("unmarshal event: %w", err)
		}

		chunk := translateEvent(&event)
		if chunk != nil {
			return chunk, nil
		}
	}

	if err := s.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

func (s *stream) Close() error {
	return s.reader.Close()
}

type anthropicEvent struct {
	Type  string          `json:"type"`
	Delta *anthropicDelta `json:"delta"`
}

type anthropicDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func translateEvent(event *anthropicEvent) *providers.StreamChunk {
	switch event.Type {
	case "message_start":
		return &providers.StreamChunk{
			Type: "message_start",
		}
	case "content_block_delta":
		if event.Delta != nil {
			return &providers.StreamChunk{
				Type:    "content_block_delta",
				Content: event.Delta.Text,
			}
		}
	case "message_stop":
		return &providers.StreamChunk{
			Type: "message_stop",
			Done: true,
		}
	}
	return nil
}
```

**Step 6: Run tests**

```bash
go test ./internal/providers/anthropic -v
```
Expected: PASS

**Step 7: Commit**

```bash
git add internal/providers/anthropic/
git commit -m "feat: implement Anthropic provider

Refactor existing Anthropic client into provider pattern.
Implements Provider interface with CreateStream.
Translates between universal and Anthropic-specific formats.
Handles SSE streaming with message events.

Part of multi-provider support implementation."
```

---

## Task 3: OpenAI Provider Implementation

**Files:**
- Create: `internal/providers/openai/client.go`
- Create: `internal/providers/openai/stream.go`
- Create: `internal/providers/openai/types.go`
- Create: `internal/providers/openai/client_test.go`

**Step 1: Write tests for OpenAI provider**

```go
// internal/providers/openai/client_test.go
package openai_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/harper/hex/internal/providers"
	"github.com/harper/hex/internal/providers/openai"
)

func TestOpenAIProviderName(t *testing.T) {
	provider := openai.NewProvider(providers.ProviderConfig{
		APIKey: "test-key",
	})

	assert.Equal(t, "openai", provider.Name())
}

func TestOpenAIValidateConfig(t *testing.T) {
	provider := openai.NewProvider(providers.ProviderConfig{})

	tests := []struct{
		name    string
		cfg     providers.ProviderConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: providers.ProviderConfig{
				APIKey: "sk-proj-test",
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			cfg: providers.ProviderConfig{
				APIKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateConfig(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOpenAIMessageTranslation(t *testing.T) {
	req := &providers.MessageRequest{
		Model: "gpt-4o",
		Messages: []providers.Message{
			{Role: "user", Content: "hello"},
		},
		MaxTokens: 1000,
	}

	openaiReq := openai.TranslateRequest(req)

	assert.Equal(t, "gpt-4o", openaiReq.Model)
	assert.Len(t, openaiReq.Messages, 1)
	assert.Equal(t, "user", openaiReq.Messages[0].Role)
	assert.Equal(t, "hello", openaiReq.Messages[0].Content)
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/providers/openai -v
```
Expected: FAIL

**Step 3: Implement OpenAI client**

```go
// internal/providers/openai/client.go
// ABOUTME: OpenAI API client implementation of Provider interface
// ABOUTME: Handles Chat Completions API with streaming support
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/harper/hex/internal/providers"
)

const (
	defaultBaseURL = "https://api.openai.com/v1/chat/completions"
)

// Provider implements the Provider interface for OpenAI
type Provider struct {
	config     providers.ProviderConfig
	httpClient *http.Client
}

// NewProvider creates a new OpenAI provider
func NewProvider(config providers.ProviderConfig) *Provider {
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{
		config:     config,
		httpClient: &http.Client{},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "openai"
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig(cfg providers.ProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("openai: API key is required")
	}
	return nil
}

// CreateStream creates a streaming request to OpenAI
func (p *Provider) CreateStream(ctx context.Context, req *providers.MessageRequest) (providers.Stream, error) {
	// Translate to OpenAI format
	openaiReq := TranslateRequest(req)

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return newStream(resp.Body), nil
}
```

**Step 4: Implement types translation**

```go
// internal/providers/openai/types.go
// ABOUTME: Type translation between hex universal format and OpenAI Chat Completions format
// ABOUTME: Handles message structure, function calling, and response parsing
package openai

import (
	"github.com/harper/hex/internal/providers"
)

// OpenAIRequest is the OpenAI-specific request format
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream"`
	Temperature float64         `json:"temperature,omitempty"`
	Tools       []OpenAITool    `json:"tools,omitempty"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAITool struct {
	Type     string                 `json:"type"` // "function"
	Function OpenAIFunctionDef      `json:"function"`
}

type OpenAIFunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// TranslateRequest converts universal format to OpenAI format
func TranslateRequest(req *providers.MessageRequest) *OpenAIRequest {
	messages := make([]OpenAIMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	tools := make([]OpenAITool, len(req.Tools))
	for i, tool := range req.Tools {
		tools[i] = OpenAITool{
			Type: "function",
			Function: OpenAIFunctionDef{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		}
	}

	return &OpenAIRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		Tools:       tools,
	}
}
```

**Step 5: Implement streaming**

```go
// internal/providers/openai/stream.go
// ABOUTME: SSE stream parsing for OpenAI Chat Completions streaming
// ABOUTME: Handles delta chunks with choices array format
package openai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/harper/hex/internal/providers"
)

type stream struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
}

func newStream(reader io.ReadCloser) *stream {
	return &stream{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
	}
}

func (s *stream) Next() (*providers.StreamChunk, error) {
	for s.scanner.Scan() {
		line := s.scanner.Text()

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return &providers.StreamChunk{
				Type: "message_stop",
				Done: true,
			}, nil
		}

		var event openAIStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil, fmt.Errorf("unmarshal event: %w", err)
		}

		chunk := translateEvent(&event)
		if chunk != nil {
			return chunk, nil
		}
	}

	if err := s.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

func (s *stream) Close() error {
	return s.reader.Close()
}

type openAIStreamEvent struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Delta        openAIDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type openAIDelta struct {
	Content string `json:"content"`
}

func translateEvent(event *openAIStreamEvent) *providers.StreamChunk {
	if len(event.Choices) == 0 {
		return nil
	}

	choice := event.Choices[0]

	if choice.FinishReason != nil {
		return &providers.StreamChunk{
			Type: "message_stop",
			Done: true,
		}
	}

	if choice.Delta.Content != "" {
		return &providers.StreamChunk{
			Type:    "content_block_delta",
			Content: choice.Delta.Content,
		}
	}

	return nil
}
```

**Step 6: Run tests**

```bash
go test ./internal/providers/openai -v
```
Expected: PASS

**Step 7: Commit**

```bash
git add internal/providers/openai/
git commit -m "feat: implement OpenAI provider

Add OpenAI Chat Completions API support.
Implements Provider interface with streaming.
Translates between universal and OpenAI formats.
Handles function calling as tools.

Part of multi-provider support implementation."
```

---

## Task 4: Gemini Provider Implementation

**Files:**
- Create: `internal/providers/gemini/client.go`
- Create: `internal/providers/gemini/stream.go`
- Create: `internal/providers/gemini/types.go`
- Create: `internal/providers/gemini/client_test.go`

**Step 1: Write tests for Gemini provider**

```go
// internal/providers/gemini/client_test.go
package gemini_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/harper/hex/internal/providers"
	"github.com/harper/hex/internal/providers/gemini"
)

func TestGeminiProviderName(t *testing.T) {
	provider := gemini.NewProvider(providers.ProviderConfig{
		APIKey: "test-key",
	})

	assert.Equal(t, "gemini", provider.Name())
}

func TestGeminiValidateConfig(t *testing.T) {
	provider := gemini.NewProvider(providers.ProviderConfig{})

	err := provider.ValidateConfig(providers.ProviderConfig{
		APIKey: "",
	})
	assert.Error(t, err)

	err = provider.ValidateConfig(providers.ProviderConfig{
		APIKey: "AIza-test",
	})
	assert.NoError(t, err)
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/providers/gemini -v
```
Expected: FAIL

**Step 3: Implement Gemini client**

```go
// internal/providers/gemini/client.go
// ABOUTME: Google Gemini API client implementation of Provider interface
// ABOUTME: Handles streamGenerateContent API with model ID in URL path
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/harper/hex/internal/providers"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1"
)

// Provider implements the Provider interface for Gemini
type Provider struct {
	config     providers.ProviderConfig
	httpClient *http.Client
}

// NewProvider creates a new Gemini provider
func NewProvider(config providers.ProviderConfig) *Provider {
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{
		config:     config,
		httpClient: &http.Client{},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "gemini"
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig(cfg providers.ProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("gemini: API key is required")
	}
	return nil
}

// CreateStream creates a streaming request to Gemini
func (p *Provider) CreateStream(ctx context.Context, req *providers.MessageRequest) (providers.Stream, error) {
	// Translate to Gemini format
	geminiReq := TranslateRequest(req)

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Model ID goes in URL path
	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s",
		p.config.BaseURL, req.Model, p.config.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return newStream(resp.Body), nil
}
```

**Step 4: Implement types translation**

```go
// internal/providers/gemini/types.go
// ABOUTME: Type translation between hex universal format and Gemini API format
// ABOUTME: Handles contents/parts structure and function declarations
package gemini

import (
	"github.com/harper/hex/internal/providers"
)

// GeminiRequest is the Gemini-specific request format
type GeminiRequest struct {
	Contents         []GeminiContent  `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	Tools            []GeminiTool     `json:"tools,omitempty"`
}

type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

type GeminiTool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations"`
}

type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// TranslateRequest converts universal format to Gemini format
func TranslateRequest(req *providers.MessageRequest) *GeminiRequest {
	contents := make([]GeminiContent, len(req.Messages))
	for i, msg := range req.Messages {
		contents[i] = GeminiContent{
			Role: msg.Role,
			Parts: []GeminiPart{
				{Text: msg.Content},
			},
		}
	}

	var tools []GeminiTool
	if len(req.Tools) > 0 {
		functionDecls := make([]FunctionDeclaration, len(req.Tools))
		for i, tool := range req.Tools {
			functionDecls[i] = FunctionDeclaration{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			}
		}
		tools = []GeminiTool{{FunctionDeclarations: functionDecls}}
	}

	return &GeminiRequest{
		Contents: contents,
		GenerationConfig: &GenerationConfig{
			MaxOutputTokens: req.MaxTokens,
			Temperature:     req.Temperature,
		},
		Tools: tools,
	}
}
```

**Step 5: Implement streaming**

```go
// internal/providers/gemini/stream.go
// ABOUTME: JSON stream parsing for Gemini streaming responses
// ABOUTME: Handles candidates array with content parts
package gemini

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/harper/hex/internal/providers"
)

type stream struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
	done    bool
}

func newStream(reader io.ReadCloser) *stream {
	return &stream{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
	}
}

func (s *stream) Next() (*providers.StreamChunk, error) {
	if s.done {
		return nil, io.EOF
	}

	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			continue
		}

		var event geminiStreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // Skip malformed lines
		}

		chunk := translateEvent(&event)
		if chunk != nil {
			if chunk.Done {
				s.done = true
			}
			return chunk, nil
		}
	}

	if err := s.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

func (s *stream) Close() error {
	return s.reader.Close()
}

type geminiStreamEvent struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

func translateEvent(event *geminiStreamEvent) *providers.StreamChunk {
	if len(event.Candidates) == 0 {
		return nil
	}

	candidate := event.Candidates[0]

	if candidate.FinishReason != "" {
		return &providers.StreamChunk{
			Type: "message_stop",
			Done: true,
		}
	}

	if len(candidate.Content.Parts) > 0 {
		text := candidate.Content.Parts[0].Text
		if text != "" {
			return &providers.StreamChunk{
				Type:    "content_block_delta",
				Content: text,
			}
		}
	}

	return nil
}
```

**Step 6: Run tests**

```bash
go test ./internal/providers/gemini -v
```
Expected: PASS

**Step 7: Commit**

```bash
git add internal/providers/gemini/
git commit -m "feat: implement Gemini provider

Add Google Gemini API support.
Implements Provider interface with streaming.
Handles model ID in URL path, API key as query param.
Translates to contents/parts structure.

Part of multi-provider support implementation."
```

---

## Task 5: OpenRouter Provider Implementation

**Files:**
- Create: `internal/providers/openrouter/client.go`
- Create: `internal/providers/openrouter/types.go`
- Create: `internal/providers/openrouter/client_test.go`

**Step 1: Write tests**

```go
// internal/providers/openrouter/client_test.go
package openrouter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/harper/hex/internal/providers"
	"github.com/harper/hex/internal/providers/openrouter"
)

func TestOpenRouterProviderName(t *testing.T) {
	provider := openrouter.NewProvider(providers.ProviderConfig{
		APIKey: "test-key",
	})

	assert.Equal(t, "openrouter", provider.Name())
}

func TestOpenRouterModelIDHandling(t *testing.T) {
	// OpenRouter uses provider/model format
	req := &providers.MessageRequest{
		Model: "anthropic/claude-sonnet-4-5",
	}

	openaiReq := openrouter.TranslateRequest(req)
	assert.Equal(t, "anthropic/claude-sonnet-4-5", openaiReq.Model)
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/providers/openrouter -v
```
Expected: FAIL

**Step 3: Implement OpenRouter client (wraps OpenAI)**

```go
// internal/providers/openrouter/client.go
// ABOUTME: OpenRouter provider wraps OpenAI format with different endpoint
// ABOUTME: OpenRouter is OpenAI-compatible proxy supporting multiple providers
package openrouter

import (
	"context"
	"fmt"

	"github.com/harper/hex/internal/providers"
	"github.com/harper/hex/internal/providers/openai"
)

const (
	defaultBaseURL = "https://openrouter.ai/api/v1/chat/completions"
)

// Provider implements the Provider interface for OpenRouter
type Provider struct {
	openaiProvider *openai.Provider
}

// NewProvider creates a new OpenRouter provider
func NewProvider(config providers.ProviderConfig) *Provider {
	// OpenRouter uses same format as OpenAI, just different endpoint
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{
		openaiProvider: openai.NewProvider(config),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "openrouter"
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig(cfg providers.ProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("openrouter: API key is required")
	}
	return nil
}

// CreateStream delegates to OpenAI provider (OpenRouter is compatible)
func (p *Provider) CreateStream(ctx context.Context, req *providers.MessageRequest) (providers.Stream, error) {
	// Model IDs in OpenRouter use provider/model format (e.g., "anthropic/claude-sonnet-4-5")
	// No translation needed, OpenRouter handles routing
	return p.openaiProvider.CreateStream(ctx, req)
}
```

**Step 4: Implement types (reuse OpenAI)**

```go
// internal/providers/openrouter/types.go
// ABOUTME: OpenRouter uses OpenAI-compatible format
// ABOUTME: Only difference is model ID format (provider/model)
package openrouter

import (
	"github.com/harper/hex/internal/providers"
	"github.com/harper/hex/internal/providers/openai"
)

// TranslateRequest reuses OpenAI translation
// OpenRouter is OpenAI-compatible
func TranslateRequest(req *providers.MessageRequest) *openai.OpenAIRequest {
	return openai.TranslateRequest(req)
}
```

**Step 5: Run tests**

```bash
go test ./internal/providers/openrouter -v
```
Expected: PASS

**Step 6: Commit**

```bash
git add internal/providers/openrouter/
git commit -m "feat: implement OpenRouter provider

Add OpenRouter support (wraps OpenAI format).
OpenRouter is OpenAI-compatible proxy to multiple providers.
Handles provider/model ID format routing.

Part of multi-provider support implementation."
```

---

## Task 6: Config System Update (YAML to TOML)

**Files:**
- Modify: `internal/core/config.go`
- Create: `internal/core/config_test.go` (expand)
- Create: `internal/core/migration.go`

**Step 1: Write tests for new config structure**

```go
// internal/core/config_test.go (add to existing)
func TestLoadConfigTOML(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
provider = "openai"
model = "gpt-4o"

[providers.anthropic]
api_key = "sk-ant-test"

[providers.openai]
api_key = "sk-proj-test"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	os.Setenv("HEX_CONFIG_PATH", configPath)
	defer os.Unsetenv("HEX_CONFIG_PATH")

	cfg, err := core.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "openai", cfg.Provider)
	assert.Equal(t, "gpt-4o", cfg.Model)
	assert.Equal(t, "sk-ant-test", cfg.ProviderConfigs["anthropic"].APIKey)
	assert.Equal(t, "sk-proj-test", cfg.ProviderConfigs["openai"].APIKey)
}

func TestConfigProviderValidation(t *testing.T) {
	cfg := &core.Config{
		Provider: "openai",
		ProviderConfigs: map[string]core.ProviderConfig{
			"anthropic": {APIKey: "sk-ant-test"},
		},
	}

	err := cfg.ValidateProvider()
	assert.Error(t, err) // openai selected but not configured

	cfg.ProviderConfigs["openai"] = core.ProviderConfig{APIKey: "sk-proj-test"}
	err = cfg.ValidateProvider()
	assert.NoError(t, err)
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/core -v
```
Expected: FAIL

**Step 3: Update config.go structure**

```go
// internal/core/config.go (update)
package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

// Config holds application configuration
type Config struct {
	Provider        string                    `mapstructure:"provider"`
	Model           string                    `mapstructure:"model"`
	ProviderConfigs map[string]ProviderConfig `mapstructure:"providers"`
	DefaultTools    []string                  `mapstructure:"default_tools"`
	PermissionMode  string                    `mapstructure:"permission_mode"`
}

// LoadConfig loads configuration from multiple sources
// Priority (highest to lowest):
// 1. Environment variables (HEX_*)
// 2. Provider-specific env vars (ANTHROPIC_API_KEY, OPENAI_API_KEY, etc.)
// 3. .env file (current directory)
// 4. ~/.hex/config.toml (migrated from config.yaml if needed)
// 5. Defaults
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	v := viper.New()

	// Set defaults
	v.SetDefault("provider", "anthropic")
	v.SetDefault("model", DefaultModel)
	v.SetDefault("permission_mode", "ask")
	v.SetDefault("default_tools", []string{"Bash", "Read", "Write", "Edit", "Grep"})

	// Environment variables
	v.SetEnvPrefix("HEX")
	v.AutomaticEnv()
	_ = v.BindEnv("provider")
	_ = v.BindEnv("model")
	_ = v.BindEnv("permission_mode")
	_ = v.BindEnv("default_tools")

	// Check for config file
	configPath := os.Getenv("HEX_CONFIG_PATH")
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			hexDir := filepath.Join(home, ".hex")

			// Check if config.yaml exists and migrate to config.toml
			yamlPath := filepath.Join(hexDir, "config.yaml")
			tomlPath := filepath.Join(hexDir, "config.toml")
			if _, err := os.Stat(yamlPath); err == nil {
				if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
					if err := migrateYAMLToTOML(yamlPath, tomlPath); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to migrate config: %v\n", err)
					}
				}
			}

			v.SetConfigName("config")
			v.SetConfigType("toml")
			v.AddConfigPath(hexDir)
			v.AddConfigPath(".")
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

	// Initialize provider configs if nil
	if cfg.ProviderConfigs == nil {
		cfg.ProviderConfigs = make(map[string]ProviderConfig)
	}

	// Check standard provider env vars
	providers := map[string]string{
		"anthropic":  os.Getenv("ANTHROPIC_API_KEY"),
		"openai":     os.Getenv("OPENAI_API_KEY"),
		"gemini":     os.Getenv("GEMINI_API_KEY"),
		"openrouter": os.Getenv("OPENROUTER_API_KEY"),
	}

	for name, envKey := range providers {
		if envKey != "" {
			if pc, ok := cfg.ProviderConfigs[name]; ok && pc.APIKey == "" {
				pc.APIKey = envKey
				cfg.ProviderConfigs[name] = pc
			} else if !ok {
				cfg.ProviderConfigs[name] = ProviderConfig{APIKey: envKey}
			}
		}
	}

	return &cfg, nil
}

// ValidateProvider ensures the selected provider is configured
func (c *Config) ValidateProvider() error {
	pc, ok := c.ProviderConfigs[c.Provider]
	if !ok {
		return fmt.Errorf("provider %s not configured", c.Provider)
	}
	if pc.APIKey == "" {
		return fmt.Errorf("API key not set for provider %s", c.Provider)
	}
	return nil
}

// GetProviderConfig returns the config for the selected provider
func (c *Config) GetProviderConfig() (ProviderConfig, error) {
	pc, ok := c.ProviderConfigs[c.Provider]
	if !ok {
		return ProviderConfig{}, fmt.Errorf("provider %s not configured", c.Provider)
	}
	return pc, nil
}
```

**Step 4: Implement YAML to TOML migration**

```go
// internal/core/migration.go
// ABOUTME: Migrates existing config.yaml to config.toml format
// ABOUTME: Preserves all existing settings and adds provider structure
package core

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// migrateYAMLToTOML converts config.yaml to config.toml
func migrateYAMLToTOML(yamlPath, tomlPath string) error {
	// Read YAML
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("read yaml: %w", err)
	}

	var yamlCfg map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &yamlCfg); err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}

	// Convert to new structure
	tomlCfg := make(map[string]interface{})

	// Copy simple fields
	if v, ok := yamlCfg["model"]; ok {
		tomlCfg["model"] = v
	}
	if v, ok := yamlCfg["permission_mode"]; ok {
		tomlCfg["permission_mode"] = v
	}
	if v, ok := yamlCfg["default_tools"]; ok {
		tomlCfg["default_tools"] = v
	}

	// Set default provider
	tomlCfg["provider"] = "anthropic"

	// Migrate API key to provider structure
	providers := make(map[string]map[string]string)
	if apiKey, ok := yamlCfg["api_key"].(string); ok && apiKey != "" {
		providers["anthropic"] = map[string]string{"api_key": apiKey}
	}
	tomlCfg["providers"] = providers

	// Write TOML
	tomlData, err := toml.Marshal(tomlCfg)
	if err != nil {
		return fmt.Errorf("marshal toml: %w", err)
	}

	if err := os.WriteFile(tomlPath, tomlData, 0644); err != nil {
		return fmt.Errorf("write toml: %w", err)
	}

	fmt.Printf("Migrated %s to %s\n", yamlPath, tomlPath)
	return nil
}
```

**Step 5: Update go.mod**

```bash
go get github.com/pelletier/go-toml/v2
go mod tidy
```

**Step 6: Run tests**

```bash
go test ./internal/core -v
```
Expected: PASS

**Step 7: Commit**

```bash
git add internal/core/config.go internal/core/migration.go internal/core/config_test.go go.mod go.sum
git commit -m "feat: migrate config from YAML to TOML with provider support

Switch from config.yaml to config.toml.
Add Provider field and per-provider API key structure.
Auto-migrate existing YAML configs to TOML.
Support standard provider env vars (ANTHROPIC_API_KEY, etc).

Part of multi-provider support implementation."
```

---

## Task 7: Database Migration for Provider Column

**Files:**
- Create: `internal/storage/migrations/000003_add_provider_column.up.sql`
- Create: `internal/storage/migrations/000003_add_provider_column.down.sql`
- Modify: `internal/storage/store.go`

**Step 1: Write migration SQL**

```sql
-- internal/storage/migrations/000003_add_provider_column.up.sql
ALTER TABLE conversations ADD COLUMN provider TEXT DEFAULT 'anthropic';
```

```sql
-- internal/storage/migrations/000003_add_provider_column.down.sql
ALTER TABLE conversations DROP COLUMN provider;
```

**Step 2: Update store.go to include provider**

```go
// internal/storage/store.go (add provider field to Conversation struct)
type Conversation struct {
	ID        string
	Provider  string  // ADD THIS
	Model     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SaveConversation (update to include provider)
func (s *Store) SaveConversation(conv *Conversation) error {
	_, err := s.db.Exec(`
		INSERT INTO conversations (id, provider, model, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			provider = excluded.provider,
			model = excluded.model,
			updated_at = excluded.updated_at
	`, conv.ID, conv.Provider, conv.Model, conv.CreatedAt, conv.UpdatedAt)
	return err
}

// GetConversation (update to retrieve provider)
func (s *Store) GetConversation(id string) (*Conversation, error) {
	var conv Conversation
	err := s.db.QueryRow(`
		SELECT id, provider, model, created_at, updated_at
		FROM conversations
		WHERE id = ?
	`, id).Scan(&conv.ID, &conv.Provider, &conv.Model, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}
```

**Step 3: Test migration**

```bash
# Run migrations
go run cmd/hex/main.go # Should auto-migrate

# Verify schema
sqlite3 ~/.hex/hex.db "PRAGMA table_info(conversations);"
# Should show provider column
```

**Step 4: Commit**

```bash
git add internal/storage/migrations/ internal/storage/store.go
git commit -m "feat: add provider column to conversations table

Add provider field to track which provider was used.
Defaults to 'anthropic' for backward compatibility.
Enables resuming conversations with original provider.

Part of multi-provider support implementation."
```

---

## Task 8: Update Orchestrator to Use Provider Interface

**Files:**
- Modify: `internal/orchestrator/stream_handler.go`
- Modify: `internal/orchestrator/orchestrator.go`
- Modify: `internal/services/agent_impl.go`

**Step 1: Update orchestrator to accept Provider**

```go
// internal/orchestrator/orchestrator.go (update)
type Orchestrator struct {
	provider     providers.Provider  // Change from core.Client
	model        string
	// ... other fields
}

func NewOrchestrator(provider providers.Provider, model string, ...) *Orchestrator {
	return &Orchestrator{
		provider: provider,
		model:    model,
		// ...
	}
}
```

**Step 2: Update stream_handler.go**

```go
// internal/orchestrator/stream_handler.go (update line 24)
func (o *Orchestrator) handleStreaming() error {
	req := &providers.MessageRequest{  // Use providers.MessageRequest
		Model:     o.model,  // Use injected model, not hardcoded
		Messages:  o.getMessageHistorySafe(),
		MaxTokens: 4096,
		Stream:    true,
	}

	stream, err := o.provider.CreateStream(o.ctx, req)
	if err != nil {
		return fmt.Errorf("create stream: %w", err)
	}
	defer stream.Close()

	// ... rest of streaming logic
}
```

**Step 3: Update agent service**

```go
// internal/services/agent_impl.go (update)
type AgentImpl struct {
	provider providers.Provider  // Rename from LLMClient
	// ...
}

func NewAgent(provider providers.Provider, ...) *AgentImpl {
	return &AgentImpl{
		provider: provider,
		// ...
	}
}
```

**Step 4: Run tests**

```bash
go test ./internal/orchestrator -v
go test ./internal/services -v
```

**Step 5: Commit**

```bash
git add internal/orchestrator/ internal/services/
git commit -m "refactor: update orchestrator to use Provider interface

Replace direct Anthropic client with Provider interface.
Remove hardcoded model from stream handler.
Inject provider at orchestrator initialization.

Part of multi-provider support implementation."
```

---

## Task 9: Update CLI Commands for Provider Selection

**Files:**
- Modify: `cmd/hex/root.go`
- Modify: `cmd/hex/print.go`

**Step 1: Add provider flag to root command**

```go
// cmd/hex/root.go (add flag)
var (
	providerFlag string
	modelFlag    string
	// ... other flags
)

func init() {
	rootCmd.PersistentFlags().StringVar(&providerFlag, "provider", "", "LLM provider (anthropic|openai|gemini|openrouter)")
	rootCmd.PersistentFlags().StringVar(&modelFlag, "model", "", "Model to use")
	// ... other flags
}
```

**Step 2: Update provider instantiation in root.go**

```go
// cmd/hex/root.go (in run function)
func run() error {
	// Load config
	cfg, err := core.LoadConfig()
	if err != nil {
		return err
	}

	// Override from flags
	if providerFlag != "" {
		cfg.Provider = providerFlag
	}
	if modelFlag != "" {
		cfg.Model = modelFlag
	}

	// Validate provider is configured
	if err := cfg.ValidateProvider(); err != nil {
		return fmt.Errorf("provider configuration error: %w", err)
	}

	// Get provider config
	providerCfg, err := cfg.GetProviderConfig()
	if err != nil {
		return err
	}

	// Convert to providers.ProviderConfig
	pCfg := providers.ProviderConfig{
		APIKey:  providerCfg.APIKey,
		BaseURL: providerCfg.BaseURL,
	}

	// Create provider factory
	factory := providers.NewFactory()
	factory.Register("anthropic", anthropic.NewProvider(pCfg))
	factory.Register("openai", openai.NewProvider(pCfg))
	factory.Register("gemini", gemini.NewProvider(pCfg))
	factory.Register("openrouter", openrouter.NewProvider(pCfg))

	// Get selected provider
	provider, err := factory.GetProvider(cfg.Provider)
	if err != nil {
		return fmt.Errorf("get provider: %w", err)
	}

	// Create orchestrator with provider
	orch := orchestrator.NewOrchestrator(provider, cfg.Model, ...)

	// ... rest of command logic
}
```

**Step 3: Update print.go similarly**

```go
// cmd/hex/print.go (similar updates)
func runPrint() error {
	// Same provider selection logic as root.go
	// ...
}
```

**Step 4: Test CLI**

```bash
# Build
make build

# Test with different providers
./hex --provider anthropic --model claude-sonnet-4-5-20250929 "test"
./hex --provider openai --model gpt-4o "test"
./hex --provider gemini --model gemini-2.5-flash "test"
./hex --provider openrouter --model anthropic/claude-sonnet-4-5 "test"
```

**Step 5: Commit**

```bash
git add cmd/hex/root.go cmd/hex/print.go
git commit -m "feat: add --provider flag and multi-provider CLI support

Add --provider flag to select provider.
Implement provider factory and instantiation in CLI.
Pass provider to orchestrator instead of direct client.

Part of multi-provider support implementation."
```

---

## Task 10: Update Cost Tracking for Multiple Providers

**Files:**
- Modify: `internal/cost/pricing.go`
- Modify: `internal/cost/tracker.go`

**Step 1: Add pricing for OpenAI and Gemini models**

```go
// internal/cost/pricing.go (expand modelPricing map)
var modelPricing = map[string]PricingModel{
	// Anthropic models
	"claude-sonnet-4-5-20250929": {
		InputTokenPrice:  3.00,
		OutputTokenPrice: 15.00,
		CacheReadPrice:   0.30,
		CacheWritePrice:  3.75,
	},
	"claude-opus-4-5-20251101": {
		InputTokenPrice:  5.00,
		OutputTokenPrice: 25.00,
		CacheReadPrice:   0.50,
		CacheWritePrice:  6.25,
	},
	"claude-haiku-4-5-20251001": {
		InputTokenPrice:  1.00,
		OutputTokenPrice: 5.00,
		CacheReadPrice:   0.10,
		CacheWritePrice:  1.25,
	},

	// OpenAI models
	"gpt-4o": {
		InputTokenPrice:  2.50,
		OutputTokenPrice: 10.00,
		CacheReadPrice:   0.00,  // OpenAI doesn't have prompt caching yet
		CacheWritePrice:  0.00,
	},
	"gpt-4o-mini": {
		InputTokenPrice:  0.15,
		OutputTokenPrice: 0.60,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"o3": {
		InputTokenPrice:  15.00,  // Reasoning models more expensive
		OutputTokenPrice: 60.00,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"o4-mini": {
		InputTokenPrice:  1.00,
		OutputTokenPrice: 4.00,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},

	// Gemini models
	"gemini-3-pro-preview": {
		InputTokenPrice:  1.25,
		OutputTokenPrice: 5.00,
		CacheReadPrice:   0.00,  // Gemini has caching but different pricing model
		CacheWritePrice:  0.00,
	},
	"gemini-2.5-flash": {
		InputTokenPrice:  0.075,
		OutputTokenPrice: 0.30,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"gemini-2.5-flash-lite": {
		InputTokenPrice:  0.025,
		OutputTokenPrice: 0.10,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"gemini-2.5-pro": {
		InputTokenPrice:  1.25,
		OutputTokenPrice: 5.00,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
}
```

**Step 2: Update tracker to handle unknown models gracefully**

```go
// internal/cost/tracker.go (update)
func (t *Tracker) TrackUsage(model string, inputTokens, outputTokens int64) error {
	pricing, err := getPricing(model)
	if err != nil {
		// Log warning but don't fail - use default pricing
		log.Printf("Warning: unknown model %s, using default pricing", model)
		pricing = &PricingModel{
			InputTokenPrice:  3.00,  // Default to Sonnet pricing
			OutputTokenPrice: 15.00,
		}
	}

	// ... rest of tracking logic
}
```

**Step 3: Test cost tracking**

```go
// internal/cost/pricing_test.go (add tests)
func TestGetPricingAllProviders(t *testing.T) {
	tests := []struct{
		model   string
		wantErr bool
	}{
		{"claude-sonnet-4-5-20250929", false},
		{"gpt-4o", false},
		{"gemini-2.5-flash", false},
		{"unknown-model", true},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			_, err := getPricing(tt.model)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

**Step 4: Run tests**

```bash
go test ./internal/cost -v
```

**Step 5: Commit**

```bash
git add internal/cost/
git commit -m "feat: add pricing for OpenAI and Gemini models

Add pricing data for GPT-4o, o3, o4-mini.
Add pricing data for Gemini 2.5 and 3.0 models.
Handle unknown models gracefully with default pricing.

Part of multi-provider support implementation."
```

---

## Task 11: Integration Testing

**Files:**
- Create: `test/integration/multi_provider_test.go`
- Create: `.scratch/scenario_multi_provider.sh`

**Step 1: Write integration test**

```go
// test/integration/multi_provider_test.go
//go:build integration
// +build integration

package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/harper/hex/internal/providers"
	"github.com/harper/hex/internal/providers/anthropic"
	"github.com/harper/hex/internal/providers/openai"
	"github.com/harper/hex/internal/providers/gemini"
	"github.com/harper/hex/internal/providers/openrouter"
)

func TestAnthropicIntegration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	provider := anthropic.NewProvider(providers.ProviderConfig{
		APIKey: apiKey,
	})

	testProvider(t, provider, "claude-haiku-4-5-20251001")
}

func TestOpenAIIntegration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	provider := openai.NewProvider(providers.ProviderConfig{
		APIKey: apiKey,
	})

	testProvider(t, provider, "gpt-4o-mini")
}

func TestGeminiIntegration(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	provider := gemini.NewProvider(providers.ProviderConfig{
		APIKey: apiKey,
	})

	testProvider(t, provider, "gemini-2.5-flash")
}

func TestOpenRouterIntegration(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set")
	}

	provider := openrouter.NewProvider(providers.ProviderConfig{
		APIKey: apiKey,
	})

	testProvider(t, provider, "anthropic/claude-haiku-4-5")
}

func testProvider(t *testing.T, provider providers.Provider, model string) {
	req := &providers.MessageRequest{
		Model: model,
		Messages: []providers.Message{
			{Role: "user", Content: "Say 'test passed' and nothing else"},
		},
		MaxTokens: 10,
		Stream:    true,
	}

	stream, err := provider.CreateStream(context.Background(), req)
	require.NoError(t, err)
	defer stream.Close()

	var content string
	for {
		chunk, err := stream.Next()
		if err != nil {
			break
		}
		content += chunk.Content
	}

	assert.Contains(t, content, "test")
}
```

**Step 2: Write scenario test script**

```bash
#!/usr/bin/env bash
# .scratch/scenario_multi_provider.sh
# Test multi-provider support with real API calls

set -e

echo "Testing multi-provider support..."

# Test Anthropic
echo "Testing Anthropic provider..."
./hex --provider anthropic --model claude-haiku-4-5-20251001 -p "Say 'Anthropic works'"

# Test OpenAI (if key available)
if [ -n "$OPENAI_API_KEY" ]; then
    echo "Testing OpenAI provider..."
    ./hex --provider openai --model gpt-4o-mini -p "Say 'OpenAI works'"
fi

# Test Gemini (if key available)
if [ -n "$GEMINI_API_KEY" ]; then
    echo "Testing Gemini provider..."
    ./hex --provider gemini --model gemini-2.5-flash -p "Say 'Gemini works'"
fi

# Test OpenRouter (if key available)
if [ -n "$OPENROUTER_API_KEY" ]; then
    echo "Testing OpenRouter provider..."
    ./hex --provider openrouter --model anthropic/claude-haiku-4-5 -p "Say 'OpenRouter works'"
fi

# Test config file provider selection
echo "provider = \"anthropic\"" > /tmp/test_config.toml
echo "model = \"claude-haiku-4-5-20251001\"" >> /tmp/test_config.toml
echo "[providers.anthropic]" >> /tmp/test_config.toml
echo "api_key = \"$ANTHROPIC_API_KEY\"" >> /tmp/test_config.toml

HEX_CONFIG_PATH=/tmp/test_config.toml ./hex -p "Say 'config works'"

echo "✅ All provider tests passed!"
```

**Step 3: Run integration tests**

```bash
# Run Go integration tests
go test -tags=integration ./test/integration -v

# Run scenario test
chmod +x .scratch/scenario_multi_provider.sh
./.scratch/scenario_multi_provider.sh
```

**Step 4: Commit**

```bash
git add test/integration/ .scratch/scenario_multi_provider.sh
git commit -m "test: add multi-provider integration tests

Add integration tests for all four providers.
Test real API calls with streaming.
Add scenario script for end-to-end testing.

Part of multi-provider support implementation."
```

---

## Task 12: Documentation and Final Verification

**Files:**
- Create: `docs/MULTI_PROVIDER.md`
- Update: `README.md`
- Update: `docs/claude-docs/11-CONFIGURATION.md`

**Step 1: Write multi-provider documentation**

```markdown
<!-- docs/MULTI_PROVIDER.md -->
# Multi-Provider Support

Hex supports multiple LLM providers: Anthropic, OpenAI, Google Gemini, and OpenRouter.

## Configuration

### Config File (~/.hex/config.toml)

```toml
# Default provider
provider = "anthropic"
model = "claude-sonnet-4-5-20250929"

[providers.anthropic]
api_key = "sk-ant-..."

[providers.openai]
api_key = "sk-proj-..."

[providers.gemini]
api_key = "AIza..."

[providers.openrouter]
api_key = "sk-or-..."
```

### Environment Variables

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-proj-..."
export GEMINI_API_KEY="AIza..."
export OPENROUTER_API_KEY="sk-or-..."
```

### CLI Usage

```bash
# Use Anthropic (default)
hex "Hello"

# Use OpenAI
hex --provider openai --model gpt-4o "Hello"

# Use Gemini
hex --provider gemini --model gemini-2.5-flash "Hello"

# Use OpenRouter
hex --provider openrouter --model anthropic/claude-sonnet-4-5 "Hello"
```

## Supported Models

### Anthropic
- claude-opus-4-5-20251101
- claude-sonnet-4-5-20250929
- claude-haiku-4-5-20251001

### OpenAI
- gpt-4o
- gpt-4o-mini
- o3 (reasoning)
- o4-mini (reasoning)

### Google Gemini
- gemini-3-pro-preview
- gemini-2.5-pro
- gemini-2.5-flash
- gemini-2.5-flash-lite

### OpenRouter
- All models via provider/model format
- Example: anthropic/claude-sonnet-4-5, openai/gpt-4o

## Cost Tracking

Hex tracks token usage and costs for all providers. View with:

```bash
hex cost
```

## Migration from YAML

Existing `~/.hex/config.yaml` will be automatically migrated to `config.toml` on first run.
```

**Step 2: Update README.md**

```markdown
<!-- README.md - add section -->
## Multi-Provider Support

Hex supports multiple LLM providers:

- **Anthropic** (Claude Opus, Sonnet, Haiku)
- **OpenAI** (GPT-4o, o3 reasoning models)
- **Google Gemini** (2.5/3.0 Pro, Flash)
- **OpenRouter** (Unified access to all providers)

Select provider with `--provider` flag or configure in `~/.hex/config.toml`.

See [Multi-Provider Documentation](docs/MULTI_PROVIDER.md) for details.
```

**Step 3: Run final verification**

```bash
# Build
make build

# Run tests
make test

# Test each provider (requires API keys)
./hex --provider anthropic --model claude-haiku-4-5-20251001 "test"
./hex --provider openai --model gpt-4o-mini "test"
./hex --provider gemini --model gemini-2.5-flash "test"
./hex --provider openrouter --model anthropic/claude-haiku-4-5 "test"

# Check cost tracking
./hex cost

# Verify config migration
cat ~/.hex/config.toml
```

**Step 4: Commit documentation**

```bash
git add docs/MULTI_PROVIDER.md README.md docs/claude-docs/
git commit -m "docs: add multi-provider documentation

Document configuration, usage, and supported models.
Update README with provider support section.
Add migration notes for YAML to TOML.

Completes multi-provider support implementation."
```

**Step 5: Final integration commit**

```bash
git add -A
git commit -m "feat: complete multi-provider support implementation

Summary of changes:
- Provider abstraction layer with factory pattern
- Anthropic provider (refactored from core)
- OpenAI provider with Chat Completions API
- Gemini provider with streamGenerateContent API
- OpenRouter provider (OpenAI-compatible proxy)
- Config migration from YAML to TOML
- Per-provider API key configuration
- Database migration for provider column
- CLI --provider flag
- Cost tracking for all providers
- Integration tests and documentation

Enables users to select between Anthropic, OpenAI, Gemini,
and OpenRouter providers with unified interface.

Backward compatible with existing Anthropic configurations."
```

---

## Execution Plan

Use **superpowers:subagent-driven-development** to execute these tasks:

1. Each task is assigned to a fresh subagent
2. Subagent implements task following TDD
3. Code review between tasks
4. Commit after each task passes tests

## Success Criteria

- [ ] All four providers implemented and tested
- [ ] Config system migrated to TOML
- [ ] Database stores provider per conversation
- [ ] CLI accepts --provider flag
- [ ] Cost tracking works for all providers
- [ ] All unit tests pass
- [ ] Integration tests pass for all providers
- [ ] Backward compatibility maintained
- [ ] Documentation complete

## Estimated Time

12 tasks × 30 minutes = 6 hours with subagent-driven development
