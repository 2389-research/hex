// ABOUTME: Tests for suggestion learning system
// ABOUTME: Validates feedback tracking and confidence adjustment logic

package suggestions

import (
	"testing"
)

func TestLearner_RecordFeedback(t *testing.T) {
	learner := NewLearner()

	// Record some feedback
	learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	learner.RecordFeedback("bash", "run_command", FeedbackRejected)
	learner.RecordFeedback("grep", "search_intent", FeedbackIgnored)

	// Check history
	history := learner.GetRecentHistory(10)
	if len(history) != 3 {
		t.Errorf("Expected 3 history events, got %d", len(history))
	}

	// Verify events
	if history[0].ToolName != "read_file" {
		t.Errorf("Expected first event for read_file, got %s", history[0].ToolName)
	}
	if history[0].Feedback != FeedbackAccepted {
		t.Errorf("Expected first event to be accepted")
	}
}

func TestLearner_AdjustSuggestion_Accepted(t *testing.T) {
	learner := NewLearner()

	// Record multiple acceptances for a tool
	for i := 0; i < 5; i++ {
		learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	}

	// Create a suggestion
	suggestion := Suggestion{
		ToolName:   "read_file",
		Confidence: 0.80,
	}

	// Apply adjustment
	learner.AdjustSuggestion(&suggestion)

	// Confidence should increase
	if suggestion.Confidence <= 0.80 {
		t.Errorf("Expected confidence to increase from 0.80, got %.2f", suggestion.Confidence)
	}

	// Should still be within valid range
	if suggestion.Confidence > 1.0 {
		t.Errorf("Confidence exceeded 1.0: %.2f", suggestion.Confidence)
	}
}

func TestLearner_AdjustSuggestion_Rejected(t *testing.T) {
	learner := NewLearner()

	// Record multiple rejections for a tool
	for i := 0; i < 5; i++ {
		learner.RecordFeedback("bash", "shell_command_like", FeedbackRejected)
	}

	// Create a suggestion
	suggestion := Suggestion{
		ToolName:   "bash",
		Confidence: 0.80,
	}

	// Apply adjustment
	learner.AdjustSuggestion(&suggestion)

	// Confidence should decrease
	if suggestion.Confidence >= 0.80 {
		t.Errorf("Expected confidence to decrease from 0.80, got %.2f", suggestion.Confidence)
	}

	// Should not go below 0
	if suggestion.Confidence < 0.0 {
		t.Errorf("Confidence went negative: %.2f", suggestion.Confidence)
	}
}

func TestLearner_AdjustSuggestion_Ignored(t *testing.T) {
	learner := NewLearner()

	// Record ignores (less impact than rejections)
	for i := 0; i < 5; i++ {
		learner.RecordFeedback("grep", "search_intent", FeedbackIgnored)
	}

	// Create a suggestion
	suggestion := Suggestion{
		ToolName:   "grep",
		Confidence: 0.75,
	}

	// Apply adjustment
	learner.AdjustSuggestion(&suggestion)

	// Confidence should decrease slightly
	if suggestion.Confidence >= 0.75 {
		t.Errorf("Expected confidence to decrease from 0.75, got %.2f", suggestion.Confidence)
	}

	// Should be clamped to valid range
	if suggestion.Confidence < 0.0 || suggestion.Confidence > 1.0 {
		t.Errorf("Confidence out of range: %.2f", suggestion.Confidence)
	}
}

func TestLearner_AdjustmentClamping(t *testing.T) {
	learner := NewLearner()

	// Try to exceed max positive adjustment
	for i := 0; i < 50; i++ {
		learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	}

	adjustments := learner.GetAdjustments()
	adjustment := adjustments["read_file"]

	if adjustment > 0.2 {
		t.Errorf("Adjustment exceeded max 0.2: %.2f", adjustment)
	}

	// Try to exceed max negative adjustment
	learner2 := NewLearner()
	for i := 0; i < 50; i++ {
		learner2.RecordFeedback("bash", "shell_command_like", FeedbackRejected)
	}

	adjustments2 := learner2.GetAdjustments()
	adjustment2 := adjustments2["bash"]

	if adjustment2 < -0.2 {
		t.Errorf("Adjustment exceeded min -0.2: %.2f", adjustment2)
	}
}

func TestLearner_HistoryLimit(t *testing.T) {
	learner := NewLearner()
	learner.maxHistory = 10 // Set small limit for testing

	// Add more events than the limit
	for i := 0; i < 20; i++ {
		learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	}

	history := learner.GetRecentHistory(100)

	if len(history) > learner.maxHistory {
		t.Errorf("History exceeded max: got %d, max %d", len(history), learner.maxHistory)
	}

	if len(history) != 10 {
		t.Errorf("Expected history to be trimmed to 10, got %d", len(history))
	}
}

func TestLearner_GetStats(t *testing.T) {
	learner := NewLearner()

	// Record mixed feedback for a tool
	learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	learner.RecordFeedback("read_file", "absolute_path", FeedbackRejected)
	learner.RecordFeedback("read_file", "absolute_path", FeedbackIgnored)

	stats := learner.GetStats()

	readFileStats, exists := stats["read_file"]
	if !exists {
		t.Fatal("Expected stats for read_file")
	}

	if readFileStats.Total != 5 {
		t.Errorf("Expected 5 total events, got %d", readFileStats.Total)
	}

	if readFileStats.Accepted != 3 {
		t.Errorf("Expected 3 accepted, got %d", readFileStats.Accepted)
	}

	if readFileStats.Rejected != 1 {
		t.Errorf("Expected 1 rejected, got %d", readFileStats.Rejected)
	}

	if readFileStats.Ignored != 1 {
		t.Errorf("Expected 1 ignored, got %d", readFileStats.Ignored)
	}

	expectedRate := 3.0 / 5.0
	if readFileStats.AcceptanceRate != expectedRate {
		t.Errorf("Expected acceptance rate %.2f, got %.2f", expectedRate, readFileStats.AcceptanceRate)
	}
}

func TestLearner_GetStats_MultiplTools(t *testing.T) {
	learner := NewLearner()

	// Record feedback for different tools
	learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	learner.RecordFeedback("bash", "run_command", FeedbackRejected)
	learner.RecordFeedback("grep", "search_intent", FeedbackAccepted)
	learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)

	stats := learner.GetStats()

	if len(stats) != 3 {
		t.Errorf("Expected stats for 3 tools, got %d", len(stats))
	}

	// Verify each tool has stats
	if _, exists := stats["read_file"]; !exists {
		t.Error("Missing stats for read_file")
	}
	if _, exists := stats["bash"]; !exists {
		t.Error("Missing stats for bash")
	}
	if _, exists := stats["grep"]; !exists {
		t.Error("Missing stats for grep")
	}
}

func TestLearner_ClearHistory(t *testing.T) {
	learner := NewLearner()

	// Add some feedback
	learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	learner.RecordFeedback("bash", "run_command", FeedbackRejected)

	// Verify it exists
	if len(learner.GetRecentHistory(10)) != 2 {
		t.Fatal("Expected 2 history events before clear")
	}

	// Clear
	learner.ClearHistory()

	// Verify cleared
	if len(learner.GetRecentHistory(10)) != 0 {
		t.Error("Expected 0 history events after clear")
	}

	if len(learner.GetAdjustments()) != 0 {
		t.Error("Expected 0 adjustments after clear")
	}
}

func TestLearner_GetRecentHistory_Limit(t *testing.T) {
	learner := NewLearner()

	// Add 10 events
	for i := 0; i < 10; i++ {
		learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	}

	// Request only 5
	history := learner.GetRecentHistory(5)

	if len(history) != 5 {
		t.Errorf("Expected 5 recent events, got %d", len(history))
	}
}

func TestLearner_GetRecentHistory_LessThanLimit(t *testing.T) {
	learner := NewLearner()

	// Add 3 events
	for i := 0; i < 3; i++ {
		learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
	}

	// Request 10 (more than available)
	history := learner.GetRecentHistory(10)

	if len(history) != 3 {
		t.Errorf("Expected 3 events, got %d", len(history))
	}
}

func TestLearner_DecayFactor(t *testing.T) {
	learner := NewLearner()

	// Record one acceptance
	learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)

	initialAdjustment := learner.GetAdjustments()["read_file"]

	// Record several ignores (which decay the adjustment)
	for i := 0; i < 5; i++ {
		learner.RecordFeedback("read_file", "absolute_path", FeedbackIgnored)
	}

	finalAdjustment := learner.GetAdjustments()["read_file"]

	// Adjustment should have decayed and gone negative
	if finalAdjustment >= initialAdjustment {
		t.Errorf("Expected adjustment to decay from %.3f, got %.3f",
			initialAdjustment, finalAdjustment)
	}
}

func TestLearner_ThreadSafety(t *testing.T) {
	learner := NewLearner()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			learner.RecordFeedback("read_file", "absolute_path", FeedbackAccepted)
			learner.GetStats()
			learner.GetRecentHistory(5)
			suggestion := Suggestion{ToolName: "read_file", Confidence: 0.8}
			learner.AdjustSuggestion(&suggestion)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without panicking, thread safety works
	stats := learner.GetStats()
	if stats["read_file"].Total != 10 {
		t.Errorf("Expected 10 events after concurrent writes, got %d", stats["read_file"].Total)
	}
}

func TestLearner_NoAdjustmentForUnknownTool(t *testing.T) {
	learner := NewLearner()

	// Don't record any feedback for this tool
	suggestion := Suggestion{
		ToolName:   "unknown_tool",
		Confidence: 0.75,
	}

	original := suggestion.Confidence
	learner.AdjustSuggestion(&suggestion)

	// Should remain unchanged
	if suggestion.Confidence != original {
		t.Errorf("Confidence changed for unknown tool: %.2f -> %.2f",
			original, suggestion.Confidence)
	}
}
