package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type (
	etcdClient[T any] struct {
		cfg    Config
		log    *monitoring.AppLogger
		client *clientv3.Client
	}

	EtcdClient[T any] interface {
		Put(ctx context.Context, key string, value T) error
		Delete(ctx context.Context, key string) error
		GetKey(ctx context.Context, key string) (*T, error)
		GetAllKeys(ctx context.Context) ([]T, error)
		Watch(ctx context.Context, listener chan T)
		Close() error
	}
)

func NewEtcdClient[T any](log *monitoring.AppLogger, cfg Config) EtcdClient[T] {
	if !cfg.Enabled {
		return nil
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.GetLogger().Error("Connection etcd error %s", zap.Error(err))
		panic(err)
	}

	return &etcdClient[T]{
		cfg:    cfg,
		log:    log,
		client: client,
	}
}

func (e *etcdClient[T]) Put(ctx context.Context, key string, value T) error {
	e.log.GetLogger().Info(fmt.Sprintf("Putting key %s .....", key), zap.Any("value", value))
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	_, err = e.client.Put(ctx, e.buildKey(key), string(data))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	e.log.GetLogger().Info("Put key successfully", zap.Any("key", key))
	return nil
}

func (e *etcdClient[T]) Delete(ctx context.Context, key string) error {
	e.log.GetLogger().Info(fmt.Sprintf("Deleting key %s .....", key))
	_, err := e.client.Delete(ctx, e.buildKey(key))
	if err != nil {
		return fmt.Errorf("failed to unregister service: %w", err)
	}

	e.log.GetLogger().Info("Deleted key successfully", zap.Any("key", key))
	return nil
}

func (e *etcdClient[T]) GetKey(ctx context.Context, key string) (*T, error) {

	res, err := e.client.Get(ctx, e.buildKey(key))
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	if len(res.Kvs) < 1 {
		return nil, fmt.Errorf("not found key")
	}

	kv := res.Kvs[0]
	var result T
	if err := json.Unmarshal(kv.Value, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshaling data: %w", err)
	}

	return &result, nil
}

func (e *etcdClient[T]) GetAllKeys(ctx context.Context) ([]T, error) {
	res, err := e.client.Get(ctx, e.cfg.BasePath, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys : %w", err)
	}

	result := make([]T, 0)

	for _, kv := range res.Kvs {
		var value T
		if err := json.Unmarshal(kv.Value, &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshaling data: %w", err)
		}

		result = append(result, value)
	}

	return result, nil
}

func (e *etcdClient[T]) Watch(ctx context.Context, listener chan T) {
	watcher := e.client.Watch(context.Background(), e.cfg.BasePath, clientv3.WithPrefix())
	for resp := range watcher {
		for _, ev := range resp.Events {
			key := string(ev.Kv.Key)
			keyName := key[len(e.cfg.BasePath)+1:] // skip the trailing slash

			switch ev.Type {
			case clientv3.EventTypePut:
				var value T
				if err := json.Unmarshal(ev.Kv.Value, &value); err != nil {
					e.log.GetLogger().Error("Error unmarshaling value: %v", zap.Error(err))
					continue
				}

				listener <- value

				e.log.GetLogger().Info("Key updated", zap.Any("key name", keyName))
			case clientv3.EventTypeDelete:
				e.log.GetLogger().Info("Key deleted", zap.Any("key name", keyName))
			}
		}
	}
}
func (e *etcdClient[T]) Close() error {
	return e.client.Close()
}

func (e *etcdClient[T]) buildKey(key string) string {
	if e.cfg.BasePath != "" {
		return strings.Join([]string{e.cfg.BasePath, key}, "/")
	}
	return key
}
