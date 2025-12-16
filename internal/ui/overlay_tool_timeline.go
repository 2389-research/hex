package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToolTimelineOverlay displays all tool calls in the conversation in chronological order
type ToolTimelineOverlay struct {
	model    *Model          // Reference to get messages and results
	viewport viewport.Model  // Embedded viewport for scrolling
	width    int
	height   int
}

// NewToolTimelineOverlay creates a new tool timeline overlay
func NewToolTimelineOverlay(model *Model) *ToolTimelineOverlay {
	return &ToolTimelineOverlay{
		model:    model,
		viewport: viewport.New(0, 0), // Initialized in OnPush
	}
}

// IsFullscreen returns true (this is a fullscreen overlay)
func (o *ToolTimelineOverlay) IsFullscreen() bool {
	return true
}

// GetDesiredHeight returns -1 (fullscreen wants all available height)
func (o *ToolTimelineOverlay) GetDesiredHeight() int {
	return -1
}

// GetHeader returns the overlay header
func (o *ToolTimelineOverlay) GetHeader() string {
	return "Tool Timeline"
}

// GetContent returns the timeline of all tool calls in the conversation
func (o *ToolTimelineOverlay) GetContent() string {
	var b strings.Builder

	// Count total tool calls
	toolCount := 0
	for _, msg := range o.model.Messages {
		for _, block := range msg.ContentBlock {
			if block.Type == "tool_use" {
				toolCount++
			}
		}
	}

	// Empty state
	if toolCount == 0 {
		return "No tool calls in this conversation"
	}

	// Iterate through all messages and find tool_use blocks
	for _, msg := range o.model.Messages {
		for _, block := range msg.ContentBlock {
			if block.Type == "tool_use" {
				// Format: [HH:MM:SS] <status_icon> toolname("params")
				timestamp := msg.Timestamp.Format("15:04:05")
				icon, style := o.model.getToolStatus(block.ID)
				paramPreview := getToolParamPreview(block.Name, block.Input)

				toolLine := fmt.Sprintf("[%s] %s %s(%s)",
					timestamp,
					icon,
					block.Name,
					paramPreview,
				)
				b.WriteString(style.Render(toolLine))
				b.WriteString("\n")

				// Find the corresponding tool result in history
				var toolOutput string
				var hasResult bool
				for _, tr := range o.model.toolResultHistory {
					if tr.ToolUseID == block.ID {
						hasResult = true
						if tr.Result != nil {
							toolOutput = tr.Result.Output
						}
						break
					}
				}

				// Show output with tree prefix
				if hasResult {
					if toolOutput != "" {
						// Split output into lines and prefix each with └─
						outputLines := strings.Split(strings.TrimRight(toolOutput, "\n"), "\n")
						for i, line := range outputLines {
							if i == 0 {
								b.WriteString("└─ ")
							} else {
								b.WriteString("   ")
							}
							b.WriteString(line)
							b.WriteString("\n")
						}
					} else {
						// Empty output
						b.WriteString("└─ (no output)\n")
					}
				} else {
					// Pending - no result yet
					b.WriteString("└─ (pending approval)\n")
				}

				// Add spacing between tool calls
				b.WriteString("\n")
			}
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

// GetFooter returns the overlay footer
func (o *ToolTimelineOverlay) GetFooter() string {
	// Count total tool calls
	toolCount := 0
	for _, msg := range o.model.Messages {
		for _, block := range msg.ContentBlock {
			if block.Type == "tool_use" {
				toolCount++
			}
		}
	}

	if toolCount == 1 {
		return "1 tool call • Esc to close"
	}
	return fmt.Sprintf("%d tool calls • Esc to close", toolCount)
}

// OnPush initializes the viewport with dimensions
func (o *ToolTimelineOverlay) OnPush(width, height int) {
	o.width = width
	o.height = height

	// Guard against negative dimensions on small terminals
	vw := width - 4
	vh := height - 6
	if vw < 1 {
		vw = 1
	}
	if vh < 1 {
		vh = 1
	}

	// Initialize viewport (leave space for header and footer)
	o.viewport = viewport.New(vw, vh)
	o.viewport.SetContent(o.GetContent())

	// Auto-scroll to bottom to show most recent tools
	o.viewport.GotoBottom()
}

// OnPop cleans up
func (o *ToolTimelineOverlay) OnPop() {
	// Nothing to clean up
}

// SetHeight updates the viewport height
func (o *ToolTimelineOverlay) SetHeight(height int) {
	o.height = height
	vh := height - 6
	if vh < 1 {
		vh = 1
	}
	o.viewport.Height = vh
	o.viewport.SetContent(o.GetContent())
}

// Update handles viewport messages
func (o *ToolTimelineOverlay) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	o.viewport, cmd = o.viewport.Update(msg)

	// Update viewport dimensions and content on window size changes
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		o.width = wsm.Width
		o.height = wsm.Height
		vw := wsm.Width - 4
		vh := wsm.Height - 6
		if vw < 1 {
			vw = 1
		}
		if vh < 1 {
			vh = 1
		}
		o.viewport.Width = vw
		o.viewport.Height = vh
		o.viewport.SetContent(o.GetContent())
	}

	return cmd
}

// HandleKey processes keyboard input
func (o *ToolTimelineOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		// Pop will be handled by caller
		return true, nil

	case tea.KeyUp, tea.KeyDown:
		// Viewport navigation
		cmd := o.Update(msg)
		return true, cmd

	default:
		// Let viewport handle other keys (like PageUp/PageDown, Home/End, etc.)
		// and capture them so they don't leak through
		cmd := o.Update(msg)
		return true, cmd
	}
}

// Render returns the complete fullscreen view
func (o *ToolTimelineOverlay) Render(width, height int) string {
	var b strings.Builder

	// Update content if it changed (e.g., new tool results came in)
	o.viewport.SetContent(o.GetContent())

	// Header with border
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("cyan"))

	closeHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Ctrl+O or Esc to close")

	header := headerStyle.Render(o.GetHeader())
	// Use lipgloss.Width for accurate visual width calculation
	headerWidth := lipgloss.Width(header)
	closeHintWidth := lipgloss.Width(closeHint)
	padding := max(0, width-headerWidth-closeHintWidth-8)
	headerLine := fmt.Sprintf("┏━━ %s %s %s ┓",
		header,
		strings.Repeat("━", padding),
		closeHint)
	b.WriteString(headerLine)
	b.WriteString("\n\n")

	// Viewport content
	b.WriteString(o.viewport.View())
	b.WriteString("\n\n")

	// Footer with border
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	b.WriteString(footerStyle.Render(o.GetFooter()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("┗%s┛", strings.Repeat("━", width-2)))

	return b.String()
}

// Cancel dismisses the tool timeline overlay (no cleanup needed)
func (o *ToolTimelineOverlay) Cancel() tea.Cmd {
	return nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
