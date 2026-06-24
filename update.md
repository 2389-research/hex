# Spec: Hex Skill System v2

## Summary

Add a first-class skill runtime to Hex that is compatible with Claude Code / Codex-style skill systems while preserving Hex’s existing Markdown skill support. Hex already has an `internal/skills` package with a loader, registry, skill parser, and `Skill` tool adapter, but the current system behaves mostly like “load this Markdown snippet into context.” The new system should treat skills as reusable, discoverable, invokable, policy-aware workflow packages.

The target model:

```text
AGENTS.md / CLAUDE.md / project memory = ambient repository instructions
skills/*/SKILL.md                   = invokable workflow capabilities
tools                               = executable world-touching actions
mux                                 = orchestration loop
hex                                 = product/runtime/UI/persistence shell
```

## Background

Hex already has a skill loader. It scans builtin, user, plugin, and project directories, with later sources overriding earlier ones, and it sorts loaded skills by priority.

The existing `Skill` model supports YAML frontmatter fields for `name`, `description`, `tags`, `activationPatterns`, `model`, `priority`, `dependencies`, `version`, plus markdown content and source metadata.

Hex already exposes skills as a safe, no-approval `Skill` tool that loads a named skill and returns formatted skill content to the model.  The registry supports name lookup, listing, tag search, pattern matching, and text search.

Claude Code skills use `SKILL.md` files with YAML frontmatter plus Markdown instructions; Claude can load them automatically when relevant or users can invoke them directly, and skill content is loaded on demand rather than always injected like `CLAUDE.md`. ([Claude API Docs][1]) Codex describes agent skills as directories with `SKILL.md` plus optional scripts, references, assets, and agent metadata; it uses progressive disclosure by initially exposing only skill name, description, and file path, then loading the full `SKILL.md` only when the skill is selected. ([OpenAI Developers][2])

## Goals

1. Support Claude/Codex-style directory skills with `SKILL.md`.
2. Keep current single-file `.md` Hex skills working.
3. Make skills available to the mux runner, not just legacy/non-mux paths.
4. Add explicit slash-command invocation: `/skill-name [args...]`.
5. Add automatic skill selection via descriptions and activation patterns.
6. Add progressive disclosure so the model initially sees compact skill cards, not full skill bodies.
7. Add skill-scoped tool policy: allowed tools, disallowed tools, command allowlists, model/effort hints.
8. Add skill management commands: list, show, search, doctor, create.
9. Add trust and security boundaries for skills with scripts or executable content.
10. Preserve Hex’s existing plugin and project skill loading patterns.

## Non-goals

Do not build a package manager in the first implementation. Do not execute arbitrary skill scripts automatically. Do not make skills depend on Anthropic or OpenAI APIs; compatibility should be format-level and behavior-level, not provider-specific. Do not move the full skill runtime into `mux` yet; first make it solid in Hex, then extract generic pieces later.

## Desired user experience

### List skills

```bash
hex skills list
```

Example:

```text
NAME                            SOURCE    PRIORITY  INVOCATION  TAGS
verification-before-completion  builtin   100       auto,user   verification,quality
go-test-debugging               project   80        auto,user   go,testing,debugging
summarize-changes               user      50        user        git,review
```

### Show a skill

```bash
hex skills show go-test-debugging
```

Output should include:

```text
Name
Description
Source
Path
Priority
Tags
Activation patterns
Allowed tools
Disallowed tools
Model/effort hints
Trust state
Content preview
```

### Invoke a skill directly

```bash
hex /go-test-debugging "fix the failing mux orchestrator test"
```

Inside interactive mode:

```text
/go-test-debugging failing TestRunContinuesAfterToolUse
```

### Automatic activation

User:

```text
go test ./orchestrator is failing; fix it
```

Hex should inject a compact skill card:

```markdown
## Relevant Skills

- go-test-debugging: Diagnose and fix failing Go tests. Use Skill(command="go-test-debugging") before editing.
- verification-before-completion: Verify changes before claiming completion. Use before final response.
```

The model can then call:

```json
{"command": "go-test-debugging"}
```

to load the full instructions.

## Skill file formats

### Format A: existing single-file Hex skill

Keep this supported:

```text
.hex/skills/verification-before-completion.md
~/.hex/skills/go-test-debugging.md
```

Example:

```markdown
---
name: verification-before-completion
description: Verify all requested work before claiming completion.
tags: [verification, quality]
activationPatterns:
  - "done"
    - "complete"
    priority: 100
    ---

    Before saying work is complete:
    1. Check the diff.
    2. Run relevant tests.
    3. Report what was verified.
    ```

### Format B: directory skill

Add this:

```text
.hex/skills/go-test-debugging/
├── SKILL.md
├── scripts/
│   └── collect-failures.sh
├── references/
│   └── go-testing.md
├── templates/
│   └── final-report.md
└── assets/
    └── example-output.txt
    ```

    This matches the Claude/Codex direction: skill directories with `SKILL.md` as the entrypoint and optional supporting files. Claude Code documents supporting files such as templates, examples, scripts, and reference documentation beside `SKILL.md`; Codex describes optional `scripts`, `references`, `assets`, and agent metadata. ([Claude API Docs][1])

## Discovery paths

Hex should load skills from these locations, in this order:

```text
1. Builtin Hex skills
   ./skills
      /usr/local/share/hex/skills
         /opt/hex/skills
            ~/.hex/builtin-skills

            2. User skills
               ~/.hex/skills
                  ~/.claude/skills
                     ~/.codex/skills

                     3. Plugin skills
                        <plugin>/skills

                        4. Project skills
                           .hex/skills
                              .claude/skills
                                 .agents/skills
                                 ```

                                 Project-local skills should override user and builtin skills **for Hex-native names**. Codex’s current docs say repository skills are discovered from `.agents/skills` directories from the current working directory up to the repository root, and that same-named skills are not merged; for Hex, use deterministic Hex precedence for now, and add qualified names for nested conflicts later. ([OpenAI Developers][2])

                                 Add nested discovery in a later phase:

                                 ```text
                                 repo/.hex/skills
                                 repo/.claude/skills
                                 repo/.agents/skills
                                 repo/apps/web/.hex/skills
                                 repo/apps/web/.claude/skills
                                 repo/apps/web/.agents/skills
                                 ```

                                 Nested skills with name collisions should appear as qualified names:

                                 ```text
                                 deploy
                                 apps/web:deploy
                                 services/api:deploy
                                 ```

## Skill manifest schema

Extend `internal/skills.Skill`.

Current struct:

```go
type Skill struct {
        Name               string
            Description        string
                Tags               []string
                    ActivationPatterns []string
                        Model              string
                            Priority           int
                                Dependencies       []string
                                    Version            string
                                        Content            string
                                            FilePath           string
                                                Source             string
}
```

Proposed:

```go
type Skill struct {
        Name        string   `yaml:"name"`
            Description string  `yaml:"description"`
                WhenToUse   string  `yaml:"when_to_use"`

                    Tags               []string `yaml:"tags"`
                        ActivationPatterns []string `yaml:"activationPatterns"`
                            Paths              []string `yaml:"paths"`

                                Version      string   `yaml:"version"`
                                    Priority     int      `yaml:"priority"`
                                        Dependencies []string `yaml:"dependencies"`

                                            Model  string `yaml:"model"`
                                                Effort string `yaml:"effort"`
                                                    Context string `yaml:"context"` // "inline" | "fork"
                                                        Agent   string `yaml:"agent"`

                                                            UserInvocable          *bool `yaml:"user-invocable"`
                                                                DisableModelInvocation bool  `yaml:"disable-model-invocation"`

                                                                    AllowedTools    []string `yaml:"allowed-tools"`
                                                                        DisallowedTools []string `yaml:"disallowed-tools"`
                                                                            AllowedCommands []string `yaml:"allowed-commands"`
                                                                                DisallowedCommands []string `yaml:"disallowed-commands"`

                                                                                    Arguments    []string `yaml:"arguments"`
                                                                                        ArgumentHint string   `yaml:"argument-hint"`

                                                                                            Shell string `yaml:"shell"` // "bash" | "powershell" later

                                                                                                Content string `yaml:"-"`
                                                                                                    FilePath string `yaml:"-"`
                                                                                                        RootDir string `yaml:"-"`
                                                                                                            Source string `yaml:"-"`

                                                                                                                ContentHash string `yaml:"-"`
}
```

Compatibility aliases to support:

```yaml
activationPatterns   # existing Hex
activation-patterns  # kebab alias

allowedTools         # existing-style camel
allowed-tools        # Claude/Codex-style kebab

disallowedTools
disallowed-tools

userInvocable
user-invocable

disableModelInvocation
disable-model-invocation

whenToUse
when_to_use
```

Claude Code’s frontmatter includes behavior fields such as `disable-model-invocation`, `user-invocable`, `allowed-tools`, `disallowed-tools`, `model`, `effort`, `context`, `agent`, `paths`, `shell`, and argument substitution hints. ([Claude API Docs][1]) Hex does not need to implement all semantics immediately, but it should parse and preserve them from day one.

## Example skill

```markdown
---
name: go-test-debugging
description: Diagnose and fix failing Go tests using minimal changes and targeted verification.
when_to_use: Use when the user mentions go test failures, panic output, race failures, or package-level test failures.
tags: [go, testing, debugging]
activationPatterns:
  - "go test"
    - "failing test"
      - "panic:"
        - "race detector"
        paths:
          - "**/*.go"
          priority: 80
          dependencies:
            - verification-before-completion
            allowed-tools:
              - read_file
                - grep
                  - glob
                    - edit
                      - bash
                        - todo_write
                        disallowed-tools:
                          - write_file
                          allowed-commands:
                            - "go test ./..."
                              - "go test ./... -race"
                                - "go test ./... -run *"
                                  - "go test ./... -count=1"
                                  model: inherit
                                  effort: high
                                  ---

# Go Test Debugging

Use the smallest failing test command first.

## Procedure

1. Reproduce the failure.
2. Identify the smallest package or test that fails.
3. Read only the files needed to explain the failure.
4. Make the smallest code change.
5. Re-run the exact failing test.
6. Re-run the package test.
7. Summarize root cause, patch, and verification.

## Stop Conditions

Stop and ask if:
- the failure is nondeterministic after two retries
- fixing requires architecture outside the requested area
- a command needs network access
```

## Runtime behavior

### Session startup

At Hex startup:

```text
1. Load config.
2. Load AGENTS.md / CLAUDE.md / project memory as ambient instructions.
3. Load skill registry.
4. Build initial compact skill index.
5. Register Skill tool with all other tools.
6. Run mux agent.
```

The mux runner currently loads `AGENTS.md` context into the system prompt and appends project memory.  Reuse that flow and insert skill summaries after ambient repo instructions but before task-specific user prompts.

### Progressive disclosure

Initial context should not include full skill bodies. It should include compact skill cards:

```go
type SkillCard struct {
        Name string
            Description string
                WhenToUse string
                    Source string
                        Path string
                            UserInvocable bool
                                ModelInvocable bool
}
```

Budget:

```go
MaxInitialSkillCards = 20
MaxInitialSkillIndexChars = 8000
MaxDescriptionChars = 512
```

This mirrors Codex’s progressive disclosure behavior: Codex starts with a compact list of skill metadata and loads full `SKILL.md` only after selecting a skill; its docs also describe an initial context budget for the skill list. ([OpenAI Developers][2])

### Automatic skill selection

Implement:

```go
type Selector struct {
        Registry *Registry
            MaxSkills int
                MaxChars int
}

func (s *Selector) Select(prompt string, touchedPaths []string) []*Skill
```

Selection signals:

```text
1. Explicit slash command: strongest
2. Exact skill mention by name
3. activationPatterns regex match
4. path glob match
5. description/when_to_use lexical match
6. priority
```

Current Hex already supports regex activation matching via `FindByPattern`.

### Skill invocation

The `Skill` tool currently accepts:

```json
{"command": "skill-name"}
```

Keep that, but extend:

```json
{
      "command": "go-test-debugging",
        "arguments": "failing TestRunLoop",
          "reason": "User asked to fix a Go test failure"
}
```

Update schema in `internal/tools/registry.go` from:

```go
case "Skill":
    properties: command
    ```

    to:

    ```go
    case "Skill":
        properties:
              command: string
                    arguments: string
                          reason: string
                          ```

                          Hex already has a `Skill` schema entry, but it only exposes `command`.

### Argument substitution

Support these substitutions in skill content:

```text
$ARGUMENTS
$ARGUMENTS[0]
$0
$1
$name
${HEX_SESSION_ID}
${HEX_RUN_ID}
${HEX_WORKSPACE_ROOT}
${HEX_MODEL}
```

Start with `$ARGUMENTS` and positional `$0`, `$1`. Add named arguments later.

## Mux runner integration

Current mux runner constructs tools manually:

```go
baseTools := []tools.Tool{
        tools.NewReadTool(),
            tools.NewWriteTool(),
                tools.NewEditTool(),
                    tools.NewBashTool(),
                        tools.NewGrepTool(),
                            tools.NewGlobTool(),
}
```

and then appends `TaskTool`.

Change to include skill initialization:

```go
func getHexToolsWithMuxSubagents(llmClient llm.Client) ([]tools.Tool, error) {
        taskTool := tools.NewTaskTool()

            skillRegistry, skillTool := initializeSkills(pluginSkillPaths)
                _ = skillRegistry // eventually pass to SkillInjector

                    baseTools := []tools.Tool{
                                tools.NewReadTool(),
                                        tools.NewWriteTool(),
                                                tools.NewEditTool(),
                                                        tools.NewBashTool(),
                                                                tools.NewGrepTool(),
                                                                        tools.NewGlobTool(),
                                                                                skillTool,
                                                                                    }

                                                                                        toolFactory := func() []tools.Tool {
                                                                                                    return []tools.Tool{
                                                                                                                    tools.NewReadTool(),
                                                                                                                                tools.NewWriteTool(),
                                                                                                                                            tools.NewEditTool(),
                                                                                                                                                        tools.NewBashTool(),
                                                                                                                                                                    tools.NewGrepTool(),
                                                                                                                                                                                tools.NewGlobTool(),
                                                                                                                                                                                            skillTool,
                                                                                                                                                                                                    }
                                                                                                                                                                                                        }

                                                                                                                                                                                                            agentRunner := adapter.NewAgentRunner(llmClient, toolFactory)
                                                                                                                                                                                                                taskTool.SetMuxRunner(agentRunner)

                                                                                                                                                                                                                    return append(baseTools, taskTool), nil
}
```

Current `initializeSkills()` creates the loader, loads skills, registers them, and returns the registry plus a `tools.Tool` adapter.

## Skill-scoped policy

When a skill is active, apply skill policy to the current turn.

Effective tools:

```text
available = sessionTools
available = available ∩ skill.allowed-tools       if allowed-tools is non-empty
available = available - skill.disallowed-tools
```

For tools with params, apply command/path policy too:

```text
bash:
  if command matches allowed-commands → auto or normal mode
        if command matches disallowed-commands → deny
              otherwise → existing permission mode

              read_file/write_file/edit:
                enforce workspace path policy
                  optionally apply skill.paths
                  ```

                  This requires the permission checker to become param-aware. The current Hex permission checker is tool-name oriented and ignores params in `Check`.

                  Add:

                  ```go
                  type SkillPolicy struct {
                          SkillName string
                              AllowedTools []string
                                  DisallowedTools []string
                                      AllowedCommands []string
                                          DisallowedCommands []string
                  }

                  type PolicyInput struct {
                          ToolName string
                              Params map[string]interface{}
                                  ActiveSkill *Skill
                                      WorkspaceRoot string
                  }

                  type PolicyDecision struct {
                          Allowed bool
                              RequiresPrompt bool
                                  Reason string
                  }
                  ```

## Security model

### Skill trust levels

```text
trusted_builtin
trusted_user
trusted_project
untrusted_plugin
untrusted_remote
```

Markdown-only project skills can be loaded after workspace trust. Skills containing executable scripts, shell injection, hooks, MCP config, or plugin metadata require explicit trust.

Create a DB table:

```sql
CREATE TABLE IF NOT EXISTS trusted_skills (
    id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
            source TEXT NOT NULL,
                path TEXT NOT NULL,
                    content_hash TEXT NOT NULL,
                        trusted_at TIMESTAMP NOT NULL
                        );
```

If a skill changes, its hash changes and trust resets.

### Script execution

Do not auto-execute scripts in Phase 1. If `SKILL.md` references `scripts/foo.sh`, the model may read it, but running it requires normal Bash/tool approval.

### Dynamic context injection

Claude Code supports inline command expansion in skills using `!` command syntax; for example, a skill can inline `git diff HEAD` before the model sees the skill content. ([Claude API Docs][1])

Hex should **not** implement automatic command expansion in Phase 1. Add a safe placeholder mechanism later:

```markdown
{{hex.git_diff}}
{{hex.git_status}}
{{hex.file "path/to/file"}}
```

These placeholders should resolve through tools/policy, not raw shell.

## CLI commands

Add `cmd/hex/skills_cmd.go`.

```bash
hex skills list [--all] [--source project|user|builtin|plugin] [--json]
hex skills show <name> [--content] [--json]
hex skills search <query> [--json]
hex skills doctor [--json]
hex skills create <name> [--directory|--file]
hex skills trust <name>
hex skills untrust <name>
```

Phase 1 only needs:

```text
list
show
search
doctor
create
```

`trust/untrust` can land with script support.

## Data structures

### Loader changes

Current loader scans `.md` files in a directory.

Update `loadFromDir`:

```go
func (l *Loader) loadFromDir(dir, source string) ([]*Skill, error) {
        // 1. Existing behavior: load *.md files, excluding SKILL.md inside dirs.
            // 2. New behavior: load */SKILL.md directory skills.
}
```

Rules:

```text
foo.md                  → skill name from frontmatter or filename
foo/SKILL.md            → skill command name from directory unless plugin root
foo/bar.md              → existing recursive file skill support, but consider deprecating
foo/scripts/*.sh        → supporting file, not separate skill
```

### Parser changes

Add:

```go
func ParseSkillFile(path string) (*Skill, error)
func ParseSkillDir(root string) (*Skill, error)
func ParseBytes(path string, root string, data []byte) (*Skill, error)
```

### Registry changes

Add qualified names and source metadata:

```go
type Registry struct {
        skills map[string]*Skill
            byQualifiedName map[string]*Skill
}
```

Add:

```go
func (r *Registry) Resolve(command string) (*Skill, error)
func (r *Registry) Cards(maxChars int) []SkillCard
func (r *Registry) FindRelevant(prompt string, paths []string, max int) []*Skill
```

## Builtin skills

Add these builtin skills in `skills/`:

```text
verification-before-completion
test-driven-development
systematic-debugging
go-test-debugging
code-review
summarize-changes
```

Hex already has `skills/verification-before-completion.md` in the repo.  Convert it to either `skills/verification-before-completion.md` with expanded frontmatter or `skills/verification-before-completion/SKILL.md`.

## Tests

### Parser tests

```text
TestParseSingleFileSkill
TestParseDirectorySkill
TestParseSkillFrontmatterAliases
TestParseSkillDefaultsNameFromDirectory
TestParseSkillRejectsMissingDescriptionForDirectorySkill
TestParseSkillAllowsExistingHexNameDescriptionFormat
```

### Loader tests

```text
TestLoaderLoadsBuiltinUserPluginProject
TestLoaderProjectOverridesUser
TestLoaderDirectorySkill
TestLoaderIgnoresSupportingMarkdownFilesAsSkills
TestLoaderLoadsClaudeSkillsDirectory
TestLoaderLoadsAgentsSkillsDirectory
```

### Registry tests

```text
TestRegistryFindByPattern
TestRegistrySearch
TestRegistryCardsRespectBudget
TestRegistryQualifiedNamesForNestedConflicts
```

### Tool tests

```text
TestSkillToolLoadsFullContent
TestSkillToolSubstitutesArguments
TestSkillToolSuggestsSimilarSkills
TestSkillToolReportsAvailableSkills
```

### Mux integration tests

```text
TestMuxRunnerIncludesSkillTool
TestMuxRunnerInjectsRelevantSkillCards
TestMuxRunnerDoesNotInjectFullSkillBodies
TestMuxSubagentReceivesSkillTool
```

### Security tests

```text
TestSkillHashChangesWhenContentChanges
TestScriptSkillRequiresTrust
TestUntrustedPluginSkillNotAutoInvoked
TestSkillAllowedToolsRestrictsToolDefinitions
TestSkillDisallowedToolsRemovesToolDefinitions
```

## Implementation phases

### Phase 1: Wire current skills into mux

Tasks:

1. Add `skillTool` to `getHexToolsWithMuxSubagents`.
2. Expand `Skill` tool schema with `arguments` and `reason`.
3. Add `hex skills list/show/search`.
4. Add tests proving mux tool definitions include `Skill`.

Acceptance:

```text
hex --print --mux "Use the verification skill" can call Skill.
hex skills list shows builtin/user/project skills.
Existing .md skills continue working.
```

### Phase 2: Directory skills

Tasks:

1. Add `foo/SKILL.md` parsing.
2. Add `RootDir`, `ContentHash`, and supporting-file metadata.
3. Add `.claude/skills` and `.agents/skills` discovery.
4. Add `hex skills doctor`.

Acceptance:

```text
.hex/skills/foo/SKILL.md loads as /foo.
.claude/skills/foo/SKILL.md loads.
.agents/skills/foo/SKILL.md loads.
foo/scripts/bar.sh is not treated as another skill.
```

### Phase 3: Progressive disclosure and auto-selection

Tasks:

1. Add `SkillCard`.
2. Add `Selector`.
3. Inject relevant skill cards into mux system prompt.
4. Enforce skill index char budget.
5. Keep full skill body loaded only through `Skill` tool or slash invocation.

Acceptance:

```text
Prompt mentioning "go test" surfaces go-test-debugging card.
Full go-test-debugging body is absent until Skill(command="go-test-debugging").
Skill card context never exceeds configured budget.
```

### Phase 4: Slash invocation

Tasks:

1. Parse `/skill-name args...` in interactive mode.
2. Parse direct slash invocation in print mode.
3. Add autocomplete/listing integration.
4. Respect `user-invocable: false`.

Acceptance:

```text
/go-test-debugging failing TestX loads skill and runs task.
Skills with user-invocable: false do not appear in slash list.
```

### Phase 5: Skill-scoped policy

Tasks:

1. Parse `allowed-tools`, `disallowed-tools`, `allowed-commands`.
2. Apply allowed/disallowed tools to mux tool definitions for the active turn.
3. Make permission checker param-aware for Bash commands.
4. Add policy audit events.

Acceptance:

```text
A skill with disallowed-tools: [write_file] cannot call write_file.
A skill with allowed-commands only auto-allows matching bash commands.
Denied tool calls return useful tool errors, not crashes.
```

### Phase 6: Trust and script safety

Tasks:

1. Add `trusted_skills` table.
2. Compute skill hashes.
3. Add `hex skills trust/untrust`.
4. Mark scripts/hooks/MCP-bearing skills as requiring trust.
5. Block automatic model invocation for untrusted executable skills.

Acceptance:

```text
Markdown-only builtin skill loads automatically.
Project skill with scripts requires trust before auto-invocation.
Changing SKILL.md invalidates trust.
```

## Open decisions

1. Should Hex prefer `.hex/skills` over `.claude/skills` when both exist with the same name? Recommendation: yes.
2. Should `.agents/skills` be treated as Codex compatibility only, or as first-class Hex project skills? Recommendation: first-class.
3. Should Hex support Claude-style `!` command dynamic injection? Recommendation: not initially; use safe Hex placeholders later.
4. Should `Skill` remain capitalized? Existing Hex schema uses `Skill`; keep it for compatibility, but consider aliasing `skill` later.
5. Should skills eventually move to mux? Recommendation: yes, after Hex v2 proves the runtime.

## Definition of done

The feature is complete when:

```text
- Existing Hex .md skills still work.
- Directory skills with SKILL.md work.
- Skills are available in mux mode.
- Skill list/search/show CLI exists.
- Relevant skills are surfaced automatically via compact cards.
- Full skill bodies are loaded only on invocation.
- Slash invocation works.
- Project skills can override builtin skills.
- Skill-scoped allowed/disallowed tools work.
- Script-bearing skills are not auto-executed and require trust.
- Tests cover parser, loader, registry, tool, mux integration, and policy.
```

## Suggested spec filename

```text
docs/plans/2026-06-20-hex-skill-system-v2.md
```

[1]: https://docs.anthropic.com/en/docs/claude-code/skills "Extend Claude with skills - Claude Code Docs"
[2]: https://developers.openai.com/codex/skills "Agent Skills – Codex | OpenAI Developers"
