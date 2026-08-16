// Package main is the planetary-tui entry point.
package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fuegoio/planetary/go/cli/internal/client"
	"github.com/fuegoio/planetary/go/cli/internal/tui"
)

func main() {
	cfg, err := client.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		os.Exit(1)
	}

	// If no token is configured, run the device-flow login before entering
	// the alt-screen TUI so the prompts render normally on the terminal.
	// Pass --login to force a re-login even when a token is present.
	forceLogin := len(os.Args) > 1 && os.Args[1] == "--login"
	if cfg.Token == "" || forceLogin {
		result, err := client.Login(context.Background(), cfg.BaseURL, true, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: login: %v\n", err)
			os.Exit(1)
		}
		cfg.Token = result.Token
		if err := client.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "error: save config: %v\n", err)
			os.Exit(1)
		}
	}

	c, err := client.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: create client: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.NewModel(c), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
