package tui

import (
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/biisal/bai/internal/tui/styles"
)

type CompSize struct {
	Height int
}

type Spinner struct {
	model       spinner.Model
	showSpinner bool
}

type Component struct {
	textArea     textarea.Model
	chatViewPort viewport.Model

	spinner Spinner
}

func NewComponent() *Component {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.SetVirtualCursor(true)
	ta.Focus()

	ta.Prompt = "┃ "

	ta.SetHeight(1)
	ta.MaxHeight = 15
	ta.DynamicHeight = true

	s := ta.Styles()
	s.Cursor.Blink = false
	s.Focused.CursorLine = lipgloss.NewStyle()
	s.Cursor.Color = nil
	ta.SetStyles(s)

	ta.ShowLineNumbers = false
	textKeymap := textarea.DefaultKeyMap()
	textKeymap.InsertNewline = key.NewBinding(key.WithKeys("shift+enter", "ctrl+j"))
	ta.KeyMap = textKeymap

	vp := viewport.New()
	vp.SelectedHighlightStyle = styles.StyleViewportSelectedHighlight
	vp.KeyMap = viewport.KeyMap{}

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Spinner.FPS = time.Second / 5

	return &Component{
		textArea:     ta,
		chatViewPort: vp,

		spinner: Spinner{model: sp},
	}
}

func (c *Component) SetCursorActive() {
	s := c.textArea.Styles()
	s.Cursor.Color = nil
	c.textArea.SetStyles(s)
}

func (c *Component) SetCursorInactive() {
	s := c.textArea.Styles()
	s.Cursor.Color = styles.StyleCursorBlurredColor
	c.textArea.SetStyles(s)
}

func (c *Component) SetValue(value string) {
	c.textArea.SetValue(value)
}

func (c *Component) SetChatContent(content string) {
	c.chatViewPort.SetContent(content)
}

func (c *Component) ScrollChatToBottom(msg ...broker.Message) {
	if len(msg) == 0 {
		c.chatViewPort.GotoBottom()
		return
	}

	m := msg[len(msg)-1]

	if m.Type != broker.EventUserMessage {
		if !c.chatViewPort.AtBottom() {
			return
		}
	}

	c.chatViewPort.GotoBottom()
}

func (c *Component) ChatViewPort(height int) string {
	c.chatViewPort.SetHeight(height)
	if c.chatViewPort.PastBottom() {
		c.chatViewPort.GotoBottom()
	}
	return c.chatViewPort.View()
}

func (c *Component) Input() (string, CompSize) {
	view := c.textArea.View()

	view = styles.StyleInput.Render(view)
	_, h := lipgloss.Size(view)
	return view, CompSize{
		Height: h,
	}
}

type FooterProps struct {
	Provider string
	ModelID  string
}

func (c Component) Footer(props FooterProps) (footer string, height int) {
	footer = styles.StyleFooter.Render(props.Provider + " - " + props.ModelID)
	_, height = lipgloss.Size(footer)
	return
}

func (c Component) SpinnerStatus() string {
	if !c.spinner.showSpinner {
		return ""
	}
	return styles.StyleFooter.Padding(1, 0).Render(c.spinner.model.View() + " working...")
}

func (c *Component) handleSpinnerTick(msg spinner.TickMsg) tea.Cmd {
	if !c.spinner.showSpinner {
		return nil
	}
	var cmd tea.Cmd
	c.spinner.model, cmd = c.spinner.model.Update(msg)
	return cmd
}
