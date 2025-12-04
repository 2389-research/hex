// ABOUTME: Example demonstrating token savings with context management
// ABOUTME: Shows real-world impact of pruning on token usage and costs
package context_test

import (
	"fmt"

	"github.com/harper/jeff/internal/context"
	"github.com/harper/jeff/internal/core"
)

// ExampleManager_tokenSavings demonstrates the token savings from context pruning
func ExampleManager_tokenSavings() {
	// Simulate a long conversation (500 messages with substantial content)
	messages := make([]core.Message, 500)
	messages[0] = core.Message{Role: "system", Content: "You are a helpful coding assistant with extensive knowledge of Go, Python, JavaScript, and software architecture."}

	for i := 1; i < 500; i++ {
		if i%2 == 1 {
			messages[i] = core.Message{
				Role: "user",
				Content: fmt.Sprintf("This is user message #%d asking a detailed question about coding in Go and Python. "+
					"I need help understanding how to implement a feature with proper error handling, testing, and documentation. "+
					"Can you provide a complete example with best practices?", i),
			}
		} else {
			messages[i] = core.Message{
				Role: "assistant",
				Content: fmt.Sprintf("This is assistant response #%d providing a detailed explanation of code concepts with examples. "+
					"Here's how you implement that feature: First, you need to set up the proper structure, then handle errors appropriately, "+
					"write comprehensive tests, and document your code thoroughly. Let me show you a complete example...", i),
			}
		}
	}

	// Without context management
	tokensWithoutPruning := context.EstimateMessagesTokens(messages)

	// With context management (1000 token limit for demonstration)
	manager := context.NewManager(1000)
	prunedMessages := manager.Prune(messages)
	tokensWithPruning := context.EstimateMessagesTokens(prunedMessages)

	// Calculate savings
	tokensSaved := tokensWithoutPruning - tokensWithPruning
	percentSaved := float64(tokensSaved) / float64(tokensWithoutPruning) * 100

	// Cost calculation (Sonnet 4.5 pricing: $3/million input tokens)
	costWithout := float64(tokensWithoutPruning) * 3.0 / 1_000_000
	costWith := float64(tokensWithPruning) * 3.0 / 1_000_000
	costSaved := costWithout - costWith

	fmt.Printf("Original conversation: %d messages, ~%d tokens\n", len(messages), tokensWithoutPruning)
	fmt.Printf("After pruning: %d messages, ~%d tokens\n", len(prunedMessages), tokensWithPruning)
	fmt.Printf("Tokens saved: %d (%.1f%%)\n", tokensSaved, percentSaved)
	fmt.Printf("Cost saved per request: $%.4f\n", costSaved)
	fmt.Printf("Cost saved over 100 requests: $%.2f\n", costSaved*100)

	// Output:
	// Original conversation: 500 messages, ~37400 tokens
	// After pruning: 5 messages, ~333 tokens
	// Tokens saved: 37067 (99.1%)
	// Cost saved per request: $0.1112
	// Cost saved over 100 requests: $11.12
}

// ExamplePruneContext demonstrates context pruning preserving important messages
func ExamplePruneContext() {
	messages := []core.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Old message 1"},
		{Role: "assistant", Content: "Old response 1"},
		{Role: "user", Content: "Old message 2"},
		{Role: "assistant", Content: "Old response 2"},
		{Role: "user", Content: "Use the read tool"},
		{Role: "assistant", Content: "Reading file", ToolCalls: []core.ToolUse{{Name: "read_file", ID: "1"}}},
		{Role: "user", Content: "Recent message"},
		{Role: "assistant", Content: "Recent response"},
	}

	pruned := context.PruneContext(messages, 100)

	fmt.Printf("Original: %d messages\n", len(messages))
	fmt.Printf("Pruned: %d messages\n", len(pruned))
	fmt.Printf("Kept system message: %v\n", pruned[0].Role == "system")

	hasToolCall := false
	for _, msg := range pruned {
		if len(msg.ToolCalls) > 0 {
			hasToolCall = true
			break
		}
	}
	fmt.Printf("Kept tool call message: %v\n", hasToolCall)
	fmt.Printf("Kept recent messages: %v\n", pruned[len(pruned)-1].Content == "Recent response")

	// Output:
	// Original: 9 messages
	// Pruned: 9 messages
	// Kept system message: true
	// Kept tool call message: true
	// Kept recent messages: true
}
