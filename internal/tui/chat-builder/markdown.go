package chatbuilder

import (
	"log/slog"
	"os"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	glamourstyles "charm.land/glamour/v2/styles"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/tui/styles"
)

const mdMargin = 1

type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
}

func ptr[T any](v T) *T { return &v }

func themed(c string, fallback *string) *string {
	if c == "" {
		return fallback
	}
	return &c
}

func markdownStyle(t *config.Theme) ansi.StyleConfig {
	s := glamourstyles.DarkStyleConfig

	s.Document.Margin = ptr(uint(mdMargin))
	s.Document.Color = themed(t.Foreground, s.Document.Color)

	s.Heading.Color = themed(t.Primary, s.Heading.Color)
	s.H1.Color = themed(t.PrimaryForeground, s.H1.Color)
	s.H1.BackgroundColor = themed(t.Primary, s.H1.BackgroundColor)
	s.H6.Color = themed(t.Primary, s.H6.Color)

	s.BlockQuote.Color = themed(t.MutedForeground, s.BlockQuote.Color)
	s.HorizontalRule.Color = themed(t.Border, s.HorizontalRule.Color)
	s.Enumeration.Color = themed(t.Primary, s.Enumeration.Color)

	s.Link.Color = themed(t.Accent, s.Link.Color)
	s.LinkText.Color = themed(t.Primary, s.LinkText.Color)

	s.Code.Color = themed(t.Warning, s.Code.Color)
	s.Code.BackgroundColor = themed(t.Muted, s.Code.BackgroundColor)
	return s
}

func NewMarkdownRenderer(width int, theme *config.Theme) *MarkdownRenderer {
	opts := []glamour.TermRendererOption{glamour.WithStyles(markdownStyle(theme))}
	if os.Getenv("GLAMOUR_STYLE") != "" {
		opts = append(opts, glamour.WithEnvironmentConfig())
	}
	if width > 0 {
		opts = append(opts, glamour.WithWordWrap(width))
	}

	r, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		slog.Error("error term renderer init", "error", err)
	}
	return &MarkdownRenderer{renderer: r}
}

func (r *MarkdownRenderer) UpdateSize(width, height int) {
	*r = *NewMarkdownRenderer(width, styles.ActiveTheme)
}

func (c *Content) toMarkdown(content string) string {
	if c.mdRenderer == nil || c.mdRenderer.renderer == nil {
		return content
	}
	text, err := c.mdRenderer.renderer.Render(content)
	if err != nil {
		return content
	}
	return text
}
