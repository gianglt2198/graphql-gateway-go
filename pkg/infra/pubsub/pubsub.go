package pubsub

import (
	"context"
)

type (
	Publisher interface {
		Publish(ctx context.Context, topic string, data []byte, attrs map[string]string) error
		Close() error
	}

	Subscriber interface {
		Subscribe(ctx context.Context, topic string, handler Handler)
		Unsubscribe(topic string) error
		Close() error
	}

	QueueSubscriber interface {
		ChanQueueSubscribe(ctx context.Context, topic string, group string, handler Handler) error
		Unsubscribe(topic string) error
		Close() error
	}

	Message struct {
		Topic string
		Data  []byte
	}

	Handler func(ctx context.Context, msg Message) error

	Client interface {
		Publisher
		Subscriber
	}

	Broker[T any] interface {
		Publisher
		QueueSubscriber
	}
)
