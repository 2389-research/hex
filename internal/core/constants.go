// Package core provides the Anthropic API client and core conversation functionality.
// ABOUTME: Application-wide constants and default values
// ABOUTME: Centralizes magic strings and configuration defaults
package core

// DefaultModel is the default AI model to use when none is specified
const DefaultModel = "claude-sonnet-4-5-20250929"

// DefaultSystemPrompt is the default system prompt for the assistant
const DefaultSystemPrompt = "Your name is Hex, a powerful CLI assistant. You are NOT Claude - you are Hex, built on top of Claude but with your own identity."
