package tui

import (
	"fmt"
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/biisal/bai/internal/git"
	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/biisal/bai/internal/tui/commands"
	"github.com/biisal/bai/internal/tui/styles"
)

type Variant int

const (
	VariantDefault Variant = iota
	VariantSuccess
	VariantError

	narrowWidth = 80
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

func (p *Prompt) UpdateSize(width int) {
	p.Width = width
}

func (p *Prompt) View() (string, CompSize) {
	buttons := make([]string, len(p.Options))
	for i, o := range p.Options {
		buttons[i] = o.Variant.style().Render(fmt.Sprintf("%s [%s]", o.Text, o.KeyBind))
	}

	separator := " "
	if p.Width < narrowWidth {
		separator = "\n\n"
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.StylePromptText.Render(p.Text+"\n"),
		styles.StylePromptText.Render(strings.Join(buttons, separator)),
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

func newGitPrompt(repo git.GitRepo) *Prompt {
	dismiss := func() tea.Msg { return commands.GitInitMsg{} }
	return &Prompt{
		Text: "Want to keep all changes in track, with a custom git repository?",
		Options: []PromptOption{
			{Text: "Yes (Recommended)", KeyBind: "y", Variant: VariantSuccess, handler: func() tea.Cmd {
				cmds := []tea.Cmd{
					dismiss,
				}
				if inited, err := repo.CheckIfGitInitialized(); err == nil && inited {
					cmds = append(
						cmds, func() tea.Msg {
							return broker.Message{
								Type:       broker.EventSystemNoticeError,
								Text:       "Already initialized",
								IsComplete: true,
							}
						},
					)
					return tea.Sequence(cmds...)
				}
				if err := repo.Init(); err != nil {
					slog.Error("git init", "error", err)
				}
				if err := repo.InsertToGitIgnore(repo.Dir()); err != nil {
					slog.Error("git insert to gitignore", "error", err)
				}
				cmds = append(
					cmds, func() tea.Msg {
						return broker.Message{
							Type:       broker.EventSystemNotice,
							Text:       "Git repository initialized successfully.",
							IsComplete: true,
						}
					},
				)
				return tea.Sequence(cmds...)
			}},
			{Text: "No (Don't please!)", KeyBind: "n", Variant: VariantError, handler: func() tea.Cmd { return dismiss }},
			{Text: "Don't ask again", KeyBind: "d", handler: func() tea.Cmd {
				cmds := []tea.Cmd{
					func() tea.Msg {
						return broker.Message{
							Type:       broker.EventSystemNotice,
							Text:       fmt.Sprintf("we won't ask again for this Directory :! but still you can re-enable it using %s", commands.GitInitCommand),
							IsComplete: true,
						}
					},
					func() tea.Msg {
						return commands.GitInitMsg{NeverAsk: true, ShowPrompt: false}
					},
				}
				return tea.Sequence(cmds...)
			}},
		},
	}
}
