package tui

import (
	"fmt"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/agent/providers"
	chatbuilder "github.com/biisal/bai/internal/tui/chat-builder"
)

type CompSize struct {
	Height int
	Width  int
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

	s := ta.Styles()
	s.Cursor.Blink = false
	s.Focused.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(s)

	ta.ShowLineNumbers = false

	vp := viewport.New()

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Spinner.FPS = time.Second / 4

	return &Component{
		textArea:     ta,
		chatViewPort: vp,

		spinner: Spinner{model: sp},
	}
}

func (c *Component) SetChatContent(content string) {
	c.chatViewPort.SetContent(content)
}

func (c *Component) ScrollChatToBottom() {
	c.chatViewPort.GotoBottom()
}

func (c *Component) ChatViewPort(height int) string {
	c.chatViewPort.SetHeight(height)
	if c.chatViewPort.PastBottom() {
		c.chatViewPort.GotoBottom()
	}
	return c.chatViewPort.View()
}

func (c Component) Input() (string, CompSize) {
	lineCount := c.textArea.LineInfo().Height
	taHeight := c.textArea.Height()
	if lineCount > taHeight {
		c.textArea.SetHeight(lineCount)
	}
	view := c.textArea.View()
	extraLines := lineCount - taHeight - c.textArea.MaxHeight

	extraLinesView := ""

	if extraLines > 0 {
		extraLinesView = fmt.Sprintf("%d More ^", extraLines)
	}

	view = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false).BorderTopForeground(lipgloss.Color("6")).Render(view)
	view = lipgloss.JoinVertical(lipgloss.Top, extraLinesView, view)
	w, h := lipgloss.Size(view)
	return view, CompSize{
		Height: h,
		Width:  w,
	}
}

type FooterProps struct {
	Provider providers.Provider
	ModelID  string
}

func (c Component) Footer(props FooterProps) string {
	spinnerView := ""

	if c.spinner.showSpinner {
		spinnerView = c.spinner.model.View()
	}

	return chatbuilder.StyleFooter.Render(spinnerView + " " + props.Provider.ID() + " - " + props.ModelID)
}

func (c *Component) handleSpinnerTick(msg spinner.TickMsg) tea.Cmd {
	if !c.spinner.showSpinner {
		return nil
	}
	var cmd tea.Cmd
	slog.Debug("spinner tick", "msg", msg)
	c.spinner.model, cmd = c.spinner.model.Update(msg)
	return cmd
}
