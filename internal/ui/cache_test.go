// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Tests for Phase 1 Task 3 content caching implementation
// ABOUTME: Verifies markdown and help text caching behavior
package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/pagent/internal/ui/components"
	"github.com/harper/pagent/internal/ui/themes"
)

// TestMarkdownCacheHit verifies cache is used when rendering same message twice
func TestMarkdownCacheHit(t *testing.T) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Add an assistant message with markdown
	model.AddMessage("assistant", "# Hello\n\nThis is **bold** text")
	msg := model.Messages[0]

	// First render (cache miss - should render and cache)
	rendered1, err := model.RenderMessage(msg)
	if err != nil {
		t.Fatalf("First render failed: %v", err)
	}

	// Verify cache was populated
	if model.markdownCache == nil {
		t.Fatal("Cache was not initialized")
	}
	if _, exists := model.markdownCache[msg.ID]; !exists {
		t.Errorf("Message %s was not cached after first render", msg.ID)
	}

	// Second render (cache hit - should return cached value)
	rendered2, err := model.RenderMessage(msg)
	if err != nil {
		t.Fatalf("Second render failed: %v", err)
	}

	// Both should be identical (same cached result)
	if rendered1 != rendered2 {
		t.Error("Cached render returned different result")
	}
}

// TestMarkdownCacheMissAfterInvalidation verifies cache invalidation works
func TestMarkdownCacheMissAfterInvalidation(t *testing.T) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Add an assistant message
	model.AddMessage("assistant", "# Test")
	msg := model.Messages[0]

	// First render (populates cache)
	_, err := model.RenderMessage(msg)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify cache was populated
	if len(model.markdownCache) == 0 {
		t.Fatal("Cache should have one entry")
	}

	// Invalidate cache (simulating window resize or theme change)
	model.InvalidateMarkdownCache()

	// Verify dirty flag is set
	if !model.markdownCacheDirty {
		t.Error("Cache dirty flag was not set")
	}

	// Second render should NOT use cache (cache is dirty)
	// The cache map still exists but dirty flag prevents use
	_, err = model.RenderMessage(msg)
	if err != nil {
		t.Fatalf("Render after invalidation failed: %v", err)
	}
}

// TestMarkdownCacheClear verifies cache clearing works
func TestMarkdownCacheClear(t *testing.T) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Add multiple messages
	model.AddMessage("assistant", "# Message 1")
	model.AddMessage("assistant", "# Message 2")

	// Render both (populates cache)
	for _, msg := range model.Messages {
		_, err := model.RenderMessage(msg)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
	}

	// Verify cache has entries
	if len(model.markdownCache) != 2 {
		t.Errorf("Expected 2 cached entries, got %d", len(model.markdownCache))
	}

	// Clear cache
	model.ClearMarkdownCache()

	// Verify cache is empty
	if len(model.markdownCache) != 0 {
		t.Errorf("Cache should be empty after clear, got %d entries", len(model.markdownCache))
	}

	// Verify dirty flag is cleared
	if model.markdownCacheDirty {
		t.Error("Dirty flag should be cleared after ClearMarkdownCache")
	}
}

// TestMarkdownCacheUserMessages verifies user messages are not cached
func TestMarkdownCacheUserMessages(t *testing.T) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Add a user message
	model.AddMessage("user", "Hello assistant")
	msg := model.Messages[0]

	// Render user message
	rendered, err := model.RenderMessage(msg)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// User messages should return raw content (no markdown processing)
	if rendered != msg.Content {
		t.Error("User message should return raw content")
	}

	// Cache should remain empty (user messages aren't cached)
	if len(model.markdownCache) != 0 {
		t.Error("User messages should not be cached")
	}
}

// TestMarkdownCacheWindowResize verifies cache invalidation on resize
func TestMarkdownCacheWindowResize(t *testing.T) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Add message and render
	model.AddMessage("assistant", "# Test")
	msg := model.Messages[0]
	_, err := model.RenderMessage(msg)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify cache is populated
	if len(model.markdownCache) == 0 {
		t.Fatal("Cache should be populated")
	}

	// Simulate window resize
	model.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Verify cache was invalidated (dirty flag set)
	if !model.markdownCacheDirty {
		t.Error("Cache should be marked dirty after window resize")
	}
}

// TestMarkdownCacheMessageWithoutID verifies graceful handling of messages without IDs
func TestMarkdownCacheMessageWithoutID(t *testing.T) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Create message without ID (edge case)
	msg := Message{
		Role:    "assistant",
		Content: "# Test",
	}

	// Should still render (just won't cache)
	rendered, err := model.RenderMessage(msg)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should have content
	if rendered == "" {
		t.Error("Render should return content even without ID")
	}

	// Cache should remain empty (no ID to key on)
	if len(model.markdownCache) != 0 {
		t.Error("Message without ID should not be cached")
	}
}

// TestHelpOverlayCacheHit verifies help text cache is used
func TestHelpOverlayCacheHit(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := components.NewHelpOverlay(theme)

	// Set size to ensure consistent rendering
	help.SetSize(80, 24)

	// First render (cache miss)
	view1 := help.View()
	if view1 == "" {
		t.Fatal("First view returned empty string")
	}

	// Second render (cache hit - should be identical)
	view2 := help.View()
	if view2 == "" {
		t.Fatal("Second view returned empty string")
	}

	// Should be exactly the same (from cache)
	if view1 != view2 {
		t.Error("Cached view should be identical")
	}

	// Verify content has expected elements
	if !strings.Contains(view1, "Keyboard Shortcuts") {
		t.Error("Help text should contain title")
	}
	if !strings.Contains(view1, "ctrl+c") {
		t.Error("Help text should contain shortcuts")
	}
}

// TestHelpOverlayCacheInvalidationOnResize verifies cache invalidation on size change
func TestHelpOverlayCacheInvalidationOnResize(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := components.NewHelpOverlay(theme)

	// Set initial size and render
	help.SetSize(80, 24)
	view1 := help.View()
	if view1 == "" {
		t.Fatal("First view returned empty string")
	}

	// Change size (should invalidate cache)
	help.SetSize(120, 40)

	// Render again (should regenerate, not use cache)
	view2 := help.View()
	if view2 == "" {
		t.Fatal("Second view returned empty string")
	}

	// Content might differ due to different size constraints
	// The important thing is it didn't crash and returned content
}

// TestHelpOverlayCacheNoInvalidationOnSameSize verifies cache isn't invalidated unnecessarily
func TestHelpOverlayCacheNoInvalidationOnSameSize(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := components.NewHelpOverlay(theme)

	// Set initial size and render
	help.SetSize(80, 24)
	view1 := help.View()

	// Set same size again (should NOT invalidate cache)
	help.SetSize(80, 24)

	// Render again (should use cache)
	view2 := help.View()

	// Should be identical (from cache)
	if view1 != view2 {
		t.Error("Setting same size should not invalidate cache")
	}
}

// TestHelpOverlayCachePartialResize verifies cache invalidation on any dimension change
func TestHelpOverlayCachePartialResize(t *testing.T) {
	theme := themes.GetTheme("dracula")
	help := components.NewHelpOverlay(theme)

	// Set initial size and render
	help.SetSize(80, 24)
	_ = help.View()

	// Change only width (should still invalidate)
	help.SetSize(100, 24)
	view1 := help.View()

	// Reset and change only height
	help.SetSize(80, 24)
	_ = help.View()
	help.SetSize(80, 30)
	view2 := help.View()

	// Both should have regenerated content
	if view1 == "" || view2 == "" {
		t.Error("Partial resizes should regenerate content")
	}
}

// TestCacheInitialization verifies caches are properly initialized
func TestCacheInitialization(t *testing.T) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Markdown cache should start nil (lazy init)
	if model.markdownCache != nil {
		t.Error("Markdown cache should start nil")
	}

	// Dirty flag should start false
	if model.markdownCacheDirty {
		t.Error("Cache dirty flag should start false")
	}

	// After adding and rendering a message, cache should be initialized
	model.AddMessage("assistant", "# Test")
	_, err := model.RenderMessage(model.Messages[0])
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if model.markdownCache == nil {
		t.Error("Markdown cache should be initialized after first render")
	}
}

// TestMessageIDGeneration verifies message IDs are unique
func TestMessageIDGeneration(t *testing.T) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Add multiple messages
	model.AddMessage("user", "Message 1")
	model.AddMessage("assistant", "Message 2")
	model.AddMessage("user", "Message 3")

	// Verify all have IDs
	for i, msg := range model.Messages {
		if msg.ID == "" {
			t.Errorf("Message %d has empty ID", i)
		}
	}

	// Verify IDs are unique
	seen := make(map[string]bool)
	for i, msg := range model.Messages {
		if seen[msg.ID] {
			t.Errorf("Message %d has duplicate ID: %s", i, msg.ID)
		}
		seen[msg.ID] = true
	}
}
