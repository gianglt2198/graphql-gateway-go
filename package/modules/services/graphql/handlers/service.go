package handlers

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/utils"
)

func RegisterHandler(
	appConfig config.AppConfig,
	subscriber pubsub.QueueSubscriber,
	exec *executor.Executor,
) error {
	ctx := context.Background()

	topic := fmt.Sprintf("%s.graphql", appConfig.Name)

	if err := subscriber.QueueSubscribe(ctx, topic, appConfig.Name, func(ctx context.Context, msg pubsub.Message) (any, error) {
		ctx = graphql.StartOperationTrace(ctx)
		params := &graphql.RawParams{
			ReadTime: graphql.TraceTiming{
				Start: time.Now(),
				End:   graphql.Now(),
			},
		}

		bodyReader := io.NopCloser(strings.NewReader(string(msg.Data)))
		if err := utils.JsonDecode(bodyReader, &params); err != nil {
			gqlErr := gqlerror.Errorf("decoded err: %+v", err)
			resp := &graphql.Response{Errors: []*gqlerror.Error{gqlErr}}
			return resp, err
		}

		return handleGraphql(ctx, params, exec), nil
	}); err != nil {
		return err
	}

	return nil
}

func handleGraphql(ctx context.Context, params *graphql.RawParams, exec *executor.Executor) *graphql.Response {
	// ctx = dataloader.NewContextWithDataLoader(ctx)
	rc, Operr := exec.CreateOperationContext(ctx, params)
	if Operr != nil {
		return exec.DispatchError(graphql.WithOperationContext(ctx, rc), Operr)
	}
	var responses graphql.ResponseHandler
	responses, ctx = exec.DispatchOperation(ctx, rc)
	return responses(ctx)
}
