// ABOUTME: Tests for the Anthropic API client
// ABOUTME: Uses VCR for recording/replaying HTTP interactions
package core_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/harper/clem/internal/core"
	"gopkg.in/dnaeon/go-vcr.v2/cassette"
	"gopkg.in/dnaeon/go-vcr.v2/recorder"
)

func TestCreateMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	// Create VCR recorder
	// First run with real API key will record, subsequent runs replay from cassette
	// To record: export CLEM_API_KEY=your-key && go test -run TestCreateMessage
	// To replay: go test -run TestCreateMessage
	r, err := recorder.New("testdata/fixtures/create_message")
	require.NoError(t, err)
	defer r.Stop()

	// Add a custom matcher that ignores timestamp differences in request body
	r.SetMatcher(func(r *http.Request, i cassette.Request) bool {
		// Match URL and method (body can vary due to timestamps)
		return r.URL.String() == i.URL && r.Method == i.Method
	})

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
