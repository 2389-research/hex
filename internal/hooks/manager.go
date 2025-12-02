// ABOUTME: Hook manager for lifecycle event automation
// ABOUTME: Coordinates hook execution and result collection

package hooks

import (
	"context"
)

// Manager manages hook execution for lifecycle events
type Manager struct {
	config   HooksConfig
	executor *Executor
}

// NewManager creates a new hook manager
func NewManager(config HooksConfig) *Manager {
	return &Manager{
		config:   config,
		executor: NewExecutor(),
	}
}

// Trigger executes all hooks for a given event
func (m *Manager) Trigger(ctx context.Context, event HookEvent, data EventData) ([]*HookResult, error) {
	hooks, ok := m.config[event]
	if !ok || len(hooks) == 0 {
		return nil, nil
	}

	results := make([]*HookResult, 0, len(hooks))
	for _, hookConfig := range hooks {
		result, err := m.executor.Run(ctx, event, hookConfig, data)
		if err != nil {
			// Log error but continue with other hooks
			result = &HookResult{
				Success: false,
				Error:   err.Error(),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// TriggerAsync executes hooks without blocking
func (m *Manager) TriggerAsync(event HookEvent, data EventData) {
	go func() {
		_, _ = m.Trigger(context.Background(), event, data)
	}()
}
