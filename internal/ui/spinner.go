// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Spinner component for showing loading and tool execution states
// ABOUTME: Provides animated spinners with different styles for various operations
package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerType represents different types of spinners
type SpinnerType int

const (
	// SpinnerTypeDefault is the standard loading spinner
	SpinnerTypeDefault SpinnerType = iota
	// SpinnerTypeToolExecution is used when executing tools
	SpinnerTypeToolExecution
	// SpinnerTypeStreaming is used when streaming responses
	SpinnerTypeStreaming
	// SpinnerTypeLoading is used for general loading states
	SpinnerTypeLoading
)

// ToolSpinner manages spinner state for tool execution
type ToolSpinner struct {
	spinner     spinner.Model
	spinnerType SpinnerType
	message     string
	active      bool
	startTime   time.Time
	tokenRate   int // tokens per second for streaming
	tokensCount int
	style       lipgloss.Style
}

var (
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	toolExecutingSpinnerStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("214"))

	streamingSpinnerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86"))

	loadingSpinnerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))
)

// NewToolSpinner creates a new spinner for tool execution
func NewToolSpinner() *ToolSpinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return &ToolSpinner{
		spinner:     s,
		spinnerType: SpinnerTypeDefault,
		active:      false,
		style:       spinnerStyle,
	}
}

// Start activates the spinner with a message
func (s *ToolSpinner) Start(spinnerType SpinnerType, message string) tea.Cmd {
	s.active = true
	s.spinnerType = spinnerType
	s.message = message
	s.startTime = time.Now()
	s.tokensCount = 0
	s.tokenRate = 0

	// Set spinner style based on type
	switch spinnerType {
	case SpinnerTypeToolExecution:
		s.spinner.Spinner = spinner.Points
		s.style = toolExecutingSpinnerStyle
	case SpinnerTypeStreaming:
		s.spinner.Spinner = spinner.MiniDot
		s.style = streamingSpinnerStyle
	case SpinnerTypeLoading:
		s.spinner.Spinner = spinner.Line
		s.style = loadingSpinnerStyle
	default:
		s.spinner.Spinner = spinner.Dot
		s.style = spinnerStyle
	}
	s.spinner.Style = s.style

	return s.spinner.Tick
}

// Stop deactivates the spinner
func (s *ToolSpinner) Stop() {
	s.active = false
	s.message = ""
}

// Update processes spinner tick messages
func (s *ToolSpinner) Update(msg tea.Msg) tea.Cmd {
	if !s.active {
		return nil
	}

	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return cmd
}

// UpdateTokens updates token count and calculates rate (for streaming)
func (s *ToolSpinner) UpdateTokens(count int) {
	s.tokensCount = count
	elapsed := time.Since(s.startTime).Seconds()
	if elapsed > 0 {
		s.tokenRate = int(float64(count) / elapsed)
	}
}

// View renders the spinner
func (s *ToolSpinner) View() string {
	if !s.active {
		return ""
	}

	view := s.spinner.View() + " " + s.message

	// Add duration for tool execution
	if s.spinnerType == SpinnerTypeToolExecution {
		elapsed := time.Since(s.startTime)
		if elapsed > time.Second {
			view += lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Render(" (" + elapsed.Round(100*time.Millisecond).String() + ")")
		}
	}

	// Add token rate for streaming
	if s.spinnerType == SpinnerTypeStreaming && s.tokenRate > 0 {
		view += lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Render(" (" + formatTokenRate(s.tokenRate) + ")")
	}

	return view
}

// IsActive returns whether the spinner is currently active
func (s *ToolSpinner) IsActive() bool {
	return s.active
}

// GetElapsed returns the elapsed time since spinner started
func (s *ToolSpinner) GetElapsed() time.Duration {
	if !s.active {
		return 0
	}
	return time.Since(s.startTime)
}

// formatTokenRate formats token rate for display
func formatTokenRate(rate int) string {
	if rate == 0 {
		return "..."
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Render(lipgloss.NewStyle().String() +
			lipgloss.NewStyle().Bold(true).Render(string(rune('0'+rate/10))) +
			string(rune('0'+rate%10)) + " tok/s")
}

// ProgressIndicator represents different states of operation progress
type ProgressIndicator struct {
	State   ProgressState
	Message string
}

// ProgressState represents the state of an operation
type ProgressState int

const (
	// StateQueued indicates a tool is waiting to execute
	StateQueued ProgressState = iota
	// StateRunning indicates a tool is currently executing
	StateRunning
	// StateCompleted indicates a tool has finished execution
	StateCompleted
	// StateFailed indicates a tool execution failed
	StateFailed
)

var (
	queuedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("35"))

	failedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// NewProgressIndicator creates a new progress indicator
func NewProgressIndicator(state ProgressState, message string) *ProgressIndicator {
	return &ProgressIndicator{
		State:   state,
		Message: message,
	}
}

// View renders the progress indicator
func (p *ProgressIndicator) View() string {
	var icon string
	var style lipgloss.Style

	switch p.State {
	case StateQueued:
		icon = "⋯"
		style = queuedStyle
	case StateRunning:
		icon = "⣾"
		style = runningStyle
	case StateCompleted:
		icon = "✓"
		style = completedStyle
	case StateFailed:
		icon = "✗"
		style = failedStyle
	}

	return style.Render(icon + " " + p.Message)
}
