package logging

import (
	"sync"

	nats "github.com/nats-io/nats.go"

	"github.com/gianglt2198/federation-go/package/config"
)

type NatsCore struct {
	subject string
	nc      *nats.Conn
	sync.Mutex
}

func NewNatsCore(config config.NATSConfig) *NatsCore {
	if config.Endpoint == "" {
		return nil
	}

	nc, err := nats.Connect(config.Endpoint)
	if err != nil {
		panic(err)
	}
	subject := "logging"
	return &NatsCore{
		subject: subject,
		nc:      nc,
	}
}

func (n *NatsCore) Write(p []byte) (int, error) {
	n.Lock()
	defer n.Unlock()

	if err := n.nc.Publish(n.subject, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (n *NatsCore) Sync() error {
	return nil // No need for sync in this implementation
}
