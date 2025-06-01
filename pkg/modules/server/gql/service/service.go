package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/99designs/gqlgen/graphql/handler/debug"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
	"github.com/gianglt2198/graphql-gateway-go/pkg/utils"
)

type GqlServer struct {
	log        *monitoring.AppLogger
	ccfg       config.BaseConfig
	gcfg       config.GqlServiceConfig
	execSchema graphql.ExecutableSchema
	// brokerClient pubsub.Broker[nats.Msg]
}

type GqlServerParams struct {
	fx.In

	Log        *monitoring.AppLogger
	CCfg       config.BaseConfig
	GCfg       config.GqlServiceConfig
	ExecSchema graphql.ExecutableSchema
	// BrokerClient pubsub.Broker[nats.Msg]
}

type GqlServerResult struct {
	fx.Out

	Server *GqlServer
}

func NewGqlServer(params GqlServerParams) GqlServerResult {
	return GqlServerResult{
		Server: &GqlServer{
			log:        params.Log,
			ccfg:       params.CCfg,
			gcfg:       params.GCfg,
			execSchema: params.ExecSchema,
			// brokerClient: params.BrokerClient,
		},
	}
}

func (g *GqlServer) Start() error {
	exec := executor.New(g.execSchema)
	exec.Use(&extension.Introspection{})
	exec.Use(&extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	exec.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	if g.gcfg.Debug {
		exec.Use(&debug.Tracer{})
	}
	exec.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		return handleGraphqlError(ctx, err)
	})
	exec.SetRecoverFunc(func(ctx context.Context, err any) error {
		e := graphql.DefaultRecover(ctx, err)
		return handleGraphqlError(ctx, NewError(e.Error(), InternalServerErrorCode))
	})

	if g.gcfg.Http.Enabled {
		go func() {
			g.log.GetLogger().Info(fmt.Sprintf("connect to :%v%v for GraqhQL", g.gcfg.Http.Port, g.gcfg.Http.Path))
			if err := registerWithSchema(g.log, registerSchema{
				GraphqlPath:       g.gcfg.Http.Path,
				GraphqlPort:       g.gcfg.Http.Port,
				PlaygroundEnabled: g.gcfg.Http.PlaygroundEnabled,
				Debug:             g.gcfg.Debug,
				Schema:            g.execSchema,
			}); err != nil {
				g.log.GetLogger().Error("RegisterPlaygroundWithSchema", zap.Error(err))
			}
		}()
	}

	return nil
}

func (g *GqlServer) Stop() error {
	// return g.brokerClient.Close()
	return nil
}

func handleGraphqlError(ctx context.Context, e error) *gqlerror.Error {
	err := graphql.DefaultErrorPresenter(ctx, e)
	if e != nil {
		var appErr Error
		if errors.As(e, &appErr) {
			err.Extensions = map[string]interface{}{
				"code":        appErr.GetCode(),
				"request_id":  utils.GetRequestIDFromCtx(ctx),
				"stack_trace": appErr.GetStackTrace(),
			}
		} else {
			err.Extensions = map[string]interface{}{
				"code":       UnknownErrorCode,
				"request_id": utils.GetRequestIDFromCtx(ctx),
			}
		}
	}
	return err
}
