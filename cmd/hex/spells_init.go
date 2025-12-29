// ABOUTME: Spell system initialization for CLI
// ABOUTME: Loads spells and provides helper functions for applying spells

package main

import (
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/spells"
)

// initializeSpellsWithDir loads spells from a specific directory (for testing)
func initializeSpellsWithDir(dir string) *spells.Registry {
	loader := &spells.Loader{
		UserDir: dir,
	}

	registry := spells.NewRegistry()

	allSpells, err := loader.LoadAll()
	if err != nil {
		logging.WarnWith("Failed to load spells", "error", err.Error())
		return registry
	}

	for _, spell := range allSpells {
		if err := registry.Register(spell); err != nil {
			logging.WarnWith("Failed to register spell", "name", spell.Name, "error", err.Error())
		}
	}

	return registry
}

// getSpellSystemPromptWithDir applies a spell from a specific directory (for testing)
func getSpellSystemPromptWithDir(dir, spellName, basePrompt, modeOverride string) (string, error) {
	loader := &spells.Loader{
		UserDir: dir,
	}
	return getSpellSystemPromptWithLoader(loader, spellName, basePrompt, modeOverride)
}

// getSpellSystemPromptWithLoader applies a spell using the provided loader
func getSpellSystemPromptWithLoader(loader *spells.Loader, spellName, basePrompt, modeOverride string) (string, error) {
	prompt, _, err := getSpellSystemPromptAndMode(loader, spellName, basePrompt, modeOverride)
	return prompt, err
}

// getSpellSystemPromptAndMode applies a spell and returns both the system prompt and effective mode
func getSpellSystemPromptAndMode(loader *spells.Loader, spellName, basePrompt, modeOverride string) (string, spells.LayerMode, error) {
	spell, err := loader.LoadByName(spellName)
	if err != nil {
		return "", "", err
	}

	var mode *spells.LayerMode
	if modeOverride != "" {
		m := spells.LayerMode(modeOverride)
		mode = &m
	}

	applied := spells.ApplySpell(spell, basePrompt, mode)

	// Determine effective mode
	effectiveMode := spell.Mode
	if mode != nil {
		effectiveMode = *mode
	}

	return applied.SystemPrompt, effectiveMode, nil
}

// getSpellWithMode loads a spell and applies it, returning prompt and mode
func getSpellWithMode(spellName, basePrompt, modeOverride string) (string, spells.LayerMode, error) {
	loader := spells.NewLoader()
	return getSpellSystemPromptAndMode(loader, spellName, basePrompt, modeOverride)
}
