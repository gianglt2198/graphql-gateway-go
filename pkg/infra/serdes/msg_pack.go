package serdes

import (
	"bytes"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

type msgPackSerializer struct {
	encoder *msgpack.Encoder
	decoder *msgpack.Decoder
	sync.Mutex
}

var _ Serializer = (*msgPackSerializer)(nil)

func NewMsgPack() Serializer {
	pack := &msgPackSerializer{
		encoder: msgpack.NewEncoder(nil),
		decoder: msgpack.NewDecoder(nil),
	}
	pack.encoder.SetCustomStructTag("json")
	pack.decoder.SetCustomStructTag("json")
	return pack
}

func (m *msgPackSerializer) Encode(data any) ([]byte, error) {
	m.Lock()
	defer m.Unlock()
	var buf bytes.Buffer
	m.encoder.ResetWriter(&buf)
	err := m.encoder.Encode(data)
	return buf.Bytes(), err
}

func (m *msgPackSerializer) Decode(data []byte, result any) error {
	m.Lock()
	defer m.Unlock()
	m.decoder.ResetReader(bytes.NewReader([]byte(data)))
	return m.decoder.Decode(&result)
}
