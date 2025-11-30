// ABOUTME: Autocomplete system with fuzzy matching for tools, files, and commands
// ABOUTME: Provides completion dropdown with keyboard navigation

package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
)

// CompletionProvider is an interface for providing completion suggestions
type CompletionProvider interface {
	// GetCompletions returns completion suggestions for the given input
	GetCompletions(input string) []Completion
}

// Completion represents a single completion suggestion
type Completion struct {
	// Value is the actual completion text
	Value string
	// Display is what gets shown in the UI (may include formatting)
	Display string
	// Description provides context about the completion
	Description string
	// Score is used for ranking (higher is better)
	Score int
}

// Autocomplete manages completion state and providers
type Autocomplete struct {
	// providers holds all registered completion providers
	providers map[string]CompletionProvider
	// active indicates if autocomplete is currently showing
	active bool
	// completions holds the current list of suggestions
	completions []Completion
	// selectedIndex is the currently highlighted completion
	selectedIndex int
	// maxCompletions limits the number of suggestions shown
	maxCompletions int
	// currentInput is the text being completed
	currentInput string
	// currentProvider is which provider is active
	currentProvider string
}

// NewAutocomplete creates a new autocomplete instance
func NewAutocomplete() *Autocomplete {
	ac := &Autocomplete{
		providers:       make(map[string]CompletionProvider),
		active:          false,
		completions:     []Completion{},
		selectedIndex:   0,
		maxCompletions:  10,
		currentInput:    "",
		currentProvider: "",
	}

	// Register default providers
	ac.RegisterProvider("tool", NewToolProvider(nil)) // Will be updated when model is set
	ac.RegisterProvider("file", NewFileProvider())
	ac.RegisterProvider("history", NewHistoryProvider())

	return ac
}

// RegisterProvider adds a completion provider
func (ac *Autocomplete) RegisterProvider(name string, provider CompletionProvider) {
	ac.providers[name] = provider
}

// GetProvider retrieves a registered provider by name
func (ac *Autocomplete) GetProvider(name string) (CompletionProvider, bool) {
	provider, ok := ac.providers[name]
	return provider, ok
}

// Show activates autocomplete for the given input and provider
func (ac *Autocomplete) Show(input string, providerName string) {
	provider, ok := ac.providers[providerName]
	if !ok {
		return
	}

	ac.currentInput = input
	ac.currentProvider = providerName
	ac.completions = provider.GetCompletions(input)

	// Limit to max completions
	if len(ac.completions) > ac.maxCompletions {
		ac.completions = ac.completions[:ac.maxCompletions]
	}

	ac.selectedIndex = 0
	ac.active = len(ac.completions) > 0
}

// Hide deactivates autocomplete
func (ac *Autocomplete) Hide() {
	ac.active = false
	ac.completions = []Completion{}
	ac.selectedIndex = 0
	ac.currentInput = ""
	ac.currentProvider = ""
}

// IsActive returns whether autocomplete is currently showing
func (ac *Autocomplete) IsActive() bool {
	return ac.active
}

// Next moves selection to next completion
func (ac *Autocomplete) Next() {
	if !ac.active || len(ac.completions) == 0 {
		return
	}
	ac.selectedIndex = (ac.selectedIndex + 1) % len(ac.completions)
}

// Previous moves selection to previous completion
func (ac *Autocomplete) Previous() {
	if !ac.active || len(ac.completions) == 0 {
		return
	}
	ac.selectedIndex--
	if ac.selectedIndex < 0 {
		ac.selectedIndex = len(ac.completions) - 1
	}
}

// GetSelected returns the currently selected completion
func (ac *Autocomplete) GetSelected() *Completion {
	if !ac.active || len(ac.completions) == 0 || ac.selectedIndex >= len(ac.completions) {
		return nil
	}
	return &ac.completions[ac.selectedIndex]
}

// GetCompletions returns all current completions
func (ac *Autocomplete) GetCompletions() []Completion {
	return ac.completions
}

// GetSelectedIndex returns the current selection index
func (ac *Autocomplete) GetSelectedIndex() int {
	return ac.selectedIndex
}

// Update refreshes completions based on new input
func (ac *Autocomplete) Update(input string) {
	if !ac.active {
		return
	}

	provider, ok := ac.providers[ac.currentProvider]
	if !ok {
		return
	}

	ac.currentInput = input
	ac.completions = provider.GetCompletions(input)

	// Limit to max completions
	if len(ac.completions) > ac.maxCompletions {
		ac.completions = ac.completions[:ac.maxCompletions]
	}

	// Reset selection if we have fewer completions
	if ac.selectedIndex >= len(ac.completions) {
		ac.selectedIndex = 0
	}

	// Hide if no completions
	if len(ac.completions) == 0 {
		ac.Hide()
	}
}

// ToolProvider provides tool name completions
type ToolProvider struct {
	// tools is a list of available tool names
	tools []string
}

// NewToolProvider creates a new tool completion provider
func NewToolProvider(tools []string) *ToolProvider {
	if tools == nil {
		tools = []string{}
	}
	return &ToolProvider{tools: tools}
}

// SetTools updates the available tools
func (tp *ToolProvider) SetTools(tools []string) {
	tp.tools = tools
}

// GetCompletions returns fuzzy-matched tool completions
func (tp *ToolProvider) GetCompletions(input string) []Completion {
	if input == "" {
		// Return all tools if no input
		completions := make([]Completion, len(tp.tools))
		for i, tool := range tp.tools {
			completions[i] = Completion{
				Value:       tool,
				Display:     tool,
				Description: "tool",
				Score:       0,
			}
		}
		return completions
	}

	// Fuzzy match
	matches := fuzzy.Find(input, tp.tools)

	completions := make([]Completion, len(matches))
	for i, match := range matches {
		completions[i] = Completion{
			Value:       match.Str,
			Display:     match.Str,
			Description: "tool",
			Score:       match.Score,
		}
	}

	return completions
}

// FileProvider provides file path completions
type FileProvider struct {
	// basePath is the directory to search for files
	basePath string
}

// NewFileProvider creates a new file completion provider
func NewFileProvider() *FileProvider {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	return &FileProvider{basePath: wd}
}

// SetBasePath updates the base directory for file searching
func (fp *FileProvider) SetBasePath(path string) {
	fp.basePath = path
}

// GetCompletions returns file path completions
func (fp *FileProvider) GetCompletions(input string) []Completion {
	// Determine directory and prefix
	dir := fp.basePath
	prefix := input

	if strings.Contains(input, "/") {
		// Split into directory and file parts
		dir = filepath.Dir(input)
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(fp.basePath, dir)
		}
		prefix = filepath.Base(input)
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []Completion{}
	}

	// Build list of files
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files unless explicitly requested
		if !strings.HasPrefix(prefix, ".") && strings.HasPrefix(name, ".") {
			continue
		}
		files = append(files, name)
	}

	// Fuzzy match if we have a prefix
	if prefix != "" {
		matches := fuzzy.Find(prefix, files)
		completions := make([]Completion, 0, len(matches))
		for _, match := range matches {
			fullPath := filepath.Join(dir, match.Str)
			display := match.Str

			// Add trailing slash for directories
			info, err := os.Stat(fullPath)
			if err == nil && info.IsDir() {
				display += "/"
			}

			completions = append(completions, Completion{
				Value:       fullPath,
				Display:     display,
				Description: "file",
				Score:       match.Score,
			})
		}
		return completions
	}

	// No prefix - return all files
	completions := make([]Completion, 0, len(files))
	for _, file := range files {
		fullPath := filepath.Join(dir, file)
		display := file

		// Add trailing slash for directories
		info, err := os.Stat(fullPath)
		if err == nil && info.IsDir() {
			display += "/"
		}

		completions = append(completions, Completion{
			Value:       fullPath,
			Display:     display,
			Description: "file",
			Score:       0,
		})
	}

	// Sort alphabetically
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Display < completions[j].Display
	})

	return completions
}

// HistoryProvider provides command history completions
type HistoryProvider struct {
	// history stores recent commands
	history []string
}

// NewHistoryProvider creates a new history completion provider
func NewHistoryProvider() *HistoryProvider {
	return &HistoryProvider{
		history: []string{},
	}
}

// AddToHistory adds a command to history
func (hp *HistoryProvider) AddToHistory(command string) {
	// Avoid duplicates
	for i, cmd := range hp.history {
		if cmd == command {
			// Move to front
			hp.history = append([]string{command}, append(hp.history[:i], hp.history[i+1:]...)...)
			return
		}
	}

	// Add to front
	hp.history = append([]string{command}, hp.history...)

	// Limit history size
	if len(hp.history) > 100 {
		hp.history = hp.history[:100]
	}
}

// GetCompletions returns history-based completions
func (hp *HistoryProvider) GetCompletions(input string) []Completion {
	if input == "" {
		// Return recent history
		completions := make([]Completion, 0, len(hp.history))
		for i, cmd := range hp.history {
			if i >= 10 { // Limit to 10 recent
				break
			}
			completions = append(completions, Completion{
				Value:       cmd,
				Display:     cmd,
				Description: "history",
				Score:       len(hp.history) - i, // More recent = higher score
			})
		}
		return completions
	}

	// Fuzzy match history
	matches := fuzzy.Find(input, hp.history)

	completions := make([]Completion, len(matches))
	for i, match := range matches {
		completions[i] = Completion{
			Value:       match.Str,
			Display:     match.Str,
			Description: "history",
			Score:       match.Score,
		}
	}

	return completions
}

// DetectProvider determines which provider to use based on input context
func DetectProvider(input string) string {
	trimmed := strings.TrimSpace(input)

	// Tool completion if starting with :tool
	if strings.HasPrefix(trimmed, ":tool ") {
		return "tool"
	}

	// File completion if looks like a path
	if strings.Contains(trimmed, "/") || strings.HasPrefix(trimmed, ".") || strings.HasPrefix(trimmed, "~") {
		return "file"
	}

	// Default to history
	return "history"
}
