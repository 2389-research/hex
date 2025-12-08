// ABOUTME: hexreplay is a CLI tool for replaying hex events from disk.
// ABOUTME: Shows chronological timeline of agent events for debugging and audit.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/2389-research/hex/internal/events"
)

func main() {
	eventFile := flag.String("events", "hex_events.jsonl", "event file to replay")
	agentFilter := flag.String("agent", "", "filter by agent ID (shows agent and descendants)")
	typeFilter := flag.String("type", "", "filter by event type")
	verbose := flag.Bool("v", false, "verbose output (show event data)")
	flag.Parse()

	file, err := os.Open(*eventFile)
	if err != nil {
		log.Fatalf("Failed to open event file: %v", err)
	}
	defer func() { _ = file.Close() }()

	fmt.Printf("Replaying events from: %s\n", *eventFile)
	if *agentFilter != "" {
		fmt.Printf("Filtering by agent: %s\n", *agentFilter)
	}
	if *typeFilter != "" {
		fmt.Printf("Filtering by type: %s\n", *typeFilter)
	}
	fmt.Println(strings.Repeat("-", 80))

	scanner := bufio.NewScanner(file)
	eventCount := 0

	for scanner.Scan() {
		var event events.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			log.Printf("Error parsing event: %v", err)
			continue
		}

		// Apply filters
		if *agentFilter != "" {
			if event.AgentID != *agentFilter && !strings.HasPrefix(event.AgentID, *agentFilter+".") {
				continue
			}
		}

		if *typeFilter != "" {
			if string(event.Type) != *typeFilter {
				continue
			}
		}

		// Display event
		timestamp := event.Timestamp.Format(time.RFC3339)
		fmt.Printf("[%s] %-20s | %s\n", timestamp, event.AgentID, event.Type)

		if *verbose && event.Data != nil {
			dataJSON, err := json.MarshalIndent(event.Data, "  ", "  ")
			if err != nil {
				fmt.Printf("  Data: %v\n", event.Data)
			} else {
				fmt.Printf("  Data: %s\n", string(dataJSON))
			}
		}

		if event.ParentID != "" && *verbose {
			fmt.Printf("  Parent: %s\n", event.ParentID)
		}

		eventCount++
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total events: %d\n", eventCount)
}
