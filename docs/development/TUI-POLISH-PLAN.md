# TUI Polish Implementation Plan - "The Incredible Clem"

**Goal**: Comprehensive TUI polish using Charm ecosystem with Dracula theme

**Design Decisions**:
- Theme: Dracula (rich & colorful)
- Forms: Huh for ALL prompts (complete replacement)
- Components: Tables, Progress bars, Lists everywhere
- Layout: Single pane with polished modes
- Scope: Comprehensive upgrade

---

## Task 1: Foundation - Add Dependencies & Dracula Theme

**Objective**: Add huh dependency and create Dracula theme configuration

**Implementation**:
1. Add `github.com/charmbracelet/huh` to go.mod
2. Create `internal/ui/theme/dracula.go` with Dracula color palette
3. Define color constants matching official Dracula spec:
   - Background: #282a36
   - Current Line: #44475a
   - Foreground: #f8f8f2
   - Comment: #6272a4
   - Cyan: #8be9fd
   - Green: #50fa7b
   - Orange: #ffb86c
   - Pink: #ff79c6
   - Purple: #bd93f9
   - Red: #ff5555
   - Yellow: #f1fa8c
4. Create theme struct with lipgloss styles for all UI elements
5. Export `DraculaTheme()` function

**Tests**:
- Test theme struct creation
- Test color constants match spec
- Test all styles are properly initialized

**Files to create**:
- `internal/ui/theme/dracula.go`
- `internal/ui/theme/dracula_test.go`

**Success criteria**:
- huh added to dependencies
- Dracula theme module with all colors defined
- All tests passing

---

## Task 2: Update Existing Styles to Dracula

**Objective**: Replace all hardcoded colors in existing UI with Dracula theme

**Implementation**:
1. Update `internal/ui/view.go` - replace all color codes with theme colors
2. Update `internal/ui/statusbar.go` - apply Dracula palette
3. Update `internal/ui/suggestions.go` - use theme colors
4. Update `internal/ui/quickactions.go` - apply theme
5. Update `internal/ui/approval.go` - use theme colors
6. Add gradient support using lipgloss for title bar and borders
7. Initialize theme in Model.Init()

**Tests**:
- Existing tests should still pass
- Visual regression tests (if applicable)

**Files to modify**:
- `internal/ui/view.go`
- `internal/ui/statusbar.go`
- `internal/ui/suggestions.go`
- `internal/ui/quickactions.go`
- `internal/ui/approval.go`
- `internal/ui/model.go`

**Success criteria**:
- All UI elements use Dracula colors
- Gradients applied to title/borders
- No hardcoded color values remain
- All tests passing

---

## Task 3: Huh Forms - Tool Approval Replacement

**Objective**: Replace custom tool approval prompt with huh confirm/select

**Implementation**:
1. Create `internal/ui/forms/approval.go`
2. Implement `ToolApprovalForm` using huh.Form with:
   - Confirm field for yes/no approval
   - Select field for approval options (approve, deny, always allow, never allow)
   - Display tool name, parameters, and risk level
3. Replace approval logic in `internal/ui/approval.go`
4. Integrate with existing Model update/view cycle
5. Handle form submission and pass result to tool execution

**Tests**:
- Test form creation with tool data
- Test approval flow (approve/deny)
- Test always allow/never allow options
- Integration test with actual tool approval

**Files to create**:
- `internal/ui/forms/approval.go`
- `internal/ui/forms/approval_test.go`

**Files to modify**:
- `internal/ui/approval.go`
- `internal/ui/model.go`
- `internal/ui/update.go`

**Success criteria**:
- Tool approval uses huh forms
- All approval options work correctly
- Smoother, more beautiful approval UX
- All tests passing

---

## Task 4: Huh Forms - Quick Actions Replacement

**Objective**: Replace quick actions modal with huh select

**Implementation**:
1. Create `internal/ui/forms/quickactions.go`
2. Implement `QuickActionsForm` using huh.Select with:
   - Filterable list of actions
   - Categories (Tools, Navigation, Settings)
   - Key bindings displayed
   - Search/filter support
3. Replace quick actions logic in `internal/ui/quickactions.go`
4. Integrate with Model
5. Add keyboard shortcuts display

**Tests**:
- Test form creation with actions
- Test filtering/search
- Test action execution
- Test keyboard navigation

**Files to create**:
- `internal/ui/forms/quickactions.go`
- `internal/ui/forms/quickactions_test.go`

**Files to modify**:
- `internal/ui/quickactions.go`
- `internal/ui/model.go`
- `internal/ui/update.go`

**Success criteria**:
- Quick actions uses huh select
- Search/filter works smoothly
- Better visual presentation
- All tests passing

---

## Task 5: Huh Forms - Settings Wizard

**Objective**: Create interactive settings wizard with huh

**Implementation**:
1. Create `internal/ui/forms/settings.go`
2. Implement multi-step settings form with:
   - Model selection (Select from available models)
   - API key configuration (Input with masking)
   - Preferences (temperature, max tokens, etc.)
   - Theme selection (if adding multiple themes)
3. Add settings command/keybinding (Ctrl+,)
4. Save settings to config file
5. Reload settings without restart

**Tests**:
- Test form creation
- Test multi-step navigation
- Test settings persistence
- Test validation

**Files to create**:
- `internal/ui/forms/settings.go`
- `internal/ui/forms/settings_test.go`

**Files to modify**:
- `internal/ui/model.go`
- `internal/ui/update.go`
- `cmd/clem/root.go` (add settings command)

**Success criteria**:
- Full settings wizard working
- Settings persist correctly
- Accessible via Ctrl+, keybinding
- All tests passing

---

## Task 6: Huh Forms - Onboarding Flow

**Objective**: Create first-run onboarding experience

**Implementation**:
1. Create `internal/ui/forms/onboarding.go`
2. Implement onboarding flow:
   - Welcome screen
   - API key setup
   - Model selection
   - Quick tutorial
   - Sample conversation offer
3. Detect first run (check for config file)
4. Show onboarding before main UI on first run
5. Add "show tutorial" option to help menu

**Tests**:
- Test first-run detection
- Test onboarding completion
- Test skip option
- Test tutorial display

**Files to create**:
- `internal/ui/forms/onboarding.go`
- `internal/ui/forms/onboarding_test.go`

**Files to modify**:
- `cmd/clem/interactive.go`
- `internal/ui/model.go`

**Success criteria**:
- Smooth first-run experience
- Users can set up API key easily
- Tutorial helpful and skippable
- All tests passing

---

## Task 7: Bubbles Components - Table for Results

**Objective**: Add table component for displaying structured data

**Implementation**:
1. Create `internal/ui/components/table.go`
2. Wrap bubbles.Table with Dracula styling
3. Implement table for:
   - Tool execution results (when tools return structured data)
   - Conversation metadata view
   - Plugin registry display
   - MCP server status
4. Add table view mode (accessible via key binding)
5. Support sorting, filtering

**Tests**:
- Test table creation with data
- Test sorting
- Test styling applied
- Test view mode switching

**Files to create**:
- `internal/ui/components/table.go`
- `internal/ui/components/table_test.go`

**Files to modify**:
- `internal/ui/model.go` (add table state)
- `internal/ui/view.go` (render tables)
- `internal/ui/update.go` (handle table interactions)

**Success criteria**:
- Tables display structured data beautifully
- Sorting and navigation work
- Dracula theme applied
- All tests passing

---

## Task 8: Bubbles Components - Progress Bars

**Objective**: Add progress indicators throughout UI

**Implementation**:
1. Create `internal/ui/components/progress.go`
2. Wrap bubbles.Progress with Dracula styling
3. Add progress bars for:
   - API streaming (show chunks received)
   - Long-running tool execution
   - Token usage (visual context fill indicator)
   - Batch operations
4. Integrate with existing streaming logic
5. Add to status bar for context usage

**Tests**:
- Test progress bar creation
- Test progress updates
- Test completion states
- Test styling

**Files to create**:
- `internal/ui/components/progress.go`
- `internal/ui/components/progress_test.go`

**Files to modify**:
- `internal/ui/streaming.go`
- `internal/ui/statusbar.go`
- `internal/ui/model.go`
- `internal/ui/view.go`

**Success criteria**:
- Progress bars show during operations
- Token usage has visual indicator
- Smooth animations
- All tests passing

---

## Task 9: Bubbles Components - List for Browsing

**Objective**: Add list component for all browsing interfaces

**Implementation**:
1. Create `internal/ui/components/list.go`
2. Wrap bubbles.List with Dracula styling
3. Implement lists for:
   - Conversation history browser
   - Skills selector
   - Commands selector
   - Plugin browser
4. Add fuzzy search/filtering
5. Add status indicators and metadata

**Tests**:
- Test list creation
- Test filtering
- Test selection
- Test navigation

**Files to create**:
- `internal/ui/components/list.go`
- `internal/ui/components/list_test.go`

**Files to modify**:
- `internal/ui/model.go`
- `internal/ui/view.go`
- `internal/ui/update.go`

**Success criteria**:
- Lists work for all browsing
- Fuzzy search functional
- Beautiful Dracula styling
- All tests passing

---

## Task 10: Bubbles Components - Help System

**Objective**: Add comprehensive help component with key bindings

**Implementation**:
1. Create `internal/ui/components/help.go`
2. Use bubbles.Help with Dracula theme
3. Implement context-aware help:
   - Show different keys based on current mode
   - Group bindings by category
   - Expandable/collapsible sections
4. Add help overlay (? key)
5. Add help view mode

**Tests**:
- Test help text generation
- Test context switching
- Test key binding display
- Test styling

**Files to create**:
- `internal/ui/components/help.go`
- `internal/ui/components/help_test.go`

**Files to modify**:
- `internal/ui/model.go`
- `internal/ui/view.go`
- `internal/ui/update.go`

**Success criteria**:
- Comprehensive help system
- Context-aware bindings
- Easy to access and navigate
- All tests passing

---

## Task 11: Visual Polish - Gradients & Animations

**Objective**: Add visual flair with gradients and smooth transitions

**Implementation**:
1. Create `internal/ui/animations/gradient.go`
2. Add gradient support for:
   - Title bar (horizontal gradient)
   - Section dividers
   - Focus indicators
3. Add smooth transitions:
   - View mode switching
   - Status changes
   - Loading states
4. Use lipgloss adaptive colors for terminal compatibility

**Tests**:
- Test gradient generation
- Test color interpolation
- Test terminal compatibility

**Files to create**:
- `internal/ui/animations/gradient.go`
- `internal/ui/animations/gradient_test.go`

**Files to modify**:
- `internal/ui/view.go`
- `internal/ui/model.go`

**Success criteria**:
- Beautiful gradients throughout
- Smooth visual transitions
- Works in different terminals
- All tests passing

---

## Task 12: Visual Polish - Borders & Spacing

**Objective**: Improve layout with better borders and spacing

**Implementation**:
1. Create `internal/ui/layout/borders.go`
2. Define consistent border styles using Dracula colors
3. Improve spacing between elements:
   - Consistent padding
   - Better margins
   - Proper alignment
4. Add visual separators for sections
5. Ensure responsive layout for different terminal sizes

**Tests**:
- Test border rendering
- Test spacing calculations
- Test responsive behavior

**Files to create**:
- `internal/ui/layout/borders.go`
- `internal/ui/layout/borders_test.go`

**Files to modify**:
- `internal/ui/view.go`
- `internal/ui/model.go`

**Success criteria**:
- Clean, professional layout
- Consistent spacing throughout
- Responsive to terminal size
- All tests passing

---

## Task 13: Feature - Conversation Browser

**Objective**: Build conversation history browser with fuzzy search

**Implementation**:
1. Create `internal/ui/browser/conversations.go`
2. Use list component from Task 9
3. Implement:
   - Load conversations from database
   - Display with metadata (date, tokens, model)
   - Fuzzy search by content/title
   - Sort by date, tokens, favorite
   - Preview pane for selected conversation
4. Add keybinding to open browser (Ctrl+O)
5. Allow loading conversation into current session

**Tests**:
- Test conversation loading
- Test search functionality
- Test sorting
- Test loading into session

**Files to create**:
- `internal/ui/browser/conversations.go`
- `internal/ui/browser/conversations_test.go`

**Files to modify**:
- `internal/ui/model.go`
- `internal/ui/view.go`
- `internal/ui/update.go`

**Success criteria**:
- Easy browsing of past conversations
- Fast fuzzy search
- Smooth loading into session
- All tests passing

---

## Task 14: Feature - Plugin/MCP Dashboard

**Objective**: Create status dashboard for plugins and MCP servers

**Implementation**:
1. Create `internal/ui/dashboard/plugins.go`
2. Use table component from Task 7
3. Display:
   - Installed plugins with status (enabled/disabled)
   - MCP servers with connection status
   - Quick actions (enable/disable, configure)
4. Add dashboard view mode (Ctrl+P)
5. Show plugin metadata and health

**Tests**:
- Test plugin listing
- Test MCP server status
- Test enable/disable actions
- Test view rendering

**Files to create**:
- `internal/ui/dashboard/plugins.go`
- `internal/ui/dashboard/plugins_test.go`

**Files to modify**:
- `internal/ui/model.go`
- `internal/ui/view.go`
- `internal/ui/update.go`

**Success criteria**:
- Clear overview of plugins/MCP
- Easy management from UI
- Status indicators accurate
- All tests passing

---

## Task 15: Feature - Real-time Token Visualization

**Objective**: Add comprehensive token usage visualization

**Implementation**:
1. Create `internal/ui/visualization/tokens.go`
2. Use progress component from Task 8
3. Implement:
   - Real-time context window fill indicator
   - Input vs output token breakdown
   - Historical usage graph (sparkline)
   - Warning at 80% capacity
4. Integrate into status bar
5. Add detailed view (Ctrl+T)

**Tests**:
- Test token tracking
- Test visualization updates
- Test warning thresholds
- Test detailed view

**Files to create**:
- `internal/ui/visualization/tokens.go`
- `internal/ui/visualization/tokens_test.go`

**Files to modify**:
- `internal/ui/statusbar.go`
- `internal/ui/model.go`
- `internal/ui/view.go`

**Success criteria**:
- Clear token usage visualization
- Warnings help prevent context overflow
- Historical view informative
- All tests passing

---

## Task 16: Integration & Testing

**Objective**: Ensure all components work together seamlessly

**Implementation**:
1. Integration tests for:
   - Theme applied consistently across all components
   - Forms integrate with existing workflows
   - Components update in sync
   - No visual glitches during transitions
2. Performance testing:
   - Large conversation rendering
   - Table with many rows
   - List filtering performance
3. Terminal compatibility testing:
   - Test in different terminal emulators
   - Test with different color settings
   - Test with various terminal sizes
4. Fix any integration issues found

**Tests**:
- Full integration test suite
- Performance benchmarks
- Terminal compatibility matrix

**Files to create**:
- `internal/ui/integration_test.go`
- `internal/ui/performance_test.go`

**Success criteria**:
- All components work together
- No performance regressions
- Works in all major terminals
- All tests passing

---

## Success Criteria for Complete Plan

- ✅ Dracula theme applied throughout
- ✅ All prompts use huh forms
- ✅ Tables display all structured data
- ✅ Progress bars for all waiting states
- ✅ Lists for all browsing interfaces
- ✅ Comprehensive help system
- ✅ Beautiful gradients and transitions
- ✅ Professional layout with good spacing
- ✅ Conversation browser with search
- ✅ Plugin/MCP dashboard
- ✅ Real-time token visualization
- ✅ All tests passing (>90% coverage)
- ✅ Performance acceptable
- ✅ Works in all major terminals

## Estimated Scope

- **16 tasks** total
- **~30-40 new files**
- **~20 modified files**
- **~5,000-7,000 lines of code**
- **~3,000-4,000 lines of tests**
