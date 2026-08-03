package tui

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/agent"
)

type CompSize struct {
	Height int
	Width  int
}

type Component struct {
	textArea textarea.Model
	viewport viewport.Model

	chats []agent.Message
}

func NewComponent() Component {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.SetVirtualCursor(true)
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetHeight(3)
	ta.SetWidth(30)
	ta.MinHeight = 3
	ta.MaxHeight = 6

	s := ta.Styles()
	s.Focused.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(s)

	ta.ShowLineNumbers = false

	vp := viewport.New(viewport.WithWidth(30), viewport.WithHeight(5))
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

	return Component{
		viewport: vp,
		textArea: ta,
	}
}

func getSize(v string) CompSize {
	w, h := lipgloss.Size(v)
	return CompSize{
		Height: h - 1,
		Width:  w - 1,
	}
}

func (c Component) Input() (string, CompSize) {
	view := c.textArea.View()
	w, h := lipgloss.Size(view)
	return view, CompSize{
		Height: h - 1,
		Width:  w - 1,
	}
}

func (c Component) UserBubble(msg string, width int) (string, CompSize) {
	bubbleStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#35383F")). // muted slate
		Foreground(lipgloss.Color("#D0D0D8")).
		Padding(1, 2).
		Width(width)

	content := bubbleStyle.Render(msg)
	return content, getSize(content)
}

func (c Component) ThinkingView(content string) string {
	view := lipgloss.NewStyle().Foreground(lipgloss.Color("#3C3C3C")).Render(content)
	return view
}
