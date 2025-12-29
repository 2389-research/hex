package ui

import tea "github.com/charmbracelet/bubbletea"

// HelpContentProvider provides content for the help overlay
type HelpContentProvider struct{}

// Header returns the help overlay header
func (p *HelpContentProvider) Header() string {
	return "Help & Keyboard Shortcuts"
}

// Content returns the help text
func (p *HelpContentProvider) Content() string {
	return `# Keyboard Shortcuts

## Navigation
- **↑/↓**: Scroll viewport
- **PageUp/PageDown**: Page up/down
- **Ctrl+D/U**: Half page down/up
- **Home/End**: Go to top/bottom

## Overlays
- **Ctrl+O**: Toggle tool timeline (all tool calls)
- **Ctrl+H**: Toggle this help screen
- **Ctrl+R**: Open conversation history
- **Escape**: Close active overlay

## Input
- **Enter**: Send message
- **Shift+Enter**: New line in message
- **Ctrl+C**: Cancel stream or close overlay

## Tools
- **Enter**: Approve tool (when prompted)
- **Escape**: Deny tool (when prompted)

## Other
- **Ctrl+L**: Clear screen
- **Ctrl+C twice**: Quit application (when idle)

# Tips

- Use overlays to view detailed information without losing context
- All overlays are scrollable with arrow keys and PageUp/PageDown
- Tool output log shows the last 10,000 lines of tool execution
- Conversation history is limited to the last 1,000 messages
`
}

// ToggleKeys returns the keys that toggle the help overlay
// Both the header close hint AND footer are auto-generated from this
func (p *HelpContentProvider) ToggleKeys() []tea.KeyType {
	return []tea.KeyType{tea.KeyCtrlH}
}

// NewHelpOverlay creates a new help overlay using the generic fullscreen wrapper
func NewHelpOverlay() *GenericFullscreenOverlay {
	provider := &HelpContentProvider{}
	return NewGenericFullscreenOverlay(provider)
}
