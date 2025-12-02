package plugins

import (
	"testing"
)

func TestURLDetection(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		isGit  bool
		isHTTP bool
	}{
		{"GitHub HTTPS", "https://github.com/user/repo.git", true, true},
		{"GitHub HTTP", "https://github.com/user/repo", true, true},
		{"GitLab", "https://gitlab.com/user/repo.git", true, true},
		{"Bitbucket", "https://bitbucket.org/user/repo.git", true, true},
		{"SSH", "git@github.com:user/repo.git", true, false},
		{"Git protocol", "git://github.com/user/repo.git", true, false},
		{"HTTP tarball", "https://example.com/plugin.tar.gz", false, true},
		{"Local path", "./my-plugin", false, false},
		{"Absolute path", "/path/to/plugin", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGit := isGitURL(tt.url)
			if gotGit != tt.isGit {
				t.Errorf("isGitURL(%q) = %v, want %v", tt.url, gotGit, tt.isGit)
			}

			gotHTTP := isHTTPURL(tt.url)
			if gotHTTP != tt.isHTTP {
				t.Errorf("isHTTPURL(%q) = %v, want %v", tt.url, gotHTTP, tt.isHTTP)
			}
		})
	}
}

func TestExtractPluginName(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/user/my-plugin.git", "my-plugin"},
		{"https://github.com/user/my-plugin", "my-plugin"},
		{"git@github.com:user/my-plugin.git", "my-plugin"},
		{"https://example.com/some-plugin.git", "some-plugin"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := extractPluginNameFromURL(tt.url)
			if got != tt.want {
				t.Errorf("extractPluginNameFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestInstallerCreation(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := tmpDir + "/plugins"
	stateFile := tmpDir + "/state.json"

	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}

	installer := NewInstaller(loader)
	if installer == nil {
		t.Fatal("NewInstaller() returned nil")
	}
}

func TestInstallerLocalInstall(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := tmpDir + "/plugins"
	stateFile := tmpDir + "/state.json"

	// Create test plugin in temp location
	pluginName := "test-plugin"
	sourcePath := createTestPlugin(t, tmpDir, pluginName, "1.0.0", []string{"skills/test.md"})

	// Create installer
	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}
	installer := NewInstaller(loader)

	// Install from local path
	err = installer.Install(sourcePath)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Verify installation
	state := loader.State()
	if !state.IsInstalled(pluginName) {
		t.Error("Plugin should be installed")
	}
	if !state.IsEnabled(pluginName) {
		t.Error("Plugin should be enabled after install")
	}
}

func TestInstallerUninstall(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := tmpDir + "/plugins"
	stateFile := tmpDir + "/state.json"

	// Create and install test plugin
	pluginName := "test-plugin"
	sourcePath := createTestPlugin(t, tmpDir, pluginName, "1.0.0", []string{})

	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}
	installer := NewInstaller(loader)

	// Install
	if err := installer.Install(sourcePath); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Uninstall
	if err := installer.Uninstall(pluginName); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	// Verify uninstallation
	if loader.State().IsInstalled(pluginName) {
		t.Error("Plugin should not be installed after uninstall")
	}
}

func TestInstallerEnableDisable(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := tmpDir + "/plugins"
	stateFile := tmpDir + "/state.json"

	// Create and install test plugin
	pluginName := "test-plugin"
	sourcePath := createTestPlugin(t, tmpDir, pluginName, "1.0.0", []string{})

	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}
	installer := NewInstaller(loader)

	// Install
	if err := installer.Install(sourcePath); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Disable
	if err := installer.Disable(pluginName); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}
	if loader.State().IsEnabled(pluginName) {
		t.Error("Plugin should be disabled")
	}

	// Enable
	if err := installer.Enable(pluginName); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}
	if !loader.State().IsEnabled(pluginName) {
		t.Error("Plugin should be enabled")
	}
}

func TestInstallerInvalidOperations(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := tmpDir + "/plugins"
	stateFile := tmpDir + "/state.json"

	loader, _ := NewLoader(pluginsDir, stateFile)
	installer := NewInstaller(loader)

	t.Run("uninstall non-existent", func(t *testing.T) {
		err := installer.Uninstall("nonexistent")
		if err == nil {
			t.Error("Uninstall() should error on non-existent plugin")
		}
	})

	t.Run("enable non-existent", func(t *testing.T) {
		err := installer.Enable("nonexistent")
		if err == nil {
			t.Error("Enable() should error on non-existent plugin")
		}
	})

	t.Run("disable non-existent", func(t *testing.T) {
		err := installer.Disable("nonexistent")
		if err == nil {
			t.Error("Disable() should error on non-existent plugin")
		}
	})

	t.Run("update non-existent", func(t *testing.T) {
		err := installer.Update("nonexistent")
		if err == nil {
			t.Error("Update() should error on non-existent plugin")
		}
	})

	t.Run("install invalid path", func(t *testing.T) {
		err := installer.Install("/nonexistent/path")
		if err == nil {
			t.Error("Install() should error on invalid path")
		}
	})
}
