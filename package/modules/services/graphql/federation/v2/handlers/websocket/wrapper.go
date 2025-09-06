package fwebsocket

import (
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gofiber/contrib/websocket"
)

// wsConnectionWrapper is a wrapper around websocket.Conn that allows
// writing from multiple goroutines
type wsConnectionWrapper struct {
	conn         *websocket.Conn
	mu           sync.Mutex
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func newWSConnectionWrapper(conn *websocket.Conn, readTimeout, writeTimeout time.Duration) *wsConnectionWrapper {
	return &wsConnectionWrapper{
		conn:         conn,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

func (w *wsConnectionWrapper) ReadJSON(v any) error {
	return w.conn.ReadJSON(v)
}

func (w *wsConnectionWrapper) WriteText(text string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writeTimeout > 0 {
		err := w.conn.SetWriteDeadline(time.Now().Add(w.writeTimeout))
		if err != nil {
			return err
		}
	}

	return w.conn.WriteMessage(websocket.TextMessage, []byte(text))
}

func (w *wsConnectionWrapper) WriteJSON(v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.conn.WriteJSON(v)
}

func (w *wsConnectionWrapper) WriteCloseFrame(code ws.StatusCode, reason string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.conn.CloseHandler()(int(code), reason)
}

func (w *wsConnectionWrapper) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.Close()
}

func (w *wsConnectionWrapper) Read() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.LocalAddr().String()
}
