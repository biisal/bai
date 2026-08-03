package tui

import (
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case broker.Message:
		switch msg.Type {
		case broker.EventUserMessage:
			newContent, _ := m.componets.UserBubble(msg.Text, m.Width)
			m.ChatContent.WriteString(newContent)
			m.ChatContent.WriteString("\n\n")
		case broker.EventAgentMessageChunk:
			m.ChatContent.WriteString(msg.Text)
		case broker.EventAgentThinking:
			newContent := m.componets.ThinkingView(msg.Text)
			m.ThinkingContent.WriteString(newContent)
		case broker.EventAgentStartThinking:
			m.ThinkingContent.WriteString("\n\n")
		case broker.EventAgentStopThinking:
			m.ChatContent.WriteString(m.ThinkingContent.String())
			m.ChatContent.WriteString("\n\n")
			m.ThinkingContent.Reset()
		case broker.EventAgentError:
			m.ChatContent.WriteString("\n\n [Error] : ")
			m.ChatContent.WriteString(msg.Text)
			m.ChatContent.WriteString("\n\n")
		}

		return m, waitForMsg(m.messages)
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		m.componets.textArea.SetWidth(msg.Width)
		m.componets.viewport.SetWidth(msg.Width)

		// m.componets.viewport.SetHeight(m.Height - m.componets.textArea.Height())

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			text := m.componets.textArea.Value()
			m.componets.textArea.SetValue("")
			return m, m.sendMessage(text)

		default:
			var cmd tea.Cmd
			m.componets.textArea, cmd = m.componets.textArea.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}
