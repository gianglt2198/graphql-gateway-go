package fwebsocket

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// wsConnectionWrapper is a wrapper around websocket.Conn that allows
// writing from multiple goroutines
type wsConnectionWrapper struct {
	conn         net.Conn
	mu           sync.Mutex
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func newWSConnectionWrapper(conn net.Conn, readTimeout, writeTimeout time.Duration) *wsConnectionWrapper {
	return &wsConnectionWrapper{
		conn:         conn,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

func (w *wsConnectionWrapper) ReadJSON(v any) error {
	if w.readTimeout > 0 {
		err := w.conn.SetReadDeadline(time.Now().Add(w.readTimeout))
		if err != nil {
			return err
		}
	}

	text, err := wsutil.ReadClientText(w.conn)
	if err != nil {
		return err
	}

	return json.Unmarshal(text, v)
}

func (c *wsConnectionWrapper) WriteText(text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.writeTimeout > 0 {
		err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
		if err != nil {
			return err
		}
	}

	return wsutil.WriteServerText(c.conn, []byte(text))
}

func (w *wsConnectionWrapper) WriteJSON(v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if w.writeTimeout > 0 {
		err := w.conn.SetWriteDeadline(time.Now().Add(w.writeTimeout))
		if err != nil {
			return err
		}
	}

	return wsutil.WriteServerText(w.conn, data)
}

func (w *wsConnectionWrapper) WriteCloseFrame(code ws.StatusCode, reason string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writeTimeout > 0 {
		err := w.conn.SetWriteDeadline(time.Now().Add(w.writeTimeout))
		if err != nil {
			return err
		}
	}

	return ws.WriteFrame(w.conn, ws.NewCloseFrame(ws.NewCloseFrameBody(code, reason)))
}

func (w *wsConnectionWrapper) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.Close()
}
