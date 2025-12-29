// ABOUTME: Applies spell configuration to session settings
// ABOUTME: Handles system prompt layering and config merging

package spells

// AppliedSpell contains the result of applying a spell to a session
type AppliedSpell struct {
	// SpellName is the name of the spell that was applied
	SpellName string

	// SystemPrompt is the final system prompt after applying mode logic
	SystemPrompt string

	// Tool configuration
	EnabledTools  []string
	DisabledTools []string

	// Response configuration
	MaxTokens      int
	ResponseFormat string
	ResponseStyle  string

	// Reasoning configuration
	ReasoningEffort ReasoningEffort
	ShowThinking    bool

	// Sampling configuration
	Temperature float64
}

// ApplySpell applies a spell to the session configuration
// basePrompt is the existing system prompt (from CLAUDE.md, etc.)
// modeOverride allows the user to override the spell's default mode
func ApplySpell(spell *Spell, basePrompt string, modeOverride *LayerMode) *AppliedSpell {
	result := &AppliedSpell{
		SpellName: spell.Name,
	}

	// Determine effective mode
	mode := spell.Mode
	if modeOverride != nil {
		mode = *modeOverride
	}

	// Apply system prompt based on mode
	switch mode {
	case LayerModeReplace:
		result.SystemPrompt = spell.SystemPrompt
	case LayerModeLayer:
		result.SystemPrompt = layerPrompts(basePrompt, spell.SystemPrompt)
	default:
		// Default to layer behavior when mode is empty or unknown
		result.SystemPrompt = layerPrompts(basePrompt, spell.SystemPrompt)
	}

	// Apply tools config
	result.EnabledTools = spell.Config.Tools.Enabled
	result.DisabledTools = spell.Config.Tools.Disabled

	// Apply response config
	result.MaxTokens = spell.Config.Response.MaxTokens
	result.ResponseFormat = spell.Config.Response.Format
	result.ResponseStyle = spell.Config.Response.Style

	// Apply reasoning config
	result.ReasoningEffort = spell.Config.Reasoning.Effort
	result.ShowThinking = spell.Config.Reasoning.ShowThinking

	// Apply sampling config
	result.Temperature = spell.Config.Sampling.Temperature

	return result
}

// layerPrompts combines base and spell prompts with proper separator
func layerPrompts(base, spell string) string {
	if base != "" && spell != "" {
		return base + "\n\n" + spell
	} else if spell != "" {
		return spell
	}
	return base
}
