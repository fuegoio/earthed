// Package commands implements the sunred CLI subcommands.
package commands

import (
	"fmt"
	"os"

	"github.com/fuegoio/sunred/go/cli/internal/client"
	"github.com/fuegoio/sunred/go/sdk/sunred"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sunred",
	Short: "Sunred RSS reader CLI",
	Long:  "Sunred is a self-hosted RSS reader. This CLI lets you manage feeds, entries, folders, and API tokens from the terminal.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// mustClient loads the config and creates a client. It exits on error.
func mustClient() (*client.Config, *sunred.ClientWithResponses) {
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
