// ABOUTME: Tests that hex honors a custom OpenRouter base URL from provider config.
// ABOUTME: Covers both the default mux print path and the legacy --legacy print path.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/mux/llm"
)

// openRouterStubServer returns an httptest server that records whether it was
// hit and the request path, replying with a minimal valid chat completion so a
// correctly-routed CreateMessage succeeds.
func openRouterStubServer(t *testing.T, hit *bool, gotPath *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*hit = true
		*gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    "chatcmpl-stub",
			"model": "anthropic/claude-3.5-sonnet",
			"choices": []map[string]any{
				{
					"message":       map[string]any{"role": "assistant", "content": "Hello!"},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]any{"prompt_tokens": 1, "completion_tokens": 1},
		})
	}))
}

// TestCreateMuxLLMClientOpenRouterHonorsBaseURL covers the default print path
// (hex -p), which builds the LLM client via createMuxLLMClient.
func TestCreateMuxLLMClientOpenRouterHonorsBaseURL(t *testing.T) {
	var hit bool
	var gotPath string
	server := openRouterStubServer(t, &hit, &gotPath)
	defer server.Close()

	cfg := &core.Config{
		ProviderConfigs: map[string]core.ProviderConfig{
			"openrouter": {APIKey: "test-key", BaseURL: server.URL},
		},
	}

	client, _, err := createMuxLLMClient(cfg, "openrouter", "anthropic/claude-3.5-sonnet")
	if err != nil {
		t.Fatalf("createMuxLLMClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// We do not assert on the error: when base_url is honored the stub replies
	// 200, so the meaningful assertion is whether our stub received the request.
	_, msgErr := client.CreateMessage(ctx, &llm.Request{
		Messages: []llm.Message{llm.NewUserMessage("Hello")},
	})

	if !hit {
		t.Fatalf("custom OpenRouter base_url was ignored: stub server never received the request (CreateMessage err: %v)", msgErr)
	}
	if gotPath != "/chat/completions" {
		t.Errorf("expected request to /chat/completions, got %s", gotPath)
	}
}

// TestCreateProviderOpenRouterHonorsBaseURL covers the legacy print path
// (hex -p --legacy), which builds the provider via createProvider.
func TestCreateProviderOpenRouterHonorsBaseURL(t *testing.T) {
	var hit bool
	var gotPath string
	server := openRouterStubServer(t, &hit, &gotPath)
	defer server.Close()

	cfg := &core.Config{
		ProviderConfigs: map[string]core.ProviderConfig{
			"openrouter": {APIKey: "test-key", BaseURL: server.URL},
		},
	}

	provider, err := createProvider(cfg, "openrouter")
	if err != nil {
		t.Fatalf("createProvider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, msgErr := provider.CreateMessage(ctx, core.MessageRequest{
		Model:     "anthropic/claude-3.5-sonnet",
		Messages:  []core.Message{{Role: "user", Content: "Hello"}},
		MaxTokens: 16,
	})

	if !hit {
		t.Fatalf("custom OpenRouter base_url was ignored (legacy path): stub server never received the request (CreateMessage err: %v)", msgErr)
	}
	if gotPath != "/chat/completions" {
		t.Errorf("expected request to /chat/completions, got %s", gotPath)
	}
}
