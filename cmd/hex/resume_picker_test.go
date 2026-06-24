// ABOUTME: Tests resume picker setup behavior before the interactive program starts.
// ABOUTME: Verifies the picker receives the bounded recent conversation list.
package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestConversationsForPickerLimitsToTwenty(t *testing.T) {
	conversations := make([]*services.Conversation, 25)
	for i := range conversations {
		conversations[i] = &services.Conversation{
			ID:        fmt.Sprintf("conv-%02d", i),
			Title:     fmt.Sprintf("Conversation %02d", i),
			UpdatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}

	limited := conversationsForPicker(conversations)

	assert.Len(t, limited, 20)
	assert.Equal(t, "conv-00", limited[0].ID)
	assert.Equal(t, "conv-19", limited[19].ID)
}
