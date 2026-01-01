// ABOUTME: Adapter that wraps mux's llm.Client to implement hex's Provider interface
// ABOUTME: Translates between hex's core types and mux's llm types for all providers

package providers

import (
	"context"
	"fmt"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/mux/llm"
)

// MuxAdapter wraps a mux llm.Client to implement hex's Provider interface
type MuxAdapter struct {
	client llm.Client
	name   string
}

// NewMuxAdapter creates a new adapter wrapping a mux llm.Client
func NewMuxAdapter(name string, client llm.Client) *MuxAdapter {
	return &MuxAdapter{
		client: client,
		name:   name,
	}
}

// Name returns the provider name
func (a *MuxAdapter) Name() string {
	return a.name
}

// ValidateConfig validates the provider configuration
func (a *MuxAdapter) ValidateConfig(cfg ProviderConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("%s: API key is required", a.name)
	}
	return nil
}

// CreateMessage sends a synchronous message request
func (a *MuxAdapter) CreateMessage(ctx context.Context, req core.MessageRequest) (*core.MessageResponse, error) {
	// Translate core.MessageRequest → llm.Request
	llmReq := translateToMuxRequest(req)

	// Call mux client
	resp, err := a.client.CreateMessage(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", a.name, err)
	}

	// Translate llm.Response → core.MessageResponse
	return translateFromMuxResponse(resp), nil
}

// CreateMessageStream sends a streaming message request
func (a *MuxAdapter) CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error) {
	// Translate core.MessageRequest → llm.Request
	llmReq := translateToMuxRequest(req)

	// Call mux client for streaming
	eventChan, err := a.client.CreateMessageStream(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", a.name, err)
	}

	// Create output channel
	chunks := make(chan *core.StreamChunk, 10)

	// Goroutine to translate events
	go func() {
		defer close(chunks)

		for event := range eventChan {
			chunk := translateStreamEvent(event)
			if chunk != nil {
				select {
				case chunks <- chunk:
				case <-ctx.Done():
					return
				}

				if chunk.Done {
					return
				}
			}
		}
	}()

	return chunks, nil
}

// translateToMuxRequest converts core.MessageRequest to llm.Request
func translateToMuxRequest(req core.MessageRequest) *llm.Request {
	muxReq := &llm.Request{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		System:    req.System,
		Messages:  make([]llm.Message, len(req.Messages)),
	}

	// Translate messages
	for i, msg := range req.Messages {
		muxMsg := llm.Message{
			Role: llm.Role(msg.Role),
		}

		// Handle content blocks if present
		if len(msg.ContentBlock) > 0 {
			muxMsg.Blocks = make([]llm.ContentBlock, len(msg.ContentBlock))
			for j, block := range msg.ContentBlock {
				muxMsg.Blocks[j] = translateContentBlockToMux(block)
			}
		} else {
			muxMsg.Content = msg.Content
		}

		muxReq.Messages[i] = muxMsg
	}

	// Translate tools
	if len(req.Tools) > 0 {
		muxReq.Tools = make([]llm.ToolDefinition, len(req.Tools))
		for i, tool := range req.Tools {
			muxReq.Tools[i] = llm.ToolDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.InputSchema,
			}
		}
	}

	return muxReq
}

// translateContentBlockToMux converts core.ContentBlock to llm.ContentBlock
func translateContentBlockToMux(block core.ContentBlock) llm.ContentBlock {
	muxBlock := llm.ContentBlock{
		Type: llm.ContentType(block.Type),
		Text: block.Text,
		ID:   block.ID,
		Name: block.Name,
	}

	if block.Input != nil {
		muxBlock.Input = block.Input
	}

	if block.ToolUseID != "" {
		muxBlock.ToolUseID = block.ToolUseID
		// For tool_result, content goes in Text field
		if block.Content != "" {
			muxBlock.Text = block.Content
		}
	}

	return muxBlock
}

// translateFromMuxResponse converts llm.Response to core.MessageResponse
func translateFromMuxResponse(resp *llm.Response) *core.MessageResponse {
	coreResp := &core.MessageResponse{
		ID:         resp.ID,
		Type:       "message",
		Role:       "assistant",
		Model:      resp.Model,
		StopReason: string(resp.StopReason),
		Usage: core.Usage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		},
	}

	// Translate content blocks
	coreResp.Content = make([]core.Content, len(resp.Content))
	for i, block := range resp.Content {
		coreResp.Content[i] = core.Content{
			Type:  string(block.Type),
			Text:  block.Text,
			ID:    block.ID,
			Name:  block.Name,
			Input: block.Input,
		}
	}

	return coreResp
}

// translateStreamEvent converts llm.StreamEvent to core.StreamChunk
func translateStreamEvent(event llm.StreamEvent) *core.StreamChunk {
	switch event.Type {
	case llm.EventMessageStart:
		return &core.StreamChunk{
			Type: "message_start",
		}

	case llm.EventContentStart:
		if event.Block != nil {
			return &core.StreamChunk{
				Type:  "content_block_start",
				Index: event.Index,
				ContentBlock: &core.Content{
					Type: string(event.Block.Type),
					ID:   event.Block.ID,
					Name: event.Block.Name,
				},
			}
		}

	case llm.EventContentDelta:
		chunk := &core.StreamChunk{
			Type:  "content_block_delta",
			Index: event.Index,
		}
		if event.Text != "" {
			chunk.Delta = &core.Delta{
				Type: "text_delta",
				Text: event.Text,
			}
		} else if event.Block != nil && event.Block.Input != nil {
			// Tool input JSON delta - serialize the partial input
			chunk.Delta = &core.Delta{
				Type: "input_json_delta",
			}
		}
		return chunk

	case llm.EventContentStop:
		return &core.StreamChunk{
			Type:  "content_block_stop",
			Index: event.Index,
		}

	case llm.EventMessageDelta:
		chunk := &core.StreamChunk{
			Type: "message_delta",
		}
		if event.Response != nil {
			chunk.Delta = &core.Delta{
				Type:       "message_delta",
				StopReason: string(event.Response.StopReason),
			}
			chunk.Usage = &core.Usage{
				InputTokens:  event.Response.Usage.InputTokens,
				OutputTokens: event.Response.Usage.OutputTokens,
			}
		}
		return chunk

	case llm.EventMessageStop:
		return &core.StreamChunk{
			Type: "message_stop",
			Done: true,
		}

	case llm.EventError:
		return &core.StreamChunk{
			Type: "error",
			Done: true,
			Delta: &core.Delta{
				Type: "error",
				Text: event.Error.Error(),
			},
		}
	}

	return nil
}

// Ensure MuxAdapter implements Provider interface
var _ Provider = (*MuxAdapter)(nil)
