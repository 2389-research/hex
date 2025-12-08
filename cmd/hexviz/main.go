// ABOUTME: main.go implements the hexviz CLI tool for visualizing multi-agent execution
// ABOUTME: Provides tree, timeline, and cost views with filtering and HTML export capabilities

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"
	"time"
)

// Event represents a single event from the event store
type Event struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agent_id"`
	ParentID  string                 `json:"parent_id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// EventType represents the type of event
type EventType string

// AgentNode represents a node in the agent hierarchy tree
type AgentNode struct {
	ID          string
	ParentID    string
	Task        string
	Model       string
	Cost        float64
	Duration    time.Duration
	Usage       UsageStats
	Children    []*AgentNode
	StartedAt   time.Time
	CompletedAt *time.Time
}

// UsageStats tracks token usage
type UsageStats struct {
	InputTokens  int64
	OutputTokens int64
	CacheReads   int64
	CacheWrites  int64
}

// ModelPricing contains pricing information per million tokens
type ModelPricing struct {
	InputTokenPrice  float64
	OutputTokenPrice float64
	CacheReadPrice   float64
	CacheWritePrice  float64
}

var (
	eventFile   = flag.String("events", "hex_events.jsonl", "event file to visualize")
	viewMode    = flag.String("view", "tree", "view mode: tree, timeline, or cost")
	agentFilter = flag.String("agent", "", "filter by agent ID (e.g., 'root' or 'root.1')")
	typeFilter  = flag.String("type", "", "filter by event type (e.g., 'ToolCall')")
	htmlOutput  = flag.String("html", "", "optional HTML output file for interactive view")
)

func main() {
	flag.Parse()

	// Load events
	events, err := loadEvents(*eventFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading events: %v\n", err)
		os.Exit(1)
	}

	// Apply filters
	if *agentFilter != "" {
		events = filterByAgent(events, *agentFilter)
	}
	if *typeFilter != "" {
		events = filterByType(events, *typeFilter)
	}

	// Render based on view mode
	var output string
	switch *viewMode {
	case "tree":
		output = renderTreeView(events)
	case "timeline":
		output = renderTimelineView(events)
	case "cost":
		output = renderCostView(events)
	default:
		fmt.Fprintf(os.Stderr, "Unknown view mode: %s (use tree, timeline, or cost)\n", *viewMode)
		os.Exit(1)
	}

	// Output to console
	fmt.Print(output)

	// Export to HTML if requested
	if *htmlOutput != "" {
		if err := exportHTML(events, *htmlOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting HTML: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "\nHTML exported to: %s\n", *htmlOutput)
	}
}

// loadEvents reads events from a JSON Lines file
func loadEvents(filename string) ([]Event, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var events []Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			// Skip malformed lines
			continue
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return events, nil
}

// filterByAgent filters events to only include the specified agent and its descendants
func filterByAgent(events []Event, agentID string) []Event {
	var filtered []Event
	for _, e := range events {
		if e.AgentID == agentID || strings.HasPrefix(e.AgentID, agentID+".") {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// filterByType filters events by type
func filterByType(events []Event, eventType string) []Event {
	var filtered []Event
	for _, e := range events {
		if string(e.Type) == eventType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// buildAgentTree constructs a hierarchical tree from flat events
func buildAgentTree(events []Event) *AgentNode {
	// Create map of agent ID to node
	nodes := make(map[string]*AgentNode)

	// First pass: create all nodes
	for _, e := range events {
		if e.Type == "AgentStart" {
			node := &AgentNode{
				ID:        e.AgentID,
				ParentID:  e.ParentID,
				Task:      extractTask(e.Data),
				Model:     extractModel(e.Data),
				Children:  make([]*AgentNode, 0),
				StartedAt: e.Timestamp,
			}
			nodes[e.AgentID] = node
		}
	}

	// Second pass: update nodes with stop event data
	for _, e := range events {
		if e.Type == "AgentStop" {
			if node, exists := nodes[e.AgentID]; exists {
				node.CompletedAt = &e.Timestamp
				node.Duration = e.Timestamp.Sub(node.StartedAt)
				node.Usage = extractUsage(e.Data)
				node.Cost = calculateAgentCost(node.Model, node.Usage)
			}
		}
	}

	// Third pass: build parent-child relationships
	var root *AgentNode
	for agentID, node := range nodes {
		if !strings.Contains(agentID, ".") {
			// This is the root
			root = node
		} else {
			// Find parent
			parentID := node.ParentID
			if parentID == "" {
				// Infer parent from ID
				lastDot := strings.LastIndex(agentID, ".")
				parentID = agentID[:lastDot]
			}
			if parent, exists := nodes[parentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return root
}

// renderTreeView creates a tree visualization
func renderTreeView(events []Event) string {
	tree := buildAgentTree(events)
	if tree == nil {
		return "No agents found\n"
	}

	var sb strings.Builder
	renderTreeNode(&sb, tree, "", true, true)

	// Add total tree cost
	totalCost := calculateTreeCost(tree)
	sb.WriteString(fmt.Sprintf("\nTotal Tree Cost: $%.4f\n", totalCost))

	return sb.String()
}

// renderTreeNode recursively renders a tree node
func renderTreeNode(sb *strings.Builder, node *AgentNode, prefix string, isRoot bool, isLast bool) {
	// Render current node
	if isRoot {
		fmt.Fprintf(sb, "%s\n", node.ID)
	} else {
		connector := "├─"
		if isLast {
			connector = "└─"
		}
		fmt.Fprintf(sb, "%s%s %s\n", prefix, connector, node.ID)
	}

	// Add details
	detailPrefix := prefix
	if !isRoot {
		if isLast {
			detailPrefix += "  "
		} else {
			detailPrefix += "│ "
		}
	}

	if node.Task != "" {
		fmt.Fprintf(sb, "%s├── Task: %q\n", detailPrefix, node.Task)
	}
	if node.Cost > 0 {
		fmt.Fprintf(sb, "%s├── Cost: $%.2f\n", detailPrefix, node.Cost)
	}
	if node.Duration > 0 {
		fmt.Fprintf(sb, "%s└── Duration: %s\n", detailPrefix, formatDuration(node.Duration))
	}

	// Render children
	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1
		childPrefix := detailPrefix
		if !isRoot {
			if isLast {
				childPrefix = prefix + "  "
			} else {
				childPrefix = prefix + "│ "
			}
		}
		sb.WriteString("\n")
		renderTreeNode(sb, child, childPrefix, false, childIsLast)
	}
}

// calculateTreeCost recursively calculates total cost for tree
func calculateTreeCost(node *AgentNode) float64 {
	total := node.Cost
	for _, child := range node.Children {
		total += calculateTreeCost(child)
	}
	return total
}

// renderTimelineView creates a chronological event listing
func renderTimelineView(events []Event) string {
	// Sort events by timestamp
	sorted := make([]Event, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	var sb strings.Builder
	for _, e := range sorted {
		timestamp := e.Timestamp.Format("2006-01-02 15:04:05")

		// Format event details
		detail := formatEventDetail(e)

		sb.WriteString(fmt.Sprintf("[%s] %s: %s", timestamp, e.AgentID, string(e.Type)))
		if detail != "" {
			sb.WriteString(fmt.Sprintf(" - %s", detail))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatEventDetail extracts relevant details from event data
func formatEventDetail(e Event) string {
	switch e.Type {
	case "AgentStart":
		if task := extractTask(e.Data); task != "" {
			return fmt.Sprintf("%q", task)
		}
	case "ToolCall":
		if toolName, ok := e.Data["tool_name"].(string); ok {
			if input, ok := e.Data["input"].(map[string]interface{}); ok {
				if path, ok := input["path"].(string); ok {
					return fmt.Sprintf("%s(path=%q)", toolName, path)
				}
				return toolName
			}
			return toolName
		}
	case "ToolResult":
		if toolName, ok := e.Data["tool_name"].(string); ok {
			success := "success"
			if s, ok := e.Data["success"].(bool); ok && !s {
				success = "failed"
			}
			return fmt.Sprintf("%s (%s)", toolName, success)
		}
	}
	return ""
}

// renderCostView creates a cost breakdown display
func renderCostView(events []Event) string {
	var sb strings.Builder

	sb.WriteString("Agent Cost Breakdown:\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Group events by agent
	agentEvents := make(map[string][]Event)
	for _, e := range events {
		agentEvents[e.AgentID] = append(agentEvents[e.AgentID], e)
	}

	// Sort agent IDs for consistent output
	agentIDs := make([]string, 0, len(agentEvents))
	for id := range agentEvents {
		agentIDs = append(agentIDs, id)
	}
	sort.Strings(agentIDs)

	totalCost := 0.0

	for _, agentID := range agentIDs {
		events := agentEvents[agentID]

		// Find start and stop events
		var model string
		var usage UsageStats

		for _, e := range events {
			switch e.Type {
			case "AgentStart":
				model = extractModel(e.Data)
			case "AgentStop":
				usage = extractUsage(e.Data)
			}
		}

		if usage.InputTokens == 0 && usage.OutputTokens == 0 {
			continue // Skip agents with no usage
		}

		pricing, _ := getPricing(model)

		inputCost := calculateCost(usage.InputTokens, pricing.InputTokenPrice)
		outputCost := calculateCost(usage.OutputTokens, pricing.OutputTokenPrice)
		cacheCost := calculateCost(usage.CacheReads, pricing.CacheReadPrice) +
			calculateCost(usage.CacheWrites, pricing.CacheWritePrice)
		agentTotal := inputCost + outputCost + cacheCost

		sb.WriteString(fmt.Sprintf("%s\n", agentID))
		sb.WriteString(fmt.Sprintf("  Model: %s\n", model))
		sb.WriteString(fmt.Sprintf("  Input:  %s tokens ($%.4f)\n", formatTokens(usage.InputTokens), inputCost))
		sb.WriteString(fmt.Sprintf("  Output: %s tokens ($%.4f)\n", formatTokens(usage.OutputTokens), outputCost))
		if usage.CacheReads > 0 {
			sb.WriteString(fmt.Sprintf("  Cache:  %s reads  ($%.4f)\n", formatTokens(usage.CacheReads), cacheCost))
		}
		if usage.CacheWrites > 0 {
			sb.WriteString(fmt.Sprintf("  Cache:  %s writes ($%.4f)\n", formatTokens(usage.CacheWrites), cacheCost))
		}
		sb.WriteString(fmt.Sprintf("  Total: $%.4f\n", agentTotal))
		sb.WriteString("\n")

		totalCost += agentTotal
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("Total Cost: $%.4f\n", totalCost))

	return sb.String()
}

// exportHTML generates an interactive HTML visualization
func exportHTML(events []Event, filename string) error {
	tree := buildAgentTree(events)

	data := struct {
		Tree      *AgentNode
		Events    []Event
		Timestamp string
	}{
		Tree:      tree,
		Events:    events,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("viz").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// Helper functions

func extractTask(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	if task, ok := data["task"].(string); ok {
		return task
	}
	return ""
}

func extractModel(data map[string]interface{}) string {
	if data == nil {
		return "unknown"
	}
	if model, ok := data["model"].(string); ok {
		return model
	}
	return "unknown"
}

func extractUsage(data map[string]interface{}) UsageStats {
	stats := UsageStats{}

	if data == nil {
		return stats
	}

	usage, ok := data["usage"].(map[string]interface{})
	if !ok {
		return stats
	}

	if v, ok := usage["input_tokens"].(float64); ok {
		stats.InputTokens = int64(v)
	}
	if v, ok := usage["output_tokens"].(float64); ok {
		stats.OutputTokens = int64(v)
	}
	if v, ok := usage["cache_read_tokens"].(float64); ok {
		stats.CacheReads = int64(v)
	}
	if v, ok := usage["cache_write_tokens"].(float64); ok {
		stats.CacheWrites = int64(v)
	}

	return stats
}

func getPricing(model string) (ModelPricing, error) {
	// Pricing per million tokens
	pricingMap := map[string]ModelPricing{
		"claude-sonnet-4-5-20250929": {
			InputTokenPrice:  3.0,
			OutputTokenPrice: 15.0,
			CacheReadPrice:   0.30,
			CacheWritePrice:  3.75,
		},
		"claude-sonnet-4-20250514": {
			InputTokenPrice:  3.0,
			OutputTokenPrice: 15.0,
			CacheReadPrice:   0.30,
			CacheWritePrice:  3.75,
		},
		"claude-opus-4-20250514": {
			InputTokenPrice:  15.0,
			OutputTokenPrice: 75.0,
			CacheReadPrice:   1.50,
			CacheWritePrice:  18.75,
		},
	}

	if pricing, ok := pricingMap[model]; ok {
		return pricing, nil
	}

	// Default to Sonnet pricing for unknown models
	return pricingMap["claude-sonnet-4-5-20250929"], nil
}

func calculateCost(tokens int64, pricePerMillion float64) float64 {
	return float64(tokens) * pricePerMillion / 1_000_000.0
}

func calculateAgentCost(model string, usage UsageStats) float64 {
	pricing, _ := getPricing(model)

	inputCost := calculateCost(usage.InputTokens, pricing.InputTokenPrice)
	outputCost := calculateCost(usage.OutputTokens, pricing.OutputTokenPrice)
	cacheCost := calculateCost(usage.CacheReads, pricing.CacheReadPrice) +
		calculateCost(usage.CacheWrites, pricing.CacheWritePrice)

	return inputCost + outputCost + cacheCost
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

func formatTokens(tokens int64) string {
	if tokens == 0 {
		return "0"
	}

	// Add thousand separators
	s := fmt.Sprintf("%d", tokens)
	n := len(s)
	if n <= 3 {
		return s
	}

	var result strings.Builder
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}

	return result.String()
}

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Hex Execution Visualization</title>
    <meta charset="utf-8">
    <style>
        body {
            font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
            background-color: #1e1e1e;
            color: #d4d4d4;
            padding: 20px;
            line-height: 1.6;
        }
        h1 {
            color: #4ec9b0;
            border-bottom: 2px solid #4ec9b0;
            padding-bottom: 10px;
        }
        .timestamp {
            color: #808080;
            font-size: 0.9em;
        }
        .agent-tree {
            margin: 20px 0;
        }
        .agent-node {
            margin: 10px 0;
            padding: 10px;
            background-color: #2d2d2d;
            border-left: 3px solid #4ec9b0;
            border-radius: 3px;
        }
        .agent-id {
            font-weight: bold;
            color: #4ec9b0;
        }
        .task {
            color: #ce9178;
            font-style: italic;
        }
        .cost {
            color: #b5cea8;
            font-weight: bold;
        }
        .duration {
            color: #9cdcfe;
        }
        .child {
            margin-left: 30px;
        }
        .total-cost {
            margin-top: 20px;
            padding: 15px;
            background-color: #2d2d2d;
            border: 2px solid #4ec9b0;
            border-radius: 5px;
            font-size: 1.2em;
            text-align: center;
        }
    </style>
</head>
<body>
    <h1>Hex Agent Execution Visualization</h1>
    <div class="timestamp">Generated: {{.Timestamp}}</div>

    <div class="agent-tree">
        {{if .Tree}}
            {{template "node" .Tree}}
        {{else}}
            <p>No agent tree found</p>
        {{end}}
    </div>

    {{if .Tree}}
    <div class="total-cost">
        Total Cost: <span class="cost">${{printf "%.4f" .Tree.Cost}}</span>
    </div>
    {{end}}
</body>
</html>

{{define "node"}}
<div class="agent-node">
    <div class="agent-id">{{.ID}}</div>
    {{if .Task}}
    <div class="task">"{{.Task}}"</div>
    {{end}}
    {{if gt .Cost 0.0}}
    <div class="cost">Cost: ${{printf "%.4f" .Cost}}</div>
    {{end}}
    {{if gt .Duration 0}}
    <div class="duration">Duration: {{.Duration}}</div>
    {{end}}
</div>
{{if .Children}}
<div class="child">
    {{range .Children}}
        {{template "node" .}}
    {{end}}
</div>
{{end}}
{{end}}
`
