# Context Management in Hex

Hex implements intelligent context management to handle long conversations efficiently and avoid hitting Claude's token limits.

## Overview

Claude models have large context windows (up to 200k tokens for Sonnet), but:
- Sending full conversation history on every request wastes tokens and costs money
- Very long contexts can degrade model performance
- Some conversations naturally exceed even 200k tokens

Hex's context management system automatically:
- **Estimates token usage** in real-time
- **Prunes old messages** when approaching limits
- **Preserves important context** (system messages, tool uses, recent exchanges)
- **Warns users** when context is getting full
- **Optionally summarizes** removed messages (future enhancement)

## Token Estimation

Hex uses a simple but effective heuristic to estimate tokens:

```
tokens ≈ characters / 4
```

This is accurate enough for most purposes. Each message also has ~4 tokens of overhead for role markers and formatting.

You can see estimated token usage in the status bar at the bottom of the screen.

## Context Pruning

When your conversation approaches the context limit, Hex automatically prunes messages using a smart strategy:

### What Gets Kept

1. **System message** (always kept if present) - contains important instructions
2. **Recent messages** (last 4+ exchanges) - most relevant for current conversation
3. **Messages with tool calls** - important for understanding tool usage history
4. **Last user message** - ensures continuity

### What Gets Removed

- Old messages from the middle of the conversation
- Messages that don't contain tool calls
- Simple back-and-forth exchanges that are less critical

### Example

Original conversation (200 messages, ~150k tokens):
```
[System] You are a helpful coding assistant
[User] Help me with Python
[Assistant] Sure, what do you need?
... 196 more messages ...
[User] Now explain async/await
[Assistant] Let me explain...
```

After pruning (keeping within 100k tokens):
```
[System] You are a helpful coding assistant
... middle messages removed ...
[User] Earlier you helped with file operations (message #87 had tool call)
[Assistant] Yes, we used the write_file tool
... last 4 exchanges kept ...
[User] Now explain async/await
[Assistant] Let me explain...
```

## Configuration

### Command-Line Flags

#### `--max-context-tokens <n>`

Set the maximum context window size in tokens.

**Default:** 180,000 (allows some buffer below the 200k limit)

**Examples:**
```bash
# Use smaller context window for faster responses
hex --max-context-tokens 50000

# Use nearly full Sonnet context
hex --max-context-tokens 195000

# Conservative limit for older models
hex --max-context-tokens 100000
```

#### `--context-strategy <strategy>`

Choose how to manage context when it gets full.

**Options:**
- `keep-all` - Never prune (will error if exceeding limit)
- `prune` - Automatically remove old messages (default)
- `summarize` - Summarize removed messages (future enhancement)

**Default:** `prune`

**Examples:**
```bash
# Never prune - good for critical conversations
hex --context-strategy keep-all

# Auto-prune old messages (default)
hex --context-strategy prune

# Summarize removed context (when implemented)
hex --context-strategy summarize
```

## Status Bar Indicators

The status bar shows context usage:

### Normal Operation
```
claude-sonnet-4-5  ●  15k↓ 8k↑  [chat]
```
- `15k↓` = ~15,000 tokens estimated input
- `8k↑` = ~8,000 tokens estimated output
- Green dot (●) = Connected to API

### Approaching Limit (>50% full)
```
claude-sonnet-4-5  ●  95k↓ 45k↑  [███████░░░]  [chat]
```
- Visual bar shows context usage
- Fills up as you approach limit

### Near Limit (>90% full)
```
claude-sonnet-4-5  ●  172k↓ 18k↑  [█████████▓]  ⚠ Context 95% full - pruning recommended
```
- Yellow warning appears
- Bar shows nearly full
- Consider starting a new conversation or using `--continue` with a fresh context

## Token Cost Implications

Context management directly affects your API costs:

### Without Pruning
Long conversation (500 messages):
- Input: ~200k tokens per request
- Output: ~500 tokens per request
- **Cost per request:** ~$0.60 (at $3/million input tokens)
- **10 requests:** ~$6.00

### With Pruning (80k token limit)
Same conversation, pruned:
- Input: ~80k tokens per request
- Output: ~500 tokens per request
- **Cost per request:** ~$0.24
- **10 requests:** ~$2.40
- **Savings:** 60% or ~$3.60

## Best Practices

### 1. Use Reasonable Context Limits

```bash
# For most conversations
hex --max-context-tokens 100000

# For code generation (needs more context)
hex --max-context-tokens 150000

# For quick Q&A (save money)
hex --max-context-tokens 50000
```

### 2. Monitor Context Usage

Watch the status bar. When you see the context bar filling up:
- Consider wrapping up the current topic
- Start a new conversation for new topics
- Use `--resume` to continue specific conversations

### 3. Start Fresh for New Topics

Instead of:
```bash
hex --continue  # Could load 150k tokens of old context
```

Consider:
```bash
hex  # Fresh conversation for new topic
```

### 4. Resume Important Conversations

For ongoing work:
```bash
# List conversations
hex list

# Resume specific conversation
hex --resume conv-12345
```

## How Pruning Works Internally

When you send a message, Hex:

1. **Estimates tokens** for all messages
2. **Checks if pruning needed** (`EstimateMessagesTokens` > `MaxTokens`)
3. **If yes:**
   - Keeps system message (first message if role="system")
   - Keeps last N messages (recent context)
   - Keeps messages with tool calls (important context)
   - Removes messages from the middle
4. **Sends pruned context** to API
5. **Stores full conversation** in database (nothing is lost)

Your full conversation history is always preserved in the database. Pruning only affects what's sent to the API.

## Future Enhancements

### Summarization (Phase 6B+)

Instead of just removing messages, Hex will summarize them:

```
[System] Previous conversation summary:
User asked about Python file handling. We discussed os.path, pathlib, and
demonstrated reading/writing files. Later covered error handling with try/except.
---
[Recent messages continue...]
```

This preserves context while using fewer tokens.

### RAG Integration (Experimental)

Retrieve relevant context from past conversations:

```bash
hex --enable-rag
```

This will:
- Generate embeddings for all messages
- Store in SQLite vector extension
- Retrieve relevant past context when needed
- Requires OpenAI API key for embeddings

**Status:** Planned but not yet implemented

## Troubleshooting

### "Context limit exceeded" Error

If you see this error:
1. Check your `--max-context-tokens` setting
2. Try reducing it or using `--context-strategy prune`
3. Consider starting a new conversation

### Context Pruning Too Aggressive

If important context is being removed:
1. Increase `--max-context-tokens`
2. Use more descriptive messages (helps identify important context)
3. Use tool calls when doing important operations (they're preserved)

### Status Bar Not Updating

Token estimates update when:
- Messages are added
- Context manager is set
- UI refreshes

If you don't see updates, check that the context manager was initialized (should see in debug logs).

## API Reference

### Manager

```go
// Create a context manager
manager := context.NewManager(180000) // max tokens

// Check if pruning needed
if manager.ShouldPrune(messages) {
    pruned := manager.Prune(messages)
}

// Get usage stats
usage := manager.GetUsage(messages)
fmt.Printf("Using %d tokens (%.1f%%)\n",
    usage.EstimatedTokens, usage.PercentUsed)
```

### Summarizer (Future)

```go
// Create summarizer
summarizer := context.NewSummarizer(apiClient)

// Summarize old messages
summary, err := summarizer.SummarizeMessages(ctx, oldMessages)

// Create summary message
summaryMsg := context.CreateSummaryMessage(summary)
```

## Examples

### Example 1: Conservative Token Usage

```bash
hex --max-context-tokens 50000 --context-strategy prune
```

Aggressively prunes to keep costs low. Good for:
- Quick questions
- Simple tasks
- High-volume usage

### Example 2: Maximum Context

```bash
hex --max-context-tokens 195000 --context-strategy keep-all
```

Use nearly full context window. Good for:
- Complex analysis requiring full history
- Code refactoring across many files
- When token cost isn't a concern

### Example 3: Balanced (Default)

```bash
hex  # Uses defaults: 180000 tokens, prune strategy
```

Good balance of context and cost. Suitable for most use cases.

## See Also

- [Architecture](./ARCHITECTURE.md) - System design
- [Tools](./TOOLS.md) - Tool system documentation
- [User Guide](./USER_GUIDE.md) - General usage guide
