// ABOUTME: Session persistence layer for saving/loading conversation history
// Provides file-based JSON storage for sessions with CRUD operations and
// automatic directory management. Sessions are stored as individual JSON files
// in ~/.hex/sessions/ or a configurable directory.

package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SessionStorage manages persistence of Session objects to disk.
// Each session is stored as an individual JSON file named {id}.json.
type SessionStorage struct {
	dir string // Directory where session files are stored
}

// NewSessionStorage creates a new SessionStorage instance at the given directory.
// Creates the directory (and any parent directories) if it doesn't exist.
// Returns an error if the directory cannot be created.
func NewSessionStorage(dir string) (*SessionStorage, error) {
	// Expand ~ to home directory if present
	if strings.HasPrefix(dir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dir = filepath.Join(home, dir[1:])
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return &SessionStorage{dir: dir}, nil
}

// Save persists a session to disk as JSON.
// The session is stored in a file named {session.ID}.json with indented formatting
// for human readability. Returns an error if the session cannot be serialized
// or written to disk.
func (s *SessionStorage) Save(session *Session) error {
	if session == nil {
		return fmt.Errorf("cannot save nil session")
	}
	if session.ID == "" {
		return fmt.Errorf("cannot save session with empty ID")
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to file
	filename := filepath.Join(s.dir, session.ID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load retrieves a session by its ID from disk.
// Returns the session if found, or an error if the file doesn't exist
// or cannot be parsed.
func (s *SessionStorage) Load(id string) (*Session, error) {
	if id == "" {
		return nil, fmt.Errorf("cannot load session with empty ID")
	}

	filename := filepath.Join(s.dir, id+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %w", err)
	}

	return &session, nil
}

// List returns all sessions stored in the directory, sorted by UpdatedAt
// descending (newest first). Sessions that fail to parse are skipped with
// a warning printed to stderr. Returns an empty slice if no sessions exist.
func (s *SessionStorage) List() ([]*Session, error) {
	// Glob for all JSON files
	pattern := filepath.Join(s.dir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list session files: %w", err)
	}

	sessions := make([]*Session, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			// Skip files that can't be read
			fmt.Fprintf(os.Stderr, "warning: failed to read session file %s: %v\n", file, err)
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			// Skip files that can't be parsed
			fmt.Fprintf(os.Stderr, "warning: failed to parse session file %s: %v\n", file, err)
			continue
		}

		sessions = append(sessions, &session)
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// Delete removes a session file from disk by its ID.
// Returns an error if the file doesn't exist or cannot be deleted.
func (s *SessionStorage) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("cannot delete session with empty ID")
	}

	filename := filepath.Join(s.dir, id+".json")
	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session not found: %s", id)
		}
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}
