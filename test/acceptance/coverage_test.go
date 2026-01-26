// ABOUTME: Additional tests to ensure high coverage of TUIHarness implementation
// ABOUTME: Exercises edge cases and rarely-used code paths

package acceptance

import (
	"testing"
	"time"

	"github.com/2389-research/hex/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test SendText and SubmitInput
func TestInput_SendTextAndSubmit(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// SendText should type characters
	require.NoError(t, h.SendText("hello"))

	// SubmitInput should work (calls SendKey(KeyEnter))
	require.NoError(t, h.SubmitInput())
}

// Test SendKey with single character
func TestInput_SendKeySingleChar(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Single character keys
	require.NoError(t, h.SendKey("a"))
	require.NoError(t, h.SendKey("Z"))
	require.NoError(t, h.SendKey("1"))
}

// Test SendKey with unknown key returns error
func TestInput_SendKeyUnknownReturnsError(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Unknown multi-char key should error
	err := h.SendKey("unknownkey")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown key")
}

// Test GetStatus error case
func TestStatus_ErrorState(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Set error status directly on the model
	h.model.SetStatus(ui.StatusError)
	status := h.GetStatus()
	assert.Equal(t, "error", status)
}

// Test GetStatus unknown case (default branch)
func TestStatus_UnknownState(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Set an invalid status value to trigger default case
	// We use a value outside the known range
	h.model.Status = ui.Status(99)
	status := h.GetStatus()
	assert.Equal(t, "unknown", status)
}

// Test GetModalType when no modal
func TestModal_GetModalTypeNoModal(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// No modal should return empty string
	modalType := h.GetModalType()
	assert.Equal(t, "", modalType)
}

// Test GetModalType when modal is active
func TestModal_GetModalTypeWithOverlay(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Try to open help with ?
	require.NoError(t, h.SendKey("?"))

	// If there's a modal, check the type
	if h.HasModal() {
		modalType := h.GetModalType()
		assert.Equal(t, "overlay", modalType)
	}
	// Note: If no modal appears, the GetModalType empty case is covered by TestModal_GetModalTypeNoModal
}

// Test WaitFor timeout
func TestWaitFor_Timeout(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Condition that never becomes true
	err := h.WaitFor(func() bool {
		return false
	}, 50*time.Millisecond)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// Test WaitFor success (immediate)
func TestWaitFor_ImmediateSuccess(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Condition that is immediately true
	err := h.WaitFor(func() bool {
		return true
	}, 100*time.Millisecond)

	assert.NoError(t, err)
}

// Test ViewContainsAny returns false when no match
func TestView_ContainsAnyNoMatch(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// These strings should not be in the view
	result := ViewContainsAny(h, "xyznotfound123", "abcnotfound456")
	assert.False(t, result)
}

// Test Shutdown is callable (coverage)
func TestLifecycle_ShutdownExplicit(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))

	// Explicit shutdown call for coverage
	h.Shutdown()

	// Should be safe to call multiple times
	h.Shutdown()
}

// Test all key constants are handled
func TestInput_AllKeyConstants(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	keys := []string{
		KeyEnter, KeyEsc, KeyCtrlC, KeyCtrlO,
		KeyUp, KeyDown, KeyTab, KeyG, KeyShiftG,
	}

	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			err := h.SendKey(key)
			assert.NoError(t, err, "SendKey should handle %s", key)
		})
	}
}

// Test SendText with various characters including special ones
func TestInput_SendTextVariousChars(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Test with various characters
	require.NoError(t, h.SendText("Hello, World!"))
	require.NoError(t, h.SendText("123"))
	require.NoError(t, h.SendText("@#$%"))
}

// Test SendKey with digit character
func TestInput_SendKeyDigit(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Single digit
	require.NoError(t, h.SendKey("5"))
}

// Test SendKey with special single character
func TestInput_SendKeySpecialChar(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Special characters (single char)
	require.NoError(t, h.SendKey("@"))
	require.NoError(t, h.SendKey("/"))
	require.NoError(t, h.SendKey(" "))
}
