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
		if _, err := m.gateway.StreamChat(m.ctx, text); err != nil {
			return broker.Message{Type: broker.EventAgentError, Text: err.Error()}
		}
		return nil
	}
}

func (m *Model) MatchCommand() tea.Cmd {
	if !m.commands.ShowList {
		return nil
	}

	switch item := m.commands.List.SelectedItem().(type) {
	case commands.ConversationItem:
		slog.Debug("match_conversation", "name", item.Title())

		messages, err := m.gateway.GetMessagesByConversationID(m.ctx, item.Conversation.ID)
		if err != nil {
			slog.Error("match_conversation_get_messages", "err", err)
			return func() tea.Msg {
				m.broker.Publish(m.ctx, broker.Message{
					Type: broker.EventSystemNoticeError,
					Text: fmt.Sprintf("Failed to get messages: %v\n", err),
				})
				return nil
			}
		}
		if err := m.gateway.SetActiveConversation(m.ctx, item.Conversation.ID, nil); err != nil {
			slog.Error("match_conversation_set_active", "err", err)
			m.broker.Publish(m.ctx, broker.Message{
				Type: broker.EventSystemNoticeError,
				Text: fmt.Sprintf("Failed to set active conversation: %v\n", err),
			})
			return nil
		}
		m.commands.ShowList = false
		slog.Debug("match_conversation", "messages", messages)
		m.content.ReRenderFromDbConversation(messages)
		return nil
	case commands.CommandItem:
		if item.Name == "exit" {
			m.commands.ShowList = false
			m.broker.Publish(m.ctx, broker.Message{
				Type: broker.EventSystemNotice,
				Text: "Bye.. See you soon!\n",
			})
			return func() tea.Msg {
				return tea.Quit()
			}
		}
		slog.Debug("match_command", "name", item.Name)
		newInput := fmt.Sprintf("/%s ", item.Name)
		m.componets.textArea.SetValue(newInput)
		m.commands.Sync(newInput)
		return nil

	case commands.ModelItem:
		slog.Debug("match_model", "provider", item.ProviderName, "model", item.ModelID)
		if err := m.gateway.AddOrUpdateProvider(m.ctx, item.ProviderName, item.ModelID); err != nil {
			return nil
		}
		m.commands.ShowList = false
		return func() tea.Msg {
			m.broker.Publish(m.ctx, broker.Message{
				Type: broker.EventSystemNotice,
				Text: fmt.Sprintf("Model changed to: %s/%s\n", item.ProviderName, item.ModelID),
			})
			return nil
		}
	}
	return nil
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
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case broker.Message:
		m.content.AddSegment(msg.Type, msg.Text)

		return m, waitForMsg(m.messages)
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		m.content.ReRender()

	case tea.KeyPressMsg:
		if m.commands.HandleKeyPress(msg.String()) {
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			text := m.componets.textArea.Value()
			if text == "" {
				return m, nil
			}
			m.componets.textArea.SetValue("")
			if !m.commands.IsCommand(text) {
				return m, m.sendMessage(text)
			}
			return m, m.MatchCommand()
		}
	}
	var textCmd, listCmd tea.Cmd

	m.componets.textArea, textCmd = m.componets.textArea.Update(msg)
	m.commands.List, listCmd = m.commands.List.Update(msg)

	m.commands.Sync(m.componets.textArea.Value())

	cmds = append(cmds, textCmd, listCmd)
	return m, tea.Batch(cmds...)
}
