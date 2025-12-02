---
name: review
description: Perform comprehensive code review on changes
args:
  file: Specific file to review (optional)
  scope: Review scope - "changes", "file", or "commit" (optional)
---

# Code Review

You are performing a thorough code review {{if .file}}of `{{.file}}`{{else}}of the current changes{{end}}.

## Review Process

### 1. Understand the Changes
- Run `git status` to see modified files
- Run `git diff` to see actual changes (or `git diff HEAD~1` for last commit)
{{if .file}}- Focus on `{{.file}}`{{end}}
- Understand what problem the changes solve

### 2. Code Quality Review

Check for:

**Correctness**
- [ ] Does the code do what it's supposed to do?
- [ ] Are edge cases handled?
- [ ] Are error conditions properly managed?
- [ ] Is there proper input validation?

**Readability**
- [ ] Are names clear and descriptive?
- [ ] Is the code self-documenting?
- [ ] Are complex sections commented?
- [ ] Is formatting consistent?

**Maintainability**
- [ ] Is the code DRY (Don't Repeat Yourself)?
- [ ] Are functions/methods single-purpose?
- [ ] Is complexity reasonable?
- [ ] Would another developer understand this?

**Testing**
- [ ] Are there tests for new functionality?
- [ ] Do tests cover edge cases?
- [ ] Are tests clear and maintainable?
- [ ] Do all tests pass?

**Security**
- [ ] Is user input sanitized?
- [ ] Are there SQL injection risks?
- [ ] Are credentials/secrets hardcoded?
- [ ] Is authentication/authorization correct?

**Performance**
- [ ] Are there obvious performance issues?
- [ ] Is database usage efficient?
- [ ] Are expensive operations cached if needed?
- [ ] Are there resource leaks?

### 3. Architecture Review

Consider:
- [ ] Does this fit with existing patterns?
- [ ] Are responsibilities properly separated?
- [ ] Are dependencies reasonable?
- [ ] Does this introduce technical debt?

### 4. Breaking Changes

Check:
- [ ] Is the API backward compatible?
- [ ] Are database migrations needed?
- [ ] Will this affect other systems?
- [ ] Is documentation updated?

## Output Format

Provide feedback as:

```markdown
## Code Review Summary

### Overview
[Brief summary of changes and their purpose]

### Strengths
- [What's done well]
- [Good patterns used]

### Issues Found

#### Critical Issues
- **[Location]**: [Problem description]
  - **Impact**: [What could go wrong]
  - **Fix**: [How to resolve]

#### Suggestions
- **[Location]**: [Improvement opportunity]
  - **Why**: [Rationale]
  - **How**: [Specific suggestion]

### Testing Gaps
- [What tests are missing]
- [Edge cases not covered]

### Recommendation
☐ Approve as-is
☐ Approve with minor suggestions
☐ Request changes before merge

### Next Steps
[What needs to be done]
```

## Guidelines

- **Be Specific**: Reference exact line numbers or code sections
- **Be Constructive**: Explain why and how to improve
- **Be Thorough**: Don't skip the checklist items
- **Be Honest**: Better to catch issues now than in production
- **Be Helpful**: Provide code examples for suggestions

Perform the review now.
