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
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/pubsub"
	"github.com/2389-research/hex/internal/services"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/hex/internal/ui/forms"
	tea "github.com/charmbracelet/bubbletea"
)

// truncateDebug truncates string for debug logging
func truncateDebug(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// Update handles Bubbletea messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Reset gg sequence on any non-rune key event
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type != tea.KeyRunes {
		m.lastKeyWasG = false
	}

	switch msg := msg.(type) {
	// Handle mouse events
	case tea.MouseMsg:
		// If Shift is held, pass through to terminal for text selection
		if msg.Shift {
			return m, nil
		}

		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.Viewport.ScrollUp(3)
			return m, nil
		case tea.MouseButtonWheelDown:
			m.Viewport.ScrollDown(3)
			return m, nil
		case tea.MouseButtonNone:
			// Handle mouse hover
			if msg.Action == tea.MouseActionMotion {
				m.updateHoveredMessage(msg.X, msg.Y)
				return m, nil
			}
		}

	// Phase 4 Task 3: Handle conversation events
	case conversationEventMsg:
		return m.handleConversationEvent(msg)

	// Phase 4 Task 3: Handle message events
	case messageEventMsg:
		return m.handleMessageEvent(msg)

	// Handle subscription errors
	case subscriptionErrorMsg:
		// Display subscription failures so users know data may be stale
		m.ErrorMessage = fmt.Sprintf("Subscription error: %v", msg.err)

		// Attempt to restart subscriptions after error
		if m.convSvc != nil && m.msgSvc != nil && m.eventCtx != nil {
			if os.Getenv("HEX_DEBUG") != "" {
				_, _ = fmt.Fprintf(os.Stderr, "[SUBSCRIPTION] Restarting after error: %v\n", msg.err)
			}

			// Restart subscriptions
			convEvents := m.convSvc.Subscribe(m.eventCtx)
			msgEvents := m.msgSvc.Subscribe(m.eventCtx)

			return m, tea.Batch(
				waitForConversationEvent(convEvents),
				waitForMessageEvent(msgEvents),
			)
		}
		return m, nil

	// Task 6: Handle streaming chunks
	case *StreamChunkMsg:
		return m.handleStreamChunk(msg)

	case *streamStartMsg:
		// Store the stream channel and start reading
		m.streamChan = msg.channel
		m.SetStatus(StatusStreaming)
		m.Streaming = true // Set streaming flag for queue logic
		// Show thinking indicator while waiting for first response chunk
		if m.streamingDisplay != nil {
			m.streamingDisplay.SetThinking(true, "")
		}
		// Add assistant message placeholder immediately to preserve message ordering
		// This ensures the assistant response appears after the user message that triggered it,
		// even if the user sends more messages while streaming
		// We add directly to m.Messages to bypass the empty content check in AddMessage
		m.Messages = append(m.Messages, Message{
			Role:      "assistant",
			Content:   "",
			Timestamp: time.Now(),
		})
		m.updateViewport()
		return m, m.readStreamChunks(m.streamCtx, m.streamChan)

	// Handle tool decision from queue-based approval overlay
	case ToolDecisionMsg:
		_, _ = fmt.Fprintf(os.Stderr, "[TOOL_DECISION] received decision=%d\n", msg.Decision)
		return m, m.HandleToolDecision(msg.Decision)

	// Handle completion of a single queued tool execution
	case toolQueueExecutionMsg:
		_, _ = fmt.Fprintf(os.Stderr, "[QUEUE_EXEC_DONE] tool_use_id=%s\n", msg.result.ToolUseID)

		// Add result to queue and history
		if m.activeToolQueue != nil {
			m.activeToolQueue.AddResult(msg.result)
		}
		m.toolResultHistory = append(m.toolResultHistory, msg.result)

		// Update cached most recent tool ID for tail preview display
		m.updateMostRecentToolID()

		// Display result in UI
		resultText := formatToolResult(msg.result.Result)
		m.AddMessage("tool", resultText)

		// Update tool log
		if msg.result.Result != nil {
			toolName := msg.result.Result.ToolName
			if toolName == "" {
				toolName = "unknown"
			}
			m.appendToolLogLine(fmt.Sprintf("─── %s ───", toolName))
			if msg.result.Result.Output != "" {
				m.appendToolLogOutput(msg.result.Result.Output)
			} else if msg.result.Result.Error != "" {
				m.appendToolLogLine("Error: " + msg.result.Result.Error)
			}
		}

		m.updateViewport()

		// Continue processing next tool in queue
		return m, m.ProcessNextTool()

	// Handle quick actions form results from huh
	case *forms.QuickActionsResultMsg:
		return m.handleQuickActionsResult(msg), nil

	case tea.KeyMsg:
		// PRIORITY 1: Route input to overlay manager FIRST (modal behavior)
		// Overlays capture ALL input when active, except special global hotkeys
		if m.overlayManager != nil && m.overlayManager.HasActive() {
			// Check for special global hotkeys that should always work
			isGlobalHotkey := false
			switch msg.Type {
			case tea.KeyCtrlO, tea.KeyCtrlH, tea.KeyCtrlR:
				// Global overlay toggle hotkeys - process below
				isGlobalHotkey = true
			}

			if !isGlobalHotkey {
				// Let overlay handle the key first
				var handled bool
				handled, cmd = m.overlayManager.HandleKey(msg)
				if handled {
					// For tool approval overlays, HandleKey returns ToolDecisionMsg
					// which is processed by HandleToolDecision - it handles everything
					// including popping the overlay. Don't interfere here.
					active := m.overlayManager.GetActive()
					if _, isToolApproval := active.(*ToolApprovalOverlay); isToolApproval {
						// Just return the cmd (ToolDecisionMsg) - queue system handles the rest
						return m, cmd
					}

					// For non-tool-approval overlays, handle Escape/CtrlC to pop them
					if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
						if active != nil {
							cmd = active.Cancel()
						}
						m.overlayManager.Pop()
						m.adjustViewportForOverlay()
					}

					// Update scrollable overlays with viewport messages
					active = m.overlayManager.GetActive()
					if active != nil {
						if scrollable, ok := active.(Scrollable); ok {
							scrollCmd := scrollable.Update(msg)
							if scrollCmd != nil {
								if cmd != nil {
									return m, tea.Batch(cmd, scrollCmd)
								}
								return m, scrollCmd
							}
						}
					}

					return m, cmd
				}
			}
		}

		// Intro screen: Switch to chat mode
		// For Tab key specifically, return early to avoid double-processing
		// For other keys, continue processing to allow typing
		if m.CurrentView == ViewModeIntro {
			m.CurrentView = ViewModeChat
			// Return early only for Tab to avoid hitting NextView() below
			if msg.Type == tea.KeyTab {
				return m, nil
			}
			// For other keys, continue processing below
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

		// Handle Ctrl+O: toggle tool timeline overlay
		if msg.Type == tea.KeyCtrlO {
			active := m.overlayManager.GetActive()
			if active == m.toolTimelineOverlay {
				// Already open, close it
				m.overlayManager.Pop()
				m.adjustViewportForOverlay()
			} else if active == nil {
				// No overlay open, open timeline
				m.overlayManager.Push(m.toolTimelineOverlay, m.Width, m.Height)
				m.adjustViewportForOverlay()
			}
			// If another overlay is open, do nothing (don't interrupt it)
			return m, nil
		}

		// Handle Ctrl+H: toggle help overlay
		if msg.Type == tea.KeyCtrlH {
			active := m.overlayManager.GetActive()
			if active == m.helpOverlay {
				// Already open, close it
				m.overlayManager.Pop()
				m.adjustViewportForOverlay()
			} else if active == nil {
				// No overlay open, open help
				m.overlayManager.Push(m.helpOverlay, m.Width, m.Height)
				m.adjustViewportForOverlay()
			}
			// If another overlay is open, do nothing (don't interrupt it)
			return m, nil
		}

		// Handle Ctrl+R: toggle history overlay
		if msg.Type == tea.KeyCtrlR {
			active := m.overlayManager.GetActive()
			if active == m.historyOverlay {
				// Already open, close it
				m.overlayManager.Pop()
				m.adjustViewportForOverlay()
			} else if active == nil {
				// No overlay open, open history
				m.overlayManager.Push(m.historyOverlay, m.Width, m.Height)
				m.adjustViewportForOverlay()
			}
			// If another overlay is open, do nothing (don't interrupt it)
			return m, nil
		}

		// Handle Esc key - fallback handling for when no overlay is active
		// Overlays are handled at the top of KeyMsg processing
		// This only handles Esc when no overlay is active
		if msg.Type == tea.KeyEsc {
			if m.helpVisible {
				m.ToggleHelp()
				return m, nil
			}
			if m.SearchMode {
				m.ExitSearchMode()
				return m, nil
			}
			if m.quickActionsMode {
				m.quickActionsMode = false
				return m, nil
			}
			// Escape doesn't quit - user must use Ctrl+C or exit command
			return m, nil
		}

		// Handle Ctrl+C: cancel active work on first press, quit on second press when idle
		// Note: Overlay Ctrl+C is handled at the top of KeyMsg processing
		if msg.Type == tea.KeyCtrlC {
			// First priority: if input box has content, clear it
			if strings.TrimSpace(m.Input.Value()) != "" {
				m.Input.Reset()
				m.updateInputHeight()
				m.pendingQuit = false // Reset quit state
				return m, nil
			}

			// If streaming, first Ctrl+C cancels the stream
			if m.Streaming {
				m.cancelStream()
				m.Streaming = false
				m.StreamingText = ""
				m.Status = StatusIdle
				m.pendingQuit = false // Reset quit state
				return m, nil
			}

			// If in quick actions mode, first Ctrl+C cancels it
			if m.quickActionsMode {
				m.quickActionsMode = false
				m.pendingQuit = false
				return m, nil
			}

			// If in search mode, first Ctrl+C cancels it
			if m.SearchMode {
				m.ExitSearchMode()
				m.pendingQuit = false
				return m, nil
			}

			// Check if this is within the confirmation window (2 seconds)
			if m.pendingQuit && time.Since(m.pendingQuitTime) < 2*time.Second {
				// Second Ctrl+C within timeout - actually quit
				m.cancelStream()
				if m.eventCancel != nil {
					m.eventCancel()
				}
				return m, tea.Quit
			}

			// First Ctrl+C when idle - show warning and wait for confirmation
			m.pendingQuit = true
			m.pendingQuitTime = time.Now()
			return m, nil
		}

		// Reset pending quit on any other key press
		if m.pendingQuit {
			m.pendingQuit = false
		}

		// Handle autocomplete completion acceptance
		// (Up/Down navigation now handled in AutocompleteOverlay)
		if m.autocomplete != nil && m.autocomplete.IsActive() {
			switch msg.Type {
			case tea.KeyTab, tea.KeyEnter:
				// Accept autocomplete selection
				selected := m.autocomplete.GetSelected()
				if selected != nil {
					m.Input.SetValue(selected.Value)
					m.autocomplete.Hide()
					// Pop the overlay from stack
					if m.overlayManager.GetActive() == m.autocompleteOverlay {
						m.overlayManager.Pop()
						m.adjustViewportForOverlay()
					}
				}
				return m, nil
			case tea.KeyEsc:
				// Cancel autocomplete
				m.autocomplete.Hide()
				// Pop the overlay from stack
				if m.overlayManager.GetActive() == m.autocompleteOverlay {
					m.overlayManager.Pop()
					m.adjustViewportForOverlay()
				}
				return m, nil
			}
		}

		// Handle Tab - trigger autocomplete or switch views
		if msg.Type == tea.KeyTab {
			// If textarea is focused and has content, show autocomplete
			if m.Input.Focused() && m.Input.Value() != "" {
				provider := DetectProvider(m.Input.Value())
				m.autocomplete.Show(m.Input.Value(), provider)
				// Push autocomplete overlay if not already active
				if m.overlayManager.GetActive() != m.autocompleteOverlay {
					m.overlayManager.Push(m.autocompleteOverlay, m.Width, m.Height)
					m.adjustViewportForOverlay()
				}
				return m, nil
			}
			// Otherwise, switch views
			m.NextView()
			return m, nil
		}

		// Handle Shift+Enter: insert newline in textarea for multi-line input
		// Most terminals send Ctrl+J (ASCII 10, line feed) for Shift+Enter
		if msg.Type == tea.KeyCtrlJ {
			m.Input.InsertString("\n")
			m.updateInputHeight()
			return m, nil
		}

		// Handle Enter key
		if msg.Type == tea.KeyEnter {
			if m.SearchMode {
				// Execute search (placeholder for now)
				m.ExitSearchMode()
				return m, nil
			}
			// Handle Alt+Enter as fallback: insert newline in textarea
			// This supports terminals that don't send Ctrl+J for Shift+Enter
			if msg.Alt {
				// Insert a newline at cursor position
				m.Input.InsertString("\n")
				m.updateInputHeight()
				return m, nil
			}
			// Plain Enter sends the message
			input := strings.TrimSpace(m.Input.Value())

			// Check for exit commands
			if input == "exit" || input == "/exit" {
				// Cancel any active stream before quitting
				m.cancelStream()
				// Cancel event subscriptions before quitting
				if m.eventCancel != nil {
					m.eventCancel()
				}
				return m, tea.Quit
			}

			// Check for clear command
			if input == "/clear" {
				m.ClearContext()
				return m, nil
			}

			if input != "" {
				// Add to input history
				m.addToInputHistory(input)

				// Check if waiting for response - if so, queue the message
				if m.waitingForResponse {
					// Only allow one queued message
					if m.queuedMessage != "" {
						// Already have a queued message - ignore
						return m, nil
					}
					m.queuedMessage = input
					m.Input.Reset()
					m.updateInputHeight() // Reset height to 1 line after clearing
					m.updateViewportPreserveScroll()
					return m, nil
				}

				// Not waiting - process immediately
				m.AddMessage("user", input)
				m.Input.Reset()
				m.updateInputHeight()       // Reset height to 1 line after clearing
				m.waitingForResponse = true // Block further input until response complete
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

		// Handle backspace in search mode
		if msg.Type == tea.KeyBackspace {
			if m.SearchMode && len(m.SearchQuery) > 0 {
				m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
				return m, nil
			}
		}

		// Handle j/k scrolling BEFORE textarea input handling
		// This allows scrolling even when input is focused
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			r := msg.Runes[0]
			if r == 'j' && !m.SearchMode && m.Input.Value() == "" {
				// Scroll down when input is empty
				m.Viewport.ScrollDown(1)
				return m, nil
			}
			if r == 'k' && !m.SearchMode && m.Input.Value() == "" {
				// Scroll up when input is empty
				m.Viewport.ScrollUp(1)
				return m, nil
			}
		}

		// Handle UP arrow to edit queued message
		// If there's a queued message and input is empty, pull it back for editing
		if m.Input.Focused() && msg.Type == tea.KeyUp && m.queuedMessage != "" && m.Input.Value() == "" {
			m.Input.SetValue(m.queuedMessage)
			m.queuedMessage = ""
			m.updateViewportPreserveScroll()
			return m, nil
		}

		// Handle up/down arrows for input history navigation
		// Only when textarea is focused and cursor is on first/last visual row
		// This allows Up/Down to scroll within multiline input normally
		if m.Input.Focused() && len(m.inputHistory) > 0 {
			lineInfo := m.Input.LineInfo()
			cursorOnFirstRow := lineInfo.RowOffset == 0
			cursorOnLastRow := lineInfo.RowOffset == lineInfo.Height-1

			if msg.Type == tea.KeyUp && cursorOnFirstRow {
				// Navigate to older history only when on first row
				if m.navigateHistoryUp() {
					return m, nil
				}
			}
			if msg.Type == tea.KeyDown && cursorOnLastRow {
				// Navigate to newer history only when on last row
				if m.navigateHistoryDown() {
					return m, nil
				}
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
		// Re-entrance guard: prevent infinite loops from WindowSizeMsg
		if m.processingWindowSize {
			return m, nil
		}
		m.processingWindowSize = true
		defer func() { m.processingWindowSize = false }()

		m.Width = msg.Width
		m.Height = msg.Height
		m.Input.SetWidth(msg.Width - 4)
		m.Viewport.Width = msg.Width - 4
		m.baseViewportHeight = msg.Height - 8 // Store base height
		m.adjustViewportForOverlay()          // Set actual height based on active overlay
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
		// NOTE: Async approval form handles its own sizing via form.Run()
	}

	// Phase 6C: Update spinner
	if m.spinner != nil {
		cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update input (only if not in search mode and no queued message)
	if !m.SearchMode && m.queuedMessage == "" {
		oldValue := m.Input.Value()
		m.Input, cmd = m.Input.Update(msg)
		cmds = append(cmds, cmd)

		newValue := m.Input.Value()
		if newValue != oldValue {
			// Phase 6C Task 4: Update autocomplete as user types
			if m.autocomplete != nil {
				if m.autocomplete.IsActive() {
					// If the input no longer starts with /, hide autocomplete
					if !strings.HasPrefix(strings.TrimSpace(newValue), "/") {
						m.autocomplete.Hide()
						// Pop overlay if active
						if m.overlayManager.GetActive() == m.autocompleteOverlay {
							m.overlayManager.Pop()
							m.adjustViewportForOverlay()
						}
					} else {
						// Already active - just update with new input
						m.autocomplete.Update(newValue)
					}
				} else if strings.HasPrefix(strings.TrimSpace(newValue), "/") {
					// Auto-show autocomplete when typing starts with /
					provider := DetectProvider(newValue)
					m.autocomplete.Show(newValue, provider)
					// Push overlay if not already active
					if m.overlayManager.GetActive() != m.autocompleteOverlay {
						m.overlayManager.Push(m.autocompleteOverlay, m.Width, m.Height)
						m.adjustViewportForOverlay()
					}
				}
			}

			// Auto-grow input height based on content (up to MaxHeight)
			m.updateInputHeight()
		}
	}

	// Update viewport - but don't pass key messages when input is focused
	// This prevents arrow keys from scrolling the conversation while typing
	shouldUpdateViewport := true
	if _, isKeyMsg := msg.(tea.KeyMsg); isKeyMsg && m.Input.Focused() {
		shouldUpdateViewport = false
	}
	if shouldUpdateViewport {
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateViewport renders messages into viewport with throttling for smooth performance
// Limits updates to 60fps max to reduce CPU overhead for expensive renders
// Always scrolls to bottom - use updateViewportPreserveScroll() to keep position
func (m *Model) updateViewport() {
	m.updateViewportInternal(true)
}

// updateViewportPreserveScroll renders messages without scrolling to bottom
func (m *Model) updateViewportPreserveScroll() {
	m.updateViewportInternal(false)
}

// adjustViewportForOverlay recalculates viewport height based on active bottom overlay
// This should be called when overlays are pushed/popped, not during View()
func (m *Model) adjustViewportForOverlay() {
	if m.baseViewportHeight == 0 || m.overlayManager == nil {
		return // Not initialized yet
	}

	// Check for active bottom (non-fullscreen) overlay
	active := m.overlayManager.GetActive()
	if active == nil {
		// No overlay - use base height
		m.Viewport.Height = m.baseViewportHeight
		return
	}

	// Fullscreen overlays don't affect viewport height (they replace viewport entirely)
	if _, isFullscreen := active.(FullscreenOverlay); isFullscreen {
		m.Viewport.Height = m.baseViewportHeight
		return
	}

	// Bottom overlay - reduce viewport height
	bottomOverlayHeight := active.GetDesiredHeight()
	if bottomOverlayHeight > m.Height/2 {
		bottomOverlayHeight = m.Height / 2 // Cap at half screen
	}

	adjustedHeight := m.baseViewportHeight - bottomOverlayHeight
	if adjustedHeight < 5 {
		adjustedHeight = 5 // Minimum height for viewport
	}
	m.Viewport.Height = adjustedHeight
}

// updateViewportInternal is the shared implementation
func (m *Model) updateViewportInternal(scrollToBottom bool) {
	// Performance: Always throttle viewport updates to 60fps max (16.67ms = 60fps)
	// Expensive renders happen anytime (glamour markdown, large messages), not just streaming
	timeSinceLastUpdate := time.Since(m.lastViewportUpdate)
	if timeSinceLastUpdate < 16*time.Millisecond {
		// Too soon since last update - skip this update
		// This prevents excessive CPU usage from rapid updateViewport() calls
		return
	}

	// Record the start of this update BEFORE doing the expensive work
	// This ensures subsequent calls are throttled even if this render takes time
	m.lastViewportUpdate = time.Now()

	// Render viewport content
	var content strings.Builder

	// Prepend intro screen if ShowIntro is true
	if m.ShowIntro {
		introContent := m.renderIntroView()
		content.WriteString(introContent)
		content.WriteString("\n\n")
		// Debug: Log intro content
		if os.Getenv("HEX_VIEW_DEBUG") != "" {
			f, _ := os.OpenFile("/tmp/hex-view-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				introLines := strings.Split(introContent, "\n")
				_, _ = fmt.Fprintf(f, "DEBUG intro: %d lines\n", len(introLines))
				for i := 0; i < min(5, len(introLines)); i++ {
					_, _ = fmt.Fprintf(f, "DEBUG intro[%d]='%s'\n", i, truncateDebug(introLines[i], 50))
				}
				_ = f.Close()
			}
		}
	}

	// Render messages with Neo-Terminal style
	// Safe: BubbleTea guarantees single-threaded Update calls,
	// so Messages slice won't be modified during this loop.
	// Taking pointers is safe because no concurrent reallocation can occur.
	for i := range m.Messages {
		msg := &m.Messages[i] // Get pointer to allow cache updates

		// Skip internal messages that shouldn't be displayed to users:
		// - "tool" role messages (internal tool result tracking)
		// - "user" messages with only tool_result ContentBlocks (API protocol)
		if msg.Role == "tool" {
			continue
		}
		if msg.Role == "user" && msg.Content == "" && len(msg.ContentBlock) > 0 {
			// Check if this is a tool_result only message
			allToolResults := true
			for _, block := range msg.ContentBlock {
				if block.Type != "tool_result" {
					allToolResults = false
					break
				}
			}
			if allToolResults {
				continue // Skip displaying tool_result user messages
			}
		}

		// Get timestamp (use current time if not set for backward compatibility)
		timestamp := msg.Timestamp
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		// Handle messages with ContentBlock (tool results, tool uses)
		if msg.Content == "" && len(msg.ContentBlock) > 0 {
			// Render structured content blocks instead of skipping
			blockContent := m.renderContentBlocks(msg.ContentBlock)
			if blockContent != "" {
				neoMessage := m.renderNeoTerminalMessage(msg.Role, blockContent, timestamp)
				content.WriteString(neoMessage)
			}
			continue
		}

		// Render message content (uses cache for performance)
		messageContent := msg.Content
		if msg.Role == "assistant" {
			// Use glamour for assistant messages
			rendered, err := m.RenderMessage(msg)
			if err == nil {
				// Strip extra newlines from glamour output
				messageContent = strings.TrimSpace(rendered)
			}
		}

		// Render with Neo-Terminal style
		neoMessage := m.renderNeoTerminalMessage(msg.Role, messageContent, timestamp)
		content.WriteString(neoMessage)

		// Add spacing between messages (message already has trailing newline)
		if i < len(m.Messages)-1 {
			content.WriteString("\n")
		}
	}

	// Streaming text is now handled by updating the placeholder message in m.Messages
	// No need to render it separately here (that would cause duplication)

	// Show work display if streaming and display is available
	if m.Streaming && m.streamingDisplay != nil {
		workDisplay := m.streamingDisplay.GetWorkDisplay()
		if workDisplay != "" {
			content.WriteString("\n\n" + workDisplay)
		}
	}

	m.Viewport.SetContent(content.String())
	// Only scroll to bottom if we have messages (not just intro)
	// When only intro is shown, keep scroll at top so full logo is visible
	if scrollToBottom && len(m.Messages) > 0 {
		m.Viewport.GotoBottom()
	}
}

// Task 6: Streaming Integration Functions

// handleStreamError cleans up state when a stream error occurs
func (m *Model) handleStreamError(err error) (tea.Model, tea.Cmd) {
	if os.Getenv("HEX_DEBUG") != "" {
		logging.Debug("Stream error", "error", err)
	}
	m.ClearStreamingText()

	// Remove empty assistant message placeholder if error occurred before any content
	if len(m.Messages) > 0 {
		lastMsg := &m.Messages[len(m.Messages)-1]
		if lastMsg.Role == "assistant" && strings.TrimSpace(lastMsg.Content) == "" {
			// Remove the empty placeholder message
			m.Messages = m.Messages[:len(m.Messages)-1]
		}
	}

	m.SetStatus(StatusError)
	m.ErrorMessage = err.Error()
	m.streamChan = nil
	m.streamCancel = nil
	m.streamCtx = nil
	m.waitingForResponse = false // Allow new input after error
	m.updateViewport()
	return m, nil
}

// handleContentBlockStart processes the start of a content block (tool_use)
func (m *Model) handleContentBlockStart(chunk *core.StreamChunk) (tea.Model, tea.Cmd) {
	// Only handle tool_use blocks - for text blocks, just continue reading
	if chunk.ContentBlock == nil || chunk.ContentBlock.Type != "tool_use" {
		return m, m.continueReading()
	}

	_, _ = fmt.Fprintf(os.Stderr, "[STREAM_TOOL_START] tool_use detected: id=%s, name=%s\n",
		chunk.ContentBlock.ID, chunk.ContentBlock.Name)

	// Start assembling tool use - Input will come in delta events
	m.assemblingToolUse = &core.ToolUse{
		Type:  "tool_use",
		ID:    chunk.ContentBlock.ID,
		Name:  chunk.ContentBlock.Name,
		Input: make(map[string]interface{}), // Will be populated when JSON is complete
	}
	m.toolInputJSONBuf = "" // Reset JSON buffer

	// Update streaming display to show tool call in progress
	if m.streamingDisplay != nil {
		m.streamingDisplay.StartToolCall(chunk.ContentBlock.ID, chunk.ContentBlock.Name)
	}
	m.updateViewport()

	// Continue reading from stream
	return m, m.continueReading()
}

// handleContentBlockDelta processes delta events for content blocks (text or tool input JSON)
func (m *Model) handleContentBlockDelta(chunk *core.StreamChunk) (tea.Model, tea.Cmd) {
	if chunk.Delta == nil {
		// Even with nil delta, continue reading the stream
		return m, m.continueReading()
	}

	// Handle input_json_delta for tool use parameters
	if chunk.Delta.Type == "input_json_delta" && m.assemblingToolUse != nil {
		// Accumulate non-empty JSON chunks
		if chunk.Delta.PartialJSON != "" {
			m.toolInputJSONBuf += chunk.Delta.PartialJSON
		}
		// ALWAYS continue reading - even empty chunks are valid stream events
		return m, m.continueReading()
	}

	// Handle text delta for assistant message
	if chunk.Delta.Text != "" {
		if os.Getenv("HEX_DEBUG") != "" {
			logging.Debug("Stream text delta", "text", chunk.Delta.Text, "length", len(chunk.Delta.Text))
		}
		m.AppendStreamingText(chunk.Delta.Text)
		if m.streamingDisplay != nil {
			// Turn off thinking indicator once we receive actual content
			if m.streamingDisplay.IsWaitingForTokens() {
				m.streamingDisplay.SetThinking(false, "")
			}
			m.streamingDisplay.AppendText(chunk.Delta.Text)
		}
		m.SetStatus(StatusStreaming)
		m.updateViewport()
		// Continue reading
		return m, m.continueReading()
	}

	// For any other delta type, continue reading
	return m, m.continueReading()
}

// handleContentBlockStop processes the completion of a content block (tool parameters complete)
func (m *Model) handleContentBlockStop() (tea.Model, tea.Cmd) {
	if m.assemblingToolUse == nil {
		// No tool being assembled, just a text block - continue reading for message_stop
		return m, m.continueReading()
	}

	// Parse accumulated JSON into Input map
	if m.toolInputJSONBuf != "" {
		var input map[string]interface{}
		if err := json.Unmarshal([]byte(m.toolInputJSONBuf), &input); err == nil {
			m.assemblingToolUse.Input = input
			_, _ = fmt.Fprintf(os.Stderr, "[STREAM_TOOL_INPUT] parsed input for tool_use_id=%s, input=%+v\n",
				m.assemblingToolUse.ID, input)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "[STREAM_TOOL_INPUT_ERROR] failed to parse input JSON for tool_use_id=%s: %v\n",
				m.assemblingToolUse.ID, err)
		}
	}

	// Mark tool call as complete in streaming display
	if m.streamingDisplay != nil {
		m.streamingDisplay.CompleteToolCall(m.assemblingToolUse.ID)
	}
	m.updateViewport()

	// Tool use is complete, append to pending tools list
	m.pendingToolUses = append(m.pendingToolUses, m.assemblingToolUse)
	fmt.Fprintf(os.Stderr, "[STREAM_TOOL_COMPLETE] tool_use complete, added to pending (total pending: %d): id=%s, name=%s\n",
		len(m.pendingToolUses), m.assemblingToolUse.ID, m.assemblingToolUse.Name)

	m.assemblingToolUse = nil
	m.toolInputJSONBuf = ""

	// Continue reading from stream
	return m, m.continueReading()
}

// handleMessageDelta processes usage metadata updates
func (m *Model) handleMessageDelta(chunk *core.StreamChunk) (tea.Model, tea.Cmd) {
	if chunk.Usage != nil {
		m.UpdateTokens(chunk.Usage.InputTokens, chunk.Usage.OutputTokens)
	}
	// Continue reading from stream
	return m, m.continueReading()
}

// handleMessageStop processes stream completion and handles tool approval or text commit
func (m *Model) handleMessageStop() (tea.Model, tea.Cmd) {
	_, _ = fmt.Fprintf(os.Stderr, "[STREAM_STOP] message stream ended, pendingToolUses count=%d\n", len(m.pendingToolUses))

	if os.Getenv("HEX_DEBUG") != "" {
		logging.Debug("Message stream stopped", "pending_tools", len(m.pendingToolUses), "streaming_text_len", len(m.StreamingText))
	}

	// Always clear stream state when message completes, regardless of path
	m.streamChan = nil
	m.streamCancel = nil
	m.streamCtx = nil

	// Path 1: Stream completed with tool uses - need approval
	if len(m.pendingToolUses) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "[STREAM_STOP_WITH_TOOLS] creating assistant message with %d tool_use block(s)\n", len(m.pendingToolUses))

		if os.Getenv("HEX_DEBUG") != "" {
			logging.Debug("Creating assistant message with tool uses", "tool_count", len(m.pendingToolUses), "text_length", len(m.StreamingText))
		}

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

			if os.Getenv("HEX_DEBUG") != "" {
				inputJSON, _ := json.Marshal(toolUse.Input)
				logging.Debug("Adding tool use to message", "tool_id", toolUse.ID, "tool_name", toolUse.Name, "input", string(inputJSON))
			}
		}

		// Update the existing placeholder message with content blocks
		// The placeholder was added in streamStartMsg to preserve message ordering
		if len(m.Messages) > 0 && m.Messages[len(m.Messages)-1].Role == "assistant" {
			// Update the last assistant message (the placeholder) with content blocks
			m.Messages[len(m.Messages)-1].ContentBlock = blocks
			m.Messages[len(m.Messages)-1].Content = "" // Clear any streaming text
		} else {
			// Fallback: add new message if no placeholder exists
			m.Messages = append(m.Messages, Message{
				Role:         "assistant",
				ContentBlock: blocks,
			})
		}

		// Save assistant message with tool_use blocks to database
		// Serialize ContentBlock to JSON for storage
		if blocksJSON, err := json.Marshal(blocks); err == nil {
			if err := m.saveMessage("assistant", string(blocksJSON)); err != nil {
				m.ErrorMessage = "Failed to save assistant message: " + err.Error()
			}
		}

		if os.Getenv("HEX_DEBUG") != "" {
			logging.Debug("Assistant message updated with tool uses", "block_count", len(blocks))
		}
		m.StreamingText = ""

		// Hide intro after first assistant response
		if m.ShowIntro {
			m.ShowIntro = false
		}

		// Reset streaming display (work is done, now waiting for approval)
		if m.streamingDisplay != nil {
			m.streamingDisplay.Reset()
		}

		_, _ = fmt.Fprintf(os.Stderr, "[STREAM_STOP_WITH_TOOLS] assistant message added to history (total messages: %d)\n", len(m.Messages))

		// Dump messages after adding assistant message with tool_use blocks
		m.dumpMessages("AFTER stream completion with tool_use blocks")

		// NEW QUEUE SYSTEM: Create queue from pending tools and start processing
		// Get permission mode from checker (default to "ask" if no checker)
		permMode := "ask"
		if m.toolExecutor != nil && m.toolExecutor.GetPermissionChecker() != nil {
			permMode = m.toolExecutor.GetPermissionChecker().GetMode().String()
		}
		m.activeToolQueue = NewToolQueue(m.pendingToolUses, m.approvalRules, permMode)
		m.pendingToolUses = nil // Clear - queue now owns them

		needsApproval, autoApprove, autoDeny := m.activeToolQueue.CountByAction()
		_, _ = fmt.Fprintf(os.Stderr, "[QUEUE_CREATED] queue has %d tools: %d need approval, %d auto-approve, %d auto-deny\n",
			m.activeToolQueue.Len(), needsApproval, autoApprove, autoDeny)

		m.updateViewport()
		return m, m.ProcessNextTool()
	}

	// Path 2: No tools, just commit regular text
	m.CommitStreamingText()

	// Save assistant message to database
	if len(m.Messages) > 0 {
		lastMsg := m.Messages[len(m.Messages)-1]
		if lastMsg.Role == "assistant" && lastMsg.Content != "" {
			if err := m.saveMessage("assistant", lastMsg.Content); err != nil {
				m.ErrorMessage = "Failed to save assistant message: " + err.Error()
			}
		}
	}

	// Hide intro after first assistant response
	if m.ShowIntro {
		m.ShowIntro = false
	}

	// Reset streaming display
	if m.streamingDisplay != nil {
		m.streamingDisplay.Reset()
	}

	m.SetStatus(StatusIdle)
	m.Streaming = false          // Clear streaming flag
	m.waitingForResponse = false // Allow new input
	// Stream state already cleared at start of handleMessageStop()

	m.updateViewport()

	// Check if there's a queued message to process
	return m, m.processQueuedMessage()
}

// handleStreamChunk processes a streaming chunk message
func (m *Model) handleStreamChunk(msg *StreamChunkMsg) (tea.Model, tea.Cmd) {
	// Handle errors
	if msg.Error != nil {
		return m.handleStreamError(msg.Error)
	}

	chunk := msg.Chunk

	// Debug logging: Log chunk type
	if os.Getenv("HEX_DEBUG") != "" {
		chunkJSON, _ := json.Marshal(chunk)
		logging.Debug("Stream chunk received", "type", chunk.Type, "chunk", string(chunkJSON))
	}

	// Handle tool_use content block start
	if chunk.Type == "content_block_start" {
		return m.handleContentBlockStart(chunk)
	}

	// Handle content block deltas (text or tool input JSON)
	if chunk.Type == "content_block_delta" {
		return m.handleContentBlockDelta(chunk)
	}

	// Handle content_block_stop - tool parameters complete
	if chunk.Type == "content_block_stop" {
		return m.handleContentBlockStop()
	}

	// Handle usage metadata (message_delta event)
	if chunk.Type == "message_delta" {
		return m.handleMessageDelta(chunk)
	}

	// Handle message completion
	if chunk.Type == "message_stop" || chunk.Done {
		return m.handleMessageStop()
	}

	// For other chunk types, continue reading
	if m.streamChan != nil {
		return m, m.readStreamChunks(m.streamCtx, m.streamChan)
	}

	return m, nil
}

// streamMessage starts streaming a message from the API
func (m *Model) streamMessage(_ string) tea.Cmd {
	// Get pruned message history (automatically compacts if near context limit)
	messages := m.GetPrunedMessages()

	// Get tool definitions from registry
	var tools []core.ToolDefinition
	if m.toolRegistry != nil {
		tools = m.toolRegistry.GetDefinitions()
	}

	// Create API request
	// Always include Hex identity in system prompt
	systemPrompt := core.DefaultSystemPrompt
	if m.systemPrompt != "" {
		systemPrompt = core.DefaultSystemPrompt + "\n\n" + m.systemPrompt
	}

	req := core.MessageRequest{
		Model:     m.Model,
		Messages:  messages,
		MaxTokens: 4096,
		Stream:    true,
		System:    systemPrompt, // Phase 6C: Use system prompt from template with Hex identity
		Tools:     tools,        // Include registered tools
	}

	// Cancel any existing stream context before creating a new one
	m.cancelStream()

	// Create cancellable context for this stream BEFORE the async command
	ctx, cancel := context.WithCancel(context.Background())
	// Store these in the model NOW (synchronously in Update context)
	m.streamCtx = ctx
	m.streamCancel = cancel

	// Capture API client reference
	apiClient := m.apiClient

	return func() tea.Msg {
		// Defensive check: ensure context hasn't been cancelled between
		// context creation and when this async command executes
		select {
		case <-ctx.Done():
			return &StreamChunkMsg{Error: ctx.Err()}
		default:
		}

		// Start stream with the context we created
		streamChan, err := apiClient.CreateMessageStream(ctx, req)
		if err != nil {
			return &StreamChunkMsg{Error: err}
		}

		return &streamStartMsg{channel: streamChan}
	}
}

// processQueuedMessage checks if there's a queued message and processes it
func (m *Model) processQueuedMessage() tea.Cmd {
	if m.queuedMessage == "" {
		// No queued message
		return nil
	}

	// Get and clear the queued message
	message := m.queuedMessage
	m.queuedMessage = ""

	// Add the queued message to conversation history now that it's being processed
	m.AddMessage("user", message)
	m.waitingForResponse = true // Block input while processing queued message
	m.updateViewport()

	// Save to database
	if err := m.saveMessage("user", message); err != nil {
		// Log error but don't block
		m.ErrorMessage = "Failed to save queued message: " + err.Error()
	}

	// Process the queued message
	if m.apiClient != nil {
		return m.streamMessage(message)
	}

	return nil
}

// continueReading is a helper that safely continues reading from the stream
// Consolidates nil checks for both context and channel
func (m *Model) continueReading() tea.Cmd {
	if m.streamChan == nil || m.streamCtx == nil {
		return nil
	}
	return m.readStreamChunks(m.streamCtx, m.streamChan)
}

// readStreamChunks reads from the stream channel and returns messages
// Uses context to handle cancellation and prevent goroutine leaks
func (m *Model) readStreamChunks(ctx context.Context, streamChan <-chan *core.StreamChunk) tea.Cmd {
	return func() tea.Msg {
		// Defensive check for nil channel or context
		if streamChan == nil || ctx == nil {
			return &StreamChunkMsg{
				Chunk: &core.StreamChunk{
					Type: "message_stop",
					Done: true,
				},
			}
		}

		// Read next chunk with context cancellation support
		// This prevents blocking forever if context is cancelled
		select {
		case chunk, ok := <-streamChan:
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

		case <-ctx.Done():
			// Context cancelled - clean shutdown
			return &StreamChunkMsg{
				Error: ctx.Err(),
			}
		}
	}
}

// Task 7: Storage Integration Functions

// saveMessage saves a message to the database via service layer
func (m *Model) saveMessage(role, content string) error {
	if m.msgSvc == nil {
		return nil // No message service, skip saving
	}

	msg := &services.Message{
		ConversationID: m.ConversationID,
		Role:           role,
		Content:        content,
	}

	return m.msgSvc.Add(context.Background(), msg)
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

// updateConversationTitle updates the conversation title via service layer
func (m *Model) updateConversationTitle(title string) error {
	if m.convSvc == nil {
		return nil
	}

	// Get current conversation
	conv, err := m.convSvc.Get(context.Background(), m.ConversationID)
	if err != nil {
		return fmt.Errorf("get conversation: %w", err)
	}

	// Update title
	conv.Title = title
	return m.convSvc.Update(context.Background(), conv)
}

// Task 12: Tool result formatting

// renderContentBlocks formats ContentBlock array for display (tool results, tool uses)
func (m *Model) renderContentBlocks(blocks []core.ContentBlock) string {
	if len(blocks) == 0 {
		return ""
	}

	var b strings.Builder
	for i, block := range blocks {
		switch block.Type {
		case "tool_result":
			// Render tool result with visual indicator
			b.WriteString(m.theme.ToolResult.Render("🔧 Tool Result"))
			if block.ToolUseID != "" {
				b.WriteString(m.theme.Muted.Render(fmt.Sprintf(" [%s]", block.ToolUseID)))
			}
			b.WriteString(":\n")
			if block.Content != "" {
				// Truncate very long outputs
				content := block.Content
				if len(content) > 1000 {
					content = content[:997] + "..."
				}
				b.WriteString(content)
			}

		case "tool_use":
			// Render compact tool use indicator with status-aware icon and color
			// The collapsed log view is rendered once at the end of all tool_use blocks
			paramPreview := getToolParamPreview(block.Name, block.Input)
			icon, style := m.getToolStatus(block.ID)
			toolLine := fmt.Sprintf("%s %s(%s)", icon, block.Name, paramPreview)
			b.WriteString(style.Render(toolLine))

		case "text":
			// Regular text block
			b.WriteString(block.Text)
		}

		// Add spacing between blocks
		if i < len(blocks)-1 {
			b.WriteString("\n")
		}
	}

	// Task 7: Only show collapsed preview on the most recent tool with a result
	// Check if any of these blocks contains the most recent tool (cached for performance)
	hasMostRecentTool := false
	if m.mostRecentToolID != "" {
		for _, block := range blocks {
			if block.Type == "tool_use" && block.ID == m.mostRecentToolID {
				hasMostRecentTool = true
				break
			}
		}
	}

	if hasMostRecentTool {
		b.WriteString("\n")
		// Add collapsed tool log (last 3 lines of chunk output)
		collapsedLog, hiddenLines := m.renderCollapsedToolLog()
		if collapsedLog != "" {
			b.WriteString(collapsedLog)
		}
		// Combine "+ N lines" with Ctrl+O hint on one line
		if hiddenLines > 0 {
			b.WriteString(m.theme.Muted.Render(fmt.Sprintf("+ %d lines · Ctrl+O for details", hiddenLines)))
		} else {
			b.WriteString(m.theme.Muted.Render("Ctrl+O for details"))
		}
	}

	return b.String()
}

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

// getToolParamPreview extracts a compact parameter preview for tool display
func getToolParamPreview(toolName string, input map[string]interface{}) string {
	// Get the most relevant parameter based on tool type
	var preview string
	switch toolName {
	case "bash":
		if cmd, ok := input["command"].(string); ok {
			preview = cmd
		}
	case "read_file", "Read":
		if path, ok := input["file_path"].(string); ok {
			preview = path
		} else if path, ok := input["path"].(string); ok {
			preview = path
		}
	case "write_file", "Write", "edit", "Edit":
		if path, ok := input["file_path"].(string); ok {
			preview = path
		}
	case "grep", "Grep":
		if pattern, ok := input["pattern"].(string); ok {
			preview = pattern
		}
	case "glob", "Glob":
		if pattern, ok := input["pattern"].(string); ok {
			preview = pattern
		}
	default:
		// For unknown tools, try to get first string parameter
		for _, val := range input {
			if str, ok := val.(string); ok && str != "" {
				preview = str
				break
			}
		}
	}

	// Truncate and quote the preview
	if preview == "" {
		return ""
	}
	// Escape newlines and truncate
	preview = strings.ReplaceAll(preview, "\n", "\\n")
	preview = strings.ReplaceAll(preview, "\t", "\\t")
	if len(preview) > 40 {
		preview = preview[:37] + "..."
	}
	return fmt.Sprintf("%q", preview)
}

// Phase 6C Task 6: Quick Actions Key Handler

// handleQuickActionsKey is deprecated - quick actions are now handled by huh forms
// This function is kept for backward compatibility but should not be called
// func (m *Model) handleQuickActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
// 	// Quick actions are now handled by huh forms in LaunchQuickActionsForm()
// 	return m, nil
// }

// Phase 4 Task 3: Event Handlers

// handleConversationEvent processes conversation events from the service layer
func (m *Model) handleConversationEvent(msg conversationEventMsg) (tea.Model, tea.Cmd) {
	_, _ = fmt.Fprintf(os.Stderr, "[CONV_EVENT] type=%d, id=%s\n", msg.event.Type, msg.event.Payload.ID)

	// Update UI state based on event type
	switch msg.event.Type {
	case pubsub.Created:
		// New conversation created
		_, _ = fmt.Fprintf(os.Stderr, "[CONV_EVENT] conversation created: %s\n", msg.event.Payload.ID)

	case pubsub.Updated:
		// Conversation updated - refresh if it's the current conversation
		if msg.event.Payload.ID == m.ConversationID {
			_, _ = fmt.Fprintf(os.Stderr, "[CONV_EVENT] current conversation updated\n")
			// Update favorite status
			m.IsFavorite = msg.event.Payload.IsFavorite
			m.updateViewport()
		}

	case pubsub.Deleted:
		// Conversation deleted
		_, _ = fmt.Fprintf(os.Stderr, "[CONV_EVENT] conversation deleted: %s\n", msg.event.Payload.ID)
	}

	// Continue listening for events
	if m.convSvc != nil && m.eventCtx != nil {
		convEvents := m.convSvc.Subscribe(m.eventCtx)
		return m, waitForConversationEvent(convEvents)
	}

	return m, nil
}

// handleMessageEvent processes message events from the service layer
func (m *Model) handleMessageEvent(msg messageEventMsg) (tea.Model, tea.Cmd) {
	_, _ = fmt.Fprintf(os.Stderr, "[MSG_EVENT] type=%d, convID=%s, role=%s\n",
		msg.event.Type, msg.event.Payload.ConversationID, msg.event.Payload.Role)

	// Update UI state based on event type
	switch msg.event.Type {
	case pubsub.Created:
		// New message created - refresh if it's for the current conversation
		if msg.event.Payload.ConversationID == m.ConversationID {
			_, _ = fmt.Fprintf(os.Stderr, "[MSG_EVENT] message added to current conversation\n")
			// Note: We don't add the message here because it's already added
			// during normal flow. This event is mainly for other consumers
			// or for syncing state across multiple UI instances
		}

	case pubsub.Updated:
		// Message updated (less common)
		_, _ = fmt.Fprintf(os.Stderr, "[MSG_EVENT] message updated\n")

	case pubsub.Deleted:
		// Message deleted (less common)
		_, _ = fmt.Fprintf(os.Stderr, "[MSG_EVENT] message deleted\n")
	}

	// Continue listening for events
	if m.msgSvc != nil && m.eventCtx != nil {
		msgEvents := m.msgSvc.Subscribe(m.eventCtx)
		return m, waitForMessageEvent(msgEvents)
	}

	return m, nil
}

// updateHoveredMessage determines which message (if any) the mouse is hovering over
func (m *Model) updateHoveredMessage(x, y int) {
	// The viewport starts after the top status bar (1 line)
	// Y coordinate 0 = top status bar, Y coordinate 1+ = viewport content

	// Check if mouse is in the viewport area (not in header/footer/input)
	// Rough layout: top bar (1), viewport (most), input area (~3-5), footer (1)
	viewportStartY := 1
	viewportEndY := m.Height - 6 // Approximate, leaves room for input and footer

	if y < viewportStartY || y > viewportEndY {
		// Mouse is outside viewport
		m.hoveredMessageIndex = -1
		return
	}

	// Get the viewport's current scroll position
	viewportY := y - viewportStartY + m.Viewport.YOffset

	// Now we need to map viewportY to a message index
	// This requires knowing the line-by-line layout of the viewport content
	// For now, we'll do a simple approach: parse the current viewport content

	// Get all visible messages (excluding tool-only messages)
	var visibleMessages []struct {
		index     int
		timestamp time.Time
		startLine int
		endLine   int
	}

	currentLine := 0

	// Account for intro screen if showing
	if m.ShowIntro {
		introContent := m.renderIntroView()
		introLines := strings.Count(introContent, "\n") + 2 // +2 for spacing
		currentLine += introLines
	}

	// Process each message to determine its line range
	for i := range m.Messages {
		msg := &m.Messages[i]

		// Skip internal messages
		if msg.Role == "tool" {
			continue
		}
		if msg.Role == "user" && msg.Content == "" && len(msg.ContentBlock) > 0 {
			allToolResults := true
			for _, block := range msg.ContentBlock {
				if block.Type != "tool_result" {
					allToolResults = false
					break
				}
			}
			if allToolResults {
				continue
			}
		}

		startLine := currentLine

		// Count lines in this message
		var messageContent string
		if msg.Content == "" && len(msg.ContentBlock) > 0 {
			messageContent = m.renderContentBlocks(msg.ContentBlock)
		} else {
			messageContent = msg.Content
			if msg.Role == "assistant" {
				rendered, err := m.RenderMessage(msg)
				if err == nil {
					messageContent = strings.TrimSpace(rendered)
				}
			}
		}

		if messageContent != "" {
			messageLines := strings.Count(messageContent, "\n") + 1
			currentLine += messageLines

			// Add spacing between messages
			if i < len(m.Messages)-1 {
				currentLine += 1
			}

			visibleMessages = append(visibleMessages, struct {
				index     int
				timestamp time.Time
				startLine int
				endLine   int
			}{
				index:     i,
				timestamp: msg.Timestamp,
				startLine: startLine,
				endLine:   currentLine - 1,
			})
		}
	}

	// Find which message the mouse is hovering over
	m.hoveredMessageIndex = -1
	for _, vm := range visibleMessages {
		if viewportY >= vm.startLine && viewportY <= vm.endLine {
			m.hoveredMessageIndex = vm.index
			m.hoveredMessageTime = vm.timestamp
			return
		}
	}
}
