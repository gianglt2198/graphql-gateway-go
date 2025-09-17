package dataloader

import (
	"context"
	"fmt"

	"github.com/vikstrous/dataloadgen"
)

type GenericLoader[K comparable, V any] struct {
	loader *dataloadgen.Loader[K, V]
}

func NewGenericLoader[K comparable, V any](fetch FetchFunc[K, V]) *GenericLoader[K, V] {
	return &GenericLoader[K, V]{
		loader: dataloadgen.NewLoader(
			fetch,
			dataloadgen.WithBatchCapacity(100),
		),
	}
}

type FetchFunc[K comparable, V any] func(ctx context.Context, keys []K) ([]V, []error)

func GetGenericLoader[K comparable, V any](ctx context.Context, fetch FetchFunc[K, V]) *GenericLoader[K, V] {
	dataLoader, ok := GetDataLoaderFromContext(ctx)
	if !ok {
		return NewGenericLoader(fetch)
	}

	loaderKey := fmt.Sprintf("%p", fetch)
	loader, _ := dataLoader.Loader.LoadOrStore(loaderKey, NewGenericLoader(fetch))

	return loader.(*GenericLoader[K, V])
}

func (g *GenericLoader[K, V]) Load(ctx context.Context, id K) (V, error) {
	return g.loader.Load(ctx, id)
}

func (g *GenericLoader[K, V]) LoadMany(ctx context.Context, ids []K) ([]V, []error) {
	values, err := g.loader.LoadAll(ctx, ids)
	if err != nil {
		return nil, []error{err}
	}
	return values, nil
}
