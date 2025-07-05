package psnats

import (
	"context"
	"encoding/json"
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
var _ pubsub.Broker = (*natsProvider)(nil)

type NatsParams struct {
	fx.In

	Log      *monitoring.Logger
	Config   config.NATSConfig
	SeqModel serdes.Serializer
}

func New(params NatsParams) *natsProvider {
	provider := connect(params.Log, params.Config)
	provider.factory = NewMessageFactory(provider, params.SeqModel)
	return provider
}

func connect(log *monitoring.Logger, cfg config.NATSConfig) *natsProvider {
	if !cfg.Enabled {
		return nil
	}

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
		chans:         make(map[string]chan *nats.Msg),
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

	subject := n.factory.Subject(topic)

	if _, exists := n.subscriptions[subject]; exists {
		n.log.GetLogger().Warn("Subscription already exists for topic", zap.String("topic", subject))
		return
	}

	sub, err := n.nc.Subscribe(subject, func(msg *nats.Msg) {
		data, err := n.factory.ReadMessage(msg)
		if err != nil {
			n.log.GetLogger().Error("Error reading message",
				zap.String("topic", msg.Subject),
				zap.Error(err))
		}
		ctx = applyHeadersToContext(ctx, msg)

		resp, err := handler(ctx, pubsub.Message{Topic: msg.Subject, Data: data})
		if err != nil {
			n.log.GetLogger().Error("Error processing message",
				zap.String("topic", msg.Subject),
				zap.Error(err))
		}
		if resp != nil {
			if err := msg.RespondMsg(natsResponse(resp)); err != nil {
				n.log.GetLogger().Error("Error responding to message",
					zap.String("topic", msg.Subject),
					zap.Error(err))
			}
		}
	})

	if err != nil {
		n.log.GetLogger().Error("Failed to subscribe to topic",
			zap.String("topic", subject),
			zap.Error(err))
		return
	}

	n.subscriptions[subject] = sub

	go func() {
		<-ctx.Done()
		_ = n.Unsubscribe(subject)
	}()

	n.log.GetLogger().Info("Subscribed to topic", zap.String("topic", topic))
}

func (n *natsProvider) QueueSubscribe(ctx context.Context, topic string, group string, handler pubsub.Handler) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	subject := n.factory.Subject(topic)

	if _, exists := n.subscriptions[subject]; exists {
		n.log.GetLogger().Warn("Subscription already exists for topic", zap.String("topic", subject))
		return nil
	}

	ch := make(chan *nats.Msg)

	sub, err := n.nc.ChanQueueSubscribe(subject, group, ch)

	if err != nil {
		n.log.GetLogger().Error("Failed to subscribe to topic",
			zap.String("topic", subject),
			zap.Error(err))
		return errors.Wrap(err, "failed to subscribe to topic"+topic)
	}

	n.subscriptions[subject] = sub
	n.chans[subject] = ch

	go func() {
		defer utils.RecoverFn()
		for msg := range ch {
			data, err := n.factory.ReadMessage(msg)
			if err != nil {
				n.log.GetLogger().Error("Error reading message",
					zap.String("topic", msg.Subject),
					zap.Error(err))
			}

			ctx = applyHeadersToContext(ctx, msg)

			resp, err := handler(ctx, pubsub.Message{Topic: msg.Subject, Data: data})
			if err != nil {
				n.log.GetLogger().Error("Error processing message",
					zap.String("topic", msg.Subject),
					zap.Error(err))
			}
			if resp != nil {
				if err := msg.RespondMsg(natsResponse(resp)); err != nil {
					n.log.GetLogger().Error("Error responding to message",
						zap.String("topic", msg.Subject),
						zap.Error(err))
				}
			}
		}
	}()

	go func() {
		<-ctx.Done()
		_ = n.Unsubscribe(subject)
	}()

	n.log.GetLogger().Info("Subscribed to topic", zap.String("topic", subject))
	return nil
}

func (n *natsProvider) Unsubscribe(topic string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	subject := n.factory.Subject(topic)

	if v, ok := n.chans[subject]; ok {
		close(v)
		delete(n.chans, subject)
	}

	sub, exists := n.subscriptions[subject]
	if !exists {
		return fmt.Errorf("no subscription found for topic: %s", subject)
	}

	err := sub.Unsubscribe()
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from topic %s: %w", subject, err)
	}

	delete(n.subscriptions, subject)
	n.log.GetLogger().Info("Unsubscribed from topic", zap.String("topic", subject))

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

func (n *natsProvider) Request(ctx context.Context, pattern string, data any, attrs map[string]string, timeout time.Duration, res any) error {
	msg, err := n.factory.NewMessage(ctx, pattern, data, attrs)
	if err != nil {
		return errors.Wrap(err, "new message error")
	}
	headers := getHeaders(msg)
	n.log.DebugC(ctx, "request to subject", zap.String("subject", msg.Subject), zap.String("type", "request"), zap.Any("headers", headers))
	resp, err := n.nc.RequestMsg(msg, timeout*time.Second)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp.Data, res); err != nil {
		return errors.Wrap(err, "unmarshal response error")
	}
	return nil
}

func natsResponse(resp any) *nats.Msg {
	var data []byte
	var err error
	data, err = json.Marshal(resp)
	if err != nil {
		data = []byte(`{"error": "internal server error"}`)
	}
	return &nats.Msg{
		Data: data,
	}
}

func applyHeadersToContext(ctx context.Context, msg *nats.Msg) context.Context {
	headers := msg.Header

	for k, v := range headers {
		if k == "start_time" {
			startTime, err := time.Parse(time.RFC3339Nano, v[0])
			if err != nil {
				return ctx
			}
			ctx = context.WithValue(ctx, "start_time", startTime)
			continue
		}
		ctx = context.WithValue(ctx, k, v[0])
	}

	return ctx
}
