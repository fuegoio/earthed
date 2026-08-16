package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fuegoio/planetary/go/sdk/planetary"
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
		m.entriesOffset = 0
		m.err = ""
		return m, nil

	case markReadMsg:
		if msg.err == nil {
			for i, e := range m.entries {
				if e.Id == msg.entryID {
					m.entries[i].Status = string(msg.status)
				}
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
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey dispatches key events to the focused panel.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Search mode intercepts everything except ctrl+c.
	if m.searching {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.handleSearchKey(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		return m, tea.Quit
	case "/":
		m.searching = true
		m.searchQuery = ""
		m.focus = focusEntries
		return m, nil
	case "r":
		m.loading = true
		return m, tea.Batch(loadFeeds(m.client), loadFolders(m.client))
	}

	if m.focus == focusSidebar {
		return m.handleSidebarKey(msg)
	}
	return m.handleEntriesKey(msg)
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel: restore the previous sidebar selection.
		m.searching = false
		m.searchQuery = ""
		if len(m.items) > 0 {
			m.loading = true
			m.entries = nil
			return m, loadEntriesByParams(m.client, entriesParamsForItem(m.items[m.sidebarCursor]))
		}
		return m, nil
	case "enter":
		// Confirm: run the search.
		m.loading = true
		m.entries = nil
		m.entriesCursor = 0
		m.entriesOffset = 0
		if m.searchQuery == "" {
			// Empty query: exit search and reload current selection.
			m.searching = false
			if len(m.items) > 0 {
				return m, loadEntriesByParams(m.client, entriesParamsForItem(m.items[m.sidebarCursor]))
			}
			return m, nil
		}
		return m, searchEntries(m.client, m.searchQuery)
	case "backspace", "ctrl+h":
		runes := []rune(m.searchQuery)
		if len(runes) > 0 {
			m.searchQuery = string(runes[:len(runes)-1])
		}
		// Live search on delete.
		if m.searchQuery != "" {
			m.loading = true
			m.entries = nil
			m.entriesCursor = 0
			m.entriesOffset = 0
			return m, searchEntries(m.client, m.searchQuery)
		}
		return m, nil
	default:
		// Append printable characters.
		if msg.Type == tea.KeyRunes {
			m.searchQuery += string(msg.Runes)
			m.loading = true
			m.entries = nil
			m.entriesCursor = 0
			m.entriesOffset = 0
			return m, searchEntries(m.client, m.searchQuery)
		}
	}
	return m, nil
}

func (m Model) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	prev := m.sidebarCursor
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
		m.focus = focusEntries
		m.loading = true
		m.entries = nil
		return m, loadEntriesByParams(m.client, entriesParamsForItem(m.items[m.sidebarCursor]))
	}
	// If the cursor moved, refresh the entry panel for the new selection.
	if m.sidebarCursor != prev && len(m.items) > 0 {
		m.loading = true
		m.entries = nil
		return m, loadEntriesByParams(m.client, entriesParamsForItem(m.items[m.sidebarCursor]))
	}
	return m, nil
}

func (m Model) handleEntriesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "ctrl+d":
		pageSize := m.entriesPageSize()
		m.entriesCursor = min(m.entriesCursor+pageSize/2, len(m.entries)-1)
	case "ctrl+u":
		pageSize := m.entriesPageSize()
		m.entriesCursor = max(m.entriesCursor-pageSize/2, 0)
	case "enter", "o":
		// Open in browser and mark as read.
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.entriesCursor]
		var cmds []tea.Cmd
		if entry.Status != "read" {
			cmds = append(cmds, setEntryStatus(m.client, entry.Id, planetary.UpdateEntriesRequestStatusRead))
		}
		cmds = append(cmds, openURL(entry.Url))
		m.clampEntriesOffset()
		return m, tea.Batch(cmds...)
	case "u":
		// Toggle read / unread.
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.entriesCursor]
		newStatus := planetary.UpdateEntriesRequestStatusRead
		if entry.Status == "read" {
			newStatus = planetary.UpdateEntriesRequestStatusUnread
		}
		return m, setEntryStatus(m.client, entry.Id, newStatus)
	case "s":
		// Toggle star.
		if len(m.entries) == 0 {
			return m, nil
		}
		entry := m.entries[m.entriesCursor]
		return m, toggleStar(m.client, entry.Id, !entry.Starred)
	case "esc", "h", "left":
		m.focus = focusSidebar
	case "q":
		return m, tea.Quit
	}
	m.clampEntriesOffset()
	return m, nil
}

// ---- View ------------------------------------------------------------------

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.err != "" {
		return errStyle.Render("error: " + m.err + "\n\nr refresh · q quit")
	}

	sidebarLines := m.renderSidebarLines()
	mainLines := m.renderMainLines()

	// Join sidebar and main manually: each sidebar line gets '│' appended,
	// then the corresponding main line follows. This avoids all lipgloss
	// Width/border interactions with ANSI-colored content.
	// Cap at m.height to prevent output from scrolling past the visible area.
	totalRows := max(len(sidebarLines), len(mainLines))
	if m.height > 0 && totalRows > m.height {
		totalRows = m.height
	}
	innerWidth := sidebarWidth - 1
	borderColor := lipgloss.NewStyle().Foreground(lipgloss.Color("237"))



	var out strings.Builder
	for i := 0; i < totalRows; i++ {
		// sidebar cell
		sl := ""
		if i < len(sidebarLines) {
			sl = sidebarLines[i]
		} else {
			sl = strings.Repeat(" ", innerWidth)
		}
		out.WriteString(sl)
		out.WriteString(borderColor.Render("│"))

		// main cell
		if i < len(mainLines) {
			out.WriteString(mainLines[i])
		}

		if i < totalRows-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}

// ---- Sidebar ---------------------------------------------------------------

// renderSidebarLines returns one string per terminal row (no trailing newline),
// each exactly innerWidth visible columns wide.
func (m Model) renderSidebarLines() []string {
	innerWidth := sidebarWidth - 1
	var lines []string

	addLine := func(rendered string) {
		lines = append(lines, sidebarLine(rendered, innerWidth))
	}

	// Header
	addLine(headerStyle.Render(truncate("☀ Planetary", innerWidth-2)))
	addLine(strings.Repeat("─", innerWidth))

	// Search input box replaces the nav items when searching.
	if m.searching {
		prompt := searchInputStyle.Render("/") + " "
		cursor := searchCursorStyle.Render("▌")
		query := searchInputStyle.Render(m.searchQuery)
		addLine(prompt + query + cursor)
		addLine("")
	}

	if !m.searching && m.loading && len(m.items) == 0 {
		addLine(dimStyle.Render("  Loading…"))
	}

	feedsLabelShown := false
	for i, item := range m.items {
		// When searching, skip the nav items (All/Unread/Starred) — only show feeds.
		if m.searching && (item.kind == sidebarAll || item.kind == sidebarUnread || item.kind == sidebarStarred) {
			continue
		}
		// Emit the "FEEDS" section label before the first feed or folder.
		if !feedsLabelShown && (item.kind == sidebarFeed || item.kind == sidebarFolder) {
			addLine("")
			label := padRight(" FEEDS", innerWidth)
			addLine(sectionLabelStyle.Render(label))
			feedsLabelShown = true
		}

		isFocused := m.focus == focusSidebar && i == m.sidebarCursor
		isActive := m.focus == focusEntries && i == m.sidebarCursor

		indent := strings.Repeat("  ", item.depth)
		var icon string
		switch item.kind {
		case sidebarAll:
			icon = "≡ "
		case sidebarUnread:
			icon = "○ "
		case sidebarStarred:
			icon = "✦ "
		case sidebarFolder:
			icon = "▸ "
		default:
			icon = "  "
		}

		maxLabel := innerWidth - utf8.RuneCountInString(indent) - 1 - utf8.RuneCountInString(icon)
		label := truncate(item.label, maxLabel)
		plain := padRight(" "+indent+icon+label, innerWidth)

		var rendered string
		switch {
		case isFocused:
			rendered = selectedFocusStyle.Render(plain)
		case isActive:
			rendered = selectedStyle.Render(plain)
		case item.kind == sidebarFolder:
			rendered = folderStyle.Render(plain)
		default:
			rendered = plain
		}
		addLine(rendered)
	}

	// blank rows to fill height (use actual line count as ground truth)
	for len(lines) < m.height-1 {
		addLine("")
	}

	// help line at bottom
	addLine(mutedStyle.Render(truncate("  r reload · q quit", innerWidth)))

	return lines
}

// sidebarLine pads a pre-rendered (possibly ANSI-colored) string to innerWidth
// visible columns by appending plain spaces. It does NOT pass the string
// through any lipgloss Width() render, avoiding the wrapping bug.
func sidebarLine(rendered string, innerWidth int) string {
	visible := ansiStripWidth(rendered)
	need := innerWidth - visible
	if need <= 0 {
		return rendered
	}
	return rendered + strings.Repeat(" ", need)
}

// ansiStripWidth returns the visible (column) width of s, ignoring ANSI escapes.
func ansiStripWidth(s string) int {
	w := 0
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		w++
	}
	return w
}

// ---- Main panel ------------------------------------------------------------

// renderMainLines returns the main panel content as a slice of lines.
func (m Model) renderMainLines() []string {
	mainWidth := m.width - sidebarWidth - 1
	if mainWidth < 10 {
		mainWidth = 10
	}

	var content string
	if m.loading {
		content = dimStyle.Render("  Loading…")
	} else {
		content = m.renderEntryList(mainWidth)
	}

	return strings.Split(content, "\n")
}

func (m Model) renderEntryList(width int) string {
	var sb strings.Builder

	title := "Entries"
	if m.searching {
		title = "Search"
	} else if len(m.items) > 0 && m.sidebarCursor < len(m.items) {
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

	headerLine := headerStyle.Render(title)
	if len(m.entries) > 0 {
		headerLine += " " + dimStyle.Render(formatPos(m.entriesCursor+1, len(m.entries)))
	}
	sb.WriteString(headerLine + "\n")
	sb.WriteString(strings.Repeat("─", width) + "\n")

	if len(m.entries) == 0 {
		sb.WriteString("\n" + dimStyle.Render("  No entries here.") + "\n")
	} else {
		isFocus := m.focus == focusEntries

		// Fixed column widths (all in visible chars):
		//   prefix: " "(1) + dot(2) + star(2) = 5
		//   date:   "2006-01-02"(10) + " "(1) = 11
		//   feed:   up to 20 chars + " "(1) = 21
		//   title:  remainder
		const prefixLen = 5
		const dateLen = 10
		const feedLen = 20
		const gaps = 3 // spaces: after title, after feed, trailing
		titleWidth := width - prefixLen - dateLen - feedLen - gaps
		if titleWidth < 10 {
			titleWidth = 10
		}

		pageSize := m.entriesPageSize()
		start := m.entriesOffset
		end := start + pageSize
		if end > len(m.entries) {
			end = len(m.entries)
		}

		for idx := start; idx < end; idx++ {
			e := m.entries[idx]
			isSelected := idx == m.entriesCursor

			feedName := ""
			if f, ok := m.feedsByID[e.FeedId]; ok {
				feedName = f.Title
			}

			// buildPlain assembles a raw (no-ANSI) row of exact width.
			buildPlain := func(dotCh, starCh string) string {
				t := padRight(truncate(e.Title, titleWidth), titleWidth)
				f := padRight(truncate(feedName, feedLen), feedLen)
				d := e.PublishedAt.Format("2006-01-02")
				return padRight(" "+dotCh+starCh+t+" "+f+" "+d+" ", width)
			}

			dotPlain := "● "
			if e.Status == "read" {
				dotPlain = "· "
			}
			starPlain := "✦ "
			if !e.Starred {
				starPlain = "  "
			}

			var line string
			switch {
			case isSelected && isFocus:
				line = selectedFocusStyle.Render(buildPlain(dotPlain, starPlain))
			case isSelected:
				line = selectedStyle.Render(buildPlain(dotPlain, starPlain))
			case e.Status == "read":
				line = dimStyle.Render(buildPlain(dotPlain, starPlain))
			default:
				// unread: colorize dot and star inline, plain text otherwise
				dot := unreadDotStyle.Render("● ") // 2 visible cols
				star := "  "
				if e.Starred {
					star = starStyle.Render("✦ ")
				}
				t := padRight(truncate(e.Title, titleWidth), titleWidth)
				f := padRight(truncate(feedName, feedLen), feedLen)
				d := dimStyle.Render(e.PublishedAt.Format("2006-01-02"))
				line = " " + dot + star + t + " " + dimStyle.Render(f) + " " + d + " "
			}

			sb.WriteString(line + "\n")
		}
	}

	var helpText string
	switch {
	case m.searching:
		helpText = "type to search · esc cancel · enter confirm · / from anywhere"
	case m.focus == focusSidebar:
		helpText = "j/k navigate · l open · / search · q quit"
	default:
		helpText = "j/k · ^d/^u · g/G · enter/o open · u read/unread · s star · / search · h sidebar · q quit"
	}
	sb.WriteString("\n" + helpStyle.Render(helpText))

	return sb.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatPos(cur, total int) string {
	return fmt.Sprintf("%d/%d", cur, total)
}

// entriesPageSize returns how many entry rows fit in the main panel.
// Layout: 1 header + 1 separator + N entries + 1 blank + 1 help = height rows.
func (m Model) entriesPageSize() int {
	n := m.height - 4 // header, sep, blank, help
	if n < 1 {
		n = 1
	}
	return n
}

// clampEntriesOffset adjusts entriesOffset so that entriesCursor stays visible.
func (m *Model) clampEntriesOffset() {
	pageSize := m.entriesPageSize()
	if m.entriesCursor < m.entriesOffset {
		m.entriesOffset = m.entriesCursor
	}
	if m.entriesCursor >= m.entriesOffset+pageSize {
		m.entriesOffset = m.entriesCursor - pageSize + 1
	}
	if m.entriesOffset < 0 {
		m.entriesOffset = 0
	}
}


