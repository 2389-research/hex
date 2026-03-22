// ABOUTME: Project memory scanner for cross-session context awareness
// ABOUTME: Detects project type, build/test commands, and structure; persists to .hex/project.json
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProjectInfo holds detected project metadata
type ProjectInfo struct {
	Language     string   `json:"language"`
	BuildCommand string   `json:"build_command,omitempty"`
	TestCommand  string   `json:"test_command,omitempty"`
	Structure    []string `json:"structure,omitempty"`
	DetectedAt   string   `json:"detected_at"`
}

// DetectProject scans a directory and detects project characteristics
func DetectProject(dir string) (*ProjectInfo, error) {
	proj := &ProjectInfo{
		DetectedAt: time.Now().Format(time.RFC3339),
	}
	proj.Language = detectLanguage(dir)
	proj.BuildCommand = detectBuildCommand(dir, proj.Language)
	proj.TestCommand = detectTestCommand(proj.Language)
	proj.Structure = detectStructure(dir)
	return proj, nil
}

func detectLanguage(dir string) string {
	markers := map[string]string{
		"go.mod":           "go",
		"Cargo.toml":       "rust",
		"package.json":     "javascript",
		"pyproject.toml":   "python",
		"setup.py":         "python",
		"requirements.txt": "python",
		"Gemfile":          "ruby",
		"build.gradle":     "java",
		"pom.xml":          "java",
		"mix.exs":          "elixir",
		"pubspec.yaml":     "dart",
		"Package.swift":    "swift",
	}
	for file, lang := range markers {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return lang
		}
	}
	return "unknown"
}

func detectBuildCommand(dir, language string) string {
	if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
		return "make build"
	}
	switch language {
	case "go":
		return "go build ./..."
	case "rust":
		return "cargo build"
	case "javascript":
		return "npm run build"
	case "java":
		if _, err := os.Stat(filepath.Join(dir, "build.gradle")); err == nil {
			return "gradle build"
		}
		return "mvn compile"
	default:
		return ""
	}
}

func detectTestCommand(language string) string {
	switch language {
	case "go":
		return "go test ./..."
	case "rust":
		return "cargo test"
	case "javascript":
		return "npm test"
	case "python":
		return "pytest"
	case "ruby":
		return "bundle exec rspec"
	case "java":
		return "mvn test"
	case "elixir":
		return "mix test"
	default:
		return ""
	}
}

func detectStructure(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && entry.Name() != "node_modules" && entry.Name() != "vendor" {
			dirs = append(dirs, entry.Name()+"/")
		}
	}
	return dirs
}

// Save persists project info to hexDir/project.json
func Save(hexDir string, proj *ProjectInfo) error {
	if err := os.MkdirAll(hexDir, 0755); err != nil {
		return fmt.Errorf("create .hex directory: %w", err)
	}
	data, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal project info: %w", err)
	}
	return os.WriteFile(filepath.Join(hexDir, "project.json"), data, 0644)
}

// Load reads project info from hexDir/project.json
func Load(hexDir string) (*ProjectInfo, error) {
	data, err := os.ReadFile(filepath.Join(hexDir, "project.json"))
	if err != nil {
		return nil, err
	}
	var proj ProjectInfo
	if err := json.Unmarshal(data, &proj); err != nil {
		return nil, fmt.Errorf("unmarshal project info: %w", err)
	}
	return &proj, nil
}

// IsStale returns true if the project info is older than maxAge
func IsStale(proj *ProjectInfo, maxAge time.Duration) bool {
	if proj == nil {
		return true
	}
	detected, err := time.Parse(time.RFC3339, proj.DetectedAt)
	if err != nil {
		return true
	}
	return time.Since(detected) > maxAge
}

// ToPromptContext generates a brief context string for the system prompt
func (p *ProjectInfo) ToPromptContext() string {
	var parts []string
	if p.Language != "" && p.Language != "unknown" {
		parts = append(parts, fmt.Sprintf("Language: %s", p.Language))
	}
	if p.BuildCommand != "" {
		parts = append(parts, fmt.Sprintf("Build: %s", p.BuildCommand))
	}
	if p.TestCommand != "" {
		parts = append(parts, fmt.Sprintf("Test: %s", p.TestCommand))
	}
	if len(p.Structure) > 0 {
		parts = append(parts, fmt.Sprintf("Directories: %s", strings.Join(p.Structure, ", ")))
	}
	if len(parts) == 0 {
		return ""
	}
	return "Project context: " + strings.Join(parts, ". ")
}
