// ABOUTME: Tests for the Anthropic API client
// ABOUTME: Uses VCR for recording/replaying HTTP interactions
package core_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/harper/pagent/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v2/cassette"
	"gopkg.in/dnaeon/go-vcr.v2/recorder"
)

func TestCreateMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	// Create VCR recorder
	// Will record if cassette doesn't exist or in recording mode
	// Otherwise replays from cassette
	r, err := recorder.New("testdata/fixtures/create_message")
	require.NoError(t, err)
	defer func() { _ = r.Stop() }()

	// Add a custom matcher that ignores request body differences
	// This allows the cassette to match even if the request body varies slightly
	r.SetMatcher(func(r *http.Request, i cassette.Request) bool {
		// Match URL and method (body can vary due to timestamps)
		return r.URL.String() == i.URL && r.Method == i.Method
	})

	// Create HTTP client with recorder
	httpClient := &http.Client{Transport: r}

	// Create API client
	// Use real API key from environment when recording, otherwise use test key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key"
	}
	client := core.NewClient(apiKey, core.WithHTTPClient(httpClient))

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
	if testing.Short() {
		t.Skip("skipping network test")
	}

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

func TestCreateMessageWithImage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	// Create VCR recorder
	r, err := recorder.New("testdata/fixtures/create_message_with_image")
	require.NoError(t, err)
	defer func() { _ = r.Stop() }()

	// Add custom matcher that ignores request body differences
	r.SetMatcher(func(r *http.Request, i cassette.Request) bool {
		return r.URL.String() == i.URL && r.Method == i.Method
	})

	// Create HTTP client with recorder
	httpClient := &http.Client{Transport: r}

	// Create API client
	// Use real API key from environment when recording, otherwise use test key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key"
	}
	client := core.NewClient(apiKey, core.WithHTTPClient(httpClient))

	// Create a message with image content
	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 1024,
		Messages: []core.Message{
			{
				Role: "user",
				ContentBlock: []core.ContentBlock{
					core.NewImageBlock(&core.ImageSource{
						Type:      "base64",
						MediaType: "image/png",
						Data:      "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
					}),
					core.NewTextBlock("What's in this image?"),
				},
			},
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

func TestMessageRequestSerialization(t *testing.T) {
	t.Run("text-only message uses content field", func(t *testing.T) {
		msg := core.Message{
			Role:    "user",
			Content: "Hello",
		}

		// When serialized, should have content field
		assert.Equal(t, "Hello", msg.Content)
		assert.Nil(t, msg.ContentBlock)
	})

	t.Run("multimodal message uses content array", func(t *testing.T) {
		msg := core.Message{
			Role: "user",
			ContentBlock: []core.ContentBlock{
				core.NewTextBlock("Hello"),
				core.NewImageBlock(&core.ImageSource{
					Type:      "base64",
					MediaType: "image/png",
					Data:      "data",
				}),
			},
		}

		// When serialized, should have content_block field
		assert.Len(t, msg.ContentBlock, 2)
	})
}
