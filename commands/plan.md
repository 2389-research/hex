---
name: plan
description: Create a detailed implementation plan with bite-sized tasks
args:
  feature: The feature or task to plan (optional)
---

# Implementation Plan

You are creating a detailed implementation plan for {{if .feature}}{{.feature}}{{else}}the requested work{{end}}.

## Planning Process

Follow these steps to create a comprehensive plan:

1. **Understand Requirements**
   - Analyze the task or feature request
   - Identify acceptance criteria
   - Note any constraints or dependencies

2. **Break Down Work**
   - Divide into logical, independent tasks
   - Each task should be completable in one sitting
   - Order tasks by dependencies

3. **Specify Details**
   - For each task, provide:
     - Exact file paths that need changes
     - Specific functions/modules to modify
     - Expected inputs and outputs
     - Verification steps

4. **Plan Testing**
   - Identify what tests need to be written
   - Specify test scenarios and edge cases
   - Include integration and end-to-end tests

## Output Format

Structure your plan as:

```markdown
## Implementation Plan: [Feature Name]

### Overview
Brief description of the feature and approach

### Tasks

#### Task 1: [Task Name]
**Files**: `path/to/file.go`, `path/to/test.go`
**Changes**:
- Add function X to handle Y
- Update interface Z to include new method

**Tests**:
- Test case A: verify X with input Y
- Test case B: verify edge case Z

**Verification**: Run `go test ./...` and verify output contains...

#### Task 2: [Task Name]
...

### Dependencies
- Task 2 depends on Task 1
- Task 3 can run in parallel with Task 2

### Success Criteria
- [ ] All tests pass
- [ ] Feature works as specified
- [ ] Documentation updated
- [ ] No regressions introduced
```

## Important Guidelines

- **Be Specific**: Use exact file paths, not "the auth file"
- **Be Bite-Sized**: Tasks should be 15-30 minutes each
- **Be Testable**: Every task should have clear verification steps
- **Assume Zero Context**: Write for someone who doesn't know the codebase
- **Think TDD**: Write tests before implementation when possible

Create the plan now.
