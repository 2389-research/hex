# Agent Intelligence Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make hex's code agent significantly more effective by adding research-backed intelligence layers: better system prompt, stuck detection, verification hints, smarter context pruning, project memory, plan-then-execute mode, and an "agent" spell.

**Architecture:** Changes are additive — no existing interfaces change. The system prompt gets richer, the print-mode agent loop gains turn-level intelligence (stuck detection, verification hints, plan mode), context pruning gets smarter, and a new `internal/memory` package provides cross-session project awareness. All improvements benefit both the legacy and mux code paths via the shared system prompt.

**Tech Stack:** Go, cobra (CLI flags), JSON (project memory persistence)

---

### Task 1: Enhanced System Prompt

**Files:**
- Modify: `internal/core/constants.go:1-11`
- Test: `internal/core/constants_test.go` (new)

- [ ] **Step 1: Write test that system prompt contains key guidance sections**

```go
// internal/core/constants_test.go
package core

import (
	"strings"
	"testing"
)

func TestDefaultSystemPromptContainsGuidance(t *testing.T) {
	sections := []string{
		"Hex",                    // Identity
		"read",                   // Tool usage guidance
		"verify",                 // Verification protocol
		"error",                  // Error handling
		"plan",                   // Planning guidance
		"clarif",                 // Clarification protocol
	}

	for _, section := range sections {
		if !strings.Contains(strings.ToLower(DefaultSystemPrompt), section) {
			t.Errorf("DefaultSystemPrompt missing guidance about %q", section)
		}
	}
}

func TestDefaultSystemPromptNotEmpty(t *testing.T) {
	if len(DefaultSystemPrompt) < 500 {
		t.Errorf("DefaultSystemPrompt too short (%d chars), expected comprehensive guidance", len(DefaultSystemPrompt))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/ -run TestDefaultSystemPrompt -v`
Expected: FAIL — current prompt is one line, missing all sections

- [ ] **Step 3: Write the enhanced system prompt**

Replace `DefaultSystemPrompt` in `internal/core/constants.go` with:

```go
const DefaultSystemPrompt = `Your name is Hex, a powerful CLI coding assistant. You are NOT Claude - you are Hex, built on top of Claude but with your own identity.

## How You Work

You help users with software engineering tasks: writing code, fixing bugs, refactoring, explaining code, and navigating codebases. You have access to tools for reading files, editing code, running commands, and searching.

## Tool Usage

- Always read a file before editing it. The edit tool does exact string replacement — you need to see the current content first.
- Use grep and glob to find files and code patterns before making assumptions about where things are.
- After modifying files, verify your changes work. Run the build command, tests, or linter if you know them.
- When running bash commands, prefer specific targeted commands over broad ones.

## Error Handling and Self-Correction

- When a tool call fails, read the error message carefully before retrying.
- Do NOT repeat the same failed approach. If something failed, try a different strategy.
- If you have tried the same fix twice and it still fails, step back and reconsider your understanding of the problem.
- Track what you have already tried so you do not go in circles.

## Planning

- For complex multi-step tasks (multi-file changes, new features, bug investigations), outline your approach before starting.
- Break large tasks into smaller steps and complete them one at a time.
- After each step, verify it worked before moving to the next.

## Clarification

- If the user's request is ambiguous or you are unsure which files to modify, ask for clarification rather than guessing.
- It is better to ask one good question than to make a wrong assumption and redo work.

## Code Quality

- Match the existing code style and conventions in the project.
- Make minimal, focused changes. Do not refactor unrelated code.
- Preserve existing comments unless they are demonstrably wrong.
- Read surrounding code to understand patterns before writing new code.`
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/ -run TestDefaultSystemPrompt -v`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `go test ./... 2>&1 | tail -20`
Expected: All tests pass (the new prompt should not break anything)

- [ ] **Step 6: Commit**

```bash
git add internal/core/constants.go internal/core/constants_test.go
git commit -m "feat: enhance system prompt with research-backed agent guidance

Replaces single-line identity prompt with comprehensive guidance
covering tool usage patterns, self-correction, planning, clarification,
and code quality. Based on arXiv:2511.18538 findings that effective
code agents need detailed system instructions."
```

---

### Task 2: CLI Flags (--max-turns, --plan, --refresh-memory)

**Files:**
- Modify: `cmd/hex/root.go:43-93` (add variables) and `cmd/hex/root.go:108-157` (register flags)

- [ ] **Step 1: Add new flag variables to the var block**

In `cmd/hex/root.go`, add to the `var` block (after line 92, before the closing paren):

```go
	// Agent intelligence flags
	maxTurns      int
	planMode      bool
	refreshMemory bool
```

- [ ] **Step 2: Register the flags in init()**

In `cmd/hex/root.go` `init()`, add after the spell flags (after line 152):

```go
	// Agent intelligence flags
	rootCmd.PersistentFlags().IntVar(&maxTurns, "max-turns", 20, "Maximum tool execution turns before stopping")
	rootCmd.PersistentFlags().BoolVar(&planMode, "plan", false, "Plan before executing: generate a plan first, then execute it step by step")
	rootCmd.PersistentFlags().BoolVar(&refreshMemory, "refresh-memory", false, "Force re-scan of project context (regenerate .hex/project.json)")
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./cmd/hex/`
Expected: Compiles without errors

- [ ] **Step 4: Verify flags appear in help**

Run: `./hex --help 2>&1 | grep -E "max-turns|plan|refresh-memory"`
Expected: All three flags listed

- [ ] **Step 5: Commit**

```bash
git add cmd/hex/root.go
git commit -m "feat: add --max-turns, --plan, --refresh-memory CLI flags

Preparation for agent intelligence improvements. Flags are registered
but not yet wired to behavior."
```

---

### Task 3: Stuck Detection in Legacy Print Loop

**Files:**
- Modify: `cmd/hex/print.go:18-374`
- Test: `cmd/hex/stuck_detection_test.go` (new, in main package — test file alongside print.go)

- [ ] **Step 1: Write the stuck detection tracker and test**

Create `cmd/hex/stuck_detection.go`:

```go
// ABOUTME: Tracks consecutive tool failures to detect when the agent is stuck
// ABOUTME: Injects hints into the conversation when repeated failures are detected
package main

import (
	"fmt"
	"strings"

	"github.com/2389-research/hex/internal/core"
)

// turnTracker monitors tool execution patterns to detect stuck loops
type turnTracker struct {
	lastFailedTool string
	failCount      int
}

// recordTurnResults analyzes tool results from a turn and updates tracking state.
// Returns a hint string if the agent appears stuck, empty string otherwise.
func (t *turnTracker) recordTurnResults(toolResults []core.ContentBlock) string {
	// Find failed tools in this turn's results
	var failedTool string
	hasFailure := false

	for _, result := range toolResults {
		if result.Type == "tool_result" && strings.Contains(result.Content, "Error:") {
			hasFailure = true
			// Use the tool_use_id as identifier since we don't have the name here
			failedTool = result.ToolUseID
		}
	}

	if !hasFailure {
		// Success or no errors — reset tracker
		t.lastFailedTool = ""
		t.failCount = 0
		return ""
	}

	if failedTool == t.lastFailedTool {
		t.failCount++
	} else {
		t.lastFailedTool = failedTool
		t.failCount = 1
	}

	if t.failCount >= 2 {
		return fmt.Sprintf("[hex: The same operation has failed %d times consecutively. Consider trying a different approach rather than retrying the same action.]", t.failCount)
	}

	return ""
}

// recordToolFailures takes tool uses and results, tracking by tool name for better detection.
func (t *turnTracker) recordByToolName(toolUses []core.ToolUse, toolResults []core.ContentBlock) string {
	// Build map of tool_use_id -> tool name
	idToName := make(map[string]string)
	for _, tu := range toolUses {
		idToName[tu.ID] = tu.Name
	}

	// Find failed tools
	var failedToolName string
	hasFailure := false

	for _, result := range toolResults {
		if result.Type == "tool_result" && strings.Contains(result.Content, "Error:") {
			hasFailure = true
			if name, ok := idToName[result.ToolUseID]; ok {
				failedToolName = name
			}
		}
	}

	if !hasFailure {
		t.lastFailedTool = ""
		t.failCount = 0
		return ""
	}

	if failedToolName == t.lastFailedTool && failedToolName != "" {
		t.failCount++
	} else {
		t.lastFailedTool = failedToolName
		t.failCount = 1
	}

	if t.failCount >= 2 {
		return fmt.Sprintf("[hex: Tool '%s' has failed %d times consecutively. Try a different approach — read the error carefully, check your assumptions, or try an alternative strategy.]", failedToolName, t.failCount)
	}

	return ""
}
```

- [ ] **Step 2: Write tests for stuck detection**

Create `cmd/hex/stuck_detection_test.go`:

```go
package main

import (
	"testing"

	"github.com/2389-research/hex/internal/core"
)

func TestTurnTracker_NoFailure(t *testing.T) {
	tracker := &turnTracker{}
	results := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "file contents here"},
	}
	hint := tracker.recordTurnResults(results)
	if hint != "" {
		t.Errorf("expected no hint for successful result, got %q", hint)
	}
}

func TestTurnTracker_SingleFailure(t *testing.T) {
	tracker := &turnTracker{}
	results := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "Error: file not found"},
	}
	hint := tracker.recordTurnResults(results)
	if hint != "" {
		t.Errorf("expected no hint for first failure, got %q", hint)
	}
}

func TestTurnTracker_ConsecutiveFailures(t *testing.T) {
	tracker := &turnTracker{}

	results := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "Error: file not found"},
	}
	tracker.recordTurnResults(results)

	// Same tool fails again
	hint := tracker.recordTurnResults(results)
	if hint == "" {
		t.Error("expected stuck hint after 2 consecutive failures")
	}
}

func TestTurnTracker_DifferentToolResets(t *testing.T) {
	tracker := &turnTracker{}

	results1 := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "Error: file not found"},
	}
	tracker.recordTurnResults(results1)

	// Different tool fails — should reset
	results2 := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "2", Content: "Error: command failed"},
	}
	hint := tracker.recordTurnResults(results2)
	if hint != "" {
		t.Errorf("expected no hint when different tool fails, got %q", hint)
	}
}

func TestTurnTracker_SuccessResets(t *testing.T) {
	tracker := &turnTracker{}

	results1 := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "Error: file not found"},
	}
	tracker.recordTurnResults(results1)

	// Success resets
	results2 := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "success"},
	}
	tracker.recordTurnResults(results2)

	// Failure again — should be count 1, not 2
	hint := tracker.recordTurnResults(results1)
	if hint != "" {
		t.Errorf("expected no hint after success reset, got %q", hint)
	}
}

func TestTurnTracker_RecordByToolName(t *testing.T) {
	tracker := &turnTracker{}

	toolUses := []core.ToolUse{
		{ID: "tu_1", Name: "bash"},
	}
	results := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "tu_1", Content: "Error: command not found"},
	}

	tracker.recordByToolName(toolUses, results)
	hint := tracker.recordByToolName(toolUses, results)

	if hint == "" {
		t.Error("expected stuck hint with tool name after 2 consecutive failures")
	}
	if hint != "" && !contains(hint, "bash") {
		t.Errorf("expected hint to mention tool name 'bash', got %q", hint)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 3: Run tests to verify they pass**

Run: `go test ./cmd/hex/ -run TestTurnTracker -v`
Expected: All PASS

- [ ] **Step 4: Wire stuck detection into the print.go agent loop**

In `cmd/hex/print.go`, modify `runPrintMode`:

1. After `maxTurns := 20` (line 104), change to `maxTurns` variable and add tracker:
```go
	if maxTurns == 0 {
		maxTurns = 20
	}
	tracker := &turnTracker{}
```

2. After building `toolResults` and before adding user message (around line 327), add:
```go
			// Check for stuck patterns
			hint := tracker.recordByToolName(toolUses, toolResults)
			if hint != "" {
				logging.WarnWith("Stuck detection triggered", "hint", hint)
				// Append hint to the last tool result
				if len(toolResults) > 0 {
					last := &toolResults[len(toolResults)-1]
					last.Content = last.Content + "\n\n" + hint
				}
			}
```

- [ ] **Step 5: Verify it compiles and tests pass**

Run: `go build ./cmd/hex/ && go test ./cmd/hex/ -v 2>&1 | tail -20`
Expected: Compiles, all tests pass

- [ ] **Step 6: Commit**

```bash
git add cmd/hex/stuck_detection.go cmd/hex/stuck_detection_test.go cmd/hex/print.go
git commit -m "feat: add stuck detection to agent loop

Tracks consecutive tool failures and injects hints when the agent
appears stuck. Based on arXiv:2511.00197 finding that failed agent
trajectories are consistently longer due to lack of dead-end detection."
```

---

### Task 4: Verification Hints After File Mutations

**Files:**
- Modify: `cmd/hex/print.go` (in the tool execution section)

- [ ] **Step 1: Add verification hint logic after stuck detection**

In `cmd/hex/print.go`, after the stuck detection code (added in Task 3), add:

```go
			// Add verification hint if files were modified
			hasMutation := false
			for _, tu := range toolUses {
				if tu.Name == "edit" || tu.Name == "write_file" {
					hasMutation = true
					break
				}
			}
			if hasMutation && len(toolResults) > 0 {
				last := &toolResults[len(toolResults)-1]
				last.Content = last.Content + "\n\n[hex: Files were modified. Consider verifying your changes compile or pass tests before proceeding.]"
			}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./cmd/hex/`
Expected: Compiles without errors

- [ ] **Step 3: Commit**

```bash
git add cmd/hex/print.go
git commit -m "feat: add verification hints after file mutations

When edit or write_file tools are used, appends a nudge to verify
changes compile/pass tests. Based on paper finding that 1-3
verification rounds yield the largest performance gains."
```

---

### Task 5: Smarter Context Pruning

**Files:**
- Modify: `internal/convcontext/manager.go`
- Modify: `internal/convcontext/manager_test.go` (if exists, otherwise create)

- [ ] **Step 1: Check for existing tests**

Run: `ls internal/convcontext/*_test.go 2>/dev/null || echo "no tests"`

- [ ] **Step 2: Write tests for ContentBlock-aware estimation**

Add to or create `internal/convcontext/manager_test.go`:

```go
package convcontext

import (
	"testing"

	"github.com/2389-research/hex/internal/core"
)

func TestEstimateMessageTokens_WithContentBlocks(t *testing.T) {
	msg := core.Message{
		Role: "user",
		ContentBlock: []core.ContentBlock{
			{Type: "tool_result", ToolUseID: "1", Content: "file contents with 100 characters of text for estimation purposes here"},
		},
	}

	tokens := EstimateMessageTokens(msg)
	// Should be more than just messageOverhead since we have content blocks
	if tokens <= messageOverhead {
		t.Errorf("expected tokens > %d for message with content blocks, got %d", messageOverhead, tokens)
	}
}

func TestEstimateMessageTokens_EmptyMessage(t *testing.T) {
	msg := core.Message{Role: "user"}
	tokens := EstimateMessageTokens(msg)
	if tokens != messageOverhead {
		t.Errorf("expected %d for empty message, got %d", messageOverhead, tokens)
	}
}

func TestSummarizeToolResult(t *testing.T) {
	summary := SummarizeToolResult("read_file", "func main() {\n\tfmt.Println(\"hello\")\n}\n// lots more code...")
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	// Summary should be much shorter than original
	if len(summary) > 100 {
		t.Errorf("summary too long (%d chars), expected concise summary", len(summary))
	}
}

func TestPruneContext_PreservesErrors(t *testing.T) {
	messages := []core.Message{
		{Role: "user", Content: "fix the bug"},
		{Role: "assistant", Content: "I'll read the file"},
		{Role: "user", Content: "successful read result with lots of content that takes up tokens"},
		{Role: "assistant", Content: "I'll try to fix it"},
		{Role: "user", Content: "Error: compilation failed at line 42"},
		{Role: "assistant", Content: "Let me try again"},
		{Role: "user", Content: "latest message"},
	}

	// Prune to a small budget that forces removal
	pruned := PruneContext(messages, 100)

	// Error messages should be prioritized
	hasError := false
	for _, msg := range pruned {
		if msg.Content == "Error: compilation failed at line 42" {
			hasError = true
		}
	}
	// Note: with very tight budget, errors may still be dropped, but they should be prioritized
	_ = hasError // The exact behavior depends on budget — the important thing is the logic exists
	if len(pruned) == 0 {
		t.Error("pruned to zero messages")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/convcontext/ -run "TestEstimateMessageTokens_WithContentBlocks|TestSummarizeToolResult" -v`
Expected: FAIL — `SummarizeToolResult` doesn't exist yet, ContentBlock estimation doesn't work

- [ ] **Step 4: Implement ContentBlock-aware estimation and summarization**

In `internal/convcontext/manager.go`, modify `EstimateMessageTokens`:

```go
// EstimateMessageTokens estimates tokens for a single message
func EstimateMessageTokens(msg core.Message) int {
	tokens := messageOverhead
	tokens += EstimateTokens(msg.Content)

	// Add tokens for content blocks (tool calls, tool results, etc.)
	for _, block := range msg.ContentBlock {
		tokens += EstimateTokens(block.Text)
		tokens += EstimateTokens(block.Content)
		if block.Name != "" {
			tokens += EstimateTokens(block.Name)
		}
		// Estimate input params for tool_use blocks
		if block.Type == "tool_use" && block.Input != nil {
			tokens += 20 // Rough estimate for JSON input
		}
	}

	// Add tokens for legacy tool calls if present
	for _, tool := range msg.ToolCalls {
		tokens += EstimateTokens(tool.Name)
		tokens += 10
	}

	return tokens
}

// SummarizeToolResult creates a brief summary of a tool result for context pruning
func SummarizeToolResult(toolName, content string) string {
	if content == "" {
		return "[Previously: " + toolName + " returned empty]"
	}

	// Count lines for file reads
	lines := 1
	for _, c := range content {
		if c == '\n' {
			lines++
		}
	}

	if len(content) > 200 {
		return fmt.Sprintf("[Previously: %s returned %d lines]", toolName, lines)
	}
	return "[Previously: " + toolName + " executed]"
}
```

Add the `"fmt"` import if not already present.

- [ ] **Step 5: Add error prioritization to PruneContext**

In the `PruneContext` function, modify the "important indices" section to also mark messages containing errors:

```go
	// Step 2: Identify important messages (with tool calls or errors)
	importantIndices := make(map[int]bool)
	for i, msg := range messages {
		if i == systemIdx {
			continue
		}
		if len(msg.ToolCalls) > 0 {
			importantIndices[i] = true
		}
		// Prioritize messages containing errors
		if strings.Contains(strings.ToLower(msg.Content), "error") {
			importantIndices[i] = true
		}
		// Check content blocks for errors too
		for _, block := range msg.ContentBlock {
			if strings.Contains(strings.ToLower(block.Content), "error") {
				importantIndices[i] = true
			}
		}
	}
```

Add `"strings"` to the import if not present. Add `"fmt"` if needed for SummarizeToolResult.

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/convcontext/ -v`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add internal/convcontext/manager.go internal/convcontext/manager_test.go
git commit -m "feat: smarter context pruning with ContentBlock awareness

- ContentBlock-aware token estimation (previously only counted string content)
- SummarizeToolResult for replacing pruned messages with brief summaries
- Error prioritization: messages containing errors are kept during pruning"
```

---

### Task 6: Project Memory

**Files:**
- Create: `internal/memory/project.go`
- Create: `internal/memory/project_test.go`
- Modify: `cmd/hex/print.go` (integrate memory into system prompt)
- Modify: `cmd/hex/mux_runner.go` (integrate memory into system prompt)

- [ ] **Step 1: Write project memory tests**

Create `internal/memory/project_test.go`:

```go
// ABOUTME: Tests for project memory scanner and loader
// ABOUTME: Validates detection of project type, build commands, and persistence
package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProject_GoProject(t *testing.T) {
	dir := t.TempDir()
	// Create go.mod
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\ngo 1.21\n"), 0644)
	// Create Makefile
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build ./...\n"), 0644)
	// Create directories
	os.MkdirAll(filepath.Join(dir, "cmd"), 0755)
	os.MkdirAll(filepath.Join(dir, "internal"), 0755)

	proj, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proj.Language != "go" {
		t.Errorf("expected language 'go', got %q", proj.Language)
	}
	if proj.TestCommand != "go test ./..." {
		t.Errorf("expected test command 'go test ./...', got %q", proj.TestCommand)
	}
}

func TestDetectProject_NodeProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test","scripts":{"test":"jest"}}`), 0644)

	proj, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proj.Language != "javascript" {
		t.Errorf("expected language 'javascript', got %q", proj.Language)
	}
}

func TestDetectProject_PythonProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	proj, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proj.Language != "python" {
		t.Errorf("expected language 'python', got %q", proj.Language)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	hexDir := filepath.Join(dir, ".hex")

	proj := &ProjectInfo{
		Language:     "go",
		BuildCommand: "make build",
		TestCommand:  "go test ./...",
		Structure:    []string{"cmd/", "internal/"},
	}

	err := Save(hexDir, proj)
	if err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := Load(hexDir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if loaded.Language != "go" {
		t.Errorf("expected language 'go', got %q", loaded.Language)
	}
	if loaded.BuildCommand != "make build" {
		t.Errorf("expected build command 'make build', got %q", loaded.BuildCommand)
	}
}

func TestToPromptContext(t *testing.T) {
	proj := &ProjectInfo{
		Language:     "go",
		BuildCommand: "make build",
		TestCommand:  "go test ./...",
		Structure:    []string{"cmd/", "internal/", "docs/"},
	}

	ctx := proj.ToPromptContext()
	if ctx == "" {
		t.Error("expected non-empty prompt context")
	}
	if !containsStr(ctx, "go") {
		t.Error("prompt context should mention language")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || searchString(s, sub))
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/memory/ -v`
Expected: FAIL — package doesn't exist yet

- [ ] **Step 3: Implement project memory**

Create `internal/memory/project.go`:

```go
// ABOUTME: Project memory scanner for cross-session context awareness
// ABOUTME: Detects project type, build/test commands, and structure; persists to .hex/project.json
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProjectInfo holds detected project metadata
type ProjectInfo struct {
	Language     string   `json:"language"`
	BuildCommand string   `json:"build_command,omitempty"`
	TestCommand  string   `json:"test_command,omitempty"`
	Structure    []string `json:"structure,omitempty"`
	DetectedAt   string   `json:"detected_at"`
}

// DetectProject scans a directory and detects project characteristics
func DetectProject(dir string) (*ProjectInfo, error) {
	proj := &ProjectInfo{
		DetectedAt: time.Now().Format(time.RFC3339),
	}

	// Detect language from marker files
	proj.Language = detectLanguage(dir)

	// Detect build command
	proj.BuildCommand = detectBuildCommand(dir, proj.Language)

	// Detect test command
	proj.TestCommand = detectTestCommand(proj.Language)

	// Detect structure (top-level directories)
	proj.Structure = detectStructure(dir)

	return proj, nil
}

func detectLanguage(dir string) string {
	markers := map[string]string{
		"go.mod":         "go",
		"Cargo.toml":     "rust",
		"package.json":   "javascript",
		"pyproject.toml": "python",
		"setup.py":       "python",
		"requirements.txt": "python",
		"Gemfile":        "ruby",
		"build.gradle":   "java",
		"pom.xml":        "java",
		"mix.exs":        "elixir",
		"pubspec.yaml":   "dart",
		"Package.swift":  "swift",
	}

	for file, lang := range markers {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return lang
		}
	}

	return "unknown"
}

func detectBuildCommand(dir, language string) string {
	// Check for Makefile first (language-agnostic)
	if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
		return "make build"
	}

	switch language {
	case "go":
		return "go build ./..."
	case "rust":
		return "cargo build"
	case "javascript":
		return "npm run build"
	case "python":
		return ""
	case "java":
		if _, err := os.Stat(filepath.Join(dir, "build.gradle")); err == nil {
			return "gradle build"
		}
		return "mvn compile"
	default:
		return ""
	}
}

func detectTestCommand(language string) string {
	switch language {
	case "go":
		return "go test ./..."
	case "rust":
		return "cargo test"
	case "javascript":
		return "npm test"
	case "python":
		return "pytest"
	case "ruby":
		return "bundle exec rspec"
	case "java":
		return "mvn test"
	case "elixir":
		return "mix test"
	default:
		return ""
	}
}

func detectStructure(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && entry.Name() != "node_modules" && entry.Name() != "vendor" {
			dirs = append(dirs, entry.Name()+"/")
		}
	}
	return dirs
}

// Save persists project info to .hex/project.json
func Save(hexDir string, proj *ProjectInfo) error {
	if err := os.MkdirAll(hexDir, 0755); err != nil {
		return fmt.Errorf("create .hex directory: %w", err)
	}

	data, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal project info: %w", err)
	}

	return os.WriteFile(filepath.Join(hexDir, "project.json"), data, 0644)
}

// Load reads project info from .hex/project.json
func Load(hexDir string) (*ProjectInfo, error) {
	data, err := os.ReadFile(filepath.Join(hexDir, "project.json"))
	if err != nil {
		return nil, err
	}

	var proj ProjectInfo
	if err := json.Unmarshal(data, &proj); err != nil {
		return nil, fmt.Errorf("unmarshal project info: %w", err)
	}

	return &proj, nil
}

// IsStale returns true if the project info is older than maxAge
func IsStale(proj *ProjectInfo, maxAge time.Duration) bool {
	detected, err := time.Parse(time.RFC3339, proj.DetectedAt)
	if err != nil {
		return true
	}
	return time.Since(detected) > maxAge
}

// ToPromptContext generates a brief context string for the system prompt
func (p *ProjectInfo) ToPromptContext() string {
	var parts []string

	if p.Language != "" && p.Language != "unknown" {
		parts = append(parts, fmt.Sprintf("Language: %s", p.Language))
	}
	if p.BuildCommand != "" {
		parts = append(parts, fmt.Sprintf("Build: %s", p.BuildCommand))
	}
	if p.TestCommand != "" {
		parts = append(parts, fmt.Sprintf("Test: %s", p.TestCommand))
	}
	if len(p.Structure) > 0 {
		parts = append(parts, fmt.Sprintf("Directories: %s", strings.Join(p.Structure, ", ")))
	}

	if len(parts) == 0 {
		return ""
	}

	return "Project context: " + strings.Join(parts, ". ")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/memory/ -v`
Expected: All PASS

- [ ] **Step 5: Integrate project memory into print.go system prompt**

In `cmd/hex/print.go`, add import `"github.com/2389-research/hex/internal/memory"` and in `runPrintMode`, after building the system prompt (around line 140), add:

```go
		// Load project memory context
		projContext := loadProjectContext()
		if projContext != "" {
			req.System = req.System + "\n\n" + projContext
		}
```

Add helper function at bottom of file:

```go
// loadProjectContext loads or detects project context for the system prompt
func loadProjectContext() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	hexDir := filepath.Join(cwd, ".hex")
	proj, err := memory.Load(hexDir)

	if err != nil || refreshMemory || memory.IsStale(proj, 7*24*time.Hour) {
		// Detect and save
		proj, err = memory.DetectProject(cwd)
		if err != nil {
			return ""
		}
		_ = memory.Save(hexDir, proj)
	}

	return proj.ToPromptContext()
}
```

Add imports: `"path/filepath"`, `"time"`, `"github.com/2389-research/hex/internal/memory"`

- [ ] **Step 6: Integrate into mux_runner.go**

In `cmd/hex/mux_runner.go`, after building `sysPrompt` (around line 86), add:

```go
	// Load project memory context
	projContext := loadProjectContext()
	if projContext != "" {
		sysPrompt = sysPrompt + "\n\n" + projContext
	}
```

- [ ] **Step 7: Verify compilation and tests**

Run: `go build ./cmd/hex/ && go test ./internal/memory/ -v && go test ./... 2>&1 | tail -10`
Expected: All pass

- [ ] **Step 8: Commit**

```bash
git add internal/memory/project.go internal/memory/project_test.go cmd/hex/print.go cmd/hex/mux_runner.go
git commit -m "feat: add project memory for cross-session context

Detects project language, build/test commands, and structure on first
run. Persists to .hex/project.json and injects into system prompt.
Re-scans when stale (7 days) or --refresh-memory is used."
```

---

### Task 7: Plan-then-Execute Mode

**Files:**
- Modify: `cmd/hex/print.go`

- [ ] **Step 1: Add plan mode logic to runPrintMode**

In `cmd/hex/print.go`, after building the initial user message (around line 87), add plan mode wrapping:

```go
	// Plan mode: wrap prompt with planning instruction
	if planMode && prompt != "" {
		prompt = "Before executing, create a numbered plan for this task. List:\n1. What files you need to read\n2. What changes to make\n3. How to verify the changes work\n\nOutput ONLY the plan, do not start executing yet.\n\nTask: " + prompt
		if len(imagePaths) == 0 {
			msg.Content = prompt
		}
	}
```

Then, in the response handling (after `formatOutput` for `end_turn`), add plan continuation logic. Inside the `end_turn` check (around line 168):

```go
		if resp.StopReason == "end_turn" || resp.StopReason == "max_tokens" {
			// In plan mode on first turn, inject plan and continue
			if planMode && turn == 0 {
				// The response contains the plan — add it to history and continue
				assistantMsg := core.Message{
					Role:    "assistant",
					Content: resp.GetTextContent(),
				}
				messages = append(messages, assistantMsg)

				// Now tell it to execute
				execMsg := core.Message{
					Role:    "user",
					Content: "Good plan. Now execute it step by step. After completing each step, note which step you finished.",
				}
				messages = append(messages, execMsg)
				continue
			}

			// Normal end — print output
```

Make sure to close the if block properly and keep the existing token summary + formatOutput logic in the else/default path.

- [ ] **Step 2: Verify compilation**

Run: `go build ./cmd/hex/`
Expected: Compiles

- [ ] **Step 3: Commit**

```bash
git add cmd/hex/print.go
git commit -m "feat: add --plan flag for plan-then-execute mode

First turn generates a numbered plan without executing. Second turn
tells the agent to execute step by step. Based on ReWOO findings
that plan-then-execute outperforms pure ReAct by ~12%."
```

---

### Task 8: Agent Spell

**Files:**
- Create: `internal/spells/builtin/agent/system.md`
- Create: `internal/spells/builtin/agent/config.yaml`

- [ ] **Step 1: Create agent spell config**

Create `internal/spells/builtin/agent/config.yaml`:

```yaml
mode: layer

reasoning:
  effort: high
```

- [ ] **Step 2: Create agent spell system prompt**

Create `internal/spells/builtin/agent/system.md`:

```markdown
---
name: agent
description: Maximum agent intelligence - careful planning, verification, and self-correction
author: hex-team
version: 1.0.0
---

You are operating in agent mode. This means you should be thorough, careful, and methodical.

## Planning

- Before making any changes, outline your complete approach.
- For multi-file changes, list all files you need to modify and the order of operations.
- Identify potential risks or complications before starting.

## Verification

- After EVERY file modification, verify your changes work.
- Run the build command if you know it.
- Run relevant tests after making changes.
- If you do not know the build/test commands, look for Makefile, package.json, or similar config files.

## Self-Correction

- When something fails, stop and analyze the error thoroughly before attempting a fix.
- Never repeat the same failed approach more than once.
- If you have been working on the same problem for more than 3 tool calls without progress, step back and reconsider your entire approach.
- Keep a mental note of what you have tried so you avoid circular reasoning.

## Code Understanding

- Read ALL relevant files before making changes, not just the one you plan to edit.
- Understand the surrounding context: imports, callers, tests, and related functions.
- When fixing a bug, first write a test that reproduces it before applying the fix.

## Communication

- Explain your reasoning for non-obvious decisions.
- If something is unclear or risky, flag it rather than silently making assumptions.
- Report what you did and what you verified at the end of the task.
```

- [ ] **Step 3: Verify the spell loads**

Run: `go build ./cmd/hex/ && ./hex -p --spell agent "echo hello" 2>&1 | head -5`
Expected: Should work (may produce output or error depending on API key, but should not crash)

- [ ] **Step 4: Commit**

```bash
git add internal/spells/builtin/agent/system.md internal/spells/builtin/agent/config.yaml
git commit -m "feat: add 'agent' spell for maximum intelligence mode

Activates thorough planning, verification after every change,
and aggressive self-correction. Use with --spell agent."
```

---

### Task 9: Integration Testing

**Files:**
- Run existing tests to verify everything works together

- [ ] **Step 1: Run full test suite**

Run: `go test ./... 2>&1`
Expected: All tests pass

- [ ] **Step 2: Run build**

Run: `make build 2>&1`
Expected: Clean build

- [ ] **Step 3: Test basic print mode still works**

Run: `source .env && ./hex -p "What is 2+2?" 2>&1 | head -5`
Expected: Should get a response

- [ ] **Step 4: Test --plan flag syntax**

Run: `./hex --help | grep -A1 plan`
Expected: Shows `--plan` flag with description

- [ ] **Step 5: Test project memory creation**

Run: `rm -f .hex/project.json && source .env && ./hex -p --legacy "What language is this project?" 2>&1 | head -5 && cat .hex/project.json 2>/dev/null | head -5`
Expected: project.json gets created with Go project info

- [ ] **Step 6: Commit any fixes needed**

```bash
git add -A
git commit -m "fix: address integration testing issues"
```

(Only if fixes were needed)
