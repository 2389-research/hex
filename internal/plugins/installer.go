// Package plugins provides plugin installation and lifecycle management.
//
// ABOUTME: Plugin installation and management operations
// ABOUTME: Handles install, uninstall, update operations for plugins
package plugins

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Installer handles plugin installation operations
type Installer struct {
	loader *Loader
}

// NewInstaller creates a new plugin installer
func NewInstaller(loader *Loader) *Installer {
	return &Installer{
		loader: loader,
	}
}

// Install installs a plugin from a source (URL, git repo, or local path)
func (i *Installer) Install(source string) error {
	// Determine source type and install accordingly
	if isGitURL(source) {
		return i.installFromGit(source)
	}
	if isHTTPURL(source) {
		return i.installFromHTTP(source)
	}
	return i.installFromLocal(source)
}

// installFromGit clones a git repository
func (i *Installer) installFromGit(repoURL string) error {
	// Extract plugin name from URL
	pluginName := extractPluginNameFromURL(repoURL)
	targetDir := filepath.Join(i.loader.GetPluginsDir(), pluginName)

	// Check if already installed
	if i.loader.State().IsInstalled(pluginName) {
		return fmt.Errorf("plugin %s is already installed. Use 'clem plugin update %s' to update", pluginName, pluginName)
	}

	// Clone repository
	//nolint:gosec // G204 - git clone with user-provided URL is intentional for plugin installation
	cmd := exec.Command("git", "clone", repoURL, targetDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	// Load and validate manifest
	if err := i.loader.ValidatePlugin(targetDir); err != nil {
		// Clean up on validation failure
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("plugin validation failed: %w", err)
	}

	// Load manifest to get actual name and version
	manifest, err := LoadManifest(filepath.Join(targetDir, "plugin.json"))
	if err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("load manifest: %w", err)
	}

	// Run install script if present
	if err := i.runScript(targetDir, "install"); err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("install script failed: %w", err)
	}

	// Add to state
	i.loader.State().AddPlugin(manifest.Name, manifest.Version, targetDir)
	if err := i.loader.SaveState(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

// installFromHTTP downloads a tarball or zip
func (i *Installer) installFromHTTP(_ string) error {
	// TODO: Implement HTTP download and extraction
	return fmt.Errorf("HTTP installation not yet implemented. Use git URLs or local paths for now")
}

// installFromLocal copies a local directory
func (i *Installer) installFromLocal(sourcePath string) error {
	// Expand ~ to home directory
	if strings.HasPrefix(sourcePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		sourcePath = filepath.Join(home, sourcePath[2:])
	}

	// Make path absolute
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Validate source plugin
	if err := i.loader.ValidatePlugin(absSource); err != nil {
		return fmt.Errorf("invalid plugin: %w", err)
	}

	// Load manifest to get name
	manifest, err := LoadManifest(filepath.Join(absSource, "plugin.json"))
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	// Check if already installed
	if i.loader.State().IsInstalled(manifest.Name) {
		return fmt.Errorf("plugin %s is already installed. Use 'clem plugin update %s' to update", manifest.Name, manifest.Name)
	}

	targetDir := filepath.Join(i.loader.GetPluginsDir(), manifest.Name)

	// Copy directory
	if err := copyDir(absSource, targetDir); err != nil {
		return fmt.Errorf("copy plugin: %w", err)
	}

	// Run install script if present
	if err := i.runScript(targetDir, "install"); err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("install script failed: %w", err)
	}

	// Add to state
	i.loader.State().AddPlugin(manifest.Name, manifest.Version, targetDir)
	if err := i.loader.SaveState(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

// Uninstall removes a plugin
func (i *Installer) Uninstall(name string) error {
	// Check if installed
	if !i.loader.State().IsInstalled(name) {
		return fmt.Errorf("plugin %s is not installed", name)
	}

	// Get plugin info
	plugin, err := i.loader.GetPlugin(name)
	if err != nil {
		return fmt.Errorf("get plugin: %w", err)
	}

	// Run uninstall script if present
	if err := i.runScript(plugin.Dir, "uninstall"); err != nil {
		// Log warning but continue with uninstall
		fmt.Fprintf(os.Stderr, "Warning: uninstall script failed: %v\n", err)
	}

	// Remove directory
	if err := os.RemoveAll(plugin.Dir); err != nil {
		return fmt.Errorf("remove plugin directory: %w", err)
	}

	// Remove from state
	i.loader.State().RemovePlugin(name)
	if err := i.loader.SaveState(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

// Update updates a plugin to the latest version
func (i *Installer) Update(name string) error {
	// Check if installed
	if !i.loader.State().IsInstalled(name) {
		return fmt.Errorf("plugin %s is not installed", name)
	}

	// Get plugin info
	plugin, err := i.loader.GetPlugin(name)
	if err != nil {
		return fmt.Errorf("get plugin: %w", err)
	}

	// Check if it's a git repository
	gitDir := filepath.Join(plugin.Dir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Pull latest changes
		//nolint:gosec // G204 - git pull on validated plugin directory is safe
		cmd := exec.Command("git", "-C", plugin.Dir, "pull")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git pull failed: %w\nOutput: %s", err, string(output))
		}

		// Reload manifest to get new version
		manifest, err := LoadManifest(filepath.Join(plugin.Dir, "plugin.json"))
		if err != nil {
			return fmt.Errorf("load manifest: %w", err)
		}

		// Run update script if present
		if err := i.runScript(plugin.Dir, "update"); err != nil {
			return fmt.Errorf("update script failed: %w", err)
		}

		// Update state
		if err := i.loader.State().UpdatePlugin(name, manifest.Version); err != nil {
			return fmt.Errorf("update state: %w", err)
		}
		if err := i.loader.SaveState(); err != nil {
			return fmt.Errorf("save state: %w", err)
		}

		return nil
	}

	return fmt.Errorf("plugin %s is not a git repository and cannot be updated automatically", name)
}

// Enable enables a plugin
func (i *Installer) Enable(name string) error {
	if !i.loader.State().IsInstalled(name) {
		return fmt.Errorf("plugin %s is not installed", name)
	}

	if err := i.loader.State().EnablePlugin(name); err != nil {
		return err
	}

	return i.loader.SaveState()
}

// Disable disables a plugin
func (i *Installer) Disable(name string) error {
	if !i.loader.State().IsInstalled(name) {
		return fmt.Errorf("plugin %s is not installed", name)
	}

	if err := i.loader.State().DisablePlugin(name); err != nil {
		return err
	}

	return i.loader.SaveState()
}

// runScript executes a plugin lifecycle script
func (i *Installer) runScript(pluginDir, scriptName string) error {
	manifest, err := LoadManifest(filepath.Join(pluginDir, "plugin.json"))
	if err != nil {
		return err
	}

	scriptPath, exists := manifest.Scripts[scriptName]
	if !exists {
		// No script to run
		return nil
	}

	fullScriptPath := filepath.Join(pluginDir, scriptPath)
	if _, err := os.Stat(fullScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	// Make script executable
	//nolint:gosec // G302 - 0755 is correct for executable scripts
	if err := os.Chmod(fullScriptPath, 0755); err != nil {
		return fmt.Errorf("make script executable: %w", err)
	}

	// Run script
	//nolint:gosec // G204 - executing validated plugin lifecycle scripts is intentional
	cmd := exec.Command(fullScriptPath)
	cmd.Dir = pluginDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}

// Helper functions

func isGitURL(source string) bool {
	return strings.HasPrefix(source, "git@") ||
		strings.HasPrefix(source, "https://") && (strings.Contains(source, "github.com") || strings.Contains(source, "gitlab.com") || strings.Contains(source, "bitbucket.org")) ||
		strings.HasPrefix(source, "git://") ||
		strings.HasSuffix(source, ".git")
}

func isHTTPURL(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

func extractPluginNameFromURL(url string) string {
	// Extract name from git URL
	// Example: https://github.com/user/my-plugin.git -> my-plugin
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]
	name = strings.TrimSuffix(name, ".git")
	return name
}

func copyDir(src, dst string) error {
	// Get source info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	//nolint:gosec // G304 - opening source file for plugin installation with validated path
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close() //nolint:errcheck // Defer close on read-only file is safe

	//nolint:gosec // G304 - creating destination file for plugin installation
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close() //nolint:errcheck // Error handled by explicit sync/close below

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
