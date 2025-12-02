# Claude Code Alignment - 6-Phase Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform Pagen from a productivity agent into a full-featured Claude Code-aligned agent framework with hooks, skills, slash commands, permissions, subagents, and plugins.

**Architecture:** Modular implementation of six core systems following Claude Code's architecture: lifecycle hooks for automation, skills for domain knowledge, enhanced permissions for tool control, slash commands for workflows, subagent framework for parallel execution, and plugin system for extensibility. Each phase builds on previous ones with clear interfaces and comprehensive testing.

**Tech Stack:** Go 1.24+, Viper (config), SQLite (storage), existing tool system, plugin architecture with process isolation, markdown-based skills/commands, shell hook execution.

---

## Current State Analysis

**What We Have:**
- ✅ Core tool system (11 built-in tools)
- ✅ MCP integration
- ✅ Interactive TUI with Bubbletea
- ✅ SQLite storage for conversations
- ✅ Configuration system (Viper)
- ✅ Existing `internal/hooks/` directory (empty)
- ✅ Existing `internal/plugins/` directory (empty)
- ✅ Provider architecture (Gmail provider just implemented)
- ✅ Comprehensive Claude Code documentation in `docs/claude-docs/`

**What We Need:**
- ❌ Hooks system (10 lifecycle events)
- ❌ Skills system (pattern-based activation)
- ❌ Enhanced permissions (auto/ask/deny modes)
- ❌ Slash commands system
- ❌ Subagent framework
- ❌ Plugin system architecture

---

## Phase 1: Hooks System

### Task 1.1: Hook Configuration Schema

**Files:**
- Modify: `internal/core/config.go:16-21` (add Hooks field)
- Create: `internal/hooks/types.go`
- Create: `internal/hooks/types_test.go`

**Step 1: Write the failing test**

```go
// internal/hooks/types_test.go
package hooks

import "testing"

func TestHookConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  HookConfig
        wantErr bool
    }{
        {
            name: "valid hook with command",
            config: HookConfig{
                Command: "echo 'test'",
            },
            wantErr: false,
        },
        {
            name: "empty command",
            config: HookConfig{
                Command: "",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/hooks/... -v`
Expected: FAIL with "HookConfig undefined"

**Step 3: Write minimal implementation**

```go
// internal/hooks/types.go
// ABOUTME: Type definitions for the hooks system
// ABOUTME: Defines hook events, configuration, and execution context

package hooks

import "fmt"

// HookEvent represents a lifecycle event that can trigger hooks
type HookEvent string

const (
    SessionStart      HookEvent = "SessionStart"
    SessionEnd        HookEvent = "SessionEnd"
    UserPromptSubmit  HookEvent = "UserPromptSubmit"
    ModelResponseDone HookEvent = "ModelResponseDone"
    PreToolUse        HookEvent = "PreToolUse"
    PostToolUse       HookEvent = "PostToolUse"
    PreCommit         HookEvent = "PreCommit"
    PostCommit        HookEvent = "PostCommit"
    OnError           HookEvent = "OnError"
    PlanModeEnter     HookEvent = "PlanModeEnter"
)

// HookConfig defines a single hook's configuration
type HookConfig struct {
    Command string            `yaml:"command" json:"command"`
    Matcher map[string]string `yaml:"matcher,omitempty" json:"matcher,omitempty"`
    Timeout int               `yaml:"timeout,omitempty" json:"timeout,omitempty"` // seconds
}

// Validate checks if the hook configuration is valid
func (h *HookConfig) Validate() error {
    if h.Command == "" {
        return fmt.Errorf("hook command cannot be empty")
    }
    if h.Timeout < 0 {
        return fmt.Errorf("hook timeout cannot be negative")
    }
    return nil
}

// HooksConfig maps hook events to their configurations
type HooksConfig map[HookEvent][]HookConfig

// EventData contains context passed to hooks
type EventData map[string]interface{}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/hooks/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/hooks/types.go internal/hooks/types_test.go
git commit -m "feat(hooks): add hook configuration types and validation"
```

---

### Task 1.2: Hook Executor

**Files:**
- Create: `internal/hooks/executor.go`
- Create: `internal/hooks/executor_test.go`

**Step 1: Write the failing test**

```go
// internal/hooks/executor_test.go
package hooks

import (
    "context"
    "testing"
    "time"
)

func TestExecutor_Run(t *testing.T) {
    exec := NewExecutor()

    ctx := context.Background()
    event := SessionStart
    data := EventData{"test": "value"}

    config := HookConfig{
        Command: "echo 'test'",
        Timeout: 5,
    }

    result, err := exec.Run(ctx, event, config, data)
    if err != nil {
        t.Fatalf("Run() error = %v", err)
    }

    if !result.Success {
        t.Errorf("expected success, got failure: %s", result.Error)
    }
}

func TestExecutor_Run_Timeout(t *testing.T) {
    exec := NewExecutor()

    ctx := context.Background()
    config := HookConfig{
        Command: "sleep 10",
        Timeout: 1,
    }

    _, err := exec.Run(ctx, SessionStart, config, EventData{})
    if err == nil {
        t.Error("expected timeout error, got nil")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/hooks/... -v`
Expected: FAIL with "NewExecutor undefined"

**Step 3: Write minimal implementation**

```go
// internal/hooks/executor.go
// ABOUTME: Hook execution engine for running shell commands
// ABOUTME: Handles timeouts, environment setup, and result collection

package hooks

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "time"
)

// Executor runs hook commands
type Executor struct{}

// NewExecutor creates a new hook executor
func NewExecutor() *Executor {
    return &Executor{}
}

// HookResult contains the result of hook execution
type HookResult struct {
    Success bool
    Output  string
    Error   string
    Duration time.Duration
}

// Run executes a hook command with timeout
func (e *Executor) Run(ctx context.Context, event HookEvent, config HookConfig, data EventData) (*HookResult, error) {
    timeout := time.Duration(config.Timeout) * time.Second
    if timeout == 0 {
        timeout = 30 * time.Second // Default 30s
    }

    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    // Prepare environment with event data
    jsonData, err := json.Marshal(data)
    if err != nil {
        return nil, fmt.Errorf("marshal event data: %w", err)
    }

    start := time.Now()

    cmd := exec.CommandContext(ctx, "sh", "-c", config.Command)
    cmd.Env = append(cmd.Env,
        fmt.Sprintf("CLAUDE_HOOK_EVENT=%s", event),
        fmt.Sprintf("CLAUDE_HOOK_DATA=%s", jsonData),
    )

    output, err := cmd.CombinedOutput()
    duration := time.Since(start)

    result := &HookResult{
        Success:  err == nil,
        Output:   string(output),
        Duration: duration,
    }

    if err != nil {
        result.Error = err.Error()
        return result, err
    }

    return result, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/hooks/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/hooks/executor.go internal/hooks/executor_test.go
git commit -m "feat(hooks): implement hook executor with timeout support"
```

---

### Task 1.3: Hook Manager

**Files:**
- Create: `internal/hooks/manager.go`
- Create: `internal/hooks/manager_test.go`
- Modify: `internal/core/config.go:16-21` (add Hooks field)

**Step 1: Write the failing test**

```go
// internal/hooks/manager_test.go
package hooks

import (
    "context"
    "testing"
)

func TestManager_Trigger(t *testing.T) {
    config := HooksConfig{
        SessionStart: []HookConfig{
            {Command: "echo 'session started'", Timeout: 5},
        },
    }

    mgr := NewManager(config)

    results, err := mgr.Trigger(context.Background(), SessionStart, EventData{
        "timestamp": "2025-12-02T10:00:00Z",
    })

    if err != nil {
        t.Fatalf("Trigger() error = %v", err)
    }

    if len(results) != 1 {
        t.Errorf("expected 1 result, got %d", len(results))
    }

    if !results[0].Success {
        t.Errorf("expected success, got failure: %s", results[0].Error)
    }
}

func TestManager_Trigger_NoHooks(t *testing.T) {
    mgr := NewManager(HooksConfig{})

    results, err := mgr.Trigger(context.Background(), SessionStart, EventData{})

    if err != nil {
        t.Fatalf("Trigger() error = %v", err)
    }

    if len(results) != 0 {
        t.Errorf("expected 0 results, got %d", len(results))
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/hooks/... -v`
Expected: FAIL with "NewManager undefined"

**Step 3: Write minimal implementation**

```go
// internal/hooks/manager.go
// ABOUTME: Hook manager for lifecycle event automation
// ABOUTME: Coordinates hook execution and result collection

package hooks

import (
    "context"
    "fmt"
)

// Manager manages hook execution for lifecycle events
type Manager struct {
    config   HooksConfig
    executor *Executor
}

// NewManager creates a new hook manager
func NewManager(config HooksConfig) *Manager {
    return &Manager{
        config:   config,
        executor: NewExecutor(),
    }
}

// Trigger executes all hooks for a given event
func (m *Manager) Trigger(ctx context.Context, event HookEvent, data EventData) ([]*HookResult, error) {
    hooks, ok := m.config[event]
    if !ok || len(hooks) == 0 {
        return nil, nil
    }

    var results []*HookResult
    for _, hookConfig := range hooks {
        result, err := m.executor.Run(ctx, event, hookConfig, data)
        if err != nil {
            // Log error but continue with other hooks
            result = &HookResult{
                Success: false,
                Error:   err.Error(),
            }
        }
        results = append(results, result)
    }

    return results, nil
}

// TriggerAsync executes hooks without blocking
func (m *Manager) TriggerAsync(event HookEvent, data EventData) {
    go func() {
        _, _ = m.Trigger(context.Background(), event, data)
    }()
}
```

**Step 4: Add Hooks field to Config**

```go
// Modify internal/core/config.go
import "github.com/harper/clem/internal/hooks"

type Config struct {
    APIKey         string              `mapstructure:"api_key"`
    Model          string              `mapstructure:"model"`
    DefaultTools   []string            `mapstructure:"default_tools"`
    PermissionMode string              `mapstructure:"permission_mode"`
    Hooks          hooks.HooksConfig   `mapstructure:"hooks"`
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/hooks/... -v && go test ./internal/core/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/hooks/manager.go internal/hooks/manager_test.go internal/core/config.go
git commit -m "feat(hooks): add hook manager and integrate with config"
```

---

### Task 1.4: Integrate Hooks into Main Flow

**Files:**
- Modify: `cmd/clem/main.go` (add SessionStart/SessionEnd hooks)
- Create: `internal/hooks/integration_test.go`

**Step 1: Write integration test**

```go
// internal/hooks/integration_test.go
package hooks

import (
    "context"
    "os"
    "path/filepath"
    "testing"
)

func TestIntegration_SessionLifecycle(t *testing.T) {
    tmpDir := t.TempDir()
    logFile := filepath.Join(tmpDir, "session.log")

    config := HooksConfig{
        SessionStart: []HookConfig{
            {Command: fmt.Sprintf("echo 'started' > %s", logFile), Timeout: 5},
        },
        SessionEnd: []HookConfig{
            {Command: fmt.Sprintf("echo 'ended' >> %s", logFile), Timeout: 5},
        },
    }

    mgr := NewManager(config)

    // Trigger session start
    _, err := mgr.Trigger(context.Background(), SessionStart, EventData{})
    if err != nil {
        t.Fatalf("SessionStart failed: %v", err)
    }

    // Trigger session end
    _, err = mgr.Trigger(context.Background(), SessionEnd, EventData{})
    if err != nil {
        t.Fatalf("SessionEnd failed: %v", err)
    }

    // Verify log file
    content, err := os.ReadFile(logFile)
    if err != nil {
        t.Fatalf("failed to read log: %v", err)
    }

    expected := "started\nended\n"
    if string(content) != expected {
        t.Errorf("expected %q, got %q", expected, string(content))
    }
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./internal/hooks/... -v`
Expected: PASS

**Step 3: Integrate hooks in main.go**

```go
// Modify cmd/clem/main.go - add to main() function
func main() {
    // ... existing config loading ...

    // Initialize hook manager
    hookManager := hooks.NewManager(cfg.Hooks)

    // Trigger SessionStart
    hookManager.TriggerAsync(hooks.SessionStart, hooks.EventData{
        "timestamp": time.Now().Format(time.RFC3339),
        "model":     cfg.Model,
    })

    // Ensure SessionEnd runs on exit
    defer func() {
        hookManager.TriggerAsync(hooks.SessionEnd, hooks.EventData{
            "timestamp": time.Now().Format(time.RFC3339),
        })
        time.Sleep(100 * time.Millisecond) // Give hooks time to complete
    }()

    // ... rest of main function ...
}
```

**Step 4: Test manually**

Create test config: `~/.clem/config.yaml`
```yaml
hooks:
  SessionStart:
    - command: "echo 'Session started' > /tmp/clem-session.log"
  SessionEnd:
    - command: "echo 'Session ended' >> /tmp/clem-session.log"
```

Run: `go run cmd/clem/main.go --print "test"`
Verify: `cat /tmp/clem-session.log` shows both messages

**Step 5: Commit**

```bash
git add cmd/clem/main.go internal/hooks/integration_test.go
git commit -m "feat(hooks): integrate SessionStart and SessionEnd into main flow"
```

---

## Phase 2: Skills System

### Task 2.1: Skill Schema and Discovery

**Files:**
- Create: `internal/skills/types.go`
- Create: `internal/skills/types_test.go`
- Create: `internal/skills/loader.go`
- Create: `internal/skills/loader_test.go`

**Step 1: Write the failing test**

```go
// internal/skills/types_test.go
package skills

import "testing"

func TestSkillMetadata_Validate(t *testing.T) {
    tests := []struct {
        name    string
        meta    SkillMetadata
        wantErr bool
    }{
        {
            name: "valid skill",
            meta: SkillMetadata{
                Name:        "test-skill",
                Description: "A test skill",
            },
            wantErr: false,
        },
        {
            name: "missing name",
            meta: SkillMetadata{
                Description: "A test skill",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.meta.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/skills/... -v`
Expected: FAIL with "SkillMetadata undefined"

**Step 3: Write minimal implementation**

```go
// internal/skills/types.go
// ABOUTME: Type definitions for the skills system
// ABOUTME: Defines skill metadata, content structure, and activation patterns

package skills

import (
    "fmt"
    "regexp"
)

// SkillMetadata is the frontmatter of a skill file
type SkillMetadata struct {
    Name               string   `yaml:"name"`
    Description        string   `yaml:"description"`
    Tags               []string `yaml:"tags,omitempty"`
    ActivationPatterns []string `yaml:"activationPatterns,omitempty"`
    Model              string   `yaml:"model,omitempty"`
}

// Validate checks if skill metadata is valid
func (m *SkillMetadata) Validate() error {
    if m.Name == "" {
        return fmt.Errorf("skill name is required")
    }
    if m.Description == "" {
        return fmt.Errorf("skill description is required")
    }

    // Validate activation patterns are valid regexes
    for _, pattern := range m.ActivationPatterns {
        if _, err := regexp.Compile(pattern); err != nil {
            return fmt.Errorf("invalid activation pattern %q: %w", pattern, err)
        }
    }

    return nil
}

// Skill represents a loaded skill
type Skill struct {
    Metadata    SkillMetadata
    Content     string
    FilePath    string
    activationRegexes []*regexp.Regexp
}

// Matches checks if the skill should be activated for the given input
func (s *Skill) Matches(input string) bool {
    if len(s.activationRegexes) == 0 {
        return false
    }

    for _, re := range s.activationRegexes {
        if re.MatchString(input) {
            return true
        }
    }

    return false
}

// CompilePatterns compiles activation patterns into regexes
func (s *Skill) CompilePatterns() error {
    s.activationRegexes = make([]*regexp.Regexp, 0, len(s.Metadata.ActivationPatterns))

    for _, pattern := range s.Metadata.ActivationPatterns {
        re, err := regexp.Compile(pattern)
        if err != nil {
            return fmt.Errorf("compile pattern %q: %w", pattern, err)
        }
        s.activationRegexes = append(s.activationRegexes, re)
    }

    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/skills/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/skills/types.go internal/skills/types_test.go
git commit -m "feat(skills): add skill metadata types and validation"
```

---

### Task 2.2: Skill Loader

**Files:**
- Create: `internal/skills/loader.go`
- Create: `internal/skills/loader_test.go`
- Create: `testdata/skills/test-skill.md` (test fixture)

**Step 1: Write the failing test**

```go
// internal/skills/loader_test.go
package skills

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoader_LoadSkill(t *testing.T) {
    // Create test skill file
    tmpDir := t.TempDir()
    skillPath := filepath.Join(tmpDir, "test-skill.md")

    content := `---
name: test-skill
description: A test skill
tags:
  - testing
activationPatterns:
  - "test.*"
---

# Test Skill

This is a test skill.
`

    if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
        t.Fatalf("failed to write test skill: %v", err)
    }

    loader := NewLoader()
    skill, err := loader.LoadSkill(skillPath)

    if err != nil {
        t.Fatalf("LoadSkill() error = %v", err)
    }

    if skill.Metadata.Name != "test-skill" {
        t.Errorf("expected name 'test-skill', got %q", skill.Metadata.Name)
    }

    if !skill.Matches("test this") {
        t.Error("expected skill to match 'test this'")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/skills/... -v`
Expected: FAIL with "NewLoader undefined"

**Step 3: Write minimal implementation**

```go
// internal/skills/loader.go
// ABOUTME: Skill loader for discovering and loading skill files
// ABOUTME: Parses YAML frontmatter and markdown content

package skills

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "gopkg.in/yaml.v3"
)

// Loader loads skills from the filesystem
type Loader struct {
    searchPaths []string
}

// NewLoader creates a new skill loader
func NewLoader() *Loader {
    return &Loader{
        searchPaths: []string{},
    }
}

// AddSearchPath adds a directory to search for skills
func (l *Loader) AddSearchPath(path string) {
    l.searchPaths = append(l.searchPaths, path)
}

// LoadSkill loads a single skill from a file
func (l *Loader) LoadSkill(path string) (*Skill, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read skill file: %w", err)
    }

    // Parse frontmatter
    parts := strings.SplitN(string(content), "---", 3)
    if len(parts) < 3 {
        return nil, fmt.Errorf("invalid skill format: missing frontmatter")
    }

    var meta SkillMetadata
    if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
        return nil, fmt.Errorf("parse frontmatter: %w", err)
    }

    if err := meta.Validate(); err != nil {
        return nil, fmt.Errorf("invalid metadata: %w", err)
    }

    skill := &Skill{
        Metadata: meta,
        Content:  strings.TrimSpace(parts[2]),
        FilePath: path,
    }

    if err := skill.CompilePatterns(); err != nil {
        return nil, fmt.Errorf("compile patterns: %w", err)
    }

    return skill, nil
}

// LoadAll discovers and loads all skills from search paths
func (l *Loader) LoadAll() ([]*Skill, error) {
    var skills []*Skill

    for _, searchPath := range l.searchPaths {
        err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }

            if !info.IsDir() && filepath.Ext(path) == ".md" {
                skill, err := l.LoadSkill(path)
                if err != nil {
                    // Log warning but continue
                    return nil
                }
                skills = append(skills, skill)
            }

            return nil
        })

        if err != nil {
            return nil, fmt.Errorf("walk %s: %w", searchPath, err)
        }
    }

    return skills, nil
}
```

**Step 4: Add gopkg.in/yaml.v3 dependency**

Run: `go get gopkg.in/yaml.v3 && go mod tidy`

**Step 5: Run test to verify it passes**

Run: `go test ./internal/skills/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/skills/loader.go internal/skills/loader_test.go go.mod go.sum
git commit -m "feat(skills): implement skill loader with frontmatter parsing"
```

---

## Phase 3: Enhanced Permissions

### Task 3.1: Permission Types and Rules

**Files:**
- Create: `internal/permissions/types.go`
- Create: `internal/permissions/types_test.go`

**Step 1: Write the failing test**

```go
// internal/permissions/types_test.go
package permissions

import "testing"

func TestPermissionMode_String(t *testing.T) {
    tests := []struct {
        mode PermissionMode
        want string
    }{
        {ModeAuto, "auto"},
        {ModeAsk, "ask"},
        {ModeDeny, "deny"},
    }

    for _, tt := range tests {
        if got := tt.mode.String(); got != tt.want {
            t.Errorf("mode.String() = %v, want %v", got, tt.want)
        }
    }
}

func TestToolPermissions_Check(t *testing.T) {
    perms := &ToolPermissions{
        Mode:       ModeAsk,
        AllowList:  []string{"Read", "Write"},
        DenyList:   []string{"Bash"},
    }

    tests := []struct {
        tool     string
        expected PermissionDecision
    }{
        {"Read", PermissionAllow},
        {"Write", PermissionAllow},
        {"Bash", PermissionDeny},
        {"Edit", PermissionAsk},
    }

    for _, tt := range tests {
        if got := perms.Check(tt.tool); got != tt.expected {
            t.Errorf("Check(%q) = %v, want %v", tt.tool, got, tt.expected)
        }
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/permissions/... -v`
Expected: FAIL with "PermissionMode undefined"

**Step 3: Write minimal implementation**

```go
// internal/permissions/types.go
// ABOUTME: Permission system types and decision logic
// ABOUTME: Three-mode permission system: auto, ask, deny

package permissions

import "strings"

// PermissionMode defines how tools are approved
type PermissionMode int

const (
    ModeAuto PermissionMode = iota // All tools auto-approved
    ModeAsk                         // Ask user for each tool
    ModeDeny                        // All tools denied
)

func (m PermissionMode) String() string {
    switch m {
    case ModeAuto:
        return "auto"
    case ModeAsk:
        return "ask"
    case ModeDeny:
        return "deny"
    default:
        return "unknown"
    }
}

// ParseMode converts string to PermissionMode
func ParseMode(s string) PermissionMode {
    switch strings.ToLower(s) {
    case "auto":
        return ModeAuto
    case "deny":
        return ModeDeny
    default:
        return ModeAsk
    }
}

// PermissionDecision represents a permission check result
type PermissionDecision int

const (
    PermissionDeny PermissionDecision = iota
    PermissionAsk
    PermissionAllow
)

// ToolPermissions defines permission rules for tools
type ToolPermissions struct {
    Mode       PermissionMode
    AllowList  []string // Tools always allowed
    DenyList   []string // Tools always denied
    ToolRules  map[string]PermissionMode // Per-tool overrides
}

// Check determines if a tool should be allowed, asked, or denied
func (p *ToolPermissions) Check(toolName string) PermissionDecision {
    // Check deny list first
    for _, denied := range p.DenyList {
        if denied == toolName {
            return PermissionDeny
        }
    }

    // Check allow list
    for _, allowed := range p.AllowList {
        if allowed == toolName {
            return PermissionAllow
        }
    }

    // Check per-tool rules
    if mode, ok := p.ToolRules[toolName]; ok {
        switch mode {
        case ModeAuto:
            return PermissionAllow
        case ModeDeny:
            return PermissionDeny
        default:
            return PermissionAsk
        }
    }

    // Fall back to global mode
    switch p.Mode {
    case ModeAuto:
        return PermissionAllow
    case ModeDeny:
        return PermissionDeny
    default:
        return PermissionAsk
    }
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/permissions/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/permissions/types.go internal/permissions/types_test.go
git commit -m "feat(permissions): add three-mode permission system"
```

---

## Phase 4: Slash Commands

### Task 4.1: Command Discovery and Loading

**Files:**
- Create: `internal/commands/types.go`
- Create: `internal/commands/loader.go`
- Create: `internal/commands/loader_test.go`

**Step 1: Write the failing test**

```go
// internal/commands/loader_test.go
package commands

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoader_LoadCommand(t *testing.T) {
    tmpDir := t.TempDir()
    cmdPath := filepath.Join(tmpDir, "test.md")

    content := "This is a test command.\n\nIt has multiple lines."

    if err := os.WriteFile(cmdPath, []byte(content), 0644); err != nil {
        t.Fatalf("failed to write test command: %v", err)
    }

    loader := NewLoader()
    loader.AddSearchPath(tmpDir)

    cmd, err := loader.LoadCommand("test")
    if err != nil {
        t.Fatalf("LoadCommand() error = %v", err)
    }

    if cmd.Name != "test" {
        t.Errorf("expected name 'test', got %q", cmd.Name)
    }

    if cmd.Content != content {
        t.Errorf("content mismatch")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/commands/... -v`
Expected: FAIL with "NewLoader undefined"

**Step 3: Write minimal implementation**

```go
// internal/commands/types.go
// ABOUTME: Slash command type definitions
// ABOUTME: Represents commands that expand to prompts

package commands

// Command represents a slash command
type Command struct {
    Name     string
    Content  string
    FilePath string
}

// internal/commands/loader.go
// ABOUTME: Command loader for discovering slash commands
// ABOUTME: Searches configured directories for .md files

package commands

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// Loader loads slash commands from the filesystem
type Loader struct {
    searchPaths []string
}

// NewLoader creates a new command loader
func NewLoader() *Loader {
    return &Loader{
        searchPaths: []string{},
    }
}

// AddSearchPath adds a directory to search for commands
func (l *Loader) AddSearchPath(path string) {
    l.searchPaths = append(l.searchPaths, path)
}

// LoadCommand loads a command by name
func (l *Loader) LoadCommand(name string) (*Command, error) {
    for _, searchPath := range l.searchPaths {
        cmdPath := filepath.Join(searchPath, name+".md")

        if _, err := os.Stat(cmdPath); err == nil {
            content, err := os.ReadFile(cmdPath)
            if err != nil {
                return nil, fmt.Errorf("read command: %w", err)
            }

            return &Command{
                Name:     name,
                Content:  string(content),
                FilePath: cmdPath,
            }, nil
        }
    }

    return nil, fmt.Errorf("command %q not found", name)
}

// LoadAll loads all available commands
func (l *Loader) LoadAll() ([]*Command, error) {
    var commands []*Command
    seen := make(map[string]bool)

    for _, searchPath := range l.searchPaths {
        entries, err := os.ReadDir(searchPath)
        if err != nil {
            continue // Skip inaccessible directories
        }

        for _, entry := range entries {
            if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
                continue
            }

            name := strings.TrimSuffix(entry.Name(), ".md")
            if seen[name] {
                continue // Already loaded from earlier search path
            }

            cmd, err := l.LoadCommand(name)
            if err != nil {
                continue
            }

            commands = append(commands, cmd)
            seen[name] = true
        }
    }

    return commands, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/commands/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/commands/types.go internal/commands/loader.go internal/commands/loader_test.go
git commit -m "feat(commands): implement slash command loader"
```

---

## Phase 5: Subagent Framework

### Task 5.1: Subagent Types and Context Isolation

**Files:**
- Create: `internal/subagents/types.go`
- Create: `internal/subagents/types_test.go`

**Step 1: Write the failing test**

```go
// internal/subagents/types_test.go
package subagents

import "testing"

func TestSubagentType_Validate(t *testing.T) {
    tests := []struct {
        name    string
        saType  SubagentType
        wantErr bool
    }{
        {"valid general", TypeGeneral, false},
        {"valid explore", TypeExplore, false},
        {"valid plan", TypePlan, false},
        {"valid code review", TypeCodeReview, false},
        {"invalid", SubagentType("invalid"), true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.saType.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/subagents/... -v`
Expected: FAIL with "SubagentType undefined"

**Step 3: Write minimal implementation**

```go
// internal/subagents/types.go
// ABOUTME: Subagent type definitions and context isolation
// ABOUTME: Four subagent types with specialized capabilities

package subagents

import "fmt"

// SubagentType defines the type of subagent
type SubagentType string

const (
    TypeGeneral    SubagentType = "general-purpose"
    TypeExplore    SubagentType = "Explore"
    TypePlan       SubagentType = "Plan"
    TypeCodeReview SubagentType = "code-reviewer"
)

// Validate checks if the subagent type is valid
func (t SubagentType) Validate() error {
    switch t {
    case TypeGeneral, TypeExplore, TypePlan, TypeCodeReview:
        return nil
    default:
        return fmt.Errorf("invalid subagent type: %s", t)
    }
}

// SubagentConfig defines configuration for a subagent
type SubagentConfig struct {
    Type        SubagentType
    Description string
    Prompt      string
    Model       string // Optional model override
}

// SubagentResult contains the result of subagent execution
type SubagentResult struct {
    Success bool
    Output  string
    Error   string
}

// Context represents isolated execution context for a subagent
type Context struct {
    WorkingDir string
    Tools      []string
    Config     map[string]interface{}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/subagents/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/subagents/types.go internal/subagents/types_test.go
git commit -m "feat(subagents): add subagent types and context isolation"
```

---

## Phase 6: Plugin System

### Task 6.1: Plugin Manifest and Discovery

**Files:**
- Create: `internal/plugins/types.go`
- Create: `internal/plugins/manifest.go`
- Create: `internal/plugins/manifest_test.go`

**Step 1: Write the failing test**

```go
// internal/plugins/manifest_test.go
package plugins

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoadManifest(t *testing.T) {
    tmpDir := t.TempDir()
    manifestPath := filepath.Join(tmpDir, "plugin.yaml")

    content := `name: test-plugin
version: 1.0.0
description: A test plugin
author: Test Author
entry_point: ./plugin.sh
hooks:
  - SessionStart
tools:
  - name: test_tool
    description: A test tool
`

    if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
        t.Fatalf("failed to write manifest: %v", err)
    }

    manifest, err := LoadManifest(manifestPath)
    if err != nil {
        t.Fatalf("LoadManifest() error = %v", err)
    }

    if manifest.Name != "test-plugin" {
        t.Errorf("expected name 'test-plugin', got %q", manifest.Name)
    }

    if len(manifest.Hooks) != 1 {
        t.Errorf("expected 1 hook, got %d", len(manifest.Hooks))
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/plugins/... -v`
Expected: FAIL with "LoadManifest undefined"

**Step 3: Write minimal implementation**

```go
// internal/plugins/types.go
// ABOUTME: Plugin system type definitions
// ABOUTME: Defines plugin manifest structure and tool registration

package plugins

// Manifest defines a plugin's metadata and capabilities
type Manifest struct {
    Name        string       `yaml:"name"`
    Version     string       `yaml:"version"`
    Description string       `yaml:"description"`
    Author      string       `yaml:"author"`
    EntryPoint  string       `yaml:"entry_point"`
    Hooks       []string     `yaml:"hooks,omitempty"`
    Tools       []ToolManifest `yaml:"tools,omitempty"`
    Skills      []string     `yaml:"skills,omitempty"`
    Commands    []string     `yaml:"commands,omitempty"`
}

// ToolManifest defines a tool provided by a plugin
type ToolManifest struct {
    Name        string `yaml:"name"`
    Description string `yaml:"description"`
}

// internal/plugins/manifest.go
// ABOUTME: Plugin manifest loader
// ABOUTME: Parses and validates plugin.yaml files

package plugins

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

// LoadManifest loads a plugin manifest from a file
func LoadManifest(path string) (*Manifest, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read manifest: %w", err)
    }

    var manifest Manifest
    if err := yaml.Unmarshal(content, &manifest); err != nil {
        return nil, fmt.Errorf("parse manifest: %w", err)
    }

    // Validate required fields
    if manifest.Name == "" {
        return nil, fmt.Errorf("plugin name is required")
    }
    if manifest.Version == "" {
        return nil, fmt.Errorf("plugin version is required")
    }
    if manifest.EntryPoint == "" {
        return nil, fmt.Errorf("plugin entry_point is required")
    }

    return &manifest, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/plugins/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/plugins/types.go internal/plugins/manifest.go internal/plugins/manifest_test.go
git commit -m "feat(plugins): add plugin manifest loading and validation"
```

---

## Documentation and Integration

### Task 7.1: Update Configuration Documentation

**Files:**
- Create: `docs/HOOKS.md`
- Create: `docs/SKILLS.md`
- Create: `docs/PERMISSIONS.md`
- Create: `docs/SLASH_COMMANDS.md`
- Create: `docs/SUBAGENTS.md`
- Create: `docs/PLUGINS.md`
- Modify: `README.md` (add references to new systems)

**Step 1: Create documentation**

Each doc should follow this structure:
```markdown
# [System Name]

## Overview
[What it is and why it exists]

## Configuration
[How to configure via config.yaml]

## Usage
[How to use the system]

## Examples
[Concrete examples]

## API Reference
[For developers]
```

**Step 2: Update README.md**

Add section:
```markdown
## Advanced Features

### Hooks System
Automate workflows with lifecycle hooks. See [HOOKS.md](docs/HOOKS.md).

### Skills System
Extend Claude with domain knowledge modules. See [SKILLS.md](docs/SKILLS.md).

### Enhanced Permissions
Fine-grained control over tool execution. See [PERMISSIONS.md](docs/PERMISSIONS.md).

### Slash Commands
Reusable prompt templates and workflows. See [SLASH_COMMANDS.md](docs/SLASH_COMMANDS.md).

### Subagents
Parallel task execution with context isolation. See [SUBAGENTS.md](docs/SUBAGENTS.md).

### Plugin System
Distribute and install community extensions. See [PLUGINS.md](docs/PLUGINS.md).
```

**Step 3: Commit**

```bash
git add docs/*.md README.md
git commit -m "docs: add comprehensive documentation for all six systems"
```

---

## Testing Strategy

Each phase includes:
1. **Unit tests**: Test individual components
2. **Integration tests**: Test system interactions
3. **End-to-end tests**: Test full workflows

**Coverage targets:**
- Phase 1 (Hooks): 80%
- Phase 2 (Skills): 80%
- Phase 3 (Permissions): 85%
- Phase 4 (Commands): 75%
- Phase 5 (Subagents): 70%
- Phase 6 (Plugins): 75%

---

## Plan complete and saved to `docs/plans/2025-12-02-claude-code-alignment.md`.

**Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration with quality gates

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach would you prefer, Doctor Biz?**
