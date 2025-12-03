---
name: document
description: Generate comprehensive documentation for code
args:
  target: What to document - function, file, module, or API (optional)
  type: Documentation type - "api", "guide", "reference" (optional)
---

# Documentation Generation

You are creating documentation {{if .target}}for {{.target}}{{else}}for the code{{end}}{{if .type}} ({{.type}} documentation){{end}}.

## Documentation Principles

### Write for Your Audience

**Code Comments**: For developers reading the code
**API Docs**: For users of your library/API
**User Guides**: For end users of the application
**Architecture Docs**: For team members and future maintainers

### Be Clear and Concise

- Use simple language
- Avoid jargon (or explain it)
- Be specific and concrete
- Include examples
- Keep it up to date

## Documentation Types

### 1. Code Comments ({{if eq .type "code"}}← FOCUS{{end}})

**File-level comments:**
```go
// ABOUTME: User authentication and session management
// ABOUTME: Handles login, logout, and session validation using JWT tokens

package auth
```

**Function comments (GoDoc style):**
```go
// CalculateTotal computes the total price including tax and discounts.
// It takes a list of items and returns the final amount to charge.
//
// The calculation applies discounts first, then adds tax on the discounted amount.
//
// Example:
//     items := []Item{{Price: 10.00, Qty: 2}, {Price: 5.00, Qty: 1}}
//     total := CalculateTotal(items)  // Returns 25.00
//
// Returns an error if any item has invalid price or quantity.
func CalculateTotal(items []Item) (float64, error) {
    // implementation
}
```

**Inline comments:**
```go
// Only use inline comments for non-obvious code
// Prefer self-documenting code over comments

// Calculate tax using jurisdiction-specific rate (varies by state)
taxRate := getTaxRate(user.state)
```

### 2. API Documentation ({{if eq .type "api"}}← FOCUS{{end}})

**REST API:**
```markdown
## Authentication API

### POST /auth/login

Authenticate a user and return a session token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiresAt": "2024-12-25T10:30:00Z",
  "user": {
    "id": "123",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid email or password format
- `401 Unauthorized`: Invalid credentials
- `429 Too Many Requests`: Rate limit exceeded

**Example:**
```bash
curl -X POST https://api.example.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'
```
```

**Library API:**
```markdown
## Package: auth

### func ValidateToken(token string) (*User, error)

Validates a JWT token and returns the associated user.

**Parameters:**
- `token` (string): JWT token to validate

**Returns:**
- `*User`: User associated with the token
- `error`: Error if token is invalid or expired

**Example:**
```go
user, err := auth.ValidateToken(token)
if err != nil {
    return fmt.Errorf("invalid token: %w", err)
}
fmt.Printf("User: %s\n", user.Email)
```

**Errors:**
- `ErrInvalidToken`: Token format is invalid
- `ErrExpiredToken`: Token has expired
- `ErrUserNotFound`: User no longer exists
```

### 3. User Guides ({{if eq .type "guide"}}← FOCUS{{end}})

**Getting Started:**
```markdown
# Getting Started with Clem

## Installation

### macOS
```bash
brew install clem
```

### Linux
```bash
curl -fsSL https://get.hex.dev | sh
```

## Quick Start

1. Set up your API key:
```bash
export ANTHROPIC_API_KEY=your-key-here
```

2. Start a conversation:
```bash
hex "Help me write a function to parse CSV files"
```

3. Resume a previous session:
```bash
hex --resume abc123
```

## Common Tasks

### Running Tests
```bash
hex "Run the test suite and fix any failures"
```

### Code Review
```bash
hex "Review the changes in git and provide feedback"
```
```

### 4. Architecture Documentation ({{if eq .type "architecture"}}← FOCUS{{end}})

```markdown
# System Architecture

## Overview
Hex is a Claude-powered CLI tool with three main components:
- CLI Interface (Cobra + BubbleTea)
- Core Engine (API client + tool system)
- Persistence Layer (SQLite)

## Component Diagram
```
┌─────────────┐
│   CLI/TUI   │ (Cobra, BubbleTea)
└──────┬──────┘
       │
┌──────▼──────┐
│  Core       │ (Client, Session, Tools)
│  Engine     │
└──────┬──────┘
       │
┌──────▼──────┐
│  Storage    │ (SQLite, Repositories)
└─────────────┘
```

## Key Design Decisions

### Tool System
Tools are modular and implement the `Tool` interface:
- `Execute(ctx, params)`: Run tool logic
- `RequiresApproval(params)`: Permission check

This allows easy addition of new tools.

### Session Persistence
Sessions are stored in SQLite with:
- Full conversation history
- Tool execution results
- Context management

This enables resuming sessions and conversation history.
```

## Documentation Checklist

For **Code Comments**:
- [ ] File has ABOUTME header
- [ ] Public functions documented
- [ ] Complex logic explained
- [ ] Examples provided
- [ ] Parameters and returns described

For **API Documentation**:
- [ ] All endpoints documented
- [ ] Request/response formats shown
- [ ] Error codes explained
- [ ] Examples provided
- [ ] Authentication described

For **User Guides**:
- [ ] Installation instructions
- [ ] Quick start tutorial
- [ ] Common tasks covered
- [ ] Troubleshooting section
- [ ] Examples are runnable

For **Architecture Docs**:
- [ ] System overview
- [ ] Component relationships
- [ ] Key decisions documented
- [ ] Diagrams included
- [ ] Trade-offs explained

## Output Format

Generate documentation in this structure:

```markdown
# [Title]

## Overview
[What this is and what it does]

## [Section 1]
[Content with examples]

### Example
```[language]
[code example]
```

**Output:**
```
[expected output]
```

## [Section 2]
...

## Common Patterns
[Frequent use cases]

## Troubleshooting
[Common issues and solutions]

## See Also
- [Related docs]
```

## Documentation Best Practices

1. **Show, Don't Just Tell**
   - Include working examples
   - Show expected output
   - Provide error examples

2. **Keep It Current**
   - Update docs when code changes
   - Remove outdated information
   - Version documentation

3. **Be Complete But Concise**
   - Cover all important cases
   - Don't repeat information
   - Link to related docs

4. **Test Your Examples**
   - Run all code examples
   - Verify commands work
   - Check links aren't broken

5. **Consider Your Reader**
   - What do they already know?
   - What are they trying to do?
   - What might confuse them?

## Special Sections

### README.md
Should include:
- Project description
- Installation
- Quick start
- Examples
- Contributing guide
- License

### CHANGELOG.md
Format:
```markdown
# Changelog

## [1.2.0] - 2024-12-01

### Added
- New feature X

### Changed
- Updated behavior of Y

### Fixed
- Bug in Z

### Deprecated
- Function A (use B instead)
```

### CONTRIBUTING.md
Should cover:
- How to set up development environment
- How to run tests
- Code style requirements
- PR process
- Where to ask questions

Begin generating documentation now.
