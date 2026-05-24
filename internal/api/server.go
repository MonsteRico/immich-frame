package api

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
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

	"github.com/MonsteRico/immich-frame/internal/auth"
	"github.com/MonsteRico/immich-frame/internal/cache"
	"github.com/MonsteRico/immich-frame/internal/config"
	"github.com/MonsteRico/immich-frame/internal/immich"
	"github.com/MonsteRico/immich-frame/internal/playback"
	setupstate "github.com/MonsteRico/immich-frame/internal/setup"
)

//go:embed static
var embeddedStatic embed.FS

type Server struct {
	mu               sync.Mutex
	Config           config.Config
	Secrets          config.Secrets
	State            config.State
	Paths            config.Paths
	Cache            *cache.Store
	Queue            *playback.Queue
	Hub              *Hub
	Setup            *setupstate.Manager
	Sessions         *auth.Manager
	ImmichHTTPClient *http.Client
	FrameDist        string
	SetupDist        string
	OnSetupComplete  func()
}

type FrameState struct {
	playback.State
	Overlays config.OverlayConfig `json:"overlays"`
	Setup    SetupPublicState     `json:"setup"`
}

type SetupPublicState struct {
	Configured          bool   `json:"configured"`
	Status              string `json:"status"`
	SetupCodeRequired   bool   `json:"setupCodeRequired"`
	SetupCode           string `json:"setupCode,omitempty"`
	SetupURL            string `json:"setupUrl"`
	Hostname            string `json:"hostname"`
	IPAddress           string `json:"ipAddress,omitempty"`
	AdminPasswordExists bool   `json:"adminPasswordExists"`
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
	mux.HandleFunc("/api/setup/state", s.setupState)
	mux.HandleFunc("/api/setup/claim", s.setupClaim)
	mux.HandleFunc("/api/setup/admin-password", s.setupAdminPassword)
	mux.HandleFunc("/api/setup/complete", s.setupComplete)
	mux.HandleFunc("/api/auth/session", s.authSession)
	mux.HandleFunc("/api/auth/login", s.authLogin)
	mux.HandleFunc("/api/auth/logout", s.authLogout)
	mux.HandleFunc("/api/status", s.status)
	mux.HandleFunc("/api/settings", s.settings)
	mux.HandleFunc("/api/immich/test", s.immichTest)
	mux.HandleFunc("/api/immich/albums", s.immichAlbums)
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

func (s *Server) RecordSyncStatus(lastError string, syncedAt time.Time) {
	s.mu.Lock()
	s.State.LastError = lastError
	if !syncedAt.IsZero() {
		s.State.LastSync = syncedAt
	}
	_ = config.SaveState(s.Paths.StateFile, s.State)
	s.mu.Unlock()
}

func (s *Server) RuntimeInputs() (config.Config, config.Secrets, config.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Config, s.Secrets, s.State
}

func (s *Server) snapshot() FrameState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return FrameState{State: s.Queue.State(), Overlays: s.Config.Overlays, Setup: s.setupPublicState(true)}
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
	writeJSON(w, sanitizeFrameState(s.snapshot(), isLoopbackRequest(r)))
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
	includeCode := isLoopbackRequest(r)
	writeSSE(w, sanitizeFrameState(s.snapshot(), includeCode))
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case state := <-ch:
			writeSSE(w, sanitizeFrameState(state, includeCode))
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
	if !isLoopbackRequest(r) && !s.hasSession(r, auth.ScopeAdmin) {
		http.Error(w, "media access requires admin authentication", http.StatusForbidden)
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

func (s *Server) setupState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	state := s.setupPublicState(isLoopbackRequest(r))
	s.mu.Unlock()
	writeJSON(w, state)
}

func (s *Server) setupClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	state, ok, err := s.Setup.Claim(strings.TrimSpace(req.Code))
	if err != nil {
		http.Error(w, "setup state unavailable", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "setup code was not accepted", http.StatusUnauthorized)
		return
	}
	session, err := s.Sessions.Create(auth.ScopeSetup)
	if err != nil {
		http.Error(w, "session unavailable", http.StatusInternalServerError)
		return
	}
	s.setSessionCookie(w, session)
	s.mu.Lock()
	s.State = state
	response := s.setupPublicState(false)
	s.mu.Unlock()
	writeJSON(w, response)
}

func (s *Server) setupAdminPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.hasSession(r, auth.ScopeSetup, auth.ScopeAdmin) {
		http.Error(w, "setup authorization required", http.StatusUnauthorized)
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.Secrets.AdminPasswordHash = hash
	err = config.SaveSecrets(s.Paths.SecretsFile, s.Secrets)
	s.mu.Unlock()
	if err != nil {
		http.Error(w, "secrets unavailable", http.StatusInternalServerError)
		return
	}
	session, err := s.Sessions.Create(auth.ScopeAdmin)
	if err != nil {
		http.Error(w, "session unavailable", http.StatusInternalServerError)
		return
	}
	s.setSessionCookie(w, session)
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) setupComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.hasSession(r, auth.ScopeSetup, auth.ScopeAdmin) {
		http.Error(w, "setup authorization required", http.StatusUnauthorized)
		return
	}
	s.mu.Lock()
	ready, reason := s.setupReadyLocked()
	s.mu.Unlock()
	if !ready {
		http.Error(w, reason, http.StatusConflict)
		return
	}
	state, err := s.Setup.Complete()
	if err != nil {
		http.Error(w, "setup state unavailable", http.StatusInternalServerError)
		return
	}
	s.mu.Lock()
	s.State = state
	response := s.setupPublicState(false)
	s.mu.Unlock()
	s.PublishState()
	if s.OnSetupComplete != nil {
		s.OnSetupComplete()
	}
	writeJSON(w, response)
}

func (s *Server) authSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, admin := s.session(r, auth.ScopeAdmin)
	_, setup := s.session(r, auth.ScopeSetup)
	writeJSON(w, map[string]bool{"admin": admin, "setup": setup})
}

func (s *Server) authLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.mu.Lock()
	hash := s.Secrets.AdminPasswordHash
	s.mu.Unlock()
	if hash == "" || !auth.VerifyPassword(hash, req.Password) {
		http.Error(w, "invalid admin password", http.StatusUnauthorized)
		return
	}
	session, err := s.Sessions.Create(auth.ScopeAdmin)
	if err != nil {
		http.Error(w, "session unavailable", http.StatusInternalServerError)
		return
	}
	s.setSessionCookie(w, session)
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) authLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if cookie, err := r.Cookie("immich_frame_session"); err == nil {
		s.Sessions.Destroy(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "immich_frame_session", Value: "", Path: "/", MaxAge: -1, SameSite: http.SameSiteLaxMode})
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.hasSession(r, auth.ScopeSetup, auth.ScopeAdmin) {
		http.Error(w, "admin authentication required", http.StatusUnauthorized)
		return
	}
	s.mu.Lock()
	response := s.portalStatusLocked()
	s.mu.Unlock()
	writeJSON(w, response)
}

func (s *Server) settings(w http.ResponseWriter, r *http.Request) {
	if !s.hasSession(r, auth.ScopeSetup, auth.ScopeAdmin) {
		http.Error(w, "admin authentication required", http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.mu.Lock()
		response := settingsResponse{Config: s.Config, HasImmichAPIKey: s.Secrets.ImmichAPIKey != "", Status: s.portalStatusLocked()}
		s.mu.Unlock()
		writeJSON(w, response)
	case http.MethodPut:
		var req settingsRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		s.mu.Lock()
		cfg := s.Config
		secrets := s.Secrets
		applySettings(&cfg, &secrets, req)
		if err := cfg.Validate(); err != nil {
			s.mu.Unlock()
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		nextValidation := s.State.ImmichValidation
		if !nextValidation.Matches(cfg.Immich.URL, secrets.ImmichAPIKey) {
			nextValidation = config.ImmichValidationState{}
		}
		if err := config.Save(s.Paths.ConfigFile, cfg); err != nil {
			s.mu.Unlock()
			http.Error(w, "config unavailable", http.StatusInternalServerError)
			return
		}
		if req.ImmichAPIKey != nil {
			if err := config.SaveSecrets(s.Paths.SecretsFile, secrets); err != nil {
				s.mu.Unlock()
				http.Error(w, "secrets unavailable", http.StatusInternalServerError)
				return
			}
		}
		if nextValidation != s.State.ImmichValidation {
			s.State.ImmichValidation = nextValidation
			if err := config.SaveState(s.Paths.StateFile, s.State); err != nil {
				s.mu.Unlock()
				http.Error(w, "state unavailable", http.StatusInternalServerError)
				return
			}
		}
		s.Config = cfg
		s.Secrets = secrets
		response := settingsResponse{Config: s.Config, HasImmichAPIKey: s.Secrets.ImmichAPIKey != "", Status: s.portalStatusLocked()}
		s.mu.Unlock()
		writeJSON(w, response)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) immichTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.hasSession(r, auth.ScopeSetup, auth.ScopeAdmin) {
		http.Error(w, "setup or admin authorization required", http.StatusUnauthorized)
		return
	}
	var req struct {
		URL    string `json:"url"`
		APIKey string `json:"apiKey"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.mu.Lock()
	if strings.TrimSpace(req.URL) == "" {
		req.URL = s.Config.Immich.URL
	}
	if strings.TrimSpace(req.APIKey) == "" {
		req.APIKey = s.Secrets.ImmichAPIKey
	}
	s.mu.Unlock()
	client, err := immich.NewClient(req.URL, req.APIKey, immich.WithHTTPClient(s.immichHTTPClient()))
	if err != nil {
		writeImmichError(w, err)
		return
	}
	info, err := client.TestConnection(r.Context())
	if err != nil {
		writeImmichError(w, err)
		return
	}
	s.mu.Lock()
	s.State.ImmichValidation = config.NewImmichValidation(req.URL, req.APIKey, info.Version, info.KeyName, time.Now())
	err = config.SaveState(s.Paths.StateFile, s.State)
	status := s.portalStatusLocked()
	s.mu.Unlock()
	if err != nil {
		http.Error(w, "state unavailable", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "version": info.Version, "keyName": info.KeyName, "status": status})
}

func (s *Server) immichAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.hasSession(r, auth.ScopeSetup, auth.ScopeAdmin) {
		http.Error(w, "setup or admin authorization required", http.StatusUnauthorized)
		return
	}
	s.mu.Lock()
	url := s.Config.Immich.URL
	apiKey := s.Secrets.ImmichAPIKey
	s.mu.Unlock()
	client, err := immich.NewClient(url, apiKey, immich.WithHTTPClient(s.immichHTTPClient()))
	if err != nil {
		writeImmichError(w, err)
		return
	}
	albums, err := client.ListAlbums(r.Context())
	if err != nil {
		writeImmichError(w, err)
		return
	}
	writeJSON(w, map[string]any{"albums": albums})
}

type settingsResponse struct {
	Config          config.Config `json:"config"`
	HasImmichAPIKey bool          `json:"hasImmichApiKey"`
	Status          PortalStatus  `json:"status"`
}

type settingsRequest struct {
	Config       config.Config `json:"config"`
	ImmichAPIKey *string       `json:"immichApiKey,omitempty"`
}

func applySettings(cfg *config.Config, secrets *config.Secrets, req settingsRequest) {
	next := req.Config
	next.Server = cfg.Server
	next.Source.LocalFolder = cfg.Source.LocalFolder
	next.Filters = cfg.Filters
	next.Sync = cfg.Sync
	next.Weather = cfg.Weather
	*cfg = next
	if req.ImmichAPIKey != nil {
		secrets.ImmichAPIKey = strings.TrimSpace(*req.ImmichAPIKey)
	}
}

type PortalStatus struct {
	Setup           SetupPublicState `json:"setup"`
	Configured      bool             `json:"configured"`
	HasImmichAPIKey bool             `json:"hasImmichApiKey"`
	Immich          ImmichStatus     `json:"immich"`
	SourceMode      string           `json:"sourceMode"`
	CacheCount      int              `json:"cacheCount"`
	LastError       string           `json:"lastError,omitempty"`
}

type ImmichStatus struct {
	URL                string     `json:"url,omitempty"`
	Configured         bool       `json:"configured"`
	Validated          bool       `json:"validated"`
	ValidationRequired bool       `json:"validationRequired"`
	ValidatedAt        *time.Time `json:"validatedAt,omitempty"`
	Version            string     `json:"version,omitempty"`
	KeyName            string     `json:"keyName,omitempty"`
}

func (s *Server) portalStatusLocked() PortalStatus {
	cacheCount := 0
	if s.Cache != nil {
		cacheCount = len(s.Cache.List())
	}
	validated := s.State.ImmichValidation.Matches(s.Config.Immich.URL, s.Secrets.ImmichAPIKey)
	immichStatus := ImmichStatus{
		URL:                s.Config.Immich.URL,
		Configured:         strings.TrimSpace(s.Config.Immich.URL) != "" && s.Secrets.ImmichAPIKey != "",
		Validated:          validated,
		ValidationRequired: !validated,
	}
	if validated {
		immichStatus.ValidatedAt = &s.State.ImmichValidation.ValidatedAt
		immichStatus.Version = s.State.ImmichValidation.Version
		immichStatus.KeyName = s.State.ImmichValidation.KeyName
	}
	return PortalStatus{
		Setup:           s.setupPublicState(false),
		Configured:      s.State.SetupComplete,
		HasImmichAPIKey: s.Secrets.ImmichAPIKey != "",
		Immich:          immichStatus,
		SourceMode:      s.Config.Source.Mode,
		CacheCount:      cacheCount,
		LastError:       s.State.LastError,
	}
}

func (s *Server) setupReadyLocked() (bool, string) {
	if s.Secrets.AdminPasswordHash == "" {
		return false, "create the local admin password before finishing setup"
	}
	if strings.TrimSpace(s.Config.Immich.URL) == "" {
		return false, "enter and save the Immich URL before finishing setup"
	}
	if strings.TrimSpace(s.Secrets.ImmichAPIKey) == "" {
		return false, "enter and save a dedicated Immich API key before finishing setup"
	}
	if !s.State.ImmichValidation.Matches(s.Config.Immich.URL, s.Secrets.ImmichAPIKey) {
		return false, "test the saved Immich URL and API key successfully before finishing setup"
	}
	switch s.Config.Source.Mode {
	case "random":
		return true, ""
	case "album":
		if strings.TrimSpace(s.Config.Source.Album.ID) != "" {
			return true, ""
		}
		return false, "choose an Immich album or random library mode before finishing setup"
	default:
		return false, "choose an Immich album or random library mode before finishing setup"
	}
}

func (s *Server) setupPublicState(includeCode bool) SetupPublicState {
	status := s.State.SetupStatus
	if status == "" {
		if s.State.SetupComplete {
			status = string(setupstate.StatusConfigured)
		} else {
			status = string(setupstate.StatusSetupCodeRequired)
		}
	}
	setupURL := fmt.Sprintf("http://%s:%d/setup", s.Config.Server.Hostname, s.Config.Server.Port)
	state := SetupPublicState{
		Configured:          s.State.SetupComplete,
		Status:              status,
		SetupCodeRequired:   !s.State.SetupComplete,
		SetupURL:            setupURL,
		Hostname:            s.Config.Server.Hostname,
		IPAddress:           firstPrivateIPv4(),
		AdminPasswordExists: s.Secrets.AdminPasswordHash != "",
	}
	if includeCode && !s.State.SetupComplete {
		state.SetupCode = s.State.SetupCode
	}
	return state
}

func (s *Server) session(r *http.Request, scopes ...auth.Scope) (auth.Session, bool) {
	cookie, err := r.Cookie("immich_frame_session")
	if err != nil || cookie.Value == "" || s.Sessions == nil {
		return auth.Session{}, false
	}
	return s.Sessions.Validate(cookie.Value, scopes...)
}

func (s *Server) hasSession(r *http.Request, scopes ...auth.Scope) bool {
	_, ok := s.session(r, scopes...)
	return ok
}

func (s *Server) setSessionCookie(w http.ResponseWriter, session auth.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     "immich_frame_session",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func sanitizeFrameState(state FrameState, includeSetupCode bool) FrameState {
	if !includeSetupCode {
		state.Setup.SetupCode = ""
	}
	return state
}

func (s *Server) immichHTTPClient() *http.Client {
	if s.ImmichHTTPClient != nil {
		return s.ImmichHTTPClient
	}
	return nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return false
	}
	return true
}

func writeImmichError(w http.ResponseWriter, err error) {
	var apiErr *immich.APIError
	status := http.StatusBadGateway
	body := map[string]string{"error": err.Error(), "kind": "immich_error"}
	if errors.As(err, &apiErr) {
		body["kind"] = string(apiErr.Kind)
		switch apiErr.Kind {
		case immich.ErrorInvalidURL, immich.ErrorInvalidKey, immich.ErrorInvalidInput:
			status = http.StatusBadRequest
		case immich.ErrorPermission:
			status = http.StatusForbidden
		case immich.ErrorNetwork, immich.ErrorUnavailable:
			status = http.StatusBadGateway
		default:
			status = http.StatusBadGateway
		}
	}
	writeJSONStatus(w, status, body)
}

func firstPrivateIPv4() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			ip = ip.To4()
			if ip != nil && !ip.IsLoopback() {
				return ip.String()
			}
		}
	}
	return ""
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

func writeJSONStatus(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
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
