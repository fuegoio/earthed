package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestViewSidebarText(t *testing.T) {
	m := NewModel(nil)

	// Simulate getting a window size.
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = m2.(Model)

	view := m.View()
	if view == "" {
		t.Fatal("View() returned empty string after WindowSizeMsg")
	}

	lines := strings.Split(view, "\n")
	t.Logf("total lines: %d", len(lines))
	for i, l := range lines {
		if i >= 5 {
			break
		}
		t.Logf("line[%d] len=%d visible=%d: %q", i, len(l), ansiStripWidth(l), l)
	}

	// Row 0 should contain "Planetary" from the sidebar header.
	if len(lines) == 0 {
		t.Fatal("no lines in view")
	}
	visible0 := stripAnsiChars(lines[0])
	if !strings.Contains(visible0, "Planetary") {
		t.Errorf("row 0 does not contain 'Planetary'; got: %q", visible0)
	}

	// Sidebar lines should contain nav items.
	sidebarLines := m.renderSidebarLines()
	t.Logf("sidebar lines: %d", len(sidebarLines))
	for i, l := range sidebarLines {
		if i >= 5 {
			break
		}
		t.Logf("sidebar[%d] len=%d visible=%d: %q", i, len(l), ansiStripWidth(l), l)
	}

	// Main lines
	mainLines := m.renderMainLines()
	t.Logf("main lines: %d", len(mainLines))
	for i, l := range mainLines {
		if i >= 5 {
			break
		}
		t.Logf("main[%d] len=%d visible=%d: %q", i, len(l), ansiStripWidth(l), l)
	}
}

func TestAnsiStripWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"\x1b[1mhello\x1b[m", 5},
		{"\x1b[38;5;99mhello\x1b[m", 5},
		{"\x1b[1m\x1b[38;5;99m  Planetary\x1b[m", 11},
		{"─", 1},                     // 3-byte UTF-8 rune, 1 visible column
		{strings.Repeat("─", 27), 27}, // 81 bytes, 27 visible columns
		{"\x1b[38;5;245m" + strings.Repeat("─", 27) + "\x1b[m", 27},
	}
	for _, tt := range tests {
		got := ansiStripWidth(tt.input)
		if got != tt.want {
			t.Errorf("ansiStripWidth(%q) = %d, want %d", tt.input[:min(len(tt.input), 40)], got, tt.want)
		}
	}
}



// stripAnsiChars returns the plain text of s with ANSI escape sequences removed.
func stripAnsiChars(s string) string {
	var out strings.Builder
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
		out.WriteRune(r)
	}
	return out.String()
}
