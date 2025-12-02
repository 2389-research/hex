package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var setupCmd = &cobra.Command{
	Use:   "setup-token [token]",
	Short: "Configure API token",
	Long: `Configure your Anthropic API token.

Get your API key from: https://console.anthropic.com/

This command will save your API key to ~/.clem/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(_ *cobra.Command, args []string) error {
	var apiKey string

	if len(args) > 0 {
		apiKey = args[0]
	} else {
		fmt.Println("Usage: clem setup-token <your-api-key>")
		fmt.Println("\nGet your API key from: https://console.anthropic.com/")
		return nil
	}

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	// Create .clem directory
	clemDir := filepath.Join(home, ".clem")
	if err := os.MkdirAll(clemDir, 0750); err != nil {
		return fmt.Errorf("create .clem dir: %w", err)
	}

	// Write config
	configPath := filepath.Join(clemDir, "config.yaml")
	config := map[string]string{
		"api_key": apiKey,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("✓ API key configured successfully\n")
	fmt.Printf("  Saved to: %s\n", configPath)
	return nil
}
