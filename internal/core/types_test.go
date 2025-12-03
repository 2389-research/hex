package core_test

import (
	"testing"

	"github.com/2389-research/hex/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestContentBlock(t *testing.T) {
	t.Run("creates text content block", func(t *testing.T) {
		block := core.NewTextBlock("Hello, world!")
		assert.Equal(t, "text", block.Type)
		assert.Equal(t, "Hello, world!", block.Text)
		assert.Nil(t, block.Source)
	})

	t.Run("creates image content block", func(t *testing.T) {
		imgSrc := &core.ImageSource{
			Type:      "base64",
			MediaType: "image/png",
			Data:      "iVBORw0KGgoAAAANSUhEUgAAAAUA",
		}
		block := core.NewImageBlock(imgSrc)
		assert.Equal(t, "image", block.Type)
		assert.Equal(t, "", block.Text)
		assert.NotNil(t, block.Source)
		assert.Equal(t, "base64", block.Source.Type)
		assert.Equal(t, "image/png", block.Source.MediaType)
	})
}

func TestMultiModalMessage(t *testing.T) {
	t.Run("creates message with text and images", func(t *testing.T) {
		blocks := []core.ContentBlock{
			core.NewImageBlock(&core.ImageSource{
				Type:      "base64",
				MediaType: "image/png",
				Data:      "data1",
			}),
			core.NewTextBlock("What's in this image?"),
		}

		msg := core.Message{
			Role:         "user",
			Content:      "", // Empty when using ContentBlocks
			ContentBlock: blocks,
		}

		assert.Equal(t, "user", msg.Role)
		assert.Len(t, msg.ContentBlock, 2)
		assert.Equal(t, "image", msg.ContentBlock[0].Type)
		assert.Equal(t, "text", msg.ContentBlock[1].Type)
	})
}

func TestMessageBackwardCompatibility(t *testing.T) {
	t.Run("text-only message still works", func(t *testing.T) {
		msg := core.Message{
			Role:    "user",
			Content: "Simple text message",
		}

		assert.Equal(t, "Simple text message", msg.Content)
		assert.Nil(t, msg.ContentBlock)
	})
}

func TestMessageJSONSerialization(t *testing.T) {
	t.Run("text-only message serializes content as string", func(t *testing.T) {
		msg := core.Message{
			Role:    "user",
			Content: "Hello, world!",
		}

		data, err := msg.MarshalJSON()
		require.NoError(t, err)

		// Should contain content as a string
		assert.Contains(t, string(data), `"content":"Hello, world!"`)
		assert.Contains(t, string(data), `"role":"user"`)
	})

	t.Run("multimodal message serializes content as array", func(t *testing.T) {
		msg := core.Message{
			Role: "user",
			ContentBlock: []core.ContentBlock{
				core.NewTextBlock("What's in this image?"),
				core.NewImageBlock(&core.ImageSource{
					Type:      "base64",
					MediaType: "image/png",
					Data:      "abc123",
				}),
			},
		}

		data, err := msg.MarshalJSON()
		require.NoError(t, err)

		// Should contain content as an array
		assert.Contains(t, string(data), `"content":[`)
		assert.Contains(t, string(data), `"type":"text"`)
		assert.Contains(t, string(data), `"type":"image"`)
	})

	t.Run("unmarshals text content correctly", func(t *testing.T) {
		jsonData := `{"role":"user","content":"Hello"}`

		var msg core.Message
		err := msg.UnmarshalJSON([]byte(jsonData))
		require.NoError(t, err)

		assert.Equal(t, "user", msg.Role)
		assert.Equal(t, "Hello", msg.Content)
		assert.Nil(t, msg.ContentBlock)
	})

	t.Run("unmarshals array content correctly", func(t *testing.T) {
		jsonData := `{"role":"user","content":[{"type":"text","text":"Hello"}]}`

		var msg core.Message
		err := msg.UnmarshalJSON([]byte(jsonData))
		require.NoError(t, err)

		assert.Equal(t, "user", msg.Role)
		assert.Len(t, msg.ContentBlock, 1)
		assert.Equal(t, "text", msg.ContentBlock[0].Type)
		assert.Equal(t, "Hello", msg.ContentBlock[0].Text)
	})
}
