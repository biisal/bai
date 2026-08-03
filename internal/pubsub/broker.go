package broker

import "context"

type Message struct {
	Type EventType
	Text string
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
	case <-ctx.Done():
	}
}
