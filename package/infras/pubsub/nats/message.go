package psnats

import (
	"context"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	nats "github.com/nats-io/nats.go"

	"github.com/gianglt2198/federation-go/package/infras/serdes"
	"github.com/gianglt2198/federation-go/package/utils"
)

type MessageFactory interface {
	NewMessage(ctx context.Context, pattern string, data any, attrs map[string]string) (*nats.Msg, error)
	Subject(pattern string) string
	ReadMessage(msg *nats.Msg) ([]byte, error)
}

type messageFactory struct {
	provider    *natsProvider
	serdesModel serdes.Serializer
}

func NewMessageFactory(provider *natsProvider, serdesModel serdes.Serializer) MessageFactory {
	return &messageFactory{
		provider:    provider,
		serdesModel: serdesModel,
	}
}

func (f *messageFactory) NewMessage(ctx context.Context, pattern string, in any, attrs map[string]string) (*nats.Msg, error) {
	subject := f.Subject(pattern)
	msg := nats.NewMsg(subject)
	f.setDefaultHeaders(ctx, msg)

	for k, v := range attrs {
		msg.Header.Set(k, attrs[v])
	}

	var (
		data []byte
		err  error
	)

	data, err = f.serdesModel.Encode(in)
	if err != nil {
		return nil, err
	}

	// for _, model := range f.serdesModels {
	// 	data, err = model.Encode(in)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	msg.Data = data
	return msg, nil
}

func (f *messageFactory) ReadMessage(msg *nats.Msg) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	in := msg.Data

	// for i := len(f.serdesModels) - 1; i == 0; i++ {
	// 	err = f.serdesModels[i].Decode(in, data)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	in = data
	// }

	err = f.serdesModel.Decode(in, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (f *messageFactory) Subject(pattern string) string {
	fragments := []string{}
	if f.provider.cfg.BasePath != "" {
		fragments = append(fragments, f.provider.cfg.BasePath)
	}
	fragments = append(fragments, pattern)
	return strings.Join(fragments, ".")
}

func (f *messageFactory) setDefaultHeaders(ctx context.Context, msg *nats.Msg) {
	utils.ApplyTraceIDWithContext(ctx)
	utils.ApplySpanIDWithContext(ctx)
	utils.ApplyRequestIDWithContext(ctx)
	ctx = graphql.StartOperationTrace(ctx)
	userID := utils.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "system"
	}
	msg.Header.Set("user_id", userID)
	msg.Header.Set("from", f.provider.cfg.Name)
	msg.Header.Set("start_time", graphql.GetStartTime(ctx).Format(time.RFC3339Nano))
}
