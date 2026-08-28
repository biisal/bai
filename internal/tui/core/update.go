package tui

import (
	"context"
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/biisal/bai/internal/tui/commands"
)

func (m Model) streamChat(ctx context.Context, text string) tea.Cmd {
	return func() tea.Msg {
		m.broker.Publish(m.ctx, broker.Message{Type: broker.EventUserMessage, Text: text, IsComplete: true})
		if _, err := m.gateway.StreamChat(ctx, text); err != nil {
			m.broker.Publish(m.ctx, broker.Message{Type: broker.EventAgentError, Text: err.Error(), IsComplete: true})
		}
		m.broker.Publish(m.ctx, broker.Message{Type: broker.EventStreamDone, IsComplete: true})
		return nil
	}
}

func (m *Model) MatchCommand() tea.Cmd {
	if !m.commands.ShowList {
		return nil
	}

	switch item := m.commands.List.SelectedItem().(type) {
	case commands.ConversationItem:
		messages, err := m.gateway.GetMessagesByConversationID(m.ctx, item.Conversation.ID)
		if err != nil {
			slog.Error("match_conversation_get_messages", "err", err)
			return func() tea.Msg {
				m.broker.Publish(m.ctx, broker.Message{
					Type:       broker.EventSystemNoticeError,
					Text:       fmt.Sprintf("Failed to get messages: %v\n", err),
					IsComplete: true,
				})
				return nil
			}
		}
		if err := m.gateway.SetActiveConversation(m.ctx, item.Conversation.ID, nil); err != nil {
			slog.Error("match_conversation_set_active", "err", err)
			m.broker.Publish(m.ctx, broker.Message{
				Type:       broker.EventSystemNoticeError,
				Text:       fmt.Sprintf("Failed to set active conversation: %v\n", err),
				IsComplete: true,
			})
			return nil
		}
		m.commands.ShowList = false
		m.content.ReRenderFromDbConversation(messages)
		m.components.SetChatContent(m.content.Render())
		m.components.ScrollChatToBottom()
		return nil

	case commands.CommandItem:
		if item.Name == "models" || item.Name == "sessions" {
			newInput := fmt.Sprintf("/%s ", item.Name)
			m.components.textArea.SetValue(newInput)
			m.commands.Sync(newInput)
			return nil
		}

		return m.commands.ExecuteCommand(item.Name, commands.CommandContext{
			Gateway:    m.gateway,
			Broker:     m.broker,
			Content:    m.content,
			Components: m.components,
			ShowList:   &m.commands.ShowList,
		})

	case commands.ModelItem:
		if err := m.gateway.AddOrUpdateProvider(m.ctx, item.ProviderName, item.ModelID); err != nil {
			return nil
		}
		m.commands.ShowList = false
		return func() tea.Msg {
			m.broker.Publish(m.ctx, broker.Message{
				Type:       broker.EventSystemNotice,
				Text:       fmt.Sprintf("Model changed to: %s/%s\n", item.ProviderName, item.ModelID),
				IsComplete: true,
			})
			return nil
		}
	}
	return nil
}

func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h

	m.components.textArea.SetWidth(w)

	// using viewport
	m.content.SetSize(w-1, h)
	m.commands.SetSize(w)

	m.components.chatViewPort.SetWidth(w)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case spinner.TickMsg:
		return m, m.components.handleSpinnerTick(msg)

	case broker.Message:
		m.content.AddSegment(msg.Type, msg.Text, msg.IsComplete)
		m.components.SetChatContent(m.content.Render())
		m.components.ScrollChatToBottom()

		if msg.Type == broker.EventStreamDone || msg.Type == broker.EventAgentError {
			m.components.spinner.showSpinner = false
		}
		return m, waitForMsg(m.messages)
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		m.content.ReRender()
		m.components.SetChatContent(m.content.Render())

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.chatCtx != nil {
				m.chatCtx.cancel()
				m.chatCtx = nil
				return m, nil
			}
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			text := m.components.textArea.Value()
			if text == "" {
				return m, nil
			}
			m.components.textArea.SetValue("")
			if !m.commands.IsCommand(text) {
				ctx, cancel := context.WithCancel(m.ctx)
				m.chatCtx = &chatContext{ctx: ctx, cancel: cancel}
				m.components.spinner.showSpinner = true
				return m, tea.Batch(
					func() tea.Msg { return m.components.spinner.model.Tick() },
					m.streamChat(ctx, text),
				)
			}
			return m, m.MatchCommand()
		}
	}
	var textCmd, listCmd, vpCmd tea.Cmd

	m.components.textArea, textCmd = m.components.textArea.Update(msg)
	m.commands.List, listCmd = m.commands.List.Update(msg)
	m.components.chatViewPort, vpCmd = m.components.chatViewPort.Update(msg)

	m.commands.Sync(m.components.textArea.Value())

	cmds = append(cmds, textCmd, listCmd, vpCmd)
	return m, tea.Batch(cmds...)
}
