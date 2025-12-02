# Plan Mode Workflow

## Overview

Plan Mode is a special workflow in Claude Code where the agent focuses on creating detailed implementation plans rather than immediately executing code. This document explains what plan mode is, when to use it, and how it works.

**Key concept**: Plan mode separates **planning** from **execution**, allowing for thorough design before implementation.

## What is Plan Mode?

Plan mode is activated via the `/superpowers:write-plan` slash command. In this mode:

```
Normal Mode:          Plan Mode:
┌─────────────────┐   ┌─────────────────┐
│ User Request    │   │ User Request    │
│       ↓         │   │       ↓         │
│ Design + Code   │   │ Design Only     │
│       ↓         │   │       ↓         │
│ Implementation  │   │ Detailed Plan   │
│       ↓         │   │       ↓         │
│ Testing         │   │ Exit Plan Mode  │
│       ↓         │   │       ↓         │
│ Completion      │   │ Await Approval  │
└─────────────────┘   └─────────────────┘
                              ↓
                      ┌─────────────────┐
                      │ (Optional)      │
                      │ Execute Plan    │
                      └─────────────────┘
```

## When to Use Plan Mode

### Use Plan Mode When:

✅ **Complex multi-phase projects**
- Refactoring large systems
- Adding major features with multiple components
- Database migrations with data integrity concerns

✅ **Need stakeholder approval**
- Changes require review before implementation
- Multiple team members need to understand approach
- Budget/time constraints require planning

✅ **Unclear requirements**
- Need to explore solution space
- Multiple valid approaches exist
- Tradeoffs need to be evaluated

✅ **Learning/teaching**
- Want to understand approach before seeing code
- Teaching implementation patterns
- Documenting architectural decisions

### Don't Use Plan Mode When:

❌ **Simple, obvious tasks**
- Fixing typos
- Adding single functions
- Updating documentation

❌ **Urgent fixes**
- Production bugs need immediate fixes
- Quick iterations preferred

❌ **Exploratory work**
- Not sure what's needed yet
- Discovery phase (use Explore subagent instead)

## Invoking Plan Mode

### Via Slash Command

```bash
/superpowers:write-plan
```

This loads the plan mode skill and switches context to planning.

### User Phrases That Trigger Planning

- "Create a plan for..."
- "Design an approach for..."
- "Outline how we should..."
- "What's the implementation strategy for..."

## Plan Mode Workflow

### Step 1: Understanding Requirements

In plan mode, Claude first ensures complete understanding:

```
1. Read user request
2. Identify ambiguities
3. Ask clarifying questions (AskUserQuestion)
4. Gather context from codebase
5. Research relevant patterns/libraries
```

**Example interaction**:

```
User: "Add user authentication"

Claude in Plan Mode:
"Before creating the plan, I need to clarify:

1. Authentication method: OAuth2, JWT, or sessions?
2. User storage: Existing database or new?
3. Required features: Registration, password reset, 2FA?
4. Timeline: MVP or full featured?"

[Uses AskUserQuestion for structured responses]
```

### Step 2: Creating the Plan

The plan should include:

#### Required Sections

**1. Overview**
- What we're building
- Why this approach
- Key decisions and tradeoffs

**2. Prerequisites**
- Required tools/libraries
- Database changes
- Configuration updates

**3. Implementation Phases**

Each phase should specify:
- **What**: Specific tasks
- **Where**: File paths and line numbers (if modifying existing)
- **How**: Code examples or pseudocode
- **Why**: Rationale for approach
- **Verify**: How to test this phase

**Example phase**:

```markdown
### Phase 1: Database Schema

**What**: Create users table with authentication fields

**Where**: db/migrations/001_create_users.sql

**How**:
​```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  last_login TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
​```

**Why**: Email as primary identifier is industry standard. Password hashing security handled in application layer.

**Verify**:
- Run migration: `psql -f db/migrations/001_create_users.sql`
- Check table: `\d users` should show structure
- Test uniqueness: Inserting duplicate email should fail
```

**4. Dependencies and Order**

Show what depends on what:

```
Phase 1: Database Schema
    ↓
Phase 2: User Model (depends on schema)
    ↓
Phase 3: Auth Routes (depends on model)
    ↓
Phase 4: Frontend Integration (depends on routes)
```

**5. Testing Strategy**

- Unit tests for each component
- Integration tests for flows
- Security tests for vulnerabilities

**6. Rollback Plan**

- How to undo changes if needed
- Database rollback scripts
- Feature flags to disable

**7. Estimated Effort**

- Time per phase
- Complexity assessment
- Potential blockers

### Step 3: Exiting Plan Mode

When plan is complete, use the `ExitPlanMode` tool:

```python
ExitPlanMode(plan="""
# Implementation Plan: User Authentication

## Overview
Implement JWT-based authentication with email/password login...

## Prerequisites
- bcrypt library for password hashing
- jsonwebtoken library for JWT
- PostgreSQL database

## Implementation Phases

### Phase 1: Database Schema
[Detailed steps...]

### Phase 2: User Model
[Detailed steps...]

...

## Testing Strategy
[Test plan...]

## Estimated Effort
Total: 8-10 hours
- Phase 1: 1 hour
- Phase 2: 2 hours
- Phase 3: 3 hours
- Phase 4: 2 hours
- Testing: 2 hours
""")
```

This presents the plan to the user and exits plan mode.

### Step 4: User Decision

After seeing the plan, the user can:

1. **Approve and execute**: "Looks good, proceed with implementation"
2. **Request changes**: "Change Phase 2 to use Redis instead"
3. **Defer execution**: "Thanks, I'll implement this myself"
4. **Ask questions**: "Why JWT over sessions?"

## Plan Quality Checklist

A good plan should be:

- [ ] **Specific**: File paths, function names, exact changes
- [ ] **Ordered**: Clear sequence of phases
- [ ] **Justified**: Explains why chosen approach
- [ ] **Testable**: Verification steps for each phase
- [ ] **Reversible**: Rollback strategy included
- [ ] **Realistic**: Time estimates and complexity noted
- [ ] **Complete**: No "TODO: figure this out later"

## Example: Good vs Bad Plans

### ❌ Bad Plan

```markdown
## Plan: Add Authentication

1. Set up database
2. Create user model
3. Add login endpoint
4. Test it

Done!
```

**Problems**:
- No file paths
- No code examples
- No "why" explanations
- No testing details
- No dependencies
- No time estimates

### ✅ Good Plan

```markdown
## Implementation Plan: JWT Authentication System

## Overview
Implement stateless authentication using JWT tokens with 24-hour expiry.
Chose JWT over sessions for horizontal scalability and API-first design.

## Prerequisites
- Install: `npm install bcrypt jsonwebtoken`
- Database: PostgreSQL (already configured)
- Environment: Add JWT_SECRET to .env

## Phase 1: Database Schema (1 hour)

**File**: `db/migrations/001_create_users.sql`

**Changes**:
​```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);
​```

**Why**: Simple schema, email as unique identifier, password stored as hash only.

**Verify**:
- Run: `npm run migrate`
- Check: `psql -c "\d users"`
- Should show 4 columns

## Phase 2: User Model (2 hours)

**File**: `src/models/User.js` (new file)

**Code**:
​```javascript
const bcrypt = require('bcrypt');
const db = require('../db');

class User {
  static async create(email, password) {
    const hash = await bcrypt.hash(password, 12);
    const result = await db.query(
      'INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING *',
      [email, hash]
    );
    return result.rows[0];
  }

  static async findByEmail(email) {
    const result = await db.query(
      'SELECT * FROM users WHERE email = $1',
      [email]
    );
    return result.rows[0];
  }

  static async verifyPassword(user, password) {
    return bcrypt.compare(password, user.password_hash);
  }
}

module.exports = User;
​```

**Why**:
- bcrypt with 12 rounds (industry standard)
- Parameterized queries prevent SQL injection
- Static methods for cleaner API

**Verify**:
- Run tests: `npm test src/models/User.test.js`
- Should pass: create user, find user, verify password

## Phase 3: Auth Routes (3 hours)

**File**: `src/routes/auth.js` (new file)

**Changes**: [Detailed implementation...]

**Dependencies**: Requires Phase 2 User model

...

## Testing Strategy

### Unit Tests
- User.create() hashes password
- User.verifyPassword() works correctly
- JWT generation includes correct claims

### Integration Tests
- Full login flow: POST /auth/login → returns JWT
- Token verification: Valid token → authorized
- Invalid credentials → 401 error

### Security Tests
- SQL injection attempts fail
- Bcrypt prevents timing attacks
- JWT signature verification

## Rollback

If issues arise:
1. Drop users table: `DROP TABLE users;`
2. Remove auth routes from server
3. Remove dependencies from package.json
4. Deploy previous version

## Estimated Effort

Total: 6-8 hours

- Phase 1: Database (1 hour)
- Phase 2: User model (2 hours)
- Phase 3: Auth routes (3 hours)
- Testing: 2 hours

Potential blockers:
- Database migration conflicts (add 1 hour)
- Password policy requirements (add 1 hour)
```

## Integration with Other Tools

### Plan Mode + Subagents

```python
# In plan mode, can use subagents for research:

Task(
    description="Research auth patterns",
    prompt="Research modern authentication patterns for Node.js APIs...",
    subagent_type="Explore"
)

# Then incorporate findings into plan
```

### Plan Mode + AskUserQuestion

```python
# Clarify requirements during planning:

AskUserQuestion(questions=[
    {
        "question": "Which authentication flow?",
        "header": "Auth type",
        "multiSelect": False,
        "options": [
            {"label": "JWT", "description": "Stateless, good for APIs"},
            {"label": "Sessions", "description": "Stateful, traditional"},
            {"label": "OAuth2", "description": "Third-party providers"}
        ]
    }
])
```

### Plan Mode + Skills

Plan mode itself is a skill (`superpowers:write-plan`). It can invoke other skills:

```
Plan Mode
  → Uses brainstorming skill for ideation
  → Uses systematic-debugging skill to analyze existing code
  → Produces implementation plan
```

## Executing the Plan

After plan approval, execution can happen:

**Option 1: Manual execution**
- User implements following the plan
- Plan serves as documentation

**Option 2: Automated execution**
- User: "Execute the plan"
- Claude implements each phase sequentially
- Tests after each phase
- Reports progress

**Option 3: Subagent-driven execution**
```python
# Use /superpowers:execute-plan
# Or manually:
for phase in plan.phases:
    Task(
        description=f"Implement {phase.name}",
        prompt=f"Execute this phase:\n{phase.details}",
        subagent_type="general-purpose"
    )
```

## Best Practices

### DO:

✅ **Be specific** - Include file paths, line numbers, exact code
✅ **Explain tradeoffs** - Why this approach over alternatives
✅ **Include examples** - Code snippets, not just descriptions
✅ **Define success** - How to verify each phase works
✅ **Estimate realistically** - Include buffer for unknowns
✅ **Plan for failure** - Rollback and recovery steps

### DON'T:

❌ **Be vague** - "Set up the database" (what tables? what schema?)
❌ **Assume knowledge** - Explain why decisions were made
❌ **Skip verification** - Every phase needs testing steps
❌ **Ignore dependencies** - Show what must happen before what
❌ **Over-promise** - Better to under-estimate than over-estimate

## Common Plan Mode Patterns

### Pattern 1: Database-First

```
1. Design schema
2. Create models
3. Build API layer
4. Add frontend
5. Test end-to-end
```

### Pattern 2: Outside-In

```
1. Define API contract (what frontend needs)
2. Implement endpoints (mock data)
3. Add business logic
4. Implement database
5. Wire everything together
```

### Pattern 3: Vertical Slice

```
1. Implement one complete feature (database → frontend)
2. Verify it works end-to-end
3. Repeat for next feature
```

## Troubleshooting

### Problem: Plan Too Vague

**Symptom**: User asks "How exactly do I do X?"

**Fix**: Add more detail - code examples, file paths, exact commands

### Problem: Plan Too Detailed

**Symptom**: Plan is 50 pages long

**Fix**: Break into multiple plans, focus on high-level phases

### Problem: Plan Can't Be Executed

**Symptom**: Missing steps, unclear dependencies

**Fix**: Walk through plan step-by-step, verify each phase is complete

## Summary

**Plan Mode**:
- Activated via `/superpowers:write-plan`
- Focuses on creating detailed implementation plans
- Exits via `ExitPlanMode` tool
- Separates planning from execution

**Good plans include**:
1. Overview and rationale
2. Prerequisites
3. Phased implementation with specific steps
4. Testing strategy
5. Rollback plan
6. Time estimates

**Use when**: Complex projects, need approval, unclear requirements
**Don't use when**: Simple tasks, urgent fixes, exploratory work

---

**See Also:**
- 03-TOOL-SYSTEM.md - ExitPlanMode tool documentation
- 05-SUBAGENT-SYSTEM.md - Subagent system overview
- 14-DECISION-FRAMEWORK.md - When to plan vs execute
- 25-PROMPTING-STRATEGIES.md - Writing effective plans
- 07-SKILLS-SYSTEM.md - superpowers:write-plan skill
