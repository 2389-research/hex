// ABOUTME: Tests for the WebSearch tool that searches DuckDuckGo and returns results
// ABOUTME: Validates search functionality, domain filtering, result limiting, and error handling

package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Mock DuckDuckGo HTML response
const mockDDGResponse = `
<!DOCTYPE html>
<html>
<body>
	<div class="results">
		<div class="result">
			<a class="result__a" href="https://example.com/page1">Example Page 1</a>
			<a class="result__snippet">This is the first example result snippet</a>
		</div>
		<div class="result">
			<a class="result__a" href="https://test.org/page2">Test Page 2</a>
			<a class="result__snippet">This is the second test result snippet</a>
		</div>
		<div class="result">
			<a class="result__a" href="https://example.com/page3">Example Page 3</a>
			<a class="result__snippet">This is the third example result snippet</a>
		</div>
		<div class="result">
			<a class="result__a" href="https://blocked.com/page4">Blocked Page 4</a>
			<a class="result__snippet">This should be blocked</a>
		</div>
		<div class="result">
			<a class="result__a" href="https://another.net/page5">Another Page 5</a>
			<a class="result__snippet">This is the fifth result snippet</a>
		</div>
	</div>
</body>
</html>
`

func TestWebSearchTool_Name(t *testing.T) {
	tool := NewWebSearchTool()
	expected := "web_search"
	if tool.Name() != expected {
		t.Errorf("Name() = %q, want %q", tool.Name(), expected)
	}
}

func TestWebSearchTool_Description(t *testing.T) {
	tool := NewWebSearchTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(strings.ToLower(desc), "search") {
		t.Errorf("Description() = %q, should mention 'search'", desc)
	}
}

func TestWebSearchTool_RequiresApproval(t *testing.T) {
	tool := NewWebSearchTool()

	tests := []struct {
		name   string
		params map[string]interface{}
		want   bool
	}{
		{
			name:   "always requires approval - empty params",
			params: map[string]interface{}{},
			want:   true,
		},
		{
			name: "always requires approval - with query",
			params: map[string]interface{}{
				"query": "golang testing",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tool.RequiresApproval(tt.params); got != tt.want {
				t.Errorf("RequiresApproval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebSearchTool_Execute_MissingQuery(t *testing.T) {
	tool := NewWebSearchTool()
	ctx := context.Background()

	params := map[string]interface{}{}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Execute() with missing query should fail")
	}
	if !strings.Contains(result.Error, "query") {
		t.Errorf("Error should mention 'query': %s", result.Error)
	}
}

func TestWebSearchTool_Execute_InvalidLimit(t *testing.T) {
	tool := NewWebSearchTool()
	ctx := context.Background()

	tests := []struct {
		name  string
		limit interface{}
	}{
		{"negative limit", -5},
		{"zero limit", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"query": "test",
				"limit": tt.limit,
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.Success {
				t.Error("Execute() with invalid limit should fail")
			}
			if !strings.Contains(result.Error, "limit") {
				t.Errorf("Error should mention 'limit': %s", result.Error)
			}
		})
	}
}

func TestWebSearchTool_Execute_BasicSearch(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/html/") {
			t.Errorf("Expected path to contain /html/, got %s", r.URL.Path)
		}
		query := r.URL.Query().Get("q")
		if query != "golang testing" {
			t.Errorf("Expected query 'golang testing', got %q", query)
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(mockDDGResponse))
	}))
	defer server.Close()

	// Create tool with custom HTTP client that uses our mock server
	tool := newWebSearchToolWithURL(server.URL + "/html/")
	ctx := context.Background()

	params := map[string]interface{}{
		"query": "golang testing",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result == nil {
		t.Fatal("Execute() returned nil result")
	}

	// Check that we got results
	output := result.Output
	if !strings.Contains(output, "Example Page 1") {
		t.Errorf("Result should contain 'Example Page 1', got: %s", output)
	}
	if !strings.Contains(output, "https://example.com/page1") {
		t.Errorf("Result should contain URL, got: %s", output)
	}
	if !strings.Contains(output, "first example result snippet") {
		t.Errorf("Result should contain snippet, got: %s", output)
	}
}

func TestWebSearchTool_Execute_WithLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(mockDDGResponse))
	}))
	defer server.Close()

	tool := newWebSearchToolWithURL(server.URL + "/html/")
	ctx := context.Background()

	params := map[string]interface{}{
		"query": "test",
		"limit": 2,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Count results by counting occurrences of result markers
	output := result.Output
	resultCount := strings.Count(output, "###")
	if resultCount > 2 {
		t.Errorf("Expected at most 2 results with limit=2, got %d results in: %s", resultCount, output)
	}
}

func TestWebSearchTool_Execute_WithAllowedDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(mockDDGResponse))
	}))
	defer server.Close()

	tool := newWebSearchToolWithURL(server.URL + "/html/")
	ctx := context.Background()

	params := map[string]interface{}{
		"query":           "test",
		"allowed_domains": []interface{}{"example.com"},
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.Output
	if !strings.Contains(output, "example.com") {
		t.Errorf("Result should contain allowed domain example.com, got: %s", output)
	}
	if strings.Contains(output, "test.org") {
		t.Errorf("Result should not contain non-allowed domain test.org, got: %s", output)
	}
}

func TestWebSearchTool_Execute_WithBlockedDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(mockDDGResponse))
	}))
	defer server.Close()

	tool := newWebSearchToolWithURL(server.URL + "/html/")
	ctx := context.Background()

	params := map[string]interface{}{
		"query":            "test",
		"blocked_domains": []interface{}{"blocked.com"},
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content := result.Output
	if strings.Contains(content, "blocked.com") {
		t.Errorf("Result should not contain blocked domain blocked.com, got: %s", content)
	}
	if !strings.Contains(content, "example.com") {
		t.Errorf("Result should contain non-blocked domain example.com, got: %s", content)
	}
}

func TestWebSearchTool_Execute_NoResults(t *testing.T) {
	emptyResponse := `
<!DOCTYPE html>
<html>
<body>
	<div class="results">
		<!-- No results -->
	</div>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(emptyResponse))
	}))
	defer server.Close()

	tool := newWebSearchToolWithURL(server.URL + "/html/")
	ctx := context.Background()

	params := map[string]interface{}{
		"query": "nonexistent query",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() with no results should not error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Execute() returned nil result")
	}

	content := result.Output
	if !strings.Contains(content, "No results found") && !strings.Contains(content, "0 results") {
		t.Errorf("Result should indicate no results found, got: %s", content)
	}
}

func TestWebSearchTool_Execute_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // Wait for context cancellation
	}))
	defer server.Close()

	tool := newWebSearchToolWithURL(server.URL + "/html/")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	params := map[string]interface{}{
		"query": "test",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Execute() with cancelled context should fail")
	}
	if !strings.Contains(result.Error, "search failed") {
		t.Errorf("Error should mention 'search failed': %s", result.Error)
	}
}

func TestWebSearchTool_Execute_DomainFilteringCaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(mockDDGResponse))
	}))
	defer server.Close()

	tool := newWebSearchToolWithURL(server.URL + "/html/")
	ctx := context.Background()

	params := map[string]interface{}{
		"query":           "test",
		"allowed_domains": []interface{}{"EXAMPLE.COM"}, // Uppercase
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content := result.Output
	if !strings.Contains(content, "example.com") {
		t.Errorf("Domain filtering should be case-insensitive, got: %s", content)
	}
}
