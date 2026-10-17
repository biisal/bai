package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/biisal/bai/internal/tui/styles"
	"github.com/charmbracelet/x/ansi"
)

// chatTop is the screen row where the chat viewport starts (below the dir info line).
const chatTop = 1

type point struct{ line, col int }

// selection is a mouse text selection in chat content coordinates.
type selection struct {
	anchor, head point
	dragging     bool
	active       bool
}

// ordered returns the selection endpoints, start first.
func (s selection) ordered() (point, point) {
	a, b := s.anchor, s.head
	if b.line < a.line || (b.line == a.line && b.col < a.col) {
		a, b = b, a
	}
	return a, b
}

// colRange returns the [start,end) selected columns on a content line.
func (s selection) colRange(line, width int) (int, int, bool) {
	if !s.active {
		return 0, 0, false
	}
	a, b := s.ordered()
	if line < a.line || line > b.line {
		return 0, 0, false
	}
	start, end := 0, width
	if line == a.line {
		start = a.col
	}
	if line == b.line {
		end = b.col + 1
	}
	return start, end, end > start
}

func (c *Component) pointAt(x, y int) point {
	row := min(max(y-chatTop, 0), max(c.chatViewPort.Height()-1, 0))
	return point{c.chatViewPort.YOffset() + row, max(x, 0)}
}

// handleMouse updates the selection; on release it returns a clipboard command.
func (c *Component) handleMouse(msg tea.MouseMsg) tea.Cmd {
	m := msg.Mouse()
	switch msg.(type) {
	case tea.MouseClickMsg:
		c.sel = selection{}
		if m.Button != tea.MouseLeft || m.Y < chatTop || m.Y >= chatTop+c.chatViewPort.Height() {
			return nil
		}
		p := c.pointAt(m.X, m.Y)
		c.sel = selection{anchor: p, head: p, dragging: true}
	case tea.MouseMotionMsg:
		if c.sel.dragging {
			c.sel.head = c.pointAt(m.X, m.Y)
			c.sel.active = c.sel.head != c.sel.anchor
		}
	case tea.MouseReleaseMsg:
		if !c.sel.dragging {
			return nil
		}
		active := c.sel.active
		text := ""
		if active {
			text = c.selectedText()
		}
		c.sel = selection{}
		if !active {
			return nil
		}
		// Native write first: some terminals ignore OSC52. Its error is fine
		// to drop, OSC52 still covers SSH and headless sessions.
		_ = clipboard.WriteAll(text)
		return tea.Batch(tea.SetClipboard(text), Notify("✓ Copied to clipboard!"))
	}
	return nil
}

func (c *Component) selectedText() string {
	lines := strings.Split(ansi.Strip(c.chatViewPort.GetContent()), "\n")
	a, b := c.sel.ordered()
	var out []string
	for i := a.line; i <= b.line && i < len(lines); i++ {
		seg := ""
		if s, e, ok := c.sel.colRange(i, ansi.StringWidth(lines[i])); ok {
			seg = strings.TrimRight(ansi.Cut(lines[i], s, e), " ")
		}
		if i > a.line && softWrapped(lines[i-1], lines[i], c.chatViewPort.Width()) {
			out[len(out)-1] += " " + strings.TrimLeft(seg, " ")
			continue
		}
		out = append(out, seg)
	}
	return strings.Join(out, "\n")
}

// softWrapped reports whether next only starts a new line because its first
// word did not fit on prev, i.e. the break came from wrapping, not the author.
func softWrapped(prev, next string, width int) bool {
	words := strings.Fields(next)
	prev = strings.TrimRight(prev, " ")
	if prev == "" || len(words) == 0 {
		return false
	}
	// -1 for the right padding of chat blocks.
	return ansi.StringWidth(prev)+1+ansi.StringWidth(words[0]) > width-1
}

// highlightSelection styles the selected cells of the rendered viewport.
func (c *Component) highlightSelection(view string) string {
	if !c.sel.active {
		return view
	}
	lines := strings.Split(view, "\n")
	for i, l := range lines {
		if s, e, ok := c.sel.colRange(c.chatViewPort.YOffset()+i, ansi.StringWidth(l)); ok {
			lines[i] = lipgloss.StyleRanges(l, lipgloss.NewRange(s, e, styles.StyleViewportSelectedHighlight))
		}
	}
	return strings.Join(lines, "\n")
}
