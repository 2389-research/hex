// ABOUTME: JSON stream parsing for Gemini streaming responses
// ABOUTME: Handles candidates array with content parts
package gemini

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/2389-research/hex/internal/providers"
)

type stream struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
	done    bool
}

func newStream(reader io.ReadCloser) *stream {
	return &stream{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
	}
}

func (s *stream) Next() (*providers.StreamChunk, error) {
	if s.done {
		return nil, io.EOF
	}

	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			continue
		}

		var event geminiStreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // Skip malformed lines
		}

		chunk := translateEvent(&event)
		if chunk != nil {
			if chunk.Done {
				s.done = true
			}
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

type geminiStreamEvent struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

func translateEvent(event *geminiStreamEvent) *providers.StreamChunk {
	if len(event.Candidates) == 0 {
		return nil
	}

	candidate := event.Candidates[0]

	if candidate.FinishReason != "" {
		return &providers.StreamChunk{
			Type: "message_stop",
			Done: true,
		}
	}

	if len(candidate.Content.Parts) > 0 {
		text := candidate.Content.Parts[0].Text
		if text != "" {
			return &providers.StreamChunk{
				Type:    "content_block_delta",
				Content: text,
			}
		}
	}

	return nil
}
