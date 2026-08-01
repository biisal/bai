package tui

import (
	"strings"

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

func (c Component) Input() (string, CompSize) {
	view := c.textArea.View()
	w, h := lipgloss.Size(view)
	return view, CompSize{
		Height: h - 1,
		Width:  w - 1,
	}
}

func (c Component) Chats(msg agent.Message) (string, CompSize) {
	var s strings.Builder
	for _, v := range c.chats {
		s.WriteString(v.Content + "\n")
	}

	c.viewport.SetContent(s.String())
	view := c.viewport.View()
	w, h := lipgloss.Size(view)
	return view, CompSize{
		Height: h - 1,
		Width:  w - 1,
	}
}
