package telemetry

import (
	"io"
	"os"
)

// HubWriter is a custom io.Writer that writes to standard output AND broadcasts to the WebSocket Hub.
// This is the absolute cleanest way to intercept slog JSON output without modifying any worker logic!
type HubWriter struct {
	stdout io.Writer
	hub    *Hub
}

func NewHubWriter(hub *Hub) *HubWriter {
	return &HubWriter{
		stdout: os.Stdout,
		hub:    hub,
	}
}

func (w *HubWriter) Write(p []byte) (n int, err error) {
	// 1. Write the log to the actual terminal
	n, err = w.stdout.Write(p)
	
	// 2. Copy the bytes (since slog might reuse the buffer internally)
	msg := make([]byte, len(p))
	copy(msg, p)
	
	// 3. Broadcast the exact JSON string to the WebSockets!
	// We use a non-blocking select to ensure that if the WebSocket hub is overloaded,
	// we just drop the UI log. We NEVER want the backend worker to slow down just because a browser is slow!
	select {
	case w.hub.broadcast <- msg:
	default:
	}
	
	return n, err
}
