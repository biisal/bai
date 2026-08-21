package chatbuilder

import (
	"log/slog"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/domain"
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

	updateStyleWidth(width)
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
	if c.active == nil {
		return c.rendered.String()
	}
	var out strings.Builder
	out.WriteString(c.rendered.String())
	out.WriteString(renderSegment(c.active, c.width))
	return out.String()
}

func renderSegment(seg *Segment, width int) string {
	var style lipgloss.Style
	switch seg.Kind {
	case broker.EventAgentThinking:
		style = StyleAgentThinking
	case broker.EventAgentError:
		style = StyleError
	case broker.EventAgentResponse:
		style = StyleAgentResponse
	case broker.EventUserMessage:
		style = StyleUserInput
	case broker.EventSystemNotice:
		style = StyleSystemNotice
	case broker.EventSystemNoticeError:
		style = StyleError
	case broker.EventToolFileReading:
		style = StyleToolFileReading
	case broker.EventToolFileWriting:
		style = StyleToolFileWriting
	case broker.EventToolBash:
		style = StyleToolBash
	}
	return style.Render(seg.buf.String())
}

func (c *Content) flushActive() {
	if c.active != nil {
		c.rendered.WriteString(renderSegment(c.active, c.width))
		c.blocks = append(c.blocks, c.active)
	}

	c.rendered.WriteString("\n\n")
	c.active = nil
}

func (c *Content) ReRender() {
	out := strings.Builder{}
	slog.Info("re-rendering content", "blocks", c.blocks)
	for _, block := range c.blocks {
		out.WriteString(renderSegment(block, c.width))
		out.WriteString("\n\n")
	}
	c.rendered.Reset()
	c.rendered.WriteString(out.String())
}

func (c *Content) ReRenderFromDbConversation(messages []domain.Message) {
	var segments []*Segment
	for _, msg := range messages {
		for _, part := range msg.Parts {
			var kind broker.EventType
			switch msg.Role {
			case domain.RoleUser:
				kind = broker.EventUserMessage
			case domain.RoleAssistant:
				switch part.Type {
				case domain.PartReasoningType:
					kind = broker.EventAgentThinking
				default:
					kind = broker.EventAgentResponse
				}
			}

			seg := &Segment{Kind: kind, buf: strings.Builder{}}
			switch part.Type {
			case domain.PartTextType:
				seg.buf.WriteString(part.Data.(domain.TextPartData).Text)
			case domain.PartReasoningType:
				seg.buf.WriteString(part.Data.(domain.ReasoningPartData).Thinking)
			}
			segments = append(segments, seg)
		}
	}
	c.blocks = segments
	c.active = nil
	c.ReRender()
}
