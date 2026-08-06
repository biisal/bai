package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	inputView, _ := m.componets.Input()
	chatView := lipgloss.NewStyle().Width(m.Width).Render(m.ChatContent.String() + "\n" + m.ThinkingContent.String() + "\n" + inputView)

	v := tea.NewView(chatView + inputView)

	return v
}
