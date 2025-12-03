# Hex - Product Roadmap

**Vision:** Production-ready Claude Code alternative in Go with full feature parity

---

## Phase 1: Foundation ✅ COMPLETE

**Duration:** 1 week (completed 2025-11-25)

**What Was Built:**
- CLI framework (Cobra)
- Configuration system (Viper + .env)
- Anthropic API client (non-streaming)
- Print mode (--print)
- Basic commands (setup-token, doctor)
- Comprehensive test suite

**Deliverables:**
- ✅ Working binary (`./hex`)
- ✅ 15+ passing tests
- ✅ Integration tests
- ✅ Documentation (README, PHASE1.md)

**What Works:**
```bash
./hex --version                      # Version info
./hex --print "test"                 # Non-interactive query
./hex --print --output-format json   # JSON output
./hex setup-token <key>              # Configure auth
./hex doctor                         # Health check
```

---

## Phase 2: Interactive Mode 🚧 NEXT

**Duration:** ~1 week (estimated)

**Goal:** Full-featured TUI with streaming, storage, and history

**Tasks:**
1. SQLite storage schema (4-6h)
2. Streaming API client (4-6h)
3. Bubbletea TUI (8-12h)
4. Conversation history (4-6h)
5. Background tasks (2-4h)

**Deliverables:**
- [ ] Interactive TUI with Bubbletea
- [ ] Streaming responses
- [ ] SQLite conversation storage
- [ ] --continue and --resume flags
- [ ] Status bar with metrics

**What Will Work:**
```bash
./hex                                # Interactive TUI
./hex --continue                     # Resume last conversation
./hex --resume                       # Pick conversation to resume
./hex --resume abc-123               # Resume specific conversation
```

**Tech Stack:**
- Bubbletea (TUI framework)
- Lipgloss (styling)
- Glamour (markdown rendering)
- modernc.org/sqlite (storage)

---

## Phase 3: Tool Execution 📋 PLANNED

**Duration:** 2-3 weeks (estimated)

**Goal:** Full tool execution with safety and permissions

**Tools to Implement:**
1. **Read** - Read files with line ranges
2. **Write** - Write files (must read first)
3. **Edit** - String replacement in files
4. **Bash** - Execute shell commands (sandboxed)
5. **Grep** - Search code with ripgrep
6. **Glob** - Find files by pattern

**Safety Features:**
- Permission system (prompt before execution)
- Sandboxed bash (--add-dir for access control)
- Workspace trust dialogs
- Tool allowlist/denylist (--allowed-tools, --disallowed-tools)

**Deliverables:**
- [ ] All 6 core tools implemented
- [ ] Permission system
- [ ] Sandbox enforcement
- [ ] Tool execution tests

**What Will Work:**
```bash
# In interactive mode, Claude can:
# - Read and edit files
# - Run bash commands
# - Search code
# - Find files by pattern
```

---

## Phase 4: MCP Integration 📋 PLANNED

**Duration:** 3-4 weeks (estimated)

**Goal:** Full MCP (Model Context Protocol) support

**Components:**
1. **MCP Client**
   - stdio transport (process-based)
   - sse transport (HTTP streaming)
   - http transport (request/response)

2. **Server Management**
   - Add/remove servers
   - Server health checks
   - Configuration scopes (user/project/local)

3. **MCP CLI**
   - Query live MCP state
   - List tools and resources
   - Invoke tools directly

**Deliverables:**
- [ ] MCP client implementation
- [ ] All 3 transports (stdio/sse/http)
- [ ] Server management commands
- [ ] MCP CLI mode (--mcp-cli)
- [ ] MCP state persistence

**What Will Work:**
```bash
# MCP server management
./hex mcp add myserver https://example.com
./hex mcp list
./hex mcp get myserver
./hex mcp remove myserver

# MCP CLI mode
./hex --mcp-cli servers
./hex --mcp-cli tools
./hex --mcp-cli call myserver/tool '{"arg":"value"}'
./hex --mcp-cli resources
./hex --mcp-cli read myserver/resource
```

---

## Phase 5: Plugin System 📋 PLANNED

**Duration:** 2-3 weeks (estimated)

**Goal:** Extensibility via plugins

**Features:**
1. **Plugin Loading**
   - Marketplace support
   - Local plugin directories
   - Plugin validation

2. **Plugin Capabilities**
   - Slash commands
   - Agent definitions
   - Skills
   - Lifecycle hooks

3. **Plugin Management**
   - Install/uninstall
   - Enable/disable
   - Update from marketplaces

**Deliverables:**
- [ ] Plugin loader
- [ ] Plugin manifest schema
- [ ] Plugin management commands
- [ ] Marketplace integration
- [ ] Plugin documentation

**What Will Work:**
```bash
# Plugin management
./hex plugin install superpowers
./hex plugin list
./hex plugin enable superpowers
./hex plugin disable superpowers
./hex plugin uninstall superpowers

# In session: slash commands from plugins
/brainstorm
/execute-plan
/write-plan
```

---

## Phase 6: Advanced Features 📋 FUTURE

**Duration:** 4-6 weeks (estimated)

**Features:**
1. **IDE Integration**
   - Auto-detect VS Code/Cursor/Windsurf
   - Enhanced file navigation
   - Editor integration

2. **Multi-Platform Support**
   - AWS Bedrock
   - Google Vertex AI
   - Anthropic Foundry

3. **Advanced Output**
   - JSON Schema validation
   - Structured output
   - Stream-JSON mode

4. **Session Management**
   - Fork sessions
   - Teleport (remote sessions)
   - Session sync across devices

**Deliverables:**
- [ ] IDE connection
- [ ] Multi-platform auth
- [ ] JSON Schema support
- [ ] Advanced session features

---

## Phase 7: Production Polish 📋 FUTURE

**Duration:** 2-3 weeks (estimated)

**Focus:** Performance, security, and UX

**Tasks:**
1. **Performance**
   - Optimize startup time
   - Reduce memory usage
   - Parallel MCP connections
   - Response caching

2. **Security**
   - Security audit
   - Secrets management
   - Input validation
   - Sandboxing hardening

3. **User Experience**
   - Better error messages
   - Onboarding flow
   - Interactive tutorials
   - Comprehensive help

4. **Documentation**
   - User guide
   - API documentation
   - Plugin development guide
   - Troubleshooting guide

**Deliverables:**
- [ ] Performance benchmarks
- [ ] Security audit report
- [ ] Complete user documentation
- [ ] Distribution packages

---

## Phase 8: Distribution 📋 FUTURE

**Duration:** 1-2 weeks (estimated)

**Goal:** Easy installation for all users

**Distribution Channels:**
1. **Binary Releases**
   - GitHub Releases
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

2. **Package Managers**
   - Homebrew (macOS/Linux)
   - apt (Debian/Ubuntu)
   - yum (RedHat/Fedora)
   - Chocolatey (Windows)
   - Scoop (Windows)

3. **Container Images**
   - Docker Hub
   - GitHub Container Registry
   - Multi-arch support

**Deliverables:**
- [ ] CI/CD for binary builds
- [ ] Homebrew formula
- [ ] APT/YUM repositories
- [ ] Docker images
- [ ] Installation scripts

---

## Complete Feature Comparison

### Claude Code (Original) vs Hex (Clean Implementation)

| Feature | Original | Phase | Status |
|---------|----------|-------|--------|
| **Core** | | | |
| Print mode | ✅ | 1 | ✅ Complete |
| Interactive TUI | ✅ | 2 | 📋 Next |
| Streaming responses | ✅ | 2 | 📋 Next |
| Model selection | ✅ | 1 | ✅ Complete |
| Configuration | ✅ | 1 | ✅ Complete |
| **Tools** | | | |
| Read | ✅ | 3 | 📋 Planned |
| Write | ✅ | 3 | 📋 Planned |
| Edit | ✅ | 3 | 📋 Planned |
| Bash | ✅ | 3 | 📋 Planned |
| Grep | ✅ | 3 | 📋 Planned |
| Glob | ✅ | 3 | 📋 Planned |
| **Storage** | | | |
| Conversation history | ✅ | 2 | 📋 Next |
| --continue | ✅ | 2 | 📋 Next |
| --resume | ✅ | 2 | 📋 Next |
| Session fork | ✅ | 6 | 📋 Future |
| **MCP** | | | |
| MCP client | ✅ | 4 | 📋 Planned |
| stdio transport | ✅ | 4 | 📋 Planned |
| sse transport | ✅ | 4 | 📋 Planned |
| http transport | ✅ | 4 | 📋 Planned |
| MCP CLI mode | ✅ | 4 | 📋 Planned |
| Server management | ✅ | 4 | 📋 Planned |
| **Plugins** | | | |
| Plugin system | ✅ | 5 | 📋 Planned |
| Marketplaces | ✅ | 5 | 📋 Planned |
| Slash commands | ✅ | 5 | 📋 Planned |
| Skills | ✅ | 5 | 📋 Planned |
| **Advanced** | | | |
| IDE integration | ✅ | 6 | 📋 Future |
| Multi-platform auth | ✅ | 6 | 📋 Future |
| JSON Schema | ✅ | 6 | 📋 Future |
| Structured output | ✅ | 6 | 📋 Future |

---

## Timeline Summary

| Phase | Duration | Status | Target |
|-------|----------|--------|--------|
| 1. Foundation | 1 week | ✅ Complete | Done |
| 2. Interactive | 1 week | 📋 Next | Week 2 |
| 3. Tools | 2-3 weeks | 📋 Planned | Weeks 3-5 |
| 4. MCP | 3-4 weeks | 📋 Planned | Weeks 6-9 |
| 5. Plugins | 2-3 weeks | 📋 Planned | Weeks 10-12 |
| 6. Advanced | 4-6 weeks | 📋 Future | Weeks 13-18 |
| 7. Polish | 2-3 weeks | 📋 Future | Weeks 19-21 |
| 8. Distribution | 1-2 weeks | 📋 Future | Weeks 22-23 |
| **Total** | **~6 months** | | |

---

## Success Metrics

### Phase Completion Criteria

**Each phase is complete when:**
- [ ] All planned features implemented
- [ ] All tests passing (unit + integration)
- [ ] Documentation updated
- [ ] No known critical bugs
- [ ] Performance acceptable
- [ ] Code reviewed

### Final Success Criteria

**Product is production-ready when:**
- [ ] Feature parity with Claude Code
- [ ] Performance ≥ original (startup, response time)
- [ ] Security audit passed
- [ ] Comprehensive test coverage (>80%)
- [ ] Complete user documentation
- [ ] Distribution packages available
- [ ] Active use by beta testers
- [ ] Positive user feedback

---

## Contributing

### For Phase 2 (Current)

**To implement Phase 2:**
1. Read `NEXT_STEPS.md` for task breakdown
2. Follow detailed plan in `docs/plans/2025-11-26-hex-phase2-interactive.md`
3. Use TDD approach (test first, implement to pass)
4. Submit each task for review before proceeding

**Preferred approach:**
```bash
claude /superpowers:execute-plan docs/plans/2025-11-26-hex-phase2-interactive.md
```

### For Future Phases

**Process:**
1. Create detailed implementation plan (like Phase 2)
2. Break down into bite-sized tasks
3. Write tests first for each task
4. Implement to pass tests
5. Review and refactor
6. Document as you go

---

## Questions?

**For technical questions:**
- Check `docs/` directory for design docs
- Review Phase 1 code for patterns
- Ask Claude for specific guidance

**For architectural decisions:**
- Refer to CLAUDE.md for clean-room principles
- Review `/Users/harper/Public/src/2389/cc-deobfuscate/spec/ARCHITECTURE.md` for original system
- Follow 4-layer architecture (Presentation → Application → Domain → Infrastructure)

---

**Last Updated:** 2025-11-27
**Current Phase:** Phase 2 (Interactive Mode)
**Next Milestone:** Working TUI with streaming responses
