// ABOUTME: Bridge between mux hooks and hex's hook engine.
// ABOUTME: Forwards mux lifecycle events to hex's shell-based hook system.

package hooks

import (
	"context"

	muxhooks "github.com/2389-research/mux/hooks"
)

// MuxBridge bridges mux's hook system to hex's hook engine.
// It registers handlers on a mux HookManager that forward events
// to the hex hook engine for shell command execution.
type MuxBridge struct {
	hexEngine  *Engine
	muxManager *muxhooks.Manager
}

// NewMuxBridge creates a bridge between mux hooks and the hex engine.
// Returns both the bridge and the mux HookManager to wire into agents.
func NewMuxBridge(hexEngine *Engine) (*MuxBridge, *muxhooks.Manager) {
	muxMgr := muxhooks.NewManager()
	bridge := &MuxBridge{
		hexEngine:  hexEngine,
		muxManager: muxMgr,
	}
	bridge.registerHandlers()
	return bridge, muxMgr
}

// Manager returns the mux hook manager for wiring into agents.
func (b *MuxBridge) Manager() *muxhooks.Manager {
	return b.muxManager
}

// registerHandlers wires up mux hooks to forward to hex engine.
func (b *MuxBridge) registerHandlers() {
	// SessionStart: mux → hex
	b.muxManager.OnSessionStart(func(_ context.Context, event *muxhooks.SessionStartEvent) error {
		// Map mux event to hex event
		// Note: mux has SessionID, Source, Prompt; hex has ProjectPath, ModelID
		// We use Source as a stand-in; hex's SessionStart is fired separately with full data
		return nil // Hex fires its own SessionStart with more complete data
	})

	// SessionEnd: mux → hex
	b.muxManager.OnSessionEnd(func(_ context.Context, event *muxhooks.SessionEndEvent) error {
		// Note: hex's SessionEnd needs ProjectPath and MessageCount
		// We don't have this from mux's event, so let hex fire its own
		return nil // Hex fires its own SessionEnd with more complete data
	})

	// Stop: mux → hex
	b.muxManager.OnStop(func(_ context.Context, event *muxhooks.StopEvent) error {
		// Map mux Stop to hex Stop
		// Mux has: SessionID, FinalText, Continue
		// Hex has: ResponseLength, TokensUsed, ToolsUsed, IsSubagent
		return b.hexEngine.FireStop(
			len(event.FinalText), // ResponseLength
			0,                    // TokensUsed - not available from mux
			nil,                  // ToolsUsed - not available from mux
			false,                // IsSubagent - root agent stop
		)
	})

	// SubagentStart: mux → hex (new event type needed in hex)
	b.muxManager.OnSubagentStart(func(_ context.Context, event *muxhooks.SubagentStartEvent) error {
		// Hex doesn't have SubagentStart yet, fire as notification for now
		return b.hexEngine.FireNotification(
			"info",
			"Subagent started: "+event.Name,
			"mux",
		)
	})

	// SubagentStop: mux → hex
	b.muxManager.OnSubagentStop(func(_ context.Context, event *muxhooks.SubagentStopEvent) error {
		// Map mux SubagentStop to hex SubagentStop
		// Mux has: ParentID, ChildID, Name, Error
		// Hex has: TaskDescription, SubagentType, ResponseLength, TokensUsed, Success, ExecutionTime
		success := event.Error == nil
		return b.hexEngine.FireSubagentStop(
			event.Name, // TaskDescription (using name as description)
			"mux",      // SubagentType
			0,          // ResponseLength - not available
			0,          // TokensUsed - not available
			success,    // Success
			0,          // ExecutionTime - not available
		)
	})
}
