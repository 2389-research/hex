package animations

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestNewGradient(t *testing.T) {
	tests := []struct {
		name     string
		start    lipgloss.Color
		end      lipgloss.Color
		steps    int
		wantStep int
	}{
		{
			name:     "normal gradient",
			start:    lipgloss.Color("#ff0000"),
			end:      lipgloss.Color("#0000ff"),
			steps:    10,
			wantStep: 10,
		},
		{
			name:     "minimum steps enforced",
			start:    lipgloss.Color("#ff0000"),
			end:      lipgloss.Color("#0000ff"),
			steps:    1,
			wantStep: 2,
		},
		{
			name:     "negative steps defaults to minimum",
			start:    lipgloss.Color("#ff0000"),
			end:      lipgloss.Color("#0000ff"),
			steps:    -5,
			wantStep: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGradient(tt.start, tt.end, tt.steps)
			if g == nil {
				t.Fatal("NewGradient returned nil")
			}
			if g.Steps != tt.wantStep {
				t.Errorf("Steps = %d, want %d", g.Steps, tt.wantStep)
			}
			if g.StartColor != tt.start {
				t.Errorf("StartColor = %v, want %v", g.StartColor, tt.start)
			}
			if g.EndColor != tt.end {
				t.Errorf("EndColor = %v, want %v", g.EndColor, tt.end)
			}
		})
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name  string
		hex   string
		wantR int
		wantG int
		wantB int
	}{
		{
			name:  "red color with hash",
			hex:   "#ff0000",
			wantR: 255,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "blue color without hash",
			hex:   "0000ff",
			wantR: 0,
			wantG: 0,
			wantB: 255,
		},
		{
			name:  "green color",
			hex:   "#00ff00",
			wantR: 0,
			wantG: 255,
			wantB: 0,
		},
		{
			name:  "mixed color",
			hex:   "#8be9fd",
			wantR: 139,
			wantG: 233,
			wantB: 253,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := parseHexColor(tt.hex)
			if r != tt.wantR {
				t.Errorf("R = %d, want %d", r, tt.wantR)
			}
			if g != tt.wantG {
				t.Errorf("G = %d, want %d", g, tt.wantG)
			}
			if b != tt.wantB {
				t.Errorf("B = %d, want %d", b, tt.wantB)
			}
		})
	}
}

func TestInterpolateColor(t *testing.T) {
	tests := []struct {
		name  string
		start lipgloss.Color
		end   lipgloss.Color
		ratio float64
	}{
		{
			name:  "start position",
			start: lipgloss.Color("#ff0000"),
			end:   lipgloss.Color("#0000ff"),
			ratio: 0.0,
		},
		{
			name:  "end position",
			start: lipgloss.Color("#ff0000"),
			end:   lipgloss.Color("#0000ff"),
			ratio: 1.0,
		},
		{
			name:  "middle position",
			start: lipgloss.Color("#ff0000"),
			end:   lipgloss.Color("#0000ff"),
			ratio: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpolateColor(tt.start, tt.end, tt.ratio)
			if result == "" {
				t.Error("interpolateColor returned empty string")
			}
			// Verify it's a valid hex color
			if !strings.HasPrefix(string(result), "#") {
				t.Errorf("result doesn't start with #: %s", result)
			}
			if len(result) != 7 {
				t.Errorf("result length = %d, want 7", len(result))
			}
		})
	}
}

func TestGradientRender(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "simple text",
			text: "Hello",
		},
		{
			name: "empty text",
			text: "",
		},
		{
			name: "single character",
			text: "X",
		},
		{
			name: "long text",
			text: "This is a much longer text for gradient testing",
		},
	}

	gradient := NewGradient(
		lipgloss.Color("#ff0000"),
		lipgloss.Color("#0000ff"),
		10,
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gradient.Render(tt.text)
			if tt.text == "" && result != "" {
				t.Error("Expected empty result for empty text")
			}
			if tt.text != "" && result == "" {
				t.Error("Expected non-empty result for non-empty text")
			}
		})
	}
}

func TestGradientRenderBar(t *testing.T) {
	gradient := NewGradient(
		lipgloss.Color("#ff0000"),
		lipgloss.Color("#0000ff"),
		10,
	)

	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "normal width",
			width: 20,
		},
		{
			name:  "single character",
			width: 1,
		},
		{
			name:  "zero width",
			width: 0,
		},
		{
			name:  "negative width",
			width: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gradient.RenderBar(tt.width)
			if tt.width <= 0 && result != "" {
				t.Errorf("Expected empty bar for width %d", tt.width)
			}
			if tt.width > 0 && result == "" {
				t.Errorf("Expected non-empty bar for width %d", tt.width)
			}
		})
	}
}

func TestNewTransition(t *testing.T) {
	duration := 100 * time.Millisecond
	trans := NewTransition(duration, 0.0, 1.0)

	if trans == nil {
		t.Fatal("NewTransition returned nil")
	}
	if trans.Duration != duration {
		t.Errorf("Duration = %v, want %v", trans.Duration, duration)
	}
	if trans.StartVal != 0.0 {
		t.Errorf("StartVal = %f, want 0.0", trans.StartVal)
	}
	if trans.EndVal != 1.0 {
		t.Errorf("EndVal = %f, want 1.0", trans.EndVal)
	}
}

func TestTransitionValue(t *testing.T) {
	duration := 100 * time.Millisecond
	trans := NewTransition(duration, 0.0, 100.0)

	// Test initial value
	val := trans.Value()
	if val < 0 || val > 100 {
		t.Errorf("Value out of range: %f", val)
	}

	// Wait for transition to complete
	time.Sleep(duration + 10*time.Millisecond)

	// Test final value
	val = trans.Value()
	if val != 100.0 {
		t.Errorf("Final value = %f, want 100.0", val)
	}
}

func TestTransitionIsComplete(t *testing.T) {
	duration := 50 * time.Millisecond
	trans := NewTransition(duration, 0.0, 1.0)

	// Should not be complete immediately
	if trans.IsComplete() {
		t.Error("Transition should not be complete immediately")
	}

	// Wait for completion
	time.Sleep(duration + 10*time.Millisecond)

	// Should be complete now
	if !trans.IsComplete() {
		t.Error("Transition should be complete after duration")
	}
}

func TestEaseInOutCubic(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{input: 0.0, want: 0.0},
		{input: 1.0, want: 1.0},
		{input: 0.5, want: 0.5},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := easeInOutCubic(tt.input)
			// Allow small floating point error
			if result < tt.want-0.1 || result > tt.want+0.1 {
				t.Errorf("easeInOutCubic(%f) = %f, want ~%f", tt.input, result, tt.want)
			}
		})
	}
}

func TestNewPulseEffect(t *testing.T) {
	base := lipgloss.Color("#ff0000")
	highlight := lipgloss.Color("#00ff00")
	period := 1 * time.Second

	pulse := NewPulseEffect(base, highlight, period)

	if pulse == nil {
		t.Fatal("NewPulseEffect returned nil")
	}
	if pulse.BaseColor != base {
		t.Errorf("BaseColor = %v, want %v", pulse.BaseColor, base)
	}
	if pulse.HighlightColor != highlight {
		t.Errorf("HighlightColor = %v, want %v", pulse.HighlightColor, highlight)
	}
	if pulse.Period != period {
		t.Errorf("Period = %v, want %v", pulse.Period, period)
	}
}

func TestPulseEffectCurrentColor(t *testing.T) {
	pulse := NewPulseEffect(
		lipgloss.Color("#ff0000"),
		lipgloss.Color("#00ff00"),
		100*time.Millisecond,
	)

	color := pulse.CurrentColor()
	if color == "" {
		t.Error("CurrentColor returned empty string")
	}
}

func TestPulseEffectRenderWithPulse(t *testing.T) {
	pulse := NewPulseEffect(
		lipgloss.Color("#ff0000"),
		lipgloss.Color("#00ff00"),
		100*time.Millisecond,
	)

	result := pulse.RenderWithPulse("Test")
	if result == "" {
		t.Error("RenderWithPulse returned empty string")
	}
}

func TestNewShimmerEffect(t *testing.T) {
	base := lipgloss.Color("#ff0000")
	shimmer := lipgloss.Color("#00ff00")
	period := 1 * time.Second

	effect := NewShimmerEffect(base, shimmer, period)

	if effect == nil {
		t.Fatal("NewShimmerEffect returned nil")
	}
	if effect.BaseColor != base {
		t.Errorf("BaseColor = %v, want %v", effect.BaseColor, base)
	}
	if effect.ShimmerColor != shimmer {
		t.Errorf("ShimmerColor = %v, want %v", effect.ShimmerColor, shimmer)
	}
	if effect.Period != period {
		t.Errorf("Period = %v, want %v", effect.Period, period)
	}
	if effect.ShimmerWidth != 0.3 {
		t.Errorf("ShimmerWidth = %f, want 0.3", effect.ShimmerWidth)
	}
}

func TestShimmerEffectRender(t *testing.T) {
	shimmer := NewShimmerEffect(
		lipgloss.Color("#ff0000"),
		lipgloss.Color("#00ff00"),
		100*time.Millisecond,
	)

	tests := []struct {
		name string
		text string
	}{
		{name: "normal text", text: "Hello World"},
		{name: "empty text", text: ""},
		{name: "single char", text: "X"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shimmer.RenderWithShimmer(tt.text)
			if tt.text == "" && result != "" {
				t.Error("Expected empty result for empty text")
			}
			if tt.text != "" && result == "" {
				t.Error("Expected non-empty result for non-empty text")
			}
		})
	}
}

func TestNewFadeIn(t *testing.T) {
	color := lipgloss.Color("#ff0000")
	duration := 100 * time.Millisecond

	fade := NewFadeIn(color, duration)

	if fade == nil {
		t.Fatal("NewFadeIn returned nil")
	}
	if fade.Color != color {
		t.Errorf("Color = %v, want %v", fade.Color, color)
	}
	if fade.Duration != duration {
		t.Errorf("Duration = %v, want %v", fade.Duration, duration)
	}
	if !fade.FadeIn {
		t.Error("FadeIn should be true")
	}
}

func TestNewFadeOut(t *testing.T) {
	color := lipgloss.Color("#ff0000")
	duration := 100 * time.Millisecond

	fade := NewFadeOut(color, duration)

	if fade == nil {
		t.Fatal("NewFadeOut returned nil")
	}
	if fade.FadeIn {
		t.Error("FadeIn should be false")
	}
}

func TestFadeEffectOpacity(t *testing.T) {
	t.Run("fade in", func(t *testing.T) {
		fade := NewFadeIn(lipgloss.Color("#ff0000"), 100*time.Millisecond)

		// Initial opacity should be low
		opacity := fade.CurrentOpacity()
		if opacity < 0 || opacity > 1 {
			t.Errorf("Opacity out of range: %f", opacity)
		}

		// Wait for completion
		time.Sleep(110 * time.Millisecond)

		// Final opacity should be 1.0
		opacity = fade.CurrentOpacity()
		if opacity != 1.0 {
			t.Errorf("Final opacity = %f, want 1.0", opacity)
		}
	})

	t.Run("fade out", func(t *testing.T) {
		fade := NewFadeOut(lipgloss.Color("#ff0000"), 100*time.Millisecond)

		// Wait for completion
		time.Sleep(110 * time.Millisecond)

		// Final opacity should be 0.0
		opacity := fade.CurrentOpacity()
		if opacity != 0.0 {
			t.Errorf("Final opacity = %f, want 0.0", opacity)
		}
	})
}

func TestFadeEffectIsComplete(t *testing.T) {
	fade := NewFadeIn(lipgloss.Color("#ff0000"), 50*time.Millisecond)

	// Should not be complete immediately
	if fade.IsComplete() {
		t.Error("Fade should not be complete immediately")
	}

	// Wait for completion
	time.Sleep(60 * time.Millisecond)

	// Should be complete now
	if !fade.IsComplete() {
		t.Error("Fade should be complete after duration")
	}
}
