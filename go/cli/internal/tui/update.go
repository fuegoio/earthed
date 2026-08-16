package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Update handles messages and keyboard input.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case loadFeedsMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.feeds = msg.feeds
		m.err = ""
		m.rebuildSidebar()
		// load entries for whatever sidebar item is currently selected
		if len(m.items) > 0 && m.sidebarCursor < len(m.items) {
			m.loading = true
			return m, loadEntriesByParams(m.client, entriesParamsForItem(m.items[m.sidebarCursor]))
		}
		return m, nil

	case loadFoldersMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.folders = msg.folders
		m.err = ""
		m.rebuildSidebar()
		return m, nil

	case loadEntriesMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.entries = msg.entries
		m.entriesCursor = 0
		m.showEntry = false
		m.err = ""
		return m, nil

	case markReadMsg:
		if msg.err == nil {
			// update local state so the unread dot disappears immediately
			for i, e := range m.entries {
				if e.Id == msg.entryID {
					m.entries[i].Status = "read"
				}
			}
			if m.current.Id == msg.entryID {
				m.current.Status = "read"
			}
		}
		return m, nil

	case toggleStarMsg:
		if msg.err == nil {
			for i, e := range m.entries {
				if e.Id == msg.entryID {
					m.entries[i].Starred = msg.starred
				}
			}
			if m.current.Id == msg.entryID {
				m.current.Starred = msg.starred
			}
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey dispatches key events to the focused panel.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// global keys
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if !m.showEntry {
			return m, tea.Quit
		}
	case "r":
		m.loading = true
		return m, tea.Batch(loadFeeds(m.client), loadFolders(m.client))
	}

	if m.focus == focusSidebar {
		return m.handleSidebarKey(msg)
	}
	return m.handleEntriesKey(msg)
}

func (m Model) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.sidebarCursor > 0 {
			m.sidebarCursor--
		}
	case "down", "j":
		if m.sidebarCursor < len(m.items)-1 {
			m.sidebarCursor++
		}
	case "g":
		m.sidebarCursor = 0
	case "G":
		m.sidebarCursor = max(0, len(m.items)-1)
	case "enter", "l", "right":
		if len(m.items) == 0 {
			return m, nil
		}
		item := m.items[m.sidebarCursor]
		m.focus = focusEntries
		m.loading = true
		m.entries = nil
		m.showEntry = false
		return m, loadEntriesByParams(m.client, entriesParamsForItem(item))
	}
	return m, nil
}

func (m Model) handleEntriesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showEntry {
		return m.handleEntryDetailKey(msg)
	}

	switch msg.String() {
	case "up", "k":
		if m.entriesCursor > 0 {
			m.entriesCursor--
		}
	case "down", "j":
		if m.entriesCursor < len(m.entries)-1 {
			m.entriesCursor++
		}
	case "g":
		m.entriesCursor = 0
	case "G":
		m.entriesCursor = max(0, len(m.entries)-1)
	case "enter", "l", "right":
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.entriesCursor]
		m.current = entry
		// mark as read and open in browser
		var cmds []tea.Cmd
		if entry.Status != "read" {
			cmds = append(cmds, markRead(m.client, entry.Id))
		}
		cmds = append(cmds, openURL(entry.Url))
		m.showEntry = true
		return m, tea.Batch(cmds...)
	case "s":
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.entriesCursor]
		newStarred := !entry.Starred
		return m, toggleStar(m.client, entry.Id, newStarred)
	case "esc", "h", "left":
		m.focus = focusSidebar
		m.showEntry = false
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleEntryDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "h", "left", "backspace", "q":
		m.showEntry = false
	case "s":
		newStarred := !m.current.Starred
		return m, toggleStar(m.client, m.current.Id, newStarred)
	case "o":
		return m, openURL(m.current.Url)
	}
	return m, nil
}

// ---- View ------------------------------------------------------------------

// View renders the full TUI layout.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	if m.err != "" {
		return errStyle.Render("error: " + m.err + "\n\nr refresh · q quit")
	}

	sidebar := m.renderSidebar()
	main := m.renderMain()

	layout := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
	return layout
}

// renderSidebar renders the left navigation panel.
func (m Model) renderSidebar() string {
	innerWidth := sidebarWidth - 1 // subtract the right border

	var sb strings.Builder

	// header
	title := padRight(truncate("  Planetary", innerWidth), innerWidth)
	sb.WriteString(headerStyle.Width(innerWidth).Render(title))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", innerWidth) + "\n")

	if m.loading && len(m.items) == 0 {
		sb.WriteString(dimStyle.Render("  Loading…"))
		sb.WriteString("\n")
	}

	for i, item := range m.items {
		isFocused := m.focus == focusSidebar && i == m.sidebarCursor
		isActive := m.focus == focusEntries && i == m.sidebarCursor

		indent := strings.Repeat("  ", item.depth)
		var icon string
		switch item.kind {
		case sidebarAll:
			icon = " ≡ "
		case sidebarUnread:
			icon = " ○ "
		case sidebarStarred:
			icon = " ★ "
		case sidebarFolder:
			icon = " ▸ "
		case sidebarFeed:
			icon = "   "
		}

		label := truncate(item.label, innerWidth-len(indent)-len(icon))
		line := padRight(indent+icon+label, innerWidth)

		switch {
		case isFocused:
			sb.WriteString(selectedFocusStyle.Width(innerWidth).Render(line))
		case isActive:
			sb.WriteString(selectedStyle.Width(innerWidth).Render(line))
		case item.kind == sidebarFolder:
			sb.WriteString(folderStyle.Width(innerWidth).Render(line))
		default:
			sb.WriteString(normalStyle.Width(innerWidth).Render(line))
		}
		sb.WriteString("\n")
	}

	// fill remaining vertical space so the border reaches the bottom
	usedLines := 2 + len(m.items) // header + sep + items
	if m.loading && len(m.items) == 0 {
		usedLines++
	}
	for i := usedLines; i < m.height-1; i++ {
		sb.WriteString(strings.Repeat(" ", innerWidth) + "\n")
	}

	// help line at bottom
	help := padRight("  r reload · q quit", innerWidth)
	sb.WriteString(mutedStyle.Width(innerWidth).Render(help))

	content := sb.String()
	return sidebarBorderStyle.Height(m.height).Render(content)
}

// renderMain renders the right content panel.
func (m Model) renderMain() string {
	mainWidth := m.width - sidebarWidth - 1 // -1 for the sidebar border
	if mainWidth < 10 {
		mainWidth = 10
	}

	if m.loading {
		return lipgloss.NewStyle().Width(mainWidth).Padding(1, 2).Render(dimStyle.Render("Loading…"))
	}

	if m.showEntry {
		return m.renderEntryDetail(mainWidth)
	}

	return m.renderEntryList(mainWidth)
}

func (m Model) renderEntryList(width int) string {
	var sb strings.Builder

	// section title from current sidebar item
	title := "Entries"
	if len(m.items) > 0 && m.sidebarCursor < len(m.items) {
		item := m.items[m.sidebarCursor]
		switch item.kind {
		case sidebarAll:
			title = "All Entries"
		case sidebarUnread:
			title = "Unread"
		case sidebarStarred:
			title = "Starred"
		default:
			title = item.label
		}
	}

	sb.WriteString(headerStyle.Render(title) + "\n")
	sb.WriteString(strings.Repeat("─", width) + "\n")

	if len(m.entries) == 0 {
		sb.WriteString("\n")
		sb.WriteString(dimStyle.Padding(0, 2).Render("No entries here."))
		sb.WriteString("\n")
	} else {
		isFocus := m.focus == focusEntries
		for i, e := range m.entries {
			isSelected := i == m.entriesCursor

			// status dot: blue circle for unread
			dot := "  "
			if e.Status != "read" {
				dot = unreadDotStyle.Render("● ")
			}

			// star indicator
			star := "  "
			if e.Starred {
				star = starStyle.Render("★ ")
			}

			dateStr := e.PublishedAt.Format("01/02")
			dateCol := dimStyle.Render(dateStr) + " "

			// available width for title: width - dot(2) - star(2) - date(6) - padding(2)
			titleWidth := width - 2 - 2 - 6 - 2
			if titleWidth < 10 {
				titleWidth = 10
			}
			titleStr := truncate(e.Title, titleWidth)
			titleStr = padRight(titleStr, titleWidth)

			line := fmt.Sprintf(" %s%s%s%s", dot, star, titleStr, dateCol)

			if isSelected {
				if isFocus {
					sb.WriteString(selectedFocusStyle.Width(width).Render(line))
				} else {
					sb.WriteString(selectedStyle.Width(width).Render(line))
				}
			} else {
				if e.Status != "read" {
					sb.WriteString(normalStyle.Width(width).Render(line))
				} else {
					sb.WriteString(dimStyle.Width(width).Render(line))
				}
			}
			sb.WriteString("\n")
		}
	}

	// help bar
	var helpText string
	if m.focus == focusSidebar {
		helpText = "tab/l open · q quit"
	} else {
		helpText = "j/k navigate · enter open in browser · s star · h sidebar · q quit"
	}
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render(helpText))

	return sb.String()
}

func (m Model) renderEntryDetail(width int) string {
	e := m.current

	var sb strings.Builder

	// title
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Width(width-4).Padding(0, 2).Render(e.Title))
	sb.WriteString("\n\n")

	// meta row
	meta := ""
	if e.Author != nil && *e.Author != "" {
		meta += "by " + *e.Author + "  "
	}
	meta += e.PublishedAt.Format("2006-01-02 15:04")
	if e.Starred {
		meta += "  " + starStyle.Render("★ starred")
	} else {
		meta += "  " + mutedStyle.Render("☆ not starred")
	}
	sb.WriteString(dimStyle.Padding(0, 2).Render(meta))
	sb.WriteString("\n")

	// url
	urlStr := truncate(e.Url, width-4)
	sb.WriteString(dimStyle.Padding(0, 2).Render(urlStr))
	sb.WriteString("\n\n")

	sb.WriteString(strings.Repeat("─", width))
	sb.WriteString("\n\n")

	// body
	content := ""
	if e.Description != nil {
		content = *e.Description
	}
	// strip HTML tags naively
	content = stripTags(content)
	// wrap to width
	content = wordWrap(content, width-4)

	if len(content) > 4000 {
		content = content[:4000] + "\n\n[truncated]"
	}
	sb.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(content))
	sb.WriteString("\n\n")

	sb.WriteString(helpStyle.Render("esc/h back · o open in browser · s toggle star"))

	return sb.String()
}

// stripTags removes HTML tags from content for plain-text display.
func stripTags(s string) string {
	var out strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			out.WriteRune(r)
		}
	}
	// collapse multiple blank lines
	result := out.String()
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(result)
}

// wordWrap wraps long lines to width characters.
func wordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var out strings.Builder
	for _, para := range strings.Split(s, "\n") {
		if para == "" {
			out.WriteString("\n")
			continue
		}
		words := strings.Fields(para)
		col := 0
		for i, w := range words {
			wl := utf8.RuneCountInString(w)
			if col == 0 {
				out.WriteString(w)
				col = wl
			} else if col+1+wl > width {
				out.WriteString("\n")
				out.WriteString(w)
				col = wl
			} else {
				out.WriteString(" ")
				out.WriteString(w)
				col += 1 + wl
			}
			_ = i
		}
		out.WriteString("\n")
	}
	return strings.TrimRight(out.String(), "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
