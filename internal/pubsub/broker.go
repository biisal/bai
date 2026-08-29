package broker

import (
	"context"
	"log/slog"
)

type Message struct {
	Type       EventType
	Text       string
	IsComplete bool
}

type Broker struct {
	msgChan chan Message
}

type Service interface {
	Subscribe() <-chan Message
	Publish(ctx context.Context, msg Message)
}

func New() Service {
	return &Broker{
		msgChan: make(chan Message, 32),
	}
}

func (b *Broker) Subscribe() <-chan Message {
	return b.msgChan
}

func (b *Broker) Publish(ctx context.Context, msg Message) {
	select {
	case b.msgChan <- msg:
		if len(b.msgChan) > 0 {
			slog.Debug("broker: queued", "queued", len(b.msgChan), "type", msg.Type)
		}
	case <-ctx.Done():
		slog.Debug("broker: publish dropped, ctx done", "type", msg.Type)
	}
}
