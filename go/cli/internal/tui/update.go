package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and keyboard input.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case loadFeedsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.feeds = msg.feeds
		m.err = ""
		return m, nil

	case loadEntriesMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.entries = msg.entries
		m.err = ""
		m.cursor = 0
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey processes keyboard input based on the current view.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	switch m.view {
	case viewFeeds:
		return m.handleFeedsKey(msg)
	case viewEntries:
		return m.handleEntriesKey(msg)
	case viewEntry:
		return m.handleEntryKey(msg)
	}

	return m, nil
}

func (m Model) handleFeedsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.feeds)-1 {
			m.cursor++
		}
	case "enter", "l":
		if len(m.feeds) > 0 {
			m.feedID = m.feeds[m.cursor].Id
			m.view = viewEntries
			m.loading = true
			m.entries = nil
			return m, loadEntries(m.client, m.feedID)
		}
	case "r":
		m.loading = true
		return m, loadFeeds(m.client)
	}

	return m, nil
}

func (m Model) handleEntriesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
	case "enter", "l":
		if len(m.entries) > 0 {
			m.current = m.entries[m.cursor]
			m.view = viewEntry
		}
	case "esc", "h", "backspace":
		m.view = viewFeeds
		m.cursor = 0
		m.entries = nil
	}

	return m, nil
}

func (m Model) handleEntryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "h", "backspace", "q":
		m.view = viewEntries
	}

	return m, nil
}

// View renders the current screen.
func (m Model) View() string {
	if m.err != "" {
		return errStyle.Render("error: " + m.err)
	}

	if m.loading {
		return dimStyle.Render("Loading...")
	}

	switch m.view {
	case viewFeeds:
		return m.renderFeeds()
	case viewEntries:
		return m.renderEntries()
	case viewEntry:
		return m.renderEntry()
	}

	return ""
}

func (m Model) renderFeeds() string {
	s := titleStyle.Render("Planetary — Feeds") + "\n\n"

	if len(m.feeds) == 0 {
		s += dimStyle.Render("No feeds. Subscribe to one with: planetary feeds add <url>")
		return s
	}

	for i, f := range m.feeds {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "> "
			style = selectedStyle
		}

		star := ""
		if f.ParsingErrorCount > 0 {
			star = " !"
		}

		s += style.Render(fmt.Sprintf("%s%d. %s%s", cursor, f.Id, f.Title, star)) + "\n"
	}

	s += "\n" + dimStyle.Render("↑/↓ navigate · enter open · r refresh · q quit")
	return s
}

func (m Model) renderEntries() string {
	s := titleStyle.Render("Planetary — Entries") + "\n\n"

	if len(m.entries) == 0 {
		s += dimStyle.Render("No entries in this feed.")
		return s
	}

	for i, e := range m.entries {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "> "
			style = selectedStyle
		}

		star := ""
		if e.Starred {
			star = " *"
		}

		s += style.Render(fmt.Sprintf("%s%d. %s%s", cursor, e.Id, e.Title, star)) + "\n"
	}

	s += "\n" + dimStyle.Render("↑/↓ navigate · enter read · esc back · q quit")
	return s
}

func (m Model) renderEntry() string {
	e := m.current
	s := titleStyle.Render(e.Title) + "\n\n"

	if e.Author != nil && *e.Author != "" {
		s += dimStyle.Render("By " + *e.Author) + "\n"
	}
	s += dimStyle.Render(e.PublishedAt.Format("2006-01-02 15:04")) + "\n"
	s += dimStyle.Render(e.Url) + "\n\n"

	content := ""
	if e.Description != nil {
		content = *e.Description
	}
	if len(content) > 2000 {
		content = content[:2000] + "..."
	}
	s += content + "\n"

	s += "\n" + dimStyle.Render("esc back · q quit")
	return s
}
