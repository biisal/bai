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
		switch msg.Type {
		case broker.EventSystemNotice:
			// TODO: Update the ui
		}
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
			if !m.commands.ShowList {
				m.componets.textArea.SetValue("")
				return m, m.sendMessage(text)
			}
			item, ok := m.commands.List.SelectedItem().(item[any])
			if ok && m.commands.ShowList {
				extras, extrasOk := item.extras.(struct {
					ProviderID string
					ModelID    string
				})
				if ok && m.commands.current == ListModels && extrasOk {
					m.gateway.AddOrUpdateProvider(m.ctx, item.title, extras.ProviderID, extras.ModelID)
					m.broker.Publish(m.ctx, broker.Message{
						Type: broker.EventAgentResponse,
						Text: "Model changed",
					})
				}
			}

		default:
			var textCmd, listCmd tea.Cmd
			m.componets.textArea, textCmd = m.componets.textArea.Update(msg)
			text := m.componets.textArea.Value()

			m.commands.Sync(text)

			m.commands.List, listCmd = m.commands.List.Update(msg)
			return m, tea.Batch(textCmd, listCmd)
		}
	}
	return m, nil
}
