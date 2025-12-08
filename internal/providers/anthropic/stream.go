// ABOUTME: SSE stream parsing for Anthropic API responses
// ABOUTME: Handles message_start, content_block_delta, message_stop events
package anthropic

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

		var event anthropicEvent
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

type anthropicEvent struct {
	Type  string          `json:"type"`
	Delta *anthropicDelta `json:"delta"`
}

type anthropicDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func translateEvent(event *anthropicEvent) *providers.StreamChunk {
	switch event.Type {
	case "message_start":
		return &providers.StreamChunk{
			Type: "message_start",
		}
	case "content_block_delta":
		if event.Delta != nil {
			return &providers.StreamChunk{
				Type:    "content_block_delta",
				Content: event.Delta.Text,
			}
		}
	case "message_stop":
		return &providers.StreamChunk{
			Type: "message_stop",
			Done: true,
		}
	}
	return nil
}
