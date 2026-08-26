package styles

import "charm.land/lipgloss/v2"

var (
	StyleAgentThinking = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("8"))

	StyleToolFileReading = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("10"))
	StyleToolFileWriting = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("11"))
	StyleToolBash        = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("14"))

	StyleAgentResponse = lipgloss.NewStyle().Padding(0, 1)
	StyleError         = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("1"))
	StyleUserInput     = lipgloss.NewStyle().Padding(1).Background(lipgloss.Color("236"))
	StyleSystemNotice  = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("3"))
	StyleFooter        = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("244"))

	StyleViewportSelectedHighlight = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	StyleInput                     = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false).BorderTopForeground(lipgloss.Color("6"))
)

func UpdateChatStyleWidth(w int) {
	StyleAgentThinking = StyleAgentThinking.Width(w)
	StyleToolFileReading = StyleToolFileReading.Width(w)
	StyleToolFileWriting = StyleToolFileWriting.Width(w)
	StyleToolBash = StyleToolBash.Width(w)
	StyleAgentResponse = StyleAgentResponse.Width(w)
	StyleError = StyleError.Width(w)
	StyleUserInput = StyleUserInput.Width(w)
	StyleSystemNotice = StyleSystemNotice.Width(w)
	StyleFooter = StyleFooter.Width(w)
}
