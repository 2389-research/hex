// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Bubbletea update function for handling events
// ABOUTME: Processes keyboard input, window resize, streaming chunks
package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/pagent/internal/core"
	"github.com/harper/pagent/internal/storage"
	"github.com/harper/pagent/internal/tools"
	"github.com/harper/pagent/internal/ui/components"
)

// Update handles Bubbletea messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Reset gg sequence on any non-rune key event
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type != tea.KeyRunes {
		m.lastKeyWasG = false
	}

	switch msg := msg.(type) {
	// Task 6: Handle streaming chunks
	case *StreamChunkMsg:
		return m.handleStreamChunk(msg)

	case *streamStartMsg:
		// Store the stream channel and start reading
		m.streamChan = msg.channel
		m.SetStatus(StatusStreaming)
		m.UpdateViewport()
		return m, m.readStreamChunks(m.streamChan)

	// Task 12: Handle tool execution results
	case toolExecutionMsg:
		m.executingTool = false
		m.currentToolID = ""

		// Phase 6C: Stop spinner
		if m.spinner != nil {
			m.spinner.Stop()
		}

		if msg.err != nil {
			m.ErrorMessage = "Tool execution error: " + msg.err.Error()
			m.AddMessage("tool", "Tool error: "+msg.err.Error())
		} else {
			// Store result
			m.toolResults = append(m.toolResults, ToolResult{
				ToolUseID: msg.toolUseID,
				Result:    msg.result,
			})

			// Display result in UI
			resultMsg := formatToolResult(msg.result)
			m.AddMessage("tool", resultMsg)
		}

		m.UpdateViewport()

		// Send tool results back to API and continue conversation
		return m, m.sendToolResults()

	// Handle batch tool execution results
	case toolBatchExecutionMsg:
		m.executingTool = false
		m.executingToolUses = nil // Clear executing tools

		// Phase 6C: Stop spinner
		if m.spinner != nil {
			m.spinner.Stop()
		}

		// Store all results
		m.toolResults = append(m.toolResults, msg.results...)

		// Display results in UI
		for _, result := range msg.results {
			resultMsg := formatToolResult(result.Result)
			m.AddMessage("tool", resultMsg)
		}

		m.UpdateViewport()

		// Send ALL tool results back to API in one user message
		return m, m.sendToolResults()

	case tea.KeyMsg:
		// Phase 6C Task 6: Handle quick actions mode first
		if m.quickActionsMode {
			return m.handleQuickActionsKey(msg)
		}

		// Task 12: Handle tool approval keys
		// Phase 2: Use Huh forms for approval
		if m.toolApprovalMode && m.huhApproval != nil {
			// Let Huh form handle all input
			approvalModel, cmd := m.huhApproval.Update(msg)
			if approval, ok := approvalModel.(*components.HuhApproval); ok {
				m.huhApproval = approval

				// Check if form is complete
				if approval.IsComplete() {
					if approval.IsApproved() {
						m.ExitHuhApprovalMode()
						return m, m.ApproveToolUse()
					}
					m.ExitHuhApprovalMode()
					return m, m.DenyToolUse()
				}
			}
			return m, cmd
		}

		// Fallback to old approval mode if Huh not available
		if m.toolApprovalMode {
			switch msg.Type {
			case tea.KeyRunes:
				if len(msg.Runes) > 0 {
					r := msg.Runes[0]
					switch r {
					case 'y', 'Y', 'a', 'A':
						return m, m.ApproveToolUse()
					case 'n', 'N', 'd', 'D':
						return m, m.DenyToolUse()
					case 'v', 'V':
						// Phase 6C: Toggle approval details
						if m.approvalPrompt != nil {
							m.approvalPrompt.ToggleDetails()
						}
						return m, nil
					}
				}
			case tea.KeyEsc:
				// Esc also denies the tool
				return m, m.DenyToolUse()
			}
			// In approval mode, block other key handling
			return m, nil
		}

		// Phase 6C: Handle new keyboard shortcuts
		// Ctrl+L: Clear screen
		if msg.Type == tea.KeyCtrlL {
			m.ClearScreen()
			return m, nil
		}

		// Ctrl+K: Clear conversation
		if msg.Type == tea.KeyCtrlK {
			m.ClearConversation()
			return m, nil
		}

		// Ctrl+S: Save conversation
		if msg.Type == tea.KeyCtrlS {
			err := m.SaveConversation()
			if err != nil {
				m.ErrorMessage = "Failed to save conversation: " + err.Error()
			} else if m.statusBar != nil {
				m.statusBar.SetCustomMessage("Conversation saved!")
			}
			return m, nil
		}

		// Ctrl+E: Export conversation
		if msg.Type == tea.KeyCtrlE {
			exported := m.ExportConversation()
			// For now, just show it was exported (in future, could write to file)
			if m.statusBar != nil {
				m.statusBar.SetCustomMessage("Conversation exported (" + fmt.Sprintf("%d", len(exported)) + " bytes)")
			}
			return m, nil
		}

		// Ctrl+T: Toggle typewriter mode
		if msg.Type == tea.KeyCtrlT {
			m.ToggleTypewriter()
			if m.statusBar != nil {
				if m.typewriterMode {
					m.statusBar.SetCustomMessage("Typewriter mode ON")
				} else {
					m.statusBar.SetCustomMessage("Typewriter mode OFF")
				}
			}
			return m, nil
		}

		// Ctrl+F: Toggle favorite (only in chat mode)
		if msg.Type == tea.KeyCtrlF && m.CurrentView == ViewModeChat {
			err := m.ToggleFavorite()
			if err != nil {
				m.ErrorMessage = "Failed to toggle favorite: " + err.Error()
			} else if m.statusBar != nil {
				if m.IsFavorite {
					m.statusBar.SetCustomMessage("⭐ Added to favorites")
				} else {
					m.statusBar.SetCustomMessage("Removed from favorites")
				}
			}
			return m, nil
		}

		// Handle Esc key - dismiss suggestions, quit, or exit search mode
		if msg.Type == tea.KeyEsc {
			// Phase 6C Task 8: Dismiss suggestions first
			if m.showSuggestions {
				m.DismissSuggestions()
				return m, nil
			}
			if m.SearchMode {
				m.ExitSearchMode()
				return m, nil
			}
			if m.helpVisible {
				m.ToggleHelp()
				return m, nil
			}
			// Cancel any active stream before quitting
			if m.streamCancel != nil {
				m.streamCancel()
			}
			return m, tea.Quit
		}

		// Handle Ctrl+C to quit
		if msg.Type == tea.KeyCtrlC {
			// Cancel any active stream before quitting
			if m.streamCancel != nil {
				m.streamCancel()
			}
			return m, tea.Quit
		}

		// Phase 6C Task 4: Handle autocomplete navigation
		if m.autocomplete != nil && m.autocomplete.IsActive() {
			switch msg.Type {
			case tea.KeyTab:
				// FIX: Accept autocomplete selection with Tab
				selected := m.autocomplete.GetSelected()
				if selected != nil {
					m.Input.SetValue(selected.Value)
					m.autocomplete.Hide()
				}
				return m, nil
			case tea.KeyDown:
				m.autocomplete.Next()
				return m, nil
			case tea.KeyUp:
				m.autocomplete.Previous()
				return m, nil
			case tea.KeyEnter:
				// Accept completion
				selected := m.autocomplete.GetSelected()
				if selected != nil {
					m.Input.SetValue(selected.Value)
					m.autocomplete.Hide()
				}
				return m, nil
			case tea.KeyEsc:
				// FIX: Cancel autocomplete AND dismiss suggestions if active
				m.autocomplete.Hide()
				if m.showSuggestions {
					m.DismissSuggestions()
				}
				return m, nil
			}
		}

		// Handle Tab - accept suggestion, trigger autocomplete, or switch views
		if msg.Type == tea.KeyTab {
			// Phase 6C Task 8: Accept suggestion if visible
			if m.showSuggestions && len(m.suggestions) > 0 {
				m.AcceptSuggestion()
				return m, nil
			}

			// If textarea is focused and has content, show autocomplete
			if m.Input.Focused() && m.Input.Value() != "" {
				provider := DetectProvider(m.Input.Value())
				m.autocomplete.Show(m.Input.Value(), provider)
				return m, nil
			}
			// Otherwise, switch views
			m.NextView()
			return m, nil
		}

		// Handle Enter key
		if msg.Type == tea.KeyEnter {
			if m.SearchMode {
				// Execute search (placeholder for now)
				m.ExitSearchMode()
				return m, nil
			}
			if !msg.Alt {
				// Send message
				input := strings.TrimSpace(m.Input.Value())
				if input != "" {
					m.AddMessage("user", input)
					m.Input.Reset()
					m.UpdateViewport()

					// Task 7: Save user message to database
					if err := m.saveMessage("user", input); err != nil {
						// Log error but don't block user
						m.ErrorMessage = "Failed to save message: " + err.Error()
					}

					// Task 7: Update conversation title from first message
					if len(m.Messages) == 1 {
						title := generateConversationTitle(input)
						_ = m.updateConversationTitle(title)
					}

					// Task 6: Trigger streaming
					if m.apiClient != nil {
						return m, m.streamMessage(input)
					}
				}
			}
		}

		// Handle backspace in search mode
		if msg.Type == tea.KeyBackspace {
			if m.SearchMode && len(m.SearchQuery) > 0 {
				m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
				return m, nil
			}
		}

		// Handle rune keys for vim-like navigation and help
		// ONLY activate vim navigation when textarea is NOT focused
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && !m.Input.Focused() {
			r := msg.Runes[0]

			// Handle search mode input
			if m.SearchMode {
				m.SearchQuery += string(r)
				return m, nil
			}

			// Phase 6C: Handle '?' for help
			if r == '?' {
				m.ToggleHelp()
				return m, nil
			}

			// Phase 6C Task 6: Handle ':' for quick actions
			if r == ':' {
				m.EnterQuickActionsMode()
				return m, nil
			}

			// Vim-like navigation
			switch r {
			case '/':
				m.EnterSearchMode()
				return m, nil

			case 'j':
				// Scroll down
				m.Viewport.ScrollDown(1)
				return m, nil

			case 'k':
				// Scroll up
				m.Viewport.ScrollUp(1)
				return m, nil

			case 'g':
				// Handle 'gg' to go to top
				if m.lastKeyWasG {
					m.Viewport.GotoTop()
					m.lastKeyWasG = false
					return m, nil
				}
				m.lastKeyWasG = true
				return m, nil

			case 'G':
				// Go to bottom
				m.Viewport.GotoBottom()
				m.lastKeyWasG = false
				return m, nil

			default:
				m.lastKeyWasG = false
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Input.SetWidth(msg.Width - 4)
		m.Viewport.Width = msg.Width - 4
		m.Viewport.Height = msg.Height - 8
		if !m.Ready {
			m.Ready = true
			m.UpdateViewport()
		}
		// Phase 6C: Update component widths
		if m.statusBar != nil {
			m.statusBar.SetWidth(msg.Width)
		}
		if m.approvalPrompt != nil {
			m.approvalPrompt.SetWidth(msg.Width)
		}
	}

	// Phase 6C: Update spinner
	if m.spinner != nil {
		cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update input (only if not in search mode)
	if !m.SearchMode {
		oldValue := m.Input.Value()
		m.Input, cmd = m.Input.Update(msg)
		cmds = append(cmds, cmd)

		// Phase 6C Task 4: Update autocomplete as user types
		if m.autocomplete != nil && m.autocomplete.IsActive() {
			newValue := m.Input.Value()
			if newValue != oldValue {
				m.autocomplete.Update(newValue)
			}
		}

		// Phase 6C Task 8: Update suggestions as user types
		newValue := m.Input.Value()
		if newValue != oldValue {
			m.AnalyzeSuggestions()
		}
	}

	// Update viewport
	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// UpdateViewport renders messages into viewport
func (m *Model) UpdateViewport() {
	var content strings.Builder
	for _, msg := range m.Messages {
		if msg.Role == "user" {
			content.WriteString("You: " + msg.Content + "\n\n")
		} else {
			// Use glamour for assistant messages
			rendered, err := m.RenderMessage(msg)
			if err == nil {
				content.WriteString("Assistant:\n" + rendered + "\n")
			} else {
				content.WriteString("Assistant: " + msg.Content + "\n\n")
			}

			// Render embedded component if present (Phase 3)
			if msg.Component != nil {
				switch comp := msg.Component.(type) {
				case *components.Table:
					componentView := comp.View()
					content.WriteString("\n" + componentView + "\n\n")
				case *components.Progress:
					componentView := comp.View()
					content.WriteString(componentView + "\n\n")
				case tea.Model:
					// Generic tea.Model support
					componentView := comp.View()
					content.WriteString(componentView + "\n\n")
				}
			}
		}
	}

	// Append streaming text if present
	if m.StreamingText != "" {
		content.WriteString("\nAssistant:\n")
		// Render streaming text with glamour if available
		rendered, err := m.RenderMessage(Message{
			Role:    "assistant",
			Content: m.StreamingText,
		})
		if err == nil {
			content.WriteString(rendered)
		} else {
			content.WriteString(m.StreamingText)
		}
	}

	m.Viewport.SetContent(content.String())
	m.Viewport.GotoBottom()
}

// Task 6: Streaming Integration Functions

// handleStreamChunk processes a streaming chunk message
func (m *Model) handleStreamChunk(msg *StreamChunkMsg) (tea.Model, tea.Cmd) {
	// Handle errors
	if msg.Error != nil {
		m.ClearStreamingText()
		m.SetStatus(StatusError)
		m.ErrorMessage = msg.Error.Error()
		m.streamChan = nil
		m.streamCancel = nil
		m.streamCtx = nil
		m.UpdateViewport()
		return m, nil
	}

	chunk := msg.Chunk

	// Task 12: Handle tool_use content blocks
	if chunk.Type == "content_block_start" && chunk.ContentBlock != nil {
		if chunk.ContentBlock.Type == "tool_use" {
			// DON'T commit streaming text yet - it will be included in the same
			// assistant message as the tool_use blocks when the stream ends
			// (see lines 584-603 which creates one message with both text and tool_use blocks)

			// Start assembling tool use - Input will come in delta events
			m.assemblingToolUse = &core.ToolUse{
				Type:  "tool_use",
				ID:    chunk.ContentBlock.ID,
				Name:  chunk.ContentBlock.Name,
				Input: make(map[string]interface{}), // Will be populated when JSON is complete
			}
			m.toolInputJSONBuf = "" // Reset JSON buffer
			// Continue processing stream to get input_json_delta events
		}
	}

	// Handle input_json_delta events to build tool parameters
	if chunk.Type == "content_block_delta" && chunk.Delta != nil && m.assemblingToolUse != nil {
		if chunk.Delta.Type == "input_json_delta" && chunk.Delta.PartialJSON != "" {
			// Accumulate JSON chunks
			m.toolInputJSONBuf += chunk.Delta.PartialJSON
		}
	}

	// Handle content_block_stop - tool parameters are complete
	if chunk.Type == "content_block_stop" && m.assemblingToolUse != nil {
		// Parse accumulated JSON into Input map
		if m.toolInputJSONBuf != "" {
			var input map[string]interface{}
			if err := json.Unmarshal([]byte(m.toolInputJSONBuf), &input); err == nil {
				m.assemblingToolUse.Input = input
			}
		}

		// Tool use is complete, append to pending tools list
		// Don't handle yet - wait for message_stop to ensure full response is received
		m.pendingToolUses = append(m.pendingToolUses, m.assemblingToolUse)
		m.assemblingToolUse = nil
		m.toolInputJSONBuf = ""
		// Continue streaming to get rest of response
	}

	// Handle content deltas
	if chunk.Delta != nil && chunk.Delta.Text != "" {
		m.AppendStreamingText(chunk.Delta.Text)
		// Phase 6C: Update streaming display
		if m.streamingDisplay != nil {
			m.streamingDisplay.AppendText(chunk.Delta.Text)
		}
		m.SetStatus(StatusStreaming)
		m.UpdateViewport()
		// Continue reading from stream
		if m.streamChan != nil {
			return m, m.readStreamChunks(m.streamChan)
		}
		return m, nil
	}

	// Handle usage metadata (message_delta event)
	if chunk.Type == "message_delta" && chunk.Usage != nil {
		m.UpdateTokens(chunk.Usage.InputTokens, chunk.Usage.OutputTokens)
		// Continue reading from stream
		if m.streamChan != nil {
			return m, m.readStreamChunks(m.streamChan)
		}
		return m, nil
	}

	// Handle message completion
	if chunk.Type == "message_stop" || chunk.Done {
		// Commit streaming text, including tool_use blocks if present
		if len(m.pendingToolUses) > 0 {
			// Create assistant message with both text and ALL tool_use content blocks
			blocks := []core.ContentBlock{}

			// Add text block if there's any text content
			if m.StreamingText != "" {
				blocks = append(blocks, core.NewTextBlock(m.StreamingText))
			}

			// Add ALL tool_use blocks
			for _, toolUse := range m.pendingToolUses {
				blocks = append(blocks, core.ContentBlock{
					Type:  "tool_use",
					ID:    toolUse.ID,
					Name:  toolUse.Name,
					Input: toolUse.Input,
				})
			}

			// Add assistant message with content blocks
			assistantMsg := Message{
				Role:         "assistant",
				ContentBlock: blocks,
			}
			m.Messages = append(m.Messages, assistantMsg)
			m.StreamingText = ""

			// Show tool approval dialog
			// Phase 2: Use Huh approval instead of plain bool
			m.EnterHuhApprovalMode()
		} else {
			// No tool, just commit regular text
			m.CommitStreamingText()
		}

		m.SetStatus(StatusIdle)
		m.streamChan = nil
		m.streamCancel = nil
		m.streamCtx = nil

		m.UpdateViewport()
		return m, nil
	}

	// For other chunk types, continue reading
	if m.streamChan != nil {
		return m, m.readStreamChunks(m.streamChan)
	}

	return m, nil
}

// streamMessage starts streaming a message from the API
func (m *Model) streamMessage(_ string) tea.Cmd {
	// Build message history for API, filtering out "tool" role messages
	// (Anthropic API only accepts user/assistant/system roles)
	messages := make([]core.Message, 0, len(m.Messages))
	for _, msg := range m.Messages {
		// Skip messages with "tool" role - they're for UI display only
		if msg.Role == "tool" {
			continue
		}
		messages = append(messages, core.Message{
			Role:         msg.Role,
			Content:      msg.Content,
			ContentBlock: msg.ContentBlock, // Include content blocks (for tool_result blocks)
		})
	}

	// Get tool definitions from registry
	var tools []core.ToolDefinition
	if m.toolRegistry != nil {
		tools = m.toolRegistry.GetDefinitions()
	}

	// Create API request
	req := core.MessageRequest{
		Model:     m.Model,
		Messages:  messages,
		MaxTokens: 4096,
		Stream:    true,
		System:    m.systemPrompt, // Phase 6C: Use system prompt from template
		Tools:     tools,          // Include registered tools
	}

	// Cancel any existing stream context before creating a new one
	if m.streamCancel != nil {
		m.streamCancel()
	}

	// Create cancellable context for this stream BEFORE the async command
	ctx, cancel := context.WithCancel(context.Background())
	// Store these in the model NOW (synchronously in Update context)
	m.streamCtx = ctx
	m.streamCancel = cancel

	// Capture API client reference
	apiClient := m.apiClient

	return func() tea.Msg {
		// Start stream with the context we created
		streamChan, err := apiClient.CreateMessageStream(ctx, req)
		if err != nil {
			return &StreamChunkMsg{Error: err}
		}

		return &streamStartMsg{channel: streamChan}
	}
}

// readStreamChunks reads from the stream channel and returns messages
func (m *Model) readStreamChunks(streamChan <-chan *core.StreamChunk) tea.Cmd {
	return func() tea.Msg {
		// Defensive check for nil channel
		if streamChan == nil {
			return &StreamChunkMsg{
				Chunk: &core.StreamChunk{
					Type: "message_stop",
					Done: true,
				},
			}
		}

		// Read next chunk
		chunk, ok := <-streamChan
		if !ok {
			// Channel closed, stream is done
			return &StreamChunkMsg{
				Chunk: &core.StreamChunk{
					Type: "message_stop",
					Done: true,
				},
			}
		}

		// Return this chunk
		return &StreamChunkMsg{Chunk: chunk}
	}
}

// Task 7: Storage Integration Functions

// saveMessage saves a message to the database if DB is available
func (m *Model) saveMessage(role, content string) error {
	if m.db == nil {
		return nil // No database, skip saving
	}

	msg := &storage.Message{
		ConversationID: m.ConversationID,
		Role:           role,
		Content:        content,
	}

	return storage.CreateMessage(m.db, msg)
}

// generateConversationTitle generates a title from the first user message
func generateConversationTitle(content string) string {
	// Truncate to ~50 chars
	title := content
	if len(title) > 50 {
		title = title[:47] + "..."
	}
	// Clean up whitespace
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, "\n", " ")
	return title
}

// updateConversationTitle updates the conversation title in the database
func (m *Model) updateConversationTitle(title string) error {
	if m.db == nil {
		return nil
	}
	return storage.UpdateConversationTitle(m.db, m.ConversationID, title)
}

// Task 12: Tool result formatting

// formatToolResult formats a tool result for display
func formatToolResult(result *tools.Result) string {
	if result == nil {
		return "Tool result: (nil)"
	}

	if !result.Success {
		return "Tool " + result.ToolName + " failed: " + result.Error
	}

	output := "Tool " + result.ToolName + " succeeded"
	if result.Output != "" {
		// Truncate long outputs
		if len(result.Output) > 500 {
			output += ":\n" + result.Output[:497] + "..."
		} else {
			output += ":\n" + result.Output
		}
	}

	return output
}

// Phase 6C Task 6: Quick Actions Key Handler

// handleQuickActionsKey handles keyboard input when quick actions mode is active
func (m *Model) handleQuickActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit quick actions mode
		m.ExitQuickActionsMode()
		return m, nil

	case tea.KeyEnter:
		// Execute the selected action
		err := m.ExecuteQuickAction()
		if err != nil {
			// Show error in status bar
			m.ErrorMessage = err.Error()
			if m.statusBar != nil {
				m.statusBar.SetCustomMessage("Error: " + err.Error())
			}
		}
		return m, nil

	case tea.KeyBackspace:
		// Remove last character
		if len(m.quickActionsInput) > 0 {
			m.quickActionsInput = m.quickActionsInput[:len(m.quickActionsInput)-1]
			m.UpdateQuickActionsInput(m.quickActionsInput)
		}
		return m, nil

	case tea.KeyRunes:
		// Add typed character to input
		if len(msg.Runes) > 0 {
			m.quickActionsInput += string(msg.Runes[0])
			m.UpdateQuickActionsInput(m.quickActionsInput)
		}
		return m, nil
	}

	return m, nil
}
