// ABOUTME: Tests for the WebFetch tool implementation
// ABOUTME: Validates URL fetching, HTML-to-markdown conversion, and error handling

package tools

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWebFetchTool_Name(t *testing.T) {
	tool := NewWebFetchTool()
	if tool.Name() != "web_fetch" {
		t.Errorf("Expected name 'web_fetch', got '%s'", tool.Name())
	}
}

func TestWebFetchTool_Description(t *testing.T) {
	tool := NewWebFetchTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	// Should mention fetching and URL
	if !strings.Contains(strings.ToLower(desc), "fetch") && !strings.Contains(strings.ToLower(desc), "url") {
		t.Errorf("Description should mention fetching or URL: %s", desc)
	}
}

func TestWebFetchTool_RequiresApproval(t *testing.T) {
	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"url":    "https://example.com",
		"prompt": "extract title",
	}

	// Should always require approval (network access)
	if !tool.RequiresApproval(params) {
		t.Error("WebFetch should always require approval")
	}

	// Even with empty params
	if !tool.RequiresApproval(map[string]interface{}{}) {
		t.Error("WebFetch should require approval even with empty params")
	}
}

func TestWebFetchTool_MissingURL(t *testing.T) {
	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"prompt": "extract title",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure when URL is missing")
	}
	if !strings.Contains(result.Error, "url") {
		t.Errorf("Error should mention 'url': %s", result.Error)
	}
}

func TestWebFetchTool_MissingPrompt(t *testing.T) {
	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"url": "https://example.com",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure when prompt is missing")
	}
	if !strings.Contains(result.Error, "prompt") {
		t.Errorf("Error should mention 'prompt': %s", result.Error)
	}
}

func TestWebFetchTool_InvalidURL(t *testing.T) {
	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"url":    "not-a-valid-url",
		"prompt": "extract title",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for invalid URL")
	}
	if !strings.Contains(result.Error, "url") && !strings.Contains(result.Error, "scheme") {
		t.Errorf("Error should mention 'url' or 'scheme': %s", result.Error)
	}
}

func TestWebFetchTool_FetchHTML(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>Test Title</h1><p>Test paragraph</p></body></html>")
	}))
	defer server.Close()

	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"url":    server.URL,
		"prompt": "extract content",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	content := result.Output
	if content == "" {
		t.Fatal("Expected non-empty output")
	}

	// Should convert HTML to markdown
	if !strings.Contains(content, "Test Title") {
		t.Errorf("Content should contain 'Test Title': %s", content)
	}
	if !strings.Contains(content, "Test paragraph") {
		t.Errorf("Content should contain 'Test paragraph': %s", content)
	}
	// Should not contain raw HTML tags
	if strings.Contains(content, "<html>") || strings.Contains(content, "<body>") {
		t.Errorf("Content should not contain raw HTML tags: %s", content)
	}
}

func TestWebFetchTool_FetchPlainText(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "This is plain text content.\nSecond line.")
	}))
	defer server.Close()

	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"url":    server.URL,
		"prompt": "extract content",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	content := result.Output
	if content == "" {
		t.Fatal("Expected non-empty output")
	}

	if !strings.Contains(content, "This is plain text content.") {
		t.Errorf("Content should contain plain text: %s", content)
	}
	if !strings.Contains(content, "Second line.") {
		t.Errorf("Content should contain second line: %s", content)
	}
}

func TestWebFetchTool_HTTP404(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found")
	}))
	defer server.Close()

	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"url":    server.URL,
		"prompt": "extract content",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for 404 response")
	}
	if !strings.Contains(result.Error, "404") {
		t.Errorf("Error should mention 404: %s", result.Error)
	}
}

func TestWebFetchTool_HTTP500(t *testing.T) {
	// Create test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"url":    server.URL,
		"prompt": "extract content",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for 500 response")
	}
	if !strings.Contains(result.Error, "500") {
		t.Errorf("Error should mention 500: %s", result.Error)
	}
}

func TestWebFetchTool_ContextCancellation(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, "Delayed response")
	}))
	defer server.Close()

	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"url":    server.URL,
		"prompt": "extract content",
	}

	// Create context that cancels immediately
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for cancelled context")
	}
	if !strings.Contains(result.Error, "context") && !strings.Contains(result.Error, "timeout") {
		t.Errorf("Error should mention context or timeout: %s", result.Error)
	}
}

func TestWebFetchTool_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow timeout test in short mode")
	}

	// Create test server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Second) // Long delay
	}))
	defer server.Close()

	tool := NewWebFetchTool()
	params := map[string]interface{}{
		"url":    server.URL,
		"prompt": "extract content",
	}

	// Should timeout with default timeout
	start := time.Now()
	result, err := tool.Execute(context.Background(), params)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected timeout failure")
	}

	// Should timeout in reasonable time (not 60s)
	if duration > 35*time.Second {
		t.Errorf("Timeout took too long: %v", duration)
	}
}
