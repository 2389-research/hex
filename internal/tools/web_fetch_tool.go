// ABOUTME: WebFetch tool implementation for fetching and processing web content
// ABOUTME: Handles HTTP requests, HTML-to-markdown conversion, and error handling

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

// WebFetchTool fetches content from a URL and optionally processes it
type WebFetchTool struct {
	client    *http.Client
	converter *md.Converter
}

// NewWebFetchTool creates a new WebFetch tool instance
func NewWebFetchTool() Tool {
	return &WebFetchTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		converter: md.NewConverter("", true, nil),
	}
}

// Name returns the tool name
func (t *WebFetchTool) Name() string {
	return "web_fetch"
}

// Description returns the tool description
func (t *WebFetchTool) Description() string {
	return "Fetch content from a URL and process it with a prompt"
}

// RequiresApproval always returns true since this makes network requests
func (t *WebFetchTool) RequiresApproval(_ map[string]interface{}) bool {
	return true
}

// Execute fetches the URL and returns its content
func (t *WebFetchTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Validate URL parameter
	urlStr, ok := params["url"].(string)
	if !ok || urlStr == "" {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "url parameter is required",
		}, nil
	}

	// Validate prompt parameter
	_, ok = params["prompt"].(string)
	if !ok {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "prompt parameter is required",
		}, nil
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("invalid url: %v", err),
		}, nil
	}

	if parsedURL.Scheme == "" {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "invalid url: missing scheme (http/https)",
		}, nil
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	// Set user agent
	req.Header.Set("User-Agent", "Hex/1.0")

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("failed to fetch url: %v", err),
		}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("http error: %d %s", resp.StatusCode, resp.Status),
		}, nil
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("failed to read response: %v", err),
		}, nil
	}

	content := string(body)

	// Convert HTML to markdown if content type is HTML
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml") {
		markdown, err := t.converter.ConvertString(content)
		if err != nil {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("failed to convert html to markdown: %v", err),
			}, nil
		}
		content = markdown
	}

	return &Result{
		ToolName: t.Name(),
		Success:  true,
		Output:   content,
	}, nil
}
