package chatbuilder

import (
	"strings"

	"charm.land/lipgloss/v2"
	broker "github.com/biisal/bai/internal/pubsub"
)

type Segment struct {
	Kind broker.EventType
	buf  strings.Builder
}

type Content struct {
	active   *Segment
	rendered strings.Builder
	blocks   []*Segment
	width    int
	height   int
}

func (c *Content) SetSize(width, height int) {
	c.width = width
	c.height = height
}

func NewContent() *Content {
	return &Content{
		rendered: strings.Builder{},
		blocks:   []*Segment{},
	}
}

func (c *Content) AddSegment(kind broker.EventType, text string) {
	if c.active == nil || c.active.Kind != kind {
		c.flushActive()
		c.active = &Segment{Kind: kind, buf: strings.Builder{}}
	}
	c.active.buf.WriteString(text)
}

func (c *Content) Render() string {
	var out strings.Builder
	out.WriteString(c.rendered.String())
	if c.active != nil {
		out.WriteString(renderSegment(c.active.Kind, c.active, c.width))
	}
	return out.String()
}

func renderSegment(kind broker.EventType, seg *Segment, width int) string {
	var style lipgloss.Style
	switch kind {
	case broker.EventAgentThinking:
		style = StyleThinking
	case broker.EventAgentError:
		style = StyleError
	case broker.EventAgentResponse:
		style = StyleResponse
	case broker.EventUserMessage:
		style = StyleUserInput
	case broker.EventSystemNotice:
		style = StyleSystemNotice
	}
	return style.Width(width).Render(seg.buf.String())
}

func (c *Content) flushActive() {
	if c.active != nil {
		c.rendered.WriteString(renderSegment(c.active.Kind, c.active, c.width))
		c.blocks = append(c.blocks, c.active)
	}

	c.rendered.WriteString("\n\n")
	c.active = nil
}

func (c *Content) ReRender() {
	out := strings.Builder{}
	for _, block := range c.blocks {
		out.WriteString(renderSegment(block.Kind, block, c.width))
		out.WriteString("\n\n")
	}
	c.rendered.Reset()
	c.rendered.WriteString(out.String())
}
