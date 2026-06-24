// ABOUTME: Tests model-level autocomplete trigger behavior.
// ABOUTME: Verifies non-slash providers are surfaced from normal input.
package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelAutocompleteTriggersForFilePaths(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("alpha"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.CurrentView = ViewModeChat
	model.Ready = true
	model.Width = 80
	model.Height = 24
	provider, ok := model.autocomplete.GetProvider("file")
	if !ok {
		t.Fatal("file provider missing")
	}
	provider.(*FileProvider).SetBasePath(dir)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	model = updated.(*Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(*Model)

	if !model.autocomplete.IsActive() {
		t.Fatal("file autocomplete should be active for ./ input")
	}
	if model.autocomplete.currentProvider != "file" {
		t.Fatalf("provider = %q, want file", model.autocomplete.currentProvider)
	}
}

func TestModelAutocompleteTriggersForInputHistory(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.CurrentView = ViewModeChat
	model.Ready = true
	model.Width = 80
	model.Height = 24
	model.addToInputHistory("explain alpha deployment")

	for _, r := range "alpha" {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		model = updated.(*Model)
	}

	if !model.autocomplete.IsActive() {
		t.Fatal("history autocomplete should be active for matching input history")
	}
	if model.autocomplete.currentProvider != "history" {
		t.Fatalf("provider = %q, want history", model.autocomplete.currentProvider)
	}
}
