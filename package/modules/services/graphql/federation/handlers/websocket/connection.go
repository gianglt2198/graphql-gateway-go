package fwebsocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/executor"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/wsprotocol"
	"github.com/gobwas/ws/wsutil"
	"github.com/wundergraph/graphql-go-tools/execution/graphql"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/plan"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"
	"go.uber.org/zap"
)

type SubscriptionRegistration struct {
	id            resolve.SubscriptionIdentifier
	msg           *wsprotocol.Message
	clientRequest *http.Request
}

type WebSocketConnectionHandlerOptions struct {
	Logger   *monitoring.Logger
	Executor *executor.Executor

	Request        *http.Request
	ResponseWriter http.ResponseWriter

	Connection *wsConnectionWrapper
	Protocol   wsprotocol.Protocol

	InitRequestID string
	ConnectionID  int64
}

type WebSocketConnectionHandler struct {
	ctx      context.Context
	logger   *monitoring.Logger
	executor *executor.Executor

	// request is the original client request. It is not safe for concurrent use.
	// You have to clone it before using it in a goroutine.
	request *http.Request
	writer  http.ResponseWriter

	conn     *wsConnectionWrapper
	protocol wsprotocol.Protocol

	initialPayload json.RawMessage

	initRequestID   string
	connectionID    int64
	subscriptionIDs atomic.Int64
	subscriptions   sync.Map
}

func NewWebSocketConnectionHandler(ctx context.Context, opts WebSocketConnectionHandlerOptions) *WebSocketConnectionHandler {
	return &WebSocketConnectionHandler{
		ctx: ctx,

		logger:   opts.Logger,
		executor: opts.Executor,

		request: opts.Request,
		writer:  opts.ResponseWriter,

		conn:     opts.Connection,
		protocol: opts.Protocol,

		initRequestID: opts.InitRequestID,
		connectionID:  opts.ConnectionID,
	}
}

type graphqlError struct {
	Message string `json:"message"`
	// Extensions *Extensions `json:"extensions,omitempty"`
}

func (h *WebSocketConnectionHandler) requestError(err error) error {
	if errors.As(err, &wsutil.ClosedError{}) {
		h.logger.Debug("Client closed connection")
		return err
	}
	h.logger.Warn("Handling websocket connection", zap.Error(err))
	return h.conn.WriteText(err.Error())
}

func (h *WebSocketConnectionHandler) writeErrorMessage(operationID string, err error) error {
	gqlErrors := []graphqlError{
		{Message: err.Error()},
	}
	payload, err := json.Marshal(gqlErrors)
	if err != nil {
		return fmt.Errorf("encoding GraphQL errors: %w", err)
	}
	return h.protocol.WriteGraphQLErrors(operationID, payload, nil)
}

func (h *WebSocketConnectionHandler) executeSubscription(registration *SubscriptionRegistration) {
	rw := newWebsocketResponseWriter(registration.msg.ID, h.protocol, true, h.logger)

	gqlRequest, err := h.UnmarshalOperationFromBody(registration.msg.Payload)
	if err != nil {
		_ = h.writeErrorMessage(registration.msg.ID, err)
		return
	}

	p, err := h.executor.ExecuteSubscription(h.ctx, gqlRequest, rw, registration.id)
	if err != nil {
		h.logger.Warn("Resolving GraphQL response", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		if _, ok := p.(*plan.SynchronousResponsePlan); ok {
			_ = rw.Flush()
			rw.Complete()
		}
	}
}

func (h *WebSocketConnectionHandler) UnmarshalOperationFromBody(body []byte) (*graphql.Request, error) {
	buf := bytes.NewBuffer(make([]byte, len(body))[:0])
	err := json.Compact(buf, body)
	if err != nil {
		return nil, err
	}

	var gqlRequest graphql.Request
	if err := json.Unmarshal(buf.Bytes(), &gqlRequest); err != nil {
		return nil, fmt.Errorf("unmarshalling GraphQL request: %w", err)
	}
	return &gqlRequest, nil
}

// registerSubscription registers a new subscription with the given message. This method is not safe for concurrent use.
func (h *WebSocketConnectionHandler) registerSubscription(msg *wsprotocol.Message) (*SubscriptionRegistration, error) {
	if msg.ID == "" {
		return nil, fmt.Errorf("missing id in subscribe")
	}
	_, exists := h.subscriptions.Load(msg.ID)
	if exists {
		return nil, fmt.Errorf("subscription with id %q already exists", msg.ID)
	}

	subscriptionID := h.subscriptionIDs.Add(1)
	h.subscriptions.Store(msg.ID, subscriptionID)

	registration := &SubscriptionRegistration{
		id: resolve.SubscriptionIdentifier{
			ConnectionID:   h.connectionID,
			SubscriptionID: subscriptionID,
		},
		msg: msg,
		// executeSubscription is running on a worker pool, so we have to clone the request
		// before passing it to the worker pool. The original request is not safe for concurrent use and
		// is needed later to construct the operation context and to clone the resolver context.
		clientRequest: h.request.Clone(h.request.Context()),
	}

	return registration, nil
}

func (h *WebSocketConnectionHandler) handleComplete(msg *wsprotocol.Message) error {
	value, exists := h.subscriptions.Load(msg.ID)
	if !exists {
		return h.requestError(fmt.Errorf("no subscription was registered for ID %q", msg.ID))
	}
	h.subscriptions.Delete(msg.ID)
	subscriptionID, ok := value.(int64)
	if !ok {
		return h.requestError(fmt.Errorf("invalid subscription state for ID %q", msg.ID))
	}
	id := resolve.SubscriptionIdentifier{
		ConnectionID:   h.connectionID,
		SubscriptionID: subscriptionID,
	}
	return h.executor.Resolver.AsyncCompleteSubscription(id)
}

func (h *WebSocketConnectionHandler) HandleMessage(handler *WebSocketConnectionHandler, msg *wsprotocol.Message) error {
	switch msg.Type {
	case wsprotocol.MessageTypePing:
		return handler.protocol.Pong(msg)
	case wsprotocol.MessageTypePong:
		// No action needed for pong, just return nil
		return nil
	case wsprotocol.MessageTypeSubscribe:
		registration, err := handler.registerSubscription(msg)
		if err != nil {
			return handler.requestError(err)
		}
		handler.executeSubscription(registration)
	case wsprotocol.MessageTypeComplete:
		_ = handler.handleComplete(msg)
		return nil
	case wsprotocol.MessageTypeTerminate:
		return errClientTerminatedConnection
	default:
		return fmt.Errorf("unknown message type: %d", msg.Type)
	}

	return nil
}

func (h *WebSocketConnectionHandler) Initialize() (err error) {
	h.initialPayload, err = h.protocol.Initialize()
	if err != nil {
		_ = h.requestError(fmt.Errorf("error initializing session"))
		return err
	}

	return nil
}

func (h *WebSocketConnectionHandler) Complete(rw *websocketResponseWriter) {
	h.subscriptions.Delete(rw.id)
	err := rw.protocol.Complete(rw.id)
	if err != nil {
		return
	}
	_ = rw.Flush()
}

func (h *WebSocketConnectionHandler) Close(unsubscribe bool) {
	if unsubscribe {
		// Remove any pending IDs associated with this connection
		err := h.executor.Resolver.AsyncUnsubscribeClient(h.connectionID)
		if err != nil {
			h.logger.Debug("Unsubscribing client", zap.Error(err))
		}
	}

	err := h.conn.Close()
	if err != nil {
		h.logger.Debug("Closing websocket connection", zap.Error(err))
	}
}
