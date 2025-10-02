package main

import (
	"fmt"
	"net/http"
	"time"
)

// GET /api/v1/health - simple liveness probe
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// GET /api/v1/health/stream - SSE stream for connectivity monitoring
func healthStreamHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // for nginx

	// Support flush
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("streaming unsupported"))
		return
	}

	// Initial comment to open the stream
	_, _ = fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	// Heartbeats every 5s
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Close when client goes away
	ctx := r.Context()
	for {
		select {
		case t := <-ticker.C:
			// Send an SSE event with current time
			_, _ = fmt.Fprintf(w, "event: ping\n")
			_, _ = fmt.Fprintf(w, "data: {\"ts\": %d}\n\n", t.Unix())
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}
