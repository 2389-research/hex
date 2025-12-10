// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Bubbletea model for interactive chat UI
// ABOUTME: Manages state, messages, input, viewport, and streaming
package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/2389-research/hex/internal/approval"
	ctxmgr "github.com/2389-research/hex/internal/convcontext"
	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/pubsub"
	"github.com/2389-research/hex/internal/services"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/hex/internal/ui/forms"
	"github.com/2389-research/hex/internal/ui/theme"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ViewMode represents different view modes in the UI
type ViewMode int

const (
	// ViewModeIntro is the startup welcome screen
	ViewModeIntro ViewMode = iota
	// ViewModeChat is the main chat interface view
	ViewModeChat
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
	// StatusQueued indicates a message is queued while agent is busy
	StatusQueued
	// StatusError indicates an error occurred
	StatusError
)

// Message represents a chat message in the UI
type Message struct {
	Role          string
	Content       string
	ContentBlock  []core.ContentBlock // For structured content like tool_result blocks
	Timestamp     time.Time           // When the message was created
	renderedCache string              // Cached rendered markdown output
	cachedContent string              // Content that was cached (for invalidation)
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

// conversationEventMsg wraps a conversation event
type conversationEventMsg struct {
	event pubsub.Event[services.Conversation]
}

// messageEventMsg wraps a message event
type messageEventMsg struct {
	event pubsub.Event[services.Message]
}

// subscriptionErrorMsg indicates a subscription channel closed
type subscriptionErrorMsg struct {
	err error
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
	ShowIntro      bool // Show intro screen in viewport initially

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
	messageQueue []string // Queue of user messages to process after current stream

	// Phase 4: Service layer integration
	convSvc  services.ConversationService
	msgSvc   services.MessageService
	agentSvc services.AgentService

	// Task 12: Tool Execution UI
	toolRegistry      *tools.Registry
	toolExecutor      *tools.Executor
	pendingToolUses   []*core.ToolUse // Tools waiting for approval/execution (can be multiple)
	executingToolUses []*core.ToolUse // Tools currently being executed (for display)
	assemblingToolUse *core.ToolUse   // Tool being assembled from streaming chunks
	toolInputJSONBuf  string          // Buffer for accumulating input_json deltas
	toolApprovalMode     bool      // Showing approval prompt
	toolApprovalForm     tea.Model // Embedded huh form for tool approval (deprecated)
	selectedApprovalOpt  int       // Currently highlighted approval option (0-3)
	executingTool    bool      // Tool is running
	currentToolID     string          // ID of currently executing tool
	toolResults       []ToolResult    // Results to send back to API
	approvalRules     *approval.Rules // Persistent approval rules (always/never allow)

	// Phase 6C: Enhanced UI Components
	spinner          *ToolSpinner
	approvalPrompt   *ApprovalPrompt
	streamingDisplay *StreamingDisplay
	statusBar        *StatusBar
	helpVisible      bool
	typewriterMode   bool

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

	// TUI Polish: Theme
	theme *theme.Theme

	// Phase 4 Task 3: Event subscriptions
	eventCtx    context.Context
	eventCancel context.CancelFunc

	// Performance: Throttled viewport updates (60fps max)
	lastViewportUpdate time.Time // Last time viewport was updated

	// Robustness: Re-entrance guards
	processingWindowSize bool // Prevent re-entrance in WindowSizeMsg handler

	// TUI Polish: Quit confirmation
	pendingQuit     bool      // First Ctrl+C pressed, waiting for confirmation
	pendingQuitTime time.Time // When first Ctrl+C was pressed

	// TUI Polish: Input history navigation
	inputHistory      []string // History of user inputs
	inputHistoryIndex int      // Current position in history (-1 = current input, 0 = most recent)
	inputHistorySaved string   // Saved current input when navigating history

	// TUI Polish: Tool output log
	toolLogLines       []string // Accumulated output lines for current chunk
	toolLogOverlay     bool     // Whether overlay is visible
	currentToolLogName  string   // Name of currently logging tool
	currentToolLogParam string   // Parameter preview of current tool
}

// ToolResult represents a tool execution result for the API
type ToolResult struct {
	ToolUseID string
	Result    *tools.Result
}

// NewModel creates a new UI model
func NewModel(conversationID, model string) *Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 10000
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	// Apply Neo-Terminal theme colors for sophisticated aesthetics
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color(theme.Ghost))
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.AccentSky))
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SoftPaper))
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DimInk))
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DimInk))
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SoftPaper))
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DimInk))

	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to Hex! Type your message below.")

	// Initialize glamour renderer for markdown
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		renderer = nil // Explicit - RenderMessage already checks for nil
	}

	// Phase 6C: Initialize new UI components
	spinner := NewToolSpinner()
	streamingDisplay := NewStreamingDisplay()
	statusBar := NewStatusBar(model, 80)
	quickActionsRegistry := NewQuickActionsRegistry()
	autocomplete := NewAutocomplete()

	// Phase 6C Task 8: Initialize suggestion system
	suggestionDetector := NewSuggestionDetector()
	suggestionLearner := NewSuggestionLearner()

	// Tool approval: Initialize approval rules
	approvalRules, err := approval.NewRules()
	if err != nil {
		// Log error but don't fail - we can still function without persistent rules
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to load approval rules: %v\n", err)
		approvalRules = nil
	}

	// TUI Polish: Initialize Neo-Terminal theme
	neoTerminalTheme := theme.NeoTerminalTheme()

	return &Model{
		ConversationID:       conversationID,
		Model:                model,
		Messages:             []Message{},
		Input:                ta,
		Viewport:             vp,
		Width:                80,
		Height:               24,
		ShowIntro:            true,         // Show intro in viewport initially
		CurrentView:          ViewModeChat, // Start in chat mode - no key press needed
		Status:               StatusIdle,
		renderer:             renderer,
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
		theme:                neoTerminalTheme,
		approvalRules:        approvalRules,
		inputHistory:         []string{},
		inputHistoryIndex:    -1, // -1 means at current input, not browsing history
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	// Phase 4 Task 3: Start event subscriptions if services are available
	if m.convSvc != nil && m.msgSvc != nil {
		return tea.Batch(
			textarea.Blink,
			m.StartEventSubscriptions(),
			tea.EnableMouseCellMotion, // Enable mouse wheel scrolling
		)
	}
	return tea.Batch(
		textarea.Blink,
		tea.EnableMouseCellMotion, // Enable mouse wheel scrolling
	)
}

// AddMessage adds a message to the conversation
func (m *Model) AddMessage(role, content string) {
	// Never add messages with empty content
	if strings.TrimSpace(content) == "" {
		return
	}

	m.Messages = append(m.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	// Update context usage after adding message
	m.updateContextUsage()
}

// NextView cycles to the next view mode
func (m *Model) NextView() {
	switch m.CurrentView {
	case ViewModeIntro:
		m.CurrentView = ViewModeChat
	case ViewModeChat:
		m.CurrentView = ViewModeHistory
	case ViewModeHistory:
		m.CurrentView = ViewModeTools
	case ViewModeTools:
		m.CurrentView = ViewModeIntro
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
	// Note: Caller should set ErrorMessage explicitly when using StatusError
	// This allows for specific error messages rather than generic ones
}

// RenderMessage renders a message using glamour for assistant messages with caching
func (m *Model) RenderMessage(msg *Message) (string, error) {
	// Performance: Use cached render if available
	if msg.renderedCache != "" && msg.cachedContent == msg.Content {
		return msg.renderedCache, nil
	}

	content := msg.Content

	// Constrain long code blocks to prevent flooding
	content = m.constrainLongCodeBlocks(content)

	if msg.Role == "assistant" && m.renderer != nil {
		rendered, err := m.renderer.Render(content)
		if err != nil {
			return content, err
		}
		// Remove glamour's paragraph indentation (leading 2 spaces on each line)
		rendered = removeGlamourIndent(rendered)

		// Cache the rendered output
		msg.renderedCache = rendered
		msg.cachedContent = content

		return rendered, nil
	}

	// Cache non-assistant messages too (they're already "rendered")
	msg.renderedCache = content
	msg.cachedContent = content

	return content, nil
}

// removeGlamourIndent strips the 2-space paragraph indentation that glamour adds
func removeGlamourIndent(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Remove up to 2 leading spaces (glamour's paragraph indent)
		if strings.HasPrefix(line, "  ") {
			lines[i] = line[2:]
		}
	}
	return strings.Join(lines, "\n")
}

// constrainLongCodeBlocks truncates code blocks that exceed a reasonable length
func (m *Model) constrainLongCodeBlocks(content string) string {
	const maxCodeBlockLines = 30 // Show first 30 lines of code blocks
	const contextLines = 5       // Show first and last N lines

	// Simple regex-like approach: find code blocks between ``` markers
	lines := strings.Split(content, "\n")
	var result strings.Builder
	inCodeBlock := false
	codeBlockLines := []string{}
	codeBlockLang := ""

	for i, line := range lines {
		// Check if this line starts/ends a code block
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if !inCodeBlock {
				// Starting a code block
				inCodeBlock = true
				codeBlockLang = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "```"))
				codeBlockLines = []string{}
				continue
			}
			// Ending a code block
			inCodeBlock = false

			// Check if code block is too long
			if len(codeBlockLines) > maxCodeBlockLines {
				// Write truncated version
				result.WriteString("```" + codeBlockLang + "\n")

				// Show first contextLines
				for j := 0; j < contextLines && j < len(codeBlockLines); j++ {
					result.WriteString(codeBlockLines[j] + "\n")
				}

				// Show truncation indicator
				omittedLines := len(codeBlockLines) - (contextLines * 2)
				result.WriteString(m.theme.Warning.Render(fmt.Sprintf("\n... (%d lines omitted) ...\n\n", omittedLines)))

				// Show last contextLines
				start := len(codeBlockLines) - contextLines
				if start < contextLines {
					start = contextLines
				}
				for j := start; j < len(codeBlockLines); j++ {
					result.WriteString(codeBlockLines[j] + "\n")
				}

				result.WriteString("```\n")
			} else {
				// Write full code block
				result.WriteString("```" + codeBlockLang + "\n")
				for _, codeLine := range codeBlockLines {
					result.WriteString(codeLine + "\n")
				}
				result.WriteString("```\n")
			}

			codeBlockLines = []string{}
			codeBlockLang = ""
			continue
		}

		if inCodeBlock {
			codeBlockLines = append(codeBlockLines, line)
		} else {
			result.WriteString(line)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
		}
	}

	// Handle unclosed code block
	if inCodeBlock && len(codeBlockLines) > 0 {
		result.WriteString("```" + codeBlockLang + "\n")
		if len(codeBlockLines) > maxCodeBlockLines {
			for j := 0; j < contextLines && j < len(codeBlockLines); j++ {
				result.WriteString(codeBlockLines[j] + "\n")
			}
			omittedLines := len(codeBlockLines) - (contextLines * 2)
			result.WriteString(m.theme.Warning.Render(fmt.Sprintf("\n... (%d lines omitted) ...\n\n", omittedLines)))
			start := len(codeBlockLines) - contextLines
			if start < contextLines {
				start = contextLines
			}
			for j := start; j < len(codeBlockLines); j++ {
				result.WriteString(codeBlockLines[j] + "\n")
			}
		} else {
			for _, codeLine := range codeBlockLines {
				result.WriteString(codeLine + "\n")
			}
		}
		result.WriteString("```\n")
	}

	return result.String()
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

// AppendStreamingText adds a chunk to the streaming buffer and updates the last message in place
func (m *Model) AppendStreamingText(chunk string) {
	m.StreamingText += chunk

	// Update the last message in place (should be the assistant message placeholder)
	if len(m.Messages) > 0 && m.Messages[len(m.Messages)-1].Role == "assistant" {
		m.Messages[len(m.Messages)-1].Content = m.StreamingText
		// Invalidate render cache when content changes
		m.Messages[len(m.Messages)-1].renderedCache = ""
	}
}

// CommitStreamingText finalizes the streaming message
// The message is already in m.Messages (added as placeholder at stream start),
// so we just need to clear the streaming buffer
func (m *Model) CommitStreamingText() {
	// The assistant message was already added when streaming started
	// and has been updated in place during streaming.
	// Just clear the streaming buffer.
	m.StreamingText = ""

	// NOTE: We do NOT clear the tool log chunk here. The chunk persists so the user
	// can view it with Ctrl+O even after the assistant responds. The chunk is cleared
	// when the next tool use begins (in ApproveToolUse).
}

// ClearStreamingText discards streaming buffer (e.g., on error)
func (m *Model) ClearStreamingText() {
	m.StreamingText = ""
}

// cancelStream safely cancels the active stream and clears all stream-related state
// This prevents memory leaks from unclosed contexts and goroutine leaks
func (m *Model) cancelStream() {
	if m.streamCancel != nil {
		m.streamCancel()
	}
	m.streamCancel = nil
	m.streamCtx = nil
	m.streamChan = nil
}

// ClearContext resets the conversation context and UI state (for /clear command)
func (m *Model) ClearContext() {
	// Cancel any active stream
	m.cancelStream()

	// Clear all messages and streaming state
	m.Messages = []Message{}
	m.StreamingText = ""
	m.Streaming = false

	// Reset input
	m.Input.Reset()

	// Reset status and errors
	m.Status = StatusIdle
	m.ErrorMessage = ""

	// Reset token counters
	m.TokensInput = 0
	m.TokensOutput = 0

	// Reset view mode to chat
	m.CurrentView = ViewModeChat

	// Exit search mode if active
	m.SearchMode = false
	m.SearchQuery = ""

	// Reset vim navigation state
	m.lastKeyWasG = false

	// Clear tool execution state
	m.pendingToolUses = nil
	m.executingToolUses = nil
	m.assemblingToolUse = nil
	m.toolInputJSONBuf = ""
	m.toolApprovalMode = false
	m.toolApprovalForm = nil
	m.executingTool = false
	m.currentToolID = ""
	m.toolResults = nil

	// Clear streaming display
	if m.streamingDisplay != nil {
		m.streamingDisplay.Reset()
	}

	// Stop spinner if active
	if m.spinner != nil {
		m.spinner.Stop()
	}

	// Reset help and UI modes
	m.helpVisible = false
	m.typewriterMode = false

	// Clear quick actions state
	m.quickActionsMode = false
	m.quickActionsInput = ""
	m.quickActionsFiltered = nil

	// Clear suggestions state
	m.showSuggestions = false
	m.suggestions = nil
	m.lastAnalyzedInput = ""

	// Clear tool log state
	m.clearToolLogChunk()
	m.toolLogOverlay = false

	// Hide autocomplete if active
	if m.autocomplete != nil {
		m.autocomplete.Hide()
	}

	// Reset context usage tracking
	m.contextUsage = ctxmgr.ContextUsage{}

	// Show intro screen
	m.ShowIntro = true

	// Update viewport to show cleared state
	m.updateViewport()
}

// SetAPIClient sets the API client for streaming
func (m *Model) SetAPIClient(client *core.Client) {
	m.apiClient = client
}

// SetServices sets the service layer dependencies
func (m *Model) SetServices(convSvc services.ConversationService, msgSvc services.MessageService, agentSvc services.AgentService) {
	m.convSvc = convSvc
	m.msgSvc = msgSvc
	m.agentSvc = agentSvc
}

// StartEventSubscriptions initializes event subscriptions and returns commands to listen for events
func (m *Model) StartEventSubscriptions() tea.Cmd {
	// Create context for subscriptions
	m.eventCtx, m.eventCancel = context.WithCancel(context.Background())

	// Subscribe to conversation events
	convEvents := m.convSvc.Subscribe(m.eventCtx)

	// Subscribe to message events
	msgEvents := m.msgSvc.Subscribe(m.eventCtx)

	// Return Bubbletea commands that listen to both channels
	return tea.Batch(
		waitForConversationEvent(convEvents),
		waitForMessageEvent(msgEvents),
	)
}

// waitForConversationEvent waits for the next conversation event and converts it to a tea.Msg
func waitForConversationEvent(ch <-chan pubsub.Event[services.Conversation]) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			// Channel closed - return error instead of silent nil
			// This provides visibility when subscriptions fail
			return subscriptionErrorMsg{err: fmt.Errorf("conversation event subscription closed")}
		}
		return conversationEventMsg{event: event}
	}
}

// waitForMessageEvent waits for the next message event and converts it to a tea.Msg
func waitForMessageEvent(ch <-chan pubsub.Event[services.Message]) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			// Channel closed - return error instead of silent nil
			// This provides visibility when subscriptions fail
			return subscriptionErrorMsg{err: fmt.Errorf("message event subscription closed")}
		}
		return messageEventMsg{event: event}
	}
}

// Task 12: Tool System Methods

// SetToolSystem sets the tool registry and executor
func (m *Model) SetToolSystem(registry *tools.Registry, executor *tools.Executor) {
	m.toolRegistry = registry
	m.toolExecutor = executor
	// Note: ToolProvider removed from autocomplete - internal tools are not user-facing
}

// SetSlashCommands sets the available slash commands for autocomplete
func (m *Model) SetSlashCommands(commands []string, descriptions map[string]string) {
	if m.autocomplete == nil {
		return
	}

	provider, ok := m.autocomplete.GetProvider("command")
	if ok {
		if cmdProvider, ok := provider.(*SlashCommandProvider); ok {
			cmdProvider.SetCommands(commands, descriptions)
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
// Filters out "tool" role messages (for UI display only) and includes ContentBlock for tool results
func (m *Model) GetPrunedMessages() []core.Message {
	// Filter and convert to core messages
	// Skip "tool" role messages - they're for UI display only (Anthropic API only accepts user/assistant/system)
	coreMessages := make([]core.Message, 0, len(m.Messages))
	for _, msg := range m.Messages {
		if msg.Role == "tool" {
			continue
		}
		coreMessages = append(coreMessages, core.Message{
			Role:         msg.Role,
			Content:      msg.Content,
			ContentBlock: msg.ContentBlock, // Include content blocks (for tool_result blocks)
		})
	}

	// If no context manager, return all messages
	if m.contextManager == nil {
		return coreMessages
	}

	// Prune if needed
	if m.contextManager.ShouldPrune(coreMessages) {
		_, _ = fmt.Fprintf(os.Stderr, "[CONTEXT] Pruning messages: %d → %d\n", len(coreMessages), m.contextManager.MaxTokens)
		return m.contextManager.Prune(coreMessages)
	}

	return coreMessages
}

// ApproveToolUse executes ALL pending tools
func (m *Model) ApproveToolUse() tea.Cmd {
	// Guard against double-execution from button mashing
	if m.executingTool {
		return nil
	}

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
	m.toolApprovalForm = nil
	m.approvalPrompt = nil // Clear approval prompt for next time
	m.executingTool = true

	// Clear previous tool log chunk - output will be added when results come back
	// This preserves the previous chunk until a new tool use begins
	m.clearToolLogChunk()

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
	m.toolApprovalForm = nil
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
	b.WriteString("# Hex Conversation Export\n\n")
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
	// Conversation is already being saved incrementally via service layer
	return nil
}

// ToggleFavorite toggles the favorite status of the current conversation
func (m *Model) ToggleFavorite() error {
	if m.convSvc == nil {
		return fmt.Errorf("conversation service not available")
	}

	// Toggle the local state
	m.IsFavorite = !m.IsFavorite

	// Get current conversation
	conv, err := m.convSvc.Get(context.Background(), m.ConversationID)
	if err != nil {
		// Revert local state on error
		m.IsFavorite = !m.IsFavorite
		return fmt.Errorf("get conversation: %w", err)
	}

	// Update favorite status
	conv.IsFavorite = m.IsFavorite
	err = m.convSvc.Update(context.Background(), conv)
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
	// Performance: Only dump messages when HEX_DEBUG is set
	if os.Getenv("HEX_DEBUG") == "" {
		return
	}

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
	_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: cancelling old context\n")
	m.cancelStream()

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
	// Get pruned messages (automatically compacts if near context limit)
	apiMessages := m.GetPrunedMessages()
	systemPrompt := m.systemPrompt
	_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: captured %d pruned messages for API request\n", len(apiMessages))

	return func() tea.Msg {
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: async command executing with %d messages\n", len(apiMessages))

		// Get tool definitions from registry
		var tools []core.ToolDefinition
		if toolRegistry != nil {
			tools = toolRegistry.GetDefinitions()
		}

		// Always include Hex identity in system prompt
		finalSystemPrompt := core.DefaultSystemPrompt
		if systemPrompt != "" {
			finalSystemPrompt = core.DefaultSystemPrompt + "\n\n" + systemPrompt
		}

		// Create API request to continue conversation with tool results
		req := core.MessageRequest{
			Model:     model,
			Messages:  apiMessages,
			MaxTokens: 4096,
			Stream:    true,
			System:    finalSystemPrompt,
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

// handleApprovalResult processes the result from the huh approval form
func (m *Model) handleApprovalResult(msg *forms.ApprovalResultMsg) (tea.Model, tea.Cmd) {
	// Check for errors
	if msg.Error != nil {
		m.ErrorMessage = "Approval form error: " + msg.Error.Error()
		m.toolApprovalMode = false
		m.toolApprovalForm = nil
		return m, m.DenyToolUse()
	}

	// Exit approval mode
	m.toolApprovalMode = false
	m.toolApprovalForm = nil
	m.approvalPrompt = nil

	// Process the decision
	switch msg.Result.Decision {
	case forms.DecisionApprove:
		_, _ = fmt.Fprintf(os.Stderr, "[APPROVAL_FORM] User approved tool: %s\n", msg.Result.ToolUse.Name)
		return m, m.ApproveToolUse()

	case forms.DecisionDeny:
		_, _ = fmt.Fprintf(os.Stderr, "[APPROVAL_FORM] User denied tool: %s\n", msg.Result.ToolUse.Name)
		return m, m.DenyToolUse()

	case forms.DecisionAlwaysAllow:
		_, _ = fmt.Fprintf(os.Stderr, "[APPROVAL_FORM] User always allowed tool: %s\n", msg.Result.ToolUse.Name)
		// Store in persistent approval rules
		if m.approvalRules != nil {
			if err := m.approvalRules.SetAlwaysAllow(msg.Result.ToolUse.Name); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[APPROVAL_RULES] Failed to persist always-allow rule: %v\n", err)
			}
		}
		return m, m.ApproveToolUse()

	case forms.DecisionNeverAllow:
		_, _ = fmt.Fprintf(os.Stderr, "[APPROVAL_FORM] User never allowed tool: %s\n", msg.Result.ToolUse.Name)
		// Store in persistent approval rules
		if m.approvalRules != nil {
			if err := m.approvalRules.SetNeverAllow(msg.Result.ToolUse.Name); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[APPROVAL_RULES] Failed to persist never-allow rule: %v\n", err)
			}
		}
		return m, m.DenyToolUse()

	default:
		// Unknown decision, deny by default
		_, _ = fmt.Fprintf(os.Stderr, "[APPROVAL_FORM] Unknown decision, denying tool\n")
		return m, m.DenyToolUse()
	}
}

// LaunchQuickActionsForm launches the quick actions form using huh
func (m *Model) LaunchQuickActionsForm() tea.Cmd {
	// Convert registry actions to form actions
	registryActions := m.quickActionsRegistry.ListActions()
	formActions := make([]*forms.QuickAction, 0, len(registryActions))

	for _, action := range registryActions {
		// Determine category based on action name
		category := string(forms.CategoryTools)
		switch action.Name {
		case "help", "back", "forward", "quit":
			category = string(forms.CategoryNavigation)
		case "save", "export", "clear", "reset":
			category = string(forms.CategorySettings)
		}

		formActions = append(formActions, &forms.QuickAction{
			Name:        action.Name,
			Description: action.Description,
			Category:    category,
			KeyBinding:  "", // Old actions don't have explicit key bindings
			Handler:     action.Handler,
		})
	}

	// Set quick actions mode flag
	m.quickActionsMode = true

	// Run the form asynchronously
	return forms.RunQuickActionsFormAsync(formActions)
}

// handleQuickActionsResult processes the result from the huh quick actions form
func (m *Model) handleQuickActionsResult(msg *forms.QuickActionsResultMsg) tea.Model {
	// Exit quick actions mode
	m.quickActionsMode = false

	// Check for errors
	if msg.Error != nil {
		m.ErrorMessage = "Quick actions error: " + msg.Error.Error()
		if m.statusBar != nil {
			m.statusBar.SetCustomMessage("Error: " + msg.Error.Error())
		}
		return m
	}

	// Execute the selected action
	_, _ = fmt.Fprintf(os.Stderr, "[QUICK_ACTIONS] User selected action: %s\n", msg.ActionName)

	action, err := m.quickActionsRegistry.GetAction(msg.ActionName)
	if err != nil {
		m.ErrorMessage = "Action not found: " + msg.ActionName
		if m.statusBar != nil {
			m.statusBar.SetCustomMessage("Error: action not found")
		}
		return m
	}

	// Execute the action handler
	err = action.Handler("")
	if err != nil {
		m.ErrorMessage = "Action failed: " + err.Error()
		if m.statusBar != nil {
			m.statusBar.SetCustomMessage("Error: " + err.Error())
		}
		return m
	}

	if m.statusBar != nil {
		m.statusBar.SetCustomMessage("Executed: " + action.Name)
	}

	return m
}

// TUI Polish: Input History Navigation Methods

// addToInputHistory adds an input to the history (avoiding duplicates of last entry)
func (m *Model) addToInputHistory(input string) {
	// Don't add empty inputs
	if strings.TrimSpace(input) == "" {
		return
	}

	// Don't add if it's the same as the last entry
	if len(m.inputHistory) > 0 && m.inputHistory[len(m.inputHistory)-1] == input {
		return
	}

	// Add to history (most recent at the end)
	m.inputHistory = append(m.inputHistory, input)

	// Limit history size to 100 entries
	const maxHistorySize = 100
	if len(m.inputHistory) > maxHistorySize {
		m.inputHistory = m.inputHistory[len(m.inputHistory)-maxHistorySize:]
	}

	// Reset history navigation state
	m.inputHistoryIndex = -1
	m.inputHistorySaved = ""
}

// navigateHistoryUp moves to an older history entry (returns true if handled)
func (m *Model) navigateHistoryUp() bool {
	if len(m.inputHistory) == 0 {
		return false
	}

	// Save current input when starting to navigate
	if m.inputHistoryIndex == -1 {
		m.inputHistorySaved = m.Input.Value()
	}

	// Move to older entry
	newIndex := m.inputHistoryIndex + 1
	if newIndex >= len(m.inputHistory) {
		// Already at oldest entry
		return true // Still handled, just don't move further
	}

	m.inputHistoryIndex = newIndex

	// History is stored oldest-first, so we index from the end
	historyIdx := len(m.inputHistory) - 1 - m.inputHistoryIndex
	m.Input.SetValue(m.inputHistory[historyIdx])

	// Move cursor to end of input
	m.Input.CursorEnd()

	return true
}

// navigateHistoryDown moves to a newer history entry (returns true if handled)
func (m *Model) navigateHistoryDown() bool {
	if m.inputHistoryIndex == -1 {
		// Not browsing history
		return false
	}

	// Move to newer entry
	newIndex := m.inputHistoryIndex - 1

	if newIndex < 0 {
		// Restore saved input
		m.inputHistoryIndex = -1
		m.Input.SetValue(m.inputHistorySaved)
		m.inputHistorySaved = ""
		m.Input.CursorEnd()
		return true
	}

	m.inputHistoryIndex = newIndex

	// History is stored oldest-first, so we index from the end
	historyIdx := len(m.inputHistory) - 1 - m.inputHistoryIndex
	m.Input.SetValue(m.inputHistory[historyIdx])

	// Move cursor to end of input
	m.Input.CursorEnd()

	return true
}
