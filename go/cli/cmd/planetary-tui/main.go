// Package main is the planetary-tui entry point.
package main

import (
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
