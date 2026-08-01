package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
)

func (m Model) View() tea.View {
	v := tea.NewView("")
	inputView, size := m.componets.Input()
	viewPort, _ := m.componets.Chats(agent.Message{
		Content: "test",
	})

	m.componets.viewport.SetHeight(m.Height - size.Height)
	v.AltScreen = true

	v.SetContent(viewPort + "\n" + inputView)
	return v
}
