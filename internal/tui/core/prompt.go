package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Varient string

const (
	VarientDefault Varient = "default"
	VarientSuccess Varient = "success"
	VarientError   Varient = "error"
)

type PromptOption struct {
	Text    string
	Value   string
	KeyBind string
	Varient Varient
	handler func() tea.Cmd
}

type Prompt struct {
	Text    string
	Options []PromptOption
}

func NewPrompt(text string, options []PromptOption) *Prompt {
	return &Prompt{
		Text:    text,
		Options: options,
	}
}

func (p *Prompt) View() (string, CompSize) {
	headerView := lipgloss.NewStyle().Render(p.Text)
	buttonsView := ""
	for _, option := range p.Options {
		buttonsView = lipgloss.JoinHorizontal(lipgloss.Left, buttonsView, lipgloss.NewStyle().Render(option.Text))
		buttonsView += " "
	}
	view := lipgloss.JoinVertical(lipgloss.Center, headerView, buttonsView)
	w, h := lipgloss.Size(view)
	return view, CompSize{Height: h, Width: w}
}

func (p *Prompt) Update(msg tea.KeyPressMsg) tea.Cmd {
	var cmd []tea.Cmd
	for _, option := range p.Options {
		if msg.String() == option.KeyBind {
			if option.handler != nil {
				c := option.handler()
				cmd = append(cmd, c)
			}
		}
	}
	return tea.Batch(cmd...)
}
