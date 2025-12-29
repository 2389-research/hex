// ABOUTME: Tests for Spell type definitions and validation
// ABOUTME: Verifies spell validation and LayerMode constants

package spells

import (
	"testing"
)

func TestSpellValidation(t *testing.T) {
	tests := []struct {
		name    string
		spell   Spell
		wantErr bool
	}{
		{
			name: "valid spell",
			spell: Spell{
				Name:         "test",
				Description:  "Test spell",
				SystemPrompt: "You are a test assistant.",
				Mode:         LayerModeReplace,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			spell: Spell{
				Description:  "Test spell",
				SystemPrompt: "You are a test assistant.",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			spell: Spell{
				Name:         "test",
				SystemPrompt: "You are a test assistant.",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spell.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLayerModeConstants(t *testing.T) {
	if LayerModeReplace != "replace" {
		t.Errorf("LayerModeReplace = %q; want %q", LayerModeReplace, "replace")
	}
	if LayerModeLayer != "layer" {
		t.Errorf("LayerModeLayer = %q; want %q", LayerModeLayer, "layer")
	}
}
