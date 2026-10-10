package chatbuilder

import (
	"charm.land/glamour/v2"
)

type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
}

func NewMarkdownRenderer(options ...glamour.TermRendererOption) (*MarkdownRenderer, error) {
	// const styleConfig
	options = append(options, glamour.WithStylePath("dracula"), glamour.WithEnvironmentConfig())
	mdr, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return nil, err
	}
	return &MarkdownRenderer{
		renderer: mdr,
	}, nil
}

func (r *MarkdownRenderer) UpdateSize(width, height int) {
	mdr, err := NewMarkdownRenderer(glamour.WithWordWrap(width))
	if err != nil {
		// TODO handle  errror : fallback to plain text
		return
	}

	r.renderer = mdr.renderer
}

func (c *Content) toMarkdown(content string) string {
	text, err := c.mdRenderer.renderer.Render(content)
	if err != nil {
		return ""
	}
	return text
}
