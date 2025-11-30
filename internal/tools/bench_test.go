// ABOUTME: Performance benchmarks for tool execution system
// ABOUTME: Measures Read, Write, Edit, Grep, Glob, and registry overhead
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkReadToolSmallFile measures Read tool with small files
func BenchmarkReadToolSmallFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create a small file (1KB)
	content := make([]byte, 1024)
	for i := range content {
		content[i] = byte('a' + (i % 26))
	}
	if err := os.WriteFile(testFile, content, 0600); err != nil {
		b.Fatal(err)
	}

	tool := NewReadTool()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := map[string]interface{}{
			"path": testFile,
		}
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkReadToolLargeFile measures Read tool with large files
func BenchmarkReadToolLargeFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Create a 1MB file
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte('a' + (i % 26))
	}
	if err := os.WriteFile(testFile, content, 0600); err != nil {
		b.Fatal(err)
	}

	tool := NewReadTool()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := map[string]interface{}{
			"path": testFile,
		}
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWriteTool measures Write tool performance
func BenchmarkWriteTool(b *testing.B) {
	tmpDir := b.TempDir()

	tool := NewWriteTool()
	ctx := context.Background()

	content := "This is a test file content that we're writing repeatedly."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test_%d.txt", i))
		params := map[string]interface{}{
			"path":    testFile,
			"content": content,
		}
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEditTool measures Edit tool string replacement
func BenchmarkEditTool(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "edit_test.txt")

	// Create file with content to edit
	initialContent := "This is line 1\nThis is line 2\nThis is line 3\nThis is line 4\n"
	if err := os.WriteFile(testFile, []byte(initialContent), 0600); err != nil {
		b.Fatal(err)
	}

	tool := NewEditTool()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Reset file content
		_ = os.WriteFile(testFile, []byte(initialContent), 0600)
		b.StartTimer()

		params := map[string]interface{}{
			"path":       testFile,
			"old_string": "line 2",
			"new_string": "modified line 2",
		}
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGrepTool measures Grep tool search performance
func BenchmarkGrepTool(b *testing.B) {
	tmpDir := b.TempDir()

	// Create multiple files with content
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf("This is file %d\nIt contains some searchable content\nAnd some other lines\n", i)
		testFile := filepath.Join(tmpDir, fmt.Sprintf("file_%d.txt", i))
		if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
			b.Fatal(err)
		}
	}

	tool := NewGrepTool()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := map[string]interface{}{
			"pattern": "searchable",
			"path":    tmpDir,
		}
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGlobTool measures Glob tool pattern matching
func BenchmarkGlobTool(b *testing.B) {
	tmpDir := b.TempDir()

	// Create directory structure with various files
	for i := 0; i < 50; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("file_%d.txt", i))
		if err := os.WriteFile(testFile, []byte("content"), 0600); err != nil {
			b.Fatal(err)
		}
	}
	for i := 0; i < 50; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("script_%d.sh", i))
		if err := os.WriteFile(testFile, []byte("#!/bin/bash"), 0600); err != nil {
			b.Fatal(err)
		}
	}

	tool := NewGlobTool()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := map[string]interface{}{
			"pattern": "*.txt",
			"path":    tmpDir,
		}
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBashToolSimple measures Bash tool with simple command
func BenchmarkBashToolSimple(b *testing.B) {
	tool := NewBashTool()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := map[string]interface{}{
			"command":     "echo 'Hello World'",
			"description": "Print hello world",
		}
		_, err := tool.Execute(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkToolRegistryLookup measures tool lookup performance
func BenchmarkToolRegistryLookup(b *testing.B) {
	registry := NewRegistry()

	// Register some tools
	_ = registry.Register(NewReadTool())
	_ = registry.Register(NewWriteTool())
	_ = registry.Register(NewEditTool())
	_ = registry.Register(NewGrepTool())
	_ = registry.Register(NewGlobTool())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := registry.Get("read_file")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkToolRegistryListAll measures tool listing performance
func BenchmarkToolRegistryListAll(b *testing.B) {
	registry := NewRegistry()

	// Register 10 tools
	_ = registry.Register(NewReadTool())
	_ = registry.Register(NewWriteTool())
	_ = registry.Register(NewEditTool())
	_ = registry.Register(NewGrepTool())
	_ = registry.Register(NewGlobTool())
	_ = registry.Register(NewBashTool())
	_ = registry.Register(NewWebFetchTool())
	_ = registry.Register(NewWebSearchTool())
	_ = registry.Register(NewAskUserQuestionTool())
	_ = registry.Register(NewTaskTool())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.List()
	}
}

// BenchmarkExecutorApprovalOverhead measures approval system overhead
func BenchmarkExecutorApprovalOverhead(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0600); err != nil {
		b.Fatal(err)
	}

	registry := NewRegistry()
	_ = registry.Register(NewReadTool())

	// Auto-approve callback
	approver := func(_ string, _ map[string]interface{}) bool {
		return true
	}

	executor := NewExecutor(registry, approver)

	ctx := context.Background()
	params := map[string]interface{}{
		"path": testFile,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, "read_file", params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentToolExecution measures parallel tool execution
func BenchmarkConcurrentToolExecution(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := "This is test content\n"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		b.Fatal(err)
	}

	tool := NewReadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path": testFile,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tool.Execute(ctx, params)
			if err != nil {
				b.Error(err)
			}
		}
	})
}
