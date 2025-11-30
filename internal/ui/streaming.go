// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Streaming UI enhancements for displaying streaming responses
// ABOUTME: Provides token rate display, progressive rendering, and typewriter effects
package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// StreamingDisplay manages streaming text display with enhancements
type StreamingDisplay struct {
	text             string
	lastUpdateTime   time.Time
	tokensReceived   int
	tokenRate        float64
	typewriterMode   bool
	typewriterPos    int
	typewriterSpeed  time.Duration
	waitingForTokens bool
	startTime        time.Time
}

var (
	streamingIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	tokenRateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	waitingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true)

	typewriterCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)
)

// NewStreamingDisplay creates a new streaming display manager
func NewStreamingDisplay() *StreamingDisplay {
	return &StreamingDisplay{
		text:             "",
		lastUpdateTime:   time.Now(),
		tokensReceived:   0,
		tokenRate:        0,
		typewriterMode:   false,
		typewriterPos:    0,
		typewriterSpeed:  30 * time.Millisecond,
		waitingForTokens: true,
		startTime:        time.Now(),
	}
}

// Reset clears the streaming display
func (s *StreamingDisplay) Reset() {
	s.text = ""
	s.lastUpdateTime = time.Now()
	s.tokensReceived = 0
	s.tokenRate = 0
	s.typewriterPos = 0
	s.waitingForTokens = true
	s.startTime = time.Now()
}

// AppendText adds new text to the streaming buffer
func (s *StreamingDisplay) AppendText(chunk string) {
	now := time.Now()
	s.text += chunk
	s.tokensReceived++
	s.waitingForTokens = false

	// Calculate token rate (exponential moving average)
	elapsed := now.Sub(s.lastUpdateTime).Seconds()
	if elapsed > 0 {
		instantRate := 1.0 / elapsed
		if s.tokenRate == 0 {
			s.tokenRate = instantRate
		} else {
			// EMA with alpha = 0.3
			s.tokenRate = 0.3*instantRate + 0.7*s.tokenRate
		}
	}

	s.lastUpdateTime = now

	// In typewriter mode, position will be advanced by Update()
	// (no action needed here)
}

// GetText returns the current text (respecting typewriter mode)
func (s *StreamingDisplay) GetText() string {
	if s.typewriterMode {
		if s.typewriterPos < len(s.text) {
			return s.text[:s.typewriterPos]
		}
	}
	return s.text
}

// GetFullText returns all text regardless of typewriter position
func (s *StreamingDisplay) GetFullText() string {
	return s.text
}

// SetTypewriterMode enables or disables typewriter effect
func (s *StreamingDisplay) SetTypewriterMode(enabled bool) {
	s.typewriterMode = enabled
	if enabled && s.typewriterPos == 0 {
		s.typewriterPos = 0
	}
}

// ToggleTypewriterMode toggles typewriter mode on/off
func (s *StreamingDisplay) ToggleTypewriterMode() {
	s.SetTypewriterMode(!s.typewriterMode)
}

// IsTypewriterMode returns whether typewriter mode is active
func (s *StreamingDisplay) IsTypewriterMode() bool {
	return s.typewriterMode
}

// AdvanceTypewriter advances the typewriter position
func (s *StreamingDisplay) AdvanceTypewriter() {
	if s.typewriterMode && s.typewriterPos < len(s.text) {
		// Advance by 1-3 characters depending on what's next
		advance := 1
		if s.typewriterPos < len(s.text)-1 {
			// Skip faster over whitespace
			if s.text[s.typewriterPos] == ' ' || s.text[s.typewriterPos] == '\n' {
				advance = 2
			}
		}
		s.typewriterPos += advance
	}
}

// GetTokenRate returns the current token rate (tokens/second)
func (s *StreamingDisplay) GetTokenRate() float64 {
	return s.tokenRate
}

// GetTokenCount returns total tokens received
func (s *StreamingDisplay) GetTokenCount() int {
	return s.tokensReceived
}

// IsWaitingForTokens returns whether we're still waiting for the first token
func (s *StreamingDisplay) IsWaitingForTokens() bool {
	return s.waitingForTokens
}

// GetElapsedTime returns time since streaming started
func (s *StreamingDisplay) GetElapsedTime() time.Duration {
	return time.Since(s.startTime)
}

// RenderStreamingIndicator renders the streaming status indicator
func (s *StreamingDisplay) RenderStreamingIndicator() string {
	var b strings.Builder

	if s.waitingForTokens {
		// Show waiting indicator
		elapsed := time.Since(s.startTime)
		dots := int(elapsed.Milliseconds()/300) % 4
		b.WriteString(waitingStyle.Render("Waiting for response" + strings.Repeat(".", dots)))
		return b.String()
	}

	// Show streaming indicator with token rate
	b.WriteString(streamingIndicatorStyle.Render("⚡ Streaming"))

	if s.tokenRate > 0 {
		b.WriteString(" ")
		b.WriteString(tokenRateStyle.Render(
			lipgloss.NewStyle().String() + "(" + formatTokenRateFloat(s.tokenRate) + " tok/s)"))
	}

	// Add typewriter mode indicator
	if s.typewriterMode {
		progress := 0
		if len(s.text) > 0 {
			progress = (s.typewriterPos * 100) / len(s.text)
		}
		b.WriteString(tokenRateStyle.Render(
			lipgloss.NewStyle().String() + " [typewriter " + lipgloss.NewStyle().String() + string(rune('0'+progress/10)) + string(rune('0'+progress%10)) + "%]"))
	}

	return b.String()
}

// RenderWithCursor renders text with a typewriter cursor if in typewriter mode
func (s *StreamingDisplay) RenderWithCursor(text string) string {
	if !s.typewriterMode {
		return text
	}

	if s.typewriterPos < len(text) {
		// Add blinking cursor at typewriter position
		before := text[:s.typewriterPos]
		cursor := typewriterCursorStyle.Render("▋")
		return before + cursor
	}

	return text
}

// formatTokenRateFloat formats a floating-point token rate
func formatTokenRateFloat(rate float64) string {
	if rate < 1 {
		return "<1"
	}
	if rate > 999 {
		return "999+"
	}
	return lipgloss.NewStyle().String() + string(rune('0'+int(rate)/100)) + string(rune('0'+(int(rate)%100)/10)) + string(rune('0'+int(rate)%10))
}

// StreamingStats provides statistics about the streaming session
type StreamingStats struct {
	TotalTokens   int
	AverageRate   float64
	CurrentRate   float64
	Duration      time.Duration
	TypewriterPos int
	TotalChars    int
}

// GetStats returns current streaming statistics
func (s *StreamingDisplay) GetStats() StreamingStats {
	duration := time.Since(s.startTime)
	avgRate := 0.0
	if duration.Seconds() > 0 {
		avgRate = float64(s.tokensReceived) / duration.Seconds()
	}

	return StreamingStats{
		TotalTokens:   s.tokensReceived,
		AverageRate:   avgRate,
		CurrentRate:   s.tokenRate,
		Duration:      duration,
		TypewriterPos: s.typewriterPos,
		TotalChars:    len(s.text),
	}
}

// ProgressiveMarkdownRenderer handles progressive rendering of markdown during streaming
type ProgressiveMarkdownRenderer struct {
	lastRenderedLength int
	incompleteBuffer   string
}

// NewProgressiveMarkdownRenderer creates a new progressive markdown renderer
func NewProgressiveMarkdownRenderer() *ProgressiveMarkdownRenderer {
	return &ProgressiveMarkdownRenderer{
		lastRenderedLength: 0,
		incompleteBuffer:   "",
	}
}

// ShouldRender determines if we should re-render based on new content
func (r *ProgressiveMarkdownRenderer) ShouldRender(text string) bool {
	// Re-render when:
	// 1. Text length increased significantly (>10 chars)
	// 2. A complete block was received (ends with \n\n)
	// 3. A complete code block or list was received

	if len(text) < r.lastRenderedLength {
		// Reset
		r.lastRenderedLength = 0
		return true
	}

	delta := len(text) - r.lastRenderedLength

	// Re-render on significant text addition
	if delta > 10 {
		return true
	}

	// Re-render on complete blocks
	if strings.HasSuffix(text, "\n\n") {
		return true
	}

	// Re-render on complete code blocks
	if strings.Contains(text[r.lastRenderedLength:], "```") {
		return true
	}

	return false
}

// UpdateRendered updates the last rendered length
func (r *ProgressiveMarkdownRenderer) UpdateRendered(textLength int) {
	r.lastRenderedLength = textLength
}

// Reset resets the renderer state
func (r *ProgressiveMarkdownRenderer) Reset() {
	r.lastRenderedLength = 0
	r.incompleteBuffer = ""
}
