package chatbuilder

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/tui/styles"
)

func TestMarkdownRenderer(t *testing.T) {
	t.Run("zero value Content falls back to plain text", func(t *testing.T) {
		if out := (&Content{}).toMarkdown("hello"); out != "hello" {
			t.Errorf("expected plain text, got %q", out)
		}
	})

	t.Run("UpdateSize changes the wrap width", func(t *testing.T) {
		c := &Content{mdRenderer: NewMarkdownRenderer(40, styles.ActiveTheme)}
		for _, width := range []int{40, 30} {
			c.mdRenderer.UpdateSize(width, 0)
			for i, line := range strings.Split(c.toMarkdown(strings.Repeat("word ", 40)), "\n") {
				if w := lipgloss.Width(line); w > width {
					t.Errorf("width %d: line %d is %d cols wide", width, i, w)
				}
			}
		}
	})
}
