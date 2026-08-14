// Package commands implements the planetary CLI subcommands.
package commands

import (
	"fmt"
	"os"

	"github.com/fuegoio/planetary/go/cli/internal/client"
	"github.com/fuegoio/planetary/go/sdk/planetary"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "planetary",
	Short: "Planetary RSS reader CLI",
	Long:  "Planetary is a self-hosted RSS reader. This CLI lets you manage feeds, entries, categories, and API tokens from the terminal.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// mustClient loads the config and creates a client. It exits on error.
func mustClient() (*client.Config, *planetary.ClientWithResponses) {
	cfg, err := client.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		os.Exit(1)
	}

	c, err := client.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: create client: %v\n", err)
		os.Exit(1)
	}

	return cfg, c
}
