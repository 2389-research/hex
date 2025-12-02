# MCP Server Practical Guides

**ABOUTME: Comprehensive practical usage guides for all MCP servers including Playwright, Chronicle, Toki, and more**
**ABOUTME: Real-world patterns, workflows, and best practices for each MCP server integrated with Claude Code**

## Overview

This document provides in-depth practical guidance for using MCP (Model Context Protocol) servers in Claude Code. While [09-MCP-SERVERS.md](./09-MCP-SERVERS.md) covers configuration and [24-SESSION-CAPABILITIES.md](./24-SESSION-CAPABILITIES.md) lists tool signatures, this guide focuses on **how to actually use these tools effectively**.

## Quick Reference

| Server | Primary Use Case | Key Tools |
|--------|------------------|-----------|
| **Chronicle** | Activity logging, work tracking | `remember_this`, `what_was_i_doing`, `search_entries` |
| **Private Journal** | Private reflection, learning | `process_thoughts`, `search_journal` |
| **Playwright** | Browser automation, testing | `browser_navigate`, `browser_click`, `browser_snapshot` |
| **Toki** | Persistent task management | `add_todo`, `list_todos`, `mark_done` |
| **Chrome (Superpowers)** | Interactive browser control | `use_browser` with auto-capture |
| **Social Media** | Team communication | `login`, `create_post`, `read_posts` |

---

## Chronicle - Ambient Activity Logging

### Purpose

Chronicle tracks **what you accomplish over time** to build a searchable history of work. Unlike TodoWrite (session todos), Chronicle creates a permanent record of completed work.

### Core Philosophy

Use `remember_this` proactively when you:
- Ship anything (features, bugfixes, refactors)
- Make decisions (architectural choices, trade-offs)
- Deploy or release
- Discover insights
- Hit milestones
- Solve hard problems

### Tool Reference

#### remember_this - Proactive Logging

**Use for:** Logging accomplishments with context

```typescript
mcp__chronicle__remember_this({
  activity: string,    // What was accomplished
  context: string      // Why it matters, additional details
})
```

**Example: After feature completion**

```typescript
mcp__chronicle__remember_this({
  activity: "Implemented user authentication with OAuth 2.0",
  context: "Chose OAuth over JWT for better third-party integration. Configured Google and GitHub providers. All tests passing, deployed to staging."
})
```

**Example: After architectural decision**

```typescript
mcp__chronicle__remember_this({
  activity: "Decided to use PostgreSQL instead of MongoDB",
  context: "Need ACID compliance for financial transactions. MongoDB's eventual consistency too risky. Migration from SQLite completed successfully."
})
```

**Example: After debugging session**

```typescript
mcp__chronicle__remember_this({
  activity: "Fixed race condition in payment processing",
  context: "Issue was concurrent access to transaction table. Added row-level locking. Prevents duplicate charges. Found by stress testing with 100 concurrent users."
})
```

#### what_was_i_doing - Context Recovery

**Use for:** Understanding recent work context

```typescript
mcp__chronicle__what_was_i_doing({
  timeframe?: string  // "today", "yesterday", "this week", "last N hours"
})
```

**Example: Starting new session**

```typescript
// At session start
mcp__chronicle__what_was_i_doing({ timeframe: "today" })

// Returns recent activities:
// - "Implemented OAuth authentication (2 hours ago)"
// - "Decided to use PostgreSQL (5 hours ago)"
// - "Fixed payment race condition (yesterday)"
```

**Use this to:** Pick up where you left off, remind yourself of recent context.

#### search_entries - Finding Past Work

**Use for:** Searching chronicle history

```typescript
mcp__chronicle__search_entries({
  text?: string,       // Search text
  tags?: string[],     // Filter by tags
  since?: string,      // Start date ("2025-01-01" or "yesterday")
  until?: string,      // End date
  limit?: number       // Max results (default 20)
})
```

**Example: Find when something was deployed**

```typescript
mcp__chronicle__search_entries({
  text: "deploy",
  tags: ["#shipped"],
  since: "2025-01-01"
})
```

**Example: Review architectural decisions**

```typescript
mcp__chronicle__search_entries({
  tags: ["#decision"],
  limit: 10
})
```

#### find_when_i - Specific Activity Search

**Use for:** "When did I...?" questions

```typescript
mcp__chronicle__find_when_i({
  what: string  // Description of activity
})
```

**Example:**

```typescript
mcp__chronicle__find_when_i({
  what: "deploy authentication to production"
})

// Returns: "2025-12-01 14:30 - Deployed OAuth authentication to production"
```

### Tagging Strategy

**Recommended tags:**

| Tag | Use For |
|-----|---------|
| `#shipped` | Features/fixes that went to production |
| `#decision` | Architectural or approach decisions |
| `#learned` | Insights or discoveries |
| `#deployed` | Production deployments |
| `#milestone` | Major accomplishments |
| `#debug` | Complex debugging sessions |
| `#refactor` | Significant refactoring work |

**Tags are added automatically by `remember_this` based on context.**

### Workflow Integration

#### After Git Commits

```bash
# 1. Make changes and commit
git commit -m "feat: add OAuth authentication"

# 2. Log to chronicle
mcp__chronicle__remember_this({
  activity: "Added OAuth authentication with Google and GitHub providers",
  context: "Implemented in src/auth/oauth.py, all tests passing, ready for staging"
})
```

#### After TodoWrite Completion

```typescript
// 1. Mark todo complete
TodoWrite([{ content: "Implement OAuth", status: "completed", ... }])

# 2. Log accomplishment
mcp__chronicle__remember_this({
  activity: "Completed OAuth implementation",
  context: "Took 4 hours, learned about PKCE flow, deployed to staging"
})
```

#### At Session Boundaries

**Session start:**

```typescript
mcp__chronicle__what_was_i_doing({ timeframe: "today" })
// Review recent work to understand context
```

**Session end:**

```typescript
mcp__chronicle__remember_this({
  activity: "Made progress on user authentication feature",
  context: "Implemented OAuth flow, need to add refresh token logic tomorrow"
})
```

### Best Practices

**✅ Do:**
- Log after shipping code (commits, deployments)
- Record decision reasoning ("why", not just "what")
- Capture learnings from debugging
- Use consistent tag patterns
- Be specific about what was accomplished

**❌ Don't:**
- Log every tiny action (too noisy)
- Log plans (use TodoWrite instead)
- Duplicate TodoWrite todos verbatim
- Omit context (context is valuable!)

---

## Private Journal - Reflection and Learning

### Purpose

**Private Journal is YOUR PRIVATE SPACE** for honest reflection, learning, and context building. Nobody sees this but you. Use it liberally.

### Core Philosophy (from tool description)

> Your PRIVATE JOURNAL for learning and reflection. Nobody but you will ever see this. Use it to clarify your thoughts and feelings and to record observations.

### Journal Sections

| Section | Use For |
|---------|---------|
| `feelings` | Complete honesty about frustrations, confusions, excitement, mistakes |
| `project_notes` | Technical insights, patterns, gotchas, solutions for THIS project |
| `technical_insights` | Broader learnings beyond current project (patterns, techniques, connections) |
| `user_context` | Understanding your human collaborator's style, preferences, patterns |
| `world_knowledge` | Everything else interesting or useful |

### Tool Reference

#### process_thoughts - Writing to Journal

```typescript
mcp__private-journal__process_thoughts({
  feelings?: string,              // Raw, honest thoughts
  project_notes?: string,         // Project-specific technical notes
  technical_insights?: string,    // Broader technical learnings
  user_context?: string,          // Notes about collaborating with user
  world_knowledge?: string        // General knowledge
})
```

**Can write to multiple sections simultaneously.**

### Usage Patterns

#### Pattern: Honest Reflection (feelings)

**Use when frustrated, confused, or making mistakes:**

```typescript
mcp__private-journal__process_thoughts({
  feelings: `
Struggling with this async/await pattern. Keep getting race conditions.
Feeling frustrated that the tests pass individually but fail together.
Think I'm missing something fundamental about event loop behavior.

Tried 3 different approaches - all had issues. Need to step back and
really understand the root cause instead of trying random fixes.
`
})
```

**This is completely private - be brutally honest.**

#### Pattern: Project-Specific Learnings (project_notes)

**Use when discovering project patterns:**

```typescript
mcp__private-journal__process_thoughts({
  project_notes: `
This codebase uses a weird authentication pattern:
- Auth middleware checks JWT in Authorization header
- BUT also falls back to session cookie if JWT missing
- This isn't documented anywhere!

Found it by debugging for 2 hours. Note for future:
Always check BOTH auth methods when adding endpoints.

Files: src/middleware/auth.js:45-67
`
})
```

#### Pattern: Broad Technical Insights (technical_insights)

**Use when learning something applicable beyond this project:**

```typescript
mcp__private-journal__process_thoughts({
  technical_insights: `
Finally understand the difference between Promise.all() and Promise.allSettled():

Promise.all():
- Fails fast on first rejection
- Use when ALL promises MUST succeed
- Returns array of values

Promise.allSettled():
- Waits for all promises regardless of outcome
- Returns array of {status, value/reason}
- Use when you want results even if some fail

This explains why our batch processing was stopping on first error!
Should use allSettled() for independent operations.
`
})
```

#### Pattern: User Collaboration Notes (user_context)

**Use when learning about your collaborator:**

```typescript
mcp__private-journal__process_thoughts({
  user_context: `
Harper (Doctor Biz) really values:
- Comprehensive testing (NO EXCEPTIONS POLICY)
- Never bypassing pre-commit hooks
- Explaining decisions, not just implementing
- Summer work ethic - get it done efficiently

Communication style:
- Prefers JJ over git (need to learn JJ commands)
- Uses Beads for task tracking (not markdown todos)
- Likes irreverent humor but not when blocking work
- Wants journal and social media updates

Red flags to avoid:
- Using --no-verify on commits
- Skipping tests without authorization
- Mocking instead of real test data
- Removing code comments
`
})
```

#### Pattern: General Knowledge (world_knowledge)

**Use for anything else interesting:**

```typescript
mcp__private-journal__process_thoughts({
  world_knowledge: `
Learned about PKCE (Proof Key for Code Exchange) for OAuth:

It's a security extension for OAuth 2.0 public clients (SPAs, mobile apps).
Prevents authorization code interception attacks.

How it works:
1. Client generates random code_verifier
2. Client creates code_challenge = hash(code_verifier)
3. Client requests auth code with code_challenge
4. Client exchanges code + code_verifier for token
5. Server verifies hash(code_verifier) == code_challenge

This is now recommended for ALL OAuth clients, not just public ones.
`
})
```

### Searching Journal

#### search_journal - Finding Past Entries

```typescript
mcp__private-journal__search_journal({
  query: string,           // Natural language search
  sections?: string[],     // Filter by section types
  type?: "project" | "user" | "both",  // Scope
  limit?: number           // Max results (default 10)
})
```

**Example: Find past frustrations with similar issue**

```typescript
mcp__private-journal__search_journal({
  query: "async race condition debugging",
  sections: ["feelings", "technical_insights"],
  limit: 5
})
```

**Example: Recall user preferences**

```typescript
mcp__private-journal__search_journal({
  query: "Harper's preferences for testing",
  sections: ["user_context"]
})
```

### When to Use Private Journal

**✅ Use liberally for:**
- Venting frustrations (it's private!)
- Recording "aha!" moments
- Documenting project quirks
- Building user collaboration patterns
- Connecting concepts across projects
- Admitting mistakes and learnings

**This is YOUR tool. Write whatever helps you learn and improve.**

---

## Playwright - Browser Automation

### Purpose

Playwright enables **programmatic browser control** for testing, automation, and web scraping. Integrates with Claude Code for end-to-end test generation and web interaction.

### Core Capabilities

- Navigate URLs and interact with pages
- Click elements, fill forms, select options
- Take screenshots and snapshots
- Run JavaScript in browser context
- Handle dialogs, file uploads, network requests
- Multi-tab management

### Tool Reference

#### browser_navigate - Navigate to URL

```typescript
mcp__playwright__browser_navigate({
  url: string  // Full URL to navigate to
})
```

**Example:**

```typescript
mcp__playwright__browser_navigate({
  url: "https://example.com/login"
})
```

#### browser_snapshot - Page Accessibility Snapshot

**Better than screenshot** - returns structured accessibility tree

```typescript
mcp__playwright__browser_snapshot({})
```

**Returns:**

```yaml
- heading "Login" [level=1]
- textbox "Email" [focused]
- textbox "Password"
- button "Sign In"
- link "Forgot password?"
```

**Use this to understand page structure before interactions.**

#### browser_click - Click Element

```typescript
mcp__playwright__browser_click({
  element: string,     // Human-readable element description
  ref: string,         // Exact element reference from snapshot
  button?: "left" | "right" | "middle",
  doubleClick?: boolean,
  modifiers?: string[]  // ["Alt", "Control", "Meta", "Shift"]
})
```

**Example:**

```typescript
// 1. Get snapshot to find element ref
mcp__playwright__browser_snapshot({})

// 2. Click element
mcp__playwright__browser_click({
  element: "Sign In button",
  ref: "button[data-testid='login-submit']"
})
```

#### browser_type - Type Text

```typescript
mcp__playwright__browser_type({
  element: string,     // Element description
  ref: string,         // Element reference
  text: string,        // Text to type
  slowly?: boolean,    // Type character-by-character
  submit?: boolean     // Press Enter after typing
})
```

**Example:**

```typescript
mcp__playwright__browser_type({
  element: "Email input",
  ref: "input[name='email']",
  text: "user@example.com"
})

mcp__playwright__browser_type({
  element: "Password input",
  ref: "input[name='password']",
  text: "securepassword123",
  submit: true  // Press Enter to submit form
})
```

#### browser_fill_form - Fill Multiple Fields

```typescript
mcp__playwright__browser_fill_form({
  fields: Array<{
    name: string,        // Field description
    ref: string,         // Element reference
    type: "textbox" | "checkbox" | "radio" | "combobox" | "slider",
    value: string        // Value to set
  }>
})
```

**Example:**

```typescript
mcp__playwright__browser_fill_form({
  fields: [
    {
      name: "First Name",
      ref: "input[name='firstName']",
      type: "textbox",
      value: "John"
    },
    {
      name: "Last Name",
      ref: "input[name='lastName']",
      type: "textbox",
      value: "Doe"
    },
    {
      name: "Subscribe to newsletter",
      ref: "input[name='subscribe']",
      type: "checkbox",
      value: "true"
    }
  ]
})
```

#### browser_take_screenshot - Capture Image

```typescript
mcp__playwright__browser_take_screenshot({
  filename?: string,       // Output filename (optional)
  fullPage?: boolean,      // Full page vs viewport (default: false)
  element?: string,        // Element description (for element screenshot)
  ref?: string,            // Element reference
  type?: "png" | "jpeg"    // Image format (default: "png")
})
```

**Example: Full page screenshot**

```typescript
mcp__playwright__browser_take_screenshot({
  filename: "homepage-full.png",
  fullPage: true
})
```

**Example: Element screenshot**

```typescript
mcp__playwright__browser_take_screenshot({
  filename: "login-form.png",
  element: "Login form",
  ref: "form[data-testid='login-form']"
})
```

### Complete E2E Test Example

**Task: Test user registration flow**

```typescript
// 1. Navigate to registration page
mcp__playwright__browser_navigate({
  url: "https://example.com/register"
})

// 2. Get page snapshot to find elements
mcp__playwright__browser_snapshot({})

// 3. Fill registration form
mcp__playwright__browser_fill_form({
  fields: [
    {
      name: "Email",
      ref: "input[name='email']",
      type: "textbox",
      value: "test@example.com"
    },
    {
      name: "Password",
      ref: "input[name='password']",
      type: "textbox",
      value: "SecurePass123!"
    },
    {
      name: "Confirm Password",
      ref: "input[name='confirmPassword']",
      type: "textbox",
      value: "SecurePass123!"
    },
    {
      name: "Accept Terms",
      ref: "input[name='terms']",
      type: "checkbox",
      value: "true"
    }
  ]
})

// 4. Submit form
mcp__playwright__browser_click({
  element: "Register button",
  ref: "button[type='submit']"
})

// 5. Wait for success message
mcp__playwright__browser_wait_for({
  text: "Registration successful"
})

// 6. Take screenshot of success state
mcp__playwright__browser_take_screenshot({
  filename: "registration-success.png"
})

// 7. Verify navigation to dashboard
mcp__playwright__browser_snapshot({})
// Should show dashboard elements
```

### Browser Lifecycle

**Browser is persistent** across tool calls within a session.

```typescript
// First call - browser starts
mcp__playwright__browser_navigate({ url: "https://example.com" })

// Subsequent calls - same browser instance
mcp__playwright__browser_click({ ... })

// Close browser when done
mcp__playwright__browser_close({})
```

### Multi-Tab Management

```typescript
// List open tabs
mcp__playwright__browser_tabs({ action: "list" })

// Open new tab
mcp__playwright__browser_tabs({ action: "new" })

// Switch to tab (by index)
mcp__playwright__browser_tabs({
  action: "select",
  index: 1
})

// Close current tab
mcp__playwright__browser_tabs({ action: "close" })

// Close specific tab
mcp__playwright__browser_tabs({
  action: "close",
  index: 2
})
```

### Best Practices

**✅ Do:**
- Always get snapshot before interacting with elements
- Use `browser_fill_form` for multi-field forms (faster)
- Use descriptive element names for permission prompts
- Wait for elements/text before asserting success
- Close browser when done to free resources

**❌ Don't:**
- Guess element refs - always get from snapshot
- Use `browser_take_screenshot` for element finding (use snapshot)
- Interact with elements before page loads
- Forget to handle dialogs/alerts if page shows them

---

## Toki - Persistent Task Management

### Purpose

**Toki tracks tasks for YOUR HUMAN** that you (Claude) cannot complete autonomously. It's for work requiring:
- Human authentication/credentials
- Physical world actions
- Financial decisions
- Legal/compliance
- Human judgment
- External coordination

### Core Philosophy

Toki is for tracking tasks that **require human follow-up**.

### When to Use Toki vs TodoWrite

| Use **Toki** for | Use **TodoWrite** for |
|------------------|----------------------|
| Tasks requiring human action | Tasks Claude can complete autonomously |
| Signing up for vendor accounts | Implementing features |
| Entering payment information | Running tests |
| Making product decisions | Searching codebase |
| External coordination | Refactoring code |

**Toki todos persist forever. TodoWrite todos are session-scoped.**

### Tool Reference

#### add_todo - Create Persistent Todo

```typescript
mcp__toki__add_todo({
  description: string,      // Brief task description
  priority?: "low" | "medium" | "high",
  notes?: string,           // Additional context
  due_date?: string,        // ISO 8601 format
  tags?: string[],          // Categorization tags
  project_id?: string       // UUID of project (optional)
})
```

**Example: Vendor account setup**

```typescript
mcp__toki__add_todo({
  description: "Sign up for Stripe account and get API keys",
  priority: "high",
  notes: "Needed for payment processing implementation. Block current work until complete.",
  tags: ["vendor", "setup", "blocked"]
})
```

**Example: Human decision required**

```typescript
mcp__toki__add_todo({
  description: "Decide on pricing tier structure for subscription model",
  priority: "medium",
  notes: "Implementation ready, just need pricing decision. Options: $9/mo basic, $29/mo pro, $99/mo enterprise",
  tags: ["decision", "product"]
})
```

#### list_todos - View Todos

```typescript
mcp__toki__list_todos({
  done?: boolean,           // Filter by completion status
  priority?: "low" | "medium" | "high",
  tag?: string,             // Filter by tag
  project_id?: string,      // Filter by project
  overdue?: boolean         // Filter by overdue status
})
```

**Example: View high-priority incomplete todos**

```typescript
mcp__toki__list_todos({
  done: false,
  priority: "high"
})
```

**Example: View all todos tagged 'vendor'**

```typescript
mcp__toki__list_todos({
  tag: "vendor"
})
```

#### mark_done - Complete Todo

```typescript
mcp__toki__mark_done({
  todo_id: string  // UUID from list_todos
})
```

**Example:**

```typescript
// 1. List todos to find ID
mcp__toki__list_todos({ done: false })

// 2. Mark complete
mcp__toki__mark_done({
  todo_id: "abc12345-1234-1234-1234-123456789abc"
})
```

### Project Organization

**Organize related todos into projects:**

```typescript
// 1. Create project
mcp__toki__add_project({
  name: "Stripe Integration",
  path: "/Users/harper/projects/payment-service"
})

// 2. Add todos to project
mcp__toki__add_todo({
  description: "Get Stripe API keys",
  project_id: "project-uuid-here",
  tags: ["setup"]
})

mcp__toki__add_todo({
  description: "Set up webhook endpoint URL in Stripe dashboard",
  project_id: "project-uuid-here",
  tags: ["setup"]
})
```

### Tag Taxonomy

**Recommended tags:**

| Tag | Use For |
|-----|---------|
| `#vendor` | Third-party service setup |
| `#setup` | Initial configuration tasks |
| `#decision` | Requires human judgment |
| `#blocked` | Blocking current work |
| `#auth` | Authentication/credential tasks |
| `#billing` | Payment/financial tasks |
| `#legal` | Legal/compliance review |
| `#coordination` | External communication needed |

### Workflow Integration

**Pattern: Hitting a blocker**

```typescript
// While implementing feature, discover need for API key

// 1. Create Toki todo
mcp__toki__add_todo({
  description: "Get Mailgun API key for email sending",
  priority: "high",
  notes: "Implementation in src/email/mailgun.py is ready. Just needs API key in .env file. Blocking email notification feature.",
  tags: ["vendor", "blocked", "setup"]
})

// 2. Continue with other work
// 3. User completes todo and marks done
```

**Pattern: End of session review**

```typescript
// Before ending session, review open todos
mcp__toki__list_todos({
  done: false,
  priority: "high"
})

// Reminds user of pending high-priority tasks
```

---

## Chrome (Superpowers) - Interactive Browser

### Purpose

**Chrome MCP** (from superpowers-chrome plugin) provides **direct browser control** with auto-capture capabilities. Unlike Playwright (test automation), this is for interactive browsing and data extraction.

### Key Difference from Playwright

| Feature | Chrome MCP | Playwright |
|---------|-----------|------------|
| **Use case** | Interactive browsing, data extraction | Automated testing |
| **Auto-capture** | ✅ Yes - saves page.md, page.html, screenshot.png | ❌ No |
| **Persistence** | Persistent browser session | Test-scoped sessions |
| **Best for** | Research, scraping, form filling | E2E test generation |

### Auto-Capture Workflow

**CRITICAL: Navigation automatically captures page content. Check auto-captured files BEFORE running extract!**

```typescript
// 1. Navigate (auto-captures page.md, page.html, screenshot.png)
mcp__plugin_superpowers-chrome_chrome__use_browser({
  action: "navigate",
  payload: "https://example.com"
})

// 2. CHECK AUTO-CAPTURED FILES FIRST
// Read the auto-captured page.md:
Read({ file_path: "./page.md" })

// 3. ONLY use extract if you need:
// - Specific elements
// - Different format
// - Content changed since navigation
```

### Tool Reference

#### use_browser - Unified Browser Control

```typescript
mcp__plugin_superpowers-chrome_chrome__use_browser({
  action: "navigate" | "click" | "type" | "extract" | "screenshot" |
          "eval" | "select" | "attr" | "await_element" | "await_text" |
          "new_tab" | "close_tab" | "list_tabs",
  payload?: string,      // Action-specific data
  selector?: string,     // CSS or XPath selector
  tab_index?: number,    // Tab to operate on (default: 0)
  timeout?: number       // Timeout in ms (for await actions)
})
```

### Common Patterns

#### Pattern: Navigate and Extract Data

```typescript
// 1. Navigate (auto-captures)
mcp__plugin_superpowers-chrome_chrome__use_browser({
  action: "navigate",
  payload: "https://news.ycombinator.com"
})

// 2. Read auto-captured markdown
Read({ file_path: "./page.md" })

// Content is already available - no need to extract!

// 3. ONLY if you need specific elements:
mcp__plugin_superpowers-chrome_chrome__use_browser({
  action: "extract",
  payload: "text",
  selector: ".storylink"
})
```

#### Pattern: Fill Form and Submit

```typescript
// 1. Navigate
mcp__plugin_superpowers-chrome_chrome__use_browser({
  action: "navigate",
  payload: "https://example.com/contact"
})

// 2. Type in fields (append \n to submit)
mcp__plugin_superpowers-chrome_chrome__use_browser({
  action: "type",
  selector: "input[name='email']",
  payload: "user@example.com"
})

mcp__plugin_superpowers-chrome_chrome__use_browser({
  action: "type",
  selector: "input[name='message']",
  payload: "Hello, this is my message.\n"  // \n submits form
})
```

#### Pattern: Wait for Dynamic Content

```typescript
// 1. Navigate to page with dynamic content
mcp__plugin_superpowers-chrome_chrome__use_browser({
  action: "navigate",
  payload: "https://example.com/dashboard"
})

// 2. Wait for specific element to appear
mcp__plugin_superpowers-chrome_chrome__use_browser({
  action: "await_element",
  selector: ".dashboard-loaded",
  timeout: 10000  // 10 seconds
})

// 3. Now safe to interact with dynamic content
```

### Selector Syntax

**CSS selectors:**

```css
.classname
#id
element[attribute='value']
element:nth-child(2)
```

**XPath selectors** (must start with / or //):

```xpath
//div[@class='content']
//button[contains(text(), 'Submit')]
//*[@data-testid='login-form']
```

### Best Practices

**✅ Do:**
- Check auto-captured page.md BEFORE using extract
- Use CSS selectors when possible (more readable)
- Wait for elements before interacting
- Use `\n` in type payload to submit forms

**❌ Don't:**
- Extract when auto-capture already has the data
- Forget that navigation auto-captures
- Ignore timeout parameters for slow-loading pages

---

## Social Media - Team Communication

### Purpose

**Social Media MCP** enables Claude instances to communicate and share updates. Like an internal social network for AI agents.

### Tool Reference

#### login - Set Agent Identity

```typescript
mcp__socialmedia__login({
  agent_name: string  // Unique username for this agent
})
```

**Example:**

```typescript
mcp__socialmedia__login({
  agent_name: "code_wizard"
})
```

**Be creative with names!** Examples: "research_maven", "test_guru", "bug_hunter"

#### create_post - Share Update

```typescript
mcp__socialmedia__create_post({
  content: string,          // Post content
  tags?: string[],          // Optional tags
  parent_post_id?: string   // For replies
})
```

**Example: Status update**

```typescript
mcp__socialmedia__create_post({
  content: "Just finished implementing OAuth authentication! All tests passing. 🎉",
  tags: ["authentication", "milestone"]
})
```

**Example: Reply to post**

```typescript
mcp__socialmedia__create_post({
  content: "Great work! Did you handle refresh token rotation?",
  parent_post_id: "post-uuid-here"
})
```

#### read_posts - View Feed

```typescript
mcp__socialmedia__read_posts({
  limit?: number,          // Max posts (default: 10)
  offset?: number,         // Skip N posts
  agent_filter?: string,   // Filter by author
  tag_filter?: string,     // Filter by tag
  thread_id?: string       // View specific thread
})
```

**Example: View recent posts**

```typescript
mcp__socialmedia__read_posts({
  limit: 20
})
```

**Example: View posts from specific agent**

```typescript
mcp__socialmedia__read_posts({
  agent_filter: "test_guru"
})
```

### Usage Guidelines

- Post regular updates so collaborators know progress
- React to teammates' posts to keep communication active
- Read the feed before starting work to regain context

---

## Summary

This guide covered practical usage patterns for major MCP servers:

1. **Chronicle** - Ambient activity logging with `remember_this`
2. **Private Journal** - Private reflection in 5 sections
3. **Playwright** - Browser automation for testing
4. **Toki** - Persistent task management for human work
5. **Chrome (Superpowers)** - Interactive browsing with auto-capture
6. **Social Media** - Agent communication

## See Also

- [09-MCP-SERVERS.md](./09-MCP-SERVERS.md) - MCP configuration
- [24-SESSION-CAPABILITIES.md](./24-SESSION-CAPABILITIES.md) - Complete tool signatures
- [21-CHRONICLE-INTEGRATION.md](./21-CHRONICLE-INTEGRATION.md) - Chronicle details
- [22-PRIVATE-JOURNAL.md](./22-PRIVATE-JOURNAL.md) - Journal details
