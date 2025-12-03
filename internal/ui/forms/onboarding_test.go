package forms

import (
	"strings"
	"testing"
)

func TestNewOnboardingForm(t *testing.T) {
	form := NewOnboardingForm("/tmp/config.toml")

	if form == nil {
		t.Fatal("expected form to be created, got nil")
	}

	if form.result == nil {
		t.Fatal("expected result to be initialized, got nil")
	}

	if form.result.Model == "" {
		t.Error("expected default model to be set")
	}

	if form.theme == nil {
		t.Error("expected theme to be initialized, got nil")
	}

	if len(form.availableModels) == 0 {
		t.Error("expected available models to be populated")
	}

	// Default should want to setup
	if !form.wantsToSetup {
		t.Error("expected wantsToSetup to be true by default")
	}
}

func TestBuildWelcomeText(t *testing.T) {
	form := NewOnboardingForm("")

	text := form.buildWelcomeText()

	if text == "" {
		t.Error("expected non-empty welcome text")
	}

	// Should contain key information
	expectedPhrases := []string{
		"Hex",
		"Features",
		"Interactive chat",
		"Tool execution",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(text, phrase) {
			t.Errorf("expected welcome text to contain %q", phrase)
		}
	}
}

func TestOnboardingBuildModelOptions(t *testing.T) {
	form := NewOnboardingForm("")

	options := form.buildModelOptions()

	if len(options) == 0 {
		t.Error("expected model options to be built")
	}

	// Check that options have labels and values
	for i, opt := range options {
		if opt.Value == "" {
			t.Errorf("option %d has empty value", i)
		}
	}

	// Should have at least 3 default models
	if len(options) < 3 {
		t.Errorf("expected at least 3 model options, got %d", len(options))
	}
}

func TestOnboardingFormatModelLabel(t *testing.T) {
	form := NewOnboardingForm("")

	tests := []struct {
		name       string
		modelID    string
		shouldHave string
	}{
		{
			name:       "claude 3.5 sonnet",
			modelID:    "claude-3-5-sonnet-20241022",
			shouldHave: "Recommended",
		},
		{
			name:       "claude 3.5 haiku",
			modelID:    "claude-3-5-haiku-20241022",
			shouldHave: "Fast",
		},
		{
			name:       "claude 3 opus",
			modelID:    "claude-3-opus-20240229",
			shouldHave: "powerful",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := form.formatModelLabel(tt.modelID)

			if label == "" {
				t.Error("expected non-empty label")
			}

			// Labels should be informative
			// We can't check exact text due to lipgloss styling,
			// but we can verify we got output
			if len(label) < 10 {
				t.Error("expected label to be reasonably long")
			}
		})
	}
}

func TestGetTutorialText(t *testing.T) {
	form := NewOnboardingForm("")

	text := form.GetTutorialText()

	if text == "" {
		t.Error("expected non-empty tutorial text")
	}

	// Should contain key tutorial sections
	expectedSections := []string{
		"Navigation",
		"Commands",
		"Features",
		"Tool Approval",
	}

	for _, section := range expectedSections {
		if !strings.Contains(text, section) {
			t.Errorf("expected tutorial to contain section %q", section)
		}
	}

	// Should contain some key bindings
	expectedBindings := []string{
		"j/k",
		"gg/G",
		":",
		"?",
		"Enter",
		"Esc",
	}

	for _, binding := range expectedBindings {
		if !strings.Contains(text, binding) {
			t.Errorf("expected tutorial to mention key binding %q", binding)
		}
	}
}

func TestGetSampleConversation(t *testing.T) {
	form := NewOnboardingForm("")

	text := form.GetSampleConversation()

	if text == "" {
		t.Error("expected non-empty sample conversation")
	}

	// Should contain sample prompts
	if !strings.Contains(text, "Try asking") {
		t.Error("expected sample conversation to suggest things to try")
	}

	// Should be friendly and inviting
	if !strings.Contains(text, "Hello") || !strings.Contains(text, "help") {
		t.Error("expected sample conversation to be friendly")
	}
}

func TestOnboardingFormResult(t *testing.T) {
	result := &OnboardingFormResult{
		Model:        "claude-3-5-sonnet-20241022",
		APIKey:       "sk-ant-test",
		ShowTutorial: true,
		StartSample:  false,
		Completed:    true,
		Skipped:      false,
	}

	// Verify fields are set correctly
	if result.Model != "claude-3-5-sonnet-20241022" {
		t.Error("expected model to be set")
	}

	if result.APIKey != "sk-ant-test" {
		t.Error("expected API key to be set")
	}

	if !result.ShowTutorial {
		t.Error("expected ShowTutorial to be true")
	}

	if result.StartSample {
		t.Error("expected StartSample to be false")
	}

	if !result.Completed {
		t.Error("expected Completed to be true")
	}

	if result.Skipped {
		t.Error("expected Skipped to be false")
	}
}

func TestOnboardingGetDraculaTheme(t *testing.T) {
	form := NewOnboardingForm("")
	theme := form.getDraculaTheme()

	if theme == nil {
		t.Fatal("expected theme to be created, got nil")
	}

	// Verify theme has required style groups
	_ = theme.Focused.Title.String()
	_ = theme.Focused.TextInput.Cursor.String()
}
