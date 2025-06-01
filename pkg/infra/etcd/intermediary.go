package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/pubsub"
	"go.uber.org/fx"
)

const TOPIC = "service.discovery"

type (
	intermediary struct {
		provider pubsub.Client
		client   EtcdClient[ServiceInfo]
		cfg      config.IntermediaryConfig
	}

	Intermediary interface {
		GetAllKeys(ctx context.Context) (any, error)
		GetKey(ctx context.Context, key string) (any, error)
		Watch(ctx context.Context, listener chan any)
		Close() error
	}

	ServiceInfo struct {
		Name      string `json:"name"`
		SchemaURL string `json:"schema_url"`
		SDL       string `json:"sdl"`
		Type      string `json:"type"`
	}
)

type IntermediaryParams struct {
	fx.In

	Log      *monitoring.AppLogger
	Config   config.IntermediaryConfig
	Provider pubsub.Client
}

type IntermediaryResutlt struct {
	fx.Out

	Client Intermediary
}

func NewIntermediary(log *monitoring.AppLogger, cfg config.IntermediaryConfig, provider pubsub.Client) IntermediaryResutlt {
	client := NewEtcdClient[ServiceInfo](log, cfg.EtcdConfig)

	return IntermediaryResutlt{Client: &intermediary{
		client:   client,
		provider: provider,
		cfg:      cfg,
	}}
}

func (e *intermediary) GetAllKeys(ctx context.Context) (any, error) {
	return e.client.GetAllKeys(ctx)
}
func (e *intermediary) GetKey(ctx context.Context, key string) (any, error) {
	return e.client.GetKey(ctx, key)
}
func (e *intermediary) Watch(ctx context.Context, listener chan any) {
	e.provider.Subscribe(ctx, TOPIC, func(ctx context.Context, msg pubsub.Message) error {
		var value ServiceInfo
		if err := json.Unmarshal(msg.Data, &value); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
		listener <- value
		if err := e.client.Put(ctx, value.Name, value); err != nil {
			return fmt.Errorf("failed to put data to etcd: %w", err)
		}
		return nil
	})
}
func (e *intermediary) Close() error { return e.client.Close() }
