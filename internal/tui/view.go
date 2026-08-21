package tui

import (
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	inputView, intputSize := m.componets.Input()

	provider, modelID := m.gateway.Active()
	footer := m.componets.Footer(FooterProps{
		Provider: provider,
		ModelID:  modelID,
	})

	commandsView := m.commands.View()
	if commandsView != "" {
		commandsView = strings.TrimRight(commandsView, "\n")
	}

	footerHeight := lipgloss.Height(footer)

	rows := []string{"", inputView}
	totalHeight := intputSize.Height + footerHeight
	if commandsView != "" {
		rows = append(rows, commandsView)
		totalHeight += lipgloss.Height(commandsView)
	}

	viewPortHeight := m.Height - totalHeight
	chatView := m.componets.ChatViewPort(viewPortHeight)
	rows[0] = chatView

	rows = append(rows, footer)

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Top, rows...))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	slog.Debug("view", "width", m.Width, "height", m.Height)
	return v
}
