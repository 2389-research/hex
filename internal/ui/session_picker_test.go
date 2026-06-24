// ABOUTME: Tests for the resume-session picker model.
// ABOUTME: Verifies conversation selection and cancellation behavior without an interactive terminal.
package ui

import (
	"testing"
	"time"

	"github.com/2389-research/hex/internal/services"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionPickerSelectsHighlightedConversation(t *testing.T) {
	conversations := []*services.Conversation{
		{
			ID:        "conv-first",
			Title:     "First conversation",
			UpdatedAt: time.Now(),
		},
	}
	picker := NewSessionPicker(conversations)

	updated, cmd := picker.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result, ok := updated.(SessionPicker)
	require.True(t, ok)

	assert.Equal(t, "conv-first", result.GetSelectedID())
	assert.NotNil(t, cmd)
	assert.Equal(t, "", result.View())
}

func TestSessionPickerCancelLeavesSelectionEmpty(t *testing.T) {
	picker := NewSessionPicker([]*services.Conversation{
		{
			ID:        "conv-first",
			Title:     "First conversation",
			UpdatedAt: time.Now(),
		},
	})

	updated, cmd := picker.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result, ok := updated.(SessionPicker)
	require.True(t, ok)

	assert.Equal(t, "", result.GetSelectedID())
	assert.NotNil(t, cmd)
	assert.Equal(t, "", result.View())
}
