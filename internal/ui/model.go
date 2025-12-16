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
	queuedMessage      string // Single queued message to process after current operation
	waitingForResponse bool   // True from message send until response complete - blocks new input

	// Phase 4: Service layer integration
	convSvc  services.ConversationService
	msgSvc   services.MessageService
	agentSvc services.AgentService

	// Task 12: Tool Execution UI
	toolRegistry      *tools.Registry
	toolExecutor      *tools.Executor
	assemblingToolUse *core.ToolUse // Tool being assembled from streaming chunks
	toolInputJSONBuf  string        // Buffer for accumulating input_json deltas
	approvalRules     *approval.Rules // Persistent approval rules (always/never allow)

	// Tool queue system
	activeToolQueue   *ToolQueue      // Active queue during tool processing (nil when not processing)
	toolResultHistory []ToolResult    // Historical results for UI status display (persists after queue done)
	pendingToolUses   []*core.ToolUse // Tools accumulated during streaming, transferred to queue

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
	currentToolLogName  string   // Name of currently logging tool
	currentToolLogParam string   // Parameter preview of current tool

	// TUI Polish: Overlay Management
	overlayManager       *OverlayManager       // Centralized overlay management
	baseViewportHeight   int                   // Base viewport height before overlay adjustments
	autocompleteOverlay  *AutocompleteOverlay  // Autocomplete overlay instance
	toolTimelineOverlay  *ToolTimelineOverlay  // Tool timeline overlay instance
	helpOverlay          *HelpOverlay          // Help overlay instance
	historyOverlay       *HistoryOverlay       // History overlay instance
	// Note: ToolApprovalOverlay instances are created dynamically per tool

	// TUI Polish: Message hover for timestamp display
	hoveredMessageIndex int       // Index of message being hovered (-1 = none)
	hoveredMessageTime  time.Time // Timestamp of hovered message

	// Performance: Cache most recent tool ID to avoid O(n²) scans during rendering
	mostRecentToolID string // ID of most recent tool with result (updated when tool results change)
}

// ApprovalType represents how a tool was approved or denied
type ApprovalType int

const (
	ApprovalPending     ApprovalType = iota
	ApprovalManual
	ApprovalAlwaysAllow
	DenialManual
	DenialNeverAllow
)

// ToolResult represents a tool execution result for the API
type ToolResult struct {
	ToolUseID    string
	Result       *tools.Result
	ApprovalType ApprovalType
}

// NewModel creates a new UI model
func NewModel(conversationID, model string) *Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "> "
	ta.CharLimit = 10000
	ta.SetWidth(80)
	ta.SetHeight(1)  // Start with 1 line
	ta.MaxHeight = 3 // Can grow up to 3 lines
	ta.ShowLineNumbers = false

	// Minimal styling - no borders, no cursor line highlight
	ta.FocusedStyle.Base = lipgloss.NewStyle() // No border on container
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle() // No highlight
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.AccentSky))
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SoftPaper))
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DimInk))
	ta.BlurredStyle.Base = lipgloss.NewStyle() // No border on container
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle() // No highlight
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DimInk))
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SoftPaper))
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DimInk))

	// Start with a larger default height to prevent content clipping before WindowSizeMsg
	vp := viewport.New(80, 40)
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

	// Initialize Model
	m := &Model{
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
		hoveredMessageIndex:  -1, // -1 means no message hovered
	}

	// Initialize overlay manager and overlay instances
	m.overlayManager = NewOverlayManager()
	// Note: ToolApprovalOverlay instances are created dynamically per tool via PushToolApprovalOverlays
	m.autocompleteOverlay = NewAutocompleteOverlay(m)
	m.toolTimelineOverlay = NewToolTimelineOverlay(m)
	m.helpOverlay = NewHelpOverlay()
	m.historyOverlay = NewHistoryOverlay(&m.Messages)

	return m
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	// Phase 4 Task 3: Start event subscriptions if services are available
	if m.convSvc != nil && m.msgSvc != nil {
		return tea.Batch(
			textarea.Blink,
			m.StartEventSubscriptions(),
			tea.EnableMouseAllMotion, // Enable mouse (use Shift for text selection)
		)
	}
	return tea.Batch(
		textarea.Blink,
		tea.EnableMouseAllMotion, // Enable mouse (use Shift for text selection)
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

// updateInputHeight adjusts the textarea height based on content lines
// The input grows from 1 line up to MaxHeight (3) as needed
func (m *Model) updateInputHeight() {
	value := m.Input.Value()
	if value == "" {
		m.Input.SetHeight(1)
		return
	}

	// Count visual lines based on content and wrapping
	// Use the textarea's reported Width() which is the actual wrap width
	// (don't subtract for prompt - textarea handles that internally)
	inputWidth := m.Input.Width()
	if inputWidth <= 0 {
		inputWidth = 80 // fallback
	}

	// Simple line counting: actual newlines plus wrapped lines
	lines := strings.Split(value, "\n")
	totalLines := 0
	for _, line := range lines {
		// Each line takes at least 1 row
		lineLen := len([]rune(line))
		wrappedLines := 1
		if lineLen > 0 && inputWidth > 0 {
			wrappedLines = (lineLen + inputWidth - 1) / inputWidth
			if wrappedLines < 1 {
				wrappedLines = 1
			}
		}
		totalLines += wrappedLines
	}

	// Clamp to 1-3 lines
	height := totalLines
	if height < 1 {
		height = 1
	}
	if height > 3 {
		height = 3
	}

	m.Input.SetHeight(height)
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
	m.updateInputHeight() // Reset height to 1 line after clearing

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
	m.assemblingToolUse = nil
	m.toolInputJSONBuf = ""

	// Clear tool queue state
	m.activeToolQueue = nil
	m.toolResultHistory = nil
	m.mostRecentToolID = ""

	// Clear tool log state
	m.toolLogLines = nil
	m.currentToolLogName = ""
	m.currentToolLogParam = ""

	// Clear any active overlays
	if m.overlayManager != nil {
		m.overlayManager.CancelAll()
	}

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

// SetInputValue sets the value of the input textarea (for testing)
func (m *Model) SetInputValue(value string) {
	m.Input.SetValue(value)
	m.updateInputHeight()
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

// PushToolApprovalOverlays creates and pushes one overlay per pending tool.
// Overlays are pushed in reverse order so the first tool is on top of the stack.
func (m *Model) PushToolApprovalOverlays() {
	if len(m.pendingToolUses) == 0 {
		return
	}

	// Push in reverse order so first tool ends up on top
	for i := len(m.pendingToolUses) - 1; i >= 0; i-- {
		tool := m.pendingToolUses[i]
		remaining := len(m.pendingToolUses) - 1 - i // Tools that come AFTER this one
		overlay := NewToolApprovalOverlay(m, tool, remaining)
		m.overlayManager.Push(overlay, m.Width, m.Height)
	}
	m.adjustViewportForOverlay()
}

// IsToolApprovalOverlayActive returns true if the active overlay is a tool approval
func (m *Model) IsToolApprovalOverlayActive() bool {
	active := m.overlayManager.GetActive()
	if active == nil {
		return false
	}
	_, ok := active.(*ToolApprovalOverlay)
	return ok
}

// ToolDecisionMsg is sent when user makes a decision on a tool approval overlay
type ToolDecisionMsg struct {
	Decision int // 0=approve, 1=deny, 2=always allow, 3=never allow
}

// toolQueueExecutionMsg is sent when a single tool from the queue finishes executing
type toolQueueExecutionMsg struct {
	result ToolResult
}

// ProcessNextTool processes the next tool in the queue
// Returns a command to execute the tool or show an overlay
func (m *Model) ProcessNextTool() tea.Cmd {
	if m.activeToolQueue == nil {
		return nil
	}

	item := m.activeToolQueue.Current()
	if item == nil {
		// Queue done - send results to API
		return m.finalizeToolQueue()
	}

	_, _ = fmt.Fprintf(os.Stderr, "[QUEUE] processing tool %d/%d: %s (action=%d)\n",
		m.activeToolQueue.current+1, m.activeToolQueue.Len(), item.Tool.Name, item.Action)

	switch item.Action {
	case ActionAutoApprove:
		item.Outcome = OutcomeApproved
		m.activeToolQueue.Advance()
		return m.executeQueuedTool(item.Tool, ApprovalAlwaysAllow)

	case ActionAutoDeny:
		item.Outcome = OutcomeDenied
		result := denialResult(item.Tool, DenialNeverAllow)
		m.activeToolQueue.AddResult(result)
		m.toolResultHistory = append(m.toolResultHistory, result)
		m.updateMostRecentToolID()
		m.activeToolQueue.Advance()
		// Immediately process next (no async needed for denial)
		return m.ProcessNextTool()

	case ActionNeedsApproval:
		// Push single overlay for this tool
		overlay := NewToolApprovalOverlayFromQueue(m, item)
		m.overlayManager.Push(overlay, m.Width, m.Height)
		m.adjustViewportForOverlay()
		return nil // wait for user input
	}

	return nil
}

// HandleToolDecision processes the user's decision from the approval overlay
func (m *Model) HandleToolDecision(decision int) tea.Cmd {
	if m.activeToolQueue == nil {
		return nil
	}

	item := m.activeToolQueue.Current()
	if item == nil {
		return nil
	}

	// Pop the approval overlay
	m.overlayManager.Pop()
	m.adjustViewportForOverlay()

	switch decision {
	case 0: // Approve once
		item.Outcome = OutcomeApproved
		m.activeToolQueue.Advance()
		return m.executeQueuedTool(item.Tool, ApprovalManual)

	case 1: // Deny once
		item.Outcome = OutcomeDenied
		result := denialResult(item.Tool, DenialManual)
		m.activeToolQueue.AddResult(result)
		m.toolResultHistory = append(m.toolResultHistory, result)
		m.updateMostRecentToolID()
		m.activeToolQueue.Advance()
		return m.ProcessNextTool()

	case 2: // Always allow
		if m.approvalRules != nil {
			_ = m.approvalRules.SetAlwaysAllow(item.Tool.Name)
		}
		item.Outcome = OutcomeApproved
		m.activeToolQueue.Advance()
		return m.executeQueuedTool(item.Tool, ApprovalAlwaysAllow)

	case 3: // Never allow
		if m.approvalRules != nil {
			_ = m.approvalRules.SetNeverAllow(item.Tool.Name)
		}
		item.Outcome = OutcomeDenied
		result := denialResult(item.Tool, DenialNeverAllow)
		m.activeToolQueue.AddResult(result)
		m.toolResultHistory = append(m.toolResultHistory, result)
		m.updateMostRecentToolID()
		m.activeToolQueue.Advance()
		return m.ProcessNextTool()
	}

	return nil
}

// executeQueuedTool executes a single tool from the queue and returns a message when done
func (m *Model) executeQueuedTool(tool *core.ToolUse, approvalType ApprovalType) tea.Cmd {
	_, _ = fmt.Fprintf(os.Stderr, "[QUEUE_EXEC] executing tool: %s (id=%s)\n", tool.Name, tool.ID)

	// Validate that the tool_use block exists in message history
	m.validateToolUseExists(tool.ID)

	return func() tea.Msg {
		ctx := context.Background()
		result, err := m.toolExecutor.Execute(ctx, tool.Name, tool.Input)

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[QUEUE_EXEC] tool error: %v\n", err)
			return toolQueueExecutionMsg{
				result: ToolResult{
					ToolUseID:    tool.ID,
					Result:       &tools.Result{ToolName: tool.Name, Success: false, Error: err.Error()},
					ApprovalType: approvalType,
				},
			}
		}

		_, _ = fmt.Fprintf(os.Stderr, "[QUEUE_EXEC] tool success: %s\n", tool.Name)
		return toolQueueExecutionMsg{
			result: ToolResult{
				ToolUseID:    tool.ID,
				Result:       result,
				ApprovalType: approvalType,
			},
		}
	}
}

// finalizeToolQueue sends accumulated results to API and clears the queue
func (m *Model) finalizeToolQueue() tea.Cmd {
	if m.activeToolQueue == nil {
		return nil
	}

	results := m.activeToolQueue.Results()
	_, _ = fmt.Fprintf(os.Stderr, "[QUEUE] finalizing queue with %d results\n", len(results))

	m.activeToolQueue = nil // clear ephemeral state

	if len(results) == 0 {
		return nil
	}

	return m.sendToolResultsFromQueue(results)
}

// sendToolResultsFromQueue sends tool results to the API (new queue-based version)
func (m *Model) sendToolResultsFromQueue(results []ToolResult) tea.Cmd {
	if len(results) == 0 || m.apiClient == nil {
		return nil
	}

	_, _ = fmt.Fprintf(os.Stderr, "[QUEUE] sending %d tool results to API\n", len(results))

	// Build tool_result content blocks for the API
	toolResultBlocks := make([]core.ContentBlock, 0, len(results))
	for _, result := range results {
		content := formatToolResult(result.Result)
		toolResultBlocks = append(toolResultBlocks, core.NewToolResultBlock(result.ToolUseID, content))
	}

	// Add a user message with tool_result content blocks
	userMsg := Message{
		Role:         "user",
		ContentBlock: toolResultBlocks,
	}
	m.Messages = append(m.Messages, userMsg)

	// Cancel any existing stream context before creating a new one
	m.cancelStream()

	// Create cancellable context for this stream
	ctx, cancel := context.WithCancel(context.Background())
	m.streamCtx = ctx
	m.streamCancel = cancel

	// Capture necessary state for the command
	apiClient := m.apiClient
	toolRegistry := m.toolRegistry
	model := m.Model
	apiMessages := m.GetPrunedMessages()
	systemPrompt := m.systemPrompt

	return func() tea.Msg {
		var toolDefs []core.ToolDefinition
		if toolRegistry != nil {
			toolDefs = toolRegistry.GetDefinitions()
		}

		finalSystemPrompt := core.DefaultSystemPrompt
		if systemPrompt != "" {
			finalSystemPrompt = core.DefaultSystemPrompt + "\n\n" + systemPrompt
		}

		req := core.MessageRequest{
			Model:     model,
			Messages:  apiMessages,
			MaxTokens: 8192,
			System:    finalSystemPrompt,
			Tools:     toolDefs,
			Stream:    true,
		}

		streamChan, err := apiClient.CreateMessageStream(ctx, req)
		if err != nil {
			return &StreamChunkMsg{Error: err}
		}

		return &streamStartMsg{channel: streamChan}
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
