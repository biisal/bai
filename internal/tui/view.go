package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	chatbuilder "github.com/biisal/bai/internal/chat-builder"
)

func (m Model) Footer() string {
	provider, modelID := m.gateway.Active()
	return chatbuilder.StyleFooter.Render(provider.ID() + " - " + modelID)
}

func (m Model) View() tea.View {
	inputView, _ := m.componets.Input()
	chatView := lipgloss.NewStyle().Render(m.content.Render())

	v := tea.NewView(chatView + "\n\n" + inputView + m.commands.View() + "\n\n" + m.Footer())
	return v
}
