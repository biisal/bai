package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

type messageQueue struct {
	items []string
}

func (q *messageQueue) Enqueue(text string) {
	q.items = append(q.items, text)
}

func (q *messageQueue) Dequeue() (string, bool) {
	if len(q.items) == 0 {
		return "", false
	}

	text := q.items[0]
	q.items = q.items[1:]
	return text, true
}

func (q *messageQueue) Len() int {
	return len(q.items)
}

func (m *Model) startChat(text string) tea.Cmd {
	ctx, cancel := context.WithCancel(m.ctx)
	m.chatCtx = &chatContext{ctx: ctx, cancel: cancel}
	m.components.spinner.showSpinner = true
	m.components.spinner.queuedMessages = m.messageQueue.Len()
	return tea.Batch(
		func() tea.Msg { return m.components.spinner.model.Tick() },
		m.streamChat(ctx, text),
	)
}

func (m *Model) submitMessage(text string) tea.Cmd {
	if m.chatCtx != nil {
		m.messageQueue.Enqueue(text)
		m.components.spinner.queuedMessages = m.messageQueue.Len()
		return nil
	}
	return m.startChat(text)
}

func (m *Model) releaseNextQueuedMessage() tea.Cmd {
	next, ok := m.messageQueue.Dequeue()
	if !ok {
		m.chatCtx = nil
		m.components.spinner.showSpinner = false
		m.components.spinner.queuedMessages = 0
		return nil
	}

	return m.startChat(next)
}
