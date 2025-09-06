package manager

import (
	"context"
	"errors"
	"time"

	"github.com/wundergraph/cosmo/router/pkg/pubsub/datasource"
	"golang.org/x/sync/errgroup"
)

func (f *federationManager) startupProviders(ctx context.Context) error {
	const defaultStartupTimeout = 5 * time.Second

	return f.providersActionWithTimeout(ctx, func(ctx context.Context, provider datasource.Provider) error {
		return provider.Startup(ctx)
	}, defaultStartupTimeout, "pubsub provider startup timed out")
}

func (f *federationManager) shutdownProviders(ctx context.Context) error {
	const defaultShutdownTimeout = 5 * time.Second

	return f.providersActionWithTimeout(ctx, func(ctx context.Context, provider datasource.Provider) error {
		return provider.Shutdown(ctx)
	}, defaultShutdownTimeout, "pubsub provider shutdown timed out")
}

func (f *federationManager) providersActionWithTimeout(
	ctx context.Context,
	action func(ctx context.Context, provider datasource.Provider) error,
	timeout time.Duration,
	errorMessage string,
) error {
	cancellableCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	providersGroup := new(errgroup.Group)
	for _, provider := range f.pubsubProviders {
		providersGroup.Go(func() error {
			actionDone := make(chan error, 1)
			go func() {
				actionDone <- action(cancellableCtx, provider)
			}()
			select {
			case err := <-actionDone:
				return err
			case <-timer.C:
				return errors.New(errorMessage)
			}
		})
	}

	return providersGroup.Wait()
}
