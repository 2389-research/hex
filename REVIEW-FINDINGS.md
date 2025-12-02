# Code Review Findings

**Review Date**: 2025-12-01
**Packages Reviewed**:
- `internal/hooks/`
- `internal/skills/`
- `internal/permissions/`
- `internal/commands/`
- Modified files: `internal/tools/executor.go`, `cmd/clem/root.go`, `internal/ui/view.go`

**Overall Assessment**: The code is well-structured with comprehensive test coverage. Found several important issues that should be addressed, plus some minor improvements.

---

## Critical Issues (must fix)

### 1. **Goroutine leak in async hook execution** - `internal/hooks/executor.go:125-130`

**Issue**: `ExecuteAsync` launches a goroutine that discards results but has no mechanism to cancel or track completion. This creates a goroutine leak if hooks run indefinitely.

```go
func (e *Executor) ExecuteAsync(hook *HookConfig, event *Event) {
    go func() {
        _ = e.Execute(hook, event)
        // Result is discarded for async execution
    }()
}
```

**Problem**:
- No way to cancel long-running async hooks
- No panic recovery - a panic in async hook will crash the program
- No logging of async hook failures

**Recommendation**: Add panic recovery and optional result logging:
```go
func (e *Executor) ExecuteAsync(hook *HookConfig, event *Event) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                // Log panic but don't crash
            }
        }()
        result := e.Execute(hook, event)
        if !result.Success && !hook.IgnoreFailure {
            // Optionally log async hook failures
        }
    }()
}
```

### 2. **Missing mutex lock in skills/registry.go** - `internal/skills/registry.go:42-50`

**Issue**: `Get` method uses RLock correctly, but `Register` method has a race condition window.

```go
func (r *Registry) Register(skill *Skill) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if skill.Name == "" {
        return fmt.Errorf("skill has no name")
    }

    // Allow overriding (project skills override user skills, etc.)
    r.skills[skill.Name] = skill  // RACE: map write without checking
    return nil
}
```

**Problem**: While the lock is held, concurrent reads via `Get()` during map write could theoretically cause issues in Go < 1.9. More importantly, there's no validation that the skill pointer is non-nil before dereferencing.

**Recommendation**: Add nil check:
```go
if skill == nil {
    return fmt.Errorf("cannot register nil skill")
}
```

### 3. **Unchecked regex compilation in pattern matching** - `internal/skills/skill.go:128-141`

**Issue**: `MatchesPattern` compiles regex on every call but silently skips invalid patterns.

```go
for _, pattern := range s.ActivationPatterns {
    // Compile regex pattern (case-insensitive)
    re, err := regexp.Compile("(?i)" + pattern)
    if err != nil {
        // Invalid regex pattern, skip
        continue  // SILENT FAILURE
    }

    if re.MatchString(lowerMsg) {
        return true
    }
}
```

**Problem**:
- No indication to user that their regex pattern is invalid
- Compiles regex on every message check (performance issue)
- Silent failures make debugging difficult

**Recommendation**:
1. Validate patterns at parse time in `ParseBytes()`
2. Cache compiled regexes in `Skill` struct
3. Return validation errors during skill loading

---

## Important Issues (should fix)

### 4. **Missing nil check in commands/tool.go** - `internal/commands/tool.go:95-100`

**Issue**: Type assertion on `params["args"]` doesn't check if the underlying value is nil.

```go
var args map[string]interface{}
if argsParam, ok := params["args"].(map[string]interface{}); ok {
    args = argsParam  // Could be nil
} else {
    args = make(map[string]interface{})
}
```

**Problem**: If someone passes `{"args": null}`, this sets `args = nil` which will panic when passed to `template.Execute`.

**Recommendation**:
```go
if argsParam, ok := params["args"].(map[string]interface{}); ok && argsParam != nil {
    args = argsParam
} else {
    args = make(map[string]interface{})
}
```

### 5. **File descriptor leak potential in loader.go files** - `internal/skills/loader.go:135-146`

**Issue**: `filepath.Walk` can be interrupted but file descriptors opened during walk may not be cleaned up.

```go
err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err  // Propagates error, could leave FDs open
    }
    // ...
})
```

**Problem**: While this is unlikely to cause issues with just `os.FileInfo`, it's good practice to ensure cleanup.

**Recommendation**: Consider using `filepath.WalkDir` (Go 1.16+) which is more efficient and has better error handling.

### 6. **Hook execution errors swallowed in engine.go** - `internal/hooks/engine.go:72-88`

**Issue**: Only the last error is returned, earlier hook failures are lost.

```go
var lastErr error
for _, hook := range hooks {
    // ...
    if hook.Async {
        e.executor.ExecuteAsync(&hook, event)
    } else {
        result := e.executor.Execute(&hook, event)
        if !result.Success && !hook.IgnoreFailure {
            lastErr = fmt.Errorf("hook failed: %w (stderr: %s)", result.Error, result.Stderr)
            // BUG: Continues to next hook, losing context about which hook failed
        }
    }
}
return lastErr
```

**Problem**: If 3 hooks fail, only the last failure is reported. No indication which hook(s) failed.

**Recommendation**: Collect all errors and return a multi-error:
```go
var errors []error
for _, hook := range hooks {
    // ...
    if !result.Success && !hook.IgnoreFailure {
        errors = append(errors, fmt.Errorf("hook %q failed: %w", hook.Command, result.Error))
    }
}
if len(errors) > 0 {
    return fmt.Errorf("hooks failed: %v", errors)
}
```

### 7. **Environment variable duplication in executor.go** - `internal/hooks/executor.go:90-122`

**Issue**: Environment variables can be duplicated when the same key is set multiple times.

```go
env := os.Environ()  // e.g., ["FOO=bar", ...]

for k, v := range baseVars {
    env = append(env, fmt.Sprintf("%s=%s", k, v))  // Adds "FOO=baz"
}
// Now env has both "FOO=bar" and "FOO=baz"
```

**Problem**: When exec runs the command, behavior with duplicate env vars is undefined. Last one usually wins, but it's not guaranteed.

**Recommendation**: Build env from scratch or deduplicate:
```go
// Option 1: Build from scratch
env := make([]string, 0)
// Add variables in priority order

// Option 2: Deduplicate
envMap := make(map[string]string)
for _, e := range os.Environ() {
    parts := strings.SplitN(e, "=", 2)
    if len(parts) == 2 {
        envMap[parts[0]] = parts[1]
    }
}
// Override with hook-specific values
for k, v := range baseVars {
    envMap[k] = v
}
// Convert back to slice
```

### 8. **Insufficient error context in permissions/checker.go** - `internal/permissions/checker.go:44`

**Issue**: The `params` parameter is accepted but never used, making debugging permission issues harder.

```go
func (c *Checker) Check(toolName string, _ map[string]interface{}) CheckResult {
    // params are ignored, but could be useful for context
}
```

**Problem**: Future permission rules might want to check parameter values (e.g., "allow Write but only to /tmp/*"). The infrastructure doesn't support this.

**Recommendation**: Keep the signature for future extensibility, but consider logging params in debug mode.

### 9. **splitFrontmatter doesn't handle \r\n consistently** - `internal/skills/skill.go:82-106`, `internal/commands/command.go:73-106`

**Issue**: Code checks for both `---\n` and `---\r\n` at start but only checks `---` for closing delimiter.

```go
if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
    return nil, data, nil
}

// Later...
for i := 1; i < len(lines); i++ {
    line := bytes.TrimSpace(lines[i])
    if bytes.Equal(line, []byte("---")) {  // Only checks "---", not "---\r"
        endIdx = i
        break
    }
}
```

**Problem**: On Windows with CRLF line endings, the closing `---` might not be detected correctly if `TrimSpace` doesn't handle it.

**Recommendation**: Use `bytes.TrimSpace` consistently (which it does), but also document expected line ending handling. Actually, this is probably fine since `bytes.Split(data, []byte("\n"))` handles both.

---

## Minor Issues (nice to fix)

### 10. **Code duplication in skills and commands packages**

**Issue**: `splitFrontmatter` function is duplicated verbatim in:
- `internal/skills/skill.go:82-116`
- `internal/commands/command.go:73-106`

**Recommendation**: Extract to a shared utility package like `internal/frontmatter`.

### 11. **findProjectSkillsDir and findProjectCommandsDir are duplicated**

**Issue**: Nearly identical functions in:
- `internal/skills/loader.go:38-60`
- `internal/commands/loader.go:37-59`

**Recommendation**: Extract to shared utility or accept a subdirectory parameter.

### 12. **Magic numbers without constants**

**Issue**: Multiple hardcoded values:
- `priority == 0` defaults to `5` (`internal/skills/skill.go:71-73`)
- Loop limit of `10` for directory search (`internal/skills/loader.go:46`)
- Default timeout `5000` (`internal/hooks/config.go:206-210`)

**Recommendation**: Define as named constants:
```go
const (
    DefaultSkillPriority = 5
    MaxDirSearchDepth = 10
    DefaultHookTimeoutMS = 5000
)
```

### 13. **Missing nil check for tools.Result**

**Issue**: In `internal/tools/executor.go:105-112`, result is checked for nil after success check:

```go
success := err == nil && result != nil && result.Success
errMsg := ""
if err != nil {
    errMsg = err.Error()
} else if result != nil && !result.Success {  // Checks nil again
    errMsg = result.Error
}
```

**Problem**: If `Execute()` returns `(nil, nil)`, this will panic on `result.Success`. While the interface contract says this shouldn't happen, defensive coding would help.

**Recommendation**: Check result for nil first.

### 14. **Exported fields without documentation**

**Issue**: Several exported structs have undocumented fields:
- `HookConfig.Match` (`internal/hooks/config.go:21`)
- `Skill.Tags` (`internal/skills/skill.go:21`)
- `Command.Args` (`internal/commands/command.go:22`)

**Recommendation**: Add godoc comments for exported fields.

### 15. **findSimilarSkills/findSimilarCommands use naive matching**

**Issue**: Bidirectional substring matching is very loose:

```go
if strings.Contains(lowerName, lowerQuery) || strings.Contains(lowerQuery, lowerName) {
    similar = append(similar, name)
}
```

**Problem**: Query "a" matches everything containing "a". Query "test-command" matches "test" and "command" separately.

**Recommendation**: Use Levenshtein distance or fuzzy matching library for better suggestions.

### 16. **No protection against frontmatter bombs**

**Issue**: `splitFrontmatter` searches through entire file looking for closing `---`:

```go
for i := 1; i < len(lines); i++ {  // Could be millions of lines
    line := bytes.TrimSpace(lines[i])
    if bytes.Equal(line, []byte("---")) {
        endIdx = i
        break
    }
}
```

**Problem**: A malicious or corrupted file with `---\n` at the start but no closing delimiter will cause the entire file to be scanned.

**Recommendation**: Limit frontmatter search to first N lines (e.g., 100):
```go
maxFrontmatterLines := 100
for i := 1; i < len(lines) && i < maxFrontmatterLines; i++ {
```

### 17. **Template execution has no timeout**

**Issue**: `Command.Expand()` executes user templates with no timeout:

```go
if err := tmpl.Execute(&buf, args); err != nil {
    return "", fmt.Errorf("execute command template: %w", err)
}
```

**Problem**: A malicious template with infinite loop could hang the program.

**Recommendation**: Add timeout context or limit template complexity.

---

## Code Quality Notes

### Strengths

1. **Excellent test coverage**: All packages have comprehensive test files with good edge case coverage
2. **Thread-safe design**: Both registries use proper mutex locking patterns
3. **Clean separation of concerns**: Loaders, registries, and tools are well-separated
4. **Good error messages**: Helpful suggestions when tools/skills/commands not found
5. **Configuration flexibility**: Multiple locations (builtin/user/project) with proper override semantics
6. **Backward compatibility**: Permission checker and hook engine are optional in executor

### Areas for Improvement

1. **Documentation**: Could use more package-level docs explaining overall architecture
2. **Error handling**: Some errors are logged to stderr but not returned
3. **Resource cleanup**: No explicit cleanup/shutdown methods for registries or executors
4. **Metrics/Observability**: No instrumentation for monitoring hook execution, tool usage, etc.
5. **Configuration validation**: Some invalid configs are silently ignored rather than reported

---

## Recommendations

### High Priority

1. **Fix goroutine leak** in `ExecuteAsync` (Issue #1)
2. **Add panic recovery** to async hooks (Issue #1)
3. **Cache compiled regexes** in skills (Issue #3)
4. **Add nil checks** in template expansion (Issue #4)
5. **Return multi-error** from hook engine (Issue #6)

### Medium Priority

6. **Deduplicate environment variables** (Issue #7)
7. **Extract common code** (Issues #10, #11)
8. **Add constants** for magic numbers (Issue #12)
9. **Limit frontmatter search** (Issue #16)

### Low Priority

10. **Improve fuzzy matching** (Issue #15)
11. **Add template timeout** (Issue #17)
12. **Add godoc comments** (Issue #14)

---

## Integration Concerns

### Tool Executor Integration

**File**: `internal/tools/executor.go`

**Observation**: The integration of hooks into the executor looks good. The hooks are optional and won't break existing code if nil.

**Potential issue**:
- Line 96-97: `extractFilePath` only checks `file_path` parameter, but different tools use different parameter names (`path`, `filepath`, etc.)
- Hooks won't fire for many tools that use different parameter naming

**Recommendation**: Make `extractFilePath` check multiple common parameter names:
```go
func extractFilePath(params map[string]interface{}) string {
    for _, key := range []string{"file_path", "path", "filepath", "notebook_path"} {
        if val, ok := params[key].(string); ok && val != "" {
            return val
        }
    }
    return ""
}
```

### Root Command Integration

**File**: `cmd/clem/root.go`

**Only saw first 100 lines** - couldn't review full integration. Need to check:
- How hooks engine is initialized
- How skills/commands registries are wired up
- Error handling for loading failures

### UI Integration

**File**: `internal/ui/view.go`

**Only saw first 100 lines** - rendering code looks fine for what was visible. No obvious issues.

---

## Test Coverage Analysis

All packages have good test coverage:
- `hooks_test.go`: 562 lines - comprehensive
- `permissions_test.go`: 448 lines - excellent edge case coverage
- `skill_test.go`, `loader_test.go`, `registry_test.go`, `tool_test.go` - good coverage
- `command_test.go`, `loader_test.go`, `registry_test.go`, `tool_test.go` - good coverage

**Missing tests**:
- No tests for async hook execution (#1)
- No tests for regex compilation caching (#3)
- No tests for nil args in template expansion (#4)
- No tests for environment variable duplication (#7)
- No load/stress tests for concurrent access

---

## Security Considerations

1. **Command injection**: Hooks execute arbitrary shell commands - properly flagged with `//nolint:gosec // G204`
2. **Path traversal**: File loading uses user-provided paths - properly flagged with `//nolint:gosec // G304`
3. **Template injection**: Command templates execute with user data - no XSS concerns since this is CLI
4. **Resource exhaustion**: No limits on:
   - Number of skills/commands/hooks loaded
   - Size of skill/command files
   - Number of concurrent async hooks
   - Template execution time

**Recommendation**: Add resource limits in production use.

---

## Performance Notes

1. **Regex compilation** in hot path (Issue #3) - should cache
2. **filepath.Walk** could use **filepath.WalkDir** for better performance
3. **Map access** in registries is O(1) - good
4. **Sorting** happens on every `All()` call - could cache if called frequently

---

## Conclusion

The code is **production-ready with fixes**. The critical issue (#1 goroutine leak) must be fixed. The important issues (#4-#9) should be addressed before release. The minor issues can be technical debt.

Overall code quality is high with good separation of concerns, comprehensive tests, and thoughtful error handling. The main gaps are in edge case handling and resource cleanup.
