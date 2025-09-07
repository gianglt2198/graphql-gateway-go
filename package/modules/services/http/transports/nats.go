package transports

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/utils"
)

type NatsTransport struct {
	http.RoundTripper

	logger *logging.Logger
	broker pubsub.Broker
}

type NatsTransportParams struct {
	Upstream *http.Transport
	Logger   *logging.Logger
	Broker   pubsub.Broker
}

func NewNatsTransport(params NatsTransportParams) *NatsTransport {
	params.Upstream.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &NatsTransport{
		RoundTripper: params.Upstream,
		logger:       params.Logger,
		broker:       params.Broker,
	}
}

func (t *NatsTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		t.logger.Error(err.Error())
		return nil, fmt.Errorf("do request: %v", err)
	}

	ctx := utils.GetFiberUserContext(req.Context())

	var result any
	err = t.broker.Request(ctx, req.Host, buf, nil, 5*time.Second, &result)
	if err != nil {
		t.logger.Error(err.Error())
		return nil, fmt.Errorf("do request: %v", err)
	}

	b, _ := json.Marshal(result)
	resp = &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(b)),
	}
	return resp, nil
}
