// ABOUTME: Tests for mux print-mode runner helpers.
// ABOUTME: Verifies plan-mode prompts keep generated plans visible to execution.
package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hexTools "github.com/2389-research/hex/internal/tools"
)

func TestBuildMuxPlanExecutionPromptIncludesGeneratedPlan(t *testing.T) {
	plan := "1. Read files\n2. Make changes\n3. Run tests"

	prompt := buildMuxPlanExecutionPrompt(plan)

	if !strings.Contains(prompt, plan) {
		t.Fatalf("execution prompt does not include plan:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Now execute it step by step") {
		t.Fatalf("execution prompt does not ask agent to execute step by step:\n%s", prompt)
	}
}

func TestRunPrintModeWithMuxPassesMaxTurnsToAgentConfig(t *testing.T) {
	source, err := os.ReadFile("mux_runner.go")
	if err != nil {
		t.Fatalf("read mux_runner.go: %v", err)
	}

	if !strings.Contains(string(source), "MaxIterations: effectiveMaxTurns()") {
		t.Fatal("mux print mode should pass effectiveMaxTurns() to adapter.Config.MaxIterations")
	}
}

func TestGetHexToolsWithMuxSubagentsIncludesSkillTool(t *testing.T) {
	tools, err := getHexToolsWithMuxSubagents(nil, nil)
	if err != nil {
		t.Fatalf("getHexToolsWithMuxSubagents() error = %v", err)
	}

	for _, tool := range tools {
		if tool.Name() == "Skill" {
			return
		}
	}

	t.Fatalf("mux tool list did not include Skill; got %v", toolNames(tools))
}

func TestGetHexToolsWithMuxSubagentsLoadsPluginSkills(t *testing.T) {
	pluginDir := t.TempDir()
	skillContent := []byte(`---
name: plugin-skill
description: Skill from plugin path
---

# Plugin Skill

Loaded through mux tool construction.
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin-skill.md"), skillContent, 0o644); err != nil {
		t.Fatalf("write plugin skill: %v", err)
	}

	tools, err := getHexToolsWithMuxSubagents(nil, []string{pluginDir})
	if err != nil {
		t.Fatalf("getHexToolsWithMuxSubagents() error = %v", err)
	}

	for _, tool := range tools {
		if tool.Name() != "Skill" {
			continue
		}
		result, execErr := tool.Execute(context.Background(), map[string]interface{}{"command": "plugin-skill"})
		if execErr != nil {
			t.Fatalf("Skill Execute() error = %v", execErr)
		}
		if result == nil || !result.Success || !strings.Contains(result.Output, "Loaded through mux tool construction.") {
			t.Fatalf("Skill tool did not load plugin skill; result = %#v", result)
		}
		return
	}

	t.Fatalf("mux tool list did not include Skill; got %v", toolNames(tools))
}

func TestGetHexToolsWithMuxSubagentsFiltersToolAliases(t *testing.T) {
	originalEnabledTools := enabledTools
	enabledTools = []string{"Read", "Grep", "Glob", "Skill"}
	defer func() {
		enabledTools = originalEnabledTools
	}()

	tools, err := getHexToolsWithMuxSubagents(nil, nil)
	if err != nil {
		t.Fatalf("getHexToolsWithMuxSubagents() error = %v", err)
	}

	names := toolNames(tools)
	for _, name := range []string{"read_file", "grep", "glob", "Skill"} {
		if !containsToolName(names, name) {
			t.Fatalf("filtered mux tools missing %q; got %v", name, names)
		}
	}
}

func toolNames(tools []hexTools.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name())
	}
	return names
}

func containsToolName(names []string, target string) bool {
	for _, name := range names {
		if name == target {
			return true
		}
	}
	return false
}
