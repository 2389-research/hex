// ABOUTME: Bubbletea update function for handling events
// ABOUTME: Processes keyboard input, window resize, streaming chunks
package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/storage"
	"github.com/harper/clem/internal/tools"
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

		m.updateViewport()

		// Send tool results back to API and continue conversation
		return m, m.sendToolResults()

	case tea.KeyMsg:
		// Task 12: Handle tool approval keys
		if m.toolApprovalMode {
			switch msg.Type {
			case tea.KeyRunes:
				if len(msg.Runes) > 0 {
					r := msg.Runes[0]
					if r == 'y' || r == 'Y' || r == 'a' || r == 'A' {
						return m, m.ApproveToolUse()
					} else if r == 'n' || r == 'N' || r == 'd' || r == 'D' {
						return m, m.DenyToolUse()
					} else if r == 'v' || r == 'V' {
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

		// Handle Esc key - can quit or exit search mode
		if msg.Type == tea.KeyEsc {
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

		// Handle Tab to switch views
		if msg.Type == tea.KeyTab {
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

			// Vim-like navigation
			switch r {
			case '/':
				m.EnterSearchMode()
				return m, nil

			case 'j':
				// Scroll down
				m.Viewport.LineDown(1)
				return m, nil

			case 'k':
				// Scroll up
				m.Viewport.LineUp(1)
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
		m.Input, cmd = m.Input.Update(msg)
		cmds = append(cmds, cmd)
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
			content.WriteString("You: " + msg.Content + "\n\n")
		} else {
			// Use glamour for assistant messages
			rendered, err := m.RenderMessage(msg)
			if err == nil {
				content.WriteString("Assistant:\n" + rendered + "\n")
			} else {
				content.WriteString("Assistant: " + msg.Content + "\n\n")
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
		m.updateViewport()
		return m, nil
	}

	chunk := msg.Chunk

	// Task 12: Handle tool_use content blocks
	if chunk.Type == "content_block_start" && chunk.ContentBlock != nil {
		if chunk.ContentBlock.Type == "tool_use" {
			// Commit any streaming text before handling tool
			m.CommitStreamingText()

			// Extract tool use
			toolUse := &core.ToolUse{
				Type:  "tool_use",
				ID:    chunk.ContentBlock.ID,
				Name:  chunk.ContentBlock.Name,
				Input: chunk.ContentBlock.Input,
			}

			// Pause streaming and handle tool
			m.streamChan = nil // Pause stream processing
			return m, m.HandleToolUse(toolUse)
		}
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
func (m *Model) streamMessage(userInput string) tea.Cmd {
	// Build message history for API
	messages := make([]core.Message, 0, len(m.Messages))
	for _, msg := range m.Messages {
		messages = append(messages, core.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Create API request
	req := core.MessageRequest{
		Model:     m.Model,
		Messages:  messages,
		MaxTokens: 4096,
		Stream:    true,
	}

	return func() tea.Msg {
		// Create cancellable context for this stream
		m.streamCtx, m.streamCancel = context.WithCancel(context.Background())

		// Start stream
		streamChan, err := m.apiClient.CreateMessageStream(m.streamCtx, req)
		if err != nil {
			return &StreamChunkMsg{Error: err}
		}

		// Store the channel in the model (this is a bit tricky with Bubbletea)
		// We'll need to return a special message to store it
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
