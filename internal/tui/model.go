package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
)

type Model struct {
	gateway   *agent.Gateway
	componets Component
	Width     int
	Height    int
}

func InitModel(gateway *agent.Gateway) *Model {
	comp := NewComponent()
	return &Model{gateway: gateway, componets: comp}
}

func waitForMsg(msgChan chan agent.Message) tea.Cmd {
	return func() tea.Msg {
		return agent.Message(<-msgChan)
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForMsg(m.gateway.MsgChan),
	)
}
