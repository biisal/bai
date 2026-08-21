package tui

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	broker "github.com/biisal/bai/internal/pubsub"
	chatbuilder "github.com/biisal/bai/internal/tui/chat-builder"
	"github.com/biisal/bai/internal/tui/commands"
)

type chatContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type Model struct {
	gateway         *agent.Gateway
	broker          broker.Service
	messages        <-chan broker.Message
	componets       Component
	Width           int
	Height          int
	ChatContent     *strings.Builder
	ThinkingContent *strings.Builder
	ctx             context.Context
	content         *chatbuilder.Content

	chatCtx *chatContext

	spinner     spinner.Model
	showSpinner bool

	commands *commands.Commands
}

func InitModel(ctx context.Context, gateway *agent.Gateway, broker broker.Service, providers []config.ProviderConfig) *Model {
	comp := NewComponent()
	commands := commands.NewCommands(ctx, providers, gateway)

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Spinner.FPS = time.Second / 4
	return &Model{
		gateway:   gateway,
		messages:  broker.Subscribe(),
		componets: comp, ctx: ctx,
		ChatContent:     &strings.Builder{},
		ThinkingContent: &strings.Builder{},
		broker:          broker,
		content:         chatbuilder.NewContent(),
		commands:        commands,
		spinner:         s,
		showSpinner:     false,
	}
}

func waitForMsg(msgChan <-chan broker.Message) tea.Cmd {
	slog.Debug("waitForMsg", "msgChan", msgChan)
	return func() tea.Msg {
		return <-msgChan
	}
}

func (m Model) Init() tea.Cmd {
	return waitForMsg(m.messages)
}
