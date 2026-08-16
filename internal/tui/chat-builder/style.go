package chatbuilder

import "charm.land/lipgloss/v2"

var (
	StyleAgentThinking = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	StyleAgentResponse = lipgloss.NewStyle()
	StyleError         = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	StyleUserInput     = lipgloss.NewStyle().Padding(1).Background(lipgloss.Color("236"))
	StyleSystemNotice  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	StyleFooter        = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)
