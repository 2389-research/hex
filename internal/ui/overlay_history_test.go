package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHistoryOverlay_IsFullscreen(t *testing.T) {
	messages := []Message{}
	overlay := NewHistoryOverlay(&messages)
	assert.True(t, overlay.IsFullscreen())
}

func TestHistoryOverlay_RefersToModelMessages(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "Hello", Timestamp: time.Now()},
	}
	overlay := NewHistoryOverlay(&messages)

	// Should reference messages, not copy
	messages = append(messages, Message{
		Role:      "assistant",
		Content:   "Hi there",
		Timestamp: time.Now(),
	})

	content := overlay.GetContent()
	assert.Contains(t, content, "Hello")
	assert.Contains(t, content, "Hi there")
}
