// ABOUTME: Print mode handler for non-interactive CLI usage
// ABOUTME: Connects CLI -> Config -> API Client for one-off queries
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/harper/clem/internal/core"
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
		modelToUse = core.DefaultModel // fallback default
	}

	// Build message content
	var msg core.Message
	msg.Role = "user"

	// If images are provided, use ContentBlock array
	if len(imagePaths) > 0 {
		contentBlocks := []core.ContentBlock{}

		// Add images first
		for _, imgPath := range imagePaths {
			imgSrc, err := core.LoadImage(imgPath)
			if err != nil {
				return fmt.Errorf("load image %s: %w", imgPath, err)
			}
			contentBlocks = append(contentBlocks, core.NewImageBlock(imgSrc))
		}

		// Add text prompt if provided
		if prompt != "" {
			contentBlocks = append(contentBlocks, core.NewTextBlock(prompt))
		}

		msg.ContentBlock = contentBlocks
	} else {
		// Text-only message
		msg.Content = prompt
	}

	// Create request
	req := core.MessageRequest{
		Model:     modelToUse,
		MaxTokens: 4096,
		Messages:  []core.Message{msg},
	}

	// Send request
	resp, err := client.CreateMessage(context.Background(), req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}

	// Format output
	return formatOutput(resp, outputFormat)
}

func formatOutput(resp *core.MessageResponse, format string) error {
	switch format {
	case "text":
		// Simple text output - just the assistant's response
		fmt.Println(resp.GetTextContent())
	case "json":
		// Full JSON response for programmatic use
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
