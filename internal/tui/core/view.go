package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/files"
	"github.com/biisal/bai/internal/tui/styles"
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

	dirInfo := styles.StyleFooter.Render(files.CurrentDirWithGitCache)
	rows := []string{dirInfo, "", inputView}
	totalHeight := intputSize.Height + footerHeight
	if commandsView != "" {
		rows = append(rows, commandsView)
		totalHeight += lipgloss.Height(commandsView)
	}

	rows = append(rows, footer)

	viewPortHeight := m.Height - totalHeight

	dirHeight := lipgloss.Height(dirInfo)
	viewPortHeight -= dirHeight

	chatView := m.componets.ChatViewPort(viewPortHeight)

	rows[1] = chatView

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Top, rows...))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}
