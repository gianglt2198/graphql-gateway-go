package fwebsocket

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/wsprotocol"
	"github.com/tidwall/gjson"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"
	"go.uber.org/zap"
)

type websocketResponseWriter struct {
	id           string
	protocol     wsprotocol.Protocol
	header       http.Header
	buf          bytes.Buffer
	writtenBytes int
	logger       *zap.Logger
}

var _ http.ResponseWriter = (*websocketResponseWriter)(nil)
var _ resolve.SubscriptionResponseWriter = (*websocketResponseWriter)(nil)

func newWebsocketResponseWriter(id string, protocol wsprotocol.Protocol, propagateErrors bool, logger *monitoring.Logger) *websocketResponseWriter {
	return &websocketResponseWriter{
		id:       id,
		protocol: protocol,
		header:   make(http.Header),
		logger:   logger.With(zap.String("subscription_id", id)),
	}
}

func (rw *websocketResponseWriter) Header() http.Header {
	return rw.header
}

func (rw *websocketResponseWriter) WriteHeader(statusCode int) {
	rw.logger.Debug("Response status code", zap.Int("status_code", statusCode))
}

func (rw *websocketResponseWriter) Complete() {
	err := rw.protocol.Complete(rw.id)
	if err != nil {
		rw.logger.Debug("Sending complete message", zap.Error(err))
	}
}

func (rw *websocketResponseWriter) Close(kind resolve.SubscriptionCloseKind) {
	err := rw.protocol.Close(kind.WSCode, kind.Reason)
	if err != nil {
		rw.logger.Debug("Sending error message", zap.Error(err))
	}
}

func (rw *websocketResponseWriter) Write(data []byte) (int, error) {
	rw.writtenBytes += len(data)
	return rw.buf.Write(data)
}

func (rw *websocketResponseWriter) Flush() error {
	if rw.buf.Len() > 0 {
		rw.logger.Debug("flushing", zap.Int("bytes", rw.buf.Len()))
		payload := rw.buf.Bytes()
		var extensions []byte
		var err error
		if len(rw.header) > 0 {
			extensions, err = json.Marshal(map[string]any{
				"response_headers": rw.header,
			})
			if err != nil {
				rw.logger.Warn("Serializing response headers", zap.Error(err))
				return err
			}
		}

		// Check if the result is an error
		errorsResult := gjson.GetBytes(payload, "errors")
		if errorsResult.Type == gjson.JSON {
			err = rw.protocol.WriteGraphQLErrors(rw.id, json.RawMessage(`[{"message":"Unable to subscribe"}]`), extensions)
		} else {
			err = rw.protocol.WriteGraphQLData(rw.id, payload, extensions)
		}
		rw.buf.Reset()
		if err != nil {
			return err
		}
	}
	return nil
}

func (rw *websocketResponseWriter) SubscriptionResponseWriter() resolve.SubscriptionResponseWriter {
	return rw
}
