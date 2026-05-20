package api

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MonsteRico/immich-frame/internal/cache"
	"github.com/MonsteRico/immich-frame/internal/config"
	"github.com/MonsteRico/immich-frame/internal/playback"
)

//go:embed static
var embeddedStatic embed.FS

type Server struct {
	Config    config.Config
	Cache     *cache.Store
	Queue     *playback.Queue
	Hub       *Hub
	FrameDist string
	SetupDist string
}

type FrameState struct {
	playback.State
	Overlays config.OverlayConfig `json:"overlays"`
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.redirectRoot)
	mux.HandleFunc("/frame", s.frame)
	mux.HandleFunc("/setup", s.setup)
	mux.HandleFunc("/assets/", s.assets)
	mux.HandleFunc("/api/state", s.state)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/display", s.display)
	mux.HandleFunc("/api/playback/next", s.playbackCommand("next"))
	mux.HandleFunc("/api/playback/previous", s.playbackCommand("previous"))
	mux.HandleFunc("/api/playback/pause", s.playbackCommand("pause"))
	mux.HandleFunc("/api/playback/resume", s.playbackCommand("resume"))
	mux.HandleFunc("/media/", s.media)
	return mux
}

func (s *Server) PublishState() {
	s.Hub.Publish(s.snapshot())
}

func (s *Server) snapshot() FrameState {
	return FrameState{State: s.Queue.State(), Overlays: s.Config.Overlays}
}

func (s *Server) redirectRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/frame", http.StatusFound)
}

func (s *Server) frame(w http.ResponseWriter, r *http.Request) {
	s.serveIndex(w, r, s.FrameDist, "static/frame/index.html")
}

func (s *Server) setup(w http.ResponseWriter, r *http.Request) {
	s.serveIndex(w, r, s.SetupDist, "static/setup/index.html")
}

func (s *Server) assets(w http.ResponseWriter, r *http.Request) {
	assetPath := strings.TrimPrefix(r.URL.Path, "/assets/")
	if s.serveDistAsset(w, r, s.FrameDist, assetPath) ||
		s.serveDistAsset(w, r, s.SetupDist, assetPath) ||
		s.serveEmbeddedAsset(w, r, "static/frame/assets", assetPath) ||
		s.serveEmbeddedAsset(w, r, "static/setup/assets", assetPath) {
		return
	}
	http.NotFound(w, r)
}

func (s *Server) state(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.snapshot())
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	ch, unsubscribe := s.Hub.Subscribe()
	defer unsubscribe()
	writeSSE(w, s.snapshot())
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case state := <-ch:
			writeSSE(w, state)
			flusher.Flush()
		}
	}
}

func (s *Server) display(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) playbackCommand(command string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var err error
		switch command {
		case "next":
			_, err = s.Queue.Next()
		case "previous":
			_, err = s.Queue.Previous()
		case "pause":
			s.Queue.Pause()
		case "resume":
			s.Queue.Resume()
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		s.PublishState()
		writeJSON(w, s.snapshot())
	}
}

func (s *Server) media(w http.ResponseWriter, r *http.Request) {
	if !isLoopbackRequest(r) {
		http.Error(w, "media access requires localhost in this build", http.StatusForbidden)
		return
	}
	assetID := strings.TrimPrefix(r.URL.Path, "/media/")
	entry, ok := s.Cache.Get(assetID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", entry.MediaType)
	http.ServeFile(w, r, entry.CachePath)
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request, distDir, embeddedPath string) {
	if distDir != "" {
		index := filepath.Join(distDir, "index.html")
		if _, err := os.Stat(index); err == nil {
			http.ServeFile(w, r, index)
			return
		}
	}
	data, err := embeddedStatic.ReadFile(embeddedPath)
	if err != nil {
		http.Error(w, "embedded UI missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (s *Server) serveDistAsset(w http.ResponseWriter, r *http.Request, distDir, path string) bool {
	if distDir == "" {
		return false
	}
	clean := filepath.Clean(path)
	if strings.HasPrefix(clean, "..") {
		return false
	}
	full := filepath.Join(distDir, "assets", clean)
	if _, err := os.Stat(full); err != nil {
		return false
	}
	http.ServeFile(w, r, full)
	return true
}

func (s *Server) serveEmbeddedAsset(w http.ResponseWriter, r *http.Request, root, assetPath string) bool {
	clean := cleanAssetPath(assetPath)
	if clean == "" {
		return false
	}
	embeddedPath := path.Join(root, clean)
	data, err := embeddedStatic.ReadFile(embeddedPath)
	if err != nil {
		return false
	}
	if contentType := mime.TypeByExtension(path.Ext(clean)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	http.ServeContent(w, r, path.Base(clean), time.Time{}, bytes.NewReader(data))
	return true
}

func cleanAssetPath(assetPath string) string {
	clean := path.Clean(strings.TrimPrefix(strings.ReplaceAll(assetPath, "\\", "/"), "/"))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return ""
	}
	return clean
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func writeSSE(w http.ResponseWriter, value any) {
	data, _ := json.Marshal(value)
	_, _ = fmt.Fprintf(w, "event: state\ndata: %s\n\n", data)
}

func isLoopbackRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip == nil || ip.IsLoopback()
}

type Hub struct {
	mu          sync.Mutex
	subscribers map[chan FrameState]struct{}
}

func NewHub() *Hub {
	return &Hub{subscribers: map[chan FrameState]struct{}{}}
}

func (h *Hub) Subscribe() (chan FrameState, func()) {
	ch := make(chan FrameState, 4)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		delete(h.subscribers, ch)
		close(ch)
		h.mu.Unlock()
	}
}

func (h *Hub) Publish(state FrameState) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subscribers {
		select {
		case ch <- state:
		default:
		}
	}
}
