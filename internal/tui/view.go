package tui

import (
	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	inputView, _ := m.componets.Input()
	m.componets.viewport.SetContent(m.ChatContent.String())

	v := tea.NewView(m.ChatContent.String() + "\n" + m.ThinkingContent.String() + "\n" + inputView)

	return v
}
