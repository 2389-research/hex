// ABOUTME: Tests for Gmail provider implementation
// ABOUTME: Verifies provider interface compliance and basic functionality

package gmail

import (
	"testing"
	"time"

	"github.com/harper/jeff/internal/providers"
	"golang.org/x/oauth2"
)

func TestGmailProvider_Interface(_ *testing.T) {
	// Verify that GmailProvider implements Provider interface
	var _ providers.Provider = (*GmailProvider)(nil)
}

func TestGmailProvider_Name(t *testing.T) {
	p := NewGmailProvider()
	if p.Name() != "gmail" {
		t.Errorf("expected name 'gmail', got %q", p.Name())
	}
}

func TestGmailProvider_SupportedTools(t *testing.T) {
	p := NewGmailProvider()
	tools := p.SupportedTools()

	// Verify we have expected number of tools
	expectedCount := 20 // 9 email + 6 calendar + 5 task
	if len(tools) != expectedCount {
		t.Errorf("expected %d tools, got %d", expectedCount, len(tools))
	}

	// Verify critical tools are present
	criticalTools := []string{
		"send_email", "search_emails", "read_email",
		"create_event", "list_events",
		"create_task", "list_tasks", "complete_task",
	}

	toolMap := make(map[string]bool)
	for _, tool := range tools {
		toolMap[tool] = true
	}

	for _, critical := range criticalTools {
		if !toolMap[critical] {
			t.Errorf("missing critical tool: %s", critical)
		}
	}
}

func TestGmailProvider_Initialize(t *testing.T) {
	p := NewGmailProvider()

	tests := []struct {
		name        string
		config      map[string]string
		shouldError bool
	}{
		{
			name: "valid config",
			config: map[string]string{
				"client_id":     "test-client-id",
				"client_secret": "test-client-secret",
			},
			shouldError: false,
		},
		{
			name: "missing client_id",
			config: map[string]string{
				"client_secret": "test-client-secret",
			},
			shouldError: true,
		},
		{
			name: "missing client_secret",
			config: map[string]string{
				"client_id": "test-client-id",
			},
			shouldError: true,
		},
		{
			name:        "empty config",
			config:      map[string]string{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Initialize(tt.config)
			if tt.shouldError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGmailProvider_Status_NotAuthenticated(t *testing.T) {
	p := NewGmailProvider()

	// Provider without token should be unhealthy
	status := p.Status()
	if status.Healthy {
		t.Error("expected unhealthy status when not authenticated")
	}
	if status.Message != "Not authenticated" {
		t.Errorf("expected 'Not authenticated' message, got %q", status.Message)
	}
}

func TestGmailProvider_ExecuteTool_NotAuthenticated(t *testing.T) {
	p := NewGmailProvider()

	// Initialize but don't authenticate
	_ = p.Initialize(map[string]string{
		"client_id":     "test",
		"client_secret": "test",
	})

	// Try to execute a tool without authentication
	result, err := p.ExecuteTool("send_email", map[string]interface{}{
		"to":      "test@example.com",
		"subject": "Test",
		"body":    "Test body",
	})

	if err == nil {
		t.Error("expected error when executing tool without authentication")
	}
	if result.Success {
		t.Error("expected failed result when not authenticated")
	}
	if result.Error != providers.ErrNotAuthenticated {
		t.Errorf("expected ErrNotAuthenticated, got %q", result.Error)
	}
}

func TestGmailProvider_ExecuteTool_UnsupportedTool(t *testing.T) {
	p := NewGmailProvider()
	// Create a mock token to bypass auth check
	p.token = &oauth2.Token{
		AccessToken: "mock-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	result, err := p.ExecuteTool("unsupported_tool", map[string]interface{}{})

	if err == nil {
		t.Error("expected error for unsupported tool")
	}
	if result.Success {
		t.Error("expected failed result for unsupported tool")
	}
	if result.Error != providers.ErrNotImplemented {
		t.Errorf("expected ErrNotImplemented, got %q", result.Error)
	}
}

func TestGmailProvider_Capabilities(t *testing.T) {
	p := NewGmailProvider()
	caps := p.Capabilities()

	// Verify rate limits are set
	if caps.RateLimits == nil {
		t.Fatal("expected rate limits to be set")
	}

	// Check specific rate limits
	if _, ok := caps.RateLimits["send_email"]; !ok {
		t.Error("expected send_email rate limit")
	}

	// Verify features are set
	if len(caps.Features) == 0 {
		t.Error("expected features to be set")
	}

	// Verify max results is set
	if caps.MaxResults == 0 {
		t.Error("expected MaxResults to be set")
	}
}
