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
	active     *Segment
	rendered   strings.Builder
	blocks     []*Segment
	width      int
	height     int
	mdRenderer *MarkdownRenderer

	// Markdown is rendered a whole document at a time, so re-rendering the
	// active segment on every delta makes one response cost O(n²). settled holds
	// the rendering of the part of it that has already reached a markdown block
	// boundary and settledLen how much of its buffer that covers; Render only
	// has to render the tail that is still being written.
	settled    strings.Builder
	settledLen int
}

func (c *Content) SetSize(width, height int) {
	c.width = width
	c.height = height

	styles.UpdateChatStyleWidth(width)

	c.mdRenderer.UpdateSize(width, height)
	c.resetSettled()
}

// resetSettled drops the cached prefix of the active segment, which has to be
// rendered again whenever the width or the theme changes.
func (c *Content) resetSettled() {
	c.settled.Reset()
	c.settledLen = 0
}

func NewContent() *Content {
	return &Content{
		blocks:     []*Segment{},
		mdRenderer: NewMarkdownRenderer(0, styles.ActiveTheme),
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

	if c.active.Kind == broker.EventAgentResponse {
		c.settle()
	}
}

// settle renders and caches the settled part of the active segment. It only ever
// moves forward, so calling it more often than needed costs nothing.
func (c *Content) settle() {
	text := c.active.buf.String()
	end := settledMarkdown(text)
	if end <= c.settledLen {
		return
	}
	c.settled.WriteString(c.textToMarkdown(text[c.settledLen:end]))
	c.settledLen = end
}

func (c *Content) Render() string {
	if c.active == nil {
		return c.rendered.String()
	}
	var out strings.Builder
	out.Grow(c.rendered.Len() + c.settled.Len() + c.active.buf.Len() + 1)
	out.WriteString(c.rendered.String())

	if c.active.Kind == broker.EventAgentResponse {
		out.WriteString(c.settled.String())
		if tail := c.active.buf.String()[c.settledLen:]; tail != "" {
			out.WriteString(c.textToMarkdown(tail))
		}
		return out.String()
	}

	out.WriteString(renderSegment(c.active, c.textToMarkdown))
	return out.String()
}

// settledMarkdown returns the length of the longest prefix of text that ends on a
// markdown block boundary. Only such a prefix renders the same on its own as it
// does inside the finished document: a blank line inside an open code fence is
// not a boundary, and neither is the blank line a trailing newline leaves behind.
func settledMarkdown(text string) int {
	var (
		stable int
		blank  = -1 // end of the last blank line, kept until another line follows
		fence  bool
		offset int
	)
	for line := range strings.SplitSeq(text, "\n") {
		if blank >= 0 {
			stable, blank = blank, -1
		}
		offset += len(line) + 1
		switch {
		case isFenceLine(line):
			fence = !fence
		case !fence && line == "":
			blank = offset
		}
	}
	return stable
}

func isFenceLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

func renderSegment(seg *Segment, toMarkdown func(string) string) string {
	var style lipgloss.Style
	content := seg.buf.String()
	switch seg.Kind {
	case broker.EventAgentThinking:
		style = styles.StyleAgentThinking
	case broker.EventAgentError:
		style = styles.StyleError
	case broker.EventAgentResponse:
		return toMarkdown(content)
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
		c.rendered.WriteString(renderSegment(c.active, c.textToMarkdown))
		c.blocks = append(c.blocks, c.active)
	}

	c.rendered.WriteString("\n")
	c.active = nil
	c.resetSettled()
}

func (c *Content) ReRender() {
	c.mdRenderer.UpdateSize(c.width, c.height)
	c.resetSettled()
	c.rendered.Reset()
	c.rendered.WriteString(Intro())
	for _, block := range c.blocks {
		c.rendered.WriteString(renderSegment(block, c.textToMarkdown))
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
