package chatbuilder

import (
	"strings"

	"github.com/biisal/bai/internal/tui/styles"
)

var helps = []struct{ key, val string }{
	{"escape", "interrupt"},
	{"ctrl+c", "exit"},
	{"/", "commands"},
}

func Intro() string {
	var sb strings.Builder

	sb.WriteString("\n\n")

	logo := styles.StyleIntroLogo.Render("bai")
	version := styles.StyleIntroVersion.Render("v0.0.1")

	sb.WriteString(styles.StyleSystemNotice.Render(logo + " " + version))

	for _, h := range helps {
		key := styles.StyleIntroHelpKey.Render(h.key)
		val := styles.StyleIntroHelpVal.Render(h.val)
		sb.WriteString("\n")
		sb.WriteString(styles.StyleIntroHelpLine.Render(key + ": " + val))
	}

	sb.WriteString("\n\n")

	return sb.String()
}
