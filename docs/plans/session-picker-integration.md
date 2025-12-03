# Session Picker Integration Plan

## Goal
Integrate the session picker into the startup flow so users can interactively select which conversation to resume.

## Current Behavior
- `--continue` flag: Resumes latest conversation
- `--resume <ID>` flag: Resumes specific conversation by ID
- No flags: Always starts new conversation

## New Behavior
When no flags are passed:
1. Check if there are any existing conversations
2. If none, start new conversation (current behavior)
3. If any exist, show interactive session picker
4. User selects conversation or chooses "New Session"
5. Continue with selected conversation

## Implementation

### Step 1: Create showSessionPicker() function

```go
// showSessionPicker displays an interactive picker for resuming conversations
// Returns the selected conversation ID, or empty string for new session
func showSessionPicker(db *sql.DB) (string, error) {
    // Get recent conversations (limit 50)
    conversations, err := storage.ListConversations(db, 50, 0)
    if err != nil {
        return "", fmt.Errorf("list conversations: %w", err)
    }

    // If no conversations, return empty (start new)
    if len(conversations) == 0 {
        return "", nil
    }

    // Create and run session picker
    picker := ui.NewSessionPicker(conversations)

    // Run as modal tea.Program
    p := tea.NewProgram(picker, tea.WithAltScreen())
    finalModel, err := p.Run()
    if err != nil {
        return "", fmt.Errorf("run session picker: %w", err)
    }

    // Extract result
    if result, ok := finalModel.(ui.SessionPicker); ok {
        if result.IsNewSession() {
            return "", nil  // User chose new session
        }
        return result.SelectedID(), nil
    }

    return "", fmt.Errorf("unexpected model type")
}
```

### Step 2: Integrate into runInteractive()

Add after database open (line ~156):

```go
// Task 7: Handle --continue or --resume flags, or show picker
if continueFlag {
    // ... existing continue logic ...
} else if resumeID != "" {
    // ... existing resume logic ...
} else {
    // NEW: Show session picker if no flags specified
    selectedID, err := showSessionPicker(db)
    if err != nil {
        logging.Warn("Failed to show session picker: %v, starting new session", err)
    } else if selectedID != "" {
        // User selected a conversation
        resumeID = selectedID  // Set resumeID to use existing resume logic
        // Fall through to resume logic below
    }
    // If selectedID is empty, user chose new session - continue as normal
}

// Now handle resumeID if it was set (either by flag or picker)
if resumeID != "" {
    // Existing resume logic...
}
```

### Step 3: Update storage package if needed

Check if `storage.ListConversations()` exists:
- If yes, use it
- If no, create it:

```go
// ListConversations returns recent conversations, ordered by updated_at DESC
func ListConversations(db *sql.DB, limit, offset int) ([]*Conversation, error) {
    query := `
        SELECT id, title, model, is_favorite, created_at, updated_at
        FROM conversations
        ORDER BY updated_at DESC
        LIMIT ? OFFSET ?
    `

    rows, err := db.Query(query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Scan()

    var conversations []*Conversation
    for rows.Next() {
        var conv Conversation
        err := rows.Scan(
            &conv.ID,
            &conv.Title,
            &conv.Model,
            &conv.IsFavorite,
            &conv.Created At,
            &conv.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        conversations = append(conversations, &conv)
    }

    return conversations, rows.Err()
}
```

## Testing Plan

### Manual Testing
1. **No conversations exist**:
   ```bash
   rm ~/.pagent/pagent.db
   ./pagent
   # Should start new session (current behavior)
   ```

2. **Conversations exist**:
   ```bash
   # Start and quit a conversation
   ./pagent
   > Hello
   > ctrl+c

   # Start again
   ./pagent
   # Should show session picker with list
   ```

3. **Select conversation**:
   - Arrow keys navigate
   - Enter selects
   - Resume that conversation

4. **New session**:
   - Navigate to "✨ New Session"
   - Enter selects
   - Start fresh conversation

5. **With flags** (should bypass picker):
   ```bash
   ./pagent --continue   # No picker, resume latest
   ./pagent --resume conv-123  # No picker, resume specific
   ```

### Edge Cases
- Empty database → start new
- Database error → warn, start new
- Picker error → warn, start new
- ESC in picker → new session
- Print mode → no picker (already handled)

## Success Criteria
✅ Picker shows when no flags and conversations exist
✅ Can select conversation with arrow keys + Enter
✅ Can choose "New Session"
✅ ESC creates new session
✅ Flags bypass picker (--continue, --resume)
✅ No conversations → no picker, start new
✅ Errors handled gracefully
✅ Works in tmux/screen

## Rollback
If issues arise, the feature can be disabled by:
1. Adding `--no-picker` flag
2. Or reverting the showSessionPicker() call
3. All existing flag behavior is preserved
