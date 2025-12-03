// Package approval manages persistent approval rules for tool execution.
// ABOUTME: Stores user's "Always Allow" and "Never Allow" decisions per tool name
package approval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// RuleDecision represents whether a tool is always allowed or never allowed
type RuleDecision string

const (
	// RuleAlwaysAllow means the tool can execute without prompting
	RuleAlwaysAllow RuleDecision = "always_allow"
	// RuleNeverAllow means the tool is permanently blocked
	RuleNeverAllow RuleDecision = "never_allow"
)

// Rules manages persistent approval decisions for tools
type Rules struct {
	mu    sync.RWMutex
	rules map[string]RuleDecision // tool name -> decision
	path  string                  // file path for persistence
}

// rulesFile represents the JSON structure for persistence
type rulesFile struct {
	Rules map[string]RuleDecision `json:"rules"`
}

// NewRules creates a new approval rules manager
func NewRules() (*Rules, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}

	rulesPath := filepath.Join(home, ".clem", "approval_rules.json")

	r := &Rules{
		rules: make(map[string]RuleDecision),
		path:  rulesPath,
	}

	// Load existing rules if file exists
	if err := r.load(); err != nil {
		// If file doesn't exist, that's OK - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("load rules: %w", err)
		}
	}

	return r, nil
}

// Check returns the rule for a tool, or empty string if no rule exists
func (r *Rules) Check(toolName string) RuleDecision {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.rules[toolName]
}

// SetAlwaysAllow marks a tool as always allowed
func (r *Rules) SetAlwaysAllow(toolName string) error {
	return r.set(toolName, RuleAlwaysAllow)
}

// SetNeverAllow marks a tool as never allowed
func (r *Rules) SetNeverAllow(toolName string) error {
	return r.set(toolName, RuleNeverAllow)
}

// Remove removes any rule for a tool (resets to prompt)
func (r *Rules) Remove(toolName string) error {
	r.mu.Lock()
	delete(r.rules, toolName)
	r.mu.Unlock()
	return r.save()
}

// set updates a rule and persists to disk
func (r *Rules) set(toolName string, decision RuleDecision) error {
	r.mu.Lock()
	r.rules[toolName] = decision
	r.mu.Unlock()
	return r.save()
}

// save persists rules to disk
func (r *Rules) save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create rules directory: %w", err)
	}

	// Marshal to JSON
	rf := rulesFile{Rules: r.rules}
	data, err := json.MarshalIndent(rf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rules: %w", err)
	}

	// Write to file
	if err := os.WriteFile(r.path, data, 0600); err != nil {
		return fmt.Errorf("write rules file: %w", err)
	}

	return nil
}

// load reads rules from disk
func (r *Rules) load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return err
	}

	var rf rulesFile
	if err := json.Unmarshal(data, &rf); err != nil {
		return fmt.Errorf("unmarshal rules: %w", err)
	}

	r.mu.Lock()
	r.rules = rf.Rules
	if r.rules == nil {
		r.rules = make(map[string]RuleDecision)
	}
	r.mu.Unlock()

	return nil
}

// List returns all current rules (for debugging/settings UI)
func (r *Rules) List() map[string]RuleDecision {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	rulesCopy := make(map[string]RuleDecision, len(r.rules))
	for k, v := range r.rules {
		rulesCopy[k] = v
	}
	return rulesCopy
}
