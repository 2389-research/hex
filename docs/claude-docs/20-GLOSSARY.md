# Glossary

Comprehensive reference of technical terms, acronyms, tools, and concepts used in Claude Code.

## A

**ABOUTME Comment**
```
Two-line file header explaining purpose:
  # ABOUTME: User authentication endpoints
  # ABOUTME: Handles login, logout, token refresh via JWT

Makes files greppable for purpose: grep "ABOUTME:" **/*.py
```

**Agent**
```
Claude Code instance executing tasks.
Can spawn subagents for isolated work.
```

**API (Application Programming Interface)**
```
Interface for programs to communicate.
Common types:
  - REST API: HTTP-based, JSON responses
  - GraphQL: Query language for APIs
  - RPC: Remote procedure calls
```

**AST (Abstract Syntax Tree)**
```
Tree representation of code structure.
Used by ast-grep (sg) for code analysis.

Example Python AST:
  def add(a, b):
      return a + b

  AST:
    FunctionDef(name='add')
      arguments: [a, b]
      body: Return(a + b)
```

**Atomic Commit**
```
Single-purpose commit that:
  - Does one thing
  - Passes all tests
  - Can be reverted independently

Example: "feat: add login endpoint"
NOT: "Add login, fix bug, update deps"
```

**Autonomous Action** 🟢
```
Low-risk operation that proceeds without asking.
Examples: Fix linting, run tests, read files
```

## B

**Bash**
```
Unix shell and command language.
Claude Code tool for running terminal commands.

Usage: Terminal operations only
NOT for: File operations (use Read/Write/Edit)
```

**Binary Search (Debugging)**
```
Debugging technique:
1. Comment out half the code
2. Does bug still occur?
   - YES → Bug in active half
   - NO → Bug in commented half
3. Repeat until found
```

**Breaking Change**
```
Code change that breaks existing functionality.
Examples:
  - Changing function signature
  - Removing API endpoint
  - Changing data format

Require:
  - Major version bump
  - Migration guide
  - User notification
```

## C

**CI/CD (Continuous Integration/Continuous Deployment)**
```
Automated build and deployment pipeline.

CI: Automatically test code on commit
CD: Automatically deploy passing code

Tools: GitHub Actions, GitLab CI, Jenkins
```

**CLI (Command-Line Interface)**
```
Text-based interface for programs.
Examples: git, uv, pytest, claude
```

**Team policies**
```
Curated instructions that customize Claude Code's behavior.

Layers:
  - Organization-wide defaults
  - Project-specific overrides

Priority: Project > Global > Defaults
```

**Co-Authored-By**
```
Git commit trailer showing collaboration.

Example:
  Co-Authored-By: Claude <noreply@anthropic.com>

Added to all Claude Code commits.
```

**Collaborative Action** 🟡
```
Medium-risk operation requiring proposal first.
Examples: Multi-file changes, new features
```

**Context Window**
```
Total information Claude Code can consider.

Includes:
  - System instructions
  - Team policy docs
  - Conversation history
  - File contents
  - Tool outputs

Limit: 200,000 tokens
```

**Conventional Commits**
```
Commit message format:

  <type>: <summary>

  [optional body]

Types: feat, fix, refactor, test, docs, chore, style, perf
```

## D

**Defense in Depth**
```
Multiple layers of security:
1. User approval (permission system)
2. Sandboxing (isolation)
3. Input validation (sanitization)
4. Pre-commit hooks (quality checks)
5. Code review (human verification)
```

**Dependency**
```
External package/library used by project.

Types:
  - Production: Required to run (requests, fastapi)
  - Development: Required to develop (pytest, ruff)

Managed by: uv in pyproject.toml
```

**Dependency Graph**
```
Map of module dependencies.

Example:
  A imports B, C
  B imports D
  C imports D
  D imports nothing

  Graph: A → B → D
         A → C → D
```

**Diff**
```
Comparison showing changes between files/commits.

Example:
  - old line (removed)
  + new line (added)
    unchanged line
```

## E

**Edit Tool**
```
Claude Code tool for surgical file modifications.

Usage: Change specific code sections
Requires: Exact old_string match
Must: Read file first
```

**End-to-End Test (E2E)**
```
Test simulating complete user workflow.

Example:
  1. User visits site
  2. User logs in
  3. User creates order
  4. User checks out
  5. User receives confirmation

Tests entire system together.
```

**Environment Variable**
```
System variable affecting program behavior.

Examples:
  DATABASE_URL=postgresql://localhost/db
  API_KEY=secret-key-here
  DEBUG=true

Storage: .env file (gitignored)
```

## F

**Fast-Forward Merge**
```
Git merge that moves branch pointer forward.

Before:  main: A → B
         feat:      → C → D

After:   main: A → B → C → D

No merge commit created.
Clean linear history.
```

**Fixture (Testing)**
```
Setup code for tests.

Example:
  @pytest.fixture
  def database():
      db = create_test_database()
      yield db
      db.cleanup()

  def test_user_creation(database):
      user = database.create_user("alice")
      assert user.id is not None
```

**Forbidden Flags**
```
Git flags that MUST NOT be used:

  --no-verify     (bypass hooks)
  --no-hooks      (bypass hooks)
  --no-pre-commit-hook (bypass hooks)

Why: Quality gates exist for safety
Action: Fix issues, don't bypass
```

## G

**Glob**
```
Pattern matching for file names.

Patterns:
  *.py         - All .py files in current dir
  **/*.py      - All .py files recursively
  src/**/*.js  - All .js in src/ tree
  test_*.py    - Files starting with test_

Claude Code tool: Glob(pattern="**/*.py")
```

**Grep**
```
Search tool for finding text in files.

Claude Code tool:
  Grep(pattern="def login", path="src/")

Supports:
  - Regex patterns
  - File filtering
  - Line numbers
  - Context lines (-A, -B, -C)
```

**Git**
```
Version control system.

Core concepts:
  - Commit: Snapshot of code
  - Branch: Parallel development
  - Merge: Combine branches
  - Remote: Server copy
  - Clone: Copy repository
```

## H

**HEAD**
```
Pointer to current commit.

  HEAD      - Current commit
  HEAD~1    - Previous commit
  HEAD~2    - Two commits ago
  HEAD^^    - Two commits ago (alternative)
```

**HEREDOC (Here Document)**
```
Multi-line string in bash.

Syntax:
  command <<'EOF'
  line 1
  line 2
  EOF

Claude Code uses for commit messages:
  git commit -m "$(cat <<'EOF'
  feat: add feature

  Details here

  Co-Authored-By: Claude
  EOF
  )"
```

**Hook (Git)**
```
Script that runs automatically on git events.

Pre-commit: Before commit is created
  - Run linters
  - Run formatters
  - Run tests
  - Check for secrets

Pre-push: Before push to remote
  - Run full test suite
  - Check coverage

Commit-msg: Validate commit message
  - Check format
  - Check length
```

## I

**Integration Test**
```
Test of multiple components together.

Example:
  def test_user_registration_flow():
      # Tests: API + Database + Email
      response = api.post("/register", data={...})
      user = db.get_user("alice")
      assert email.was_sent_to("alice@example.com")

Between unit and E2E in scope.
```

**Interactive Mode**
```
CLI mode requiring user input.

Examples:
  git rebase -i    (interactive rebase)
  git add -i       (interactive staging)
  vim              (text editor)

⚠️ Claude Code CANNOT use interactive tools
   (no keyboard input capability)
```

## J

**JJ (Jujutsu)**
```
Version control system (git alternative).

Modern approach to version control.
```

**JSON (JavaScript Object Notation)**
```
Data format for structured information.

Example:
  {
    "name": "Alice",
    "email": "alice@example.com",
    "age": 30
  }

Used in: APIs, config files, package manifests
```

**JWT (JSON Web Token)**
```
Authentication token format.

Structure: header.payload.signature

Example:
  eyJhbGc...  (header)
  eyJzdWI...  (payload: user ID, expiry)
  SflKxwR...  (signature: verification)

Stateless authentication.
```

## L

**Linter**
```
Tool that analyzes code for errors/style issues.

Python: ruff, pylint, flake8
JavaScript: eslint, biome
TypeScript: tslint, biome

Runs in pre-commit hooks.
```

**Lock File**
```
File specifying exact dependency versions.

Python: uv.lock
Node.js: package-lock.json, yarn.lock

Ensures reproducible builds:
  - Same versions on all machines
  - Same versions in CI/CD
  - Prevents unexpected updates
```

## M

**Main Branch**
```
Primary branch in git repository.

Alternatives: master, trunk, develop

Protected in team settings.
```

**MCP (Model Context Protocol)**
```
Protocol for extending Claude Code with external tools.

MCP Servers provide:
  - Web browsing (playwright, chrome)
  - Private journal
  - Social media
  - Time tracking (chronicle)
  - Custom integrations

Each server = isolated functionality.
```

**Merge Conflict**
```
Git conflict when same code changed in both branches.

Markers:
  <<<<<<< HEAD
  our changes
  =======
  their changes
  >>>>>>> feature

Resolution:
  1. Choose one version, or
  2. Combine both, or
  3. Write new solution
  4. Remove markers
  5. git add file
  6. git merge --continue
```

**Mock (Testing)**
```
Fake object replacing real dependency in tests.
Why: Mocks test mock behavior, not real code.
Use: Real dependencies in tests.
```

## N

**npm (Node Package Manager)**
```
Package manager for JavaScript.

Commands:
  npm install    - Install dependencies
  npm test       - Run tests
  npm run build  - Build project

Alternative: yarn, pnpm
Python equivalent: uv, pip
```

## O

**OAuth (Open Authorization)**
```
Authorization framework for third-party access.

Flow:
  1. User clicks "Login with Google"
  2. Redirected to Google
  3. User approves
  4. App receives token
  5. App accesses user data

No password sharing.
```

## P

**Package Manager**
```
Tool for installing dependencies.

Python: uv (preferred), pip, poetry
JavaScript: npm, yarn, pnpm
Ruby: gem, bundler
Rust: cargo

Claude Code uses: uv for Python
```

**Permission System**
```
Claude Code's risk-based approval system.

Automatic: Low-risk operations
Prompt: High-risk operations

User controls all risky actions.
```

**Pre-Commit Hook**
```
See Hook (Git).

Runs before commit created:
  - Format code
  - Lint code
  - Run tests
  - Check secrets

MUST NOT bypass (--no-verify forbidden).
```

**Pull Request (PR)**
```
Proposal to merge changes.

Contains:
  - Summary of changes
  - Test plan
  - Screenshots (if UI)
  - Breaking changes
  - Deployment notes

Workflow:
  1. Create feature branch
  2. Commit changes
  3. Push branch
  4. Open PR
  5. Code review
  6. Address feedback
  7. Merge
```

**PyPI (Python Package Index)**
```
Repository of Python packages.

URL: https://pypi.org

Usage:
  uv add requests  # Installs from PyPI
```

## Q

**Quality Gates**
```
Automated checks ensuring code quality.

Gates:
  - Tests pass
  - Linting passes
  - Type checking passes
  - Coverage threshold met
  - Security scan clean

NEVER bypass quality gates.
```

## R

**Read Tool**
```
Claude Code tool for reading files.

Usage: Read(file_path="/path/to/file.py")

Features:
  - Offset/limit for large files
  - Images (PNG, JPG)
  - PDFs
  - Jupyter notebooks

MUST use before Edit/Write on existing files.
```

**Refactoring**
```
Restructuring code without changing behavior.

Examples:
  - Extract function
  - Rename variable
  - Simplify logic
  - Remove duplication

Rules:
  1. Tests pass before refactoring
  2. Refactor
  3. Tests still pass
  4. Commit
```

**Regression**
```
Bug in previously working code.

Caused by: Recent changes
Prevention: Tests, code review

Regression test:
  Test that proves bug existed and is now fixed.
```

**Remote (Git)**
```
Server copy of repository.

Common remotes:
  origin - Primary remote (usually GitHub)
  upstream - Original repo (for forks)

Commands:
  git remote -v        - List remotes
  git remote add name url  - Add remote
  git push origin main - Push to remote
  git pull origin main - Pull from remote
```

**REPL (Read-Eval-Print Loop)**
```
Interactive programming environment.

Examples:
  python           # Python REPL
  node             # Node.js REPL
  irb              # Ruby REPL

⚠️ Claude Code cannot use REPLs
   (interactive input not supported)
```

**Repository (Repo)**
```
Project tracked by version control.

Contains:
  - Source code
  - Git history
  - Configuration files
  - Documentation

Locations:
  - Local: On your machine
  - Remote: On GitHub/GitLab
```

**REST (Representational State Transfer)**
```
API architectural style.

HTTP Methods:
  GET /users       - List users
  GET /users/123   - Get user
  POST /users      - Create user
  PUT /users/123   - Update user
  DELETE /users/123 - Delete user

Stateless, resource-based.
```

## S

**Sandbox**
```
Isolated environment for safe execution.

With sandbox:
  ✅ Limited file access
  ✅ Controlled network
  ❌ Can't affect host system

Without sandbox:
  ✅ Full system access
  ⚠️ Can modify anything

Control: dangerouslyDisableSandbox parameter
```

**Secret**
```
Sensitive credential that must be protected.

Examples:
  - API keys
  - Passwords
  - Private keys
  - OAuth tokens
  - Database URLs with passwords

Storage:
  ✅ .env file (gitignored)
  ✅ Environment variables
  ✅ Secret management service
  ❌ NEVER in git
```

**Skill**
```
Specialized workflow in Claude Code.

Examples:
  - test-driven-development
  - systematic-debugging
  - subagent-driven-development
  - brainstorming
  - code-review

Usage: Skill(command="test-driven-development")
```

**Slash Command**
```
Custom command in .claude/commands/

Example:
  /review-pr 123   - Review pull request
  /deploy staging  - Deploy to staging

Expands to full prompt for Claude Code.
```

**Staging Area (Git)**
```
Files prepared for next commit.

Commands:
  git add file.py      - Stage file
  git add .            - Stage all changes
  git reset file.py    - Unstage file
  git diff --staged    - See staged changes
```

**Subagent**
```
Fresh Claude Code instance for isolated work.

Benefits:
  - Clean context (0 tokens)
  - Independent execution
  - Can run in parallel
  - Fresh perspective

Usage: superpowers:subagent-driven-development skill
```

## T

**TDD (Test-Driven Development)**
```
Development methodology:

1. RED: Write failing test
2. GREEN: Write minimal code to pass
3. REFACTOR: Improve code quality

Benefits:
  - Tests prove code works
  - Tests describe behavior
  - High test coverage
  - Better design
```

**Technical Debt**
```
Future cost of shortcuts taken now.

Examples:
  - Skipped tests
  - Quick hacks
  - Duplicated code
  - Poor documentation

Result: Harder to maintain/extend code

Prevention: Do it right the first time.
```

**Terminal**
```
Text-based interface to system.

Also called: Shell, console, command line

Common shells:
  - bash (Bourne Again Shell)
  - zsh (Z Shell)
  - fish (Friendly Interactive Shell)
```

**Token**
```
Unit of text for language models.

~1 token = ~4 characters
~1 token = ~0.75 words

Claude Code budget: 200,000 tokens/conversation

Example:
  "Hello world" = ~2 tokens
  100-line file = ~2,000-3,000 tokens
```

**Tool**
```
Function Claude Code can call.

Core tools:
  - Bash: Run commands
  - Read: Read files
  - Write: Create files
  - Edit: Modify files
  - Grep: Search content
  - Glob: Find files
  - TodoWrite: Track tasks
  - Skill: Execute workflows

MCP tools: browser, journal, etc.
```

**Traffic Light System** 🟢🟡🔴
```
Decision framework for actions:

🟢 Green: Proceed automatically
   (Low risk, reversible)

🟡 Yellow: Propose first, then proceed
   (Medium risk, multiple files)

🔴 Red: Ask permission before proceeding
   (High risk, security, data loss)
```

**Type System**
```
System for specifying data types.

Python (optional):
  def add(a: int, b: int) -> int:
      return a + b

TypeScript (required):
  function add(a: number, b: number): number {
      return a + b;
  }

Benefits:
  - Catch errors early
  - Better IDE support
  - Self-documenting code
```

## U

**Unit Test**
```
Test of single function/class in isolation.

Example:
  def test_add():
      assert add(2, 3) == 5

Characteristics:
  - Fast (<1ms)
  - No external dependencies
  - Tests one thing
```

**UV**
```
Fast Python package manager.

Replacement for: pip, poetry, virtualenv, pyenv

Commands:
  uv init          - Initialize project
  uv add package   - Add dependency
  uv add --dev pkg - Add dev dependency
  uv sync          - Install from lock file
  uv run cmd       - Run in project env
  uv lock          - Update lock file

10-100× faster than pip.
```

**UUID (Universally Unique Identifier)**
```
128-bit unique identifier.

Format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
Example: 123e4567-e89b-12d3-a456-426614174000

Used for: Database IDs, session tokens, etc.

Practically unique (collision ~impossible).
```

## V

**Version Control**
```
System for tracking code changes over time.

Systems:
  - Git (most popular)
  - JJ/Jujutsu (modern alternative)
  - Mercurial, SVN (older)

Benefits:
  - History of all changes
  - Collaboration
  - Branching/merging
  - Recovery from mistakes
```

**Virtual Environment (venv)**
```
Isolated Python environment.

Creation:
  uv venv          # Claude Code way
  python -m venv   # Old way

Benefits:
  - Project-specific dependencies
  - No conflicts between projects
  - Reproducible environment

Location: .venv/ directory (gitignored)
```

## W

**Worktree (Git)**
```
Additional working directory for same repo.

Usage:
  git worktree add ../repo-feature feature-branch

Benefits:
  - Multiple branches checked out simultaneously
  - No stashing needed to switch
  - Isolated work environments

Cleanup:
  git worktree remove ../repo-feature
```

**Write Tool**
```
Claude Code tool for creating/replacing files.

Usage: Write(file_path="...", content="...")

Rules:
  - MUST Read existing files first
  - Prefer Edit for partial changes
  - Use for new files or complete rewrites

Overwrites existing content.
```

## X

**XML (eXtensible Markup Language)**
```
Data format using tags.

Example:
  <user>
    <name>Alice</name>
    <email>alice@example.com</email>
  </user>

Alternative: JSON (more common in APIs)
Used in: Config files, SOAP APIs, RSS
```

## Y

**YAML (YAML Ain't Markup Language)**
```
Human-readable data format.

Example:
  user:
    name: Alice
    email: alice@example.com
    roles:
      - admin
      - user

Used in:
  - Config files (.pre-commit-config.yaml)
  - CI/CD (GitHub Actions)
  - Docker Compose
```

## Z

**Zero-Day**
```
Security vulnerability unknown to vendor.

Timeline:
  Day 0: Vulnerability discovered
  Day 1+: Vendor creates patch
  Day X: Patch deployed

Risk highest at Day 0 (no patch available).
```

---

## File Locations

### Global Configuration

```
~/.claude/
  guidance.md            - Global instructions
  docs/                  - Personal documentation
    python.md
    source-control.md
    using-uv.md
    docker-uv.md
  commands/              - Global slash commands
  skills/                - Global skills
```

### Project Configuration

```
.claude/
  guidance.md            - Project instructions
  commands/              - Project slash commands
  skills/                - Project skills
```

### Git Configuration

```
.git/                    - Git repository data
.gitignore               - Files to ignore
.pre-commit-config.yaml  - Pre-commit hooks
```

### Python Project

```
pyproject.toml           - Project metadata, dependencies
uv.lock                  - Dependency lock file
.venv/                   - Virtual environment (gitignored)
src/                     - Source code
tests/                   - Test files
```

### Documentation

```
README.md                - Project overview
CHANGELOG.md             - Version history
API.md                   - API documentation
ARCHITECTURE.md          - System design
LICENSE                  - License terms
```

---

## Common Acronyms Quick Reference

```
API     Application Programming Interface
AST     Abstract Syntax Tree
CI/CD   Continuous Integration/Continuous Deployment
CLI     Command-Line Interface
E2E     End-to-End
HTTP    HyperText Transfer Protocol
JSON    JavaScript Object Notation
JWT     JSON Web Token
MCP     Model Context Protocol
OAuth   Open Authorization
PR      Pull Request
REPL    Read-Eval-Print Loop
REST    Representational State Transfer
TDD     Test-Driven Development
URL     Uniform Resource Locator
UUID    Universally Unique Identifier
XML     eXtensible Markup Language
YAML    YAML Ain't Markup Language
```

---

This glossary provides quick reference for all technical terms encountered when working with Claude Code. For deeper explanations, see the relevant documentation sections.
