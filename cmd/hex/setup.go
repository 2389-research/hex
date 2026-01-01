// ABOUTME: Manual setup subcommand to re-run the configuration wizard
// ABOUTME: Allows users to reconfigure their LLM provider and API key at any time

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run the setup wizard to configure hex",
	Long: `Run the interactive setup wizard to configure your LLM provider and API key.

This wizard will help you:
  - Choose an AI provider (Anthropic, OpenAI, Gemini, OpenRouter, Ollama)
  - Configure your API key
  - Save your configuration to ~/.hex/config.toml

You can run this command at any time to reconfigure hex.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := RunWizard(); err != nil {
			if err == ErrSetupCancelled {
				fmt.Println("Setup cancelled.")
				return nil
			}
			return err
		}
		fmt.Println("Setup complete! Run 'hex' to start chatting.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
