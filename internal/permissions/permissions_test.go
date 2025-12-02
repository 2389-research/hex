// ABOUTME: Comprehensive tests for permission system
// ABOUTME: Tests mode parsing, rule checking, and permission decisions

package permissions

import (
	"testing"
)

// TestParseMode tests permission mode parsing
func TestParseMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Mode
		wantErr bool
	}{
		{"ask lowercase", "ask", ModeAsk, false},
		{"ask uppercase", "ASK", ModeAsk, false},
		{"ask with spaces", "  ask  ", ModeAsk, false},
		{"auto lowercase", "auto", ModeAuto, false},
		{"auto uppercase", "AUTO", ModeAuto, false},
		{"deny lowercase", "deny", ModeDeny, false},
		{"deny uppercase", "DENY", ModeDeny, false},
		{"invalid mode", "invalid", ModeAsk, true},
		{"empty string", "", ModeAsk, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestModeString tests mode string representation
func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeAsk, "ask"},
		{ModeAuto, "auto"},
		{ModeDeny, "deny"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("Mode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRulesIsToolAllowed tests tool allow/deny rules
func TestRulesIsToolAllowed(t *testing.T) {
	tests := []struct {
		name        string
		allowed     []string
		disallowed  []string
		toolName    string
		wantAllowed bool
	}{
		// No rules - everything allowed
		{"no rules - read allowed", nil, nil, "Read", true},
		{"no rules - bash allowed", nil, nil, "Bash", true},

		// Allowed list only
		{"allowed list - read allowed", []string{"Read"}, nil, "Read", true},
		{"allowed list - bash denied", []string{"Read"}, nil, "Bash", false},
		{"allowed list - case insensitive", []string{"read"}, nil, "Read", true},
		{"allowed list - with suffix", []string{"read"}, nil, "read_file", true},

		// Disallowed list only
		{"disallowed list - read allowed", nil, []string{"Bash"}, "Read", true},
		{"disallowed list - bash denied", nil, []string{"Bash"}, "Bash", false},
		{"disallowed list - case insensitive", nil, []string{"bash"}, "Bash", false},

		// Both lists - disallowed takes precedence
		{"both lists - disallowed wins", []string{"Bash"}, []string{"Bash"}, "Bash", false},
		{"both lists - allowed and not disallowed", []string{"Read"}, []string{"Bash"}, "Read", true},
		{"both lists - not in allowed", []string{"Read"}, []string{"Bash"}, "Write", false},

		// Tool name variations
		{"variation - ReadFile vs Read", []string{"Read"}, nil, "ReadFile", true},
		{"variation - read_file vs Read", []string{"Read"}, nil, "read_file", true},
		{"variation - READFILE vs read", []string{"read"}, nil, "READFILE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := NewRules(tt.allowed, tt.disallowed)
			got := rules.IsToolAllowed(tt.toolName)
			if got != tt.wantAllowed {
				t.Errorf("IsToolAllowed(%q) = %v, want %v", tt.toolName, got, tt.wantAllowed)
			}
		})
	}
}

// TestCheckerModeAsk tests ask mode behavior
func TestCheckerModeAsk(t *testing.T) {
	checker := NewChecker(ModeAsk, NewRules(nil, nil))

	result := checker.Check("Read", nil)

	if result.Allowed {
		t.Error("ModeAsk should not auto-allow")
	}
	if !result.RequiresPrompt {
		t.Error("ModeAsk should require prompt")
	}
	if result.ToolName != "Read" {
		t.Errorf("ToolName = %q, want %q", result.ToolName, "Read")
	}
}

// TestCheckerModeAuto tests auto mode behavior
func TestCheckerModeAuto(t *testing.T) {
	checker := NewChecker(ModeAuto, NewRules(nil, nil))

	result := checker.Check("Bash", nil)

	if !result.Allowed {
		t.Error("ModeAuto should auto-allow")
	}
	if result.RequiresPrompt {
		t.Error("ModeAuto should not require prompt")
	}
}

// TestCheckerModeDeny tests deny mode behavior
func TestCheckerModeDeny(t *testing.T) {
	checker := NewChecker(ModeDeny, NewRules(nil, nil))

	result := checker.Check("Read", nil)

	if result.Allowed {
		t.Error("ModeDeny should deny all tools")
	}
	if result.RequiresPrompt {
		t.Error("ModeDeny should not require prompt")
	}
}

// TestCheckerWithAllowedTools tests allow list filtering
func TestCheckerWithAllowedTools(t *testing.T) {
	rules := NewRules([]string{"Read", "Write"}, nil)
	checker := NewChecker(ModeAuto, rules)

	tests := []struct {
		toolName string
		want     bool
	}{
		{"Read", true},
		{"Write", true},
		{"Bash", false},
		{"Edit", false},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			result := checker.Check(tt.toolName, nil)
			if result.Allowed != tt.want {
				t.Errorf("Check(%q).Allowed = %v, want %v", tt.toolName, result.Allowed, tt.want)
			}
		})
	}
}

// TestCheckerWithDisallowedTools tests deny list filtering
func TestCheckerWithDisallowedTools(t *testing.T) {
	rules := NewRules(nil, []string{"Bash", "Write"})
	checker := NewChecker(ModeAuto, rules)

	tests := []struct {
		toolName string
		want     bool
	}{
		{"Read", true},
		{"Bash", false},
		{"Write", false},
		{"Edit", true},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			result := checker.Check(tt.toolName, nil)
			if result.Allowed != tt.want {
				t.Errorf("Check(%q).Allowed = %v, want %v", tt.toolName, result.Allowed, tt.want)
			}
		})
	}
}

// TestCheckerPrecedence tests that disallowed takes precedence over allowed
func TestCheckerPrecedence(t *testing.T) {
	rules := NewRules([]string{"Bash"}, []string{"Bash"})
	checker := NewChecker(ModeAuto, rules)

	result := checker.Check("Bash", nil)

	if result.Allowed {
		t.Error("Disallowed should take precedence over allowed")
	}
}

// TestConfigValidate tests config validation
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			"default config",
			DefaultConfig(),
			false,
		},
		{
			"auto mode with allowed tools",
			&Config{Mode: ModeAuto, AllowedTools: []string{"Read"}},
			false,
		},
		{
			"ask mode with disallowed tools",
			&Config{Mode: ModeAsk, DisallowedTools: []string{"Bash"}},
			false,
		},
		{
			"both allowed and disallowed (valid)",
			&Config{Mode: ModeAsk, AllowedTools: []string{"Read"}, DisallowedTools: []string{"Bash"}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfigToChecker tests converting config to checker
func TestConfigToChecker(t *testing.T) {
	config := &Config{
		Mode:            ModeAuto,
		AllowedTools:    []string{"Read"},
		DisallowedTools: []string{"Bash"},
	}

	checker, err := config.ToChecker()
	if err != nil {
		t.Fatalf("ToChecker() error = %v", err)
	}

	if checker.GetMode() != ModeAuto {
		t.Errorf("checker mode = %v, want %v", checker.GetMode(), ModeAuto)
	}

	// Test that rules work
	if !checker.Check("Read", nil).Allowed {
		t.Error("Read should be allowed")
	}
	if checker.Check("Bash", nil).Allowed {
		t.Error("Bash should be denied (in disallowed list)")
	}
	if checker.Check("Write", nil).Allowed {
		t.Error("Write should be denied (not in allowed list)")
	}
}

// TestCheckerHelpers tests helper methods
func TestCheckerHelpers(t *testing.T) {
	tests := []struct {
		name            string
		mode            Mode
		wantAutoApprove bool
		wantDenyAll     bool
	}{
		{"ModeAsk", ModeAsk, false, false},
		{"ModeAuto", ModeAuto, true, false},
		{"ModeDeny", ModeDeny, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewChecker(tt.mode, nil)

			if got := checker.ShouldAutoApprove(); got != tt.wantAutoApprove {
				t.Errorf("ShouldAutoApprove() = %v, want %v", got, tt.wantAutoApprove)
			}
			if got := checker.ShouldDenyAll(); got != tt.wantDenyAll {
				t.Errorf("ShouldDenyAll() = %v, want %v", got, tt.wantDenyAll)
			}
		})
	}
}

// TestRulesHasLists tests HasAllowList and HasDenyList
func TestRulesHasLists(t *testing.T) {
	tests := []struct {
		name          string
		allowed       []string
		disallowed    []string
		wantAllowList bool
		wantDenyList  bool
	}{
		{"no lists", nil, nil, false, false},
		{"allow list only", []string{"Read"}, nil, true, false},
		{"deny list only", nil, []string{"Bash"}, false, true},
		{"both lists", []string{"Read"}, []string{"Bash"}, true, true},
		{"empty lists", []string{}, []string{}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := NewRules(tt.allowed, tt.disallowed)

			if got := rules.HasAllowList(); got != tt.wantAllowList {
				t.Errorf("HasAllowList() = %v, want %v", got, tt.wantAllowList)
			}
			if got := rules.HasDenyList(); got != tt.wantDenyList {
				t.Errorf("HasDenyList() = %v, want %v", got, tt.wantDenyList)
			}
		})
	}
}

// TestNormalizeToolName tests tool name normalization
func TestNormalizeToolName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Read", "read"},
		{"read_file", "readfile"},
		{"ReadFile", "readfile"},
		{"BASH_TOOL", "bashtool"},
		{"write", "write"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := normalizeToolName(tt.input); got != tt.want {
				t.Errorf("normalizeToolName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestMatchToolName tests flexible tool name matching
func TestMatchToolName(t *testing.T) {
	tests := []struct {
		pattern  string
		toolName string
		want     bool
	}{
		{"Read", "Read", true},
		{"read", "Read", true},
		{"Read", "read_file", true},
		{"read", "ReadFile", true},
		{"Bash", "bash_tool", true},
		{"Write", "Read", false},
		{"Edit", "EditFile", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_vs_"+tt.toolName, func(t *testing.T) {
			if got := matchToolName(tt.pattern, tt.toolName); got != tt.want {
				t.Errorf("matchToolName(%q, %q) = %v, want %v", tt.pattern, tt.toolName, got, tt.want)
			}
		})
	}
}

// TestConfigString tests config string representation
func TestConfigString(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			"default config",
			DefaultConfig(),
		},
		{
			"with allowed tools",
			&Config{Mode: ModeAuto, AllowedTools: []string{"Read", "Write"}},
		},
		{
			"with disallowed tools",
			&Config{Mode: ModeAsk, DisallowedTools: []string{"Bash"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.config.String()
			if str == "" {
				t.Error("String() returned empty string")
			}
			// Just verify it doesn't panic and returns something
		})
	}
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Run("nil rules to checker", func(t *testing.T) {
		checker := NewChecker(ModeAsk, nil)
		result := checker.Check("Read", nil)
		// Should handle gracefully with default empty rules
		if !result.RequiresPrompt {
			t.Error("Should require prompt with nil rules")
		}
	})

	t.Run("empty tool name", func(t *testing.T) {
		rules := NewRules([]string{"Read"}, nil)
		// Empty tool name should not match
		if rules.IsToolAllowed("") {
			t.Error("Empty tool name should not be allowed")
		}
	})

	t.Run("whitespace in tool lists", func(t *testing.T) {
		rules := NewRules([]string{"  Read  ", "Write"}, []string{" Bash "})
		if !rules.IsToolAllowed("Read") {
			t.Error("Should handle whitespace in allowed list")
		}
		if rules.IsToolAllowed("Bash") {
			t.Error("Should handle whitespace in disallowed list")
		}
	})
}
