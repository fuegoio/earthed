package tui

import (
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
		return m, toggleStar(m.client, entry.Id, !entry.Starred)
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
		return m, toggleStar(m.client, m.current.Id, !m.current.Starred)
	case "o":
		return m, openURL(m.current.Url)
	}
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
	addLine(headerStyle.Render(truncate("  Planetary", innerWidth-2)))
	addLine(strings.Repeat("─", innerWidth))

	if m.loading && len(m.items) == 0 {
		addLine(dimStyle.Render("  Loading…"))
	}

	for i, item := range m.items {
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
			icon = "★ "
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

	// blank rows to fill height
	usedLines := 2 + len(m.items)
	if m.loading && len(m.items) == 0 {
		usedLines++
	}
	for i := usedLines; i < m.height-1; i++ {
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
	} else if m.showEntry {
		content = m.renderEntryDetail(mainWidth)
	} else {
		content = m.renderEntryList(mainWidth)
	}

	return strings.Split(content, "\n")
}

func (m Model) renderEntryList(width int) string {
	var sb strings.Builder

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
		sb.WriteString("\n" + dimStyle.Render("  No entries here.") + "\n")
	} else {
		isFocus := m.focus == focusEntries

		// date column is always 5 chars "MM/DD"
		const dateLen = 5
		// prefix: " " (1) + dot (2) + star (2) = 5
		const prefixLen = 5
		titleWidth := width - prefixLen - dateLen - 2 // 2 = spaces around date
		if titleWidth < 10 {
			titleWidth = 10
		}

		for i, e := range m.entries {
			isSelected := i == m.entriesCursor

			// Build each component as plain text for accurate width accounting.
			dot := "  "
			if e.Status != "read" {
				dot = unreadDotStyle.Render("● ") // 2 visible cols
			}
			star := "  "
			if e.Starred {
				star = starStyle.Render("★ ") // 2 visible cols
			}

			titleStr := padRight(truncate(e.Title, titleWidth), titleWidth)
			dateStr := dimStyle.Render(e.PublishedAt.Format("01/02"))

			// Assemble: 1 space + dot(2) + star(2) + title(titleWidth) + " " + date(5) + " "
			line := " " + dot + star + titleStr + " " + dateStr + " "

			if isSelected && isFocus {
				// For selection, rebuild the line as plain text so the background
				// style fills the entire row without ANSI width confusion.
				plainTitle := padRight(truncate(e.Title, titleWidth), titleWidth)
				dotPlain := "● "
				if e.Status == "read" {
					dotPlain = "  "
				}
				starPlain := "★ "
				if !e.Starred {
					starPlain = "  "
				}
				datePlain := e.PublishedAt.Format("01/02")
				plainLine := padRight(" "+dotPlain+starPlain+plainTitle+" "+datePlain+" ", width)
				line = selectedFocusStyle.Render(plainLine)
			} else if isSelected {
				plainTitle := padRight(truncate(e.Title, titleWidth), titleWidth)
				dotPlain := "● "
				if e.Status == "read" {
					dotPlain = "  "
				}
				starPlain := "★ "
				if !e.Starred {
					starPlain = "  "
				}
				datePlain := e.PublishedAt.Format("01/02")
				plainLine := padRight(" "+dotPlain+starPlain+plainTitle+" "+datePlain+" ", width)
				line = selectedStyle.Render(plainLine)
			} else if e.Status == "read" {
				line = dimStyle.Render(" " + "  " + "  " + titleStr + " " + e.PublishedAt.Format("01/02") + " ")
			}

			sb.WriteString(line + "\n")
		}
	}

	var helpText string
	if m.focus == focusSidebar {
		helpText = "l/enter open list · q quit"
	} else {
		helpText = "j/k navigate · enter open in browser · s star · h sidebar · q quit"
	}
	sb.WriteString("\n" + helpStyle.Render(helpText))

	return sb.String()
}

func (m Model) renderEntryDetail(width int) string {
	e := m.current
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(headerStyle.Render(truncate(e.Title, width-4)))
	sb.WriteString("\n\n")

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
	sb.WriteString(dimStyle.Render("  "+meta) + "\n")
	sb.WriteString(dimStyle.Render("  "+truncate(e.Url, width-4)) + "\n\n")
	sb.WriteString(strings.Repeat("─", width) + "\n\n")

	content := ""
	if e.Description != nil {
		content = *e.Description
	}
	content = stripTags(content)
	content = wordWrap(content, width-4)
	if len(content) > 4000 {
		content = content[:4000] + "\n\n[truncated]"
	}

	// indent body
	for _, line := range strings.Split(content, "\n") {
		sb.WriteString("  " + line + "\n")
	}

	sb.WriteString("\n" + helpStyle.Render("esc/h back · o open in browser · s toggle star"))
	return sb.String()
}

// stripTags removes HTML tags for plain-text display.
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
	result := out.String()
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(result)
}

// wordWrap wraps long lines to width columns.
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
		for _, w := range words {
			wl := utf8.RuneCountInString(w)
			if col == 0 {
				out.WriteString(w)
				col = wl
			} else if col+1+wl > width {
				out.WriteString("\n" + w)
				col = wl
			} else {
				out.WriteString(" " + w)
				col += 1 + wl
			}
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


