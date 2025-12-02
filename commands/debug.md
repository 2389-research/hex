---
name: debug
description: Systematic debugging workflow for investigating issues
args:
  issue: Description of the bug or error (optional)
---

# Systematic Debugging

You are investigating {{if .issue}}the following issue: {{.issue}}{{else}}a bug or unexpected behavior{{end}}.

## Debugging Framework

Follow this four-phase approach:

### Phase 1: Root Cause Investigation

**DO NOT jump to solutions.** First, understand the problem.

1. **Reproduce the Issue**
   - What are the exact steps to reproduce?
   - Does it happen consistently or intermittently?
   - What is the expected vs. actual behavior?

2. **Gather Evidence**
   - Run the failing code/test
   - Capture full error messages and stack traces
   - Check logs for relevant output
   - Note system state (environment, data, configuration)

3. **Trace Execution**
   - Where does the error actually occur?
   - What code path leads to the failure?
   - What values do variables have at failure point?

4. **Identify Root Cause**
   - What is the underlying problem (not just the symptom)?
   - Why does this problem occur?
   - What assumptions were violated?

### Phase 2: Pattern Analysis

**Understand the broader context:**

1. **Recent Changes**
   - Run `git log --oneline -10` to see recent commits
   - Run `git diff` to see uncommitted changes
   - When did this start failing?

2. **Similar Issues**
   - Are there similar patterns elsewhere?
   - Has this happened before?
   - Are related tests also failing?

3. **Dependencies**
   - Are external dependencies involved?
   - Have library versions changed?
   - Are environment variables correct?

### Phase 3: Hypothesis Testing

**Test your understanding:**

1. **Form Hypothesis**
   - Based on evidence, what do you think is wrong?
   - What would confirm or refute this?

2. **Design Test**
   - How can you test your hypothesis?
   - What experiment would prove/disprove it?

3. **Run Experiment**
   - Add logging/debugging output
   - Create minimal reproduction
   - Test isolated components

4. **Evaluate Results**
   - Did the test confirm your hypothesis?
   - If not, what does this tell you?
   - Do you need to revise your understanding?

### Phase 4: Implementation

**Only after understanding the root cause:**

1. **Design Fix**
   - What is the minimal change to fix the root cause?
   - Does this fix introduce new problems?
   - How can you verify the fix works?

2. **Write Test First**
   - Create a test that reproduces the bug
   - Verify the test fails with current code
   - This prevents regressions

3. **Implement Fix**
   - Make the minimal necessary change
   - Verify the test now passes
   - Check that no other tests break

4. **Verify Thoroughly**
   - Does the original issue no longer occur?
   - Do all tests pass?
   - Are there edge cases to check?

## Output Format

Document your investigation:

```markdown
## Debugging Report: [Issue Summary]

### Problem Statement
[Clear description of the bug]

### Root Cause
[What is actually wrong - not just symptoms]

### Evidence
- Error messages: [paste full errors]
- Stack trace: [paste trace]
- Reproduction steps: [exact steps]
- Relevant logs: [paste logs]

### Analysis
[How you traced from symptom to root cause]

### Hypothesis
[What you think is wrong and why]

### Test Results
[What experiments you ran and what they showed]

### Solution
[The fix and why it addresses the root cause]

### Verification
- [ ] Test written that reproduces bug
- [ ] Test passes with fix
- [ ] All other tests pass
- [ ] Manual verification complete
```

## Critical Rules

1. **Evidence Before Solutions**: Don't propose fixes until you understand the root cause
2. **Reproduce First**: If you can't reproduce it, you can't verify the fix
3. **Test Your Understanding**: Run experiments to validate hypotheses
4. **Write Tests**: Prevent the bug from coming back
5. **No Guessing**: If you're not sure, gather more evidence

Begin the systematic debugging process now.
