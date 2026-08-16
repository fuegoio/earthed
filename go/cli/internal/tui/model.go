// Package tui implements the planetary TUI — an interactive feed browser.
package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fuegoio/planetary/go/sdk/planetary"
)

// focus tracks which panel has keyboard focus.
type focus int

const (
	focusSidebar focus = iota
	focusEntries
)

// sidebarItemKind identifies a navigation destination in the sidebar.
type sidebarItemKind int

const (
	sidebarAll sidebarItemKind = iota
	sidebarUnread
	sidebarStarred
	sidebarFeed
	sidebarFolder
)

type sidebarItem struct {
	kind     sidebarItemKind
	label    string
	feedID   int64
	folderID int64
	depth    int
}

// Model holds the TUI state.
type Model struct {
	client  *planetary.ClientWithResponses
	focus   focus
	width   int
	height  int
	loading bool
	err     string

	// sidebar data
	feeds         []planetary.Feed
	folders       []planetary.Folder
	items         []sidebarItem
	sidebarCursor int

	// entries panel
	entries       []planetary.Entry
	entriesCursor int
	showEntry     bool
	current       planetary.Entry
}

// ---- styles ----------------------------------------------------------------

const sidebarWidth = 28

var (
	sidebarBorderStyle = lipgloss.NewStyle().
				BorderRight(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("237"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("255")).
			Bold(true)

	selectedFocusStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("57")).
				Foreground(lipgloss.Color("255")).
				Bold(true)

	normalStyle = lipgloss.NewStyle()

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	starStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	unreadDotStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)

	folderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Bold(true)
)

// NewModel creates a new TUI model with the given client.
func NewModel(client *planetary.ClientWithResponses) Model {
	return Model{
		client: client,
		focus:  focusSidebar,
	}
}

// Init sends the initial command to load sidebar data.
func (m Model) Init() tea.Cmd {
	return tea.Batch(loadFeeds(m.client), loadFolders(m.client))
}

// ---- messages --------------------------------------------------------------

type loadFeedsMsg struct {
	feeds []planetary.Feed
	err   error
}

type loadFoldersMsg struct {
	folders []planetary.Folder
	err     error
}

type loadEntriesMsg struct {
	entries []planetary.Entry
	err     error
}

type markReadMsg struct {
	entryID int64
	err     error
}

type toggleStarMsg struct {
	entryID int64
	starred bool
	err     error
}

// ---- commands --------------------------------------------------------------

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

func loadFolders(client *planetary.ClientWithResponses) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.ListFoldersWithResponse(context.Background())
		if err != nil {
			return loadFoldersMsg{err: err}
		}
		if resp.JSON200 == nil {
			return loadFoldersMsg{err: fmt.Errorf("API error (status %d)", resp.StatusCode())}
		}
		return loadFoldersMsg{folders: *resp.JSON200}
	}
}

func loadEntriesByParams(client *planetary.ClientWithResponses, params *planetary.ListEntriesParams) tea.Cmd {
	return func() tea.Msg {
		if params.Limit == nil {
			params.Limit = ptr(int64(200))
		}
		resp, err := client.ListEntriesWithResponse(context.Background(), params)
		if err != nil {
			return loadEntriesMsg{err: err}
		}
		if resp.JSON200 == nil {
			return loadEntriesMsg{err: fmt.Errorf("API error (status %d)", resp.StatusCode())}
		}
		return loadEntriesMsg{entries: *resp.JSON200}
	}
}

func markRead(client *planetary.ClientWithResponses, entryID int64) tea.Cmd {
	return func() tea.Msg {
		ids := []int64{entryID}
		_, err := client.UpdateEntriesWithResponse(context.Background(), planetary.UpdateEntriesRequest{
			EntryIds: &ids,
			Status:   planetary.UpdateEntriesRequestStatusRead,
		})
		return markReadMsg{entryID: entryID, err: err}
	}
}

func toggleStar(client *planetary.ClientWithResponses, entryID int64, starred bool) tea.Cmd {
	return func() tea.Msg {
		_, err := client.ToggleEntryStarredWithResponse(context.Background(), entryID, planetary.ToggleEntryStarredRequest{
			Starred: starred,
		})
		return toggleStarMsg{entryID: entryID, starred: starred, err: err}
	}
}

func openURL(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		_ = cmd.Start()
		return nil
	}
}

// ---- sidebar helpers -------------------------------------------------------

// rebuildSidebar constructs the flat sidebar item list from feeds and folders.
func (m *Model) rebuildSidebar() {
	items := []sidebarItem{
		{kind: sidebarAll, label: "All"},
		{kind: sidebarUnread, label: "Unread"},
		{kind: sidebarStarred, label: "Starred"},
	}

	// feeds without a folder (depth 0)
	for _, f := range m.feeds {
		if f.FolderId == nil {
			items = append(items, sidebarItem{kind: sidebarFeed, label: f.Title, feedID: f.Id})
		}
	}

	// folders with their feeds nested (depth 1)
	for _, folder := range m.folders {
		items = append(items, sidebarItem{kind: sidebarFolder, label: folder.Title, folderID: folder.Id})
		for _, f := range m.feeds {
			if f.FolderId != nil && *f.FolderId == folder.Id {
				items = append(items, sidebarItem{kind: sidebarFeed, label: f.Title, feedID: f.Id, depth: 1})
			}
		}
	}

	m.items = items
}

// entriesParamsForItem returns API params matching the selected sidebar item.
func entriesParamsForItem(item sidebarItem) *planetary.ListEntriesParams {
	switch item.kind {
	case sidebarAll:
		return &planetary.ListEntriesParams{}
	case sidebarUnread:
		s := planetary.ListEntriesParamsStatusUnread
		return &planetary.ListEntriesParams{Status: &s}
	case sidebarStarred:
		return &planetary.ListEntriesParams{Starred: ptr(true)}
	case sidebarFeed:
		return &planetary.ListEntriesParams{FeedId: &item.feedID}
	case sidebarFolder:
		return &planetary.ListEntriesParams{FolderId: &item.folderID}
	}
	return &planetary.ListEntriesParams{}
}

// ---- utility ---------------------------------------------------------------

func ptr[T any](v T) *T { return &v }

// truncate cuts s to at most n visible runes, appending "…" when cut.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n-1]) + "…"
}

// padRight right-pads s to w visible runes.
func padRight(s string, w int) string {
	need := w - utf8.RuneCountInString(s)
	if need <= 0 {
		return s
	}
	return s + strings.Repeat(" ", need)
}
