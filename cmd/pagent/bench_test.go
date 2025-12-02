// ABOUTME: Performance benchmarks for CLI startup and initialization
// ABOUTME: Measures command execution time and startup overhead
package main

import (
	"os"
	"os/exec"
	"testing"
)

// BenchmarkStartupHelp measures time to display help
func BenchmarkStartupHelp(b *testing.B) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "clem_bench", ".")
	if err := buildCmd.Run(); err != nil {
		b.Skipf("failed to build: %v", err)
		return
	}
	defer func() { _ = os.Remove("clem_bench") }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("./clem_bench", "--help")
		if err := cmd.Run(); err != nil {
			b.Fatalf("failed to run: %v", err)
		}
	}
}

// BenchmarkStartupVersion measures time to display version
func BenchmarkStartupVersion(b *testing.B) {
	buildCmd := exec.Command("go", "build", "-o", "clem_bench", ".")
	if err := buildCmd.Run(); err != nil {
		b.Skipf("failed to build: %v", err)
		return
	}
	defer func() { _ = os.Remove("clem_bench") }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("./clem_bench", "version")
		if err := cmd.Run(); err != nil {
			b.Fatalf("failed to run: %v", err)
		}
	}
}
