// ABOUTME: WebSearch tool that searches DuckDuckGo and returns formatted results
// ABOUTME: Supports domain filtering (allowed/blocked) and result limiting

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// WebSearchTool searches the web using DuckDuckGo and returns formatted results
type WebSearchTool struct {
	baseURL    string
	httpClient *http.Client
}

// SearchResult represents a single web search result
type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

// NewWebSearchTool creates a new web search tool instance
func NewWebSearchTool() Tool {
	return &WebSearchTool{
		baseURL:    "https://html.duckduckgo.com/html/",
		httpClient: &http.Client{},
	}
}

// newWebSearchToolWithURL creates a tool with a custom base URL (for testing)
func newWebSearchToolWithURL(baseURL string) *WebSearchTool {
	return &WebSearchTool{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// Name returns the tool's identifier
func (t *WebSearchTool) Name() string {
	return "web_search"
}

// Description returns a human-readable description of the tool
func (t *WebSearchTool) Description() string {
	return "Search the web using DuckDuckGo and return results"
}

// RequiresApproval returns true since this tool makes network requests
func (t *WebSearchTool) RequiresApproval(_ map[string]interface{}) bool {
	// Always requires approval since it makes network requests
	return true
}

// Execute performs the web search and returns formatted results
func (t *WebSearchTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Extract and validate query parameter
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "query parameter is required and must be a non-empty string",
		}, nil
	}

	// Extract optional limit parameter
	limit := 10 // default
	if limitParam, ok := params["limit"]; ok {
		switch v := limitParam.(type) {
		case int:
			limit = v
		case float64:
			limit = int(v)
		default:
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    "limit parameter must be a number",
			}, nil
		}
		if limit <= 0 {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    "limit must be greater than 0",
			}, nil
		}
	}

	// Extract optional domain filters
	allowedDomains := extractDomains(params["allowed_domains"])
	blockedDomains := extractDomains(params["blocked_domains"])

	// Perform the search
	results, err := t.search(ctx, query)
	if err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("search failed: %v", err),
		}, nil
	}

	// Filter results by domains
	filtered := filterResults(results, allowedDomains, blockedDomains)

	// Apply limit
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Format results as markdown
	output := formatResults(filtered, query)

	return &Result{
		ToolName: t.Name(),
		Success:  true,
		Output:   output,
	}, nil
}

func (t *WebSearchTool) search(ctx context.Context, query string) ([]SearchResult, error) {
	// Build search URL
	searchURL := t.baseURL + "?q=" + url.QueryEscape(query)

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Hex/1.0)")

	// Execute request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	// Parse HTML response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	results, err := parseResults(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	return results, nil
}

func parseResults(htmlContent string) ([]SearchResult, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			// Check if this is a result div (class contains "result" but not "results")
			for _, attr := range n.Attr {
				if attr.Key == "class" {
					// Match "result" class but not "results" (container)
					classes := strings.Fields(attr.Val)
					for _, class := range classes {
						if class == "result" {
							result := &SearchResult{}
							// Parse children of this result div
							traverseResult(n, result)
							if result.URL != "" {
								results = append(results, *result)
							}
							// Don't traverse children again
							return
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return results, nil
}

func traverseResult(n *html.Node, result *SearchResult) {
	if n.Type == html.ElementNode && n.Data == "a" {
		isTitle := false
		isSnippet := false

		for _, attr := range n.Attr {
			if attr.Key == "class" {
				if strings.Contains(attr.Val, "result__a") {
					isTitle = true
					// Extract href
					for _, a := range n.Attr {
						if a.Key == "href" {
							result.URL = a.Val
							break
						}
					}
				} else if strings.Contains(attr.Val, "result__snippet") {
					isSnippet = true
				}
			}
		}

		if isTitle && n.FirstChild != nil {
			result.Title = extractText(n)
		} else if isSnippet && n.FirstChild != nil {
			result.Snippet = extractText(n)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverseResult(c, result)
	}
}

func extractText(n *html.Node) string {
	var text strings.Builder
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return strings.TrimSpace(text.String())
}

func extractDomains(param interface{}) []string {
	if param == nil {
		return nil
	}

	slice, ok := param.([]interface{})
	if !ok {
		return nil
	}

	domains := make([]string, 0, len(slice))
	for _, item := range slice {
		if str, ok := item.(string); ok {
			domains = append(domains, strings.ToLower(str))
		}
	}
	return domains
}

func filterResults(results []SearchResult, allowedDomains, blockedDomains []string) []SearchResult {
	if len(allowedDomains) == 0 && len(blockedDomains) == 0 {
		return results
	}

	filtered := make([]SearchResult, 0, len(results))
	for _, result := range results {
		domain := extractDomain(result.URL)
		if domain == "" {
			continue
		}

		domain = strings.ToLower(domain)

		// Check blocked domains first
		if containsDomain(blockedDomains, domain) {
			continue
		}

		// If allowed domains specified, check membership
		if len(allowedDomains) > 0 && !containsDomain(allowedDomains, domain) {
			continue
		}

		filtered = append(filtered, result)
	}

	return filtered
}

func extractDomain(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func containsDomain(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func formatResults(results []SearchResult, query string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Search Results for: %s\n\n", query))

	if len(results) == 0 {
		sb.WriteString("No results found.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Found %d results:\n\n", len(results)))

	for i, result := range results {
		sb.WriteString(fmt.Sprintf("### %d. %s\n", i+1, result.Title))
		sb.WriteString(fmt.Sprintf("**URL**: %s\n\n", result.URL))
		if result.Snippet != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", result.Snippet))
		}
		sb.WriteString("---\n\n")
	}

	return sb.String()
}
