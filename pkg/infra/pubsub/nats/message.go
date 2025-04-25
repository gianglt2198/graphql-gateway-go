package psnats

import (
	"strings"

	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/serdes"
	"github.com/gofiber/utils"
	nats "github.com/nats-io/nats.go"
)

type MessageFactory interface {
	NewMessage(pattern string, data any, attrs map[string]string) (*nats.Msg, error)
	Subject(pattern string) string
}

type messageFactory struct {
	provider     *natsProvider
	serdesModels []serdes.Serializer
}

func NewMessageFactory(provider *natsProvider, serdesModels []serdes.Serializer) MessageFactory {
	return &messageFactory{
		provider:     provider,
		serdesModels: serdesModels,
	}
}

func (f *messageFactory) NewMessage(pattern string, in any, attrs map[string]string) (*nats.Msg, error) {
	subject := f.Subject(pattern)
	msg := nats.NewMsg(subject)
	f.setDefaultHeaders(msg)

	for k, v := range attrs {
		msg.Header.Set(k, attrs[v])
	}

	var (
		data []byte
		err  error
	)

	for _, model := range f.serdesModels {
		data, err = model.Encode(in)
		if err != nil {
			return nil, err
		}
	}

	msg.Data = data
	return msg, nil
}

func (f *messageFactory) Subject(pattern string) string {
	fragments := []string{}
	if f.provider.cfg.BasePath != "" {
		fragments = append(fragments, f.provider.cfg.BasePath)
	}
	fragments = append(fragments, pattern)
	return strings.Join(fragments, ".")
}

func (f *messageFactory) setDefaultHeaders(msg *nats.Msg) {
	msg.Header.Set("id", utils.UUID())
	msg.Header.Set("from", f.provider.cfg.Name)
}
