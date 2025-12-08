// ABOUTME: Print mode handler for non-interactive CLI usage with tool support
// ABOUTME: Implements multi-turn tool execution loop like Claude Code
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/cost"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/tools"
)

func runPrintMode(prompt string) error {
	if prompt == "" && len(imagePaths) == 0 {
		return fmt.Errorf("prompt or image required in print mode")
	}

	// Create context for the entire print mode execution
	ctx := context.Background()

	// Load config
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Determine which provider to use (from flag, config, or default)
	providerName := provider
	if providerName == "" && cfg.Provider != "" {
		providerName = cfg.Provider
	}
	if providerName == "" {
		providerName = "anthropic" // Default
	}

	// Validate and determine model BEFORE creating provider (fail fast)
	modelToUse := model
	if modelToUse == "" {
		// Anthropic has a hardcoded default as fallback
		// Other providers require explicit --model flag
		if providerName == "anthropic" {
			if cfg.Model != "" {
				modelToUse = cfg.Model
			} else {
				modelToUse = core.DefaultModel
			}
		} else {
			return fmt.Errorf("--model flag is required when using --provider=%s\n\nExample models:\n  anthropic: claude-sonnet-4-5-20250929, claude-opus-4-5-20251101, claude-haiku-4-5-20251001\n  openai: gpt-5.1, gpt-5.1-codex, gpt-5.1-codex-mini\n  gemini: gemini-2.5-pro, gemini-2.5-flash, gemini-pro-latest\n  openrouter: anthropic/claude-sonnet-4-5, openai/gpt-5.1, google/gemini-2.5-pro", providerName)
		}
	}

	// Now create provider (after model is validated)
	client, err := createProvider(cfg, providerName)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
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

	// Multi-turn tool execution loop with token tracking
	maxTurns := 20
	var totalInputTokens, totalOutputTokens int

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
		resp, err := client.CreateMessage(ctx, req)
		if err != nil {
			return fmt.Errorf("API error: %w", err)
		}

		// Track token usage
		if resp.Usage.InputTokens > 0 || resp.Usage.OutputTokens > 0 {
			totalInputTokens += resp.Usage.InputTokens
			totalOutputTokens += resp.Usage.OutputTokens
			logging.DebugWith("Turn token usage",
				"turn", turn+1,
				"input_tokens", resp.Usage.InputTokens,
				"output_tokens", resp.Usage.OutputTokens,
				"total_input", totalInputTokens,
				"total_output", totalOutputTokens,
			)
		}

		logging.Debug("Response received", "stop_reason", resp.StopReason, "content_blocks", len(resp.Content))

		// Check stop reason
		if resp.StopReason == "end_turn" || resp.StopReason == "max_tokens" {
			// Print token usage summary if we have metrics
			if totalInputTokens > 0 || totalOutputTokens > 0 {
				logging.InfoWith("Total token usage",
					"input_tokens", totalInputTokens,
					"output_tokens", totalOutputTokens,
					"total_tokens", totalInputTokens+totalOutputTokens,
				)
			}
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

			// Execute tools in parallel and collect results
			logging.InfoWith("Executing tools in parallel", "count", len(toolUses))

			type toolResult struct {
				block core.ContentBlock
				index int
			}

			resultChan := make(chan toolResult, len(toolUses))
			var wg sync.WaitGroup

			// Launch all tool executions concurrently
			for i, toolUse := range toolUses {
				wg.Add(1)

				// Capture loop variables
				idx := i
				tu := toolUse

				go func() {
					defer wg.Done()

					// Recover from panics in tool execution
					defer func() {
						if r := recover(); r != nil {
							logging.WarnWith("Tool execution panicked", "tool", tu.Name, "panic", r)
							resultChan <- toolResult{
								block: core.ContentBlock{
									Type:      "tool_result",
									ToolUseID: tu.ID,
									Content:   fmt.Sprintf("Error: tool panicked: %v", r),
								},
								index: idx,
							}
						}
					}()

					// Check context cancellation before executing
					select {
					case <-ctx.Done():
						resultChan <- toolResult{
							block: core.ContentBlock{
								Type:      "tool_result",
								ToolUseID: tu.ID,
								Content:   fmt.Sprintf("Error: %v", ctx.Err()),
							},
							index: idx,
						}
						return
					default:
					}

					logging.InfoWith("Executing tool", "name", tu.Name, "id", tu.ID)

					result, err := executor.Execute(ctx, tu.Name, tu.Input)
					if err != nil {
						resultChan <- toolResult{
							block: core.ContentBlock{
								Type:      "tool_result",
								ToolUseID: tu.ID,
								Content:   fmt.Sprintf("Error: %v", err),
							},
							index: idx,
						}
						return
					}

					resultChan <- toolResult{
						block: core.ContentBlock{
							Type:      "tool_result",
							ToolUseID: tu.ID,
							Content:   result.Output,
						},
						index: idx,
					}
				}()
			}

			// Wait for all tools to complete
			go func() {
				wg.Wait()
				close(resultChan)
			}()

			// Collect results in a map to avoid zero-value issues
			resultsMap := make(map[int]toolResult, len(toolUses))
			for tr := range resultChan {
				resultsMap[tr.index] = tr
			}

			// Build toolResults array in original order
			toolResults := make([]core.ContentBlock, 0, len(toolUses))
			for i := 0; i < len(toolUses); i++ {
				if tr, ok := resultsMap[i]; ok {
					toolResults = append(toolResults, tr.block)
				} else {
					// Handle missing result - this indicates a serious problem
					logging.WarnWith("Tool execution did not complete", "index", i, "tool_use_id", toolUses[i].ID)
					toolResults = append(toolResults, core.ContentBlock{
						Type:      "tool_result",
						ToolUseID: toolUses[i].ID,
						Content:   "Error: tool execution did not complete",
					})
				}
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
		// Print token usage summary if we have metrics
		if totalInputTokens > 0 || totalOutputTokens > 0 {
			logging.InfoWith("Total token usage",
				"input_tokens", totalInputTokens,
				"output_tokens", totalOutputTokens,
				"total_tokens", totalInputTokens+totalOutputTokens,
			)
		}

		// Print cost summary if agent ID is set
		agentID := os.Getenv("HEX_AGENT_ID")
		if agentID != "" {
			cost.PrintCostSummary(agentID)
		}

		return formatOutput(resp, outputFormat)
	}

	// Print token usage even on max turns error
	if totalInputTokens > 0 || totalOutputTokens > 0 {
		logging.InfoWith("Total token usage (partial)",
			"input_tokens", totalInputTokens,
			"output_tokens", totalOutputTokens,
			"total_tokens", totalInputTokens+totalOutputTokens,
		)
	}

	// Print cost summary if agent ID is set (even on error)
	agentID := os.Getenv("HEX_AGENT_ID")
	if agentID != "" {
		cost.PrintCostSummary(agentID)
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
