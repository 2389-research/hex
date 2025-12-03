// Package forms provides beautiful huh-based forms for the hex TUI.
// ABOUTME: Huh-based onboarding flow for first-run experience
// ABOUTME: Guides new users through setup with welcome, API key, model selection, and tutorial
package forms

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/harper/hex/internal/ui/theme"
)

// OnboardingFormResult contains the result of the onboarding flow
type OnboardingFormResult struct {
	Model        string
	APIKey       string
	ShowTutorial bool
	StartSample  bool
	Completed    bool
	Skipped      bool
}

// OnboardingForm creates a first-run onboarding experience using huh
type OnboardingForm struct {
	result *OnboardingFormResult
	theme  *theme.Theme

	// Available models for selection
	availableModels []string

	// Config file path for saving
	configPath string

	// Internal string field for confirmation
	wantsToSetup bool
}

// NewOnboardingForm creates a new onboarding wizard
func NewOnboardingForm(configPath string) *OnboardingForm {
	// Default available models
	availableModels := []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
	}

	result := &OnboardingFormResult{
		Model:        "claude-3-5-sonnet-20241022", // Default to latest Sonnet
		ShowTutorial: true,
		StartSample:  false,
		Completed:    false,
		Skipped:      false,
	}

	return &OnboardingForm{
		result:          result,
		theme:           theme.DraculaTheme(),
		availableModels: availableModels,
		configPath:      configPath,
		wantsToSetup:    true,
	}
}

// Run displays the onboarding flow and returns the result
func (f *OnboardingForm) Run() (*OnboardingFormResult, error) {
	// Create multi-step onboarding form
	form := huh.NewForm(
		// Step 1: Welcome Screen
		huh.NewGroup(
			huh.NewNote().
				Title("👋 Welcome to Hex!").
				Description(f.buildWelcomeText()),
			huh.NewConfirm().
				Title("Ready to get started?").
				Description("Let's set up your Hex configuration").
				Affirmative("Yes, let's go!").
				Negative("Skip for now").
				Value(&f.wantsToSetup),
		),

		// Step 2: API Key Setup (conditional on wants to setup)
		huh.NewGroup(
			huh.NewInput().
				Title("🔑 Enter your Anthropic API Key").
				Description("Get your key from: https://console.anthropic.com/").
				Value(&f.result.APIKey).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("API key is required")
					}
					if !strings.HasPrefix(s, "sk-ant-") {
						return fmt.Errorf("API key should start with 'sk-ant-'")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return !f.wantsToSetup
		}),

		// Step 3: Model Selection (conditional on wants to setup)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("🤖 Choose Your AI Model").
				Description("Select which Claude model to use").
				Options(f.buildModelOptions()...).
				Value(&f.result.Model),
		).WithHideFunc(func() bool {
			return !f.wantsToSetup
		}),

		// Step 4: Tutorial Offer
		huh.NewGroup(
			huh.NewConfirm().
				Title("📚 Show Quick Tutorial?").
				Description("Learn the basics: navigation, commands, and features").
				Affirmative("Yes, show me").
				Negative("No thanks").
				Value(&f.result.ShowTutorial),
		),

		// Step 5: Sample Conversation Offer
		huh.NewGroup(
			huh.NewConfirm().
				Title("💬 Try a Sample Conversation?").
				Description("Start with a pre-made example to see Hex in action").
				Affirmative("Yes, let's try it").
				Negative("No, I'll start fresh").
				Value(&f.result.StartSample),
		),
	).WithTheme(f.getDraculaTheme())

	// Run the form
	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Update result based on choices
	f.result.Skipped = !f.wantsToSetup
	f.result.Completed = f.wantsToSetup

	return f.result, nil
}

// buildWelcomeText creates the welcome message
func (f *OnboardingForm) buildWelcomeText() string {
	return `Hex is your intelligent command-line assistant powered by Claude.

Features:
  • Interactive chat with Claude AI
  • Tool execution and approval workflow
  • Conversation history and management
  • Vim-style navigation (j/k, gg/G)
  • Quick actions palette (:)
  • Smart suggestions and autocomplete

Let's get you set up in just a few steps!`
}

// buildModelOptions creates huh options for available models
func (f *OnboardingForm) buildModelOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(f.availableModels))

	for _, model := range f.availableModels {
		label := f.formatModelLabel(model)
		options = append(options, huh.NewOption(label, model))
	}

	return options
}

// formatModelLabel creates a friendly label for a model ID
func (f *OnboardingForm) formatModelLabel(modelID string) string {
	switch {
	case strings.Contains(modelID, "sonnet") && strings.Contains(modelID, "3-5"):
		return f.theme.Emphasized.Render("Claude 3.5 Sonnet") +
			f.theme.Muted.Render(" - Recommended: Best balance")
	case strings.Contains(modelID, "haiku") && strings.Contains(modelID, "3-5"):
		return f.theme.Emphasized.Render("Claude 3.5 Haiku") +
			f.theme.Muted.Render(" - Fast and efficient")
	case strings.Contains(modelID, "opus"):
		return f.theme.Emphasized.Render("Claude 3 Opus") +
			f.theme.Muted.Render(" - Most powerful")
	case strings.Contains(modelID, "sonnet"):
		return f.theme.Emphasized.Render("Claude 3 Sonnet") +
			f.theme.Muted.Render(" - Balanced")
	case strings.Contains(modelID, "haiku"):
		return f.theme.Emphasized.Render("Claude 3 Haiku") +
			f.theme.Muted.Render(" - Fast")
	default:
		return modelID
	}
}

// getDraculaTheme returns a huh theme configured with Dracula colors
func (f *OnboardingForm) getDraculaTheme() *huh.Theme {
	t := huh.ThemeBase()

	colors := f.theme.Colors

	// Configure with Dracula colors
	t.Focused.Base = t.Focused.Base.
		BorderForeground(colors.Purple)

	t.Focused.Title = t.Focused.Title.
		Foreground(colors.Purple).
		Bold(true)

	t.Focused.Description = t.Focused.Description.
		Foreground(colors.Foreground)

	t.Focused.SelectSelector = t.Focused.SelectSelector.
		Foreground(colors.Pink)

	t.Focused.SelectedOption = t.Focused.SelectedOption.
		Foreground(colors.Cyan).
		Bold(true)

	t.Focused.UnselectedOption = t.Focused.UnselectedOption.
		Foreground(colors.Comment)

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.
		Foreground(colors.Pink)

	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.
		Foreground(colors.Comment)

	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.
		Foreground(colors.Purple)

	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(colors.Background).
		Background(colors.Green).
		Bold(true)

	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(colors.Foreground).
		Background(colors.CurrentLine)

	return t
}

// GetTutorialText returns the tutorial content to display
func (f *OnboardingForm) GetTutorialText() string {
	return `# Hex Quick Tutorial

## Navigation
- j/k - Scroll down/up (Vim-style)
- Ctrl+D/U - Page down/up
- gg/G - Go to top/bottom

## Commands
- : - Open quick actions menu
- ? - Show help
- / - Search in conversation
- Tab - Cycle views or autocomplete

## Features
- Enter - Send message
- Ctrl+L - Clear screen
- Ctrl+K - Clear conversation
- Ctrl+S - Save conversation
- Ctrl+E - Export conversation
- Ctrl+F - Toggle favorite
- Esc - Exit modes or quit

## Tool Approval
When Claude needs to use a tool, you'll see an approval prompt with:
- ✓ Approve - Run this time
- ✗ Deny - Skip this time
- ✓✓ Always Allow - Never ask again
- ✗✗ Never Allow - Block permanently

Press '?' anytime for detailed help!`
}

// GetSampleConversation returns a sample conversation to start with
func (f *OnboardingForm) GetSampleConversation() string {
	return `Hello! I'm ready to help. Try asking me:
- "What files are in the current directory?"
- "Search for TODO comments in my code"
- "Explain how async/await works in JavaScript"
- "Help me write a README for my project"

What would you like to do?`
}
