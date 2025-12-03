package theme_test

import (
	"fmt"

	"github.com/harper/hex/internal/ui/theme"
)

// ExampleDraculaTheme demonstrates basic usage of the Dracula theme
func ExampleDraculaTheme() {
	// Create a new Dracula theme instance
	t := theme.DraculaTheme()

	// Use the theme to style text
	title := t.Title.Render("Welcome to Hex")
	subtitle := t.Subtitle.Render("A beautiful CLI interface")
	body := t.Body.Render("This is some body text")

	fmt.Println(title)
	fmt.Println(subtitle)
	fmt.Println(body)
}

// ExampleDraculaTheme_statusIndicators demonstrates status styling
func ExampleDraculaTheme_statusIndicators() {
	t := theme.DraculaTheme()

	// Different status indicators
	success := t.Success.Render("✓ Operation successful")
	errorMsg := t.Error.Render("✗ Operation failed")
	warning := t.Warning.Render("⚠ Warning: Check this")
	info := t.Info.Render("ℹ Information")

	fmt.Println(success)
	fmt.Println(errorMsg)
	fmt.Println(warning)
	fmt.Println(info)
}

// ExampleDraculaTheme_toolStates demonstrates tool execution states
func ExampleDraculaTheme_toolStates() {
	t := theme.DraculaTheme()

	// Tool execution states
	approval := t.ToolApproval.Render("Approve tool execution?")
	executing := t.ToolExecuting.Render("⏳ Executing bash command...")
	success := t.ToolSuccess.Render("✓ Tool executed successfully")
	failed := t.ToolError.Render("✗ Tool execution failed")

	fmt.Println(approval)
	fmt.Println(executing)
	fmt.Println(success)
	fmt.Println(failed)
}

// ExampleDraculaTheme_lists demonstrates list item styling
func ExampleDraculaTheme_lists() {
	t := theme.DraculaTheme()

	// List items with different states
	normalItem := t.ListItem.Render("Regular list item")
	selectedItem := t.ListItemSelected.Render("Selected item")
	activeItem := t.ListItemActive.Render("Active item")

	fmt.Println(normalItem)
	fmt.Println(selectedItem)
	fmt.Println(activeItem)
}

// ExampleDraculaTheme_code demonstrates code styling
func ExampleDraculaTheme_code() {
	t := theme.DraculaTheme()

	// Code elements
	inlineCode := t.Code.Render("inline_code()")
	keyword := t.Keyword.Render("func")
	str := t.String.Render(`"hello world"`)
	num := t.Number.Render("42")

	fmt.Printf("%s %s = %s + %s\n", keyword, inlineCode, str, num)
}

// ExampleDraculaTheme_directColorAccess demonstrates using colors directly
func ExampleDraculaTheme_directColorAccess() {
	t := theme.DraculaTheme()

	// Access colors directly for custom styling
	fmt.Printf("Purple: %s\n", t.Colors.Purple)
	fmt.Printf("Pink: %s\n", t.Colors.Pink)
	fmt.Printf("Cyan: %s\n", t.Colors.Cyan)
	fmt.Printf("Green: %s\n", t.Colors.Green)
}
