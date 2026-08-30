package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/config"
)

var (
	StyleAgentThinking = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("8"))

	StyleToolFileReading = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("10"))
	StyleToolFileWriting = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("11"))
	StyleToolBash        = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("14"))

	StyleAgentResponse = lipgloss.NewStyle().Padding(0, 1)
	StyleError         = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("1"))
	StyleUserInput     = lipgloss.NewStyle().Padding(1).Background(lipgloss.Color("236"))
	StyleSystemNotice  = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("3"))
	StyleIntroLogo     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	StyleIntroVersion  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	StyleIntroHelpKey  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	StyleIntroHelpVal  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
	StyleIntroHelpLine = lipgloss.NewStyle().PaddingLeft(1)
	StyleFooter        = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("244"))

	StyleViewportSelectedHighlight = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	StyleInput                     = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false).BorderTopForeground(lipgloss.Color("6"))

	StyleCursorFocusedColor = lipgloss.Color("7")
	StyleCursorBlurredColor = lipgloss.Color("240")

	StyleColorBackground color.Color = nil
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

func UpdateStylesUsingConfigTheme(theme *config.Theme) {
	StyleAgentThinking = StyleAgentThinking.Foreground(lipgloss.Color(theme.MutedForeground))
	StyleToolFileReading = StyleToolFileReading.Foreground(lipgloss.Color(theme.Success))
	StyleToolFileWriting = StyleToolFileWriting.Foreground(lipgloss.Color(theme.Warning))
	StyleToolBash = StyleToolBash.Foreground(lipgloss.Color(theme.Accent))
	StyleAgentResponse = StyleAgentResponse.Foreground(lipgloss.Color(theme.Foreground))
	StyleError = StyleError.Foreground(lipgloss.Color(theme.Destructive))

	StyleUserInput = StyleUserInput.Background(lipgloss.Color(theme.Muted))

	StyleSystemNotice = StyleSystemNotice.Foreground(lipgloss.Color(theme.Warning))

	StyleIntroLogo = StyleIntroLogo.Foreground(lipgloss.Color(theme.Primary))
	StyleIntroVersion = StyleIntroVersion.Foreground(lipgloss.Color(theme.MutedForeground))
	StyleIntroHelpKey = StyleIntroHelpKey.Foreground(lipgloss.Color(theme.MutedForeground))
	StyleIntroHelpVal = StyleIntroHelpVal.Foreground(lipgloss.Color(theme.Foreground))

	StyleFooter = StyleFooter.Foreground(lipgloss.Color(theme.MutedForeground))

	StyleViewportSelectedHighlight = StyleViewportSelectedHighlight.Background(lipgloss.Color(theme.Accent))

	StyleInput = StyleInput.BorderTopForeground(lipgloss.Color(theme.Primary))

	StyleCursorFocusedColor = lipgloss.Color(theme.Primary)
	StyleCursorBlurredColor = lipgloss.Color(theme.Muted)

	if theme.Background != "" {
		StyleColorBackground = lipgloss.Color(theme.Background)
	}
}
