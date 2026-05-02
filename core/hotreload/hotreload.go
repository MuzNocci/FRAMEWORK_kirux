package hotreload

import (
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

type Hub struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[chan struct{}]struct{})}
}

func (h *Hub) add(ch chan struct{}) {
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) remove(ch chan struct{}) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

func (h *Hub) broadcast() {
	h.mu.Lock()
	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	h.mu.Unlock()
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan struct{}, 1)
	h.add(ch)
	defer h.remove(ch)

	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			fmt.Fprintf(w, "data: reload\n\n")
			flusher.Flush()
		}
	}
}

func (h *Hub) Watch(dirs ...string) {
	go func() {
		mtimes := snapshot(dirs)

		for range time.Tick(300 * time.Millisecond) {
			current := snapshot(dirs)
			if changed(mtimes, current) {
				h.broadcast()
			}
			mtimes = current
		}
	}()
}

func snapshot(dirs []string) map[string]time.Time {
	m := map[string]time.Time{}
	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			switch filepath.Ext(path) {
			case ".html", ".css", ".js":
				if info, err := d.Info(); err == nil {
					m[path] = info.ModTime()
				}
			}
			return nil
		})
	}
	return m
}

func changed(prev, curr map[string]time.Time) bool {
	for path, t := range curr {
		if p, ok := prev[path]; !ok || !t.Equal(p) {
			return true
		}
	}
	for path := range prev {
		if _, ok := curr[path]; !ok {
			return true
		}
	}
	return false
}
