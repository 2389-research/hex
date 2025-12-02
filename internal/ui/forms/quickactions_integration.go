// Package forms provides beautiful huh-based forms for the clem TUI.
// ABOUTME: Integration layer between old QuickActionsRegistry and new huh forms
// ABOUTME: Converts registry actions to form actions and runs the form
package forms

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// QuickActionsResultMsg is sent when the quick actions form completes
type QuickActionsResultMsg struct {
	ActionName string
	Error      error
}

// ConvertRegistryAction converts an old-style action to a new form action
// This is used to bridge the gap between the existing registry and the new forms
func ConvertRegistryAction(name, description, category, keyBinding string, handler func(string) error) *QuickAction {
	return &QuickAction{
		Name:        name,
		Description: description,
		Category:    category,
		KeyBinding:  keyBinding,
		Handler:     handler,
	}
}

// RunQuickActionsFormAsync runs the quick actions form in a goroutine
// and returns a tea.Cmd that will send the result when complete
func RunQuickActionsFormAsync(actions []*QuickAction) tea.Cmd {
	return func() tea.Msg {
		form := NewQuickActionsForm(actions)
		actionName, err := form.Run()

		return &QuickActionsResultMsg{
			ActionName: actionName,
			Error:      err,
		}
	}
}

// QuickActionsRegistry integration - allows Model to use the new forms
// while maintaining backward compatibility with existing registry

// OldQuickAction represents the old action structure for conversion
type OldQuickAction struct {
	Name        string
	Description string
	Usage       string
	Handler     func(args string) error
}

// ConvertOldActions converts old-style actions to new form actions
func ConvertOldActions(oldActions []*OldQuickAction) []*QuickAction {
	actions := make([]*QuickAction, 0, len(oldActions))

	for _, old := range oldActions {
		// Determine category based on action name
		category := string(CategoryTools)
		switch old.Name {
		case "help", "back", "forward", "quit":
			category = string(CategoryNavigation)
		case "save", "export", "clear", "reset":
			category = string(CategorySettings)
		}

		actions = append(actions, &QuickAction{
			Name:        old.Name,
			Description: old.Description,
			Category:    category,
			KeyBinding:  "", // Old actions didn't have key bindings
			Handler:     old.Handler,
		})
	}

	return actions
}

// QuickActionError wraps an error with context about which action failed
type QuickActionError struct {
	ActionName string
	Err        error
}

func (e *QuickActionError) Error() string {
	return fmt.Sprintf("quick action %q failed: %v", e.ActionName, e.Err)
}

func (e *QuickActionError) Unwrap() error {
	return e.Err
}
