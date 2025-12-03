// Package dashboard provides status dashboards for plugins and MCP servers.
// ABOUTME: Plugin and MCP server status dashboard with health monitoring
// ABOUTME: Displays installed plugins, MCP servers, and their connection status
package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/hex/internal/ui/theme"
)

// PluginInfo represents information about a plugin
type PluginInfo struct {
	Name    string
	Version string
	Enabled bool
	Status  string
}

// MCPServerInfo represents information about an MCP server
type MCPServerInfo struct {
	Name      string
	URL       string
	Connected bool
	Status    string
}

// PluginDashboard displays plugin and MCP server status
type PluginDashboard struct {
	theme      *theme.Theme
	plugins    []PluginInfo
	mcpServers []MCPServerInfo
	table      table.Model
	width      int
	height     int
	showingMCP bool // Toggle between plugins and MCP view
}

// NewPluginDashboard creates a new plugin dashboard
func NewPluginDashboard(t *theme.Theme) *PluginDashboard {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Version/URL", Width: 30},
		{Title: "Status", Width: 15},
		{Title: "Enabled", Width: 10},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Style table with Dracula theme
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.Colors.Purple).
		BorderBottom(true).
		Bold(true).
		Foreground(t.Colors.Purple)
	s.Selected = lipgloss.NewStyle().
		Foreground(t.Colors.Background).
		Background(t.Colors.Purple).
		Bold(true)
	tbl.SetStyles(s)

	return &PluginDashboard{
		theme:      t,
		plugins:    []PluginInfo{},
		mcpServers: []MCPServerInfo{},
		table:      tbl,
		showingMCP: false,
	}
}

// Init initializes the dashboard
func (pd *PluginDashboard) Init() tea.Cmd {
	return pd.loadPluginData
}

// Update handles messages
func (pd *PluginDashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		pd.width = msg.Width
		pd.height = msg.Height
		pd.table.SetHeight(msg.Height - 8)
		return pd, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return pd, tea.Quit

		case "tab":
			// Toggle between plugins and MCP servers
			pd.showingMCP = !pd.showingMCP
			pd.updateTable()
			return pd, nil

		case "r":
			// Refresh data
			return pd, pd.loadPluginData
		}

	case pluginDataMsg:
		pd.plugins = msg.plugins
		pd.mcpServers = msg.mcpServers
		pd.updateTable()
		return pd, nil
	}

	var cmd tea.Cmd
	pd.table, cmd = pd.table.Update(msg)
	return pd, cmd
}

// View renders the dashboard
func (pd *PluginDashboard) View() string {
	if pd.width == 0 {
		return "Loading..."
	}

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(pd.theme.Colors.Pink).
		Bold(true).
		Padding(1, 2)

	var title string
	if pd.showingMCP {
		title = "MCP Server Dashboard"
	} else {
		title = "Plugin Dashboard"
	}

	titleView := titleStyle.Render(title)

	// Table
	tableView := pd.table.View()

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(pd.theme.Colors.Comment).
		Padding(1, 2)

	help := helpStyle.Render(
		"[Tab] Toggle View • [↑/↓] Navigate • [r] Refresh • [q] Quit",
	)

	// Stats
	statsStyle := lipgloss.NewStyle().
		Foreground(pd.theme.Colors.Cyan).
		Padding(0, 2)

	var stats string
	if pd.showingMCP {
		connected := 0
		for _, mcp := range pd.mcpServers {
			if mcp.Connected {
				connected++
			}
		}
		stats = fmt.Sprintf("Total: %d | Connected: %d", len(pd.mcpServers), connected)
	} else {
		enabled := 0
		for _, plugin := range pd.plugins {
			if plugin.Enabled {
				enabled++
			}
		}
		stats = fmt.Sprintf("Total: %d | Enabled: %d", len(pd.plugins), enabled)
	}
	statsView := statsStyle.Render(stats)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleView,
		statsView,
		tableView,
		help,
	)
}

// updateTable updates the table rows based on current view
func (pd *PluginDashboard) updateTable() {
	var rows []table.Row

	if pd.showingMCP {
		// Show MCP servers
		for _, mcp := range pd.mcpServers {
			connected := "❌"
			if mcp.Connected {
				connected = "✅"
			}
			rows = append(rows, table.Row{
				mcp.Name,
				mcp.URL,
				mcp.Status,
				connected,
			})
		}
	} else {
		// Show plugins
		for _, plugin := range pd.plugins {
			enabled := "❌"
			if plugin.Enabled {
				enabled = "✅"
			}
			rows = append(rows, table.Row{
				plugin.Name,
				plugin.Version,
				plugin.Status,
				enabled,
			})
		}
	}

	pd.table.SetRows(rows)
}

// Messages

type pluginDataMsg struct {
	plugins    []PluginInfo
	mcpServers []MCPServerInfo
}

// Commands

func (pd *PluginDashboard) loadPluginData() tea.Msg {
	// Mock data for now - in a real implementation, this would query actual plugin system
	plugins := []PluginInfo{
		{
			Name:    "superpowers",
			Version: "1.0.0",
			Enabled: true,
			Status:  "Active",
		},
		{
			Name:    "document-skills",
			Version: "1.2.0",
			Enabled: true,
			Status:  "Active",
		},
		{
			Name:    "elements-of-style",
			Version: "0.9.0",
			Enabled: false,
			Status:  "Disabled",
		},
	}

	mcpServers := []MCPServerInfo{
		{
			Name:      "private-journal",
			URL:       "local://journal",
			Connected: true,
			Status:    "Connected",
		},
		{
			Name:      "chronicle",
			URL:       "local://chronicle",
			Connected: true,
			Status:    "Connected",
		},
		{
			Name:      "playwright",
			URL:       "local://playwright",
			Connected: false,
			Status:    "Disconnected",
		},
	}

	return pluginDataMsg{
		plugins:    plugins,
		mcpServers: mcpServers,
	}
}

// GetPlugins returns the current plugin list
func (pd *PluginDashboard) GetPlugins() []PluginInfo {
	return pd.plugins
}

// GetMCPServers returns the current MCP server list
func (pd *PluginDashboard) GetMCPServers() []MCPServerInfo {
	return pd.mcpServers
}

// RenderCompact renders a compact summary for embedding in other views
func (pd *PluginDashboard) RenderCompact(width int) string {
	if len(pd.plugins) == 0 && len(pd.mcpServers) == 0 {
		return ""
	}

	var lines []string

	// Plugin summary
	enabled := 0
	for _, p := range pd.plugins {
		if p.Enabled {
			enabled++
		}
	}
	pluginLine := fmt.Sprintf("Plugins: %d/%d enabled", enabled, len(pd.plugins))
	lines = append(lines, pluginLine)

	// MCP summary
	connected := 0
	for _, m := range pd.mcpServers {
		if m.Connected {
			connected++
		}
	}
	mcpLine := fmt.Sprintf("MCP Servers: %d/%d connected", connected, len(pd.mcpServers))
	lines = append(lines, mcpLine)

	style := lipgloss.NewStyle().
		Foreground(pd.theme.Colors.Comment).
		Width(width)

	return style.Render(strings.Join(lines, " | "))
}
