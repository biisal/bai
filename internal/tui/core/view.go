package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/files"
	"github.com/biisal/bai/internal/tui/styles"
)

func (m Model) View() tea.View {
	inputView, inputSize := m.components.Input()
	provider, modelID := m.gateway.Active()
	footer, footerHeight := m.components.Footer(FooterProps{
		Provider: provider.Name(),
		ModelID:  modelID,
	})

	commandsView := strings.TrimRight(m.commands.View(), "\n")
	dirInfo := styles.StyleFooter.Render(files.CurrentDirWithGitCache)
	spinner := m.components.SpinnerStatus()

	usedHeight := inputSize.Height + footerHeight + lipgloss.Height(dirInfo)
	if commandsView != "" {
		usedHeight += lipgloss.Height(commandsView)
	}
	if spinner != "" {
		usedHeight += lipgloss.Height(spinner)
	}

	chatView := m.components.ChatViewPort(m.Height - usedHeight)

	rows := []string{dirInfo, chatView}
	if spinner != "" {
		rows = append(rows, spinner)
	}
	rows = append(rows, inputView)
	if commandsView != "" {
		rows = append(rows, commandsView)
	}
	rows = append(rows, footer)

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Top, rows...))
	v.AltScreen = true
	v.ReportFocus = true
	v.BackgroundColor = styles.StyleColorBackground
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = m.gateway.ActiveConversationTitle()
	return v
}
