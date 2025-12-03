// ABOUTME: Print mode handler for non-interactive CLI usage with tool support
// ABOUTME: Implements multi-turn tool execution loop like Claude Code
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/tools"
)

func runPrintMode(prompt string) error {
	if prompt == "" && len(imagePaths) == 0 {
		return fmt.Errorf("prompt or image required in print mode")
	}

	// Load config
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Get API key
	apiKey, err := cfg.GetAPIKey()
	if err != nil {
		return fmt.Errorf("get API key: %w", err)
	}

	// Create client
	client := core.NewClient(apiKey)

	// Use model from flag if set, otherwise use config/default
	modelToUse := model
	if modelToUse == "" && cfg.Model != "" {
		modelToUse = cfg.Model
	}
	if modelToUse == "" {
		modelToUse = core.DefaultModel
	}

	// Build initial user message
	var msg core.Message
	msg.Role = "user"

	if len(imagePaths) > 0 {
		contentBlocks := []core.ContentBlock{}
		for _, imgPath := range imagePaths {
			imgSrc, err := core.LoadImage(imgPath)
			if err != nil {
				return fmt.Errorf("load image %s: %w", imgPath, err)
			}
			contentBlocks = append(contentBlocks, core.NewImageBlock(imgSrc))
		}
		if prompt != "" {
			contentBlocks = append(contentBlocks, core.NewTextBlock(prompt))
		}
		msg.ContentBlock = contentBlocks
	} else {
		msg.Content = prompt
	}

	// Setup tools - always enabled in print mode
	var registry *tools.Registry
	var executor *tools.Executor
	var toolDefs []core.ToolDefinition

	registry, executor, err = setupPrintModeTools()
	if err != nil {
		return fmt.Errorf("setup tools: %w", err)
	}
	toolDefs = registry.GetDefinitions()

	// Build conversation history
	messages := []core.Message{msg}

	// Multi-turn tool execution loop
	maxTurns := 20
	for turn := 0; turn < maxTurns; turn++ {
		logging.DebugWith("Print mode turn", "turn", turn+1, "messages", len(messages))

		// Create API request
		req := core.MessageRequest{
			Model:     modelToUse,
			MaxTokens: 4096,
			Messages:  messages,
		}

		if systemPrompt != "" {
			req.System = systemPrompt
		}

		if len(toolDefs) > 0 {
			req.Tools = toolDefs
		}

		// Send request
		resp, err := client.CreateMessage(context.Background(), req)
		if err != nil {
			return fmt.Errorf("API error: %w", err)
		}

		logging.DebugWith("Response received", "stop_reason", resp.StopReason, "content_blocks", len(resp.Content))

		// Check stop reason
		if resp.StopReason == "end_turn" || resp.StopReason == "max_tokens" {
			return formatOutput(resp, outputFormat)
		}

		if resp.StopReason == "tool_use" {
			if executor == nil {
				return fmt.Errorf("received tool_use response but no executor configured")
			}

			// Extract tool uses from response
			var toolUses []core.ToolUse
			for _, content := range resp.Content {
				if content.Type == "tool_use" {
					toolUses = append(toolUses, core.ToolUse{
						ID:    content.ID,
						Name:  content.Name,
						Input: content.Input,
					})
				}
			}

			if len(toolUses) == 0 {
				return fmt.Errorf("stop_reason was tool_use but no tool_use blocks found")
			}

			logging.InfoWith("Executing tools", "count", len(toolUses))

			// Add assistant message with tool uses
			assistantMsg := core.Message{
				Role: "assistant",
			}
			// Convert []Content to []ContentBlock for the message
			assistantContentBlocks := make([]core.ContentBlock, len(resp.Content))
			for i, c := range resp.Content {
				assistantContentBlocks[i] = core.ContentBlock{
					Type:  c.Type,
					Text:  c.Text,
					ID:    c.ID,
					Name:  c.Name,
					Input: c.Input,
				}
			}
			assistantMsg.ContentBlock = assistantContentBlocks
			messages = append(messages, assistantMsg)

			// Execute tools and collect results
			var toolResults []core.ContentBlock
			for _, toolUse := range toolUses {
				logging.InfoWith("Executing tool", "name", toolUse.Name, "id", toolUse.ID)

				result, err := executor.Execute(context.Background(), toolUse.Name, toolUse.Input)
				if err != nil {
					toolResults = append(toolResults, core.ContentBlock{
						Type:      "tool_result",
						ToolUseID: toolUse.ID,
						Content:   fmt.Sprintf("Error: %v", err),
					})
					continue
				}

				toolResults = append(toolResults, core.ContentBlock{
					Type:      "tool_result",
					ToolUseID: toolUse.ID,
					Content:   result.Output,
				})
			}

			// Add tool results as user message
			userMsg := core.Message{
				Role:         "user",
				ContentBlock: toolResults,
			}
			messages = append(messages, userMsg)

			// Continue loop
			continue
		}

		// Unknown stop reason
		logging.WarnWith("Unknown stop reason", "stop_reason", resp.StopReason)
		return formatOutput(resp, outputFormat)
	}

	return fmt.Errorf("exceeded maximum turns (%d) in tool execution loop", maxTurns)
}

func formatOutput(resp *core.MessageResponse, format string) error {
	switch format {
	case "text":
		fmt.Println(resp.GetTextContent())
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("encode JSON: %w", err)
		}
	case "stream-json":
		return fmt.Errorf("streaming not yet implemented")
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
	return nil
}
