# Built-in Slash Commands

This directory contains Clem's built-in slash commands - reusable prompt templates for common workflows.

## Available Commands

### `/plan` - Implementation Planning
Create detailed, bite-sized implementation plans with exact file paths and verification steps.

**Usage**: `/plan` or `/plan {"feature": "authentication"}`

**Use when**: Starting a new feature or breaking down complex work into manageable tasks.

---

### `/brainstorm` - Design Exploration
Interactive design refinement using the Socratic method. Explores alternatives, trade-offs, and validates approaches before implementation.

**Usage**: `/brainstorm` or `/brainstorm {"topic": "database migration strategy"}`

**Use when**: Exploring design decisions, evaluating approaches, or refining rough ideas.

---

### `/review` - Code Review
Comprehensive code review covering quality, security, testing, performance, and architecture.

**Usage**: `/review` or `/review {"file": "auth.go"}`

**Use when**: Before merging code, after implementing features, or reviewing changes.

---

### `/debug` - Systematic Debugging
Four-phase debugging framework: root cause investigation, pattern analysis, hypothesis testing, and implementation.

**Usage**: `/debug` or `/debug {"issue": "authentication fails on refresh"}`

**Use when**: Investigating bugs, unexpected behavior, or test failures.

---

### `/test` - Test Writing (TDD)
Write comprehensive tests following Test-Driven Development principles with Red-Green-Refactor cycle.

**Usage**: `/test` or `/test {"target": "authentication", "type": "unit"}`

**Use when**: Writing tests, following TDD workflow, or ensuring test coverage.

---

### `/commit` - Commit Changes
Review changes and create well-crafted commits following conventional commit format.

**Usage**: `/commit` or `/commit {"message": "feat: add OAuth support"}`

**Use when**: Ready to commit changes and want to ensure quality and proper formatting.

---

### `/refactor` - Code Refactoring
Systematic refactoring workflow with safety checks, small incremental steps, and continuous testing.

**Usage**: `/refactor` or `/refactor {"target": "auth module", "goal": "simplify"}`

**Use when**: Improving code structure, reducing complexity, or cleaning up technical debt.

---

### `/document` - Documentation Generation
Generate comprehensive documentation for code, APIs, user guides, or architecture.

**Usage**: `/document` or `/document {"target": "API endpoints", "type": "api"}`

**Use when**: Creating or updating documentation, API docs, or user guides.

---

## Command Format

All commands are markdown files with YAML frontmatter:

```markdown
---
name: command-name
description: Brief description of what the command does
args:
  argname: Description of the argument (optional)
---

Command content with {{.argname}} template variables.
```

## Template Variables

Commands support Go template syntax for arguments:

- `{{.varname}}` - Insert variable value
- `{{if .varname}}...{{else}}...{{end}}` - Conditional sections
- `{{if .varname}}{{.varname}}{{else}}default{{end}}` - Default values

## Customization

You can override built-in commands or add custom ones by creating files in:

- **Project-level**: `.claude/commands/*.md` (highest priority)
- **User-level**: `~/.clem/commands/*.md` (overrides built-in)
- **Built-in**: `commands/*.md` (lowest priority)

Example custom command (`.claude/commands/mycommand.md`):

```markdown
---
name: mycommand
description: My project-specific workflow
args:
  feature: Feature name
---

# Custom Workflow for {{.feature}}

1. Check requirements
2. Write tests
3. Implement feature
4. Review changes
5. Deploy
```

## Using Commands

Commands are invoked through the SlashCommand tool:

```
SlashCommand({
  "command": "plan",
  "args": {
    "feature": "user authentication"
  }
})
```

Or using the `/` prefix in chat:

```
/plan feature="user authentication"
```

## Best Practices

1. **Keep commands focused**: Each command should have one clear purpose
2. **Use templates**: Make commands reusable with template variables
3. **Provide examples**: Show expected usage in the description
4. **Document args**: Clearly describe what each argument does
5. **Test templates**: Verify variable substitution works correctly

## Contributing Commands

When adding new built-in commands:

1. Follow the existing format and structure
2. Include comprehensive instructions and examples
3. Use clear, actionable language
4. Test with various argument combinations
5. Document the command in this README

## See Also

- [Slash Commands Documentation](../docs/claude-docs/10-SLASH-COMMANDS.md)
- [Skills System](../internal/skills/) - For more complex, stateful workflows
- [Template System](../internal/templates/) - For session templates
