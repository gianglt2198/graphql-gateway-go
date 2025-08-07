package wsprotocol

import (
	"encoding/json"
	"fmt"

	"github.com/gobwas/ws"
	"github.com/tidwall/sjson"
)

type subscriptionsTransportWSMessageType string

const (
	subscriptionsTransportWSMessageTypeConnectionInit      = subscriptionsTransportWSMessageType("connection_init")
	subscriptionsTransportWSMessageTypeConnectionAck       = subscriptionsTransportWSMessageType("connection_ack")
	subscriptionsTransportWSMessageTypeConnectionError     = subscriptionsTransportWSMessageType("connection_error")
	subscriptionsTransportWSMessageTypeConnectionTerminate = subscriptionsTransportWSMessageType("connection_terminate")
	subscriptionsTransportWSMessageTypeKeepAlive           = subscriptionsTransportWSMessageType("ka")
	subscriptionsTransportWSMessageTypeStart               = subscriptionsTransportWSMessageType("start")
	subscriptionsTransportWSMessageTypeStop                = subscriptionsTransportWSMessageType("stop")
	subscriptionsTransportWSMessageTypeData                = subscriptionsTransportWSMessageType("data")
	subscriptionsTransportWSMessageTypeError               = subscriptionsTransportWSMessageType("error")
	subscriptionsTransportWSMessageTypeComplete            = subscriptionsTransportWSMessageType("complete")

	SubscriptionsTransportWSSubprotocol = "graphql-ws"
)

var _ Protocol = (*subscriptionsTransportWSProtocol)(nil)

type subscriptionsTransportWSMessage struct {
	ID         string                              `json:"id,omitempty"`
	Type       subscriptionsTransportWSMessageType `json:"type"`
	Payload    json.RawMessage                     `json:"payload,omitempty"`
	Extensions json.RawMessage                     `json:"extensions,omitempty"`
}

type subscriptionsTransportWSProtocol struct {
	conn ProtocolConn
}

func newSubscriptionsTransportWSProtocol(conn ProtocolConn) *subscriptionsTransportWSProtocol {
	return &subscriptionsTransportWSProtocol{
		conn: conn,
	}
}

func (p *subscriptionsTransportWSProtocol) Subprotocol() string {
	return SubscriptionsTransportWSSubprotocol
}

func (p *subscriptionsTransportWSProtocol) Initialize() (json.RawMessage, error) {
	var msg subscriptionsTransportWSMessage
	if err := p.conn.ReadJSON(&msg); err != nil {
		return nil, fmt.Errorf("failed to read connection init message: %w", err)
	}

	if msg.Type != subscriptionsTransportWSMessageTypeConnectionInit {
		return nil, fmt.Errorf("first message should be %s, got %s", subscriptionsTransportWSMessageTypeConnectionInit, msg.Type)
	}

	if err := p.conn.WriteJSON(subscriptionsTransportWSMessage{
		Type: subscriptionsTransportWSMessageTypeConnectionAck,
	}); err != nil {
		return nil, fmt.Errorf("failed to write connection ack message: %w", err)
	}

	return msg.Payload, nil
}

func (p *subscriptionsTransportWSProtocol) ReadMessage() (*Message, error) {
	var msg subscriptionsTransportWSMessage
	if err := p.conn.ReadJSON(&msg); err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var messageType MessageType
	switch msg.Type {
	case subscriptionsTransportWSMessageTypeConnectionTerminate:
		messageType = MessageTypeTerminate
	case subscriptionsTransportWSMessageTypeStart:
		messageType = MessageTypeSubscribe
	case subscriptionsTransportWSMessageTypeStop:
		messageType = MessageTypeComplete
	default:
		return nil, fmt.Errorf("unknown message type: %s", msg.Type)
	}

	return &Message{
		ID:      msg.ID,
		Type:    messageType,
		Payload: msg.Payload,
	}, nil
}

func (p *subscriptionsTransportWSProtocol) Pong(msg *Message) error {
	return p.conn.WriteJSON(subscriptionsTransportWSMessage{
		ID:      msg.ID,
		Type:    subscriptionsTransportWSMessageTypeKeepAlive,
		Payload: msg.Payload,
	})
}

func (p *subscriptionsTransportWSProtocol) WriteGraphQLData(id string, data json.RawMessage, extensions json.RawMessage) error {
	return p.conn.WriteJSON(subscriptionsTransportWSMessage{
		ID:         id,
		Type:       subscriptionsTransportWSMessageTypeData,
		Payload:    data,
		Extensions: extensions,
	})
}

func (p *subscriptionsTransportWSProtocol) WriteGraphQLErrors(id string, errors json.RawMessage, extensions json.RawMessage) error {
	// This protocol has errors inside an object, so we need to wrap it
	data, err := sjson.SetBytes([]byte(`{}`), "errors", errors)
	if err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return p.conn.WriteJSON(subscriptionsTransportWSMessage{
		ID:         id,
		Type:       subscriptionsTransportWSMessageTypeData,
		Payload:    data,
		Extensions: extensions,
	})
}

func (p *subscriptionsTransportWSProtocol) Close(code ws.StatusCode, reason string) error {
	if err := p.conn.WriteCloseFrame(code, reason); err != nil {
		return err
	}

	return nil
}

func (p *subscriptionsTransportWSProtocol) Complete(id string) error {
	return p.conn.WriteJSON(subscriptionsTransportWSMessage{
		ID:   id,
		Type: subscriptionsTransportWSMessageTypeComplete,
	})
}
