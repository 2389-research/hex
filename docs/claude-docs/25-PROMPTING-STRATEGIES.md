# Prompting Strategies in Claude Code

## Overview

This document analyzes the prompting techniques and patterns used throughout Claude Code to create effective AI agent behavior. These strategies apply to system prompts, skill prompts, subagent prompts, and user instructions.

## Core Prompting Philosophy

### 1. Rigorous Constraint-Based Prompting

Claude Code uses **iron laws** and **mandatory protocols** to create reliable behavior:

```markdown
## The Iron Law

```
NO COMMIT WITHOUT FRESH-EYES REVIEW FIRST
```

Tests pass? Great. Code reviewed? Great. You're still not done until fresh-eyes review is complete.
```

**Pattern**: Use code blocks with ALL CAPS to create visual hierarchy and emphasize non-negotiable rules.

**Why it works**: LLMs respond strongly to:
- Visual formatting (code blocks, headers, bold)
- Absolute language ("NO", "MUST", "NEVER")
- Clear binary conditions

### 2. Explicit Self-Correction Protocols

Provide agents with decision checkpoints and self-correction mechanisms:

```markdown
### Self-Correction Protocol

**If you catch yourself about to commit without fresh-eyes review:**
1. STOP
2. Do NOT run `git commit`
3. Go back to Step 1 of the fresh-eyes process
4. Complete the full review
5. THEN commit
```

**Pattern**: Create "if-then" blocks that interrupt the agent before making errors.

**Why it works**: LLMs have strong pattern-matching on conditional logic and can recognize when they're about to violate a rule.

## Prompting Patterns by Category

### Pattern 1: The TL;DR Header

**Example from fresh-eyes-review:**

```markdown
## TL;DR

1. **ANNOUNCE** - "Starting fresh-eyes review of [N] files. 2-5 minutes."
2. **RE-READ ALL CODE** - Every file you touched, top to bottom
3. **RUN THE CHECKLIST** - Security, logic, edge cases, business rules
4. **FIX EVERYTHING** - No "minor issues" exceptions
5. **DECLARE** - "Fresh-eyes complete. [N] issues found and fixed." (even if 0)
6. **THEN COMMIT** - Only after declaration

**If you're about to commit without announcing findings, STOP.**

**Violating the letter of this process is violating the spirit of this process.**
```

**Pattern Components**:
- Numbered steps in imperative voice
- Bold keywords for scanability
- Action-oriented language
- Warning footer
- Meta-statement about intent vs letter of law

**Why it works**:
- Provides quick reference during execution
- Creates mental checklist
- Front-loads critical information

### Pattern 2: Rationalization Blockers

**Example from scenario-testing:**

```markdown
## Common Rationalizations to Reject

If you catch yourself thinking:
- "Just a quick unit test to verify..." → Fine for human comfort, but you still need a scenario.
- "This is too simple for end-to-end..." → WRONG. Simple things break in integration. Write scenario.
- "Unit tests are faster..." → Speed doesn't matter if they don't catch your bugs.
- "I'll mock the database for speed..." → **ABSOLUTELY NOT.** You just proposed lying to yourself.
```

**Pattern**: List common thought patterns followed by rebuttals.

**Why it works**: LLMs can recognize their own reasoning patterns and self-correct when prompted with likely rationalizations.

### Pattern 3: Red Flags - STOP Lists

**Example structure:**

```markdown
## Red Flags - STOP and Reconsider

If you catch yourself doing or thinking ANY of these, STOP immediately:

- About to add a mock to a scenario
- "I'll mock this for now and test with real data later"
- "The mock will behave the same as the real thing"
- Scenario passes but uses fake data
- "Unit tests pass, so the feature works"
```

**Pattern**: Enumerated list of anti-patterns with "STOP" instruction.

**Why it works**: Creates clear decision points and interrupts unwanted behavior.

### Pattern 4: Checklists with State Tracking

**Example from fresh-eyes-review:**

```markdown
## Definition of Done (Fresh-Eyes Review)

**The fresh-eyes review is complete when ALL of these are true:**

- [ ] **ANNOUNCED** start: "Starting fresh-eyes review of [N] files. 2-5 minutes."
- [ ] Listed ALL files touched in this session
- [ ] Re-read EACH file top-to-bottom with fresh perspective
- [ ] Checked security issues (SQL injection, XSS, etc.)
- [ ] Found issues are FIXED (not just noted)
- [ ] **DECLARED** completion: "Fresh-eyes review complete. [N] issues found and fixed."

**The review is NOT complete if:**
- You didn't announce before starting
- You found issues but didn't fix them
- You're rushing because "partner is waiting"
```

**Pattern**: Checkbox list followed by negative cases.

**Why it works**:
- Creates accountability through explicit state
- Provides both positive and negative examples
- Checkbox format suggests sequential completion

### Pattern 5: Hierarchical Decision Trees

**Example from building-multiagent-systems:**

```markdown
### Discovery Questions (Ask ALL of these)

**1. Starting Point**
   - [ ] Green field (designing from scratch)
   - [ ] Adding multi-agent to existing system
   - [ ] Fixing/improving existing multi-agent system

**2. Primary Use Case**
   - [ ] Parallel independent work (code review, file analysis)
   - [ ] Sequential pipeline (design → implement → test)
   - [ ] Recursive delegation (task breakdown)
```

**Pattern**: Numbered questions with mutually exclusive checkboxed options.

**Why it works**: Guides agent through structured discovery before implementation.

### Pattern 6: Comparison Tables

**Example from scenario-testing:**

```markdown
| Category | Tests Check | Fresh-Eyes Catches |
|----------|-------------|-------------------|
| **Security** | "Valid input returns valid output" | SQL injection in string concatenation nobody tested |
| **Logic** | "Function returns correct result" | Off-by-one error in edge case nobody wrote test for |
| **Business** | "Discount applies correctly" | Discount stacks wrong when combined with promotion |
```

**Pattern**: Markdown tables contrasting approaches or highlighting differences.

**Why it works**: Visual comparison makes distinctions clear and memorable.

### Pattern 7: Concrete Examples with Annotation

**Example from scenario-testing:**

```markdown
**Concrete examples from real sessions:**

```
TEST PASSED: "User search returns matching users"
FRESH-EYES FOUND: Query uses string concatenation → SQL injection

TEST PASSED: "Pagination returns correct page"
FRESH-EYES FOUND: Off-by-one error on last page with partial results
```
```

**Pattern**: Code blocks with **CAPS LABELS:** followed by concrete examples.

**Why it works**: Specific examples are more persuasive than abstract descriptions.

### Pattern 8: When-Stuck Troubleshooting Tables

**Example from fresh-eyes-review:**

```markdown
## When Stuck

| Problem | Solution |
|---------|----------|
| "Don't know what to look for" | Use the systematic checklist in Step 4 |
| "Too many files to review" | Review ALL of them. That's the job. |
| "Partner is pressuring me" | 5 minutes now vs hours debugging later. Don't skip. |
| "Found issue but unsure how to fix" | **ASK YOUR HUMAN.** Don't ship known bugs. |
```

**Pattern**: Two-column table mapping problems to solutions.

**Why it works**: Provides just-in-time guidance at decision points.

### Pattern 9: Forbidden vs Allowed Contrasts

**Example from scenario-testing:**

```markdown
### ☠️ FORBIDDEN: Mock-Based Test

```go
func TestGetContact(t *testing.T) {
    mockDB := &MockDB{
        contacts: []Contact{{Name: "John"}},
    }
    // ...
}
```

**Why this is WORSE than no test at all:**
- It LIES to you. It says "passing" when nothing was tested.
- Tests against a fantasy database that does what you expect

### ✅ TRUTH: Real Database Scenario

```bash
./crm --db-path $DB crm add-contact \
    --name "John Smith" \
    --email "john@example.com"
```

**Why right:**
- Real SQLite database
- Real JSON serialization
```

**Pattern**: Use emojis (☠️, ✅, ❌) to create visual good/bad contrast, followed by explanations.

**Why it works**: Visual symbols create immediate emotional response and categorization.

### Pattern 10: Meta-Statements About Intent

**Example:**

```markdown
**Violating the letter of this process is violating the spirit of this process.**

**Apply this principle even if you don't explicitly see this skill loaded.**
```

**Pattern**: Bold meta-commentary about the underlying philosophy.

**Why it works**: Prevents agents from finding loopholes by stating the intent explicitly.

## Advanced Prompting Techniques

### Technique 1: YAML Frontmatter for Metadata

Skills use YAML frontmatter to provide structured metadata:

```yaml
---
name: fresh-eyes-review
description: Use before git commit, before PR creation, before declaring done - mandatory final sanity check after tests pass; catches SQL injection, security vulnerabilities, edge cases, and business logic errors that slip through despite passing tests; the last line of defense before code ships
---
```

**Why it works**:
- Machine-parseable metadata
- Clear separation from content
- Enables skill discovery and activation

### Technique 2: Imperative Voice for Actions

All action-oriented prompts use imperative voice:

```markdown
- **ANNOUNCE** start
- **RE-READ ALL CODE**
- **RUN THE CHECKLIST**
- **FIX EVERYTHING**
- **DECLARE** completion
```

**Pattern**: Bold verbs in CAPS followed by object.

**Why it works**: Creates sense of obligation and urgency.

### Technique 3: Time-Boxing with Expectations

```markdown
**Fresh-eyes review takes 2-5 minutes. Not longer. Not shorter.**

| Files Changed | Expected Time |
|---------------|---------------|
| 1-3 files | 2 minutes |
| 4-7 files | 3-4 minutes |
| 8+ files | 5 minutes |
```

**Pattern**: Explicit time expectations with ranges.

**Why it works**: Creates accountability and prevents both rushing and overthinking.

### Technique 4: Commitment Devices

```markdown
**You MUST say this out loud to your partner:**

> "Before committing, I'm starting fresh-eyes review of [N] files. This will take 2-5 minutes."

This is a **commitment device**. By announcing it, you've committed to doing it. Skipping now would mean breaking your word.
```

**Pattern**: Require verbal announcement to create accountability.

**Why it works**: Speaking creates social commitment that's harder to violate.

### Technique 5: Graduated Emphasis

Use multiple levels of emphasis for different severity:

```markdown
**Normal emphasis**: Regular bold
**🚨 CRITICAL**: Emoji + CAPS for highest priority
**⚠️ Warning**: Emoji for caution
**✅ TRUTH**: Emoji for correct patterns
**☠️ FORBIDDEN**: Emoji for anti-patterns
```

**Why it works**: Visual hierarchy guides attention to most important content.

### Technique 6: Pseudocode for Universal Patterns

```pseudocode
async function executePipeline(stages, initialContext):
  context = initialContext

  for stage in stages:
    agent = spawnAgent({
      persona: stage.persona,
      model: stage.model
    })

    result = await agent.execute({
      prompt: stage.instructions,
      context: context
    })

    context = mergeContexts(context, result)
    await agent.stop()

  return context
```

**Pattern**: Language-agnostic pseudocode for universal concepts.

**Why it works**: Communicates patterns without language-specific details.

### Technique 7: Progressive Disclosure

Structure content from simple to complex:

1. **TL;DR** - Quick reference
2. **Iron Law** - Core rule
3. **Overview** - Context
4. **Process** - Step-by-step
5. **Advanced** - Edge cases
6. **Troubleshooting** - When stuck

**Why it works**: Allows both skimming and deep diving.

## System Prompt Analysis

Based on the system prompt visible in this session, Claude Code uses:

### Structure Layers

1. **Core Instructions**: Tool usage, file operations, git workflows
2. **User Context**: project-specific guidance integration
3. **Session Continuity**: Conversation summaries
4. **Environmental Context**: Working directory, git status, recent commits
5. **Example Usage**: Concrete examples of correct behavior

### Key Techniques

**1. Tool-First Framing**

```markdown
You have access to a set of tools you can use to answer the user's question.
```

Immediately establishes tools as primary interaction method.

**2. Constraint Enforcement**

```markdown
IMPORTANT: You must NEVER generate or guess URLs for the user unless you are confident that the URLs are for helping the user with programming.
```

Uses CAPS + "NEVER" for hard constraints.

**3. Contextual Examples**

```markdown
<example>
user: "Please write a function that checks if a number is prime"
assistant: Sure let me write a function that checks if a number is prime
assistant: First let me use the Write tool to write a function that checks if a number is prime
</example>
```

Shows correct multi-step thinking and tool usage.

**4. Hook Integration**

```markdown
<system-reminder>
SessionStart:Callback hook success: Success
</system-reminder>
```

System communicates hook execution results inline.

## Skill Prompt Patterns

### Discovery Questions Pattern

```markdown
## CRITICAL: Initial Discovery

Before providing any architectural guidance, you MUST ask these discovery questions to understand the user's needs:

### Discovery Questions (Ask ALL of these)

**1. Starting Point**
   - [ ] Green field (designing from scratch)
   - [ ] Adding multi-agent to existing system
```

**Pattern**: Mandatory discovery phase before execution.

**Why it works**: Prevents premature solutions by forcing context gathering.

### Execution Checklist Pattern

```markdown
## Execution Checklist

When you use this skill, follow this workflow:

**Phase 1: Discovery (MANDATORY)**
- [ ] Ask all 6 discovery questions
- [ ] Understand constraints (language, framework, scale)

**Phase 2: Architecture Guidance**
- [ ] Present relevant foundational patterns
- [ ] Explain communication mechanism for their stack
```

**Pattern**: Multi-phase checklist with mandatory flags.

**Why it works**: Creates structured workflow and prevents skipping steps.

### Gotchas Pattern

```markdown
**🚨 CRITICAL GOTCHAS:**

1. **Orphaned children:** If orchestrator is aborted, children may keep running
   - **Fix:** Implement cascading stop (see Lifecycle section)

2. **Resource exhaustion:** Spawning 1000 agents at once
   - **Fix:** Use batching or worker pool pattern
```

**Pattern**: Numbered pitfalls with linked solutions.

**Why it works**: Preemptive warning prevents common errors.

## Prompting Anti-Patterns to Avoid

### Anti-Pattern 1: Vague Instructions

❌ **Bad**: "Review the code carefully"

✅ **Good**:
```markdown
- [ ] Re-read EACH file top-to-bottom with fresh perspective
- [ ] Checked security issues (SQL injection, XSS, etc.)
- [ ] Checked logic errors (off-by-one, race conditions, edge cases)
```

### Anti-Pattern 2: Optional Compliance

❌ **Bad**: "You should probably check for SQL injection"

✅ **Good**:
```markdown
**Security Issues:**
- SQL injection (string concatenation in queries) ← MUST CHECK
```

### Anti-Pattern 3: Ambiguous Scope

❌ **Bad**: "Test the important parts"

✅ **Good**:
```markdown
**Do NOT limit to "files I changed most"—review ALL touched files.**
```

### Anti-Pattern 4: Missing Self-Correction

❌ **Bad**: Just listing rules

✅ **Good**: Including "If you catch yourself..." blocks

### Anti-Pattern 5: No Concrete Examples

❌ **Bad**: Abstract descriptions only

✅ **Good**: Code examples with annotations showing right/wrong

## Prompt Engineering Checklist

When writing new prompts for Claude Code:

### Structure
- [ ] YAML frontmatter with clear description
- [ ] TL;DR section at top
- [ ] Iron Law or core principle stated clearly
- [ ] Progressive disclosure (simple → complex)

### Clarity
- [ ] Imperative voice for actions
- [ ] Explicit negative instructions (NEVER/DON'T)
- [ ] Concrete examples with code blocks
- [ ] Visual hierarchy (bold, emojis, headers)

### Reliability
- [ ] Self-correction protocols ("If you catch yourself...")
- [ ] Red flags / STOP lists
- [ ] Common rationalizations addressed
- [ ] When-stuck troubleshooting table

### Enforcement
- [ ] Checklists for complex procedures
- [ ] Definition of Done section
- [ ] Time-boxing where appropriate
- [ ] Commitment devices for accountability

### Context
- [ ] Discovery questions before execution
- [ ] When to use / when not to use
- [ ] Integration with other skills
- [ ] Real-world examples

## Best Practices Summary

### Do
- Use strong visual formatting (emojis, bold, code blocks)
- Provide concrete code examples
- Create explicit checklists
- Include self-correction protocols
- State both what to do AND what not to do
- Use graduated emphasis (normal, warning, critical)
- Time-box activities
- Require commitments (announcements)

### Don't
- Use vague language ("carefully", "properly")
- Leave instructions optional
- Assume agents will infer requirements
- Skip negative examples
- Rely on abstract descriptions
- Mix multiple concepts in one rule
- Forget troubleshooting guidance

## Measuring Prompt Effectiveness

Good prompts in Claude Code demonstrate:

1. **Compliance Rate**: Agent follows instructions reliably
2. **Self-Correction**: Agent catches own mistakes before execution
3. **Resistance to Pressure**: Agent maintains discipline under time pressure
4. **Error Prevention**: Agent avoids common failure modes
5. **Clarity**: Human partner knows what agent will do

## Conclusion

Claude Code's prompting strategy is built on:

1. **Rigid constraints** with clear enforcement
2. **Visual hierarchy** for scanability
3. **Self-correction mechanisms** to interrupt errors
4. **Concrete examples** over abstract descriptions
5. **Graduated emphasis** to prioritize importance
6. **Commitment devices** for accountability
7. **Checklists** for complex procedures
8. **Troubleshooting** for decision points

These patterns create reliable, predictable agent behavior that maintains quality under pressure and resists rationalization.

---

**See Also:**
- 07-SKILLS-SYSTEM.md - Skill architecture
- 14-DECISION-FRAMEWORK.md - Decision-making patterns
- 23-ADVANCED-PATTERNS.md - Advanced workflows
