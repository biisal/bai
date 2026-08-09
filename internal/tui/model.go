package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
	chatbuilder "github.com/biisal/bai/internal/chat-builder"
	broker "github.com/biisal/bai/internal/pubsub"
)

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
}

func InitModel(ctx context.Context, gateway *agent.Gateway, broker broker.Service) *Model {
	comp := NewComponent()
	return &Model{
		gateway:   gateway,
		messages:  broker.Subscribe(),
		componets: comp, ctx: ctx,
		ChatContent:     &strings.Builder{},
		ThinkingContent: &strings.Builder{},
		broker:          broker,
		content:         chatbuilder.NewContent(),
	}
}

func waitForMsg(msgChan <-chan broker.Message) tea.Cmd {
	return func() tea.Msg {
		return <-msgChan
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForMsg(m.messages),
	)
}
