package plugins

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateOperations(t *testing.T) {
	state := &State{
		Plugins: make(map[string]*PluginState),
	}

	// Test AddPlugin
	t.Run("AddPlugin", func(t *testing.T) {
		state.AddPlugin("test-plugin", "1.0.0", "/path/to/plugin")
		if !state.IsInstalled("test-plugin") {
			t.Error("Plugin should be installed")
		}
		if !state.IsEnabled("test-plugin") {
			t.Error("Plugin should be enabled by default")
		}

		plugin, err := state.GetPlugin("test-plugin")
		if err != nil {
			t.Fatalf("GetPlugin() error = %v", err)
		}
		if plugin.Version != "1.0.0" {
			t.Errorf("Version = %q, want %q", plugin.Version, "1.0.0")
		}
		if plugin.Path != "/path/to/plugin" {
			t.Errorf("Path = %q, want %q", plugin.Path, "/path/to/plugin")
		}
	})

	// Test DisablePlugin
	t.Run("DisablePlugin", func(t *testing.T) {
		err := state.DisablePlugin("test-plugin")
		if err != nil {
			t.Fatalf("DisablePlugin() error = %v", err)
		}
		if state.IsEnabled("test-plugin") {
			t.Error("Plugin should be disabled")
		}
	})

	// Test EnablePlugin
	t.Run("EnablePlugin", func(t *testing.T) {
		err := state.EnablePlugin("test-plugin")
		if err != nil {
			t.Fatalf("EnablePlugin() error = %v", err)
		}
		if !state.IsEnabled("test-plugin") {
			t.Error("Plugin should be enabled")
		}
	})

	// Test UpdatePlugin
	t.Run("UpdatePlugin", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // Ensure timestamp is different
		err := state.UpdatePlugin("test-plugin", "2.0.0")
		if err != nil {
			t.Fatalf("UpdatePlugin() error = %v", err)
		}

		plugin, _ := state.GetPlugin("test-plugin")
		if plugin.Version != "2.0.0" {
			t.Errorf("Version = %q, want %q", plugin.Version, "2.0.0")
		}
		if plugin.Updated.IsZero() {
			t.Error("Updated timestamp should be set")
		}
	})

	// Test RemovePlugin
	t.Run("RemovePlugin", func(t *testing.T) {
		state.RemovePlugin("test-plugin")
		if state.IsInstalled("test-plugin") {
			t.Error("Plugin should not be installed")
		}
	})

	// Test operations on non-existent plugin
	t.Run("NonExistentPlugin", func(t *testing.T) {
		_, err := state.GetPlugin("nonexistent")
		if err == nil {
			t.Error("GetPlugin() should return error for non-existent plugin")
		}

		err = state.EnablePlugin("nonexistent")
		if err == nil {
			t.Error("EnablePlugin() should return error for non-existent plugin")
		}

		err = state.DisablePlugin("nonexistent")
		if err == nil {
			t.Error("DisablePlugin() should return error for non-existent plugin")
		}

		err = state.UpdatePlugin("nonexistent", "1.0.0")
		if err == nil {
			t.Error("UpdatePlugin() should return error for non-existent plugin")
		}
	})
}

func TestStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	// Create and save state
	state := &State{
		Plugins: make(map[string]*PluginState),
	}
	state.AddPlugin("plugin1", "1.0.0", "/path/to/plugin1")
	state.AddPlugin("plugin2", "2.0.0", "/path/to/plugin2")
	_ = state.DisablePlugin("plugin2") // Test doesn't need to check error for setup operation

	err := state.Save(stateFile)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load state
	loadedState, err := LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	// Verify loaded state
	if !loadedState.IsInstalled("plugin1") {
		t.Error("plugin1 should be installed")
	}
	if !loadedState.IsEnabled("plugin1") {
		t.Error("plugin1 should be enabled")
	}
	if loadedState.IsEnabled("plugin2") {
		t.Error("plugin2 should be disabled")
	}

	plugin1, _ := loadedState.GetPlugin("plugin1")
	if plugin1.Version != "1.0.0" {
		t.Errorf("plugin1 version = %q, want %q", plugin1.Version, "1.0.0")
	}
}

func TestLoadStateNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "nonexistent.json")

	// Should create empty state
	state, err := LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	if len(state.Plugins) != 0 {
		t.Errorf("New state should have 0 plugins, got %d", len(state.Plugins))
	}
}

func TestListMethods(t *testing.T) {
	state := &State{
		Plugins: make(map[string]*PluginState),
	}

	state.AddPlugin("plugin1", "1.0.0", "/path/1")
	state.AddPlugin("plugin2", "2.0.0", "/path/2")
	state.AddPlugin("plugin3", "3.0.0", "/path/3")
	_ = state.DisablePlugin("plugin2") // Test doesn't need to check error for setup operation

	t.Run("ListAll", func(t *testing.T) {
		all := state.ListAll()
		if len(all) != 3 {
			t.Errorf("ListAll() returned %d plugins, want 3", len(all))
		}
	})

	t.Run("ListEnabled", func(t *testing.T) {
		enabled := state.ListEnabled()
		if len(enabled) != 2 {
			t.Errorf("ListEnabled() returned %d plugins, want 2", len(enabled))
		}
		// Check that disabled plugin is not in list
		for _, name := range enabled {
			if name == "plugin2" {
				t.Error("plugin2 should not be in enabled list")
			}
		}
	})
}

func TestStateSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// State file in nested non-existent directory
	stateFile := filepath.Join(tmpDir, "nested", "dir", "state.json")

	state := &State{
		Plugins: make(map[string]*PluginState),
	}
	state.AddPlugin("test", "1.0.0", "/path")

	err := state.Save(stateFile)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("State file was not created")
	}
}
