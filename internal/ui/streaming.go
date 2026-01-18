// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Streaming UI enhancements for displaying streaming responses
// ABOUTME: Provides token rate display, progressive rendering, and typewriter effects
package ui

import (
	"math"
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
	// Phase: Streaming work display
	currentToolCalls []ToolCallProgress
	thinkingActive   bool
	thinkingText     string
	// Thinking spinner animation
	spinnerFrames []string
	spinnerFrame  int
	lastSpinTime  time.Time
	// Order-preserving streaming content
	streamingLines []StreamingLine
}

var (
	streamingIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	tokenRateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	typewriterCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	toolCallStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")). // Purple
			Bold(false)

	thinkingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")). // Cyan
			Italic(true)
)

// ToolCallProgress tracks a tool call in progress
type ToolCallProgress struct {
	Name       string
	ID         string
	InProgress bool
	Complete   bool
}

// StreamingLineType identifies what kind of streaming content this is
type StreamingLineType int

const (
	// StreamingLineThinking represents thinking/reasoning content
	StreamingLineThinking StreamingLineType = iota
	// StreamingLineTool represents tool call content
	StreamingLineTool
	// StreamingLineText represents regular text content
	StreamingLineText
)

// StreamingLine represents a single line/chunk of streaming content with order preserved
type StreamingLine struct {
	Type    StreamingLineType
	Content string
	ToolID  string // For tool lines, the tool ID
}

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
		// Hexagonal/geometric spinner - distinctive for hex branding
		spinnerFrames: []string{"⬡", "⬢", "◇", "◆", "⬡", "⬢", "◇", "◆"},
		spinnerFrame:  0,
		lastSpinTime:  time.Now(),
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
	s.currentToolCalls = []ToolCallProgress{}
	s.thinkingActive = false
	s.thinkingText = ""
	s.streamingLines = nil
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

	// Track in order-preserving list (append to last text line or create new)
	if len(s.streamingLines) > 0 {
		last := &s.streamingLines[len(s.streamingLines)-1]
		if last.Type == StreamingLineText {
			last.Content += chunk
			return
		}
	}
	s.streamingLines = append(s.streamingLines, StreamingLine{
		Type:    StreamingLineText,
		Content: chunk,
	})
}

// AppendThinkingLine adds a thinking line preserving order
func (s *StreamingDisplay) AppendThinkingLine(text string) {
	s.streamingLines = append(s.streamingLines, StreamingLine{
		Type:    StreamingLineThinking,
		Content: text,
	})
	// Also update legacy fields for backward compatibility
	s.thinkingActive = true
	s.thinkingText = text
}

// AppendToolLine adds a tool usage line preserving order
func (s *StreamingDisplay) AppendToolLine(name, id string) {
	s.streamingLines = append(s.streamingLines, StreamingLine{
		Type:    StreamingLineTool,
		Content: name,
		ToolID:  id,
	})
	// Also update legacy fields for backward compatibility
	s.currentToolCalls = append(s.currentToolCalls, ToolCallProgress{
		Name:       name,
		ID:         id,
		InProgress: true,
		Complete:   false,
	})
}

// GetStreamingLines returns all streaming lines in order
func (s *StreamingDisplay) GetStreamingLines() []StreamingLine {
	return s.streamingLines
}

// GetOrderedText returns all text from streaming lines combined
func (s *StreamingDisplay) GetOrderedText() string {
	var b strings.Builder
	for _, line := range s.streamingLines {
		if line.Type == StreamingLineText {
			b.WriteString(line.Content)
		}
	}
	return b.String()
}

// RenderOrderedContent renders streaming lines in order with styles
func (s *StreamingDisplay) RenderOrderedContent() string {
	var parts []string

	for _, line := range s.streamingLines {
		switch line.Type {
		case StreamingLineThinking:
			spinner := s.getSpinnerFrame()
			display := spinner + " " + line.Content
			parts = append(parts, thinkingStyle.Render(display))
		case StreamingLineTool:
			display := "⚙ " + line.Content
			parts = append(parts, toolCallStyle.Render(display))
		case StreamingLineText:
			// Text is rendered separately via GetText() for markdown processing
			// Skip here to avoid double rendering
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n")
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

// RenderStreamingIndicator renders the streaming status indicator with animated effects
func (s *StreamingDisplay) RenderStreamingIndicator() string {
	var b strings.Builder

	if s.waitingForTokens {
		// Show waiting indicator with subtle pulse
		elapsed := time.Since(s.startTime)
		dots := int(elapsed.Milliseconds()/300) % 4

		// Pulse the color slightly for visual interest
		pulseProgress := float64(elapsed.Milliseconds()%1000) / 1000.0
		pulseIntensity := (math.Sin(pulseProgress*2*math.Pi) + 1) / 2
		grayVal := int(100 + pulseIntensity*55) // Range from #64 (100) to #9B (155)
		pulseColor := lipgloss.Color(formatHexGray(grayVal))
		pulseStyle := lipgloss.NewStyle().Foreground(pulseColor).Italic(true)

		b.WriteString(pulseStyle.Render("◉ Waiting" + strings.Repeat(".", dots)))
		return b.String()
	}

	// Show streaming indicator with animated gradient
	elapsed := time.Since(s.startTime)

	// Animate lightning bolt color
	pulseProgress := float64(elapsed.Milliseconds()%500) / 500.0
	pulseIntensity := (math.Sin(pulseProgress*2*math.Pi) + 1) / 2

	// Interpolate between cyan (#56 = 86) and bright cyan
	boltColor := lipgloss.Color(formatHexRGB(86, int(233+pulseIntensity*22), int(253+pulseIntensity*2)))
	boltStyle := lipgloss.NewStyle().Foreground(boltColor).Bold(true)

	b.WriteString(boltStyle.Render("⚡") + " ")
	b.WriteString(streamingIndicatorStyle.Render("Streaming"))

	if s.tokenRate > 0 {
		b.WriteString(" ")
		// Add visual indicator of speed with bar
		speedBar := renderSpeedBar(s.tokenRate)
		b.WriteString(speedBar)
		b.WriteString(tokenRateStyle.Render(" " + formatTokenRateFloat(s.tokenRate) + " tok/s"))
	}

	// Add typewriter mode indicator
	if s.typewriterMode {
		progress := 0
		if len(s.text) > 0 {
			progress = (s.typewriterPos * 100) / len(s.text)
		}
		b.WriteString(tokenRateStyle.Render(
			" [typewriter " + string(rune('0'+progress/10)) + string(rune('0'+progress%10)) + "%]"))
	}

	return b.String()
}

// formatHexGray creates a hex color string for a gray value
func formatHexGray(val int) string {
	if val < 0 {
		val = 0
	}
	if val > 255 {
		val = 255
	}
	return string([]byte{'#', hexDigit(val / 16), hexDigit(val % 16), hexDigit(val / 16), hexDigit(val % 16), hexDigit(val / 16), hexDigit(val % 16)})
}

// formatHexRGB creates a hex color string from RGB values
func formatHexRGB(r, g, b int) string {
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return v
	}
	r, g, b = clamp(r), clamp(g), clamp(b)
	return string([]byte{'#', hexDigit(r / 16), hexDigit(r % 16), hexDigit(g / 16), hexDigit(g % 16), hexDigit(b / 16), hexDigit(b % 16)})
}

// hexDigit converts a value 0-15 to a hex digit
func hexDigit(v int) byte {
	if v < 10 {
		return byte('0' + v)
	}
	return byte('a' + v - 10)
}

// renderSpeedBar creates a visual speed indicator
func renderSpeedBar(rate float64) string {
	// Map rate to 1-5 blocks
	bars := int(rate / 20) // 20 tok/s per bar
	if bars < 1 {
		bars = 1
	}
	if bars > 5 {
		bars = 5
	}

	// Gradient from yellow to green based on speed
	colors := []lipgloss.Color{
		lipgloss.Color("#f1fa8c"), // Yellow (slow)
		lipgloss.Color("#bef992"), // Yellow-green
		lipgloss.Color("#73daca"), // Cyan-green
		lipgloss.Color("#50fa7b"), // Green (fast)
		lipgloss.Color("#50fa7b"), // Green
	}

	var result string
	for i := 0; i < bars; i++ {
		style := lipgloss.NewStyle().Foreground(colors[i])
		result += style.Render("▮")
	}
	// Fill remaining with dim blocks
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44475a"))
	for i := bars; i < 5; i++ {
		result += dimStyle.Render("▯")
	}

	return result
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
	// Manual digit conversion for integer token rates
	return string(rune('0'+int(rate)/100)) + string(rune('0'+(int(rate)%100)/10)) + string(rune('0'+int(rate)%10))
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

// StartToolCall marks a tool call as started
func (s *StreamingDisplay) StartToolCall(id, name string) {
	s.currentToolCalls = append(s.currentToolCalls, ToolCallProgress{
		Name:       name,
		ID:         id,
		InProgress: true,
		Complete:   false,
	})
}

// CompleteToolCall marks a tool call as complete
func (s *StreamingDisplay) CompleteToolCall(id string) {
	for i := range s.currentToolCalls {
		if s.currentToolCalls[i].ID == id {
			s.currentToolCalls[i].InProgress = false
			s.currentToolCalls[i].Complete = true
			break
		}
	}
}

// SetThinking sets the thinking state and text
func (s *StreamingDisplay) SetThinking(active bool, text string) {
	s.thinkingActive = active
	s.thinkingText = text
	if active {
		// Reset spinner when starting to think
		s.lastSpinTime = time.Now()
	}
}

// getSpinnerFrame returns the current spinner frame and advances if enough time has passed
func (s *StreamingDisplay) getSpinnerFrame() string {
	if len(s.spinnerFrames) == 0 {
		return "⠋"
	}

	// Advance frame every 80ms (dots spinner interval)
	now := time.Now()
	if now.Sub(s.lastSpinTime) > 80*time.Millisecond {
		s.spinnerFrame = (s.spinnerFrame + 1) % len(s.spinnerFrames)
		s.lastSpinTime = now
	}

	return s.spinnerFrames[s.spinnerFrame]
}

// GetWorkDisplay returns the current work display (tool calls, thinking)
func (s *StreamingDisplay) GetWorkDisplay() string {
	var parts []string

	// Show thinking if active with animated spinner and pulsing effect
	if s.thinkingActive {
		spinner := s.getSpinnerFrame()

		// Create pulsing color effect for thinking state
		elapsed := time.Since(s.startTime)
		pulseProgress := float64(elapsed.Milliseconds()%1500) / 1500.0
		pulseIntensity := (math.Sin(pulseProgress*2*math.Pi) + 1) / 2

		// Pulse between cyan (#56E0F3) and purple (#BD93F9)
		// Cyan base: R=86, G=224, B=243
		// Purple accent: R=189, G=147, B=249
		r := int(86 + pulseIntensity*(189-86))
		g := int(224 - pulseIntensity*(224-147))
		b := int(243 + pulseIntensity*(249-243))
		pulseColor := lipgloss.Color(formatHexRGB(r, g, b))

		pulseStyle := lipgloss.NewStyle().
			Foreground(pulseColor).
			Italic(true)

		// Build thinking display with brain emoji and animated dots
		dots := int((elapsed.Milliseconds() / 400) % 4)
		dotStr := strings.Repeat("·", dots) + strings.Repeat(" ", 3-dots)

		display := spinner + " 🧠 Thinking" + dotStr
		if s.thinkingText != "" {
			// Show truncated thinking text for context
			thinkText := s.thinkingText
			if len(thinkText) > 40 {
				thinkText = thinkText[:37] + "..."
			}
			display += " — " + thinkText
		}
		parts = append(parts, pulseStyle.Render(display))
	}

	// Show active tool calls with distinctive styling
	for _, tool := range s.currentToolCalls {
		if tool.InProgress {
			// Use gear emoji and pulsing effect for tools
			elapsed := time.Since(s.startTime)
			pulseProgress := float64(elapsed.Milliseconds()%1000) / 1000.0
			pulseIntensity := (math.Sin(pulseProgress*2*math.Pi) + 1) / 2

			// Pulse between purple (#BD93F9) and pink (#FF79C6)
			r := int(189 + pulseIntensity*(255-189))
			g := int(147 - pulseIntensity*(147-121))
			b := int(249 - pulseIntensity*(249-198))
			toolColor := lipgloss.Color(formatHexRGB(r, g, b))

			toolStyle := lipgloss.NewStyle().
				Foreground(toolColor).
				Bold(true)

			display := "⚙ " + tool.Name
			parts = append(parts, toolStyle.Render(display))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n")
}
