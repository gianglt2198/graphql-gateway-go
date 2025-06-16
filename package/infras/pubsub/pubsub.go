package pubsub

import (
	"context"
	"time"
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
		QueueSubscribe(ctx context.Context, topic string, group string, handler Handler) error
		Unsubscribe(topic string) error
		Close() error
	}

	Broker[T any] interface {
		Request(ctx context.Context, pattern string, data any, attrs map[string]string, timeout time.Duration) (*T, error)
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
)
