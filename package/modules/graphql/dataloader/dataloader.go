package dataloader

import (
	"context"
	"sync"
)

type DataLoader struct {
	Loader sync.Map
}

type contextKey string

const dataLoaderKey contextKey = "dataloader"

func NewContextWithDataLoader(ctx context.Context) context.Context {
	return context.WithValue(ctx, dataLoaderKey, &DataLoader{})
}

func GetDataLoaderFromContext(ctx context.Context) (*DataLoader, bool) {
	loader, ok := ctx.Value(dataLoaderKey).(*DataLoader)
	return loader, ok
}
