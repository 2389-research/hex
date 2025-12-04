// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Benchmarks for Phase 1 Task 3 content caching implementation
// ABOUTME: Measures performance improvement from markdown and help text caching
package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/harper/pagent/internal/ui/components"
	"github.com/harper/pagent/internal/ui/themes"
)

// BenchmarkMarkdownRenderNoCaching benchmarks markdown rendering without cache
func BenchmarkMarkdownRenderNoCaching(b *testing.B) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Create a complex markdown message
	markdown := generateComplexMarkdown()
	model.AddMessage("assistant", markdown)
	msg := model.Messages[0]

	// Force cache invalidation on each iteration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.InvalidateMarkdownCache()
		_, err := model.RenderMessage(msg)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkMarkdownRenderWithCaching benchmarks markdown rendering with cache
func BenchmarkMarkdownRenderWithCaching(b *testing.B) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Create a complex markdown message
	markdown := generateComplexMarkdown()
	model.AddMessage("assistant", markdown)
	msg := model.Messages[0]

	// First render to populate cache
	_, err := model.RenderMessage(msg)
	if err != nil {
		b.Fatalf("Initial render failed: %v", err)
	}

	// Benchmark cached renders
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.RenderMessage(msg)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkMarkdownRenderMultipleMessages benchmarks rendering multiple messages
func BenchmarkMarkdownRenderMultipleMessages(b *testing.B) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Add 10 messages with markdown
	for i := 0; i < 10; i++ {
		model.AddMessage("assistant", generateComplexMarkdown())
	}

	// First pass to populate cache
	for _, msg := range model.Messages {
		_, err := model.RenderMessage(msg)
		if err != nil {
			b.Fatalf("Initial render failed: %v", err)
		}
	}

	// Benchmark re-rendering all messages (should hit cache)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, msg := range model.Messages {
			_, err := model.RenderMessage(msg)
			if err != nil {
				b.Fatalf("Render failed: %v", err)
			}
		}
	}
}

// BenchmarkHelpOverlayNoCaching benchmarks help overlay without cache
func BenchmarkHelpOverlayNoCaching(b *testing.B) {
	theme := themes.GetTheme("dracula")
	help := components.NewHelpOverlay(theme)
	help.SetSize(80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Force cache invalidation by changing size
		help.SetSize(80+i%2, 24)
		_ = help.View()
	}
}

// BenchmarkHelpOverlayWithCaching benchmarks help overlay with cache
func BenchmarkHelpOverlayWithCaching(b *testing.B) {
	theme := themes.GetTheme("dracula")
	help := components.NewHelpOverlay(theme)
	help.SetSize(80, 24)

	// First render to populate cache
	_ = help.View()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = help.View()
	}
}

// BenchmarkCacheLookup benchmarks the cache lookup overhead
func BenchmarkCacheLookup(b *testing.B) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Populate cache with many entries
	for i := 0; i < 100; i++ {
		model.AddMessage("assistant", fmt.Sprintf("# Message %d", i))
		msg := model.Messages[i]
		_, err := model.RenderMessage(msg)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}

	// Benchmark looking up a message in the middle
	targetMsg := model.Messages[50]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.RenderMessage(targetMsg)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkCacheInvalidation benchmarks cache invalidation operations
func BenchmarkCacheInvalidation(b *testing.B) {
	model := NewModel("test-conv", "test-model", "dracula")

	// Populate cache with many entries
	for i := 0; i < 100; i++ {
		model.AddMessage("assistant", fmt.Sprintf("# Message %d", i))
		msg := model.Messages[i]
		_, err := model.RenderMessage(msg)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.InvalidateMarkdownCache()
	}
}

// BenchmarkCacheClear benchmarks cache clearing operations
func BenchmarkCacheClear(b *testing.B) {
	model := NewModel("test-conv", "test-model", "dracula")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Populate cache
		for j := 0; j < 100; j++ {
			model.AddMessage("assistant", fmt.Sprintf("# Message %d", j))
			msg := model.Messages[j]
			_, _ = model.RenderMessage(msg)
		}
		b.StartTimer()

		// Benchmark clear operation
		model.ClearMarkdownCache()

		b.StopTimer()
		model.Messages = nil // Reset for next iteration
		b.StartTimer()
	}
}

// BenchmarkMessageIDGeneration benchmarks message ID generation overhead
func BenchmarkMessageIDGeneration(b *testing.B) {
	model := NewModel("test-conv", "test-model", "dracula")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.AddMessage("assistant", "Test message")
	}
}

// generateComplexMarkdown creates a realistic markdown document for benchmarking
func generateComplexMarkdown() string {
	var b strings.Builder

	// Title
	b.WriteString("# Complex Markdown Document\n\n")

	// Introduction with formatting
	b.WriteString("This is a **bold** statement with *italic* text and `inline code`.\n\n")

	// Code block
	b.WriteString("```go\n")
	b.WriteString("func main() {\n")
	b.WriteString("    fmt.Println(\"Hello, World!\")\n")
	b.WriteString("    for i := 0; i < 10; i++ {\n")
	b.WriteString("        fmt.Printf(\"Count: %d\\n\", i)\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	b.WriteString("```\n\n")

	// Lists
	b.WriteString("## Features\n\n")
	b.WriteString("- First feature with **emphasis**\n")
	b.WriteString("- Second feature with [link](https://example.com)\n")
	b.WriteString("- Third feature with `code`\n\n")

	// Ordered list
	b.WriteString("## Steps\n\n")
	b.WriteString("1. Initialize the system\n")
	b.WriteString("2. Configure settings\n")
	b.WriteString("3. Run the application\n\n")

	// Quote
	b.WriteString("> This is a blockquote with important information.\n")
	b.WriteString("> It can span multiple lines.\n\n")

	// Table (if supported)
	b.WriteString("| Column 1 | Column 2 | Column 3 |\n")
	b.WriteString("|----------|----------|----------|\n")
	b.WriteString("| Value 1  | Value 2  | Value 3  |\n")
	b.WriteString("| Value 4  | Value 5  | Value 6  |\n\n")

	// Nested sections
	b.WriteString("## Conclusion\n\n")
	b.WriteString("This document demonstrates various **markdown features** including:\n\n")
	b.WriteString("- Headers at multiple levels\n")
	b.WriteString("- *Text formatting* options\n")
	b.WriteString("- `Code blocks` and inline code\n")
	b.WriteString("- Lists (ordered and unordered)\n")
	b.WriteString("- Blockquotes\n")
	b.WriteString("- Tables\n")

	return b.String()
}
