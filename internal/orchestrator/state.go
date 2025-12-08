// ABOUTME: Agent orchestrator state management
// ABOUTME: Defines agent states and transitions
package orchestrator

// AgentState represents the current state of the agent
type AgentState string

const (
	// StateIdle indicates agent is not active
	StateIdle AgentState = "idle"

	// StateStreaming indicates agent is streaming a response
	StateStreaming AgentState = "streaming"

	// StateAwaitingApproval indicates agent is waiting for tool approval
	StateAwaitingApproval AgentState = "awaiting_approval"

	// StateExecutingTool indicates agent is executing a tool
	StateExecutingTool AgentState = "executing_tool"

	// StateComplete indicates agent has completed successfully
	StateComplete AgentState = "complete"

	// StateError indicates agent encountered an error
	StateError AgentState = "error"
)

// String returns the string representation of the state
func (s AgentState) String() string {
	return string(s)
}
