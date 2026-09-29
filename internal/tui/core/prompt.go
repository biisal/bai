package tui

import (
	"fmt"
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/git"
	"github.com/biisal/bai/internal/tui/commands"
	"github.com/biisal/bai/internal/tui/styles"
)

type Variant int

const (
	VariantDefault Variant = iota
	VariantSuccess
	VariantError
)

func (v Variant) style() lipgloss.Style {
	switch v {
	case VariantSuccess:
		return styles.StylePromptSuccess
	case VariantError:
		return styles.StylePromptError
	default:
		return styles.StylePromptButton
	}
}

type PromptOption struct {
	Text    string
	KeyBind string
	Variant Variant
	handler func() tea.Cmd
}

type Prompt struct {
	Text    string
	Options []PromptOption
	Width   int
}

func (p *Prompt) View() (string, CompSize) {
	buttons := make([]string, len(p.Options))
	for i, o := range p.Options {
		buttons[i] = o.Variant.style().Render(fmt.Sprintf("%s (%s)", o.Text, o.KeyBind))
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.StylePromptText.Render(p.Text),
		styles.StylePromptText.Render(strings.Join(buttons, " ")),
	)
	view := styles.StyleInput.Width(p.Width).Render(body)
	w, h := lipgloss.Size(view)
	return view, CompSize{Width: w, Height: h}
}

func (p *Prompt) Update(msg tea.KeyPressMsg) tea.Cmd {
	for _, o := range p.Options {
		if o.handler != nil && strings.EqualFold(msg.String(), o.KeyBind) {
			return o.handler()
		}
	}
	return nil
}

func newGitPrompt(repo *git.Git) *Prompt {
	dismiss := func() tea.Msg { return commands.GitInitMsg{} }
	return &Prompt{
		Text: "Initialize a git repository in this directory?",
		Options: []PromptOption{
			{Text: "Yes", KeyBind: "y", Variant: VariantSuccess, handler: func() tea.Cmd {
				if err := repo.Init(); err != nil {
					slog.Error("git init", "error", err)
				}
				if err := repo.InsertToGitIgnore(repo.Directory); err != nil {
					slog.Error("git insert to gitignore", "error", err)
				}
				return dismiss
			}},
			{Text: "No", KeyBind: "n", Variant: VariantError, handler: func() tea.Cmd { return dismiss }},
			{Text: "Don't ask again", KeyBind: "d", handler: func() tea.Cmd {
				return func() tea.Msg { return commands.GitInitMsg{NeverAsk: true} }
			}},
		},
	}
}
