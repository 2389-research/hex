// ABOUTME: First-run setup wizard for Hex CLI
// ABOUTME: Guides users through provider selection and API key configuration

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrSetupCancelled is returned when the user cancels the setup wizard
var ErrSetupCancelled = fmt.Errorf("setup cancelled")

// Wizard states
type wizardState int

const (
	stateWelcome wizardState = iota
	stateProviderChoice
	stateAPIKey
	stateComplete
	stateError
)

// Dracula color palette
var (
	wzColorPurple   = lipgloss.Color("#bd93f9")
	wzColorCyan     = lipgloss.Color("#8be9fd")
	wzColorGreen    = lipgloss.Color("#50fa7b")
	wzColorPink     = lipgloss.Color("#ff79c6")
	wzColorYellow   = lipgloss.Color("#f1fa8c")
	wzColorOrange   = lipgloss.Color("#ffb86c")
	wzColorRed      = lipgloss.Color("#ff5555")
	wzColorFg       = lipgloss.Color("#f8f8f2")
	wzColorMuted    = lipgloss.Color("#6272a4")
	wzColorDarkMute = lipgloss.Color("#44475a")

	wizardBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(wzColorPurple).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	wizardErrorBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(wzColorRed).
				Padding(1, 2).
				MarginTop(1)

	wizardOptionStyle = lipgloss.NewStyle().
				Foreground(wzColorFg)

	wizardOptionSelectedStyle = lipgloss.NewStyle().
					Foreground(wzColorGreen).
					Bold(true)

	wizardMutedStyle = lipgloss.NewStyle().
				Foreground(wzColorMuted)

	wizardHighlightStyle = lipgloss.NewStyle().
				Foreground(wzColorCyan)

	wizardSuccessStyle = lipgloss.NewStyle().
				Foreground(wzColorGreen).
				Bold(true)

	wizardErrorStyle = lipgloss.NewStyle().
				Foreground(wzColorRed).
				Bold(true)

	brandLogoStyle = lipgloss.NewStyle().
			Foreground(wzColorPurple).
			Bold(true)

	brandTaglineStyle = lipgloss.NewStyle().
				Foreground(wzColorCyan).
				Italic(true)

	brandOrgStyle = lipgloss.NewStyle().
			Foreground(wzColorMuted)

	brandVersionStyle = lipgloss.NewStyle().
				Foreground(wzColorDarkMute)

	featureIconStyle = lipgloss.NewStyle().
				Foreground(wzColorPink)

	featureTitleStyle = lipgloss.NewStyle().
				Foreground(wzColorYellow).
				Bold(true)

	featureDescStyle = lipgloss.NewStyle().
				Foreground(wzColorFg)

	tipKeyStyle = lipgloss.NewStyle().
			Foreground(wzColorOrange).
			Bold(true)
)

// providerOption describes a supported LLM provider
type providerOption struct {
	name    string // Display name
	key     string // Internal key
	desc    string // Description
	envVar  string // Environment variable name
	keyHint string // Placeholder for API key input
}

var providerOptions = []providerOption{
	{"Anthropic (Claude)", "anthropic", "Claude Sonnet 4.5, Opus 4.5, Haiku", "ANTHROPIC_API_KEY", "sk-ant-api..."},
	{"OpenAI (GPT)", "openai", "GPT-4o, GPT-4 Turbo", "OPENAI_API_KEY", "sk-..."},
	{"Google (Gemini)", "gemini", "Gemini 2.5 Pro, Flash", "GEMINI_API_KEY", "AI..."},
	{"OpenRouter", "openrouter", "Access multiple providers", "OPENROUTER_API_KEY", "sk-or-..."},
	{"Ollama (Local)", "ollama", "Run models locally - no API key needed", "", ""},
}

// WizardModel is the Bubbletea model for the setup wizard
type WizardModel struct {
	state     wizardState
	cursor    int
	textInput textinput.Model
	err       error
	errMsg    string
	width     int
	height    int
	quitting  bool

	// Results
	provider string
	apiKey   string
}

// NewWizardModel creates a new wizard model
func NewWizardModel() *WizardModel {
	ti := textinput.New()
	ti.Placeholder = "sk-ant-api..."
	ti.CharLimit = 200
	ti.Width = 60
	ti.EchoMode = textinput.EchoPassword

	return &WizardModel{
		state:     stateWelcome,
		textInput: ti,
	}
}

func (m *WizardModel) Init() tea.Cmd {
	return nil
}

func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			// Go back if possible
			switch m.state {
			case stateProviderChoice:
				m.state = stateWelcome
			case stateAPIKey:
				m.textInput.Blur()
				m.state = stateProviderChoice
			case stateError:
				m.state = stateProviderChoice
				m.err = nil
				m.errMsg = ""
			}
			return m, nil
		}

		// Handle state-specific keys
		switch m.state {
		case stateWelcome:
			if msg.String() == "enter" {
				m.state = stateProviderChoice
				m.cursor = 0
				return m, nil
			}

		case stateProviderChoice:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(providerOptions)-1 {
					m.cursor++
				}
			case "enter":
				selected := providerOptions[m.cursor]
				m.provider = selected.key

				// Ollama doesn't need an API key
				if m.provider == "ollama" {
					m.state = stateComplete
					return m, nil
				}

				// Update text input placeholder based on provider
				m.textInput.Placeholder = selected.keyHint
				m.textInput.Reset()
				m.cursor = 0
				m.state = stateAPIKey
				m.textInput.Focus()
				return m, textinput.Blink
			}

		case stateAPIKey:
			switch msg.String() {
			case "enter":
				key := strings.TrimSpace(m.textInput.Value())
				if key != "" {
					// Validate API key format
					if err := validateAPIKey(m.provider, key); err != nil {
						m.err = err
						m.errMsg = err.Error()
						m.state = stateError
						return m, nil
					}
					m.apiKey = key
					m.state = stateComplete
				}
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}

		case stateComplete:
			if msg.String() == "enter" {
				// Save config and exit
				if err := m.saveConfig(); err != nil {
					m.err = err
					m.errMsg = err.Error()
					m.state = stateError
					return m, nil
				}
				return m, tea.Quit
			}

		case stateError:
			if msg.String() == "enter" {
				m.state = stateProviderChoice
				m.err = nil
				m.errMsg = ""
			}
		}
	}

	return m, nil
}

func (m *WizardModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Only show full branding on welcome screen
	if m.state == stateWelcome {
		return m.viewWelcome()
	}

	// Compact header for other screens
	header := brandLogoStyle.Render("HEX") + " " + brandOrgStyle.Render("by 2389 Research")
	b.WriteString(header)
	b.WriteString("\n")

	switch m.state {
	case stateProviderChoice:
		b.WriteString(m.viewProviderChoice())
	case stateAPIKey:
		b.WriteString(m.viewAPIKey())
	case stateComplete:
		b.WriteString(m.viewComplete())
	case stateError:
		b.WriteString(m.viewError())
	}

	return b.String()
}

func (m *WizardModel) viewWelcome() string {
	var b strings.Builder

	// ASCII art logo
	logo := `
  ██╗  ██╗███████╗██╗  ██╗
  ██║  ██║██╔════╝╚██╗██╔╝
  ███████║█████╗   ╚███╔╝
  ██╔══██║██╔══╝   ██╔██╗
  ██║  ██║███████╗██╔╝ ██╗
  ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`
	b.WriteString(brandLogoStyle.Render(logo))
	b.WriteString("\n\n")

	// Tagline
	b.WriteString(brandTaglineStyle.Render("  Claude-powered CLI for developers"))
	b.WriteString("\n\n")

	// Organization and version
	b.WriteString(brandOrgStyle.Render("  by 2389 Research"))
	b.WriteString("  ")
	b.WriteString(brandVersionStyle.Render(fmt.Sprintf("v%s", version)))
	b.WriteString("\n")

	// Separator
	b.WriteString(wizardMutedStyle.Render("  " + strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	// Feature highlights
	features := []struct {
		icon  string
		title string
		desc  string
	}{
		{"🔧", "Code Tools", "Read, write, edit files; run shell commands"},
		{"🔍", "Search", "Grep, glob, and explore your codebase"},
		{"🌐", "Web", "Fetch URLs and search the web"},
		{"🤖", "Multi-Provider", "Anthropic, OpenAI, Gemini, OpenRouter, Ollama"},
	}

	for _, f := range features {
		b.WriteString("  ")
		b.WriteString(featureIconStyle.Render(f.icon))
		b.WriteString(" ")
		b.WriteString(featureTitleStyle.Render(f.title))
		b.WriteString(wizardMutedStyle.Render(" - "))
		b.WriteString(featureDescStyle.Render(f.desc))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Quick start tips
	b.WriteString(wizardMutedStyle.Render("  Quick Start:"))
	b.WriteString("\n")
	tips := []struct {
		key  string
		desc string
	}{
		{"hex", "Launch interactive chat"},
		{"hex -p \"question\"", "One-shot query (print mode)"},
		{"hex --continue", "Resume last conversation"},
		{"hex doctor", "Check your setup"},
	}
	for _, tip := range tips {
		b.WriteString("    ")
		b.WriteString(tipKeyStyle.Render(tip.key))
		b.WriteString(wizardMutedStyle.Render(" → "))
		b.WriteString(featureDescStyle.Render(tip.desc))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Separator
	b.WriteString(wizardMutedStyle.Render("  " + strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	// Setup prompt
	b.WriteString(featureTitleStyle.Render("  Ready to get started?"))
	b.WriteString("\n")
	b.WriteString(featureDescStyle.Render("  Let's configure your LLM provider."))
	b.WriteString("\n\n")

	// Help
	b.WriteString(wizardMutedStyle.Render("  Press "))
	b.WriteString(tipKeyStyle.Render("[Enter]"))
	b.WriteString(wizardMutedStyle.Render(" to continue, "))
	b.WriteString(tipKeyStyle.Render("[q]"))
	b.WriteString(wizardMutedStyle.Render(" to quit"))

	return b.String()
}

func (m *WizardModel) viewProviderChoice() string {
	var content strings.Builder

	content.WriteString("Choose your AI provider:\n\n")

	for i, opt := range providerOptions {
		cursor := "  "
		style := wizardOptionStyle
		if i == m.cursor {
			cursor = "▸ "
			style = wizardOptionSelectedStyle
		}

		content.WriteString(cursor)
		content.WriteString(style.Render(opt.name))
		content.WriteString("\n")
		content.WriteString("    ")
		content.WriteString(wizardMutedStyle.Render(opt.desc))
		if opt.envVar != "" {
			content.WriteString(wizardMutedStyle.Render(" (" + opt.envVar + ")"))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(wizardMutedStyle.Render("Use "))
	content.WriteString(tipKeyStyle.Render("↑/↓"))
	content.WriteString(wizardMutedStyle.Render(" or "))
	content.WriteString(tipKeyStyle.Render("j/k"))
	content.WriteString(wizardMutedStyle.Render(" to navigate, "))
	content.WriteString(tipKeyStyle.Render("Enter"))
	content.WriteString(wizardMutedStyle.Render(" to select"))

	return wizardBoxStyle.Render(content.String())
}

func (m *WizardModel) viewAPIKey() string {
	var content strings.Builder

	selected := providerOptions[m.findProviderIndex()]

	content.WriteString(fmt.Sprintf("Enter your %s API key:\n\n", selected.name))
	content.WriteString(m.textInput.View())
	content.WriteString("\n\n")

	if selected.envVar != "" {
		content.WriteString(wizardMutedStyle.Render("You can also set "))
		content.WriteString(wizardHighlightStyle.Render(selected.envVar))
		content.WriteString(wizardMutedStyle.Render(" environment variable"))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(wizardMutedStyle.Render("Press "))
	content.WriteString(tipKeyStyle.Render("Enter"))
	content.WriteString(wizardMutedStyle.Render(" to save, "))
	content.WriteString(tipKeyStyle.Render("Esc"))
	content.WriteString(wizardMutedStyle.Render(" to go back"))

	return wizardBoxStyle.Render(content.String())
}

func (m *WizardModel) viewComplete() string {
	var content strings.Builder

	content.WriteString(wizardSuccessStyle.Render("✓ Setup Complete!"))
	content.WriteString("\n\n")

	content.WriteString("Provider: ")
	content.WriteString(wizardHighlightStyle.Render(m.provider))
	content.WriteString("\n")

	if m.apiKey != "" {
		content.WriteString("API Key: ")
		content.WriteString(wizardMutedStyle.Render("••••••••" + m.apiKey[len(m.apiKey)-4:]))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(featureDescStyle.Render("Configuration will be saved to ~/.hex/config.toml"))
	content.WriteString("\n\n")

	content.WriteString(wizardMutedStyle.Render("Press "))
	content.WriteString(tipKeyStyle.Render("Enter"))
	content.WriteString(wizardMutedStyle.Render(" to launch hex"))

	return wizardBoxStyle.Render(content.String())
}

func (m *WizardModel) viewError() string {
	var content strings.Builder

	content.WriteString(wizardErrorStyle.Render("✗ Error"))
	content.WriteString("\n\n")
	content.WriteString(m.errMsg)
	content.WriteString("\n\n")
	content.WriteString(wizardMutedStyle.Render("Press "))
	content.WriteString(tipKeyStyle.Render("Enter"))
	content.WriteString(wizardMutedStyle.Render(" to try again"))

	return wizardErrorBoxStyle.Render(content.String())
}

func (m *WizardModel) findProviderIndex() int {
	for i, opt := range providerOptions {
		if opt.key == m.provider {
			return i
		}
	}
	return 0
}

func (m *WizardModel) saveConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	hexDir := filepath.Join(home, ".hex")
	if err := os.MkdirAll(hexDir, 0700); err != nil {
		return fmt.Errorf("create .hex dir: %w", err)
	}

	configPath := filepath.Join(hexDir, "config.toml")

	var content strings.Builder
	content.WriteString("# Hex Configuration\n")
	content.WriteString("# Generated by setup wizard\n\n")
	content.WriteString(fmt.Sprintf("provider = %q\n", m.provider))

	if m.apiKey != "" {
		content.WriteString(fmt.Sprintf("\n[providers.%s]\n", m.provider))
		content.WriteString(fmt.Sprintf("api_key = %q\n", m.apiKey))
	}

	return os.WriteFile(configPath, []byte(content.String()), 0600)
}

// validateAPIKey performs basic validation on API key format
func validateAPIKey(provider, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("API key is required")
	}

	// Provider-specific prefix validation (soft validation - just warnings)
	switch provider {
	case "anthropic":
		if !strings.HasPrefix(key, "sk-ant-") {
			return fmt.Errorf("anthropic API keys typically start with 'sk-ant-'")
		}
	case "openai":
		if !strings.HasPrefix(key, "sk-") {
			return fmt.Errorf("OpenAI API keys typically start with 'sk-'")
		}
	case "openrouter":
		if !strings.HasPrefix(key, "sk-or-") {
			return fmt.Errorf("OpenRouter API keys typically start with 'sk-or-'")
		}
	}

	return nil
}

// IsFirstRun checks if hex needs initial setup
func IsFirstRun() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return true
	}

	// Check for existing config file
	configPath := filepath.Join(home, ".hex", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return false
	}

	// Also check for old YAML config
	yamlPath := filepath.Join(home, ".hex", "config.yaml")
	if _, err := os.Stat(yamlPath); err == nil {
		return false
	}

	// Check for any provider API key in environment
	providerEnvVars := []string{
		"ANTHROPIC_API_KEY",
		"OPENAI_API_KEY",
		"GEMINI_API_KEY",
		"OPENROUTER_API_KEY",
		"HEX_API_KEY",
	}

	for _, envVar := range providerEnvVars {
		if os.Getenv(envVar) != "" {
			return false
		}
	}

	return true
}

// RunWizard launches the setup wizard
func RunWizard() error {
	model := NewWizardModel()

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("run wizard: %w", err)
	}

	m := finalModel.(*WizardModel)
	if m.quitting && m.state != stateComplete {
		return ErrSetupCancelled
	}

	return nil
}
