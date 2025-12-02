// ABOUTME: Gmail provider implementation for email, calendar, and tasks
// ABOUTME: Integrates with Google APIs via OAuth2 for productivity operations

// Package gmail implements the Gmail productivity provider for Pagen
package gmail

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/harper/clem/internal/providers"
	"github.com/harper/clem/internal/providers/oauth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailProvider implements the Provider interface for Google services
//
//nolint:revive // GmailProvider is intentional for clarity
type GmailProvider struct {
	config       *oauth2.Config
	token        *oauth2.Token
	clientID     string
	clientSecret string
	tokenPath    string
	ctx          context.Context
}

// NewGmailProvider creates a new Gmail provider instance
func NewGmailProvider() *GmailProvider {
	return &GmailProvider{
		ctx: context.Background(),
	}
}

// Name returns the provider name
func (g *GmailProvider) Name() string {
	return "gmail"
}

// SupportedTools returns the list of tools this provider implements
func (g *GmailProvider) SupportedTools() []string {
	return []string{
		// Email tools
		"send_email",
		"reply_email",
		"search_emails",
		"read_email",
		"archive_email",
		"mark_email_read",
		"mark_email_unread",
		"label_email",
		"delete_email",

		// Calendar tools
		"create_event",
		"update_event",
		"delete_event",
		"list_events",
		"search_events",
		"find_free_time",

		// Task tools
		"create_task",
		"update_task",
		"complete_task",
		"list_tasks",
		"delete_task",
	}
}

// Initialize sets up the provider with configuration
func (g *GmailProvider) Initialize(config map[string]string) error {
	// Extract configuration
	clientID, ok := config["client_id"]
	if !ok || clientID == "" {
		return fmt.Errorf("missing required config: client_id")
	}

	clientSecret, ok := config["client_secret"]
	if !ok || clientSecret == "" {
		return fmt.Errorf("missing required config: client_secret")
	}

	tokenPath, ok := config["token_file"]
	if !ok || tokenPath == "" {
		// Default token path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		tokenPath = filepath.Join(homeDir, ".pagen", "tokens", "gmail.json")
	}

	g.clientID = clientID
	g.clientSecret = clientSecret
	g.tokenPath = tokenPath

	// Create OAuth2 config
	g.config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.modify",
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/tasks",
		},
	}

	// Try to load existing token
	if err := g.loadToken(); err != nil {
		// Token doesn't exist yet, that's okay
		// User will need to call Authenticate()
		return nil
	}

	return nil
}

// Authenticate performs OAuth2 authentication flow
func (g *GmailProvider) Authenticate() error {
	// Generate random state for CSRF protection
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Start local callback server
	callbackServer, err := oauth.NewCallbackServer(state)
	if err != nil {
		return fmt.Errorf("failed to create callback server: %w", err)
	}

	if err := callbackServer.Start(); err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}
	defer func() { _ = callbackServer.Stop() }()

	// Set redirect URI
	g.config.RedirectURL = callbackServer.RedirectURI()

	// Generate OAuth URL
	authURL := g.config.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce)

	// Print instructions
	fmt.Printf("\n🔐 Gmail Authentication Required\n\n")
	fmt.Printf("Opening your browser to authenticate...\n")
	fmt.Printf("If the browser doesn't open automatically, visit:\n%s\n\n", authURL)

	// TODO: Open browser automatically
	// For now, user must copy/paste URL

	// Wait for callback with timeout
	select {
	case code := <-callbackServer.CodeChan:
		// Exchange code for token
		token, err := g.config.Exchange(g.ctx, code)
		if err != nil {
			return fmt.Errorf("failed to exchange code for token: %w", err)
		}

		g.token = token

		// Save token
		if err := g.saveToken(token); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		fmt.Printf("\n✅ Authentication successful!\n\n")
		return nil

	case err := <-callbackServer.ErrChan:
		return fmt.Errorf("oauth callback error: %w", err)

	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authentication timeout after 5 minutes")
	}
}

// Close cleans up provider resources
func (g *GmailProvider) Close() error {
	// Nothing to cleanup for Gmail provider
	return nil
}

// Status returns the current health status
func (g *GmailProvider) Status() providers.ProviderStatus {
	if g.token == nil {
		return providers.ProviderStatus{
			Healthy:   false,
			Message:   "Not authenticated",
			LastCheck: time.Now(),
		}
	}

	// Check if token is expired
	if g.token.Expiry.Before(time.Now()) {
		// Token is expired, but OAuth2 client will auto-refresh
		return providers.ProviderStatus{
			Healthy:   true,
			Message:   "Token expired, will auto-refresh",
			LastCheck: time.Now(),
		}
	}

	return providers.ProviderStatus{
		Healthy:   true,
		Message:   "Authenticated and ready",
		LastCheck: time.Now(),
	}
}

// Capabilities returns provider capabilities
func (g *GmailProvider) Capabilities() providers.ProviderCapabilities {
	return providers.ProviderCapabilities{
		RateLimits: map[string]int{
			// Gmail API quotas (per day, converted to per hour estimate)
			"send_email":    50,   // 500/day ÷ 10 hours
			"search_emails": 1000, // Very high limit
			"read_email":    1000,
		},
		Features: []string{
			"attachments",
			"labels",
			"threading",
			"calendar_sharing",
			"task_lists",
		},
		MaxResults: 500, // Maximum results per query
	}
}

// ExecuteTool executes a productivity tool
func (g *GmailProvider) ExecuteTool(toolName string, params map[string]interface{}) (providers.ToolResult, error) {
	// Check authentication
	if g.token == nil {
		return providers.ToolResult{
			Success: false,
			Error:   providers.ErrNotAuthenticated,
		}, fmt.Errorf(providers.ErrNotAuthenticated)
	}

	// Route to appropriate handler
	switch toolName {
	// Email tools
	case "send_email":
		return g.sendEmail(params)
	case "reply_email":
		return g.replyEmail(params)
	case "search_emails":
		return g.searchEmails(params)
	case "read_email":
		return g.readEmail(params)
	case "archive_email":
		return g.archiveEmail(params)
	case "mark_email_read":
		return g.markEmailRead(params, true)
	case "mark_email_unread":
		return g.markEmailRead(params, false)
	case "label_email":
		return g.labelEmail(params)
	case "delete_email":
		return g.deleteEmail(params)

	// Calendar tools
	case "create_event":
		return g.createEvent(params)
	case "update_event":
		return g.updateEvent(params)
	case "delete_event":
		return g.deleteEvent(params)
	case "list_events":
		return g.listEvents(params)
	case "search_events":
		return g.searchEvents(params)
	case "find_free_time":
		return g.findFreeTime(params)

	// Task tools
	case "create_task":
		return g.createTask(params)
	case "update_task":
		return g.updateTask(params)
	case "complete_task":
		return g.completeTask(params)
	case "list_tasks":
		return g.listTasks(params)
	case "delete_task":
		return g.deleteTask(params)

	default:
		return providers.ToolResult{
			Success: false,
			Error:   providers.ErrNotImplemented,
		}, fmt.Errorf("%s: %s", providers.ErrNotImplemented, toolName)
	}
}

// Token management

func (g *GmailProvider) loadToken() error {
	data, err := os.ReadFile(g.tokenPath)
	if err != nil {
		return err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return err
	}

	g.token = &token
	return nil
}

func (g *GmailProvider) saveToken(token *oauth2.Token) error {
	// Ensure directory exists
	dir := filepath.Dir(g.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return os.WriteFile(g.tokenPath, data, 0600)
}

// Helper methods

// getGmailService returns an authenticated Gmail API service
func (g *GmailProvider) getGmailService() (*gmail.Service, error) {
	client := g.config.Client(g.ctx, g.token)
	service, err := gmail.NewService(g.ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}
	return service, nil
}

// createMessage creates a Gmail message from email parameters
func (g *GmailProvider) createMessage(to, subject, body, cc, bcc string) string {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	if cc != "" {
		message.WriteString(fmt.Sprintf("Cc: %s\r\n", cc))
	}
	if bcc != "" {
		message.WriteString(fmt.Sprintf("Bcc: %s\r\n", bcc))
	}
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(body)
	return message.String()
}

// Tool implementations

func (g *GmailProvider) sendEmail(params map[string]interface{}) (providers.ToolResult, error) {
	// Extract parameters
	to, _ := params["to"].(string)
	subject, _ := params["subject"].(string)
	body, _ := params["body"].(string)
	cc, _ := params["cc"].(string)
	bcc, _ := params["bcc"].(string)

	if to == "" || subject == "" || body == "" {
		return providers.ToolResult{
			Success: false,
			Error:   "missing required parameters: to, subject, body",
		}, fmt.Errorf("missing required parameters")
	}

	// Get Gmail service
	service, err := g.getGmailService()
	if err != nil {
		return providers.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Create message
	messageStr := g.createMessage(to, subject, body, cc, bcc)
	message := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(messageStr)),
	}

	// Send message
	sent, err := service.Users.Messages.Send("me", message).Do()
	if err != nil {
		return providers.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to send email: %v", err),
		}, err
	}

	return providers.ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"message_id": sent.Id,
			"thread_id":  sent.ThreadId,
			"to":         to,
			"subject":    subject,
		},
	}, nil
}

func (g *GmailProvider) replyEmail(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) searchEmails(params map[string]interface{}) (providers.ToolResult, error) {
	// Extract parameters
	query, _ := params["query"].(string)
	maxResultsFloat, _ := params["max_results"].(float64) // JSON numbers come as float64
	maxResults := int64(maxResultsFloat)
	if maxResults == 0 {
		maxResults = 10 // Default to 10 results
	}

	// Get Gmail service
	service, err := g.getGmailService()
	if err != nil {
		return providers.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Search messages
	listCall := service.Users.Messages.List("me")
	if query != "" {
		listCall = listCall.Q(query)
	}
	listCall = listCall.MaxResults(maxResults)

	response, err := listCall.Do()
	if err != nil {
		return providers.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to search emails: %v", err),
		}, err
	}

	// Get full message details for each result
	emails := make([]map[string]interface{}, 0, len(response.Messages))
	for _, msg := range response.Messages {
		fullMsg, err := service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
		if err != nil {
			// Skip messages that fail to load
			continue
		}

		// Extract headers
		headers := make(map[string]string)
		for _, header := range fullMsg.Payload.Headers {
			headers[header.Name] = header.Value
		}

		emails = append(emails, map[string]interface{}{
			"id":        fullMsg.Id,
			"thread_id": fullMsg.ThreadId,
			"from":      headers["From"],
			"to":        headers["To"],
			"subject":   headers["Subject"],
			"date":      headers["Date"],
			"snippet":   fullMsg.Snippet,
		})
	}

	return providers.ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"emails":        emails,
			"count":         len(emails),
			"result_size":   response.ResultSizeEstimate,
			"next_page_tok": response.NextPageToken,
		},
	}, nil
}

func (g *GmailProvider) readEmail(params map[string]interface{}) (providers.ToolResult, error) {
	// Extract parameters
	messageID, _ := params["message_id"].(string)
	if messageID == "" {
		return providers.ToolResult{
			Success: false,
			Error:   "missing required parameter: message_id",
		}, fmt.Errorf("missing required parameter: message_id")
	}

	// Get Gmail service
	service, err := g.getGmailService()
	if err != nil {
		return providers.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Get full message
	msg, err := service.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		return providers.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read email: %v", err),
		}, err
	}

	// Extract headers
	headers := make(map[string]string)
	for _, header := range msg.Payload.Headers {
		headers[header.Name] = header.Value
	}

	// Extract body
	var body string
	if msg.Payload.Body.Data != "" {
		// Body is directly in the message
		bodyBytes, err := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
		if err == nil {
			body = string(bodyBytes)
		}
	} else if len(msg.Payload.Parts) > 0 {
		// Body is in parts (multipart message)
		for _, part := range msg.Payload.Parts {
			if part.MimeType == "text/plain" || part.MimeType == "text/html" {
				if part.Body.Data != "" {
					bodyBytes, err := base64.URLEncoding.DecodeString(part.Body.Data)
					if err == nil {
						body = string(bodyBytes)
						break // Use first text part found
					}
				}
			}
		}
	}

	return providers.ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"id":        msg.Id,
			"thread_id": msg.ThreadId,
			"from":      headers["From"],
			"to":        headers["To"],
			"cc":        headers["Cc"],
			"bcc":       headers["Bcc"],
			"subject":   headers["Subject"],
			"date":      headers["Date"],
			"body":      body,
			"snippet":   msg.Snippet,
			"labels":    msg.LabelIds,
		},
	}, nil
}

func (g *GmailProvider) archiveEmail(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) markEmailRead(_ map[string]interface{}, _ bool) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) labelEmail(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) deleteEmail(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) createEvent(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) updateEvent(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) deleteEvent(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) listEvents(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) searchEvents(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) findFreeTime(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) createTask(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) updateTask(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) completeTask(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) listTasks(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}

func (g *GmailProvider) deleteTask(_ map[string]interface{}) (providers.ToolResult, error) {
	return providers.ToolResult{Success: false, Error: "not yet implemented"}, fmt.Errorf("not yet implemented")
}
