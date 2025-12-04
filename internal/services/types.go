// ABOUTME: Shared types and models for service layer
// ABOUTME: Defines domain models used across services

// Package services defines business logic interfaces and domain models for the application.
package services

import "time"

// Conversation represents a chat session
type Conversation struct {
	ID               string
	Title            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PromptTokens     int64
	CompletionTokens int64
	TotalCost        float64
	SummaryMessageID *string
	IsFavorite       bool
}

// Message represents a single message in a conversation
type Message struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	Provider       string
	Model          string
	IsSummary      bool
	CreatedAt      time.Time
}

// AgentCall represents a request to the agent
type AgentCall struct {
	ConversationID string
	Prompt         string
	Attachments    []string
	MaxTokens      int
}

// AgentResult represents the agent's response
type AgentResult struct {
	Text         string
	ToolCalls    []ToolCall
	PromptTokens int64
	OutputTokens int64
}

// ToolCall represents a tool execution request
type ToolCall struct {
	ID     string
	Name   string
	Input  map[string]interface{}
	Result string
}

// StreamEvent represents a streaming update from the agent
type StreamEvent struct {
	Type         string // "text", "tool_call", "done", "error"
	Text         string
	ToolCall     *ToolCall
	Error        error
	PromptTokens int64
	OutputTokens int64
}
