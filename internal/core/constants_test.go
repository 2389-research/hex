// ABOUTME: Tests for application-wide constants
// ABOUTME: Verifies DefaultSystemPrompt contains required guidance sections
package core_test

import (
	"strings"
	"testing"

	"github.com/2389-research/hex/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSystemPromptIdentity(t *testing.T) {
	assert.Contains(t, strings.ToLower(core.DefaultSystemPrompt), "hex",
		"system prompt must establish Hex identity")
}

func TestDefaultSystemPromptToolUsage(t *testing.T) {
	assert.Contains(t, strings.ToLower(core.DefaultSystemPrompt), "read",
		"system prompt must include tool usage guidance about reading files")
}

func TestDefaultSystemPromptVerification(t *testing.T) {
	assert.Contains(t, strings.ToLower(core.DefaultSystemPrompt), "verify",
		"system prompt must include guidance about verifying changes")
}

func TestDefaultSystemPromptErrorHandling(t *testing.T) {
	assert.Contains(t, strings.ToLower(core.DefaultSystemPrompt), "error",
		"system prompt must include error handling guidance")
}

func TestDefaultSystemPromptPlanning(t *testing.T) {
	assert.Contains(t, strings.ToLower(core.DefaultSystemPrompt), "plan",
		"system prompt must include planning guidance")
}

func TestDefaultSystemPromptClarification(t *testing.T) {
	assert.Contains(t, strings.ToLower(core.DefaultSystemPrompt), "clarif",
		"system prompt must include clarification guidance")
}

func TestDefaultSystemPromptLength(t *testing.T) {
	assert.GreaterOrEqual(t, len(core.DefaultSystemPrompt), 500,
		"system prompt must be at least 500 characters to provide meaningful guidance")
}

func TestDefaultSystemPromptPersistence(t *testing.T) {
	assert.Contains(t, strings.ToLower(core.DefaultSystemPrompt), "persist",
		"system prompt must instruct agent to persist until task is solved")
}

func TestHeadlessGuidanceExists(t *testing.T) {
	assert.Contains(t, strings.ToLower(core.HeadlessGuidance), "non-interactive",
		"headless guidance must mention non-interactive mode")
	assert.Contains(t, strings.ToLower(core.HeadlessGuidance), "never ask",
		"headless guidance must tell agent not to ask questions")
}
