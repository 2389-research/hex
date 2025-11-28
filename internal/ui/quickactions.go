// ABOUTME: Quick actions menu system for keyboard shortcuts
// ABOUTME: Provides fuzzy-searchable command palette for tools and actions
package ui

import (
	"fmt"
	"strings"
	"sync"
)

// QuickAction represents a single quick action
type QuickAction struct {
	Name        string
	Description string
	Usage       string
	Handler     func(args string) error
}

// QuickActionsRegistry manages quick actions
type QuickActionsRegistry struct {
	actions map[string]*QuickAction
	mu      sync.RWMutex
}

// NewQuickActionsRegistry creates a new registry with built-in actions
func NewQuickActionsRegistry() *QuickActionsRegistry {
	r := &QuickActionsRegistry{
		actions: make(map[string]*QuickAction),
	}

	// Register built-in actions
	r.registerBuiltInActions()

	return r
}

// registerBuiltInActions registers the default quick actions
func (r *QuickActionsRegistry) registerBuiltInActions() {
	// Read file action
	_ = r.RegisterAction("read", "Read a file", "read <file>", func(args string) error {
		// This will be connected to the tool system in the model
		return fmt.Errorf("read action not yet connected to tool system")
	})

	// Grep action
	_ = r.RegisterAction("grep", "Search files with grep", "grep <pattern>", func(args string) error {
		return fmt.Errorf("grep action not yet connected to tool system")
	})

	// Web fetch action
	_ = r.RegisterAction("web", "Fetch web page", "web <url>", func(args string) error {
		return fmt.Errorf("web action not yet connected to tool system")
	})

	// Attach image action
	_ = r.RegisterAction("attach", "Attach an image", "attach <file>", func(args string) error {
		return fmt.Errorf("attach action not yet connected to tool system")
	})

	// Save conversation action
	_ = r.RegisterAction("save", "Save conversation", "save", func(args string) error {
		return fmt.Errorf("save action not yet connected to model")
	})

	// Export action
	_ = r.RegisterAction("export", "Export conversation as markdown", "export", func(args string) error {
		return fmt.Errorf("export action not yet connected to model")
	})
}

// RegisterAction adds a new quick action
func (r *QuickActionsRegistry) RegisterAction(name, description, usage string, handler func(string) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[name]; exists {
		return fmt.Errorf("action %s already registered", name)
	}

	r.actions[name] = &QuickAction{
		Name:        name,
		Description: description,
		Usage:       usage,
		Handler:     handler,
	}

	return nil
}

// GetAction retrieves an action by name
func (r *QuickActionsRegistry) GetAction(name string) (*QuickAction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	action, exists := r.actions[name]
	if !exists {
		return nil, fmt.Errorf("action %s not found", name)
	}

	return action, nil
}

// ListActions returns all registered actions
func (r *QuickActionsRegistry) ListActions() []*QuickAction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	actions := make([]*QuickAction, 0, len(r.actions))
	for _, action := range r.actions {
		actions = append(actions, action)
	}

	return actions
}

// Execute runs an action with the given arguments
func (r *QuickActionsRegistry) Execute(name, args string) error {
	action, err := r.GetAction(name)
	if err != nil {
		return err
	}

	return action.Handler(args)
}

// FuzzySearch searches actions by name using fuzzy matching
func (r *QuickActionsRegistry) FuzzySearch(query string) []*QuickAction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))

	// If query is empty, return all actions
	if query == "" {
		return r.ListActions()
	}

	// Collect matches with scores
	type match struct {
		action *QuickAction
		score  int
	}

	matches := make([]match, 0)

	for _, action := range r.actions {
		name := strings.ToLower(action.Name)

		// Score the match
		score := fuzzyScore(query, name)
		if score > 0 {
			matches = append(matches, match{
				action: action,
				score:  score,
			})
		}
	}

	// Sort by score (descending)
	// Simple bubble sort since we have few items
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[i].score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Extract just the actions
	result := make([]*QuickAction, len(matches))
	for i, m := range matches {
		result[i] = m.action
	}

	return result
}

// fuzzyScore calculates a simple fuzzy match score
// Higher score = better match
func fuzzyScore(query, target string) int {
	// Exact match gets highest score
	if query == target {
		return 1000
	}

	// Prefix match gets high score
	if strings.HasPrefix(target, query) {
		return 500
	}

	// Contains match gets medium score
	if strings.Contains(target, query) {
		return 250
	}

	// Try fuzzy match - all query chars must appear in order
	queryIdx := 0
	for i := 0; i < len(target) && queryIdx < len(query); i++ {
		if target[i] == query[queryIdx] {
			queryIdx++
		}
	}

	// If all query chars were found, give a score
	if queryIdx == len(query) {
		// Score based on how tight the match is
		return 100 + (100 / (len(target) - len(query) + 1))
	}

	return 0
}

// ParseActionCommand splits a command input into command and arguments
func ParseActionCommand(input string) (command, args string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ""
	}

	parts := strings.SplitN(input, " ", 2)
	command = parts[0]

	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	return command, args
}
