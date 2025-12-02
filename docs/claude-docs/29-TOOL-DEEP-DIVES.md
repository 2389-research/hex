# Tool Deep Dives

**ABOUTME: In-depth coverage of under-documented Claude Code tools including NotebookEdit, AskUserQuestion, ExitPlanMode**
**ABOUTME: Comprehensive usage patterns, edge cases, and practical examples for advanced tooling**

## Overview

While core tools (Read, Edit, Write, Bash, Grep, Glob) are well-documented in [03-TOOL-SYSTEM.md](./03-TOOL-SYSTEM.md), several advanced tools require deeper exploration. This document provides comprehensive guides for specialized tools that enable advanced workflows.

## NotebookEdit - Jupyter Notebook Manipulation

### Overview

**NotebookEdit** enables editing Jupyter notebooks (.ipynb files) cell by cell. Since notebooks are JSON structures containing code, markdown, and output, this tool provides cell-level granularity.

**Signature:**

```typescript
NotebookEdit({
  notebook_path: string,        // Absolute path to .ipynb file
  new_source: string,           // New content for the cell
  cell_id?: string,             // Cell ID to edit (optional)
  cell_type?: 'code' | 'markdown', // Cell type (required for insert mode)
  edit_mode?: 'replace' | 'insert' | 'delete' // Default: replace
})
```

### Cell Discovery

**Finding cell IDs:**

Notebooks use unique cell IDs for identification. Use **Read** tool to view notebook structure:

```bash
# Read notebook to see cell IDs
Read({ file_path: "/path/to/notebook.ipynb" })
```

**Example output:**

```json
{
  "cells": [
    {
      "cell_type": "markdown",
      "id": "a1b2c3d4",
      "source": ["# Data Analysis\n", "This notebook analyzes..."]
    },
    {
      "cell_type": "code",
      "id": "e5f6g7h8",
      "execution_count": 1,
      "source": ["import pandas as pd\n", "import numpy as np"],
      "outputs": []
    },
    {
      "cell_type": "code",
      "id": "i9j0k1l2",
      "execution_count": 2,
      "source": ["df = pd.read_csv('data.csv')\n", "df.head()"],
      "outputs": [...]
    }
  ]
}
```

**Extract cell IDs** from the `id` field of each cell.

### Replacing Cell Content

**Example: Update code cell**

```typescript
// Replace cell e5f6g7h8 with new imports
NotebookEdit({
  notebook_path: "/Users/harper/analysis.ipynb",
  cell_id: "e5f6g7h8",
  new_source: "import pandas as pd\nimport numpy as np\nimport matplotlib.pyplot as plt"
})
```

**Example: Update markdown cell**

```typescript
// Replace markdown cell a1b2c3d4
NotebookEdit({
  notebook_path: "/Users/harper/analysis.ipynb",
  cell_id: "a1b2c3d4",
  new_source: "# Updated Data Analysis\n\nThis notebook performs comprehensive analysis..."
})
```

### Inserting New Cells

**Insert cell after specific cell:**

```typescript
// Insert new code cell after e5f6g7h8
NotebookEdit({
  notebook_path: "/Users/harper/analysis.ipynb",
  cell_id: "e5f6g7h8",  // Insert AFTER this cell
  cell_type: "code",
  edit_mode: "insert",
  new_source: "# Load additional data\ndf2 = pd.read_csv('more_data.csv')"
})
```

**Insert cell at beginning:**

```typescript
// Insert at beginning (omit cell_id)
NotebookEdit({
  notebook_path: "/Users/harper/analysis.ipynb",
  cell_type: "markdown",
  edit_mode: "insert",
  new_source: "# Installation Instructions\n\nRun: pip install pandas numpy"
})
```

### Deleting Cells

```typescript
// Delete cell i9j0k1l2
NotebookEdit({
  notebook_path: "/Users/harper/analysis.ipynb",
  cell_id: "i9j0k1l2",
  edit_mode: "delete",
  new_source: ""  // Required but ignored for delete
})
```

### Handling Cell Outputs

**Important:** NotebookEdit modifies cell **source** only. Cell outputs remain unchanged until the notebook is re-executed.

**Workflow to update outputs:**

```bash
# 1. Edit cell source
NotebookEdit({ ... })

# 2. Execute notebook to regenerate outputs
jupyter nbconvert --execute --inplace analysis.ipynb

# Or use papermill for parameterized execution
papermill input.ipynb output.ipynb -p param1 value1
```

### Working with Multi-line Cells

**Source is an array of strings** in .ipynb format:

```json
{
  "source": [
    "# First line\n",
    "# Second line\n",
    "# Third line"
  ]
}
```

**When editing, provide full content as single string:**

```typescript
NotebookEdit({
  notebook_path: "/path/to/notebook.ipynb",
  cell_id: "abc123",
  new_source: "# First line\n# Second line\n# Third line"
})
```

The tool handles conversion to array format automatically.

### Common Patterns

#### Pattern: Add Cell Before Analysis

```typescript
// 1. Read notebook to find first analysis cell
const notebook = Read({ file_path: "analysis.ipynb" })

// 2. Find cell ID of first analysis cell (manually identify from output)
const firstAnalysisCell = "e5f6g7h8"

// 3. Insert setup cell before it
// (Insert happens AFTER specified cell, so insert before previous cell)
NotebookEdit({
  notebook_path: "analysis.ipynb",
  cell_id: "a1b2c3d4",  // Cell before first analysis
  cell_type: "code",
  edit_mode: "insert",
  new_source: "# Setup\nimport warnings\nwarnings.filterwarnings('ignore')"
})
```

#### Pattern: Update All Import Cells

```bash
# 1. Read notebook
Read({ file_path: "analysis.ipynb" })

# 2. Identify import cells from output
# 3. Update each import cell
NotebookEdit({ cell_id: "cell1", new_source: "import pandas as pd" })
NotebookEdit({ cell_id: "cell2", new_source: "from sklearn import ..." })
```

#### Pattern: Template Notebook Creation

```bash
# 1. Create empty notebook
echo '{"cells":[],"metadata":{},"nbformat":4,"nbformat_minor":4}' > template.ipynb

# 2. Add title cell
NotebookEdit({
  notebook_path: "template.ipynb",
  cell_type: "markdown",
  edit_mode: "insert",
  new_source: "# Analysis Template"
})

# 3. Add import cell
NotebookEdit({
  notebook_path: "template.ipynb",
  cell_type: "code",
  edit_mode: "insert",
  new_source: "import pandas as pd\nimport numpy as np"
})

# 4. Add analysis cell
NotebookEdit({
  notebook_path: "template.ipynb",
  cell_type: "code",
  edit_mode: "insert",
  new_source: "# Your analysis here"
})
```

### Troubleshooting

**Issue: Cell ID not found**

```
Error: Cell with id 'xyz123' not found
```

**Solution:** Re-read notebook to get current cell IDs (IDs may change after edits).

**Issue: Invalid JSON after edit**

```
Error: Invalid notebook format
```

**Solution:** Validate notebook structure:

```bash
jupyter nbconvert --to notebook --execute --inplace notebook.ipynb
```

**Issue: Outputs not updated**

**Solution:** Outputs aren't auto-regenerated. Re-execute notebook:

```bash
jupyter nbconvert --execute --inplace notebook.ipynb
```

## AskUserQuestion - Interactive Decision Making

### Overview

**AskUserQuestion** enables Claude to ask the user for input during execution. Useful for:
- Gathering preferences
- Clarifying ambiguous instructions
- Getting decisions on implementation choices
- Offering multiple approaches

**Signature:**

```typescript
AskUserQuestion({
  questions: Array<{
    question: string,           // Clear question ending with "?"
    header: string,             // Short label (max 12 chars)
    options: Array<{
      label: string,            // Option display text (1-5 words)
      description: string       // Explanation of this choice
    }>,
    multiSelect: boolean        // Allow multiple selections (default: false)
  }>
})
```

### Single Question Pattern

**Example: Choose authentication method**

```typescript
AskUserQuestion({
  questions: [{
    question: "Which authentication method should we implement?",
    header: "Auth Method",
    multiSelect: false,
    options: [
      {
        label: "OAuth 2.0",
        description: "Industry standard, supports third-party providers (Google, GitHub)"
      },
      {
        label: "JWT",
        description: "Stateless tokens, good for APIs and microservices"
      },
      {
        label: "Session-based",
        description: "Traditional cookies, simpler to implement"
      }
    ]
  }]
})
```

**User sees:**

```
┌─ Auth Method ────────────────────────────────────────────┐
│ Which authentication method should we implement?         │
│                                                           │
│ ○ OAuth 2.0                                              │
│   Industry standard, supports third-party providers      │
│                                                           │
│ ○ JWT                                                    │
│   Stateless tokens, good for APIs and microservices     │
│                                                           │
│ ○ Session-based                                          │
│   Traditional cookies, simpler to implement              │
│                                                           │
│ ○ Other (provide custom answer)                         │
└───────────────────────────────────────────────────────────┘
```

### Multi-Select Pattern

**Example: Choose features to implement**

```typescript
AskUserQuestion({
  questions: [{
    question: "Which features should we include in this release?",
    header: "Features",
    multiSelect: true,
    options: [
      {
        label: "User profiles",
        description: "Allow users to customize their profile information"
      },
      {
        label: "Dark mode",
        description: "Toggle between light and dark themes"
      },
      {
        label: "Export data",
        description: "Export user data to CSV/JSON formats"
      },
      {
        label: "Email notifications",
        description: "Send email alerts for important events"
      }
    ]
  }]
})
```

**User can select multiple options** (checkboxes instead of radio buttons).

### Multiple Questions Pattern

**Ask up to 4 questions simultaneously:**

```typescript
AskUserQuestion({
  questions: [
    {
      question: "Which database should we use?",
      header: "Database",
      multiSelect: false,
      options: [
        { label: "PostgreSQL", description: "Robust relational database" },
        { label: "MongoDB", description: "Flexible document store" },
        { label: "SQLite", description: "Lightweight, file-based" }
      ]
    },
    {
      question: "Which ORM should we use?",
      header: "ORM",
      multiSelect: false,
      options: [
        { label: "SQLAlchemy", description: "Full-featured Python ORM" },
        { label: "Peewee", description: "Lightweight and simple" },
        { label: "Raw SQL", description: "No ORM, direct queries" }
      ]
    },
    {
      question: "Enable caching?",
      header: "Caching",
      multiSelect: false,
      options: [
        { label: "Redis", description: "Fast in-memory cache" },
        { label: "Memcached", description: "Simple distributed cache" },
        { label: "No caching", description: "Skip caching for now" }
      ]
    }
  ]
})
```

### Progressive Disclosure Pattern

**Start broad, then get specific:**

```typescript
// First question - high level
AskUserQuestion({
  questions: [{
    question: "What type of application are we building?",
    header: "App Type",
    multiSelect: false,
    options: [
      { label: "Web API", description: "RESTful API backend" },
      { label: "Web App", description: "Full-stack web application" },
      { label: "CLI Tool", description: "Command-line utility" }
    ]
  }]
})

// Based on answer, ask follow-up
// If user chose "Web API":
AskUserQuestion({
  questions: [{
    question: "Which API framework?",
    header: "Framework",
    multiSelect: false,
    options: [
      { label: "FastAPI", description: "Modern, fast, async support" },
      { label: "Flask", description: "Lightweight, flexible" },
      { label: "Django REST", description: "Batteries included" }
    ]
  }]
})
```

### Best Practices

#### 1. Clear, Specific Questions

```typescript
// ❌ BAD - Vague question
question: "What should we do?"

// ✅ GOOD - Specific question
question: "Which testing framework should we use for this project?"
```

#### 2. Concise Headers

```typescript
// ❌ BAD - Too long
header: "Testing Framework Selection"  // 27 chars

// ✅ GOOD - Within limit
header: "Test Frmwrk"  // 11 chars
```

#### 3. Meaningful Descriptions

```typescript
// ❌ BAD - Redundant description
{
  label: "pytest",
  description: "pytest testing framework"
}

// ✅ GOOD - Adds value
{
  label: "pytest",
  description: "Python's most popular testing framework, supports fixtures and plugins"
}
```

#### 4. Appropriate Option Count

```typescript
// ✅ GOOD - 2-4 options (easy to compare)
options: [
  { label: "Option A", description: "..." },
  { label: "Option B", description: "..." },
  { label: "Option C", description: "..." }
]

// ⚠️ OK - 5-7 options (still manageable)
// ❌ BAD - 8+ options (overwhelming)
```

#### 5. Logical Grouping

```typescript
// If you have many related questions, group them logically
// Question 1: Architecture decisions
// Question 2: Technology choices
// Question 3: Deployment strategy
// Question 4: Monitoring approach
```

### When to Use AskUserQuestion

**✅ Use when:**
- Multiple valid approaches exist
- User preferences matter
- Requirements are ambiguous
- Trade-offs need user input

**❌ Don't use when:**
- Only one correct answer
- Question is purely technical (you should know)
- Disrupts workflow unnecessarily
- Answer is obvious from context

### Error Handling

**Accessing user's answer:**

```typescript
// The answers are provided in the `answers` parameter
AskUserQuestion({
  questions: [{
    question: "Choose deployment platform",
    header: "Platform",
    options: [...]
  }],
  answers: {
    "Platform": "AWS"  // User's selection
  }
})
```

**Handling "Other" responses:**

Users can always choose "Other" to provide custom text. Handle this:

```typescript
if (answer === "Other") {
  // User provided custom text instead of selecting an option
  // Proceed with their custom input
}
```

## ExitPlanMode - Plan Mode Workflow

### Overview

**ExitPlanMode** is used when Claude is in "plan mode" and has finished creating an implementation plan. It signals readiness to transition to execution.

**Important:** Only use this for **implementation planning tasks** requiring code. NOT for research, file reading, or codebase exploration.

**Signature:**

```typescript
ExitPlanMode({
  plan: string  // Concise plan with markdown formatting
})
```

### When to Use ExitPlanMode

**✅ Use for:**
- Planning implementation steps for a coding task
- Breaking down complex feature development
- Organizing refactoring work
- Sequencing architectural changes

**❌ Don't use for:**
- Research tasks (gathering information)
- Searching files or understanding codebase
- Reading documentation
- Exploratory analysis

### Plan Format

**Good plan structure:**

```markdown
## Implementation Plan

### 1. Database Schema Updates
- Add `users` table with columns: id, email, password_hash, created_at
- Create migration file `001_create_users.sql`
- Run migration against dev database

### 2. User Model Implementation
- Create `src/models/user.py` with User class
- Implement password hashing using bcrypt
- Add email validation

### 3. Authentication Endpoints
- POST `/api/register` - Create new user
- POST `/api/login` - Authenticate and return JWT
- GET `/api/profile` - Get current user (requires auth)

### 4. Testing
- Unit tests for User model (registration, validation)
- Integration tests for auth endpoints
- E2E test for full registration flow

### 5. Documentation
- Update API docs with new endpoints
- Add authentication guide to README
```

### Using ExitPlanMode with /superpowers:write-plan

**Typical workflow:**

1. User requests implementation planning: `/superpowers:write-plan`
2. Claude creates detailed plan
3. Claude uses ExitPlanMode to present plan and get approval
4. User approves
5. Execution begins (or user uses `/superpowers:execute-plan`)

**Example:**

```typescript
// After creating comprehensive plan
ExitPlanMode({
  plan: `
## User Authentication Implementation

### Phase 1: Database (30min)
- Create users table with SQLAlchemy model
- Add migration for schema changes
- Verify migration in dev environment

### Phase 2: Core Logic (45min)
- Implement password hashing (bcrypt)
- Create JWT token generation
- Add user registration logic

### Phase 3: API Endpoints (45min)
- POST /api/register endpoint
- POST /api/login endpoint
- Add authentication middleware

### Phase 4: Testing (60min)
- Unit tests for password hashing
- Integration tests for endpoints
- E2E test for registration flow

**Total Estimated Time: 3 hours**
**Dependencies: bcrypt, PyJWT, SQLAlchemy**
`
})
```

### Plan Mode Lifecycle

```
User asks to plan → Plan Mode Active
         ↓
Claude creates plan
         ↓
Claude calls ExitPlanMode(plan)
         ↓
User reviews plan
         ↓
User approves/requests changes
         ↓
Execute plan (manual or /superpowers:execute-plan)
```

### Handling Ambiguity in Plans

**Before using ExitPlanMode, ensure plan is unambiguous:**

```typescript
// ❌ BAD - Ambiguous plan
ExitPlanMode({
  plan: "1. Set up authentication\n2. Add user features\n3. Deploy"
})

// ✅ GOOD - Specific plan
ExitPlanMode({
  plan: `
1. Implement OAuth 2.0 with Google provider
   - Install authlib library
   - Configure OAuth credentials
   - Create /auth/google/login endpoint

2. Add user profile management
   - Display name, email, avatar
   - Edit profile endpoint
   - Profile page UI component

3. Deploy to Vercel staging
   - Set environment variables
   - Run build verification
   - Deploy to preview URL
`
})
```

**If requirements unclear, use AskUserQuestion BEFORE ExitPlanMode:**

```typescript
// 1. First, clarify approach
AskUserQuestion({
  questions: [{
    question: "Which authentication method should we implement?",
    header: "Auth Method",
    options: [
      { label: "OAuth 2.0", description: "..." },
      { label: "JWT", description: "..." },
      { label: "Session-based", description: "..." }
    ]
  }]
})

// 2. Based on answer, create specific plan
ExitPlanMode({
  plan: `... specific plan for chosen method ...`
})
```

### Common Mistakes

**❌ Using ExitPlanMode for research:**

```typescript
// WRONG - This is research, not implementation planning
ExitPlanMode({
  plan: "1. Search for vim mode implementation\n2. Read relevant files\n3. Document findings"
})

// Should just do the research directly without plan mode
```

**❌ Too vague:**

```typescript
// WRONG - Not actionable
ExitPlanMode({
  plan: "1. Fix bugs\n2. Add features\n3. Test"
})
```

**❌ Missing crucial details:**

```typescript
// WRONG - Which files? What changes exactly?
ExitPlanMode({
  plan: "1. Update authentication\n2. Change database stuff"
})
```

## ListMcpResourcesTool & ReadMcpResourceTool - MCP Resource Access

### Overview

MCP servers can provide **resources** - documents, data, or content accessible via URIs. These tools let you discover and read those resources.

**Use case:** Access documentation, templates, or data provided by MCP servers without direct tool calls.

### ListMcpResourcesTool

**Signature:**

```typescript
ListMcpResourcesTool({
  server?: string  // Optional: filter by specific server
})
```

**List all available resources:**

```typescript
ListMcpResourcesTool({})
```

**Example output:**

```json
[
  {
    "server": "playwright",
    "uri": "playwright://docs/getting-started",
    "name": "Getting Started Guide",
    "mimeType": "text/markdown",
    "description": "Introduction to Playwright automation"
  },
  {
    "server": "toki",
    "uri": "toki://templates/project",
    "name": "Project Template",
    "mimeType": "application/json",
    "description": "Default project structure template"
  }
]
```

**List resources from specific server:**

```typescript
ListMcpResourcesTool({ server: "playwright" })
```

### ReadMcpResourceTool

**Signature:**

```typescript
ReadMcpResourceTool({
  server: string,  // MCP server name
  uri: string      // Resource URI
})
```

**Read a specific resource:**

```typescript
ReadMcpResourceTool({
  server: "playwright",
  uri: "playwright://docs/getting-started"
})
```

**Example output:**

```markdown
# Getting Started with Playwright

Playwright enables reliable end-to-end testing...

## Installation

npm install @playwright/test

## Writing Your First Test

...
```

### Resource Discovery Workflow

```bash
# 1. Discover available resources
ListMcpResourcesTool({})

# 2. Identify relevant resource from output
# Example: Found "toki://templates/project"

# 3. Read specific resource
ReadMcpResourceTool({
  server: "toki",
  uri: "toki://templates/project"
})

# 4. Use content as needed
```

### Common Patterns

#### Pattern: Find and Use Documentation

```typescript
// 1. List resources to find docs
const resources = ListMcpResourcesTool({ server: "playwright" })

// 2. Read specific doc
const guide = ReadMcpResourceTool({
  server: "playwright",
  uri: "playwright://docs/selectors"
})

// 3. Use information from guide
// "Based on the selector guide, I'll use data-testid attributes..."
```

#### Pattern: Access Server-Provided Templates

```typescript
// 1. List available templates
const resources = ListMcpResourcesTool({ server: "toki" })

// 2. Read template
const template = ReadMcpResourceTool({
  server: "toki",
  uri: "toki://templates/bug-report"
})

// 3. Create todo from template
mcp__toki__add_todo({
  description: "Bug: Application crashes on startup",
  notes: template.content  // Use template structure
})
```

### When to Use vs Direct Tool Calls

**Use MCP Resource Tools when:**
- Accessing documentation/templates provided by server
- Exploring available resources from server
- Need static content (guides, schemas, etc.)

**Use direct MCP tool calls when:**
- Performing actions (create todo, log entry, etc.)
- Querying dynamic data (list todos, search journal)
- Triggering server functionality

**Example comparison:**

```typescript
// ✅ Resource tool - Reading static docs
ReadMcpResourceTool({
  server: "chronicle",
  uri: "chronicle://docs/tagging-guide"
})

// ✅ Direct tool - Querying dynamic data
mcp__chronicle__search_entries({
  text: "deployment",
  tags: ["production"]
})
```

### Troubleshooting

**Issue: Resource not found**

```
Error: Resource not found: <uri>
```

**Solution:** List resources first to verify URI:

```typescript
ListMcpResourcesTool({ server: "servername" })
```

**Issue: Server doesn't support resources**

```
Error: Server does not provide resources
```

**Solution:** Not all MCP servers provide resources. This is expected. Use direct tool calls instead.

## Summary

### Tool Selection Guide

| Task | Tool | When to Use |
|------|------|-------------|
| Edit Jupyter notebook cell | NotebookEdit | Working with .ipynb files, need cell-level edits |
| Ask user for preference | AskUserQuestion | Multiple valid approaches, need user input |
| Present implementation plan | ExitPlanMode | Finished planning code implementation, need approval |
| Discover MCP resources | ListMcpResourcesTool | Exploring what resources an MCP server provides |
| Read MCP resource | ReadMcpResourceTool | Accessing docs/templates from MCP server |

### Key Takeaways

1. **NotebookEdit** works with cell IDs - always Read notebook first to find IDs
2. **AskUserQuestion** supports 1-4 questions, 2-4 options each, with descriptions
3. **ExitPlanMode** only for implementation planning, not research tasks
4. **MCP Resource Tools** for static content, direct tools for dynamic operations

## See Also

- [03-TOOL-SYSTEM.md](./03-TOOL-SYSTEM.md) - Complete tool reference
- [02-FILE-OPERATIONS.md](./02-FILE-OPERATIONS.md) - Read, Edit, Write tools
- [28-PLAN-MODE-WORKFLOW.md](./28-PLAN-MODE-WORKFLOW.md) - Complete plan mode guide
- [24-SESSION-CAPABILITIES.md](./24-SESSION-CAPABILITIES.md) - All MCP tool signatures
