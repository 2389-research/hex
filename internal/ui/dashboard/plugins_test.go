package dashboard

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/clem/internal/ui/theme"
)

func TestNewPluginDashboard(t *testing.T) {
	th := theme.NewDraculaTheme()
	dashboard := NewPluginDashboard(th)

	if dashboard == nil {
		t.Fatal("NewPluginDashboard returned nil")
	}
	if dashboard.theme != th {
		t.Error("Theme not set correctly")
	}
	if dashboard.showingMCP {
		t.Error("Should default to showing plugins")
	}
	if len(dashboard.plugins) != 0 {
		t.Error("Plugins should be empty initially")
	}
	if len(dashboard.mcpServers) != 0 {
		t.Error("MCP servers should be empty initially")
	}
}

func TestPluginDashboardInit(t *testing.T) {
	th := theme.NewDraculaTheme()
	dashboard := NewPluginDashboard(th)

	cmd := dashboard.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestPluginDashboardUpdate(t *testing.T) {
	th := theme.NewDraculaTheme()
	dashboard := NewPluginDashboard(th)

	t.Run("window size message", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 30}
		model, _ := dashboard.Update(msg)

		pd, ok := model.(*PluginDashboard)
		if !ok {
			t.Fatal("Update should return *PluginDashboard")
		}
		if pd.width != 100 {
			t.Errorf("Width = %d, want 100", pd.width)
		}
		if pd.height != 30 {
			t.Errorf("Height = %d, want 30", pd.height)
		}
	})

	t.Run("plugin data message", func(t *testing.T) {
		plugins := []PluginInfo{
			{Name: "test", Version: "1.0", Enabled: true, Status: "Active"},
		}
		servers := []MCPServerInfo{
			{Name: "server", URL: "local://test", Connected: true, Status: "Connected"},
		}

		msg := pluginDataMsg{plugins: plugins, mcpServers: servers}
		model, _ := dashboard.Update(msg)

		pd, ok := model.(*PluginDashboard)
		if !ok {
			t.Fatal("Update should return *PluginDashboard")
		}
		if len(pd.plugins) != 1 {
			t.Errorf("Got %d plugins, want 1", len(pd.plugins))
		}
		if len(pd.mcpServers) != 1 {
			t.Errorf("Got %d MCP servers, want 1", len(pd.mcpServers))
		}
	})

	t.Run("tab key toggles view", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyTab}
		model, _ := dashboard.Update(msg)

		pd, ok := model.(*PluginDashboard)
		if !ok {
			t.Fatal("Update should return *PluginDashboard")
		}
		if !pd.showingMCP {
			t.Error("Tab should toggle to MCP view")
		}

		// Toggle again
		model, _ = pd.Update(msg)
		pd, _ = model.(*PluginDashboard)
		if pd.showingMCP {
			t.Error("Tab should toggle back to plugins view")
		}
	})

	t.Run("keyboard quit", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		_, cmd := dashboard.Update(msg)

		if cmd == nil {
			t.Error("Quit command should not be nil")
		}
	})
}

func TestPluginDashboardView(t *testing.T) {
	th := theme.NewDraculaTheme()
	dashboard := NewPluginDashboard(th)

	t.Run("view before size set", func(t *testing.T) {
		view := dashboard.View()
		if view != "Loading..." {
			t.Error("Should show loading before size is set")
		}
	})

	t.Run("view after size set", func(t *testing.T) {
		dashboard.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		view := dashboard.View()

		if view == "" {
			t.Error("View should not be empty after size set")
		}
		if view == "Loading..." {
			t.Error("Should not show loading after size set")
		}
	})

	t.Run("plugin view title", func(t *testing.T) {
		dashboard.showingMCP = false
		dashboard.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		view := dashboard.View()

		if view == "" {
			t.Error("View should not be empty")
		}
		// Title is styled, so we can't do exact match
	})

	t.Run("mcp view title", func(t *testing.T) {
		dashboard.showingMCP = true
		dashboard.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		view := dashboard.View()

		if view == "" {
			t.Error("View should not be empty")
		}
	})
}

func TestUpdateTable(t *testing.T) {
	th := theme.NewDraculaTheme()
	dashboard := NewPluginDashboard(th)

	plugins := []PluginInfo{
		{Name: "test1", Version: "1.0", Enabled: true, Status: "Active"},
		{Name: "test2", Version: "2.0", Enabled: false, Status: "Disabled"},
	}
	servers := []MCPServerInfo{
		{Name: "server1", URL: "url1", Connected: true, Status: "Connected"},
		{Name: "server2", URL: "url2", Connected: false, Status: "Disconnected"},
	}

	dashboard.plugins = plugins
	dashboard.mcpServers = servers

	t.Run("update table with plugins", func(_ *testing.T) {
		dashboard.showingMCP = false
		dashboard.updateTable()
		// Table rows are internal, just verify no panic
	})

	t.Run("update table with mcp", func(_ *testing.T) {
		dashboard.showingMCP = true
		dashboard.updateTable()
		// Table rows are internal, just verify no panic
	})
}

func TestLoadPluginData(t *testing.T) {
	th := theme.NewDraculaTheme()
	dashboard := NewPluginDashboard(th)

	msg := dashboard.loadPluginData()

	dataMsg, ok := msg.(pluginDataMsg)
	if !ok {
		t.Fatal("loadPluginData should return pluginDataMsg")
	}

	if len(dataMsg.plugins) == 0 {
		t.Error("Should load some mock plugins")
	}
	if len(dataMsg.mcpServers) == 0 {
		t.Error("Should load some mock MCP servers")
	}
}

func TestGetPlugins(t *testing.T) {
	th := theme.NewDraculaTheme()
	dashboard := NewPluginDashboard(th)

	plugins := []PluginInfo{
		{Name: "test", Version: "1.0", Enabled: true, Status: "Active"},
	}
	dashboard.plugins = plugins

	result := dashboard.GetPlugins()
	if len(result) != 1 {
		t.Errorf("Got %d plugins, want 1", len(result))
	}
	if result[0].Name != "test" {
		t.Errorf("Got plugin name %s, want test", result[0].Name)
	}
}

func TestGetMCPServers(t *testing.T) {
	th := theme.NewDraculaTheme()
	dashboard := NewPluginDashboard(th)

	servers := []MCPServerInfo{
		{Name: "server", URL: "url", Connected: true, Status: "Connected"},
	}
	dashboard.mcpServers = servers

	result := dashboard.GetMCPServers()
	if len(result) != 1 {
		t.Errorf("Got %d servers, want 1", len(result))
	}
	if result[0].Name != "server" {
		t.Errorf("Got server name %s, want server", result[0].Name)
	}
}

func TestRenderCompact(t *testing.T) {
	th := theme.NewDraculaTheme()
	dashboard := NewPluginDashboard(th)

	t.Run("empty data", func(t *testing.T) {
		compact := dashboard.RenderCompact(80)
		if compact != "" {
			t.Error("Should return empty string when no data")
		}
	})

	t.Run("with data", func(t *testing.T) {
		dashboard.plugins = []PluginInfo{
			{Name: "p1", Enabled: true},
			{Name: "p2", Enabled: false},
		}
		dashboard.mcpServers = []MCPServerInfo{
			{Name: "m1", Connected: true},
			{Name: "m2", Connected: true},
		}

		compact := dashboard.RenderCompact(80)
		if compact == "" {
			t.Error("Should return non-empty string with data")
		}
	})
}

func TestPluginInfo(t *testing.T) {
	plugin := PluginInfo{
		Name:    "test-plugin",
		Version: "1.0.0",
		Enabled: true,
		Status:  "Active",
	}

	if plugin.Name != "test-plugin" {
		t.Errorf("Name = %s, want test-plugin", plugin.Name)
	}
	if plugin.Version != "1.0.0" {
		t.Errorf("Version = %s, want 1.0.0", plugin.Version)
	}
	if !plugin.Enabled {
		t.Error("Enabled should be true")
	}
	if plugin.Status != "Active" {
		t.Errorf("Status = %s, want Active", plugin.Status)
	}
}

func TestMCPServerInfo(t *testing.T) {
	server := MCPServerInfo{
		Name:      "test-server",
		URL:       "local://test",
		Connected: true,
		Status:    "Connected",
	}

	if server.Name != "test-server" {
		t.Errorf("Name = %s, want test-server", server.Name)
	}
	if server.URL != "local://test" {
		t.Errorf("URL = %s, want local://test", server.URL)
	}
	if !server.Connected {
		t.Error("Connected should be true")
	}
	if server.Status != "Connected" {
		t.Errorf("Status = %s, want Connected", server.Status)
	}
}
