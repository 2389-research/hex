// ABOUTME: Tests for spell applicator functionality
// ABOUTME: Verifies system prompt layering, tools config, and response settings

package spells

import "testing"

func TestApplySpell_Replace(t *testing.T) {
	spell := &Spell{
		Name:         "test",
		Description:  "Test",
		SystemPrompt: "You are a test agent.",
		Mode:         LayerModeReplace,
	}

	basePrompt := "You are the default agent."
	result := ApplySpell(spell, basePrompt, nil)

	if result.SystemPrompt != spell.SystemPrompt {
		t.Errorf("SystemPrompt = %q; want %q", result.SystemPrompt, spell.SystemPrompt)
	}
}

func TestApplySpell_Layer(t *testing.T) {
	spell := &Spell{
		Name:         "test",
		Description:  "Test",
		SystemPrompt: "Additional instructions: be concise.",
		Mode:         LayerModeLayer,
	}

	basePrompt := "You are the default agent."
	result := ApplySpell(spell, basePrompt, nil)

	expected := basePrompt + "\n\n" + spell.SystemPrompt
	if result.SystemPrompt != expected {
		t.Errorf("SystemPrompt = %q; want %q", result.SystemPrompt, expected)
	}
}

func TestApplySpell_LayerEmptyBase(t *testing.T) {
	spell := &Spell{
		Name:         "test",
		Description:  "Test",
		SystemPrompt: "You are a test agent.",
		Mode:         LayerModeLayer,
	}

	basePrompt := ""
	result := ApplySpell(spell, basePrompt, nil)

	// With empty base, should just use spell prompt
	if result.SystemPrompt != spell.SystemPrompt {
		t.Errorf("SystemPrompt = %q; want %q", result.SystemPrompt, spell.SystemPrompt)
	}
}

func TestApplySpell_LayerEmptySpell(t *testing.T) {
	spell := &Spell{
		Name:         "test",
		Description:  "Test",
		SystemPrompt: "",
		Mode:         LayerModeLayer,
	}

	basePrompt := "You are the default agent."
	result := ApplySpell(spell, basePrompt, nil)

	// With empty spell prompt, should just use base prompt
	if result.SystemPrompt != basePrompt {
		t.Errorf("SystemPrompt = %q; want %q", result.SystemPrompt, basePrompt)
	}
}

func TestApplySpell_ModeOverride(t *testing.T) {
	spell := &Spell{
		Name:         "test",
		Description:  "Test",
		SystemPrompt: "You are a test agent.",
		Mode:         LayerModeLayer, // Default is layer
	}

	basePrompt := "You are the default agent."
	// Override to replace
	modeOverride := LayerModeReplace
	result := ApplySpell(spell, basePrompt, &modeOverride)

	// Should use override, not spell's default
	if result.SystemPrompt != spell.SystemPrompt {
		t.Errorf("SystemPrompt = %q; want %q (override to replace)", result.SystemPrompt, spell.SystemPrompt)
	}
}

func TestApplySpell_ModeOverrideToLayer(t *testing.T) {
	spell := &Spell{
		Name:         "test",
		Description:  "Test",
		SystemPrompt: "You are a test agent.",
		Mode:         LayerModeReplace, // Default is replace
	}

	basePrompt := "You are the default agent."
	// Override to layer
	modeOverride := LayerModeLayer
	result := ApplySpell(spell, basePrompt, &modeOverride)

	expected := basePrompt + "\n\n" + spell.SystemPrompt
	// Should use override, not spell's default
	if result.SystemPrompt != expected {
		t.Errorf("SystemPrompt = %q; want %q (override to layer)", result.SystemPrompt, expected)
	}
}

func TestApplySpell_ToolsConfig(t *testing.T) {
	spell := &Spell{
		Name:        "test",
		Description: "Test",
		Config: SpellConfig{
			Tools: ToolsConfig{
				Enabled:  []string{"bash", "read_file"},
				Disabled: []string{"web_search"},
			},
		},
	}

	result := ApplySpell(spell, "", nil)

	if len(result.EnabledTools) != 2 {
		t.Errorf("EnabledTools length = %d; want 2", len(result.EnabledTools))
	}
	if len(result.DisabledTools) != 1 {
		t.Errorf("DisabledTools length = %d; want 1", len(result.DisabledTools))
	}

	// Verify specific tools
	hasReadFile := false
	hasBash := false
	for _, tool := range result.EnabledTools {
		if tool == "read_file" {
			hasReadFile = true
		}
		if tool == "bash" {
			hasBash = true
		}
	}
	if !hasReadFile {
		t.Error("EnabledTools missing 'read_file'")
	}
	if !hasBash {
		t.Error("EnabledTools missing 'bash'")
	}

	hasWebSearch := false
	for _, tool := range result.DisabledTools {
		if tool == "web_search" {
			hasWebSearch = true
		}
	}
	if !hasWebSearch {
		t.Error("DisabledTools missing 'web_search'")
	}
}

func TestApplySpell_ResponseConfig(t *testing.T) {
	spell := &Spell{
		Name:        "test",
		Description: "Test",
		Config: SpellConfig{
			Response: ResponseConfig{
				MaxTokens: 8192,
			},
		},
	}

	result := ApplySpell(spell, "", nil)

	if result.MaxTokens != 8192 {
		t.Errorf("MaxTokens = %d; want 8192", result.MaxTokens)
	}
}

func TestApplySpell_ResponseFormat(t *testing.T) {
	spell := &Spell{
		Name:        "test",
		Description: "Test",
		Config: SpellConfig{
			Response: ResponseConfig{
				Format: "json",
				Style:  "concise",
			},
		},
	}

	result := ApplySpell(spell, "", nil)

	if result.ResponseFormat != "json" {
		t.Errorf("ResponseFormat = %q; want %q", result.ResponseFormat, "json")
	}
	if result.ResponseStyle != "concise" {
		t.Errorf("ResponseStyle = %q; want %q", result.ResponseStyle, "concise")
	}
}

func TestApplySpell_ReasoningConfig(t *testing.T) {
	spell := &Spell{
		Name:        "test",
		Description: "Test",
		Config: SpellConfig{
			Reasoning: ReasoningConfig{
				Effort:       ReasoningEffortHigh,
				ShowThinking: true,
			},
		},
	}

	result := ApplySpell(spell, "", nil)

	if result.ReasoningEffort != ReasoningEffortHigh {
		t.Errorf("ReasoningEffort = %q; want %q", result.ReasoningEffort, ReasoningEffortHigh)
	}
	if !result.ShowThinking {
		t.Error("ShowThinking = false; want true")
	}
}

func TestApplySpell_SamplingConfig(t *testing.T) {
	spell := &Spell{
		Name:        "test",
		Description: "Test",
		Config: SpellConfig{
			Sampling: SamplingConfig{
				Temperature: 0.7,
			},
		},
	}

	result := ApplySpell(spell, "", nil)

	if result.Temperature != 0.7 {
		t.Errorf("Temperature = %f; want %f", result.Temperature, 0.7)
	}
}

func TestApplySpell_SpellName(t *testing.T) {
	spell := &Spell{
		Name:        "my-custom-spell",
		Description: "Test",
	}

	result := ApplySpell(spell, "", nil)

	if result.SpellName != "my-custom-spell" {
		t.Errorf("SpellName = %q; want %q", result.SpellName, "my-custom-spell")
	}
}

func TestApplySpell_DefaultModeIsLayer(t *testing.T) {
	// When mode is empty, should default to layer behavior
	spell := &Spell{
		Name:         "test",
		Description:  "Test",
		SystemPrompt: "Additional instructions.",
		Mode:         "", // Empty mode
	}

	basePrompt := "Base prompt."
	result := ApplySpell(spell, basePrompt, nil)

	expected := basePrompt + "\n\n" + spell.SystemPrompt
	if result.SystemPrompt != expected {
		t.Errorf("SystemPrompt = %q; want %q (default to layer)", result.SystemPrompt, expected)
	}
}

func TestApplySpell_FullConfig(t *testing.T) {
	// Test with all config options populated
	spell := &Spell{
		Name:         "full-config",
		Description:  "Full config test",
		SystemPrompt: "Full test prompt.",
		Mode:         LayerModeReplace,
		Config: SpellConfig{
			Tools: ToolsConfig{
				Enabled:  []string{"bash", "read_file", "write_file"},
				Disabled: []string{"web_search", "dangerous_tool"},
			},
			Reasoning: ReasoningConfig{
				Effort:       ReasoningEffortMedium,
				ShowThinking: true,
			},
			Response: ResponseConfig{
				MaxTokens: 4096,
				Format:    "markdown",
				Style:     "detailed",
			},
			Sampling: SamplingConfig{
				Temperature: 0.5,
			},
		},
	}

	result := ApplySpell(spell, "ignored base", nil)

	if result.SpellName != "full-config" {
		t.Errorf("SpellName = %q; want 'full-config'", result.SpellName)
	}
	if result.SystemPrompt != "Full test prompt." {
		t.Errorf("SystemPrompt = %q; want 'Full test prompt.'", result.SystemPrompt)
	}
	if len(result.EnabledTools) != 3 {
		t.Errorf("EnabledTools length = %d; want 3", len(result.EnabledTools))
	}
	if len(result.DisabledTools) != 2 {
		t.Errorf("DisabledTools length = %d; want 2", len(result.DisabledTools))
	}
	if result.ReasoningEffort != ReasoningEffortMedium {
		t.Errorf("ReasoningEffort = %q; want 'medium'", result.ReasoningEffort)
	}
	if !result.ShowThinking {
		t.Error("ShowThinking = false; want true")
	}
	if result.MaxTokens != 4096 {
		t.Errorf("MaxTokens = %d; want 4096", result.MaxTokens)
	}
	if result.ResponseFormat != "markdown" {
		t.Errorf("ResponseFormat = %q; want 'markdown'", result.ResponseFormat)
	}
	if result.ResponseStyle != "detailed" {
		t.Errorf("ResponseStyle = %q; want 'detailed'", result.ResponseStyle)
	}
	if result.Temperature != 0.5 {
		t.Errorf("Temperature = %f; want 0.5", result.Temperature)
	}
}
