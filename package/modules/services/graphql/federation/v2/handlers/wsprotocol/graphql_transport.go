package wsprotocol

import (
	"encoding/json"
	"fmt"

	"github.com/gobwas/ws"
	"github.com/tidwall/sjson"
)

type graphQLWSMessageType string

const (
	graphQLWSMessageTypeConnectionInit graphQLWSMessageType = "connection_init"
	graphQLWSMessageTypeConnectionAck  graphQLWSMessageType = "connection_ack"
	graphQLWSMessageTypePing           graphQLWSMessageType = "ping"
	graphQLWSMessageTypePong           graphQLWSMessageType = "pong"
	graphQLWSMessageTypeSubscribe      graphQLWSMessageType = "subscribe"
	graphQLWSMessageTypeNext           graphQLWSMessageType = "next"
	graphQLWSMessageTypeError          graphQLWSMessageType = "error"
	graphQLWSMessageTypeComplete       graphQLWSMessageType = "complete"

	SubscriptionsGraphQLWSSubprotocol = "graphql-transport-ws"
)

var _ Protocol = (*graphqlTransportWSProtocol)(nil)

type graphqlTransportWSMessage struct {
	ID         string               `json:"id,omitempty"`
	Type       graphQLWSMessageType `json:"type"`
	Payload    json.RawMessage      `json:"payload,omitempty"`
	Extensions json.RawMessage      `json:"extensions,omitempty"`
}

type graphqlTransportWSProtocol struct {
	conn ProtocolConn
}

func newGraphQLWSProtocol(conn ProtocolConn) *graphqlTransportWSProtocol {
	return &graphqlTransportWSProtocol{
		conn: conn,
	}
}

func (p *graphqlTransportWSProtocol) Subprotocol() string {
	return SubscriptionsTransportWSSubprotocol
}

func (p *graphqlTransportWSProtocol) Initialize() (json.RawMessage, error) {
	var msg graphqlTransportWSMessage
	if err := p.conn.ReadJSON(&msg); err != nil {
		return nil, fmt.Errorf("failed to read connection init message: %w", err)
	}

	if msg.Type != graphQLWSMessageTypeConnectionInit {
		return nil, fmt.Errorf("first message should be %s, got %s", graphQLWSMessageTypeConnectionInit, msg.Type)
	}

	if err := p.conn.WriteJSON(graphqlTransportWSMessage{
		Type: graphQLWSMessageTypeConnectionAck,
	}); err != nil {
		return nil, fmt.Errorf("failed to write connection ack message: %w", err)
	}

	return msg.Payload, nil
}

func (p *graphqlTransportWSProtocol) ReadMessage() (*Message, error) {
	var msg graphqlTransportWSMessage
	if err := p.conn.ReadJSON(&msg); err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var messageType MessageType
	switch msg.Type {
	case graphQLWSMessageTypePing:
		messageType = MessageTypePing
	case graphQLWSMessageTypePong:
		messageType = MessageTypePong
	case graphQLWSMessageTypeSubscribe:
		messageType = MessageTypeSubscribe
	case graphQLWSMessageTypeComplete:
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

func (p *graphqlTransportWSProtocol) Pong(msg *Message) error {
	return p.conn.WriteJSON(graphqlTransportWSMessage{
		ID:      msg.ID,
		Type:    graphQLWSMessageTypePong,
		Payload: msg.Payload,
	})
}

func (p *graphqlTransportWSProtocol) WriteGraphQLData(id string, data json.RawMessage, extensions json.RawMessage) error {
	return p.conn.WriteJSON(graphqlTransportWSMessage{
		ID:         id,
		Type:       graphQLWSMessageTypeNext,
		Payload:    data,
		Extensions: extensions,
	})
}

func (p *graphqlTransportWSProtocol) WriteGraphQLErrors(id string, errors json.RawMessage, extensions json.RawMessage) error {
	// This protocol has errors inside an object, so we need to wrap it
	data, err := sjson.SetBytes([]byte(`{}`), "errors", errors)
	if err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return p.conn.WriteJSON(graphqlTransportWSMessage{
		ID:         id,
		Type:       graphQLWSMessageTypeError,
		Payload:    data,
		Extensions: extensions,
	})
}

func (p *graphqlTransportWSProtocol) Close(code ws.StatusCode, reason string) error {
	if err := p.conn.WriteCloseFrame(code, reason); err != nil {
		return err
	}

	return nil
}

func (p *graphqlTransportWSProtocol) Complete(id string) error {
	return p.conn.WriteJSON(graphqlTransportWSMessage{
		ID:   id,
		Type: graphQLWSMessageTypeComplete,
	})
}
