// ABOUTME: Command registry for storing and retrieving loaded commands
// ABOUTME: Thread-safe storage with lookup by name and listing functionality

package commands

import (
	"fmt"
	"sort"
	"sync"
)

// Registry stores loaded commands and provides thread-safe access
type Registry struct {
	mu       sync.RWMutex
	commands map[string]*Command
}

// NewRegistry creates an empty command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*Command),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd *Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cmd.Name == "" {
		return fmt.Errorf("cannot register command with empty name")
	}

	r.commands[cmd.Name] = cmd
	return nil
}

// RegisterAll adds multiple commands to the registry
func (r *Registry) RegisterAll(commands []*Command) error {
	for _, cmd := range commands {
		if err := r.Register(cmd); err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves a command by name
func (r *Registry) Get(name string) (*Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, exists := r.commands[name]
	if !exists {
		return nil, fmt.Errorf("command not found: %s", name)
	}

	return cmd, nil
}

// Has checks if a command exists
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.commands[name]
	return exists
}

// List returns all registered command names, sorted alphabetically
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// All returns all registered commands, sorted by name
func (r *Registry) All() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]*Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}

	// Sort by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	return commands
}

// Count returns the number of registered commands
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.commands)
}

// Clear removes all commands from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands = make(map[string]*Command)
}

// Remove removes a command by name
func (r *Registry) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[name]; !exists {
		return false
	}

	delete(r.commands, name)
	return true
}
