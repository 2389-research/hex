// ABOUTME: Regression tests for built-in help command content.
// ABOUTME: Keeps slash help aligned with implemented TUI key bindings.
package main

import (
	"os"
	"strings"
	"testing"
)

func TestHelpCommandDocumentsImplementedKeyBindings(t *testing.T) {
	content, err := os.ReadFile("../../commands/help.md")
	if err != nil {
		t.Fatalf("read help command: %v", err)
	}

	help := string(content)
	expected := []string{
		"`Tab` | Switch Chat/History/Tools/Intro views",
		"`Ctrl+H` | Toggle help panel",
		"`Ctrl+R` | Toggle conversation history",
		"`Ctrl+T` | Toggle typewriter mode",
		"`Ctrl+F` | Toggle favorite",
	}
	for _, text := range expected {
		if !strings.Contains(help, text) {
			t.Fatalf("help command missing implemented binding %q", text)
		}
	}

	unimplemented := []string{
		"`?` | Toggle help panel",
		"`Tab` | Show autocomplete / Accept suggestion",
	}
	for _, text := range unimplemented {
		if strings.Contains(help, text) {
			t.Fatalf("help command documents stale binding %q", text)
		}
	}
}
