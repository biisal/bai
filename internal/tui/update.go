package tui

import (
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"
	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/biisal/bai/internal/tui/commands"
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
			if !m.commands.ShowList {
				return m, m.sendMessage(text)
			}
			return m, m.MatchCommand()

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

func (m *Model) MatchCommand() tea.Cmd {
	if !m.commands.ShowList {
		return nil
	}

	switch item := m.commands.List.SelectedItem().(type) {
	case commands.ListItem[commands.CommandItem]:
		slog.Debug("match_command", "name", item.Fields.Name)
		newInput := fmt.Sprintf("/%s ", item.Fields.Name)
		m.componets.textArea.SetValue(newInput)
		m.commands.Sync(newInput)
		return nil

	case commands.ListItem[commands.ModelList]:
		if err := m.gateway.AddOrUpdateProvider(m.ctx, item.Title(), item.Fields.ProviderID, item.Fields.ModelID); err != nil {
			return nil
		}
		m.commands.ShowList = false
		return func() tea.Msg {
			m.broker.Publish(m.ctx, broker.Message{
				Type: broker.EventSystemNotice,
				Text: fmt.Sprintf("Model changed to: %s\n", item.Title()),
			})
			return nil
		}
	}
	return nil
}
