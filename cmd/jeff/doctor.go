package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/harper/jeff/internal/core"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check installation health",
	Long:  "Verify that Jeff is correctly installed and configured",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(_ *cobra.Command, _ []string) error {
	fmt.Println("Jeff Health Check")
	fmt.Println("==================")
	fmt.Println()

	checks := []check{
		checkHomeDirectory,
		checkConfigFile,
		checkAPIKey,
	}

	allPassed := true
	for _, check := range checks {
		if !check() {
			allPassed = false
		}
		fmt.Println()
	}

	if allPassed {
		fmt.Println("✓ All checks passed")
	} else {
		fmt.Println("⚠ Some checks failed")
	}

	return nil
}

type check func() bool

func checkHomeDirectory() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		printCheck("Home directory", false, err.Error())
		return false
	}

	jeffDir := filepath.Join(home, ".jeff")
	if _, err := os.Stat(jeffDir); os.IsNotExist(err) {
		printCheck(".jeff directory", false, "not found")
		fmt.Printf("  Run: mkdir -p %s\n", jeffDir)
		return false
	}

	printCheck(".jeff directory", true, jeffDir)
	return true
}

func checkConfigFile() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		printCheck("Config file", false, "cannot get home dir")
		return false
	}

	configPath := filepath.Join(home, ".jeff", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		printCheck("Config file", false, "not found")
		fmt.Println("  Run: jeff setup-token <your-api-key>")
		return false
	}

	printCheck("Config file", true, configPath)
	return true
}

func checkAPIKey() bool {
	cfg, err := core.LoadConfig()
	if err != nil {
		printCheck("API key", false, err.Error())
		return false
	}

	if _, err := cfg.GetAPIKey(); err != nil {
		printCheck("API key", false, "not configured")
		fmt.Println("  Run: jeff setup-token <your-api-key>")
		fmt.Println("  Or set: export PAGEN_API_KEY=<your-key>")
		return false
	}

	printCheck("API key", true, "configured")
	return true
}

func printCheck(name string, passed bool, detail string) {
	symbol := "✓"
	if !passed {
		symbol = "✗"
	}
	fmt.Printf("%s %s: %s\n", symbol, name, detail)
}
