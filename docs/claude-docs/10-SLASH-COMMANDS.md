# Slash Commands

## Overview

Slash commands are a powerful mechanism in Claude Code that allow you to create reusable prompts, workflows, and instructions. They expand into full prompts when invoked, enabling consistent, repeatable interactions.

## What Slash Commands Are

Slash commands are markdown files stored in `~/.claude/commands/` that contain prompts or instructions. When you type a slash command (e.g., `/brainstorm`), Claude Code:

1. Locates the corresponding markdown file
2. Reads the file content
3. Expands the content as if you typed it directly
4. Processes the expanded prompt

This appears in the conversation as:
```
<command-message>{name} is running…</command-message>
```

Followed by the expanded prompt content in the next message.

## Built-in Commands

Claude Code includes several built-in slash commands available in this session:

### Planning and Design

- **/brainstorm**: Interactive design refinement using Socratic method
  - References: `@skills/collaboration/brainstorming/SKILL.md`
  - Uses structured questioning to refine rough ideas

- **/plan**: Create detailed implementation plan
  - Generates comprehensive task breakdown
  - File: `plan.md`

- **/write-plan**: Create detailed implementation plan with bite-sized tasks
  - Creates plans for engineers with zero codebase context
  - File: `write-plan.md`

- **/execute-plan**: Execute plan in batches with review checkpoints
  - Loads plan, reviews, executes in controlled batches
  - File: `execute-plan.md`

### Development Workflows

- **/plan-tdd**: TDD-focused planning
  - Test-driven development approach
  - File: `plan-tdd.md`

- **/find-missing-tests**: Identify untested code
  - File: `find-missing-tests.md`

- **/do-todo**: Process todo items
  - File: `do-todo.md`

- **/do-issues**: Work through GitHub issues
  - File: `do-issues.md`

### Code Review and Security

- **/careful-review**: Thorough code review
  - Detailed examination of changes
  - File: `careful-review.md`

- **/security-review**: Security-focused code review
  - Vulnerability scanning and security analysis
  - File: `security-review.md`

- **/suspicious**: Flag suspicious code patterns
  - File: `suspicious.md`

### GitHub Integration

- **/gh-issue**: GitHub issue operations
  - File: `gh-issue.md`

- **/make-github-issues**: Bulk issue creation
  - File: `make-github-issues.md`

- **/plan-gh**: GitHub-integrated planning
  - File: `plan-gh.md`

### Documentation and Setup

- **/session-summary**: Generate session summary
  - File: `session-summary.md`

- **/setup**: Project setup instructions
  - File: `setup.md`

- **/create-dot-file**: Create configuration files
  - File: `create-dot-file.md`

### Prompt Management

- **/do-prompt-plan**: Execute prompt-based plan
  - File: `do-prompt-plan.md`

## Custom Command Creation

### Directory Structure

Custom commands are stored in: `~/.claude/commands/`

Each command is a separate markdown file.

### Basic Command File

Create a file: `~/.claude/commands/mycommand.md`

```markdown
# My Custom Command

Please analyze the codebase and provide recommendations for:

1. Code organization improvements
2. Performance optimizations
3. Security enhancements

Format the results as a prioritized list with implementation difficulty.
```

Usage: `/mycommand`

### Command with References

Commands can reference skills using the `@` syntax:

```markdown
# Refactor Command

@skills/code-quality/refactoring/SKILL.md

Please refactor the code following these principles:
- Extract complex functions
- Improve naming
- Add documentation
```

### Parameterized Commands

While commands don't have formal parameters, you can create templates:

```markdown
# Code Review Template

Please review the following:
- Code quality and maintainability
- Test coverage
- Documentation completeness
- Security considerations

Focus on: [SPECIFY AREA HERE]
```

Usage: Edit the command file before running, or create variants like:
- `/review-backend`
- `/review-frontend`
- `/review-security`

### Multi-step Workflow Commands

```markdown
# Full Feature Workflow

Execute the following steps in order:

1. **Planning Phase**
   - Analyze requirements
   - Create task breakdown
   - Identify dependencies

2. **Implementation Phase**
   - Write tests first (TDD)
   - Implement functionality
   - Document as you go

3. **Review Phase**
   - Run full test suite
   - Security review
   - Performance check

4. **Completion**
   - Create PR description
   - Update changelog
   - Mark tasks complete

Pause for approval after each phase.
```

## Command Expansion Mechanism

### How Expansion Works

1. **Trigger Detection**: When user types `/commandname`
2. **File Lookup**: Search `~/.claude/commands/commandname.md`
3. **Content Reading**: Load markdown file content
4. **Reference Resolution**: Process any `@` references to skills or files
5. **Prompt Injection**: Insert expanded content into conversation
6. **Execution**: Process as if user typed the content directly

### Expansion Indicators

You'll see these messages during expansion:

```
<command-message>brainstorm is running…</command-message>
```

This signals that the command file is being loaded and expanded.

### Reference Resolution

Commands can reference:
- **Skills**: `@skills/path/to/SKILL.md`
- **Files**: `@path/to/file.md`
- **Other commands**: Not directly, but can instruct to run them

## Slash Commands vs Skills

### Slash Commands

- **Location**: `~/.claude/commands/`
- **Format**: Simple markdown files
- **Purpose**: Quick prompts and workflows
- **Invocation**: `/commandname`
- **Expansion**: Content injected into conversation
- **Scope**: Single prompt/instruction set

### Skills

- **Location**: `~/.claude/skills/` or plugin directories
- **Format**: Structured SKILL.md with metadata
- **Purpose**: Complex, reusable capabilities
- **Invocation**: Via Skill tool or `@skills/` references
- **Expansion**: Full skill framework activated
- **Scope**: Multi-turn workflows with state

### When to Use Each

**Use Slash Commands for:**
- Quick prompt templates
- Standard workflows
- Consistent instructions
- Simple parameterized tasks

**Use Skills for:**
- Complex multi-step processes
- Stateful interactions
- Reusable capabilities across projects
- Formal plugin distribution

## Parameters and Arguments

### Pseudo-Parameters via File Variants

Create multiple command files for variations:

```
~/.claude/commands/
├── test-unit.md
├── test-integration.md
├── test-e2e.md
└── test-all.md
```

### Context-Aware Commands

Commands can reference current context:

```markdown
# Context-Aware Review

Review the current changes:

1. Run git status to see modified files
2. Run git diff to see changes
3. Analyze the changes for:
   - Correctness
   - Test coverage
   - Breaking changes

Provide feedback based on the CURRENT state of the repository.
```

### Combining Commands with Skills

```markdown
# Full Stack Feature

First, use the brainstorming skill:
@skills/collaboration/brainstorming/SKILL.md

Then create an implementation plan:
@skills/planning/implementation/SKILL.md

Execute using TDD:
@skills/development/tdd/SKILL.md

Finally, verify with:
@skills/verification/testing/SKILL.md
```

## Advanced Command Patterns

### Checklist Commands

```markdown
# Pre-Commit Checklist

Verify the following before committing:

- [ ] All tests pass (`npm test`)
- [ ] No linting errors (`npm run lint`)
- [ ] Types check (`npm run typecheck`)
- [ ] No console.log statements
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] No commented-out code
- [ ] Meaningful commit message prepared

Report status of each item.
```

### Template Generation Commands

```markdown
# Component Template

Create a new React component with:

1. TypeScript interface for props
2. Functional component implementation
3. Unit tests with React Testing Library
4. Storybook story
5. Documentation comments

Component name: [SPECIFY]
Props: [SPECIFY]
```

### Workflow Automation Commands

```markdown
# Release Workflow

Execute the release process:

1. **Verification**
   - Run full test suite
   - Check for uncommitted changes
   - Verify on main branch

2. **Version Bump**
   - Update package.json version
   - Update CHANGELOG.md
   - Commit version bump

3. **Build**
   - Run production build
   - Verify build artifacts

4. **Tag**
   - Create git tag
   - Push tag to remote

5. **Deploy**
   - Run deployment script
   - Verify deployment

Report status at each step. Stop if any step fails.
```

## Command Organization

### Recommended Directory Structure

```
~/.claude/commands/
├── planning/
│   ├── brainstorm.md
│   ├── plan.md
│   └── write-plan.md
├── development/
│   ├── tdd.md
│   ├── review.md
│   └── refactor.md
├── testing/
│   ├── unit.md
│   ├── integration.md
│   └── e2e.md
├── git/
│   ├── commit.md
│   ├── pr.md
│   └── release.md
└── project/
    ├── setup.md
    └── cleanup.md
```

Note: Commands must be referenced with full path: `/planning/brainstorm`

### Command Naming Conventions

- **Kebab-case**: Use hyphens for multi-word commands
- **Verb-first**: Start with action verbs (review-, create-, setup-)
- **Descriptive**: Clear indication of purpose
- **Consistent**: Similar commands use similar patterns

Examples:
- `/review-code`
- `/create-component`
- `/setup-project`
- `/test-integration`

## Common Command Patterns

### Investigation Commands

```markdown
# Investigate Bug

1. Reproduce the issue
2. Check recent changes (git log, git diff)
3. Review error logs and stack traces
4. Identify potential causes
5. Propose fixes with reasoning

Document findings and recommendations.
```

### Documentation Commands

```markdown
# Document Module

For the current module:

1. Add/update JSDoc comments
2. Create/update README.md
3. Add usage examples
4. Document edge cases
5. Add type definitions

Ensure documentation is comprehensive and accurate.
```

### Cleanup Commands

```markdown
# Code Cleanup

Perform cleanup:

1. Remove unused imports
2. Remove commented-out code
3. Fix formatting inconsistencies
4. Update outdated comments
5. Consolidate duplicate code

Make incremental commits for each cleanup category.
```

## Best Practices

1. **Single Responsibility**: Each command should have one clear purpose
2. **Clear Instructions**: Be explicit about expected behavior
3. **Ordered Steps**: Number multi-step commands
4. **Verification Points**: Include checkpoints for complex workflows
5. **Documentation**: Add comments explaining non-obvious instructions
6. **Reusability**: Make commands general enough for multiple uses
7. **Composability**: Commands should work well together
8. **Context-Aware**: Reference current state (git, files, tests)

## Troubleshooting

### Command Not Found

- Check file exists: `ls ~/.claude/commands/`
- Verify filename matches command (without .md extension)
- Ensure proper markdown format

### Command Not Expanding

- Check for syntax errors in markdown
- Verify file permissions (should be readable)
- Look for malformed skill references

### Unexpected Behavior

- Review expanded prompt in conversation
- Check for conflicting instructions
- Verify skill references are valid

## See Also

- [MCP Servers](09-MCP-SERVERS.md) - External tool integration
- [Configuration](11-CONFIGURATION.md) - Settings and customization
- [Superpowers Skills](https://github.com/superpowers-marketplace/superpowers) - Available skills
