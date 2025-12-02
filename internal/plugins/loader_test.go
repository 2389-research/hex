package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestFile(t *testing.T, path, content string) {
	t.Helper()
	//nolint:gosec // G301 - test helper uses standard directory permissions
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	//nolint:gosec // G306 - test helper uses standard file permissions
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func createTestPlugin(t *testing.T, dir, name, version string, skills []string) string {
	t.Helper()

	pluginDir := filepath.Join(dir, name)
	//nolint:gosec // G301 - test helper uses standard directory permissions
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create manifest
	manifest := &Manifest{
		Name:        name,
		Version:     version,
		Description: "Test plugin",
		Skills:      skills,
	}

	manifestPath := filepath.Join(pluginDir, "plugin.json")
	if err := manifest.Save(manifestPath); err != nil {
		t.Fatal(err)
	}

	// Create skill files
	for _, skill := range skills {
		skillPath := filepath.Join(pluginDir, skill)
		createTestFile(t, skillPath, "# Test Skill\nTest content")
	}

	return pluginDir
}

func TestLoaderDiscoverAll(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")
	stateFile := filepath.Join(tmpDir, "state.json")

	// Create test plugins
	createTestPlugin(t, pluginsDir, "plugin1", "1.0.0", []string{"skills/test.md"})
	createTestPlugin(t, pluginsDir, "plugin2", "2.0.0", []string{})

	// Create loader
	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}

	// Mark plugin1 as installed and enabled
	loader.State().AddPlugin("plugin1", "1.0.0", filepath.Join(pluginsDir, "plugin1"))

	// Discover all plugins
	plugins, err := loader.DiscoverAll()
	if err != nil {
		t.Fatalf("DiscoverAll() error = %v", err)
	}

	if len(plugins) != 2 {
		t.Fatalf("DiscoverAll() found %d plugins, want 2", len(plugins))
	}

	// Check plugin1
	var plugin1 *Plugin
	for _, p := range plugins {
		if p.Name == "plugin1" {
			plugin1 = p
			break
		}
	}
	if plugin1 == nil {
		t.Fatal("plugin1 not found")
	}
	if !plugin1.Enabled {
		t.Error("plugin1 should be enabled")
	}
	if !plugin1.Installed {
		t.Error("plugin1 should be marked as installed")
	}
}

func TestLoaderLoadEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")
	stateFile := filepath.Join(tmpDir, "state.json")

	// Create test plugins
	createTestPlugin(t, pluginsDir, "plugin1", "1.0.0", []string{})
	createTestPlugin(t, pluginsDir, "plugin2", "2.0.0", []string{})

	// Create loader
	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}

	// Mark only plugin1 as installed and enabled
	loader.State().AddPlugin("plugin1", "1.0.0", filepath.Join(pluginsDir, "plugin1"))

	// Load enabled plugins
	plugins, err := loader.LoadEnabled(nil)
	if err != nil {
		t.Fatalf("LoadEnabled() error = %v", err)
	}

	if len(plugins) != 1 {
		t.Fatalf("LoadEnabled() found %d plugins, want 1", len(plugins))
	}

	if plugins[0].Name != "plugin1" {
		t.Errorf("Plugin name = %q, want %q", plugins[0].Name, "plugin1")
	}
}

func TestLoaderValidatePlugin(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")
	stateFile := filepath.Join(tmpDir, "state.json")

	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}

	t.Run("valid plugin", func(t *testing.T) {
		createTestPlugin(t, pluginsDir, "valid", "1.0.0", []string{"skills/test.md"})
		pluginDir := filepath.Join(pluginsDir, "valid")

		err := loader.ValidatePlugin(pluginDir)
		if err != nil {
			t.Errorf("ValidatePlugin() error = %v", err)
		}
	})

	t.Run("missing manifest", func(t *testing.T) {
		noManifestDir := filepath.Join(pluginsDir, "no-manifest")
		//nolint:gosec // G301 - test helper uses standard directory permissions
		if err := os.MkdirAll(noManifestDir, 0755); err != nil {
			t.Fatal(err)
		}

		err := loader.ValidatePlugin(noManifestDir)
		if err == nil {
			t.Error("ValidatePlugin() should error on missing manifest")
		}
	})

	t.Run("missing skill file", func(t *testing.T) {
		badPluginDir := filepath.Join(pluginsDir, "bad-plugin")
		//nolint:gosec // G301 - test helper uses standard directory permissions
		if err := os.MkdirAll(badPluginDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create manifest referencing non-existent skill
		manifest := &Manifest{
			Name:        "bad-plugin",
			Version:     "1.0.0",
			Description: "Bad plugin",
			Skills:      []string{"skills/nonexistent.md"},
		}
		manifestPath := filepath.Join(badPluginDir, "plugin.json")
		if err := manifest.Save(manifestPath); err != nil {
			t.Fatal(err)
		}

		err := loader.ValidatePlugin(badPluginDir)
		if err == nil {
			t.Error("ValidatePlugin() should error on missing skill file")
		}
	})
}

func TestLoaderGetPlugin(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")
	stateFile := filepath.Join(tmpDir, "state.json")

	// Create test plugin
	createTestPlugin(t, pluginsDir, "test-plugin", "1.0.0", []string{})

	// Create loader
	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}

	// Add to state
	pluginPath := filepath.Join(pluginsDir, "test-plugin")
	loader.State().AddPlugin("test-plugin", "1.0.0", pluginPath)

	// Get plugin
	plugin, err := loader.GetPlugin("test-plugin")
	if err != nil {
		t.Fatalf("GetPlugin() error = %v", err)
	}

	if plugin.Name != "test-plugin" {
		t.Errorf("Name = %q, want %q", plugin.Name, "test-plugin")
	}
	if plugin.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", plugin.Version, "1.0.0")
	}
}

func TestLoaderWithActivation(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")
	stateFile := filepath.Join(tmpDir, "state.json")

	// Create plugin with activation rules
	pluginDir := filepath.Join(pluginsDir, "go-plugin")
	//nolint:gosec // G301 - test helper uses standard directory permissions
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifest := &Manifest{
		Name:        "go-plugin",
		Version:     "1.0.0",
		Description: "Go development plugin",
		Activation: &ActivationConfig{
			Languages: []string{"go"},
		},
	}
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	if err := manifest.Save(manifestPath); err != nil {
		t.Fatal(err)
	}

	// Create loader
	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}

	// Enable plugin
	loader.State().AddPlugin("go-plugin", "1.0.0", pluginDir)

	// Test with matching context
	t.Run("matching context", func(t *testing.T) {
		context := &ActivationContext{
			Languages: []string{"go", "python"},
		}
		plugins, err := loader.LoadEnabled(context)
		if err != nil {
			t.Fatalf("LoadEnabled() error = %v", err)
		}
		if len(plugins) != 1 {
			t.Errorf("LoadEnabled() found %d plugins, want 1", len(plugins))
		}
	})

	// Test with non-matching context
	t.Run("non-matching context", func(t *testing.T) {
		context := &ActivationContext{
			Languages: []string{"python", "javascript"},
		}
		plugins, err := loader.LoadEnabled(context)
		if err != nil {
			t.Fatalf("LoadEnabled() error = %v", err)
		}
		if len(plugins) != 0 {
			t.Errorf("LoadEnabled() found %d plugins, want 0", len(plugins))
		}
	})
}

func TestLoaderState(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")
	stateFile := filepath.Join(tmpDir, "state.json")

	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}

	// Test State() method
	state := loader.State()
	if state == nil {
		t.Fatal("State() returned nil")
	}

	// Test SaveState()
	state.AddPlugin("test", "1.0.0", "/path")
	if err := loader.SaveState(); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	// Reload and verify
	loader2, _ := NewLoader(pluginsDir, stateFile)
	if !loader2.State().IsInstalled("test") {
		t.Error("Plugin should be installed after reload")
	}
}

func TestGetPluginsDir(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")
	stateFile := filepath.Join(tmpDir, "state.json")

	loader, _ := NewLoader(pluginsDir, stateFile)

	got := loader.GetPluginsDir()
	if got != pluginsDir {
		t.Errorf("GetPluginsDir() = %q, want %q", got, pluginsDir)
	}
}
