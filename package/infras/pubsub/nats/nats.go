package psnats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pingcap/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/infras/serdes"
	"github.com/gianglt2198/federation-go/package/utils"
)

type (
	natsProvider struct {
		cfg     config.NATSConfig
		nc      *nats.Conn
		log     *monitoring.Logger
		factory MessageFactory

		subscriptions map[string]*nats.Subscription
		chans         map[string]chan *nats.Msg

		mu sync.RWMutex
	}
)

var _ pubsub.Client = (*natsProvider)(nil)
var _ pubsub.QueueSubscriber = (*natsProvider)(nil)
var _ pubsub.Broker[nats.Msg] = (*natsProvider)(nil)

type NatsParams struct {
	fx.In

	Log       *monitoring.Logger
	Config    config.NATSConfig
	SeqModels []serdes.Serializer
}

func New(params NatsParams) *natsProvider {
	provider := connect(params.Log, params.Config)
	provider.factory = NewMessageFactory(provider, params.SeqModels)
	return provider
}

func connect(log *monitoring.Logger, cfg config.NATSConfig) *natsProvider {
	options := []nats.Option{
		nats.Name(cfg.Name),
		nats.PingInterval(cfg.PingInterval),
	}

	if cfg.AllowReconnect {
		options = append(options, nats.MaxReconnects(cfg.MaxReconnects))
		options = append(options, nats.ConnectHandler(func(c *nats.Conn) {
			log.GetLogger().Info("Connected to nats successfully")
		}))
		options = append(options, nats.ReconnectHandler(func(c *nats.Conn) {
			log.GetLogger().Info("Reconnected to nats server")
		}))
		options = append(options, nats.DisconnectErrHandler(func(c *nats.Conn, err error) {
			log.GetLogger().Warn("Disconnected from nats server", zap.Error(err))
		}))
	}

	nc, err := nats.Connect(cfg.Endpoint, options...)
	if err != nil {
		log.GetLogger().Panic("Connection error %s", zap.Error(err))
	}

	return &natsProvider{
		cfg:           cfg,
		nc:            nc,
		log:           log,
		subscriptions: make(map[string]*nats.Subscription),
	}
}

func (n *natsProvider) Publish(ctx context.Context, pattern string, data []byte, attrs map[string]string) error {
	msg, err := n.factory.NewMessage(ctx, pattern, data, attrs)
	if err != nil {
		return errors.Wrap(err, "send event failed because encode data to json has error")
	}

	return n.nc.PublishMsg(msg)
}

func (n *natsProvider) Subscribe(ctx context.Context, topic string, handler pubsub.Handler) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.subscriptions[topic]; exists {
		n.log.GetLogger().Warn("Subscription already exists for topic", zap.String("topic", topic))
		return
	}

	sub, err := n.nc.Subscribe(topic, func(msg *nats.Msg) {

		err := handler(ctx, pubsub.Message{Topic: msg.Subject, Data: msg.Data})
		if err != nil {
			n.log.GetLogger().Error("Error processing message",
				zap.String("topic", msg.Subject),
				zap.Error(err))
		}
	})

	if err != nil {
		n.log.GetLogger().Error("Failed to subscribe to topic",
			zap.String("topic", topic),
			zap.Error(err))
		return
	}

	n.subscriptions[topic] = sub

	go func() {
		<-ctx.Done()
		_ = n.Unsubscribe(topic)
	}()

	n.log.GetLogger().Info("Subscribed to topic", zap.String("topic", topic))
}

func (n *natsProvider) QueueSubscribe(ctx context.Context, topic string, group string, handler pubsub.Handler) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.subscriptions[topic]; exists {
		n.log.GetLogger().Warn("Subscription already exists for topic", zap.String("topic", topic))
		return nil
	}

	ch := make(chan *nats.Msg)

	sub, err := n.nc.ChanQueueSubscribe(topic, group, ch)

	if err != nil {
		n.log.GetLogger().Error("Failed to subscribe to topic",
			zap.String("topic", topic),
			zap.Error(err))
		return errors.Wrap(err, "failed to subscribe to topic"+topic)
	}

	n.subscriptions[topic] = sub
	n.chans[topic] = ch

	go func() {
		defer utils.RecoverFn()
		for msg := range ch {
			err := handler(ctx, pubsub.Message{Topic: msg.Subject, Data: msg.Data})
			if err != nil {
				n.log.GetLogger().Error("Error processing message",
					zap.String("topic", msg.Subject),
					zap.Error(err))
			}
		}
	}()

	go func() {
		<-ctx.Done()
		_ = n.Unsubscribe(topic)
	}()

	n.log.GetLogger().Info("Subscribed to topic", zap.String("topic", topic))
	return nil
}

func (n *natsProvider) Unsubscribe(topic string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if v, ok := n.chans[topic]; ok {
		close(v)
		delete(n.chans, topic)
	}

	sub, exists := n.subscriptions[topic]
	if !exists {
		return fmt.Errorf("no subscription found for topic: %s", topic)
	}

	err := sub.Unsubscribe()
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from topic %s: %w", topic, err)
	}

	delete(n.subscriptions, topic)
	n.log.GetLogger().Info("Unsubscribed from topic", zap.String("topic", topic))

	return nil
}

func (n *natsProvider) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for topic, sub := range n.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			n.log.GetLogger().Error("Failed to unsubscribe during close",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}

	n.subscriptions = make(map[string]*nats.Subscription)

	n.nc.Close()
	return nil
}

func (n *natsProvider) Request(ctx context.Context, pattern string, data any, attrs map[string]string, timeout time.Duration) (*nats.Msg, error) {
	msg, err := n.factory.NewMessage(ctx, pattern, data, attrs)
	if err != nil {
		return nil, errors.Wrap(err, "new message error")
	}
	headers := getHeaders(msg)
	n.log.DebugC(ctx, "request to subject", zap.String("subject", msg.Subject), zap.String("type", "request"), zap.Any("headers", headers))
	return n.nc.RequestMsg(msg, timeout*time.Second)
}
