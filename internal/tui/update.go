package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case agent.Message:
		m.componets.chats = append(m.componets.chats, msg)
		return m, waitForMsg(m.gateway.MsgChan)
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		m.componets.textArea.SetWidth(msg.Width)
		m.componets.viewport.SetWidth(msg.Width)

		m.componets.viewport.SetHeight(m.Height - m.componets.textArea.Height())

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":

		default:
			var cmd tea.Cmd
			m.componets.textArea, cmd = m.componets.textArea.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}
