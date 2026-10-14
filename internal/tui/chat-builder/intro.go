package chatbuilder

import (
	"strings"

	"github.com/biisal/bai/internal/consts"
	"github.com/biisal/bai/internal/tui/styles"
)

var helps = []struct{ key, val string }{
	{"escape", "interrupt"},
	{"ctrl+c", "exit"},
	{"/", "commands"},
	{"@", "find files"},
}

func Intro() string {
	var sb strings.Builder

	sb.WriteString("\n\n")

	logo := styles.StyleIntroLogo.Render("bai")
	ver := styles.StyleIntroVersion.Render(consts.Version)

	sb.WriteString(styles.StyleSystemNotice.Render(logo + " " + ver))

	for _, h := range helps {
		key := styles.StyleIntroHelpKey.Render(h.key)
		val := styles.StyleIntroHelpVal.Render(h.val)
		sb.WriteString("\n")
		sb.WriteString(styles.StyleIntroHelpLine.Render(key + ": " + val))
	}

	sb.WriteString("\n\n")

	return sb.String()
}
