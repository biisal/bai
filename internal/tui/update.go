package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	broker "github.com/biisal/bai/internal/pubsub"
)

func (m *Model) sendMessage(text string) tea.Cmd {
	return func() tea.Msg {
		m.broker.Publish(m.ctx, broker.Message{Type: broker.EventUserMessage, Text: text})
		if err := m.gateway.AddUserMessageToDB(m.ctx, text); err != nil {
			return broker.Message{Type: broker.EventAgentError, Text: err.Error()}
		}

		if _, err := m.gateway.StreamChat(m.ctx, text); err != nil {
			return broker.Message{Type: broker.EventAgentError, Text: err.Error()}
		}
		return nil
	}
}

func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h

	m.componets.textArea.SetWidth(w)
	m.componets.viewport.SetWidth(w)

	m.componets.viewport.SetWidth(w)
	m.content.SetSize(w, h)
	m.commands.SetSize(w)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case broker.Message:
		m.content.AddSegment(msg.Type, msg.Text)

		return m, waitForMsg(m.messages)
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		m.content.ReRender()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			text := m.componets.textArea.Value()
			m.componets.textArea.SetValue("")
			return m, m.sendMessage(text)

		default:
			var textCmd, listCmd tea.Cmd
			var current Commad
			m.componets.textArea, textCmd = m.componets.textArea.Update(msg)
			text := m.componets.textArea.Value()

			m.commands.ShowList, current = showList(text)

			m.commands.Update(current)
			m.commands.List, listCmd = m.commands.List.Update(msg)
			return m, tea.Batch(textCmd, listCmd)
		}
	}
	return m, nil
}

func showList(text string) (show bool, command Commad) {
	if !strings.HasPrefix(text, "/") {
		return
	}
	if strings.HasPrefix(text, "/") && !strings.Contains(text, " ") {
		return true, ListCommands
	}
	if strings.HasPrefix(text, "/models ") {
		return true, ListModels
	}
	return
}
