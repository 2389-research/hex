package visualization

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/hex/internal/ui/theme"
)

func TestNewTokenVisualization(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	if vis == nil {
		t.Fatal("NewTokenVisualization returned nil")
	}
	if vis.theme != th {
		t.Error("Theme not set correctly")
	}
	if vis.detailedView {
		t.Error("Should default to compact view")
	}
	if len(vis.history) != 0 {
		t.Error("History should be empty initially")
	}
}

func TestTokenVisualizationUpdate(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	t.Run("window size message", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 30}
		model, _ := vis.Update(msg)

		tv, ok := model.(*TokenVisualization)
		if !ok {
			t.Fatal("Update should return *TokenVisualization")
		}
		if tv.width != 100 {
			t.Errorf("Width = %d, want 100", tv.width)
		}
		if tv.height != 30 {
			t.Errorf("Height = %d, want 30", tv.height)
		}
	})

	t.Run("token update message", func(t *testing.T) {
		usage := TokenUsage{
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
			MaxTokens:    10000,
			ModelName:    "test-model",
		}
		msg := TokenUpdateMsg{Usage: usage}

		model, _ := vis.Update(msg)
		tv, ok := model.(*TokenVisualization)
		if !ok {
			t.Fatal("Update should return *TokenVisualization")
		}
		if tv.current.TotalTokens != 1500 {
			t.Errorf("Total tokens = %d, want 1500", tv.current.TotalTokens)
		}
	})

	t.Run("toggle detailed view", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		model, _ := vis.Update(msg)

		tv, ok := model.(*TokenVisualization)
		if !ok {
			t.Fatal("Update should return *TokenVisualization")
		}
		if !tv.detailedView {
			t.Error("Should toggle to detailed view")
		}

		// Toggle again
		model, _ = tv.Update(msg)
		tv, _ = model.(*TokenVisualization)
		if tv.detailedView {
			t.Error("Should toggle back to compact view")
		}
	})
}

func TestTokenVisualizationView(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	t.Run("view before size set", func(t *testing.T) {
		view := vis.View()
		if view != "Loading..." {
			t.Error("Should show loading before size is set")
		}
	})

	t.Run("view after size set", func(t *testing.T) {
		vis.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		view := vis.View()

		if view == "" {
			t.Error("View should not be empty after size set")
		}
	})

	t.Run("compact view", func(t *testing.T) {
		vis.detailedView = false
		vis.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		view := vis.View()

		if view == "" {
			t.Error("Compact view should not be empty")
		}
	})

	t.Run("detailed view", func(t *testing.T) {
		vis.detailedView = true
		vis.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		view := vis.View()

		if view == "" {
			t.Error("Detailed view should not be empty")
		}
	})
}

func TestUpdateUsage(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	usage := TokenUsage{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		MaxTokens:    10000,
		ModelName:    "test-model",
	}

	vis.updateUsage(usage)

	if vis.current.TotalTokens != 1500 {
		t.Errorf("TotalTokens = %d, want 1500", vis.current.TotalTokens)
	}
	if len(vis.history) != 1 {
		t.Errorf("History length = %d, want 1", len(vis.history))
	}
}

func TestUpdateTokens(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	usage := TokenUsage{
		InputTokens:  2000,
		OutputTokens: 1000,
		TotalTokens:  3000,
		MaxTokens:    10000,
		ModelName:    "test-model",
	}

	vis.UpdateTokens(usage)

	if vis.current.TotalTokens != 3000 {
		t.Errorf("TotalTokens = %d, want 3000", vis.current.TotalTokens)
	}
}

func TestGetCurrentUsage(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	usage := TokenUsage{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		MaxTokens:    10000,
		ModelName:    "test-model",
	}

	vis.UpdateTokens(usage)
	current := vis.GetCurrentUsage()

	if current.TotalTokens != 1500 {
		t.Errorf("TotalTokens = %d, want 1500", current.TotalTokens)
	}
}

func TestGetUsagePercentage(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	tests := []struct {
		name    string
		usage   TokenUsage
		wantPct float64
	}{
		{
			name: "empty",
			usage: TokenUsage{
				TotalTokens: 0,
				MaxTokens:   0,
			},
			wantPct: 0,
		},
		{
			name: "50 percent",
			usage: TokenUsage{
				TotalTokens: 5000,
				MaxTokens:   10000,
			},
			wantPct: 0.5,
		},
		{
			name: "80 percent",
			usage: TokenUsage{
				TotalTokens: 8000,
				MaxTokens:   10000,
			},
			wantPct: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vis.UpdateTokens(tt.usage)
			pct := vis.GetUsagePercentage()

			if pct != tt.wantPct {
				t.Errorf("Percentage = %f, want %f", pct, tt.wantPct)
			}
		})
	}
}

func TestIsNearCapacity(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	tests := []struct {
		name     string
		usage    TokenUsage
		wantNear bool
	}{
		{
			name: "low usage",
			usage: TokenUsage{
				TotalTokens: 5000,
				MaxTokens:   10000,
			},
			wantNear: false,
		},
		{
			name: "near capacity",
			usage: TokenUsage{
				TotalTokens: 8500,
				MaxTokens:   10000,
			},
			wantNear: true,
		},
		{
			name: "at capacity",
			usage: TokenUsage{
				TotalTokens: 9800,
				MaxTokens:   10000,
			},
			wantNear: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vis.UpdateTokens(tt.usage)
			near := vis.IsNearCapacity()

			if near != tt.wantNear {
				t.Errorf("IsNearCapacity = %v, want %v", near, tt.wantNear)
			}
		})
	}
}

func TestRenderStatusBar(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	t.Run("empty usage", func(t *testing.T) {
		statusBar := vis.RenderStatusBar()
		if statusBar != "" {
			t.Error("Status bar should be empty with no usage")
		}
	})

	t.Run("with usage", func(t *testing.T) {
		usage := TokenUsage{
			TotalTokens: 5000,
			MaxTokens:   10000,
		}
		vis.UpdateTokens(usage)

		statusBar := vis.RenderStatusBar()
		if statusBar == "" {
			t.Error("Status bar should not be empty with usage data")
		}
	})
}

func TestGenerateSparkline(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)

	history := []TokenUsage{
		{TotalTokens: 1000, MaxTokens: 10000},
		{TotalTokens: 2000, MaxTokens: 10000},
		{TotalTokens: 3000, MaxTokens: 10000},
		{TotalTokens: 2500, MaxTokens: 10000},
		{TotalTokens: 4000, MaxTokens: 10000},
	}

	sparkline := vis.generateSparkline(history, 50)

	if sparkline == "" {
		t.Error("Sparkline should not be empty")
	}
	if len([]rune(sparkline)) != len(history) {
		t.Errorf("Sparkline length = %d, want %d", len([]rune(sparkline)), len(history))
	}
}

func TestHistoryTrimming(t *testing.T) {
	th := theme.NewDraculaTheme()
	vis := NewTokenVisualization(th)
	vis.maxHistoryLen = 5

	// Add more than maxHistoryLen
	for i := 0; i < 10; i++ {
		usage := TokenUsage{
			TotalTokens: i * 1000,
			MaxTokens:   10000,
		}
		vis.UpdateTokens(usage)
	}

	if len(vis.history) != vis.maxHistoryLen {
		t.Errorf("History length = %d, want %d", len(vis.history), vis.maxHistoryLen)
	}
}

func TestTokenUsage(t *testing.T) {
	usage := TokenUsage{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		MaxTokens:    10000,
		ModelName:    "test-model",
	}

	if usage.InputTokens != 1000 {
		t.Errorf("InputTokens = %d, want 1000", usage.InputTokens)
	}
	if usage.OutputTokens != 500 {
		t.Errorf("OutputTokens = %d, want 500", usage.OutputTokens)
	}
	if usage.TotalTokens != 1500 {
		t.Errorf("TotalTokens = %d, want 1500", usage.TotalTokens)
	}
	if usage.MaxTokens != 10000 {
		t.Errorf("MaxTokens = %d, want 10000", usage.MaxTokens)
	}
	if usage.ModelName != "test-model" {
		t.Errorf("ModelName = %s, want test-model", usage.ModelName)
	}
}
