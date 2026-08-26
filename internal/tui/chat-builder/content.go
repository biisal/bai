package chatbuilder

import (
	"encoding/json"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/agent/core/tools"
	"github.com/biisal/bai/internal/domain"
	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/biisal/bai/internal/tui/styles"
)

func bashWithDollarPrefix(text string) string {
	return "$ " + text
}

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

	styles.UpdateChatStyleWidth(width)
}

func NewContent() *Content {
	return &Content{
		rendered: strings.Builder{},
		blocks:   []*Segment{},
	}
}

func (c *Content) AddSegment(kind broker.EventType, text string, isComplete bool) {
	if text == "" {
		return
	}
	if c.active == nil || c.active.Kind != kind || isComplete {
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
	out.WriteString(renderSegment(c.active))
	return out.String()
}

func renderSegment(seg *Segment) string {
	var style lipgloss.Style
	content := seg.buf.String()
	switch seg.Kind {
	case broker.EventAgentThinking:
		style = styles.StyleAgentThinking
	case broker.EventAgentError:
		style = styles.StyleError
	case broker.EventAgentResponse:
		style = styles.StyleAgentResponse
	case broker.EventUserMessage:
		style = styles.StyleUserInput
	case broker.EventSystemNotice:
		style = styles.StyleSystemNotice
	case broker.EventSystemNoticeError:
		style = styles.StyleError
	case broker.EventToolFileReading:
		style = styles.StyleToolFileReading
	case broker.EventToolFileWriting:
		style = styles.StyleToolFileWriting
	case broker.EventToolBash:
		style = styles.StyleToolBash
		content = bashWithDollarPrefix(content)
	}
	return style.MarginBottom(1).Render(content)
}

func (c *Content) flushActive() {
	if c.active != nil {
		c.rendered.WriteString(renderSegment(c.active))
		c.blocks = append(c.blocks, c.active)
	}

	c.rendered.WriteString("\n")
	c.active = nil
}

func (c *Content) ReRender() {
	out := strings.Builder{}
	for _, block := range c.blocks {
		out.WriteString(renderSegment(block))
		out.WriteString("\n")
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
			case domain.RoleTool:
				kind = broker.EventToolBash
			case domain.RoleAssistant:
				switch part.Type {
				case domain.PartReasoningType:
					kind = broker.EventAgentThinking
				case domain.PartToolCallType:
					kind = broker.EventToolBash // TODO: update this to proper tool
				default:
					kind = broker.EventAgentResponse
				}
			}

			seg := &Segment{Kind: kind, buf: strings.Builder{}}
			switch part.Type {
			case domain.PartToolResultType:
				continue
			case domain.PartToolCallType:
				if tc, ok := part.Data.(domain.ToolCallPartData); ok {
					switch tc.Name {
					case string(tools.BashTool): // TODO: orgasnize this
						var args struct {
							Command string `json:"command"`
						}
						if err := json.Unmarshal(tc.Input, &args); err == nil && args.Command != "" {
							seg.buf.WriteString(args.Command)
						} else {
							seg.buf.WriteString(string(tc.Input))
						}
					case string(tools.EditFileTool):
						var args struct {
							Path string `json:"path"`
						}
						if err := json.Unmarshal(tc.Input, &args); err == nil && args.Path != "" {
							seg.buf.WriteString(args.Path)
						}
					}
				}
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
