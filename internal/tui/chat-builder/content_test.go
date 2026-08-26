package chatbuilder

import (
	"reflect"
	"strings"
	"testing"

	"github.com/biisal/bai/internal/domain"
	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/biisal/bai/internal/tui/styles"
	test_utils "github.com/biisal/bai/utils/tests"
)

func TestContent_AddSegment(t *testing.T) {
	tests := []struct {
		name   string
		testFn func(t *testing.T, c *Content)
	}{
		{
			name: "should work with zero value content",
			testFn: func(t *testing.T, c *Content) {
				t.Helper()

				c.AddSegment(broker.EventUserMessage, "hello", false)

				if c.active == nil {
					t.Fatal("expected an active segment")
				}
				if got := c.active.buf.String(); got != "hello" {
					t.Errorf("expected hello, got %s", got)
				}
				if c.active.Kind != broker.EventUserMessage {
					t.Errorf("expected %v, got %v", broker.EventUserMessage, c.active.Kind)
				}
				if len(c.blocks) != 0 {
					t.Errorf("expected 0 blocks, got %d", len(c.blocks))
				}
			},
		},
		{
			name: "should append multiple pieces of the same kind",
			testFn: func(t *testing.T, c *Content) {
				t.Helper()

				c.AddSegment(broker.EventUserMessage, "hello", false)
				c.AddSegment(broker.EventUserMessage, " ", false)
				c.AddSegment(broker.EventUserMessage, "world", false)

				if got := c.active.buf.String(); got != "hello world" {
					t.Errorf("expected %q, got %q", "hello world", got)
				}
				if len(c.blocks) != 0 {
					t.Errorf("expected 0 blocks, got %d", len(c.blocks))
				}
			},
		},
		{
			name: "should preserve the complete previous segment",
			testFn: func(t *testing.T, c *Content) {
				t.Helper()

				c.AddSegment(broker.EventUserMessage, "hello", false)
				c.AddSegment(broker.EventUserMessage, " world", false)

				c.AddSegment(broker.EventAgentThinking, "thinking", false)

				if len(c.blocks) != 1 {
					t.Fatalf("expected 1 block, got %d", len(c.blocks))
				}

				if got := c.blocks[0].buf.String(); got != "hello world" {
					t.Errorf("expected previous block to contain %q, got %q", "hello world", got)
				}
				if c.blocks[0].Kind != broker.EventUserMessage {
					t.Errorf(
						"expected previous block kind %v, got %v",
						broker.EventUserMessage,
						c.blocks[0].Kind,
					)
				}
			},
		},
		{
			name: "should preserve all segments across multiple kind changes",
			testFn: func(t *testing.T, c *Content) {
				t.Helper()

				c.AddSegment(broker.EventUserMessage, "user", false)
				c.AddSegment(broker.EventAgentThinking, "thinking", false)
				c.AddSegment(broker.EventUserMessage, "user again", false)

				if len(c.blocks) != 2 {
					t.Fatalf("expected 2 blocks, got %d", len(c.blocks))
				}

				if got := c.blocks[0].buf.String(); got != "user" {
					t.Errorf("expected first block %q, got %q", "user", got)
				}
				if got := c.blocks[1].buf.String(); got != "thinking" {
					t.Errorf("expected second block %q, got %q", "thinking", got)
				}

				if c.active.Kind != broker.EventUserMessage {
					t.Errorf("expected active kind %v, got %v", broker.EventUserMessage, c.active.Kind)
				}
				if got := c.active.buf.String(); got != "user again" {
					t.Errorf("expected active content %q, got %q", "user again", got)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContent()
			tt.testFn(t, c)
		})
	}
}

func TestNewContent(t *testing.T) {
	tests := []struct {
		name string
		want *Content
	}{
		{
			name: "active should be nil, blocks should be empty not nil,render must not nil",
			want: &Content{
				active:   nil,
				rendered: strings.Builder{},
				blocks:   []*Segment{},
				width:    0,
				height:   0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewContent()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getSb(text string) strings.Builder {
	sb := strings.Builder{}
	sb.WriteString(text)
	return sb
}

func Test_renderSegment(t *testing.T) {
	tests := []struct {
		name  string
		seg   *Segment
		width int
		want  string
	}{
		{
			name:  "apply user input style to text",
			seg:   &Segment{buf: getSb("test_text"), Kind: broker.EventUserMessage},
			width: 0,
			want:  styles.StyleUserInput.MarginBottom(1).Render("test_text"),
		},
		{
			name:  "apply agent response style to text",
			seg:   &Segment{buf: getSb("test_text"), Kind: broker.EventAgentResponse},
			width: 0,
			want:  styles.StyleAgentResponse.MarginBottom(1).Render("test_text"),
		},
		{
			name:  "apply error style to text",
			seg:   &Segment{buf: getSb("test_text"), Kind: broker.EventAgentError},
			width: 0,
			want:  styles.StyleError.MarginBottom(1).Render("test_text"),
		},
		{
			name:  "apply agent thinking style to text",
			seg:   &Segment{buf: getSb("test_text"), Kind: broker.EventAgentThinking},
			width: 0,
			want:  styles.StyleAgentThinking.MarginBottom(1).Render("test_text"),
		},
		{
			name:  "apply system notice style to text",
			seg:   &Segment{buf: getSb("test_text"), Kind: broker.EventSystemNotice},
			width: 0,
			want:  styles.StyleSystemNotice.MarginBottom(1).Render("test_text"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderSegment(tt.seg)
			if got != tt.want {
				t.Errorf("renderSegment() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContent_flushActive(t *testing.T) {
	tests := []struct {
		name   string
		testFn func(t *testing.T, c *Content)
	}{
		{
			name: "should append a new block if active is not nil",
			testFn: func(t *testing.T, c *Content) {
				t.Helper()
				seg := &Segment{buf: getSb("test_text"), Kind: broker.EventUserMessage}
				c.active = seg

				c.flushActive()
				if len(c.blocks) != 1 {
					t.Errorf("flushActive() = %d, want %d", len(c.blocks), 1)
				}
				test_utils.AssertSliceEqual(t, c.blocks, []*Segment{seg})
			},
		},
		{
			name: "should write to renderd with styled text if active is not nil",
			testFn: func(t *testing.T, c *Content) {
				t.Helper()
				kind := broker.EventUserMessage
				text := "test_text"
				seg := &Segment{buf: getSb(text), Kind: kind}
				c.active = seg

				if c.rendered.String() != "" {
					t.Errorf("flushActive() = %s, want empty string", c.rendered.String())
				}
				c.flushActive()

				styled := styles.StyleUserInput.MarginBottom(1).Render(text) + "\n"

				if c.rendered.String() != styled {
					t.Errorf("flushActive() = %s, want %s", c.rendered.String(), styled)
				}
			},
		},
		{
			name: "active should be nil after flushActive",
			testFn: func(t *testing.T, c *Content) {
				t.Helper()
				seg := &Segment{buf: getSb("test_text"), Kind: broker.EventUserMessage}
				c.active = seg

				c.flushActive()
				if c.active != nil {
					t.Errorf("flushActive() = %v, want nil", c.active)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContent()
			tt.testFn(t, c)
		})
	}
}

func TestContent_SetSize(t *testing.T) {
	t.Run("should chnage width and height to the main struct", func(t *testing.T) {
		c := NewContent()
		c.SetSize(80, 24)
		if c.width != 80 || c.height != 24 {
			t.Errorf("SetSize() = (%d, %d), want (80, 24)", c.width, c.height)
		}
	})
}

func TestContent_Render(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, c *Content) (want string)
	}{
		{
			name: "should return rendered content when active segment is nil",
			setup: func(t *testing.T, c *Content) string {
				c.rendered.WriteString("test_text")
				return "test_text"
			},
		},
		{
			name: "should return rendered content with active segment",
			setup: func(t *testing.T, c *Content) string {
				seg := &Segment{Kind: broker.EventAgentResponse, buf: getSb("test_agent_response")}
				c.active = seg
				c.rendered.WriteString("test_rendered")
				return "test_rendered" + renderSegment(seg)
			},
		},
		{
			name: "should render active segment when rendered content is empty",
			setup: func(t *testing.T, c *Content) string {
				seg := &Segment{
					Kind: broker.EventAgentResponse,
					buf:  getSb("test_agent_response"),
				}
				c.active = seg

				return renderSegment(seg)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContent()
			want := tt.setup(t, c)
			got := c.Render()
			if got != want {
				t.Errorf("Render() = %q, want %q", got, want)
			}
		})
	}
}

func TestContent_ReRender(t *testing.T) {
	t.Run("should style all the blocks based on the kind", func(t *testing.T) {
		c := NewContent()
		c.SetSize(80, 24)
		c.blocks = []*Segment{
			{Kind: broker.EventAgentResponse, buf: getSb("test_agent_response")},
			{Kind: broker.EventAgentThinking, buf: getSb("test_agent_response")},
		}

		if c.rendered.String() != "" {
			t.Errorf("expected empty but got %q", c.rendered.String())
		}

		c.ReRender()

		renderdSegString := strings.Builder{}

		for _, block := range c.blocks {
			renderdSegString.WriteString(renderSegment(block))
			renderdSegString.WriteString("\n")
		}

		if c.rendered.String() != renderdSegString.String() {
			t.Errorf("expected %q but got %q", renderdSegString.String(), c.rendered.String())
		}
	})
	t.Run("should return empty when no blocks", func(t *testing.T) {
		c := NewContent()
		c.SetSize(80, 24)
		c.blocks = []*Segment{}

		if c.rendered.String() != "" {
			t.Errorf("expected empty but got %q", c.rendered.String())
		}

		c.ReRender()

		renderdSegString := ""

		if c.rendered.String() != renderdSegString {
			t.Errorf("expected %q but got %q", renderdSegString, c.rendered.String())
		}
	})
}

func assertBlocks(t *testing.T, expected []*Segment, actual []*Segment) {
	if len(expected) != len(actual) {
		t.Errorf("expected %d blocks but got %d", len(expected), len(actual))
	}
	for i := range expected {
		if expected[i].Kind != actual[i].Kind {
			t.Errorf("expected block kind %v but got %v", expected[i].Kind, actual[i].Kind)
		}
		if expected[i].buf.String() != actual[i].buf.String() {
			t.Errorf("expected block content %q but got %q", expected[i].buf.String(), actual[i].buf.String())
		}
	}
}

func TestContent_RerenderFromDbConversation(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) (messages []domain.Message, expectedSegments []*Segment, want string)
	}{
		{
			name: "should render empty when no messages",
			setup: func(t *testing.T) (messages []domain.Message, expectedSegments []*Segment, want string) {
				messages = []domain.Message{}
				want = ""
				return
			},
		},
		{
			name: "active should be nil after rerender",
			setup: func(t *testing.T) (messages []domain.Message, expectedSegments []*Segment, want string) {
				return
			},
		},

		{
			name: "should render sorted segments from db conversation",
			setup: func(t *testing.T) (messages []domain.Message, expectedSegments []*Segment, want string) {
				messages = []domain.Message{
					{
						Role: domain.RoleUser,
						Parts: []domain.Part{
							{Type: domain.PartTextType, Data: domain.TextPartData{Text: "test content 2"}},
						},
					},
					{
						Role: domain.RoleUser,
						Parts: []domain.Part{
							{Type: domain.PartTextType, Data: domain.TextPartData{Text: "test content"}},
						},
					},
				}

				expectedSegments = []*Segment{
					{
						Kind: broker.EventUserMessage,
						buf:  getSb("test content 2"),
					},
					{
						Kind: broker.EventUserMessage,
						buf:  getSb("test content"),
					},
				}
				want = renderSegment(expectedSegments[0]) + "\n" + renderSegment(expectedSegments[1]) + "\n"

				return
			},
		},
		{
			name: "should render thinking parts as separate thinking segments",
			setup: func(t *testing.T) (messages []domain.Message, expectedSegments []*Segment, want string) {
				messages = []domain.Message{
					{
						Role: domain.RoleAssistant,
						Parts: []domain.Part{
							{Type: domain.PartReasoningType, Data: domain.ReasoningPartData{Thinking: "let me think"}},
							{Type: domain.PartTextType, Data: domain.TextPartData{Text: "the answer"}},
						},
					},
				}

				expectedSegments = []*Segment{
					{Kind: broker.EventAgentThinking, buf: getSb("let me think")},
					{Kind: broker.EventAgentResponse, buf: getSb("the answer")},
				}
				want = renderSegment(expectedSegments[0]) + "\n" + renderSegment(expectedSegments[1]) + "\n"

				return
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContent()
			messages, expectedSegments, want := tt.setup(t)
			c.ReRenderFromDbConversation(messages)
			if c.rendered.String() != want {
				t.Errorf("expected %q but got %q", want, c.rendered.String())
			}

			if c.active != nil {
				t.Errorf("expected active to be nil but got %v", c.active)
			}

			assertBlocks(t, expectedSegments, c.blocks)
		})
	}
}
