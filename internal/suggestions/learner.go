// ABOUTME: Learning system for tracking and adapting to user behavior
// ABOUTME: Adjusts suggestion confidence based on acceptance/rejection patterns

package suggestions

import (
	"sync"
	"time"
)

// FeedbackType represents user feedback on a suggestion
type FeedbackType int

const (
	FeedbackAccepted FeedbackType = iota
	FeedbackRejected
	FeedbackIgnored
)

// FeedbackEvent records user interaction with a suggestion
type FeedbackEvent struct {
	ToolName  string
	Pattern   string // Pattern that triggered the suggestion
	Feedback  FeedbackType
	Timestamp time.Time
}

// Learner tracks suggestion effectiveness and adjusts confidence
type Learner struct {
	mu            sync.RWMutex
	history       []FeedbackEvent
	adjustments   map[string]float64 // tool name -> confidence adjustment (-0.2 to +0.2)
	maxHistory    int
	decayFactor   float64 // How much to decay old adjustments (0.0 - 1.0)
}

// NewLearner creates a new learning system
func NewLearner() *Learner {
	return &Learner{
		history:     make([]FeedbackEvent, 0),
		adjustments: make(map[string]float64),
		maxHistory:  100,      // Keep last 100 events
		decayFactor: 0.95,     // Slowly decay adjustments
	}
}

// RecordFeedback records user feedback on a suggestion
func (l *Learner) RecordFeedback(toolName string, pattern string, feedback FeedbackType) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Add to history
	event := FeedbackEvent{
		ToolName:  toolName,
		Pattern:   pattern,
		Feedback:  feedback,
		Timestamp: time.Now(),
	}
	l.history = append(l.history, event)

	// Trim history if too large
	if len(l.history) > l.maxHistory {
		l.history = l.history[len(l.history)-l.maxHistory:]
	}

	// Update adjustment for this tool
	l.updateAdjustment(toolName, feedback)
}

// updateAdjustment calculates confidence adjustment based on feedback
func (l *Learner) updateAdjustment(toolName string, feedback FeedbackType) {
	current := l.adjustments[toolName]

	// Decay existing adjustment
	current *= l.decayFactor

	// Apply feedback
	switch feedback {
	case FeedbackAccepted:
		// Increase confidence slightly
		current += 0.02
	case FeedbackRejected:
		// Decrease confidence
		current -= 0.05
	case FeedbackIgnored:
		// Slight decrease for ignored suggestions
		current -= 0.01
	}

	// Clamp to reasonable range [-0.2, +0.2]
	if current > 0.2 {
		current = 0.2
	}
	if current < -0.2 {
		current = -0.2
	}

	l.adjustments[toolName] = current
}

// AdjustSuggestion applies learned adjustments to a suggestion
func (l *Learner) AdjustSuggestion(suggestion *Suggestion) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	adjustment, exists := l.adjustments[suggestion.ToolName]
	if exists {
		suggestion.Confidence += adjustment

		// Clamp to valid confidence range [0.0, 1.0]
		if suggestion.Confidence > 1.0 {
			suggestion.Confidence = 1.0
		}
		if suggestion.Confidence < 0.0 {
			suggestion.Confidence = 0.0
		}
	}
}

// GetStats returns statistics about suggestion effectiveness
func (l *Learner) GetStats() map[string]ToolStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := make(map[string]ToolStats)

	for _, event := range l.history {
		s, exists := stats[event.ToolName]
		if !exists {
			s = ToolStats{ToolName: event.ToolName}
		}

		s.Total++
		switch event.Feedback {
		case FeedbackAccepted:
			s.Accepted++
		case FeedbackRejected:
			s.Rejected++
		case FeedbackIgnored:
			s.Ignored++
		}

		stats[event.ToolName] = s
	}

	// Calculate acceptance rates
	for tool, s := range stats {
		if s.Total > 0 {
			s.AcceptanceRate = float64(s.Accepted) / float64(s.Total)
		}
		s.ConfidenceAdjustment = l.adjustments[tool]
		stats[tool] = s
	}

	return stats
}

// ToolStats contains statistics for a specific tool
type ToolStats struct {
	ToolName             string
	Total                int
	Accepted             int
	Rejected             int
	Ignored              int
	AcceptanceRate       float64
	ConfidenceAdjustment float64
}

// ClearHistory clears all feedback history and adjustments
func (l *Learner) ClearHistory() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.history = make([]FeedbackEvent, 0)
	l.adjustments = make(map[string]float64)
}

// GetRecentHistory returns recent feedback events (max 20)
func (l *Learner) GetRecentHistory(limit int) []FeedbackEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	start := len(l.history) - limit
	if start < 0 {
		start = 0
	}

	// Return a copy to avoid race conditions
	history := make([]FeedbackEvent, len(l.history[start:]))
	copy(history, l.history[start:])
	return history
}

// GetAdjustments returns current confidence adjustments for all tools
func (l *Learner) GetAdjustments() map[string]float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Return a copy
	adjustments := make(map[string]float64)
	for k, v := range l.adjustments {
		adjustments[k] = v
	}
	return adjustments
}
