package serdes

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

type GzipSerializer struct {
	writer *gzip.Writer
	reader *gzip.Reader
	mu     sync.Mutex
}

var _ Serializer = (*GzipSerializer)(nil)

func NewGzipSerializer() Serializer {
	gzipSeq := &GzipSerializer{}

	gzipSeq.writer = gzip.NewWriter(nil)
	gzipSeq.reader, _ = gzip.NewReader(nil)

	return gzipSeq
}

func (m *GzipSerializer) Encode(data any) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var serialized bytes.Buffer
	var err error

	encoder := json.NewEncoder(&serialized)
	err = encoder.Encode(data)

	if err != nil {
		return nil, fmt.Errorf("serialization error: %w", err)
	}

	var compressed bytes.Buffer

	m.writer.Reset(&compressed)

	_, err = m.writer.Write(serialized.Bytes())
	if err != nil {
		return nil, fmt.Errorf("compression error: %w", err)
	}

	if err := m.writer.Close(); err != nil {
		return nil, fmt.Errorf("gzip close error: %w", err)
	}

	return compressed.Bytes(), nil
}

func (m *GzipSerializer) Decode(in []byte, out any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	compressed := bytes.NewReader(in)

	if err := m.reader.Reset(compressed); err != nil {
		return fmt.Errorf("gzip reader reset error: %w", err)
	}
	defer m.reader.Close()

	decompressed, err := io.ReadAll(m.reader)
	if err != nil {
		return fmt.Errorf("decompression error: %w", err)
	}

	reader := bytes.NewReader(decompressed)

	decoder := json.NewDecoder(reader)
	return decoder.Decode(out)
}
