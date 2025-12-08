// ABOUTME: SSE stream parsing for OpenAI Chat Completions streaming
// ABOUTME: Handles delta chunks with choices array format
package openai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/2389-research/hex/internal/providers"
)

type stream struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
}

func newStream(reader io.ReadCloser) *stream {
	return &stream{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
	}
}

func (s *stream) Next() (*providers.StreamChunk, error) {
	for s.scanner.Scan() {
		line := s.scanner.Text()

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return &providers.StreamChunk{
				Type: "message_stop",
				Done: true,
			}, nil
		}

		var event openAIStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil, fmt.Errorf("unmarshal event: %w", err)
		}

		chunk := translateEvent(&event)
		if chunk != nil {
			return chunk, nil
		}
	}

	if err := s.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

func (s *stream) Close() error {
	return s.reader.Close()
}

type openAIStreamEvent struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Delta        openAIDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type openAIDelta struct {
	Content string `json:"content"`
}

func translateEvent(event *openAIStreamEvent) *providers.StreamChunk {
	if len(event.Choices) == 0 {
		return nil
	}

	choice := event.Choices[0]

	if choice.FinishReason != nil {
		return &providers.StreamChunk{
			Type: "message_stop",
			Done: true,
		}
	}

	if choice.Delta.Content != "" {
		return &providers.StreamChunk{
			Type:    "content_block_delta",
			Content: choice.Delta.Content,
		}
	}

	return nil
}
