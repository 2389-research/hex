// ABOUTME: Export command to export conversations in multiple formats
// ABOUTME: Supports Markdown, JSON, and HTML output to stdout or file
package main

import (
	"fmt"
	"os"

	"github.com/harper/clem/internal/export"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportOutput string
)

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "markdown", "Export format (markdown, json, html)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")

	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export <conversation-id>",
	Short: "Export a conversation in various formats",
	Long: `Export a conversation to Markdown, JSON, or HTML format.

The export includes:
- Conversation metadata (title, model, timestamps)
- All messages with role, content, and timestamps
- Tool calls and metadata (if present)
- System prompt (if set)

Formats:
- markdown: Clean Markdown with YAML frontmatter (default)
- json: Complete JSON structure for programmatic use
- html: Styled HTML with syntax highlighting

Examples:
  # Export as Markdown to stdout
  clem export abc123

  # Export as JSON to file
  clem export abc123 --format json --output conversation.json

  # Export as HTML
  clem export abc123 --format html --output conversation.html`,
	Args: cobra.ExactArgs(1),
	RunE: runExport,
}

func runExport(cmd *cobra.Command, args []string) error {
	conversationID := args[0]

	// Validate format
	var format export.Format
	switch exportFormat {
	case "markdown", "md":
		format = export.FormatMarkdown
	case "json":
		format = export.FormatJSON
	case "html", "htm":
		format = export.FormatHTML
	default:
		return fmt.Errorf("unknown format: %s (valid formats: markdown, json, html)", exportFormat)
	}

	// Get database
	db, err := openDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("get database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Determine output writer
	var writer *os.File
	if exportOutput == "" {
		writer = os.Stdout
	} else {
		file, err := os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer func() { _ = file.Close() }()
		writer = file
	}

	// Export
	if err := export.Export(db, conversationID, format, writer); err != nil {
		return fmt.Errorf("export conversation: %w", err)
	}

	// Print success message to stderr if writing to file
	if exportOutput != "" {
		_, _ = fmt.Fprintf(os.Stderr, "Exported conversation %s to %s\n", conversationID, exportOutput)
	}

	return nil
}
