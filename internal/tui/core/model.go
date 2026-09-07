package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/files"
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
	components      *Component
	Width           int
	Height          int
	ChatContent     *strings.Builder
	ThinkingContent *strings.Builder
	ctx             context.Context
	content         *chatbuilder.Content

	windowTitle string

	chatCtx      *chatContext
	messageQueue messageQueue

	commands *commands.Commands
}

func InitModel(ctx context.Context, gateway *agent.Gateway, broker broker.Service, providers []config.ProviderConfig) *Model {
	comp := NewComponent()
	commands := commands.NewCommands(ctx, providers, gateway)

	return &Model{
		gateway:    gateway,
		messages:   broker.Subscribe(),
		components: comp, ctx: ctx,
		ChatContent:     &strings.Builder{},
		ThinkingContent: &strings.Builder{},
		broker:          broker,
		content:         chatbuilder.NewContent(),
		commands:        commands,
		windowTitle:     fmt.Sprintf("bai - %s", files.GetBaseDir()),
	}
}

func waitForMsg(msgChan <-chan broker.Message) tea.Cmd {
	return func() tea.Msg {
		return <-msgChan
	}
}

func (m Model) Init() tea.Cmd {
	return waitForMsg(m.messages)
}
