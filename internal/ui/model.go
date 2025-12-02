// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Bubbletea model for interactive chat UI
// ABOUTME: Manages state, messages, input, viewport, and streaming
package ui

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	ctxmgr "github.com/harper/pagent/internal/context"
	"github.com/harper/pagent/internal/core"
	"github.com/harper/pagent/internal/storage"
	"github.com/harper/pagent/internal/tools"
	"github.com/harper/pagent/internal/ui/themes"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ViewMode represents different view modes in the UI
type ViewMode int

const (
	// ViewModeChat is the main chat interface view
	ViewModeChat ViewMode = iota
	// ViewModeHistory displays conversation history
	ViewModeHistory
	// ViewModeTools shows available tools and their status
	ViewModeTools
)

// Status represents the current UI status
type Status int

const (
	// StatusIdle indicates the assistant is not processing
	StatusIdle Status = iota
	// StatusTyping indicates the user is typing
	StatusTyping
	// StatusStreaming indicates the assistant is streaming a response
	StatusStreaming
	// StatusError indicates an error occurred
	StatusError
)

// Message represents a chat message in the UI
type Message struct {
	Role         string
	Content      string
	ContentBlock []core.ContentBlock // For structured content like tool_result blocks
}

// StreamChunk is an alias for core.StreamChunk for use in UI
type StreamChunk = core.StreamChunk

// Delta is an alias for core.Delta for use in UI
type Delta = core.Delta

// Usage is an alias for core.Usage for use in UI
type Usage = core.Usage

// StreamChunkMsg is a Bubbletea message carrying a streaming chunk
type StreamChunkMsg struct {
	Chunk *StreamChunk
	Error error
}

// streamStartMsg is sent when a stream is initialized
type streamStartMsg struct {
	channel <-chan *core.StreamChunk
}

// Model is the Bubbletea model for interactive mode
type Model struct {
	ConversationID string
	Model          string
	Messages       []Message
	Input          textarea.Model
	Viewport       viewport.Model
	Width          int
	Height         int
	Streaming      bool
	StreamingText  string
	Ready          bool

	// Task 5: Advanced UI Features
	CurrentView  ViewMode
	Status       Status
	ErrorMessage string
	TokensInput  int
	TokensOutput int
	SearchMode   bool
	SearchQuery  string
	renderer     *glamour.TermRenderer
	lastKeyWasG  bool // Track 'g' key for 'gg' navigation

	// Task 6: Streaming Integration
	apiClient    *core.Client
	streamChan   <-chan *core.StreamChunk
	streamCtx    context.Context
	streamCancel context.CancelFunc

	// Task 7: Storage Integration
	db *sql.DB

	// Task 12: Tool Execution UI
	toolRegistry      *tools.Registry
	toolExecutor      *tools.Executor
	pendingToolUses   []*core.ToolUse // Tools waiting for approval/execution (can be multiple)
	executingToolUses []*core.ToolUse // Tools currently being executed (for display)
	assemblingToolUse *core.ToolUse   // Tool being assembled from streaming chunks
	toolInputJSONBuf  string          // Buffer for accumulating input_json deltas
	toolApprovalMode  bool            // Showing approval prompt
	executingTool     bool            // Tool is running
	currentToolID     string          // ID of currently executing tool
	toolResults       []ToolResult    // Results to send back to API

	// Phase 6C: Enhanced UI Components
	spinner          *ToolSpinner
	approvalPrompt   *ApprovalPrompt
	streamingDisplay *StreamingDisplay
	statusBar        *StatusBar
	helpVisible      bool
	typewriterMode   bool

	// TUI Theme
	theme themes.Theme

	// Phase 6B: Context Management
	contextManager *ctxmgr.Manager
	contextUsage   ctxmgr.ContextUsage

	// Phase 6C Task 6: Quick Actions Menu
	quickActionsMode     bool                  // Quick actions menu is open
	quickActionsInput    string                // Current input in quick actions
	quickActionsFiltered []*QuickAction        // Filtered actions from search
	quickActionsRegistry *QuickActionsRegistry // Registry of available actions

	// Phase 6C Task 4: Autocomplete System
	autocomplete *Autocomplete

	// Phase 6C Task 8: Smart Suggestions
	suggestions        []*Suggestion // Current suggestions based on input
	showSuggestions    bool          // Whether to display suggestions
	suggestionDetector *SuggestionDetector
	suggestionLearner  *SuggestionLearner
	lastAnalyzedInput  string // Track last input to avoid re-analyzing

	// Phase 6C Task 3: Template System
	systemPrompt string // System prompt from template or custom

	// Phase 6C Task 5: Conversation Favorites
	IsFavorite bool // Track if current conversation is favorite
}

// ToolResult represents a tool execution result for the API
type ToolResult struct {
	ToolUseID string
	Result    *tools.Result
}

// NewModel creates a new UI model
func NewModel(conversationID, model, themeName string) *Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 10000
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to Clem! Type your message below.")

	// Initialize glamour renderer for markdown
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		renderer = nil // Explicit - RenderMessage already checks for nil
	}

	// Load theme
	theme := themes.GetTheme(themeName)

	// Phase 6C: Initialize new UI components
	spinner := NewToolSpinner()
	streamingDisplay := NewStreamingDisplay()
	statusBar := NewStatusBar(model, 80)
	quickActionsRegistry := NewQuickActionsRegistry()
	autocomplete := NewAutocomplete()

	// Phase 6C Task 8: Initialize suggestion system
	suggestionDetector := NewSuggestionDetector()
	suggestionLearner := NewSuggestionLearner()

	return &Model{
		ConversationID:       conversationID,
		Model:                model,
		Messages:             []Message{},
		Input:                ta,
		Viewport:             vp,
		Width:                80,
		Height:               24,
		CurrentView:          ViewModeChat,
		Status:               StatusIdle,
		renderer:             renderer,
		theme:                theme,
		spinner:              spinner,
		streamingDisplay:     streamingDisplay,
		statusBar:            statusBar,
		helpVisible:          false,
		typewriterMode:       false,
		quickActionsRegistry: quickActionsRegistry,
		quickActionsMode:     false,
		quickActionsInput:    "",
		quickActionsFiltered: []*QuickAction{},
		autocomplete:         autocomplete,
		suggestionDetector:   suggestionDetector,
		suggestionLearner:    suggestionLearner,
		suggestions:          []*Suggestion{},
		showSuggestions:      false,
		lastAnalyzedInput:    "",
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

// AddMessage adds a message to the conversation
func (m *Model) AddMessage(role, content string) {
	m.Messages = append(m.Messages, Message{
		Role:    role,
		Content: content,
	})
	// Update context usage after adding message
	m.updateContextUsage()
}

// NextView cycles to the next view mode
func (m *Model) NextView() {
	switch m.CurrentView {
	case ViewModeChat:
		m.CurrentView = ViewModeHistory
	case ViewModeHistory:
		m.CurrentView = ViewModeTools
	case ViewModeTools:
		m.CurrentView = ViewModeChat
	}
}

// UpdateTokens updates the token counters (cumulative across all messages in this session)
func (m *Model) UpdateTokens(input, output int) {
	m.TokensInput += input
	m.TokensOutput += output
}

// SetStatus sets the current UI status
func (m *Model) SetStatus(status Status) {
	m.Status = status
	if status == StatusError {
		m.ErrorMessage = "An error occurred"
	}
}

// RenderMessage renders a message using glamour for assistant messages
func (m *Model) RenderMessage(msg Message) (string, error) {
	if msg.Role == "assistant" && m.renderer != nil {
		rendered, err := m.renderer.Render(msg.Content)
		if err != nil {
			return msg.Content, err
		}
		return rendered, nil
	}
	return msg.Content, nil
}

// EnterSearchMode activates search mode
func (m *Model) EnterSearchMode() {
	m.SearchMode = true
	m.SearchQuery = ""
}

// ExitSearchMode deactivates search mode
func (m *Model) ExitSearchMode() {
	m.SearchMode = false
	m.SearchQuery = ""
}

// UpdateSearchQuery updates the search query
func (m *Model) UpdateSearchQuery(query string) {
	m.SearchQuery = query
}

// AppendStreamingText adds a chunk to the streaming buffer
func (m *Model) AppendStreamingText(chunk string) {
	m.StreamingText += chunk
}

// CommitStreamingText converts streaming text into a permanent assistant message
func (m *Model) CommitStreamingText() {
	if m.StreamingText != "" {
		m.AddMessage("assistant", m.StreamingText)

		// Task 7: Save assistant message to database
		if m.db != nil {
			_ = m.saveMessageInternal("assistant", m.StreamingText)
		}

		m.StreamingText = ""
	}
}

// saveMessageInternal is a helper method to save messages (called from model.go)
func (m *Model) saveMessageInternal(role, content string) error {
	if m.db == nil {
		return nil
	}

	msg := &storage.Message{
		ConversationID: m.ConversationID,
		Role:           role,
		Content:        content,
	}

	return storage.CreateMessage(m.db, msg)
}

// ClearStreamingText discards streaming buffer (e.g., on error)
func (m *Model) ClearStreamingText() {
	m.StreamingText = ""
}

// SetAPIClient sets the API client for streaming
func (m *Model) SetAPIClient(client *core.Client) {
	m.apiClient = client
}

// SetDB sets the database connection for storage
func (m *Model) SetDB(db *sql.DB) {
	m.db = db
}

// Task 12: Tool System Methods

// SetToolSystem sets the tool registry and executor
func (m *Model) SetToolSystem(registry *tools.Registry, executor *tools.Executor) {
	m.toolRegistry = registry
	m.toolExecutor = executor

	// FIX: Update autocomplete tool provider with available tools
	if m.autocomplete != nil && registry != nil {
		// Get existing provider and update it, or create new one
		provider, ok := m.autocomplete.GetProvider("tool")
		if ok {
			// Update existing provider's tool list
			if toolProvider, ok := provider.(*ToolProvider); ok {
				toolProvider.SetTools(registry.List())
			}
		} else {
			// Create new provider if it doesn't exist
			toolProvider := NewToolProvider(registry.List())
			m.autocomplete.RegisterProvider("tool", toolProvider)
		}
	}
}

// Phase 6B: Context Management Methods

// SetContextManager sets the context manager and initializes context tracking
func (m *Model) SetContextManager(manager *ctxmgr.Manager) {
	m.contextManager = manager
	m.updateContextUsage()
}

// Phase 6C: Template System Methods

// SetSystemPrompt sets the system prompt for this conversation
func (m *Model) SetSystemPrompt(prompt string) {
	m.systemPrompt = prompt
}

// updateContextUsage updates the context usage statistics and status bar
func (m *Model) updateContextUsage() {
	if m.contextManager == nil {
		return
	}

	// Convert UI messages to core.Message for estimation
	coreMessages := make([]core.Message, len(m.Messages))
	for i, msg := range m.Messages {
		coreMessages[i] = core.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	m.contextUsage = m.contextManager.GetUsage(coreMessages)

	// Update status bar with context info
	if m.statusBar != nil {
		m.statusBar.SetContextSize(m.contextManager.MaxTokens)
		m.statusBar.SetTokens(m.contextUsage.EstimatedTokens, 0) // Estimated as input for now

		// Show warning if near limit
		if m.contextUsage.NearLimit {
			m.statusBar.SetCustomMessage(fmt.Sprintf("⚠ Context %.0f%% full - pruning recommended", m.contextUsage.PercentUsed))
		} else {
			m.statusBar.ClearCustomMessage()
		}
	}
}

// GetPrunedMessages returns messages pruned to fit context limit
func (m *Model) GetPrunedMessages() []core.Message {
	if m.contextManager == nil {
		// No context manager, return all messages
		coreMessages := make([]core.Message, len(m.Messages))
		for i, msg := range m.Messages {
			coreMessages[i] = core.Message{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
		return coreMessages
	}

	// Convert to core messages
	coreMessages := make([]core.Message, len(m.Messages))
	for i, msg := range m.Messages {
		coreMessages[i] = core.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Prune if needed
	if m.contextManager.ShouldPrune(coreMessages) {
		return m.contextManager.Prune(coreMessages)
	}

	return coreMessages
}

// ApproveToolUse executes ALL pending tools
func (m *Model) ApproveToolUse() tea.Cmd {
	if len(m.pendingToolUses) == 0 || m.toolExecutor == nil {
		m.toolApprovalMode = false
		return nil
	}

	_, _ = fmt.Fprintf(os.Stderr, "[TOOL_APPROVAL] approving %d tool(s)\n", len(m.pendingToolUses))

	// Validate that all tool_use blocks exist in message history
	for _, toolUse := range m.pendingToolUses {
		m.validateToolUseExists(toolUse.ID)
	}

	// Capture tools and clear pending
	toolUses := m.pendingToolUses
	m.pendingToolUses = nil
	m.executingToolUses = toolUses // Save for status display
	m.toolApprovalMode = false
	m.approvalPrompt = nil // Clear approval prompt for next time
	m.executingTool = true

	// Phase 6C: Start spinner for tool execution
	var spinnerCmd tea.Cmd
	if m.spinner != nil {
		if len(toolUses) == 1 {
			spinnerCmd = m.spinner.Start(SpinnerTypeToolExecution, "Running "+toolUses[0].Name+"...")
		} else {
			spinnerCmd = m.spinner.Start(SpinnerTypeToolExecution, fmt.Sprintf("Running %d tools...", len(toolUses)))
		}
	}

	m.updateViewport()

	// Execute all tools in batch
	toolCmd := m.executeToolsBatch(toolUses)

	// Return batch of spinner start and tool execution
	if spinnerCmd != nil {
		return tea.Batch(spinnerCmd, toolCmd)
	}
	return toolCmd
}

// DenyToolUse rejects ALL pending tools
func (m *Model) DenyToolUse() tea.Cmd {
	if len(m.pendingToolUses) == 0 {
		m.toolApprovalMode = false
		return nil
	}

	_, _ = fmt.Fprintf(os.Stderr, "[TOOL_DENIAL] denying %d tool(s)\n", len(m.pendingToolUses))

	// Create error results for all denied tools
	for _, toolUse := range m.pendingToolUses {
		result := &tools.Result{
			ToolName: toolUse.Name,
			Success:  false,
			Error:    "User denied permission",
		}

		// Validate that tool_use exists in message history
		m.validateToolUseExists(toolUse.ID)

		m.toolResults = append(m.toolResults, ToolResult{
			ToolUseID: toolUse.ID,
			Result:    result,
		})

		m.AddMessage("tool", "Tool denied: "+toolUse.Name)
	}

	m.pendingToolUses = nil
	m.toolApprovalMode = false
	m.updateViewport()

	// Send all denial results back to API
	return m.sendToolResults()
}

// toolExecutionMsg is sent when a single tool finishes executing
type toolExecutionMsg struct {
	toolUseID string
	result    *tools.Result
	err       error
}

// toolBatchExecutionMsg is sent when a batch of tools finishes executing
type toolBatchExecutionMsg struct {
	results []ToolResult
}

// executeToolsBatch executes multiple tools sequentially and collects all results
func (m *Model) executeToolsBatch(toolUses []*core.ToolUse) tea.Cmd {
	return func() tea.Msg {
		results := make([]ToolResult, 0, len(toolUses))

		for i, toolUse := range toolUses {
			fmt.Fprintf(os.Stderr, "[BATCH_EXEC] executing tool %d/%d: %s (id=%s)\n",
				i+1, len(toolUses), toolUse.Name, toolUse.ID)

			ctx := context.Background()
			result, err := m.toolExecutor.Execute(ctx, toolUse.Name, toolUse.Input)

			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[BATCH_EXEC_ERROR] tool %s failed: %v\n", toolUse.Name, err)
				results = append(results, ToolResult{
					ToolUseID: toolUse.ID,
					Result: &tools.Result{
						ToolName: toolUse.Name,
						Success:  false,
						Error:    err.Error(),
					},
				})
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "[BATCH_EXEC_SUCCESS] tool %s succeeded\n", toolUse.Name)
				results = append(results, ToolResult{
					ToolUseID: toolUse.ID,
					Result:    result,
				})
			}
		}

		_, _ = fmt.Fprintf(os.Stderr, "[BATCH_EXEC_DONE] executed %d tools\n", len(results))
		return toolBatchExecutionMsg{results: results}
	}
}

// Phase 6C: Enhanced UI Methods

// GetAutocomplete returns the autocomplete instance
func (m *Model) GetAutocomplete() *Autocomplete {
	return m.autocomplete
}

// ToggleHelp toggles the help display
func (m *Model) ToggleHelp() {
	m.helpVisible = !m.helpVisible
}

// ToggleTypewriter toggles typewriter mode
func (m *Model) ToggleTypewriter() {
	m.typewriterMode = !m.typewriterMode
	if m.streamingDisplay != nil {
		m.streamingDisplay.ToggleTypewriterMode()
	}
}

// ClearScreen clears the viewport
func (m *Model) ClearScreen() {
	m.Viewport.SetContent("")
	m.Viewport.GotoTop()
}

// ClearConversation clears all messages
func (m *Model) ClearConversation() {
	m.Messages = []Message{}
	m.StreamingText = ""
	if m.streamingDisplay != nil {
		m.streamingDisplay.Reset()
	}
	m.updateViewport()
}

// ExportConversation exports the conversation to a string
func (m *Model) ExportConversation() string {
	var b strings.Builder
	b.WriteString("# Clem Conversation Export\n\n")
	b.WriteString(fmt.Sprintf("Model: %s\n", m.Model))
	b.WriteString(fmt.Sprintf("Conversation ID: %s\n\n", m.ConversationID))
	b.WriteString("---\n\n")

	caser := cases.Title(language.English)
	for _, msg := range m.Messages {
		b.WriteString(fmt.Sprintf("## %s\n\n", caser.String(msg.Role)))
		b.WriteString(msg.Content)
		b.WriteString("\n\n")
	}

	return b.String()
}

// SaveConversation saves the conversation (placeholder for future implementation)
func (m *Model) SaveConversation() error {
	// This would save to database or file
	// For now, it's a placeholder
	if m.db != nil {
		// Conversation is already being saved incrementally
		return nil
	}
	return nil
}

// ToggleFavorite toggles the favorite status of the current conversation
func (m *Model) ToggleFavorite() error {
	if m.db == nil {
		return fmt.Errorf("database not available")
	}

	// Toggle the local state
	m.IsFavorite = !m.IsFavorite

	// Update in database
	err := storage.SetFavorite(m.db, m.ConversationID, m.IsFavorite)
	if err != nil {
		// Revert local state on error
		m.IsFavorite = !m.IsFavorite
		return fmt.Errorf("toggle favorite: %w", err)
	}

	return nil
}

// Phase 6C Task 6: Quick Actions Methods

// EnterQuickActionsMode opens the quick actions menu
func (m *Model) EnterQuickActionsMode() {
	m.quickActionsMode = true
	m.quickActionsInput = ""
	m.quickActionsFiltered = m.quickActionsRegistry.ListActions()
}

// ExitQuickActionsMode closes the quick actions menu
func (m *Model) ExitQuickActionsMode() {
	m.quickActionsMode = false
	m.quickActionsInput = ""
	m.quickActionsFiltered = []*QuickAction{}
}

// UpdateQuickActionsInput updates the search query and filters actions
func (m *Model) UpdateQuickActionsInput(input string) {
	m.quickActionsInput = input
	m.quickActionsFiltered = m.quickActionsRegistry.FuzzySearch(input)
}

// ExecuteQuickAction executes the selected or first filtered action
func (m *Model) ExecuteQuickAction() error {
	if len(m.quickActionsFiltered) == 0 {
		return fmt.Errorf("no actions available")
	}

	// Parse command and args from input
	command, args := ParseActionCommand(m.quickActionsInput)

	// Use the command name from parsed input, or first filtered action
	actionName := command
	if actionName == "" && len(m.quickActionsFiltered) > 0 {
		actionName = m.quickActionsFiltered[0].Name
	}

	// Exit quick actions mode
	m.ExitQuickActionsMode()

	// Execute the action
	return m.quickActionsRegistry.Execute(actionName, args)
}

// dumpMessages logs all messages with their content blocks for debugging
func (m *Model) dumpMessages(label string) {
	_, _ = fmt.Fprintf(os.Stderr, "\n========== MESSAGE DUMP: %s ==========\n", label)
	for i, msg := range m.Messages {
		_, _ = fmt.Fprintf(os.Stderr, "[%d] Role: %s\n", i, msg.Role)
		if msg.Content != "" {
			_, _ = fmt.Fprintf(os.Stderr, "    Content (string): %q\n", msg.Content)
		}
		if len(msg.ContentBlock) > 0 {
			_, _ = fmt.Fprintf(os.Stderr, "    ContentBlocks (%d):\n", len(msg.ContentBlock))
			for j, block := range msg.ContentBlock {
				_, _ = fmt.Fprintf(os.Stderr, "      [%d] Type: %s\n", j, block.Type)
				if block.Text != "" {
					_, _ = fmt.Fprintf(os.Stderr, "          Text: %q\n", block.Text)
				}
				if block.ID != "" {
					_, _ = fmt.Fprintf(os.Stderr, "          ID: %s\n", block.ID)
				}
				if block.Name != "" {
					_, _ = fmt.Fprintf(os.Stderr, "          Name: %s\n", block.Name)
				}
				if block.Input != nil {
					_, _ = fmt.Fprintf(os.Stderr, "          Input: %+v\n", block.Input)
				}
				if block.ToolUseID != "" {
					_, _ = fmt.Fprintf(os.Stderr, "          ToolUseID: %s\n", block.ToolUseID)
				}
				if block.Content != "" {
					_, _ = fmt.Fprintf(os.Stderr, "          Content: %q\n", block.Content)
				}
			}
		}
	}
	_, _ = fmt.Fprintf(os.Stderr, "========================================\n\n")
}

// validateToolUseExists checks if a tool_use block with the given ID exists in message history
func (m *Model) validateToolUseExists(toolUseID string) {
	_, _ = fmt.Fprintf(os.Stderr, "[VALIDATION] Looking for tool_use with ID: %s\n", toolUseID)

	for i, msg := range m.Messages {
		if msg.Role != "assistant" {
			continue
		}
		for j, block := range msg.ContentBlock {
			if block.Type == "tool_use" && block.ID == toolUseID {
				_, _ = fmt.Fprintf(os.Stderr, "[VALIDATION] ✓ Found tool_use at message[%d].ContentBlock[%d]\n", i, j)
				return
			}
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "[VALIDATION] ✗ WARNING: tool_use with ID %s NOT FOUND in message history!\n", toolUseID)
}

// sendToolResults sends tool results back to the API and continues the conversation
func (m *Model) sendToolResults() tea.Cmd {
	if len(m.toolResults) == 0 || m.apiClient == nil {
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: no results (%d) or no client (%v)\n", len(m.toolResults), m.apiClient != nil)
		return nil
	}

	_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: preparing to send %d tool result(s)\n", len(m.toolResults))

	// Dump messages BEFORE adding tool results
	m.dumpMessages("BEFORE adding tool results")

	// Capture tool results before clearing
	results := m.toolResults
	m.toolResults = nil // Clear results

	// Build tool_result content blocks for the API
	// According to Anthropic API spec, tool results must be sent as content blocks
	// in a user message, with type="tool_result" and tool_use_id matching the original request
	toolResultBlocks := make([]core.ContentBlock, 0, len(results))
	for _, result := range results {
		content := formatToolResult(result.Result)
		toolResultBlocks = append(toolResultBlocks, core.NewToolResultBlock(result.ToolUseID, content))
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: created tool_result block for tool_use_id=%s\n", result.ToolUseID)
	}

	// Add a user message with tool_result content blocks
	// The API requires tool results to be in this specific format
	userMsg := Message{
		Role:         "user",
		ContentBlock: toolResultBlocks,
	}
	m.Messages = append(m.Messages, userMsg)
	_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: added user message with %d tool_result blocks, total messages now: %d\n", len(toolResultBlocks), len(m.Messages))

	// Dump messages AFTER adding tool results
	m.dumpMessages("AFTER adding tool results")

	// Cancel any existing stream context before creating a new one
	if m.streamCancel != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: cancelling old context\n")
		m.streamCancel()
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: no old context to cancel\n")
	}

	// Create cancellable context for this stream BEFORE the async command
	ctx, cancel := context.WithCancel(context.Background())
	// Store these in the model NOW (synchronously)
	m.streamCtx = ctx
	m.streamCancel = cancel
	_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: created new context\n")

	// Capture necessary state for the command
	apiClient := m.apiClient
	toolRegistry := m.toolRegistry
	model := m.Model
	messages := make([]Message, len(m.Messages))
	copy(messages, m.Messages)
	systemPrompt := m.systemPrompt
	_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: captured %d messages for API request\n", len(messages))

	return func() tea.Msg {
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: async command executing\n")
		// Build messages from captured state, filtering out "tool" role messages
		// (Anthropic API only accepts user/assistant/system roles)
		apiMessages := make([]core.Message, 0, len(messages))
		for _, msg := range messages {
			// Skip messages with "tool" role - they're for UI display only
			if msg.Role == "tool" {
				_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: skipping tool role message\n")
				continue
			}
			apiMessages = append(apiMessages, core.Message{
				Role:         msg.Role,
				Content:      msg.Content,
				ContentBlock: msg.ContentBlock, // Include content blocks (for tool_result blocks)
			})
		}
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: filtered to %d API messages (from %d total)\n", len(apiMessages), len(messages))

		// Get tool definitions from registry
		var tools []core.ToolDefinition
		if toolRegistry != nil {
			tools = toolRegistry.GetDefinitions()
		}

		// Create API request to continue conversation with tool results
		req := core.MessageRequest{
			Model:     model,
			Messages:  apiMessages,
			MaxTokens: 4096,
			Stream:    true,
			System:    systemPrompt,
			Tools:     tools, // Include tool definitions
		}

		// Start stream with the context we created
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: calling CreateMessageStream with %d messages\n", len(apiMessages))
		streamChan, err := apiClient.CreateMessageStream(ctx, req)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: ERROR creating stream: %v\n", err)
			return &StreamChunkMsg{Error: err}
		}

		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: stream created successfully, returning streamStartMsg\n")
		return &streamStartMsg{channel: streamChan}
	}
}

// Phase 6C Task 8: Smart Suggestion Methods

// AnalyzeSuggestions analyzes current input and updates suggestions
func (m *Model) AnalyzeSuggestions() {
	if m.suggestionDetector == nil {
		return
	}

	input := m.Input.Value()

	// Don't re-analyze if input hasn't changed
	if input == m.lastAnalyzedInput {
		return
	}

	m.lastAnalyzedInput = input

	// Get suggestions from detector
	rawSuggestions := m.suggestionDetector.AnalyzeInput(input)

	// Apply learning adjustments
	if m.suggestionLearner != nil {
		for i := range rawSuggestions {
			m.suggestionLearner.AdjustSuggestion(&rawSuggestions[i])
		}
	}

	// Convert to pointers and update model
	m.suggestions = make([]*Suggestion, len(rawSuggestions))
	for i := range rawSuggestions {
		s := rawSuggestions[i]
		m.suggestions[i] = &s
	}

	// Show suggestions if we have any
	m.showSuggestions = len(m.suggestions) > 0
}

// AcceptSuggestion accepts the top suggestion and applies it
func (m *Model) AcceptSuggestion() {
	if len(m.suggestions) == 0 {
		return
	}

	suggestion := m.suggestions[0]

	// Record acceptance
	if m.suggestionLearner != nil {
		m.suggestionLearner.RecordFeedback(
			suggestion.ToolName,
			"",
			FeedbackAccepted,
		)
	}

	// Apply the suggestion action to input
	m.Input.SetValue(suggestion.Action)

	// Clear suggestions
	m.DismissSuggestions()
}

// DismissSuggestions hides and clears suggestions
func (m *Model) DismissSuggestions() {
	// Record ignores for visible suggestions
	if m.suggestionLearner != nil && m.showSuggestions {
		for _, s := range m.suggestions {
			// FIX: Skip nil suggestions to avoid panic
			if s == nil {
				continue
			}
			m.suggestionLearner.RecordFeedback(
				s.ToolName,
				"",
				FeedbackIgnored,
			)
		}
	}

	m.showSuggestions = false
	m.suggestions = []*Suggestion{}
	m.lastAnalyzedInput = ""
}

// RejectTopSuggestion explicitly rejects the top suggestion
func (m *Model) RejectTopSuggestion() {
	if len(m.suggestions) == 0 {
		return
	}

	suggestion := m.suggestions[0]

	// Record rejection
	if m.suggestionLearner != nil {
		m.suggestionLearner.RecordFeedback(
			suggestion.ToolName,
			"",
			FeedbackRejected,
		)
	}

	// Remove the top suggestion and keep the rest
	if len(m.suggestions) > 1 {
		m.suggestions = m.suggestions[1:]
	} else {
		m.DismissSuggestions()
	}
}

// GetPendingToolUses returns the current pending tool uses for testing
func (m *Model) GetPendingToolUses() []*core.ToolUse {
	return m.pendingToolUses
}

// GetTheme returns the current theme
func (m *Model) GetTheme() themes.Theme {
	return m.theme
}
