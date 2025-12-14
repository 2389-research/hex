// ABOUTME: Adapter that wraps hex tools to implement mux's tool.Tool interface.
// ABOUTME: Allows hex's existing tools to work with mux's orchestrator.
package adapter

import (
	"context"

	"github.com/2389-research/hex/internal/tools"
	muxtool "github.com/2389-research/mux/tool"
)

// adaptedTool wraps a hex Tool to implement mux's Tool interface.
type adaptedTool struct {
	hex tools.Tool
}

// AdaptTool wraps a hex tool for use with mux.
func AdaptTool(t tools.Tool) muxtool.Tool {
	return &adaptedTool{hex: t}
}

// AdaptAll wraps multiple hex tools for use with mux.
func AdaptAll(hexTools []tools.Tool) []muxtool.Tool {
	adapted := make([]muxtool.Tool, len(hexTools))
	for i, t := range hexTools {
		adapted[i] = AdaptTool(t)
	}
	return adapted
}

func (a *adaptedTool) Name() string {
	return a.hex.Name()
}

func (a *adaptedTool) Description() string {
	return a.hex.Description()
}

func (a *adaptedTool) RequiresApproval(params map[string]any) bool {
	// Convert map[string]any to map[string]interface{}
	converted := make(map[string]interface{}, len(params))
	for k, v := range params {
		converted[k] = v
	}
	return a.hex.RequiresApproval(converted)
}

func (a *adaptedTool) Execute(ctx context.Context, params map[string]any) (*muxtool.Result, error) {
	// Convert map[string]any to map[string]interface{}
	converted := make(map[string]interface{}, len(params))
	for k, v := range params {
		converted[k] = v
	}

	result, err := a.hex.Execute(ctx, converted)
	if err != nil {
		return nil, err
	}

	// Convert hex Result to mux Result
	// Note: Metadata conversion from map[string]interface{} to map[string]any
	muxMetadata := make(map[string]any, len(result.Metadata))
	for k, v := range result.Metadata {
		muxMetadata[k] = v
	}

	return &muxtool.Result{
		ToolName: result.ToolName,
		Success:  result.Success,
		Output:   result.Output,
		Error:    result.Error,
		Metadata: muxMetadata,
	}, nil
}

// Compile-time interface assertion.
var _ muxtool.Tool = (*adaptedTool)(nil)
