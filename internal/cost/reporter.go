// ABOUTME: reporter.go provides pretty-printing of cost summaries
// ABOUTME: Displays hierarchical cost breakdowns for multi-agent systems

package cost

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// PrintCostSummary prints a formatted cost summary for an agent tree
func PrintCostSummary(rootAgentID string) {
	tracker := Global()

	// Get root cost
	_, err := tracker.GetAgentCost(rootAgentID)
	if err != nil {
		// No costs recorded, skip
		return
	}

	// Get tree cost
	treeCost, err := tracker.GetTreeCost(rootAgentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calculating tree cost: %v\n", err)
		return
	}

	// Print header
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "                        COST SUMMARY                           ")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "")

	// Print tree
	printAgentTree(tracker, rootAgentID, "", true)

	// Print totals
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "───────────────────────────────────────────────────────────────")
	fmt.Fprintf(os.Stderr, "  Total Tree Cost: $%.4f\n", treeCost)
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "")
}

// printAgentTree recursively prints the agent cost tree
func printAgentTree(tracker *CostTracker, agentID string, prefix string, isLast bool) {
	cost, err := tracker.GetAgentCost(agentID)
	if err != nil {
		return
	}

	// Print tree branch
	branch := "├─"
	if isLast {
		branch = "└─"
	}
	if prefix == "" {
		branch = ""
	}

	// Format agent info
	status := "running"
	if cost.CompletedAt != nil {
		duration := cost.CompletedAt.Sub(cost.StartedAt)
		status = fmt.Sprintf("completed in %s", duration.Round(100*1000000)) // Round to 100ms
	}

	// Print agent line
	fmt.Fprintf(os.Stderr, "%s%s Agent: %s (%s)\n", prefix, branch, agentID, status)

	// Print cost details
	detailPrefix := prefix
	if prefix != "" {
		if isLast {
			detailPrefix += "  "
		} else {
			detailPrefix += "│ "
		}
	}

	fmt.Fprintf(os.Stderr, "%s  Model: %s\n", detailPrefix, cost.Model)
	fmt.Fprintf(os.Stderr, "%s  Tokens: %s in, %s out",
		detailPrefix,
		formatTokens(cost.InputTokens),
		formatTokens(cost.OutputTokens))

	if cost.CacheReads > 0 || cost.CacheWrites > 0 {
		fmt.Fprintf(os.Stderr, ", %s cache read, %s cache write",
			formatTokens(cost.CacheReads),
			formatTokens(cost.CacheWrites))
	}
	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintf(os.Stderr, "%s  Costs: $%.4f input, $%.4f output",
		detailPrefix, cost.InputCost, cost.OutputCost)
	if cost.CacheCost > 0 {
		fmt.Fprintf(os.Stderr, ", $%.4f cache", cost.CacheCost)
	}
	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintf(os.Stderr, "%s  Total: $%.4f\n", detailPrefix, cost.TotalCost)

	// Get and sort children
	children := tracker.GetChildren(agentID)
	sort.Slice(children, func(i, j int) bool {
		return children[i].StartedAt.Before(children[j].StartedAt)
	})

	// Recursively print children
	for i, child := range children {
		isLastChild := i == len(children)-1
		childPrefix := detailPrefix
		if !isLastChild {
			childPrefix += "│ "
		} else {
			childPrefix += "  "
		}
		fmt.Fprintln(os.Stderr, detailPrefix)
		printAgentTree(tracker, child.AgentID, childPrefix, isLastChild)
	}
}

// formatTokens formats token counts with thousands separators
func formatTokens(tokens int64) string {
	if tokens == 0 {
		return "0"
	}

	// Convert to string
	s := fmt.Sprintf("%d", tokens)

	// Add thousand separators
	var result []string
	for i, digit := range reverse(s) {
		if i > 0 && i%3 == 0 {
			result = append(result, ",")
		}
		result = append(result, string(digit))
	}

	return reverse(strings.Join(result, ""))
}

// reverse reverses a string
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
