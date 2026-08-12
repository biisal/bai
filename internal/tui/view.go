package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	inputView, _ := m.componets.Input()
	chatView := lipgloss.NewStyle().Render(m.content.Render())

	v := tea.NewView(chatView + "\n\n" + inputView + m.commands.View())
	return v
}
