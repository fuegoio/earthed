package commands

import (
	"fmt"
	"os"

	"github.com/fuegoio/earthed/go/cli/internal/client"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, _ []string) {
		cfg, err := client.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Base URL: %s\n", cfg.BaseURL)
		if cfg.Token != "" {
			fmt.Printf("Token: %s...%s\n", cfg.Token[:8], cfg.Token[len(cfg.Token)-4:])
		} else {
			fmt.Println("Token: (not set)")
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := client.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		switch args[0] {
		case "base_url":
			cfg.BaseURL = args[1]
		case "token":
			cfg.Token = args[1]
		default:
			fmt.Fprintf(os.Stderr, "error: unknown key %q (valid: base_url, token)\n", args[0])
			os.Exit(1)
		}
		if err := client.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set %s\n", args[0])
	},
}

func init() {
	configCmd.AddCommand(configShowCmd, configSetCmd)
	rootCmd.AddCommand(configCmd)
}
