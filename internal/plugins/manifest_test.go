package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManifestValidation(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *Manifest
		wantError bool
	}{
		{
			name: "valid manifest",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "1.0.0",
				Description: "A test plugin",
			},
			wantError: false,
		},
		{
			name: "missing name",
			manifest: &Manifest{
				Version:     "1.0.0",
				Description: "A test plugin",
			},
			wantError: true,
		},
		{
			name: "missing version",
			manifest: &Manifest{
				Name:        "test-plugin",
				Description: "A test plugin",
			},
			wantError: true,
		},
		{
			name: "missing description",
			manifest: &Manifest{
				Name:    "test-plugin",
				Version: "1.0.0",
			},
			wantError: true,
		},
		{
			name: "invalid name with uppercase",
			manifest: &Manifest{
				Name:        "TestPlugin",
				Version:     "1.0.0",
				Description: "A test plugin",
			},
			wantError: true,
		},
		{
			name: "invalid name with spaces",
			manifest: &Manifest{
				Name:        "test plugin",
				Version:     "1.0.0",
				Description: "A test plugin",
			},
			wantError: true,
		},
		{
			name: "invalid version",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "1.0",
				Description: "A test plugin",
			},
			wantError: true,
		},
		{
			name: "valid version with v prefix",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "v1.0.0",
				Description: "A test plugin",
			},
			wantError: false,
		},
		{
			name: "valid version with pre-release",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "1.0.0-beta.1",
				Description: "A test plugin",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestLoadManifest(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create valid manifest
	validManifest := `{
  "name": "test-plugin",
  "version": "1.0.0",
  "description": "A test plugin",
  "author": "Test Author",
  "license": "MIT"
}`

	validPath := filepath.Join(tmpDir, "valid.json")
	//nolint:gosec // G306 - test helper uses standard file permissions
	if err := os.WriteFile(validPath, []byte(validManifest), 0644); err != nil {
		t.Fatal(err)
	}

	// Create invalid JSON
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	//nolint:gosec // G306 - test helper uses standard file permissions
	if err := os.WriteFile(invalidPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create invalid manifest (missing required fields)
	incompleteManifest := `{
  "name": "test-plugin"
}`
	incompletePath := filepath.Join(tmpDir, "incomplete.json")
	//nolint:gosec // G306 - test helper uses standard file permissions
	if err := os.WriteFile(incompletePath, []byte(incompleteManifest), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		path      string
		wantError bool
		checkFunc func(*testing.T, *Manifest)
	}{
		{
			name:      "valid manifest",
			path:      validPath,
			wantError: false,
			checkFunc: func(t *testing.T, m *Manifest) {
				if m.Name != "test-plugin" {
					t.Errorf("Name = %q, want %q", m.Name, "test-plugin")
				}
				if m.Version != "1.0.0" {
					t.Errorf("Version = %q, want %q", m.Version, "1.0.0")
				}
				if m.Author != "Test Author" {
					t.Errorf("Author = %q, want %q", m.Author, "Test Author")
				}
			},
		},
		{
			name:      "invalid JSON",
			path:      invalidPath,
			wantError: true,
		},
		{
			name:      "incomplete manifest",
			path:      incompletePath,
			wantError: true,
		},
		{
			name:      "non-existent file",
			path:      filepath.Join(tmpDir, "nonexistent.json"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := LoadManifest(tt.path)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadManifest() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.checkFunc != nil {
				tt.checkFunc(t, manifest)
			}
		})
	}
}

func TestManifestShouldActivate(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		context  *ActivationContext
		want     bool
	}{
		{
			name: "no activation rules - always active",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "1.0.0",
				Description: "Test",
			},
			context: nil,
			want:    true,
		},
		{
			name: "onStartup - always active",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "1.0.0",
				Description: "Test",
				Activation: &ActivationConfig{
					OnStartup: true,
				},
			},
			context: nil,
			want:    true,
		},
		{
			name: "language match",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "1.0.0",
				Description: "Test",
				Activation: &ActivationConfig{
					Languages: []string{"go", "python"},
				},
			},
			context: &ActivationContext{
				Languages: []string{"go", "javascript"},
			},
			want: true,
		},
		{
			name: "language no match",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "1.0.0",
				Description: "Test",
				Activation: &ActivationConfig{
					Languages: []string{"go", "python"},
				},
			},
			context: &ActivationContext{
				Languages: []string{"javascript", "typescript"},
			},
			want: false,
		},
		{
			name: "file pattern match",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "1.0.0",
				Description: "Test",
				Activation: &ActivationConfig{
					Files: []string{"go.mod", "package.json"},
				},
			},
			context: &ActivationContext{
				Files: []string{"go.mod", "main.go"},
			},
			want: true,
		},
		{
			name: "project type match",
			manifest: &Manifest{
				Name:        "test-plugin",
				Version:     "1.0.0",
				Description: "Test",
				Activation: &ActivationConfig{
					Projects: []string{"react", "vue"},
				},
			},
			context: &ActivationContext{
				ProjectTypes: []string{"react"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.manifest.ShouldActivate(tt.context)
			if got != tt.want {
				t.Errorf("ShouldActivate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManifestPaths(t *testing.T) {
	manifest := &Manifest{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "Test",
		Skills:      []string{"skills/test.md", "skills/another.md"},
		Commands:    []string{"commands/test.md"},
		Agents:      []string{"agents/test.md"},
		Templates:   []string{"templates/test.txt"},
	}

	pluginDir := "/path/to/plugin"

	t.Run("GetSkillPaths", func(t *testing.T) {
		paths := manifest.GetSkillPaths(pluginDir)
		if len(paths) != 2 {
			t.Errorf("GetSkillPaths() returned %d paths, want 2", len(paths))
		}
		want := filepath.Join(pluginDir, "skills/test.md")
		if paths[0] != want {
			t.Errorf("GetSkillPaths()[0] = %q, want %q", paths[0], want)
		}
	})

	t.Run("GetCommandPaths", func(t *testing.T) {
		paths := manifest.GetCommandPaths(pluginDir)
		if len(paths) != 1 {
			t.Errorf("GetCommandPaths() returned %d paths, want 1", len(paths))
		}
	})

	t.Run("GetAgentPaths", func(t *testing.T) {
		paths := manifest.GetAgentPaths(pluginDir)
		if len(paths) != 1 {
			t.Errorf("GetAgentPaths() returned %d paths, want 1", len(paths))
		}
	})

	t.Run("GetTemplatePaths", func(t *testing.T) {
		paths := manifest.GetTemplatePaths(pluginDir)
		if len(paths) != 1 {
			t.Errorf("GetTemplatePaths() returned %d paths, want 1", len(paths))
		}
	})
}

func TestFullID(t *testing.T) {
	manifest := &Manifest{
		Name:        "test-plugin",
		Version:     "1.2.3",
		Description: "Test",
	}

	got := manifest.FullID()
	want := "test-plugin@1.2.3"
	if got != want {
		t.Errorf("FullID() = %q, want %q", got, want)
	}
}

func TestContainsHelper(t *testing.T) {
	slice := []string{"foo", "bar", "baz"}

	tests := []struct {
		item string
		want bool
	}{
		{"foo", true},
		{"bar", true},
		{"Foo", true}, // case-insensitive
		{"BAR", true}, // case-insensitive
		{"qux", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.item, func(t *testing.T) {
			got := contains(slice, tt.item)
			if got != tt.want {
				t.Errorf("contains(%v, %q) = %v, want %v", slice, tt.item, got, tt.want)
			}
		})
	}
}

func TestSaveManifest(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.json")

	manifest := &Manifest{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "Test",
		Author:      "Test Author",
	}

	err := manifest.Save(path)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	loaded, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}

	if loaded.Name != manifest.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, manifest.Name)
	}
	if loaded.Author != manifest.Author {
		t.Errorf("Author = %q, want %q", loaded.Author, manifest.Author)
	}
}

func TestIsValidPluginName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"valid-plugin", true},
		{"plugin123", true},
		{"p", true},
		{"test-plugin-123", true},
		{"Invalid", false},     // uppercase
		{"test_plugin", false}, // underscore
		{"test plugin", false}, // space
		{"123plugin", false},   // starts with number
		{"-plugin", false},     // starts with hyphen
		{"test.plugin", false}, // dot
		{"test@plugin", false}, // special char
		{"", false},            // empty
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPluginName(tt.name)
			if got != tt.valid {
				t.Errorf("isValidPluginName(%q) = %v, want %v", tt.name, got, tt.valid)
			}
		})
	}
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"v1.0.0", true},
		{"0.0.1", true},
		{"10.20.30", true},
		{"1.0.0-beta", true},
		{"1.0.0-beta.1", true},
		{"1.0.0-alpha.beta.1", true},
		{"1.0", false},
		{"1", false},
		{"v1", false},
		{"1.0.0.0", false},
		{"a.b.c", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := isValidSemver(tt.version)
			if got != tt.valid {
				t.Errorf("isValidSemver(%q) = %v, want %v", tt.version, got, tt.valid)
			}
		})
	}
}
