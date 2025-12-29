package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ToolTimelineContentProvider provides content for the tool timeline overlay
type ToolTimelineContentProvider struct {
	model *Model // Reference to get messages and results
}

// Header returns the tool timeline header
func (p *ToolTimelineContentProvider) Header() string {
	return "Tool Timeline"
}

// Content returns the timeline of all tool calls in the conversation
func (p *ToolTimelineContentProvider) Content() string {
	var b strings.Builder

	// Count total tool calls
	toolCount := 0
	for _, msg := range p.model.Messages {
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
	for _, msg := range p.model.Messages {
		for _, block := range msg.ContentBlock {
			if block.Type == "tool_use" {
				// Format: [HH:MM:SS] <status_icon> toolname("params")
				timestamp := msg.Timestamp.Format("15:04:05")
				icon, style := p.model.getToolStatus(block.ID)
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
				for _, tr := range p.model.toolResultHistory {
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

// ToggleKeys returns the keys that toggle the tool timeline overlay
func (p *ToolTimelineContentProvider) ToggleKeys() []tea.KeyType {
	return []tea.KeyType{tea.KeyCtrlO}
}

// Footer returns the custom footer with tool count
func (p *ToolTimelineContentProvider) Footer() string {
	// Count total tool calls
	toolCount := 0
	for _, msg := range p.model.Messages {
		for _, block := range msg.ContentBlock {
			if block.Type == "tool_use" {
				toolCount++
			}
		}
	}

	if toolCount == 1 {
		return "1 tool call • Ctrl+O or Esc to close"
	}
	return fmt.Sprintf("%d tool calls • Ctrl+O or Esc to close", toolCount)
}

// NewToolTimelineOverlay creates a new tool timeline overlay using the generic fullscreen wrapper
func NewToolTimelineOverlay(model *Model) *GenericFullscreenOverlay {
	provider := &ToolTimelineContentProvider{model: model}
	overlay := NewGenericFullscreenOverlay(provider)
	return overlay
}

// ToolTimelineOverlayScrollToBottom is a helper to scroll to bottom after creation
// Call this after OnPush to auto-scroll to most recent tools
func ToolTimelineOverlayScrollToBottom(overlay *GenericFullscreenOverlay) {
	overlay.viewport.GotoBottom()
}
