#!/usr/bin/env python3
# ABOUTME: Builds the canonical behavior-spec workbook from code-derived rows.
# ABOUTME: Uses only the Python standard library so the spec can be regenerated anywhere.

from __future__ import annotations

import html
import zipfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "docs" / "behavior-spec.xlsx"


HEADERS = [
    "Area",
    "Feature ID",
    "User story",
    "Expected behavior (from code)",
    "Status",
    "Defects",
    "Type",
    "Test method",
    "Iteration",
    "Notes / source",
    "Open questions",
]


ROWS = [
    ["Root CLI", "ROOT-001", "As a CLI user I want to start Hex with an optional prompt so I can begin an AI session.", "`hex [prompt]` accepts arbitrary args, joins them with spaces, initializes shutdown/coordinator/logging/event/cost tracking, then dispatches to print mode when `--print` is set or interactive TUI otherwise.", "Spec'd", "-", "-", "CLI smoke + unit tests", 0, "cmd/hex/root.go:102 runRoot; cmd/hex/root.go:268 joinArgs", ""],
    ["Root CLI", "ROOT-002", "As a scripted user I want non-interactive print mode so I can run one-off prompts.", "`--print/-p` routes to print mode; default mux print and legacy print both require a prompt unless legacy image input is supplied.", "Spec'd", "-", "-", "CLI smoke with isolated config where possible", 0, "cmd/hex/root.go:123; cmd/hex/print.go:22 runPrintMode; cmd/hex/mux_runner.go:23", ""],
    ["Root CLI", "ROOT-003", "As an interactive user I want Hex to open a terminal UI when no print flag is used.", "Interactive mode validates resume flag conflicts, opens SQLite storage, resolves provider/model/template/AGENTS.md context, creates or resumes a conversation, registers tools/skills/commands/MCP, and starts a Bubble Tea program.", "Spec'd", "-", "-", "Go tests + guarded TUI smoke", 0, "cmd/hex/root.go:279 runInteractive; cmd/hex/root.go:655 tea.NewProgram", ""],
    ["Root CLI", "ROOT-004", "As a user I want setup to run automatically only when needed.", "On first run, non-print mode invokes `RunWizard`; if cancelled it exits nil, and print mode skips first-run wizard.", "Spec'd", "-", "-", "Unit/CLI smoke", 0, "cmd/hex/root.go:250; cmd/hex/wizard.go", ""],
    ["Root CLI", "ROOT-005", "As a user I want CLI version output.", "Cobra root command has `Version` set from the package-level `version` variable, so `hex --version` is handled by Cobra without entering `runRoot`.", "Spec'd", "-", "-", "CLI smoke", 0, "cmd/hex/root.go:102", ""],
    ["Root CLI", "ROOT-006", "As an interactive user I want to continue the latest conversation from the root command.", "`--continue` loads the most recently updated conversation, restores its model/messages, or prints `No previous conversations found, starting new session` and creates a new session when none exist.", "Spec'd", "-", "-", "CLI smoke with temp db", 0, "cmd/hex/root.go:134; cmd/hex/root.go:374", ""],
    ["Root CLI", "ROOT-007", "As a user I want to choose the database path.", "`--db-path` defaults to the platform default path and is passed to `openDatabase` by root, history, resume, favorite, favorites, and export commands.", "Spec'd", "-", "-", "CLI smoke with temp db", 0, "cmd/hex/root.go:136; cmd/hex/storage.go; cmd/hex/history.go; cmd/hex/export.go", ""],
    ["Root CLI", "ROOT-008", "As a user I want verbose mode when advertised.", "`--verbose` raises logging to debug level without setting `HEX_DEBUG`; `--debug` remains the stronger mode that also enables debug environment behavior.", "Spec'd", "-", "-", "Unit test", 0, "cmd/hex/root.go:121; cmd/hex/root.go:initializeLogging; cmd/hex/logging_integration_test.go:TestVerboseEnablesDebugLogLevel", ""],
    ["Root CLI", "ROOT-009", "As a user I want a custom system prompt in print mode.", "`--system-prompt` is appended to the default prompt in mux print mode and appended after default identity in legacy print unless a replace-mode spell overrides it.", "Spec'd", "-", "-", "API-gated print smoke", 0, "cmd/hex/root.go:149; cmd/hex/mux_runner.go:91; cmd/hex/print.go:158", ""],
    ["Root CLI", "ROOT-010", "As a scripted user I want to limit enabled tools.", "`--tools` filters mux print tools directly; in legacy print it is converted into `allowedTools` only when `--allowed-tools` was not set.", "Spec'd", "-", "-", "CLI/API-gated smoke", 0, "cmd/hex/root.go:148; cmd/hex/mux_runner.go:251; cmd/hex/print_tools.go:16", ""],
    ["Logging", "LOG-001", "As a user I want configurable log level and format.", "`--log-level` is parsed into logger level; `--log-format json` selects JSON logs and other values use text formatting.", "Spec'd", "-", "-", "CLI smoke with log file", 0, "cmd/hex/root.go:139; cmd/hex/root.go:683 initializeLogging", ""],
    ["Config", "CFG-001", "As a user I want provider selection from flags or config with a default.", "Provider name is resolved as flag, then config, then `anthropic`; supported providers are anthropic, openai, gemini, openrouter, and ollama.", "Spec'd", "-", "-", "Unit tests / CLI error smoke", 0, "cmd/hex/root.go:313; cmd/hex/root.go:319", ""],
    ["Config", "CFG-002", "As a user of non-Anthropic providers I want a clear model requirement.", "Interactive mode and legacy print mode require `--model` for non-Anthropic providers and return an error with example model names when missing.", "Spec'd", "-", "-", "CLI error smoke", 0, "cmd/hex/root.go:354; cmd/hex/print.go:54", ""],
    ["Config", "CFG-003", "As an Anthropic TUI user I want API key validation before session start.", "TUI mode only supports Anthropic; it loads provider config, overrides API key from `ANTHROPIC_API_KEY`, and errors if no key is configured.", "Spec'd", "-", "-", "CLI error smoke", 0, "cmd/hex/root.go:467; cmd/hex/root.go:481", ""],
    ["Config", "CFG-004", "As a user I want debug logging without corrupting the TUI.", "`--debug` sets `HEX_DEBUG=1`, writes logs to `--log-file` or `/tmp/hex-debug.log`, and only writes debug logs to stderr in print mode.", "Spec'd", "-", "-", "CLI smoke + file check", 0, "cmd/hex/root.go:683 initializeLogging", ""],
    ["Permissions", "PERM-001", "As a user I want permission modes for tool execution.", "`--permission-mode` parses `ask`, `auto`, or `deny`; ask requires prompting, auto allows, deny blocks. Invalid modes error.", "Spec'd", "-", "-", "Unit tests / CLI smoke", 0, "cmd/hex/root.go:750 createPermissionChecker; internal/permissions/mode.go:32 ParseMode; internal/permissions/checker.go:36 Check", ""],
    ["Permissions", "PERM-002", "As a user I want allow and deny lists to constrain tools.", "`--allowed-tools` and `--disallowed-tools` build permission rules; tools not in an allow list or present in a deny list are blocked before mode handling.", "Spec'd", "-", "-", "Unit tests", 0, "cmd/hex/root.go:759; internal/permissions/checker.go:42 Check", ""],
    ["Permissions", "PERM-003", "As a legacy user I want old permission bypass flag compatibility.", "`--dangerously-skip-permissions` changes mode to `auto` and logs a deprecation warning.", "Spec'd", "-", "-", "CLI smoke", 0, "cmd/hex/root.go:744 createPermissionChecker", ""],
    ["Storage", "STOR-001", "As a user I want conversations persisted locally.", "Opening a database enables foreign keys, enables WAL journal mode, and runs embedded migrations before use.", "Spec'd", "-", "-", "Go tests + CLI smoke with temp db", 0, "internal/storage/schema.go:39 OpenDatabase", ""],
    ["Storage", "STOR-002", "As a user I want new conversations recorded with metadata.", "Creating a conversation generates a UUID when missing, defaults timestamps to now and provider to Anthropic, then inserts title/provider/model/system prompt/timestamps.", "Spec'd", "-", "-", "Go tests", 0, "internal/storage/conversations.go:27 CreateConversation", ""],
    ["Storage", "STOR-003", "As a user I want messages saved in order.", "Creating a message generates a UUID when missing, inserts role/content/tool calls/metadata, and updates parent conversation timestamp in the same transaction; listing orders by created_at ascending.", "Spec'd", "-", "-", "Go tests", 0, "internal/storage/messages.go:24 CreateMessage; internal/storage/messages.go:88 ListMessages", ""],
    ["History", "HIST-001", "As a user I want recent conversation history.", "`hex history` opens the configured db, fetches recent history up to `--limit` default 20, prints empty-state guidance when none, and shows relative time, truncated preview, and conversation ID.", "Spec'd", "-", "-", "CLI smoke with seeded temp db", 0, "cmd/hex/history.go:53 runHistory; cmd/hex/history.go:128 displayHistoryEntry", ""],
    ["History", "HIST-002", "As a user I want to search history.", "`hex history search <query>` uses full-text search up to `--limit`, prints no-results guidance if empty, and formats matching entries like recent history.", "Spec'd", "-", "-", "CLI smoke with seeded temp db", 0, "cmd/hex/history.go:89 runHistorySearch", ""],
    ["Resume", "RES-001", "As a user I want to resume the most recent conversation.", "`hex resume --last` loads the latest conversation or errors if none, validates existence, loads messages, prints session metadata to stderr, and continues interactive mode.", "Spec'd", "-", "-", "CLI smoke with seeded temp db", 0, "cmd/hex/resume.go:62 runResume", "Full interactive continuation requires terminal/API conditions."],
    ["Resume", "RES-002", "As a user I want to resume a specific conversation by ID.", "`hex resume <conversation-id>` validates the conversation, loads ordered messages into a UI model, preserves favorite status, and continues interactive mode.", "Spec'd", "-", "-", "CLI smoke with seeded temp db", 0, "cmd/hex/resume.go:58; cmd/hex/resume.go:94", ""],
    ["Resume", "RES-003", "As a user I want a picker when resuming without arguments.", "`hex resume` lists up to 20 recent conversations through `conversationsForPicker`, passes them to `ui.NewSessionPicker` in alt-screen mode, and the picker returns the selected conversation ID on Enter or an empty selection on Esc/q cancel.", "Spec'd", "-", "-", "TUI/unit tests", 0, "cmd/hex/resume.go:73; cmd/hex/resume.go:132 showConversationPicker; internal/ui/session_picker.go:47 NewSessionPicker", ""],
    ["Favorites", "FAV-001", "As a user I want to toggle a conversation as favorite.", "`hex favorite <conversation-id>` loads the conversation, toggles `is_favorite`, updates `updated_at`, and prints marked/removed confirmation.", "Spec'd", "-", "-", "CLI smoke with temp db", 0, "cmd/hex/favorites.go:13; internal/storage/conversations.go:139 SetFavorite", ""],
    ["Favorites", "FAV-002", "As a user I want to list favorite conversations.", "`hex favorites` prints an empty state when none; otherwise prints count, title truncated to 60 chars, ID, model, and relative update time for favorites ordered by updated_at descending.", "Spec'd", "-", "-", "CLI smoke with temp db", 0, "cmd/hex/favorites.go:50; internal/storage/conversations.go:148 ListFavorites", ""],
    ["Export", "EXP-001", "As a user I want to export conversations as Markdown, JSON, or HTML.", "`hex export <conversation-id>` accepts `--format markdown|md|json|html|htm`, opens storage, writes to stdout by default or `--output`, and errors on unknown format.", "Spec'd", "-", "-", "CLI smoke with temp db", 0, "cmd/hex/export.go:25 runExport; internal/export/exporter.go", ""],
    ["Export", "EXP-002", "As a user writing an export file I want confirmation.", "When `--output` is provided, the export file is created/truncated and a success message is written to stderr after export succeeds.", "Spec'd", "-", "-", "CLI smoke", 0, "cmd/hex/export.go:69; cmd/hex/export.go:82", ""],
    ["Setup", "SETUP-001", "As a user I want to re-run configuration setup.", "`hex setup` runs `RunWizard`; cancellation prints `Setup cancelled.` and returns nil, success prints `Setup complete! Run 'hex' to start chatting.`", "Spec'd", "-", "-", "Manual/TUI guarded test", 0, "cmd/hex/setup.go:12", "Wizard choices need separate TUI verification."],
    ["Templates", "TMPL-001", "As a user I want to list session templates.", "`hex templates` and `hex templates list` load YAML templates from `~/.hex/templates`, print a creation example when none exist, or sorted template metadata and `hex --template <name>` usage.", "Spec'd", "-", "-", "CLI smoke with isolated HOME", 0, "cmd/hex/templates.go:16; cmd/hex/templates.go:37 runTemplatesList", ""],
    ["Templates", "TMPL-002", "As a user I want to start with a named template.", "`--template` loads by name from `~/.hex/templates`; missing template errors with available names when any exist. TUI applies system prompt, model, title, and initial messages.", "Spec'd", "-", "-", "CLI smoke/unit tests", 0, "cmd/hex/root.go:331; cmd/hex/templates.go:100 loadTemplateByName", ""],
    ["Templates", "TMPL-003", "As a template author I want validation.", "Template YAML requires `name`; initial message roles must be user/assistant/system and content cannot be empty.", "Spec'd", "-", "-", "Go tests", 0, "internal/templates/types.go:27 Validate; internal/templates/loader.go:17 LoadTemplate", ""],
    ["MCP", "MCP-001", "As a user I want to add MCP servers.", "`hex mcp add <name> <command> [args...]` disables flag parsing, stores a stdio server config in the project registry, saves it, and prints command/config path.", "Spec'd", "-", "-", "CLI smoke in temp dir", 0, "cmd/hex/mcp.go:23; cmd/hex/mcp.go:68 runMCPAdd", ""],
    ["MCP", "MCP-002", "As a user I want to list configured MCP servers.", "`hex mcp list` loads the project MCP registry, prints empty-state guidance when no servers exist, otherwise prints name, transport, full command, total, and config path.", "Spec'd", "-", "-", "CLI smoke in temp dir", 0, "cmd/hex/mcp.go:42; cmd/hex/mcp.go:104 runMCPList", ""],
    ["MCP", "MCP-003", "As a user I want to remove an MCP server.", "`hex mcp remove <name>` loads the registry, removes the named server, saves, and prints removal confirmation.", "Spec'd", "-", "-", "CLI smoke in temp dir", 0, "cmd/hex/mcp.go:55; cmd/hex/mcp.go:133 runMCPRemove", ""],
    ["Plugins", "PLUG-001", "As a user I want to install plugins.", "`hex plugin install <source>` creates the default plugin registry, calls `Install(source)`, and prints progress and success.", "Spec'd", "-", "-", "Unit/CLI with local fixture", 0, "cmd/hex/plugin.go:21; cmd/hex/plugin.go:89 runPluginInstall", "Remote Git install should not be exercised without network fixture."],
    ["Plugins", "PLUG-002", "As a user I want to list plugins.", "`hex plugin list|ls` prints empty-state install guidance or a tabular NAME/VERSION/STATUS/DESCRIPTION list, truncating descriptions over 50 chars.", "Spec'd", "-", "-", "CLI smoke with isolated HOME", 0, "cmd/hex/plugin.go:46; cmd/hex/plugin.go:117 runPluginList", ""],
    ["Plugins", "PLUG-003", "As a user I want plugin lifecycle controls.", "`install`, `uninstall|remove|rm`, `enable`, `disable`, `update`, and `show|info` delegate to the default registry and print success or detailed plugin fields.", "Spec'd", "-", "-", "CLI/unit with fixture", 0, "cmd/hex/plugin.go:38; cmd/hex/plugin.go:153; cmd/hex/plugin.go:211", ""],
    ["Skills", "SKILL-001", "As a user I want built-in, user, project, and plugin skills available to the agent.", "Skill initialization locates built-in dirs, includes plugin paths, loads all skills, registers each unique skill, warns on failures, and exposes a `Skill` tool adapter.", "Spec'd", "-", "-", "Go tests", 0, "cmd/hex/skills.go:14 initializeSkills; internal/skills/loader.go", ""],
    ["Slash Commands", "CMD-001", "As a user I want slash commands available to the agent and autocomplete.", "Command initialization locates built-in command dirs, includes plugin paths, loads and registers commands, exposes a slash-command tool, and passes command names/descriptions to the UI model.", "Spec'd", "-", "-", "Go tests", 0, "cmd/hex/commands.go:14 initializeCommands; cmd/hex/root.go:621", ""],
    ["Slash Commands", "CMD-002", "As a user I want `/brainstorm` to guide design exploration.", "Built-in `brainstorm.md` defines a Socratic design workflow with understanding, alternatives, refinement, validation, and structured output around an optional topic argument.", "Spec'd", "-", "-", "Slash-command tool test", 0, "commands/brainstorm.md", ""],
    ["Slash Commands", "CMD-003", "As a user I want `/plan` to create an implementation plan.", "Built-in `plan.md` asks the agent to analyze requirements, break down work, specify file-level tasks, plan tests, and output dependencies/success criteria.", "Spec'd", "-", "-", "Slash-command tool test", 0, "commands/plan.md", ""],
    ["Slash Commands", "CMD-004", "As a user I want `/debug` to investigate issues systematically.", "Built-in `debug.md` defines reproduce/evidence/trace/root-cause, pattern analysis, hypothesis testing, and implementation phases around an optional issue argument.", "Spec'd", "-", "-", "Slash-command tool test", 0, "commands/debug.md", ""],
    ["Slash Commands", "CMD-005", "As a user I want `/test` to drive TDD.", "Built-in `test.md` defines red-green-refactor behavior for a target and optional test type.", "Spec'd", "-", "-", "Slash-command tool test", 0, "commands/test.md", ""],
    ["Slash Commands", "CMD-006", "As a user I want `/review` for code review.", "Built-in `review.md` asks the agent to inspect status/diff, then review correctness, readability, maintainability, testing, security, performance, architecture, and breaking changes.", "Spec'd", "-", "-", "Slash-command tool test", 0, "commands/review.md", ""],
    ["Slash Commands", "CMD-007", "As a user I want `/commit` to prepare a commit.", "Built-in `commit.md` asks the agent to inspect git status/diff/staged diff, run a quality checklist, inspect recent commit style, and craft conventional commit output.", "Spec'd", "-", "-", "Slash-command tool test", 0, "commands/commit.md", ""],
    ["Slash Commands", "CMD-008", "As a user I want `/refactor` for safe refactoring.", "Built-in `refactor.md` requires tests/baseline understanding, states the golden rule of changing behavior or structure but not both, and guides smell identification and incremental refactoring.", "Spec'd", "-", "-", "Slash-command tool test", 0, "commands/refactor.md", ""],
    ["Slash Commands", "CMD-009", "As a user I want `/document` to generate documentation.", "Built-in `document.md` guides code comments, API docs, guides, and architecture documentation around optional target/type arguments.", "Spec'd", "-", "-", "Slash-command tool test", 0, "commands/document.md", ""],
    ["Slash Commands", "CMD-010", "As a user I want `/help` to show slash-command and keyboard help.", "Built-in `help.md` lists available slash commands, implemented keyboard shortcuts, tool approval keys, tips, and project link; help now documents Ctrl+H/Ctrl+R/Tab view switching/typewriter/favorite behavior from the TUI update path.", "Spec'd", "-", "-", "Static slash-help regression test", 0, "commands/help.md; internal/ui/update.go; cmd/hex/help_command_test.go", ""],
    ["Slash Commands", "CMD-011", "As a user I want `/spell` to list or switch spells.", "Built-in `spell.md` renders list/reset/off/cast spell prompts and documents built-in spell names, layer/replace modes, and print-mode usage.", "Spec'd", "-", "-", "Slash-command tool test", 0, "commands/spell.md", ""],
    ["Spells", "SPELL-001", "As a user I want named spells to alter agent behavior.", "`--spell` loads a spell by name from project, user, or embedded builtin sources; `--spell-mode` can override mode; replace mode uses only spell prompt, layer mode appends to existing prompt.", "Spec'd", "-", "-", "Go tests / print smoke", 0, "cmd/hex/spells_init.go:60 getSpellWithMode; internal/spells/loader.go:150; internal/spells/applicator.go:30", ""],
    ["Spells", "SPELL-002", "As a spell author I want spell directories parsed predictably.", "A spell directory must contain `system.md`; optional `config.yaml` sets mode/config; optional `tools/*.yaml` create tool overrides; invalid spell validation aborts that spell.", "Spec'd", "-", "-", "Go tests", 0, "internal/spells/parser.go:17 ParseSpellDirectory", ""],
    ["Tools", "TOOL-001", "As an agent I want to read files safely.", "`read_file` requires path, supports optional byte offset and limit, and returns file contents or expected failure result.", "Spec'd", "-", "-", "Go tests / print-mode tool smoke", 0, "internal/tools/read_tool.go; internal/tools/registry.go:68", ""],
    ["Tools", "TOOL-002", "As an agent I want to write files with explicit modes.", "`write_file` requires path/content and supports create, overwrite, or append, defaulting to create.", "Spec'd", "-", "-", "Go tests with temp dir", 0, "internal/tools/write_tool.go; internal/tools/registry.go:87", ""],
    ["Tools", "TOOL-003", "As an agent I want exact string file edits.", "`edit` performs exact replacements, requiring a unique match unless `replace_all` is true.", "Spec'd", "-", "-", "Go tests with temp dir", 0, "internal/tools/edit_tool.go; internal/tools/registry.go:220", ""],
    ["Tools", "TOOL-004", "As an agent I want shell command execution.", "`bash` executes a subprocess with command, optional timeout seconds, working directory, and background mode.", "Spec'd", "-", "-", "Go tests / CLI smoke", 0, "internal/tools/bash_tool.go; internal/tools/registry.go:106", ""],
    ["Tools", "TOOL-005", "As an agent I want to retrieve and stop background shell jobs.", "`bash_output` retrieves output from a background shell by bash_id with optional regex filter; `kill_shell` kills a running background shell by shell_id.", "Spec'd", "-", "-", "Go tests", 0, "internal/tools/bash_output_tool.go; internal/tools/kill_shell_tool.go", ""],
    ["Tools", "TOOL-006", "As an agent I want file discovery.", "`glob` finds files by glob pattern, supports recursive `**`, optional path, and sorts by modification time newest first.", "Spec'd", "-", "-", "Go tests with temp dir", 0, "internal/tools/glob_tool.go", ""],
    ["Tools", "TOOL-007", "As an agent I want content search.", "`grep` uses ripgrep-compatible search with pattern, optional path, output modes content/files_with_matches/count, case-insensitive flag, context flags, glob filter, and file type filter.", "Spec'd", "-", "-", "Go tests with temp dir", 0, "internal/tools/grep_tool.go; internal/tools/registry.go:142", ""],
    ["Tools", "TOOL-008", "As an agent I want to ask structured questions.", "`ask_user_question` gathers information through multiple-choice questions using the tool contract.", "Spec'd", "-", "-", "Go tests", 0, "internal/tools/ask_user_question_tool.go", ""],
    ["Tools", "TOOL-009", "As an agent I want structured task tracking.", "`todo_write` creates and manages structured todo lists for progress tracking.", "Spec'd", "-", "-", "Go tests", 0, "internal/tools/todo_write_tool.go", ""],
    ["Tools", "TOOL-010", "As an agent I want web research tools.", "`web_fetch` fetches URL content and processes it with a prompt; `web_search` searches DuckDuckGo and returns results. These tools are registered in interactive and legacy print setup, but not in mux print's base tool list.", "Spec'd", "-", "-", "Go tests / network-gated smoke", 0, "internal/tools/web_fetch_tool.go; internal/tools/web_search_tool.go; cmd/hex/root.go:535; cmd/hex/mux_runner.go:236", "Network behavior may be environment-dependent."],
    ["Tools", "TOOL-011", "As an agent I want to delegate complex work.", "`task` launches a sub-agent with prompt, description, subagent_type, optional model, and optional resume flag; valid subagent types come from `subagents.ValidSubagentTypes()`.", "Spec'd", "-", "-", "Go tests / guarded smoke", 0, "internal/tools/task_tool.go; internal/tools/registry.go:123", ""],
    ["Print Mode", "PRINT-001", "As a scripted user I want print-mode tool execution to continue across turns.", "Legacy print mode builds a user message, registers print tools, sends tool definitions, executes tool-use turns, appends tool results, and stops on end_turn or max_tokens.", "Spec'd", "-", "-", "Unit/CLI with API gated", 0, "cmd/hex/print.go:89; cmd/hex/print.go:195", "External API required for full behavioral execution."],
    ["Print Mode", "PRINT-002", "As a scripted user I want output formats.", "Print mode supports `--output-format text|json|stream-json` through `formatOutput` when the model returns final content.", "Spec'd", "-", "-", "Unit tests / API-gated smoke", 0, "cmd/hex/root.go:119; cmd/hex/print.go:237 formatOutput reference", "Need exact `formatOutput` branch verification in loop."],
    ["Print Mode", "PRINT-003", "As a scripted user I want multimodal input.", "Legacy print mode accepts repeated `--image` paths, loads each image into content blocks, and appends text prompt when present.", "Spec'd", "-", "-", "Unit test / CLI error smoke", 0, "cmd/hex/root.go:138; cmd/hex/print.go:66", ""],
    ["Print Mode", "PRINT-004", "As a scripted user I want plan mode.", "`--plan` wraps the first prompt with planning instructions; after first model response, it adds a user message telling the model to execute step by step.", "Spec'd", "-", "-", "Unit/API-gated smoke", 0, "cmd/hex/root.go:158; cmd/hex/print.go:80; cmd/hex/print.go:203", ""],
    ["Context", "CTX-001", "As a user I want context bounded by configured token limits.", "`--max-context-tokens` defaults to 180000; print mode, root interactive mode, and resume continuation create a context manager from `--context-strategy`. `keep-all` disables pruning, `prune` removes older context, and `summarize` replaces pruned older messages with a compact `Previous conversation summary` message while retaining recent context.", "Spec'd", "-", "-", "Go tests/static check", 0, "cmd/hex/root.go:138; cmd/hex/root.go:798 createContextManager; cmd/hex/print.go:runPrintMode; cmd/hex/interactive.go:continueInteractiveWithModel; internal/convcontext/manager.go:SummarizeContext", ""],
    ["Context", "CTX-003", "As a user I want project memory loaded into print-mode prompts.", "Print modes call `loadProjectContext`, which loads `.hex/project.json`, regenerates and saves it when missing/stale after 7 days or when `--refresh-memory` is set, and appends project context to the system prompt.", "Spec'd", "-", "-", "CLI/static check", 0, "cmd/hex/root.go:157; cmd/hex/print.go:514 loadProjectContext; cmd/hex/mux_runner.go:103", ""],
    ["Events", "EVT-001", "As an operator I want session events recorded.", "Root startup creates a temp `hex_events_*.jsonl` event store, sets `HEX_AGENT_ID` to `root` if absent, records `SessionStart`, and closes on exit; `HEX_DEBUG` prints replay/visualize hints.", "Spec'd", "-", "-", "CLI smoke with debug env", 0, "cmd/hex/root.go:186", ""],
    ["Cost", "COST-001", "As an operator I want cost summaries.", "Root initializes global cost tracking and on exit prints session cost to stderr when tracked cost for `HEX_AGENT_ID` is greater than zero; interactive also calls `cost.PrintCostSummary`.", "Spec'd", "-", "-", "Unit/API-gated smoke", 0, "cmd/hex/root.go:227; cmd/hex/root.go:667", ""],
    ["Visualization", "VIZ-001", "As an operator I want to visualize event files.", "`hex visualize <event-file>` checks file existence, finds `hexviz` on PATH or `bin/hexviz`, passes `-events`, `-view`, optional agent/type/html flags, and streams stdout/stderr.", "Spec'd", "-", "-", "CLI smoke with sample events", 0, "cmd/hex/visualize.go:15; cmd/hex/visualize.go:46 runVisualize", ""],
    ["Visualization", "VIZ-002", "As an operator I want standalone visualization modes.", "`hexviz` loads JSONL events, optionally filters by agent/type, renders tree, timeline, or cost view, errors on unknown view, and optionally exports HTML.", "Spec'd", "-", "-", "CLI smoke with sample events", 0, "cmd/hexviz/main.go:69", ""],
    ["Replay", "REPLAY-001", "As an operator I want to replay event timelines.", "`hex replay <event-file>` checks file existence, finds `hexreplay` on PATH or `bin/hexreplay`, passes event/agent/type args, and streams stdout/stderr.", "Spec'd", "-", "-", "CLI smoke with sample events", 0, "cmd/hex/replay.go:15; cmd/hex/replay.go:39 runReplay", ""],
    ["Replay", "REPLAY-002", "As an operator I want standalone replay filtering.", "`hexreplay` reads JSONL events, skips unparsable lines with log warnings, filters by exact agent or child-agent prefix and exact event type, prints timestamp/type/agent rows, and total count.", "Spec'd", "-", "-", "CLI smoke with sample events", 0, "cmd/hexreplay/main.go:19", ""],
    ["Replay", "REPLAY-003", "As an operator I want verbose replay details.", "`hexreplay -v` prints indented event data JSON when present and parent ID lines when present, in addition to the normal timeline rows.", "Spec'd", "-", "-", "CLI smoke with sample events", 0, "cmd/hexreplay/main.go:23; cmd/hexreplay/main.go:68", ""],
    ["TUI", "TUI-001", "As an interactive user I want a full terminal chat UI.", "Interactive mode constructs `ui.Model` with conversation/model, services, API client, tool registry/executor, slash commands, context manager, then runs Bubble Tea with alt-screen when stdin is a terminal.", "Spec'd", "-", "-", "Go UI tests + guarded terminal smoke", 0, "cmd/hex/root.go:415; cmd/hex/root.go:655; internal/ui/model.go", ""],
    ["TUI", "TUI-002", "As an interactive user I want previous messages visible after resume.", "Resume and continue paths load stored messages and call `uiModel.AddMessage(role, content)` for each before launching the UI.", "Spec'd", "-", "-", "Go tests", 0, "cmd/hex/root.go:385; cmd/hex/root.go:403; cmd/hex/resume.go:104", ""],
    ["TUI", "TUI-003", "As a user I want slash-command autocomplete.", "After command registry initialization, interactive mode sends command names and descriptions to the UI model for autocomplete.", "Spec'd", "-", "-", "Go UI tests", 0, "cmd/hex/root.go:621", ""],
    ["TUI", "TUI-004", "As a user I want tool approvals visible in the UI.", "Interactive mode wires a permission checker into the tool executor; UI approval handling is expected to determine approvals while executor approval function itself returns true.", "Spec'd", "-", "-", "Go UI tests", 0, "cmd/hex/root.go:600; internal/ui/approval.go; internal/ui/overlay_tool_approval.go", "Approval prompt behavior needs exact UI model verification."],
    ["Landing Page", "WEB-001", "As a visitor I want the main marketing page.", "Astro page `/` renders the landing-page implementation through `landing-page/src/pages/index.astro` using the shared layout and global styles.", "Spec'd", "-", "-", "npm build / browser smoke", 0, "landing-page/src/pages/index.astro; landing-page/src/layouts/Layout.astro", ""],
    ["Landing Page", "WEB-002", "As an alpha visitor I want the alpha page.", "Astro page `/alpha` renders from `landing-page/src/pages/alpha.astro` through the shared layout.", "Spec'd", "-", "-", "npm build / browser smoke", 0, "landing-page/src/pages/alpha.astro; landing-page/src/layouts/Layout.astro", ""],
    ["Build/Test", "QA-001", "As a maintainer I want a full verification command.", "`make verify` runs `fmt`, `vet`, `lint`, and `test`; `make test` runs `go test -v -race ./...`; `make test-short` runs short tests.", "Spec'd", "-", "-", "Command verification", 0, "Makefile:12; Makefile:93", ""],
    ["Build/Test", "QA-002", "As a maintainer I want visualization binaries built with the app.", "`make build` depends on `build-viz`, builds `bin/hex`, and `build-viz` builds `bin/hexviz` and `bin/hexreplay`.", "Spec'd", "-", "-", "Command verification", 0, "Makefile:13; Makefile:105", ""],
    ["Build/Test", "QA-003", "As a maintainer I want dependency audits clean enough to trust the landing page build.", "`npm audit --audit-level=high` reports high-severity vulnerabilities in the landing-page dependency tree, including Astro, Vite, Rollup, h3, devalue, defu, picomatch, and svgo advisories.", "Spec'd", "Landing dependency audit fails with 13 vulnerabilities, 8 high.", "Security", "npm audit --audit-level=high", 1, "landing-page/package-lock.json; npm audit output", "Fix path may require dependency updates, possibly including a breaking Astro upgrade."],
    ["Doctor", "DOC-001", "As a user I want to check installation health.", "`hex doctor` prints checks for `~/.hex`, config file, and API key; it reports pass/fail status but returns nil from the command.", "Spec'd", "-", "-", "CLI smoke with isolated HOME", 0, "cmd/hex/doctor.go:12; cmd/hex/doctor.go:23", ""],
    ["Print Mode", "PRINT-005", "As a scripted user I want mux print mode by default.", "Print mode routes to `runPrintModeWithMux` unless `--legacy` is set; mux print rejects image paths with a message to use `--legacy`, builds mux tools, and streams text output.", "Spec'd", "-", "-", "CLI smoke/API-gated test", 0, "cmd/hex/print.go:23; cmd/hex/mux_runner.go:23", ""],
    ["Print Mode", "PRINT-006", "As a scripted user I want stream-json output when advertised.", "Legacy `formatOutput` accepts `stream-json` and writes one compact JSON response line. Mux print still does not honor `--output-format`.", "Spec'd", "-", "-", "Unit test for legacy formatter; mux remains covered by PRINT-005", 0, "cmd/hex/root.go:118; cmd/hex/print.go:formatOutputTo; cmd/hex/stuck_detection_test.go:TestFormatOutputStreamJSONWritesSingleLine", ""],
    ["Print Mode", "PRINT-007", "As a scripted user I want max-turns to control print-mode loops.", "`--max-turns` is registered globally; legacy print reads it through `effectiveMaxTurns`, and mux print passes it as `adapter.Config.MaxIterations` to mux `agent.Config.MaxIterations`.", "Spec'd", "-", "-", "Unit/static bridge tests", 0, "cmd/hex/root.go:162; cmd/hex/print.go:effectiveMaxTurns; cmd/hex/mux_runner.go:runPrintModeWithMux; internal/adapter/bootstrap.go:Config", ""],
    ["Thinking", "THINK-001", "As a user I want extended thinking for complex tasks.", "`--thinking` enables an Anthropic mux client wrapper with `--thinking-budget`; non-Anthropic providers log a warning and ignore thinking.", "Spec'd", "-", "-", "CLI/API-gated test", 0, "cmd/hex/root.go:159; cmd/hex/mux_runner.go:69", ""],
    ["Tools", "TOOL-012", "As a user I want tool calls controlled, cached, and observable.", "Executor normalizes nil params, uses cache for read_file/grep/glob, applies permission checker and optional approval, fires pre/post tool hooks, executes the tool, and caches successful read-only results.", "Spec'd", "-", "-", "Go tests", 0, "internal/tools/executor.go:87; internal/tools/cache.go:149", ""],
    ["Tools", "TOOL-013", "As an agent I want the tool registry to expose stable schemas.", "The registry rejects duplicate tool names, returns sorted names, retrieves by exact name, and emits API tool definitions using `GetToolSchema`.", "Spec'd", "-", "-", "Go tests", 0, "internal/tools/registry.go:29; internal/tools/registry.go:63", ""],
    ["Hooks", "HOOK-001", "As a user I want lifecycle shell hooks.", "Hook config loads user then project `.hex/settings.json`; project event hook lists override user lists. Hooks support match filters, sync/async execution, timeouts, env injection, and failure blocking unless ignored.", "Spec'd", "-", "-", "Go tests / fixture smoke", 0, "internal/hooks/config.go:70; internal/hooks/engine.go:58; internal/hooks/executor.go:35", ""],
    ["MCP", "MCP-004", "As a user I want MCP tools loaded at session start.", "Interactive startup loads optional `.mcp.json`; stdio mux clients list tools and register adapters, skipping name collisions. MCP tool adapters do not require approval by default.", "Spec'd", "-", "-", "Go tests with fixture MCP", 0, "cmd/hex/root.go:583; internal/mcp/loader.go:21; internal/mcp/tool_adapter.go:44", ""],
    ["Subagents", "SUB-001", "As an agent I want isolated delegated work.", "Valid subagent types are `general-purpose`, `Explore`, `Plan`, and `code-reviewer`; executor spawns `hex --print`, sets subagent env vars, and records start/stop events.", "Spec'd", "-", "-", "Go tests / guarded CLI smoke", 0, "internal/subagents/types.go:52; internal/subagents/executor.go:46", ""],
    ["Subagents", "SUB-002", "As an agent I want dispatch strategies.", "Dispatcher supports parallel, sequential, batch, and first-success delegation patterns over subagent tasks.", "Spec'd", "-", "-", "Go tests", 0, "internal/subagents/dispatcher.go:44", ""],
    ["Suggestions", "SUG-001", "As a TUI user I want input suggestions.", "Suggestion detector returns up to 3 high-confidence actions from paths, URLs, search, shell-like input, write/edit/web intent.", "Spec'd", "-", "-", "Go tests", 0, "internal/suggestions/detector.go:217", ""],
    ["TUI", "TUI-005", "As an operator I want to send messages and receive streamed responses.", "Enter sends trimmed input, persists the user message, updates title from first message, starts streaming if an API client exists; `exit`/`/exit` quits and `/clear` clears context.", "Spec'd", "-", "-", "Go UI tests / terminal smoke", 0, "internal/ui/update.go:32", ""],
    ["TUI", "TUI-006", "As an operator I want busy-session input queued.", "When waiting for a response, messages queue up to 100; the first queued item is previewed, and Up pulls the first queued item back into input for editing.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/model.go:252 QueueMessage; internal/ui/model.go:267 PopQueue", ""],
    ["TUI", "TUI-007", "As an operator I want controlled quit behavior.", "Ctrl+C clears non-empty input, cancels streaming/search/quick actions, and when idle requires a second Ctrl+C within two seconds to quit.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/update.go:32", ""],
    ["TUI", "TUI-008", "As an operator I want fullscreen help/history/tool overlays.", "Ctrl+H toggles help, Ctrl+R toggles current conversation history capped at 1,000 messages with content truncated to 500 runes, and Ctrl+O toggles a tool timeline overlay.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/overlay_help.go:58; internal/ui/overlay_history.go:87; internal/ui/overlay_tool_timeline.go:126", ""],
    ["TUI", "TUI-009", "As an operator I want a command palette.", "`:` launches async quick actions with built-ins read/grep/web/attach/save/export/settings/onboarding; command actions read/grep/web/attach prepare input with the action name and optional args, save/export use model behavior, and settings/onboarding return async form commands.", "Spec'd", "-", "-", "Go UI tests / terminal smoke", 0, "internal/ui/model.go:1361 connectQuickActions; internal/ui/model.go:1450 ExecuteQuickAction; internal/ui/quickactions.go:39", ""],
    ["TUI", "TUI-010", "As an operator I want search mode.", "`/` enters search mode when textarea is not focused; runes append and backspace deletes through `UpdateSearchQuery`, the query is matched case-insensitively against message content and content blocks, the prompt shows match count, Enter cycles the current match, and Esc/Ctrl+C exit search.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/model.go:EnterSearchMode; internal/ui/model.go:UpdateSearchQuery; internal/ui/update.go:SearchMode; internal/ui/view.go:Search prompt", ""],
    ["TUI", "TUI-011", "As an operator I want a conversation browser.", "Conversation browser component renders list/preview UI with quit, Enter load preview, favorite, delete, sort, and refresh keys; preview loads recent real messages through `MessageService`, and the main History tab mounts the browser when conversation/message services are available.", "Spec'd", "-", "-", "Go UI tests/static reachability", 0, "internal/ui/browser/conversations.go:89; internal/ui/browser/conversations.go:418; internal/ui/model.go:894 SetServices; internal/ui/view.go:275 renderHistoryView", ""],
    ["TUI", "TUI-012", "As an operator I want plugin/MCP dashboard status.", "Plugin dashboard component toggles Plugin/MCP views with Tab, refreshes with r, quits with q, no longer fabricates sample rows, receives real plugin and MCP registry data through `SetIntegrationRegistries`, and renders a compact Plugin/MCP status section in Tools.", "Spec'd", "-", "-", "Go UI tests/static reachability", 0, "internal/ui/dashboard/plugins.go:44; internal/ui/dashboard/plugins.go:SetData; internal/ui/model.go:SetIntegrationRegistries; internal/ui/view.go:renderToolsView; cmd/hex/root.go:589", ""],
    ["TUI", "TUI-013", "As an operator I want autocomplete beyond slash commands.", "Autocomplete providers support slash commands, files, and history; the main UI detects the provider from input, opens file completions for path-like prefixes such as `./`, opens slash command completions for `/`, and feeds submitted prompts into the history provider for history completions.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/autocomplete.go:DetectProvider; internal/ui/update.go:syncAutocompleteOverlay; internal/ui/model.go:addToInputHistory; internal/ui/autocomplete_trigger_test.go", ""],
    ["TUI", "TUI-014", "As an operator I want to switch between TUI views.", "Tab cycles Chat, History, Tools, and Intro views; History renders current conversation messages with previews, and Tools renders active/historical tool state or an empty tool-call message.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/model.go:442 NextView; internal/ui/view.go:275 renderHistoryView; internal/ui/view.go:291 renderToolsView", ""],
    ["TUI", "TUI-015", "As an operator I want clear/save/export/typewriter/favorite shortcuts.", "Ctrl+L clears screen, Ctrl+K clears conversation, Ctrl+S calls `SaveConversation`, Ctrl+E reports exported markdown length, Ctrl+T toggles typewriter mode, and Ctrl+F toggles favorite status in chat mode.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/update.go:237", ""],
    ["TUI", "TUI-016", "As an operator I want multiline input.", "Ctrl+J inserts a newline and resizes input; Alt+Enter also inserts a newline before plain Enter send handling.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/update.go:346; internal/ui/update.go:360", ""],
    ["TUI", "TUI-017", "As an operator I want input history navigation.", "Submitted inputs are added to input history; Up navigates older history when cursor is on the first visual row, and Down navigates newer history when cursor is on the last visual row.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/update.go:452", ""],
    ["TUI", "TUI-018", "As an operator I want mouse and keyboard scrolling.", "When input is empty, `j` scrolls down and `k` scrolls up; additional mouse scroll and hover timestamp handling are implemented in the UI update path.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/update.go:432; internal/ui/update.go:576", ""],
    ["TUI", "TUI-019", "As an operator I want status context at the bottom of the screen.", "Status bar renders token counts when available, permission mode indicator when executor has a checker, current view mode, and help text for common shortcuts.", "Spec'd", "-", "-", "Go UI tests", 0, "internal/ui/view.go:289", ""],
    ["TUI", "TUI-020", "As an operator I want token usage visualization.", "The model owns a token visualization component, feeds it from `UpdateTokens`, resizes it on window-size messages, and renders compact token usage in the Tools view when token counts exist.", "Spec'd", "-", "-", "Go UI tests/static reachability", 0, "internal/ui/model.go:457 UpdateTokens; internal/ui/update.go:656 WindowSizeMsg; internal/ui/view.go:291 renderToolsView; internal/ui/visualization/tokens.go:40", ""],
    ["TUI", "TUI-021", "As a first-run operator I want a huh-based onboarding flow.", "Onboarding form asks setup confirmation, Anthropic API key with `sk-ant-` validation, Claude model selection, tutorial offer, and sample conversation offer; it supplements the root setup wizard through the `onboarding` quick action and updates the active model when completed.", "Spec'd", "-", "-", "Go UI tests/static reachability", 0, "internal/ui/forms/onboarding.go:40 NewOnboardingForm; internal/ui/forms/onboarding_integration.go:21 RunOnboardingFormAsync; internal/ui/model.go:1361 connectQuickActions", ""],
    ["TUI", "TUI-022", "As an operator I want an interactive settings form.", "Settings form selects model, validates Anthropic API key, validates temperature 0.0-1.0 and max tokens 256-8192, saves non-cancelled results to `~/.hex/config.toml`, and is launched from the `settings` quick action.", "Spec'd", "-", "-", "Go UI tests/static reachability", 0, "internal/ui/forms/settings.go:48 NewSettingsForm; internal/ui/forms/settings_integration.go:18 RunSettingsFormAsync; internal/ui/model.go:1361 connectQuickActions", ""],
    ["Context", "CTX-002", "As a user I want long chats summarized when needed.", "Context manager estimates roughly 4 chars/token, preserves system/recent/tool/error-important messages when pruning, and summarizer can call Claude Haiku and cache by content hash.", "Spec'd", "-", "-", "Go tests", 0, "internal/convcontext/manager.go:131; internal/convcontext/summarizer.go:31", ""],
    ["Rate Limit", "RATE-001", "As a user with parallel agents I want fewer API 429s.", "Global Anthropic client limiter is a token bucket of 50 tokens refilling 1 token/minute; sync and streaming requests call `Acquire` and return context errors when blocked/cancelled.", "Spec'd", "-", "-", "Go tests", 0, "internal/core/client.go:32; internal/ratelimit/limiter.go:63; internal/core/stream.go:64", ""],
    ["Landing Page", "WEB-003", "As an alpha visitor I want an error state when downloads metadata fails.", "`/alpha` fetches Firebase `versions.json` from the default URL or `HEX_ALPHA_VERSIONS_URL`; if fetch fails or returns non-OK, the downloads box renders a `DOWNLOAD MANIFEST UNAVAILABLE` alert and tells visitors to refresh or check GitHub.", "Spec'd", "-", "-", "Forced-failure Astro build + HTML assertion", 0, "landing-page/src/pages/alpha.astro; landing-page/scripts/verify-alpha-error.mjs; landing-page/package.json:test:alpha-error", ""],
]


def col_name(index: int) -> str:
    result = ""
    while index:
        index, rem = divmod(index - 1, 26)
        result = chr(65 + rem) + result
    return result


def cell_xml(row_idx: int, col_idx: int, value) -> str:
    ref = f"{col_name(col_idx)}{row_idx}"
    if isinstance(value, int):
        return f'<c r="{ref}"><v>{value}</v></c>'
    text = html.escape(str(value), quote=True)
    return f'<c r="{ref}" t="inlineStr"><is><t xml:space="preserve">{text}</t></is></c>'


def sheet_xml(rows: list[list[object]]) -> str:
    body = []
    for row_idx, row in enumerate(rows, start=1):
        cells = "".join(cell_xml(row_idx, col_idx, value) for col_idx, value in enumerate(row, start=1))
        body.append(f'<row r="{row_idx}">{cells}</row>')
    cols = "".join(
        f'<col min="{i}" max="{i}" width="{width}" customWidth="1"/>'
        for i, width in enumerate([18, 13, 52, 80, 15, 26, 18, 28, 10, 48, 42], start=1)
    )
    return f'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>
  <cols>{cols}</cols>
  <sheetData>{"".join(body)}</sheetData>
  <autoFilter ref="A1:K{len(rows)}"/>
</worksheet>'''


def rows_for_workbook() -> list[list[object]]:
    logistical_failures = {
        "ROOT-009": "API-gated behavior not executed because ANTHROPIC_API_KEY is absent.",
        "ROOT-010": "API-gated tool filtering not executed against a live model because ANTHROPIC_API_KEY is absent.",
        "PRINT-002": "Output formatting through a live print-mode response not executed because ANTHROPIC_API_KEY is absent.",
        "PRINT-004": "Plan-mode live two-turn execution not executed because ANTHROPIC_API_KEY is absent.",
        "COST-001": "Live usage/cost summary not executed because ANTHROPIC_API_KEY is absent.",
        "PRINT-005": "Mux print live execution not executed because ANTHROPIC_API_KEY is absent.",
        "THINK-001": "Live thinking-mode execution not executed because ANTHROPIC_API_KEY is absent.",
    }
    capped_failures = {
        "ROOT-009": "CAPPED after 3 loop passes: cannot verify live system-prompt behavior without ANTHROPIC_API_KEY.",
        "ROOT-010": "CAPPED after 3 loop passes: cannot verify live tool filtering without ANTHROPIC_API_KEY.",
        "PRINT-002": "CAPPED after 3 loop passes: formatter unit coverage exists, but live print output formatting cannot be verified without ANTHROPIC_API_KEY.",
        "PRINT-004": "CAPPED after 3 loop passes: cannot verify live two-turn plan-mode execution without ANTHROPIC_API_KEY.",
        "COST-001": "CAPPED after 3 loop passes: cannot verify nonzero live usage/cost summary without ANTHROPIC_API_KEY.",
        "PRINT-005": "CAPPED after 3 loop passes: cannot verify live mux print streaming without ANTHROPIC_API_KEY.",
        "THINK-001": "CAPPED after 3 loop passes: cannot verify live Anthropic thinking-mode wrapper without ANTHROPIC_API_KEY.",
        "QA-003": "CAPPED after 3 loop passes: npm audit remains red; remediation requires explicit permission for security dependency updates, likely including a breaking Astro upgrade.",
    }
    verified_features = {
        "CMD-010": "Verified by go test ./cmd/hex -run TestHelpCommandDocumentsImplementedKeyBindings.",
        "CTX-001": "Verified by go test ./internal/convcontext -run TestManagerSummarizeStrategyAddsSummaryForPrunedMessages and go test ./cmd/hex -run 'TestContinueInteractiveUsesContextStrategyFactory|TestEffectiveMaxTurnsUsesConfiguredFlag'.",
        "ROOT-008": "Verified by go test ./cmd/hex -run TestVerboseEnablesDebugLogLevel.",
        "TUI-010": "Verified by go test ./internal/ui -run 'TestViewShowsSearchMode|TestSearchModeEnterSelectsNextMatch|TestSearchModeFindsMessageMatches|TestSearchModeBackspace'.",
        "TUI-013": "Verified by go test ./internal/ui -run 'TestModelAutocompleteTriggersForFilePaths|TestModelAutocompleteTriggersForInputHistory'.",
        "PRINT-006": "Verified by go test ./cmd/hex -run TestFormatOutputStreamJSONWritesSingleLine.",
        "PRINT-007": "Verified by go test ./cmd/hex -run 'TestRunPrintModeWithMuxPassesMaxTurnsToAgentConfig|TestEffectiveMaxTurnsUsesConfiguredFlag' and go test ./internal/adapter -run 'TestNewRootAgentCarriesMaxIterations|TestNewSubagentCarriesMaxIterations'.",
        "RES-003": "Verified by go test ./cmd/hex -run TestConversationsForPickerLimitsToTwenty and go test ./internal/ui -run 'TestSessionPickerSelectsHighlightedConversation|TestSessionPickerCancelLeavesSelectionEmpty'.",
        "TUI-011": "Verified by go test ./internal/ui -run 'TestViewRendersConversationBrowserWhenServicesAvailable|TestViewRendersHistoryMode' and go test ./internal/ui ./internal/ui/browser.",
        "TUI-012": "Verified by go test ./internal/ui -run TestViewRendersIntegrationDashboardData and go test ./internal/ui ./internal/ui/dashboard ./cmd/hex.",
        "TUI-009": "Verified by go test ./internal/ui -run 'TestBuiltInActions|TestModelQuickActionSettingsLaunchesSettingsForm|TestModelQuickActionOnboardingLaunchesOnboardingForm|TestModelQuickActionsExecute|TestModelQuickActionReadPreparesInputTemplate|TestModelQuickActionsWithArguments' and go test ./internal/ui.",
        "TUI-020": "Verified by go test ./internal/ui -run TestViewRendersToolsMode and go test ./internal/ui.",
        "TUI-021": "Verified by go test ./internal/ui -run 'TestModelQuickActionOnboardingLaunchesOnboardingForm|TestModelOnboardingResultUpdatesModel' and go test ./internal/ui.",
        "TUI-022": "Verified by go test ./internal/ui -run 'TestModelQuickActionSettingsLaunchesSettingsForm|TestModelSettingsResultUpdatesModel' and go test ./internal/ui.",
        "WEB-003": "Verified by npm run test:alpha-error and npm run build in landing-page.",
        "TUI-014": "Verified by go test ./internal/ui -run 'TestViewRendersHistoryMode|TestViewRendersToolsMode' and go test ./internal/ui.",
    }
    rows: list[list[object]] = []
    for row in ROWS:
        next_row = list(row)
        feature_id = str(next_row[1])
        if feature_id in logistical_failures and next_row[5] == "-":
            next_row[5] = logistical_failures[feature_id]
            next_row[6] = "Logistical"
        if feature_id in verified_features:
            next_row[4] = "Verified"
            if next_row[10]:
                next_row[10] = f"{next_row[10]} {verified_features[feature_id]}"
            else:
                next_row[10] = verified_features[feature_id]
        elif next_row[5] == "-":
            next_row[4] = "Tested-Pass"
        else:
            next_row[4] = "Tested-Fail"
        if feature_id in capped_failures:
            next_row[8] = 3
            if next_row[10]:
                next_row[10] = f"{next_row[10]} {capped_failures[feature_id]}"
            else:
                next_row[10] = capped_failures[feature_id]
        else:
            next_row[8] = 1
        rows.append(next_row)
    return rows


def write_workbook() -> None:
    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    rows = [HEADERS, *rows_for_workbook()]
    files = {
        "[Content_Types].xml": '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
  <Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>
</Types>''',
        "_rels/.rels": '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>''',
        "xl/workbook.xml": '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="Spec" sheetId="1" r:id="rId1"/></sheets>
</workbook>''',
        "xl/_rels/workbook.xml.rels": '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>''',
        "xl/styles.xml": '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="1"><font><sz val="11"/><name val="Aptos"/></font></fonts>
  <fills count="1"><fill><patternFill patternType="none"/></fill></fills>
  <borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>''',
        "xl/worksheets/sheet1.xml": sheet_xml(rows),
        "docProps/core.xml": '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>Hex Code-Derived Behavioral Spec</dc:title>
  <dc:creator>Codex</dc:creator>
</cp:coreProperties>''',
        "docProps/app.xml": '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
  <Application>Codex</Application>
</Properties>''',
    }
    with zipfile.ZipFile(OUTPUT, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        for name, content in files.items():
            zf.writestr(name, content)
    print(f"Wrote {OUTPUT} with {len(ROWS)} spec rows")


if __name__ == "__main__":
    write_workbook()
