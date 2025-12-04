// ABOUTME: CLI commands for managing productivity providers
// ABOUTME: Handles provider add, use, list, test, and authentication

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/harper/jeff/internal/providers"
	"github.com/harper/jeff/internal/providers/gmail"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage productivity providers",
	Long: `Manage productivity providers (Gmail, Outlook, etc.)

Providers give Pagen access to your email, calendar, and tasks.
You must authenticate a provider before using it.`,
}

var providerAddCmd = &cobra.Command{
	Use:   "add [provider-name]",
	Short: "Add and authenticate a new provider",
	Long: `Add and authenticate a new provider.

Supported providers:
  - gmail: Google Gmail, Calendar, and Tasks

Example:
  pagen provider add gmail`,
	Args: cobra.ExactArgs(1),
	RunE: runProviderAdd,
}

var providerUseCmd = &cobra.Command{
	Use:   "use [provider-name]",
	Short: "Set the active provider",
	Long: `Set the active provider for productivity operations.

Only one provider can be active at a time (in v1).

Example:
  pagen provider use gmail`,
	Args: cobra.ExactArgs(1),
	RunE: runProviderUse,
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured providers",
	Long: `List all configured providers and their status.

Shows which provider is active and whether each is authenticated.

Example:
  pagen provider list`,
	Args: cobra.NoArgs,
	RunE: runProviderList,
}

var providerTestCmd = &cobra.Command{
	Use:   "test [provider-name]",
	Short: "Test provider authentication and connectivity",
	Long: `Test a provider's authentication and API connectivity.

Verifies that the provider is properly authenticated and can
communicate with its backend API.

Example:
  pagen provider test gmail`,
	Args: cobra.ExactArgs(1),
	RunE: runProviderTest,
}

var providerReauthCmd = &cobra.Command{
	Use:   "reauth [provider-name]",
	Short: "Re-authenticate a provider",
	Long: `Re-authenticate a provider (refresh OAuth tokens).

Use this if authentication expires or you need to grant new permissions.

Example:
  pagen provider reauth gmail`,
	Args: cobra.ExactArgs(1),
	RunE: runProviderReauth,
}

func init() {
	providerCmd.AddCommand(providerAddCmd)
	providerCmd.AddCommand(providerUseCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerTestCmd)
	providerCmd.AddCommand(providerReauthCmd)
	rootCmd.AddCommand(providerCmd)
}

func runProviderAdd(_ *cobra.Command, args []string) error {
	providerName := strings.ToLower(args[0])

	// Create provider registry
	registry := providers.NewRegistry()

	// Create provider instance based on name
	var provider providers.Provider
	switch providerName {
	case "gmail":
		provider = gmail.NewGmailProvider()
	default:
		return fmt.Errorf("unknown provider: %s\n\nSupported providers:\n  - gmail", providerName)
	}

	// Get configuration for this provider
	config, err := getProviderConfig(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider config: %w\n\nRun 'pagen init' to create configuration", err)
	}

	// Initialize provider
	if err := provider.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize provider: %w", err)
	}

	// Authenticate
	fmt.Printf("Setting up %s provider...\n\n", providerName)
	if err := provider.Authenticate(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Register provider
	if err := registry.Register(provider); err != nil {
		return fmt.Errorf("failed to register provider: %w", err)
	}

	// Save as active provider
	if err := saveActiveProvider(providerName); err != nil {
		return fmt.Errorf("failed to save active provider: %w", err)
	}

	fmt.Printf("✅ Provider '%s' added and set as active\n", providerName)
	fmt.Printf("\nYou can now use Pagen with %s!\n", providerName)
	fmt.Printf("Try: pagen \"what's on my calendar today?\"\n\n")

	return nil
}

func runProviderUse(_ *cobra.Command, args []string) error {
	providerName := strings.ToLower(args[0])

	// Load provider registry
	registry, err := loadProviderRegistry()
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	// Check if provider exists
	if !registry.HasProvider(providerName) {
		return fmt.Errorf("provider '%s' not found\n\nRun 'pagen provider list' to see available providers\nRun 'pagen provider add %s' to add it", providerName, providerName)
	}

	// Set as active
	if err := registry.SetActive(providerName); err != nil {
		return fmt.Errorf("failed to set active provider: %w", err)
	}

	// Save to config
	if err := saveActiveProvider(providerName); err != nil {
		return fmt.Errorf("failed to save active provider: %w", err)
	}

	fmt.Printf("✅ Active provider set to: %s\n", providerName)
	return nil
}

func runProviderList(_ *cobra.Command, _ []string) error {
	// Load provider registry
	registry, err := loadProviderRegistry()
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	providerInfos := registry.List()

	if len(providerInfos) == 0 {
		fmt.Println("No providers configured.")
		fmt.Println("\nTo add a provider:")
		fmt.Println("  pagen provider add gmail")
		return nil
	}

	fmt.Println("Configured Providers:")
	fmt.Println()

	for _, info := range providerInfos {
		activeMarker := " "
		if info.Active {
			activeMarker = "*"
		}

		authStatus := "not authenticated"
		if info.Authenticated {
			authStatus = "authenticated"
		}

		fmt.Printf("%s %s - %s\n", activeMarker, info.Name, authStatus)

		if len(info.SupportedTools) > 0 {
			fmt.Printf("  Tools: %d (%s, %s, ...)\n",
				len(info.SupportedTools),
				info.SupportedTools[0],
				info.SupportedTools[1])
		}
	}

	fmt.Println()
	fmt.Println("* = active provider")
	return nil
}

func runProviderTest(_ *cobra.Command, args []string) error {
	providerName := strings.ToLower(args[0])

	// Load provider registry
	registry, err := loadProviderRegistry()
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	// Get provider
	provider, err := registry.Get(providerName)
	if err != nil {
		return fmt.Errorf("provider not found: %w\n\nRun 'pagen provider add %s' to add it", err, providerName)
	}

	fmt.Printf("Testing provider: %s\n\n", providerName)

	// Check status
	status := provider.Status()
	if !status.Healthy {
		fmt.Printf("❌ Provider unhealthy: %s\n", status.Message)
		fmt.Printf("\nTry re-authenticating:\n")
		fmt.Printf("  pagen provider reauth %s\n", providerName)
		return fmt.Errorf("provider unhealthy")
	}

	fmt.Printf("✅ Provider healthy: %s\n", status.Message)

	// Show capabilities
	caps := provider.Capabilities()
	fmt.Printf("\nCapabilities:\n")
	fmt.Printf("  Max results: %d\n", caps.MaxResults)
	fmt.Printf("  Features: %s\n", strings.Join(caps.Features, ", "))

	fmt.Printf("\nSupported tools: %d\n", len(provider.SupportedTools()))

	fmt.Printf("\n✅ Provider test successful!\n")
	return nil
}

func runProviderReauth(_ *cobra.Command, args []string) error {
	providerName := strings.ToLower(args[0])

	// Load provider registry
	registry, err := loadProviderRegistry()
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	// Get provider
	provider, err := registry.Get(providerName)
	if err != nil {
		return fmt.Errorf("provider not found: %w\n\nRun 'pagen provider add %s' to add it", err, providerName)
	}

	fmt.Printf("Re-authenticating provider: %s\n\n", providerName)

	// Re-authenticate
	if err := provider.Authenticate(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("\n✅ Re-authentication successful!\n")
	return nil
}

// Helper functions

func getProviderConfig(providerName string) (map[string]string, error) {
	// TODO: Load from ~/.jeff/config.yaml
	// For now, return minimal config with defaults

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// For Gmail, we need client_id and client_secret
	// These should come from user's config file
	config := map[string]string{
		"token_file": fmt.Sprintf("%s/.jeff/tokens/%s.json", homeDir, providerName),
	}

	// Check for environment variables (for testing)
	if clientID := os.Getenv("JEFF_GMAIL_CLIENT_ID"); clientID != "" {
		config["client_id"] = clientID
	}
	if clientSecret := os.Getenv("JEFF_GMAIL_CLIENT_SECRET"); clientSecret != "" {
		config["client_secret"] = clientSecret
	}

	// Validate required fields
	if providerName == "gmail" {
		if _, ok := config["client_id"]; !ok {
			return nil, fmt.Errorf("missing client_id for Gmail\n\nSet via environment variable:\n  export JEFF_GMAIL_CLIENT_ID=your-client-id\n\nOr add to ~/.jeff/config.yaml")
		}
		if _, ok := config["client_secret"]; !ok {
			return nil, fmt.Errorf("missing client_secret for Gmail\n\nSet via environment variable:\n  export JEFF_GMAIL_CLIENT_SECRET=your-secret\n\nOr add to ~/.jeff/config.yaml")
		}
	}

	return config, nil
}

func loadProviderRegistry() (*providers.Registry, error) {
	// TODO: Load providers from saved state
	// For now, create fresh registry with available providers

	registry := providers.NewRegistry()

	// Try to load Gmail provider if it has saved tokens
	gmailProvider := gmail.NewGmailProvider()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	gmailConfig := map[string]string{
		"token_file":    fmt.Sprintf("%s/.jeff/tokens/gmail.json", homeDir),
		"client_id":     os.Getenv("JEFF_GMAIL_CLIENT_ID"),
		"client_secret": os.Getenv("JEFF_GMAIL_CLIENT_SECRET"),
	}

	// Only register if we have credentials
	if gmailConfig["client_id"] != "" && gmailConfig["client_secret"] != "" {
		if err := gmailProvider.Initialize(gmailConfig); err == nil {
			_ = registry.Register(gmailProvider)
		}
	}

	return registry, nil
}

func saveActiveProvider(_ string) error {
	// TODO: Save to ~/.jeff/config.yaml
	// For now, just create the directory structure
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	jeffDir := fmt.Sprintf("%s/.jeff", homeDir)
	if err := os.MkdirAll(jeffDir, 0700); err != nil {
		return err
	}

	tokensDir := fmt.Sprintf("%s/tokens", jeffDir)
	if err := os.MkdirAll(tokensDir, 0700); err != nil {
		return err
	}

	// TODO: Write actual config file with active provider

	return nil
}
