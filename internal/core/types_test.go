package core_test

import (
	"testing"

	"github.com/harper/clem/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	msg := core.Message{
		Role:    "user",
		Content: "Hello",
	}

	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "Hello", msg.Content)
}

func TestMessageRole(t *testing.T) {
	tests := []struct {
		role  string
		valid bool
	}{
		{"user", true},
		{"assistant", true},
		{"system", true},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			msg := core.Message{Role: tt.role}
			err := msg.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
