package controllers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/timendus/pixelbox/server"
)

var Events *SSE

func init() {
	Events = NewSSEHub()
	router := http.NewServeMux()
	router.HandleFunc("GET /", Events.Handler)
	server.RegisterRouter("/events", router)

}

type SSE struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func NewSSEHub() *SSE {
	return &SSE{clients: make(map[chan string]struct{})}
}

func (h *SSE) Subscribe() chan string {
	ch := make(chan string, 16) // buffered so slow clients don't immediately block broadcasts
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *SSE) Unsubscribe(ch chan string) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

func (h *SSE) Broadcast(msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
			// Drop if the client is too slow (prevents one client from blocking everyone)
		}
	}
}

// GET /events
func (h *SSE) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	// Initial line to establish stream
	_, _ = w.Write([]byte(": ok\n\n"))
	flusher.Flush()

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-keepAlive.C:
			_, _ = w.Write([]byte(": ping\n\n"))
			flusher.Flush()

		case msg := <-ch:
			// Basic SSE format: "data: <line>\n\n"
			// If msg contains newlines, you must prefix each line with "data: ".
			writeSSEData(w, msg)
			flusher.Flush()
		}
	}
}

func writeSSEData(w io.Writer, data string) {
	// SSE requires each line to start with "data: "
	// (this makes multiline payloads valid)
	for _, line := range strings.Split(data, "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
}
