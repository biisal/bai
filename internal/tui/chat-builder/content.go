package chatbuilder

import (
	"encoding/json"
	"strings"

	"charm.land/fantasy"
	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/agent/core/tools"
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
	c.rendered.Reset()
	c.rendered.WriteString(Intro())
	for _, block := range c.blocks {
		c.rendered.WriteString(renderSegment(block))
		c.rendered.WriteString("\n")
	}
}

func (c *Content) ReRenderFromDbConversation(messages []fantasy.Message) {
	var segments []*Segment
	for _, msg := range messages {
		if msg.Role == fantasy.MessageRoleTool {
			continue
		}
		for _, part := range msg.Content {
			var kind broker.EventType
			switch msg.Role {
			case fantasy.MessageRoleUser:
				kind = broker.EventUserMessage
			case fantasy.MessageRoleAssistant:
				switch p := part.(type) {
				case fantasy.ReasoningPart:
					kind = broker.EventAgentThinking
				case fantasy.ToolCallPart:
					switch p.ToolName {
					case tools.BashName:
						kind = broker.EventToolBash
					case tools.EditFileName, tools.WriteFileName:
						kind = broker.EventToolFileWriting
					default:
						kind = broker.EventToolFileReading
					}
				default:
					kind = broker.EventAgentResponse
				}
			}

			seg := &Segment{Kind: kind, buf: strings.Builder{}}
			switch p := part.(type) {
			case fantasy.ToolResultPart:
				continue
			case fantasy.ToolCallPart:
				var args map[string]any
				if err := json.Unmarshal([]byte(p.Input), &args); err == nil {
					if cmd, ok := args["command"].(string); ok && cmd != "" {
						seg.buf.WriteString(cmd)
					} else if path, ok := args["path"].(string); ok && path != "" {
						seg.buf.WriteString(path)
					} else {
						seg.buf.WriteString(p.Input)
					}
				} else {
					seg.buf.WriteString(p.Input)
				}
			case fantasy.TextPart:
				seg.buf.WriteString(p.Text)
			case fantasy.ReasoningPart:
				seg.buf.WriteString(p.Text)
			}
			segments = append(segments, seg)
		}
	}
	c.blocks = segments
	c.active = nil
	c.ReRender()
}
