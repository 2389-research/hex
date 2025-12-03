// ABOUTME: UI integration for smart suggestion system
// ABOUTME: Wraps internal/suggestions for use in UI model

package ui

import (
	"github.com/2389-research/hex/internal/suggestions"
)

// Suggestion is an alias for suggestions.Suggestion
type Suggestion = suggestions.Suggestion

// SuggestionDetector is an alias for suggestions.Detector
type SuggestionDetector = suggestions.Detector

// SuggestionLearner is an alias for suggestions.Learner
type SuggestionLearner = suggestions.Learner

// NewSuggestionDetector creates a new detector
func NewSuggestionDetector() *SuggestionDetector {
	return suggestions.NewDetector()
}

// NewSuggestionLearner creates a new learner
func NewSuggestionLearner() *SuggestionLearner {
	return suggestions.NewLearner()
}

// FeedbackType represents user feedback
type FeedbackType = suggestions.FeedbackType

// Feedback constants
const (
	FeedbackAccepted = suggestions.FeedbackAccepted
	FeedbackRejected = suggestions.FeedbackRejected
	FeedbackIgnored  = suggestions.FeedbackIgnored
)
