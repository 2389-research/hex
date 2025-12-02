// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Bubbletea update function for handling events
// ABOUTME: Processes keyboard input, window resize, streaming chunks
package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/storage"
	"github.com/harper/clem/internal/tools"
	"github.com/harper/clem/internal/ui/forms"
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
		m.updateViewport()
		return m, m.readStreamChunks(m.streamChan)

	// Task 12: Handle tool execution results
	case toolExecutionMsg:
		_, _ = fmt.Fprintf(os.Stderr, "[TOOL_RESULT_RECEIVED] tool_use_id=%s, err=%v\n", msg.toolUseID, msg.err)

		m.executingTool = false
		m.currentToolID = ""

		// Phase 6C: Stop spinner
		if m.spinner != nil {
			m.spinner.Stop()
		}

		if msg.err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[TOOL_RESULT_ERROR] tool_use_id=%s, error=%s\n", msg.toolUseID, msg.err.Error())
			m.ErrorMessage = "Tool execution error: " + msg.err.Error()
			m.AddMessage("tool", "Tool error: "+msg.err.Error())
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "[TOOL_RESULT_SUCCESS] tool_use_id=%s, storing result\n", msg.toolUseID)

			// Validate that tool_use exists in message history before storing result
			m.validateToolUseExists(msg.toolUseID)

			// Store result
			m.toolResults = append(m.toolResults, ToolResult{
				ToolUseID: msg.toolUseID,
				Result:    msg.result,
			})

			_, _ = fmt.Fprintf(os.Stderr, "[TOOL_RESULTS_QUEUE] current queue length: %d\n", len(m.toolResults))

			// Display result in UI
			resultMsg := formatToolResult(msg.result)
			m.AddMessage("tool", resultMsg)
		}

		m.updateViewport()

		// Send tool results back to API and continue conversation
		_, _ = fmt.Fprintf(os.Stderr, "[TOOL_RESULTS_SENDING] about to send %d tool results back to API\n", len(m.toolResults))
		return m, m.sendToolResults()

	// Handle batch tool execution results
	case toolBatchExecutionMsg:
		_, _ = fmt.Fprintf(os.Stderr, "[BATCH_RESULT_RECEIVED] received results for %d tool(s)\n", len(msg.results))

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

		m.updateViewport()

		// Send ALL tool results back to API in one user message
		_, _ = fmt.Fprintf(os.Stderr, "[BATCH_RESULTS_SENDING] sending %d tool results back to API\n", len(m.toolResults))
		return m, m.sendToolResults()

	// Handle approval form results from huh
	case *forms.ApprovalResultMsg:
		return m.handleApprovalResult(msg)

	// Handle quick actions form results from huh
	case *forms.QuickActionsResultMsg:
		return m.handleQuickActionsResult(msg)

	case tea.KeyMsg:
		// Intro screen: Any key dismisses it
		if m.CurrentView == ViewModeIntro {
			m.CurrentView = ViewModeChat
			return m, nil
		}

		// Task 12: Tool approval is now handled by huh forms
		// No need to handle keys in approval mode - huh handles its own input
		if m.toolApprovalMode {
			// In approval mode, block other key handling (form is running)
			return m, nil
		}

		// TUI Polish Task 4: Quick actions mode is now handled by huh forms
		// No need to handle keys in quick actions mode - huh handles its own input
		if m.quickActionsMode {
			// In quick actions mode, block other key handling (form is running)
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
			// Only send message on plain Enter
			// Alt+Enter is passed to textarea for multi-line input
			if !msg.Alt {
				// Send message
				input := strings.TrimSpace(m.Input.Value())
				if input != "" {
					m.AddMessage("user", input)
					m.Input.Reset()
					m.updateViewport()

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

			// TUI Polish Task 4: Handle ':' for quick actions (now using huh forms)
			if r == ':' {
				return m, m.LaunchQuickActionsForm()
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
			m.updateViewport()
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

// updateViewport renders messages into viewport
func (m *Model) updateViewport() {
	var content strings.Builder
	for _, msg := range m.Messages {
		if msg.Role == "user" {
			// Style user messages with cyan (Dracula theme)
			styledContent := m.theme.UserMessage.Render(msg.Content)
			content.WriteString(styledContent + "\n\n")
		} else {
			// Use glamour for assistant messages (no label, just content)
			// Note: glamour.Render() adds "\n" at start and "\n\n" at end
			// We strip the leading "\n" to maintain consistent "\n\n" spacing between messages
			rendered, err := m.RenderMessage(msg)
			if err == nil {
				content.WriteString(strings.TrimPrefix(rendered, "\n"))
			} else {
				content.WriteString(msg.Content + "\n\n")
			}
		}
	}

	// Append streaming text if present
	if m.StreamingText != "" {
		// Render streaming text with glamour if available
		// Note: glamour.Render() adds "\n" at start and "\n\n" at end
		// We strip the leading "\n" to maintain consistent "\n\n" spacing
		rendered, err := m.RenderMessage(Message{
			Role:    "assistant",
			Content: m.StreamingText,
		})
		if err == nil {
			content.WriteString(strings.TrimPrefix(rendered, "\n"))
		} else {
			content.WriteString(m.StreamingText + "\n\n")
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
		m.updateViewport()
		return m, nil
	}

	chunk := msg.Chunk

	// Task 12: Handle tool_use content blocks
	if chunk.Type == "content_block_start" && chunk.ContentBlock != nil {
		if chunk.ContentBlock.Type == "tool_use" {
			_, _ = fmt.Fprintf(os.Stderr, "[STREAM_TOOL_START] tool_use detected: id=%s, name=%s\n", chunk.ContentBlock.ID, chunk.ContentBlock.Name)

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
				_, _ = fmt.Fprintf(os.Stderr, "[STREAM_TOOL_INPUT] parsed input for tool_use_id=%s, input=%+v\n", m.assemblingToolUse.ID, input)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "[STREAM_TOOL_INPUT_ERROR] failed to parse input JSON for tool_use_id=%s: %v\n", m.assemblingToolUse.ID, err)
			}
		}

		// Tool use is complete, append to pending tools list
		// Don't handle yet - wait for message_stop to ensure full response is received
		m.pendingToolUses = append(m.pendingToolUses, m.assemblingToolUse)
		fmt.Fprintf(os.Stderr, "[STREAM_TOOL_COMPLETE] tool_use complete, added to pending (total pending: %d): id=%s, name=%s\n",
			len(m.pendingToolUses), m.assemblingToolUse.ID, m.assemblingToolUse.Name)
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
		m.updateViewport()
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
		_, _ = fmt.Fprintf(os.Stderr, "[STREAM_STOP] message stream ended, pendingToolUses count=%d\n", len(m.pendingToolUses))

		// Commit streaming text, including tool_use blocks if present
		if len(m.pendingToolUses) > 0 {
			_, _ = fmt.Fprintf(os.Stderr, "[STREAM_STOP_WITH_TOOLS] creating assistant message with %d tool_use block(s)\n", len(m.pendingToolUses))

			// Create assistant message with both text and ALL tool_use content blocks
			blocks := []core.ContentBlock{}

			// Add text block if there's any text content
			if m.StreamingText != "" {
				blocks = append(blocks, core.NewTextBlock(m.StreamingText))
				_, _ = fmt.Fprintf(os.Stderr, "[STREAM_STOP_WITH_TOOLS] including text block (%d chars)\n", len(m.StreamingText))
			}

			// Add ALL tool_use blocks
			for i, toolUse := range m.pendingToolUses {
				blocks = append(blocks, core.ContentBlock{
					Type:  "tool_use",
					ID:    toolUse.ID,
					Name:  toolUse.Name,
					Input: toolUse.Input,
				})
				fmt.Fprintf(os.Stderr, "[STREAM_STOP_WITH_TOOLS] added tool_use block %d/%d: id=%s, name=%s\n",
					i+1, len(m.pendingToolUses), toolUse.ID, toolUse.Name)
			}

			// Add assistant message with content blocks
			assistantMsg := Message{
				Role:         "assistant",
				ContentBlock: blocks,
			}
			m.Messages = append(m.Messages, assistantMsg)
			m.StreamingText = ""

			_, _ = fmt.Fprintf(os.Stderr, "[STREAM_STOP_WITH_TOOLS] assistant message added to history (total messages: %d)\n", len(m.Messages))

			// Dump messages after adding assistant message with tool_use blocks
			m.dumpMessages("AFTER stream completion with tool_use blocks")

			// Show tool approval dialog using huh form
			_, _ = fmt.Fprintf(os.Stderr, "[STREAM_STOP_WITH_TOOLS] launching huh approval form\n")
			m.toolApprovalMode = true
			m.updateViewport()

			// Launch huh form for the pending tools
			if len(m.pendingToolUses) == 1 {
				return m, forms.RunToolApprovalForm(m.pendingToolUses[0])
			}
			// For multiple tools, run them in batch
			return m, forms.RunToolApprovalFormBatch(m.pendingToolUses)
		}
		// No tool, just commit regular text
		m.CommitStreamingText()

		m.SetStatus(StatusIdle)
		m.streamChan = nil
		m.streamCancel = nil
		m.streamCtx = nil

		m.updateViewport()
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

// handleQuickActionsKey is deprecated - quick actions are now handled by huh forms
// This function is kept for backward compatibility but should not be called
// func (m *Model) handleQuickActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
// 	// Quick actions are now handled by huh forms in LaunchQuickActionsForm()
// 	return m, nil
// }
