// ABOUTME: Performance benchmarks for API client operations
// ABOUTME: Measures request marshaling, HTTP round trips, and client creation
package core_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/harper/clem/internal/core"
)

// BenchmarkClientCreation measures client initialization overhead
func BenchmarkClientCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = core.NewClient("test-api-key")
	}
}

// BenchmarkRequestMarshaling measures JSON encoding performance
func BenchmarkRequestMarshaling(b *testing.B) {
	req := core.MessageRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []core.Message{
			{
				Role:    "user",
				Content: "Hello, how are you today?",
			},
		},
		MaxTokens: 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkResponseUnmarshaling measures JSON decoding performance
func BenchmarkResponseUnmarshaling(b *testing.B) {
	respJSON := `{
		"id": "msg_123",
		"type": "message",
		"role": "assistant",
		"content": [
			{
				"type": "text",
				"text": "Hello! I'm doing well, thank you for asking."
			}
		],
		"model": "claude-3-5-sonnet-20241022",
		"stop_reason": "end_turn",
		"usage": {
			"input_tokens": 10,
			"output_tokens": 20
		}
	}`

	data := []byte(respJSON)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var resp core.MessageResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHTTPRoundTrip measures complete HTTP request/response cycle
func BenchmarkHTTPRoundTrip(b *testing.B) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(core.MessageResponse{
			ID:   "msg_test",
			Type: "message",
			Role: "assistant",
			Content: []core.Content{
				{Type: "text", Text: "Test response"},
			},
			Model:      "claude-3-5-sonnet-20241022",
			StopReason: "end_turn",
		})
	}))
	defer server.Close()

	client := core.NewClient("test-key", core.WithBaseURL(server.URL))
	req := core.MessageRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []core.Message{
			{Role: "user", Content: "Test"},
		},
		MaxTokens: 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.CreateMessage(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLargeMessagePayload measures performance with large messages
func BenchmarkLargeMessagePayload(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(core.MessageResponse{
			ID:         "msg_test",
			Type:       "message",
			Role:       "assistant",
			Content:    []core.Content{{Type: "text", Text: "OK"}},
			Model:      "claude-3-5-sonnet-20241022",
			StopReason: "end_turn",
		})
	}))
	defer server.Close()

	client := core.NewClient("test-key", core.WithBaseURL(server.URL))

	// Create 100KB of content
	largeContent := string(bytes.Repeat([]byte("x"), 100*1024))

	req := core.MessageRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []core.Message{
			{Role: "user", Content: largeContent},
		},
		MaxTokens: 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.CreateMessage(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCreateMessageWithTools measures request with tool definitions
func BenchmarkCreateMessageWithTools(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(core.MessageResponse{
			ID:         "msg_test",
			Type:       "message",
			Role:       "assistant",
			Content:    []core.Content{{Type: "text", Text: "Test"}},
			Model:      "claude-3-5-sonnet-20241022",
			StopReason: "end_turn",
		})
	}))
	defer server.Close()

	client := core.NewClient("test-key", core.WithBaseURL(server.URL))
	req := core.MessageRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []core.Message{
			{Role: "user", Content: "Test"},
		},
		MaxTokens: 1024,
		Tools: []core.ToolDefinition{
			{
				Name:        "read_file",
				Description: "Read a file from the filesystem",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]string{
							"type":        "string",
							"description": "Path to file",
						},
					},
					"required": []string{"path"},
				},
			},
			{
				Name:        "write_file",
				Description: "Write a file to the filesystem",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]string{
							"type":        "string",
							"description": "Path to file",
						},
						"content": map[string]string{
							"type":        "string",
							"description": "File content",
						},
					},
					"required": []string{"path", "content"},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.CreateMessage(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
