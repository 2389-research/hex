// Package export provides conversation export functionality in multiple formats.
// ABOUTME: HTML exporter for conversations with syntax highlighting
// ABOUTME: Uses Chroma for code block syntax highlighting with color themes
package export

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/harper/pagent/internal/storage"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// HTMLExporter exports conversations as HTML with syntax highlighting
type HTMLExporter struct{}

// Export implements the Exporter interface for HTML format
func (e *HTMLExporter) Export(conv *storage.Conversation, messages []*storage.Message, w io.Writer) error {
	// Write HTML header
	if err := e.writeHeader(w, conv); err != nil {
		return err
	}

	// Write conversation metadata
	if err := e.writeMetadata(w, conv); err != nil {
		return err
	}

	// Write messages
	if err := e.writeMessages(w, messages); err != nil {
		return err
	}

	// Write HTML footer
	if err := e.writeFooter(w); err != nil {
		return err
	}

	return nil
}

// writeHeader writes the HTML header with CSS
func (e *HTMLExporter) writeHeader(w io.Writer, conv *storage.Conversation) error {
	css := e.generateCSS()
	header := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
%s
    </style>
</head>
<body>
    <div class="container">
`, html.EscapeString(conv.Title), css)

	_, err := w.Write([]byte(header))
	return err
}

// writeMetadata writes conversation metadata
func (e *HTMLExporter) writeMetadata(w io.Writer, conv *storage.Conversation) error {
	metadata := fmt.Sprintf(`        <div class="metadata">
            <h1>%s</h1>
            <div class="meta-item"><strong>ID:</strong> %s</div>
            <div class="meta-item"><strong>Model:</strong> %s</div>
            <div class="meta-item"><strong>Created:</strong> %s</div>
            <div class="meta-item"><strong>Updated:</strong> %s</div>
`, html.EscapeString(conv.Title),
		html.EscapeString(conv.ID),
		html.EscapeString(conv.Model),
		html.EscapeString(conv.CreatedAt.Format("2006-01-02 15:04:05")),
		html.EscapeString(conv.UpdatedAt.Format("2006-01-02 15:04:05")))

	if conv.SystemPrompt != "" {
		metadata += fmt.Sprintf(`            <div class="meta-item"><strong>System Prompt:</strong> <pre>%s</pre></div>
`, html.EscapeString(conv.SystemPrompt))
	}

	metadata += `        </div>
`
	_, err := w.Write([]byte(metadata))
	return err
}

// writeMessages writes all messages with syntax highlighting
func (e *HTMLExporter) writeMessages(w io.Writer, messages []*storage.Message) error {
	for _, msg := range messages {
		if err := e.writeMessage(w, msg); err != nil {
			return err
		}
	}
	return nil
}

// writeMessage writes a single message
func (e *HTMLExporter) writeMessage(w io.Writer, msg *storage.Message) error {
	roleClass := fmt.Sprintf("message message-%s", msg.Role)
	roleIcon := roleToEmoji(msg.Role)
	caser := cases.Title(language.English)

	msgHTML := fmt.Sprintf(`        <div class="%s">
            <div class="message-header">
                <span class="role-icon">%s</span>
                <span class="role-name">%s</span>
                <span class="timestamp">%s</span>
            </div>
            <div class="message-content">
`,
		roleClass,
		roleIcon,
		html.EscapeString(caser.String(msg.Role)),
		html.EscapeString(msg.CreatedAt.Format("2006-01-02 15:04:05")))

	if _, err := w.Write([]byte(msgHTML)); err != nil {
		return err
	}

	// Process content with code block highlighting
	processedContent := e.processContent(msg.Content)

	if _, err := w.Write([]byte(processedContent)); err != nil {
		return err
	}

	// Write tool calls if present
	if msg.ToolCalls != "" && msg.ToolCalls != "null" {
		toolCallsHTML := `                <div class="tool-calls">
                    <strong>Tool Calls:</strong>
`
		if _, err := w.Write([]byte(toolCallsHTML)); err != nil {
			return err
		}

		highlighted, err := e.highlightCode(msg.ToolCalls, "json")
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(highlighted)); err != nil {
			return err
		}

		if _, err := w.Write([]byte("                </div>\n")); err != nil {
			return err
		}
	}

	// Close message div
	if _, err := w.Write([]byte(`            </div>
        </div>
`)); err != nil {
		return err
	}

	return nil
}

// processContent processes message content, highlighting code blocks
func (e *HTMLExporter) processContent(content string) string {
	// Find code blocks with language specifiers
	codeBlockRegex := regexp.MustCompile("(?s)```(\\w+)?\\n(.*?)```")

	// Track positions of code blocks
	matches := codeBlockRegex.FindAllStringIndex(content, -1)

	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		// FIX: Escape text BEFORE this code block (non-code content)
		if match[0] > lastEnd {
			textBefore := content[lastEnd:match[0]]
			result.WriteString(html.EscapeString(textBefore))
		}

		// Process the code block
		codeBlock := content[match[0]:match[1]]
		parts := codeBlockRegex.FindStringSubmatch(codeBlock)
		if len(parts) >= 3 {
			lang := parts[1]
			code := parts[2]
			if lang == "" {
				lang = "text"
			}
			highlighted, err := e.highlightCode(code, lang)
			if err != nil {
				result.WriteString(fmt.Sprintf("<pre><code>%s</code></pre>", html.EscapeString(code)))
			} else {
				result.WriteString(highlighted)
			}
		}

		lastEnd = match[1]
	}

	// FIX: Escape any remaining text after last code block
	if lastEnd < len(content) {
		result.WriteString(html.EscapeString(content[lastEnd:]))
	}

	return result.String()
}

// highlightCode highlights code using Chroma
func (e *HTMLExporter) highlightCode(code, lang string) (string, error) {
	// Get lexer for language
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Get style
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Create HTML formatter
	formatter := chromahtml.New(
		chromahtml.WithClasses(false),
		chromahtml.Standalone(false),
		chromahtml.WithLineNumbers(false),
		chromahtml.TabWidth(4),
	)

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return "", err
	}

	// Format to string
	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// writeFooter writes the HTML footer
func (e *HTMLExporter) writeFooter(w io.Writer) error {
	footer := `    </div>
</body>
</html>
`
	_, err := w.Write([]byte(footer))
	return err
}

// generateCSS generates the stylesheet for the HTML export
func (e *HTMLExporter) generateCSS() string {
	return `        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .metadata {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 6px;
            margin-bottom: 30px;
            border-left: 4px solid #007bff;
        }
        .metadata h1 {
            margin-top: 0;
            color: #2c3e50;
        }
        .meta-item {
            margin: 8px 0;
            font-size: 14px;
        }
        .meta-item strong {
            color: #555;
        }
        .meta-item pre {
            background: white;
            padding: 10px;
            border-radius: 4px;
            overflow-x: auto;
        }
        .message {
            margin: 20px 0;
            padding: 20px;
            border-radius: 6px;
            border: 1px solid #e0e0e0;
        }
        .message-user {
            background: #e3f2fd;
            border-left: 4px solid #2196f3;
        }
        .message-assistant {
            background: #f3e5f5;
            border-left: 4px solid #9c27b0;
        }
        .message-system {
            background: #fff3e0;
            border-left: 4px solid #ff9800;
        }
        .message-header {
            display: flex;
            align-items: center;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 1px solid rgba(0,0,0,0.1);
        }
        .role-icon {
            font-size: 20px;
            margin-right: 8px;
        }
        .role-name {
            font-weight: bold;
            margin-right: 15px;
            text-transform: capitalize;
        }
        .timestamp {
            color: #666;
            font-size: 13px;
            margin-left: auto;
        }
        .message-content {
            line-height: 1.8;
        }
        .message-content pre {
            background: #1e1e1e;
            padding: 15px;
            border-radius: 4px;
            overflow-x: auto;
            margin: 10px 0;
        }
        .message-content code {
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
            font-size: 13px;
        }
        .tool-calls {
            margin-top: 15px;
            padding-top: 15px;
            border-top: 1px solid rgba(0,0,0,0.1);
        }
        .tool-calls strong {
            display: block;
            margin-bottom: 8px;
            color: #555;
        }`
}
