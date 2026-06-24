// ABOUTME: Regression tests for CLI context-strategy wiring.
// ABOUTME: Guards interactive entry points so they share one context-manager factory.
package main

import (
	"os"
	"strings"
	"testing"
)

func TestContinueInteractiveUsesContextStrategyFactory(t *testing.T) {
	source, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatalf("read interactive.go: %v", err)
	}

	body := string(source)
	if !strings.Contains(body, "createContextManager()") {
		t.Fatal("continueInteractiveWithModel should use createContextManager")
	}
	if strings.Contains(body, "ctxmgr.NewManager(maxContextTokens)") {
		t.Fatal("continueInteractiveWithModel should not bypass contextStrategy with ctxmgr.NewManager")
	}
}
