// ABOUTME: Tests for tool registry
// ABOUTME: Validates tool registration, retrieval, and thread safety

package tools_test

import (
	"sync"
	"testing"

	"github.com/harper/hex/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	registry := tools.NewRegistry()
	assert.NotNil(t, registry)

	// Should start empty
	names := registry.List()
	assert.Empty(t, names)
}

func TestRegistry_Register(t *testing.T) {
	registry := tools.NewRegistry()

	tool := &tools.MockTool{
		NameValue:        "test_tool",
		DescriptionValue: "A test tool",
	}

	err := registry.Register(tool)
	require.NoError(t, err)

	// Verify it's in the list
	names := registry.List()
	assert.Contains(t, names, "test_tool")
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	registry := tools.NewRegistry()

	tool1 := &tools.MockTool{
		NameValue: "test_tool",
	}
	tool2 := &tools.MockTool{
		NameValue: "test_tool",
	}

	err := registry.Register(tool1)
	require.NoError(t, err)

	// Registering duplicate should fail
	err = registry.Register(tool2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_Get_Existing(t *testing.T) {
	registry := tools.NewRegistry()

	tool := &tools.MockTool{
		NameValue:        "test_tool",
		DescriptionValue: "A test tool",
	}

	err := registry.Register(tool)
	require.NoError(t, err)

	// Retrieve the tool
	retrieved, err := registry.Get("test_tool")
	require.NoError(t, err)
	assert.Equal(t, "test_tool", retrieved.Name())
	assert.Equal(t, "A test tool", retrieved.Description())
}

func TestRegistry_Get_NonExistent(t *testing.T) {
	registry := tools.NewRegistry()

	tool, err := registry.Get("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, tool)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_List(t *testing.T) {
	registry := tools.NewRegistry()

	tool1 := &tools.MockTool{NameValue: "tool_a"}
	tool2 := &tools.MockTool{NameValue: "tool_c"}
	tool3 := &tools.MockTool{NameValue: "tool_b"}

	require.NoError(t, registry.Register(tool1))
	require.NoError(t, registry.Register(tool2))
	require.NoError(t, registry.Register(tool3))

	// List should be sorted alphabetically
	names := registry.List()
	assert.Len(t, names, 3)
	assert.Equal(t, []string{"tool_a", "tool_b", "tool_c"}, names)
}

func TestRegistry_List_Empty(t *testing.T) {
	registry := tools.NewRegistry()

	names := registry.List()
	assert.Empty(t, names)
	assert.NotNil(t, names) // Should be empty slice, not nil
}

func TestRegistry_ThreadSafety(t *testing.T) {
	registry := tools.NewRegistry()

	// Concurrently register tools
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			tool := &tools.MockTool{
				NameValue: string(rune('a' + i)),
			}
			_ = registry.Register(tool)
		}(i)
	}

	wg.Wait()

	// Should have registered all tools
	names := registry.List()
	assert.Len(t, names, numGoroutines)
}

func TestRegistry_ConcurrentGetAndRegister(t *testing.T) {
	registry := tools.NewRegistry()

	// Register initial tool
	tool := &tools.MockTool{NameValue: "initial"}
	require.NoError(t, registry.Register(tool))

	// Concurrently get and register
	var wg sync.WaitGroup
	numReaders := 5
	numWriters := 5

	// Readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = registry.Get("initial")
			}
		}()
	}

	// Writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tool := &tools.MockTool{
				NameValue: string(rune('a' + i)),
			}
			_ = registry.Register(tool)
		}(i)
	}

	wg.Wait()

	// Should still be able to get the initial tool
	retrieved, err := registry.Get("initial")
	require.NoError(t, err)
	assert.Equal(t, "initial", retrieved.Name())
}
