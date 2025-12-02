// ABOUTME: Commands for managing and using session templates
// ABOUTME: Provides 'clem templates list' and --template flag integration

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/harper/pagent/internal/templates"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage session templates",
	Long: `Manage YAML-based session templates.

Templates are loaded from ~/.clem/templates/ and can define:
- System prompts
- Initial messages
- Enabled tools
- Model preferences
- Token limits

Use --template flag with root command to start a session with a template.`,
	RunE: runTemplatesList,
}

var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `List all available session templates from ~/.clem/templates/.`,
	RunE:  runTemplatesList,
}

func init() {
	templatesCmd.AddCommand(templatesListCmd)
	rootCmd.AddCommand(templatesCmd)
}

func runTemplatesList(_ *cobra.Command, _ []string) error {
	// Get templates directory
	dir, err := templates.GetTemplatesDir()
	if err != nil {
		return fmt.Errorf("get templates directory: %w", err)
	}

	// Load templates
	templateMap, err := templates.LoadTemplates(dir)
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	// Check if any templates exist
	if len(templateMap) == 0 {
		fmt.Printf("No templates found in %s\n\n", dir)
		fmt.Println("To create a template, add a YAML file to the templates directory.")
		fmt.Println("Example template structure:")
		fmt.Println()
		fmt.Println("  name: My Template")
		fmt.Println("  description: Template description")
		fmt.Println("  system_prompt: You are a helpful assistant")
		fmt.Println("  initial_messages:")
		fmt.Println("    - role: user")
		fmt.Println("      content: Hello")
		fmt.Println("  tools_enabled:")
		fmt.Println("    - read_file")
		fmt.Println("    - bash")
		fmt.Println()
		return nil
	}

	// Sort templates by name for consistent output
	names := make([]string, 0, len(templateMap))
	for name := range templateMap {
		names = append(names, name)
	}
	sort.Strings(names)

	// Display templates
	fmt.Printf("Available templates (%d found in %s):\n\n", len(templateMap), dir)

	for _, name := range names {
		template := templateMap[name]
		fmt.Printf("  %s\n", name)

		if template.Description != "" {
			fmt.Printf("    Description: %s\n", template.Description)
		}

		if template.Model != "" {
			fmt.Printf("    Model: %s\n", template.Model)
		}

		if len(template.ToolsEnabled) > 0 {
			fmt.Printf("    Tools: %s\n", strings.Join(template.ToolsEnabled, ", "))
		}

		if len(template.InitialMessages) > 0 {
			fmt.Printf("    Initial messages: %d\n", len(template.InitialMessages))
		}

		if template.SystemPrompt != "" {
			// Show first 60 chars of system prompt
			prompt := template.SystemPrompt
			if len(prompt) > 60 {
				prompt = prompt[:57] + "..."
			}
			fmt.Printf("    System prompt: %s\n", prompt)
		}

		fmt.Println()
	}

	fmt.Printf("Use with: clem --template <name>\n")

	return nil
}

// loadTemplateByName loads a template by name from the templates directory
func loadTemplateByName(name string) (*templates.Template, error) {
	dir, err := templates.GetTemplatesDir()
	if err != nil {
		return nil, fmt.Errorf("get templates directory: %w", err)
	}

	templateMap, err := templates.LoadTemplates(dir)
	if err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}

	template, exists := templateMap[name]
	if !exists {
		// List available templates in error message
		available := make([]string, 0, len(templateMap))
		for n := range templateMap {
			available = append(available, n)
		}
		sort.Strings(available)

		if len(available) > 0 {
			return nil, fmt.Errorf("template %q not found. Available templates: %s",
				name, strings.Join(available, ", "))
		}
		return nil, fmt.Errorf("template %q not found. No templates available in %s", name, dir)
	}

	return template, nil
}

// createExampleTemplates creates example template files if they don't exist
func createExampleTemplates() error {
	dir, err := templates.EnsureTemplatesDir()
	if err != nil {
		return err
	}

	// Code review template
	codeReviewPath := fmt.Sprintf("%s/code-review.yaml", dir)
	if _, err := os.Stat(codeReviewPath); os.IsNotExist(err) {
		codeReview := `name: code-review
description: Interactive code review session with best practices focus
system_prompt: |
  You are an expert code reviewer focused on:
  - Code quality and maintainability
  - Security vulnerabilities
  - Performance issues
  - Best practices and design patterns
  - Clear, constructive feedback

  Provide specific, actionable suggestions with examples.
initial_messages:
  - role: assistant
    content: |
      I'm ready to help review your code. I'll focus on:

      1. Security vulnerabilities
      2. Performance issues
      3. Code quality and maintainability
      4. Best practices

      Please share the code you'd like me to review using :read or paste it directly.
tools_enabled:
  - read_file
  - grep
  - glob
  - bash
model: claude-sonnet-4-5-20250929
max_tokens: 8192
`
		if err := os.WriteFile(codeReviewPath, []byte(codeReview), 0600); err != nil {
			return fmt.Errorf("write code-review template: %w", err)
		}
	}

	// Debug session template
	debugPath := fmt.Sprintf("%s/debug-session.yaml", dir)
	if _, err := os.Stat(debugPath); os.IsNotExist(err) {
		debugSession := `name: debug-session
description: Systematic debugging session with root cause analysis
system_prompt: |
  You are a systematic debugging expert. Your approach:

  1. Understand the problem - gather symptoms and context
  2. Form hypotheses - what could cause this?
  3. Test hypotheses - verify each possibility
  4. Identify root cause - find the actual issue
  5. Fix and verify - implement solution and test

  Ask clarifying questions and use tools to investigate thoroughly.
initial_messages:
  - role: assistant
    content: |
      Let's debug this systematically. I'll help you:

      1. Understand the symptoms and gather context
      2. Form and test hypotheses
      3. Identify the root cause
      4. Implement and verify the fix

      What issue are you experiencing? Please share:
      - Error messages or unexpected behavior
      - Steps to reproduce
      - Relevant code or logs
tools_enabled:
  - read_file
  - grep
  - glob
  - bash
  - edit_file
model: claude-sonnet-4-5-20250929
max_tokens: 8192
`
		if err := os.WriteFile(debugPath, []byte(debugSession), 0600); err != nil {
			return fmt.Errorf("write debug-session template: %w", err)
		}
	}

	// Refactoring template
	refactorPath := fmt.Sprintf("%s/refactor.yaml", dir)
	if _, err := os.Stat(refactorPath); os.IsNotExist(err) {
		refactor := `name: refactor
description: Safe refactoring session with test-driven approach
system_prompt: |
  You are a refactoring expert who prioritizes safety and maintainability.

  Your approach:
  1. Understand current code structure and behavior
  2. Ensure tests exist (or write them first)
  3. Make small, incremental changes
  4. Run tests after each change
  5. Verify behavior is preserved

  Never make breaking changes without explicit approval.
initial_messages:
  - role: assistant
    content: |
      I'm ready to help refactor your code safely. My approach:

      1. First, understand the current code
      2. Ensure we have tests (write them if needed)
      3. Make small, incremental improvements
      4. Test after each change

      What code would you like to refactor? Please share the file or describe what needs improvement.
tools_enabled:
  - read_file
  - edit_file
  - grep
  - glob
  - bash
model: claude-sonnet-4-5-20250929
max_tokens: 8192
`
		if err := os.WriteFile(refactorPath, []byte(refactor), 0600); err != nil {
			return fmt.Errorf("write refactor template: %w", err)
		}
	}

	return nil
}
