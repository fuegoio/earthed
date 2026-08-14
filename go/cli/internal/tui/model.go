// Package tui implements the planetary TUI — an interactive feed browser.
package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fuegoio/planetary/go/sdk/planetary"
)

// view is the current screen state.
type view int

const (
	viewFeeds view = iota
	viewEntries
	viewEntry
)

// Model holds the TUI state.
type Model struct {
	client   *planetary.ClientWithResponses
	view     view
	feeds    []planetary.Feed
	entries  []planetary.Entry
	current  planetary.Entry
	cursor   int
	loading  bool
	err      string
	width    int
	height   int
	feedID   int64
}

// styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			Padding(0, 1)
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
	normalStyle = lipgloss.NewStyle()
	dimStyle    = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// NewModel creates a new TUI model with the given client.
func NewModel(client *planetary.ClientWithResponses) Model {
	return Model{
		client: client,
		view:   viewFeeds,
	}
}

// Init sends the initial command to load feeds.
func (m Model) Init() tea.Cmd {
	return loadFeeds(m.client)
}

// loadFeedsMsg carries the result of loading feeds.
type loadFeedsMsg struct {
	feeds []planetary.Feed
	err   error
}

// loadFeeds fetches all feeds from the API.
func loadFeeds(client *planetary.ClientWithResponses) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.ListFeedsWithResponse(context.Background())
		if err != nil {
			return loadFeedsMsg{err: err}
		}
		if resp.JSON200 == nil {
			return loadFeedsMsg{err: fmt.Errorf("API error (status %d)", resp.StatusCode())}
		}
		return loadFeedsMsg{feeds: *resp.JSON200}
	}
}

// loadEntriesMsg carries the result of loading entries for a feed.
type loadEntriesMsg struct {
	entries []planetary.Entry
	err     error
}

// loadEntries fetches entries for a given feed.
func loadEntries(client *planetary.ClientWithResponses, feedID int64) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.ListEntriesWithResponse(context.Background(), &planetary.ListEntriesParams{
			FeedId: &feedID,
			Limit:  ptr(int64(100)),
		})
		if err != nil {
			return loadEntriesMsg{err: err}
		}
		if resp.JSON200 == nil {
			return loadEntriesMsg{err: fmt.Errorf("API error (status %d)", resp.StatusCode())}
		}
		return loadEntriesMsg{entries: *resp.JSON200}
	}
}

func ptr[T any](v T) *T { return &v }
