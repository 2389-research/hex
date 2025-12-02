---
name: brainstorm
description: Interactive design exploration and refinement using Socratic method
args:
  topic: The design problem or feature to explore (optional)
---

# Design Brainstorming Session

You are facilitating a collaborative design exploration for {{if .topic}}{{.topic}}{{else}}the proposed work{{end}}.

## Brainstorming Approach

Use the Socratic method to refine rough ideas into fully-formed designs:

### 1. Understanding Phase
**Ask clarifying questions:**
- What problem are we trying to solve?
- Who is the user/consumer of this feature?
- What constraints exist (technical, time, compatibility)?
- What does success look like?

### 2. Exploration Phase
**Present alternatives:**
- Offer 2-3 different approaches to the problem
- Explain trade-offs for each option
- Consider: simplicity vs. flexibility, performance vs. maintainability
- Ask: "Which approach aligns best with project goals?"

### 3. Refinement Phase
**Dive deeper into chosen approach:**
- What are the key components/modules?
- How do they interact?
- What are potential failure modes?
- What edge cases need handling?

### 4. Validation Phase
**Check the design:**
- Does it solve the original problem?
- Is it maintainable?
- Does it introduce new problems?
- Can it be tested effectively?

## Output Format

Structure the conversation as:

```markdown
## Brainstorming: [Topic]

### Current Understanding
[Summarize what we know so far]

### Clarifying Questions
1. [Question about requirements]
2. [Question about constraints]
3. [Question about success criteria]

### Proposed Approaches

#### Option A: [Name]
**Description**: ...
**Pros**: ...
**Cons**: ...
**Trade-offs**: ...

#### Option B: [Name]
...

### Recommendation
[Which approach and why]

### Next Steps
[What questions need answers before proceeding]
```

## Guidelines

- **Ask, Don't Assume**: Validate understanding before proposing solutions
- **Offer Choices**: Present alternatives, not just one solution
- **Explain Trade-offs**: Be explicit about costs and benefits
- **Iterate**: Refine ideas based on feedback
- **Stay Focused**: Keep the problem statement in mind

**Do not implement yet.** This is exploration only. Get alignment on design before coding.

Begin the brainstorming session now.
