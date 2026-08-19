package commands

import (
	"context"
	"fmt"

	"github.com/fuegoio/earthed/go/cli/internal/client"
	"github.com/spf13/cobra"
)

var (
	loginNoBrowser bool
	loginBaseURL   string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in via the browser (device flow)",
	Long: `Log in by opening a browser to a confirmation page where you approve
the CLI's access. The CLI polls the API until you approve, then stores the
resulting token in the config file.

If --no-browser is set, or a browser can't be opened, the confirmation URL
and code are printed so you can open them on any device where you're signed
in.`,
	Run: func(cmd *cobra.Command, _ []string) {
		cfg, err := client.LoadConfig()
		if err != nil {
			client.ExitOnError(fmt.Errorf("load config: %w", err))
		}
		if loginBaseURL != "" {
			cfg.BaseURL = loginBaseURL
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		result, err := client.Login(ctx, cfg.BaseURL, !loginNoBrowser, cmd.OutOrStdout())
		if err != nil {
			client.ExitOnError(err)
		}

		cfg.Token = result.Token
		if err := client.SaveConfig(cfg); err != nil {
			client.ExitOnError(fmt.Errorf("save config: %w", err))
		}
		fmt.Printf("\nToken saved to config. You're logged in.\n")
	},
}

func init() {
	loginCmd.Flags().BoolVar(&loginNoBrowser, "no-browser", false, "Print the URL instead of opening a browser")
	loginCmd.Flags().StringVar(&loginBaseURL, "url", "", "API base URL (overrides config)")
	rootCmd.AddCommand(loginCmd)
}
