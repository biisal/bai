package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/files"
	"github.com/biisal/bai/internal/tui/styles"
)

func (m Model) View() tea.View {
	var promptView string
	var promptSize CompSize
	if m.components.showPrompt {
		promptView, promptSize = m.components.prompt.View()
	} else {
		promptView, promptSize = m.components.Input()
	}
	provider, modelID := m.gateway.Active()
	footer, footerHeight := m.components.Footer(FooterProps{
		Provider: provider.Name(),
		ModelID:  modelID,
	})

	commandsView := strings.TrimRight(m.commands.View(), "\n")
	dirInfo := styles.StyleFooter.Render(files.CurrentDirWithGitCache)
	spinner := m.components.SpinnerStatus()

	usedHeight := promptSize.Height + footerHeight + lipgloss.Height(dirInfo)
	if commandsView != "" {
		usedHeight += lipgloss.Height(commandsView)
	}
	if spinner != "" {
		usedHeight += lipgloss.Height(spinner)
	}

	queueView := m.components.QueueView(m.Width)
	if queueView != "" {
		usedHeight += lipgloss.Height(queueView)
	}

	chatView := m.components.ChatViewPort(m.Height - usedHeight)

	rows := []string{dirInfo, chatView}
	if spinner != "" {
		rows = append(rows, spinner)
	}

	if queueView != "" {
		rows = append(rows, queueView)
	}
	rows = append(rows, promptView)
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
