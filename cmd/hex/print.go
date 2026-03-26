// ABOUTME: Print mode handler for non-interactive CLI usage with tool support
// ABOUTME: Implements multi-turn tool execution loop like Claude Code
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/cost"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/memory"
	"github.com/2389-research/hex/internal/tools"
)

func runPrintMode(prompt string) error {
	// Use mux by default, fall back to legacy if requested
	if !useLegacyMode {
		return runPrintModeWithMux(prompt)
	}

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
			return fmt.Errorf("--model flag is required when using --provider=%s\n\nExample models:\n  anthropic: claude-sonnet-4-5-20250929, claude-opus-4-5-20251101, claude-haiku-4-5-20251001\n  openai: gpt-5.1, gpt-5.1-codex, gpt-5.1-codex-mini\n  gemini: gemini-2.5-pro, gemini-2.5-flash, gemini-pro-latest\n  openrouter: anthropic/claude-sonnet-4-5, openai/gpt-5.1, google/gemini-2.5-pro\n  ollama: llama3.2, codellama, mistral, mixtral", providerName)
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
			imgSrc, loadErr := core.LoadImage(imgPath)
			if loadErr != nil {
				return fmt.Errorf("load image %s: %w", imgPath, loadErr)
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

	// Plan mode: wrap prompt with planning instruction
	if planMode && prompt != "" && len(imagePaths) == 0 {
		msg.Content = "Before executing, create a numbered plan for this task. List:\n1. What files you need to read\n2. What changes to make\n3. How to verify the changes work\n\nOutput ONLY the plan, do not start executing yet.\n\nTask: " + prompt
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
	maxTurns := 50
	if maxTurns == 0 {
		maxTurns = 50
	}
	tracker := &turnTracker{}
	var totalInputTokens, totalOutputTokens int

	for turn := 0; turn < maxTurns; turn++ {
		logging.DebugWith("Print mode turn", "turn", turn+1, "messages", len(messages))

		// Create API request
		req := core.MessageRequest{
			Model:     modelToUse,
			MaxTokens: 16384,
			Messages:  messages,
		}

		// Build effective system prompt
		effectiveSystemPrompt := systemPrompt
		var spellIsReplaceMode bool

		// Apply spell if specified
		if spellName != "" {
			spellPrompt, effectiveMode, err := getSpellWithMode(spellName, effectiveSystemPrompt, spellMode)
			if err != nil {
				return fmt.Errorf("load spell %q: %w", spellName, err)
			}
			effectiveSystemPrompt = spellPrompt
			spellIsReplaceMode = effectiveMode == "replace"
			logging.InfoWith("Applied spell", "name", spellName, "mode", string(effectiveMode))
		}

		// Include Hex identity in system prompt UNLESS spell is in replace mode
		if spellIsReplaceMode {
			// Replace mode: use only the spell's system prompt
			req.System = effectiveSystemPrompt
		} else if effectiveSystemPrompt != "" {
			req.System = core.DefaultSystemPrompt + "\n\n" + effectiveSystemPrompt
		} else {
			req.System = core.DefaultSystemPrompt
		}

		// Print mode is always non-interactive — append headless guidance
		req.System = req.System + core.HeadlessGuidance

		// Load project memory context
		projContext := loadProjectContext()
		if projContext != "" {
			req.System = req.System + "\n\n" + projContext
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
			// In plan mode on first turn, inject plan and continue to execution
			if planMode && turn == 0 {
				assistantMsg := core.Message{
					Role:    "assistant",
					Content: resp.GetTextContent(),
				}
				messages = append(messages, assistantMsg)

				execMsg := core.Message{
					Role:    "user",
					Content: "Good plan. Now execute it step by step. After completing each step, note which step you finished.",
				}
				messages = append(messages, execMsg)
				continue
			}

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

			// Per-tool output limits to preserve context budget
			toolLimits := map[string]int{
				"read_file":  12000,
				"bash":       8000,
				"grep":       5000,
				"glob":       3000,
				"edit":       5000,
				"write_file": 3000,
			}
			const defaultLimit = 15000
			for i := range toolResults {
				limit := defaultLimit
				if toolResults[i].ToolUseID != "" {
					for _, tu := range toolUses {
						if tu.ID == toolResults[i].ToolUseID {
							if l, ok := toolLimits[tu.Name]; ok {
								limit = l
							}
							break
						}
					}
				}
				if len(toolResults[i].Content) > limit {
					orig := len(toolResults[i].Content)
					toolResults[i].Content = toolResults[i].Content[:limit] + fmt.Sprintf("\n\n[truncated: showing first %d of %d chars. Use offset/limit parameters to read specific sections.]", limit, orig)
				}
			}

			// Check for stuck patterns
			hint := tracker.recordByToolName(toolUses, toolResults)
			if hint != "" {
				logging.WarnWith("Stuck detection triggered", "hint", hint)
				if len(toolResults) > 0 {
					toolResults[len(toolResults)-1].Content = toolResults[len(toolResults)-1].Content + "\n\n" + hint
				}
			}

			// Add verification hint if files were modified
			hasMutation := false
			for _, tu := range toolUses {
				if tu.Name == "edit" || tu.Name == "write_file" {
					hasMutation = true
					break
				}
			}
			if hasMutation && len(toolResults) > 0 {
				toolResults[len(toolResults)-1].Content = toolResults[len(toolResults)-1].Content + "\n\n[hex: Files were modified. Consider verifying your changes compile or pass tests before proceeding.]"
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

	// Max turns reached — output best effort instead of erroring
	// The agent may have already written files; exit cleanly so verifiers can check
	if totalInputTokens > 0 || totalOutputTokens > 0 {
		logging.InfoWith("Total token usage (max turns reached)",
			"input_tokens", totalInputTokens,
			"output_tokens", totalOutputTokens,
			"total_tokens", totalInputTokens+totalOutputTokens,
		)
	}

	agentID := os.Getenv("HEX_AGENT_ID")
	if agentID != "" {
		cost.PrintCostSummary(agentID)
	}

	// Output whatever we have rather than failing
	fmt.Fprintf(os.Stderr, "Warning: reached maximum turns (%d)\n", maxTurns)
	return nil
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

// loadProjectContext loads or detects project context for the system prompt
func loadProjectContext() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	hexDir := filepath.Join(cwd, ".hex")
	proj, err := memory.Load(hexDir)

	if err != nil || refreshMemory || memory.IsStale(proj, 7*24*time.Hour) {
		proj, err = memory.DetectProject(cwd)
		if err != nil {
			return ""
		}
		_ = memory.Save(hexDir, proj)
	}

	return proj.ToPromptContext()
}
