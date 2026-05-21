package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MonsteRico/immich-frame/internal/auth"
	"github.com/MonsteRico/immich-frame/internal/cache"
	"github.com/MonsteRico/immich-frame/internal/config"
	setupstate "github.com/MonsteRico/immich-frame/internal/setup"
	"github.com/MonsteRico/immich-frame/internal/source"
)

func TestServeEmbeddedFrameIndexWhenNoDistIsConfigured(t *testing.T) {
	server := Server{}
	request := httptest.NewRequest(http.MethodGet, "/frame", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET /frame status = %d, want 200", response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", contentType)
	}
	if !strings.Contains(response.Body.String(), "<html") {
		t.Fatal("embedded frame index response should contain HTML")
	}
}

func TestSetupClaimAdminPasswordAndSettingsNeverRevealAPIKey(t *testing.T) {
	server := newTestServer(t)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/setup/claim", map[string]string{"code": "123456"}, ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("claim status = %d body=%s", rec.Code, rec.Body.String())
	}
	cookie := rec.Result().Cookies()[0].String()

	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/setup/admin-password", map[string]string{"password": "correct horse"}, cookie))
	if rec.Code != http.StatusOK {
		t.Fatalf("password status = %d body=%s", rec.Code, rec.Body.String())
	}
	adminCookie := rec.Result().Cookies()[0].String()

	cfg := config.DefaultConfig()
	cfg.Immich.URL = "https://immich.example.com"
	cfg.Source.Mode = "random"
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, jsonRequest(http.MethodPut, "/api/settings", settingsRequest{
		Config:       cfg,
		ImmichAPIKey: ptr("raw-secret-key"),
	}, adminCookie))
	if rec.Code != http.StatusOK {
		t.Fatalf("settings status = %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "raw-secret-key") {
		t.Fatal("settings response leaked Immich API key")
	}

	secrets, err := config.LoadSecrets(server.Paths.SecretsFile)
	if err != nil {
		t.Fatal(err)
	}
	if secrets.ImmichAPIKey != "raw-secret-key" {
		t.Fatalf("saved api key = %q", secrets.ImmichAPIKey)
	}
	if secrets.AdminPasswordHash == "" || secrets.AdminPasswordHash == "correct horse" {
		t.Fatalf("admin password hash was not saved safely: %q", secrets.AdminPasswordHash)
	}
}

func TestLoginRequiredForLANMediaAccess(t *testing.T) {
	server := newTestServer(t)
	sourcePath := filepath.Join(t.TempDir(), "photo.jpg")
	if err := os.WriteFile(sourcePath, []byte("photo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := server.Cache.Ensure([]source.Candidate{{ID: "asset-1", SourcePath: sourcePath, MediaType: "image/jpeg"}}); err != nil {
		t.Fatal(err)
	}
	hash, err := auth.HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	server.Secrets.AdminPasswordHash = hash

	req := httptest.NewRequest(http.MethodGet, "/media/asset-1", nil)
	req.RemoteAddr = "192.168.1.10:1234"
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("unauth media status = %d", rec.Code)
	}

	login := httptest.NewRecorder()
	server.Handler().ServeHTTP(login, jsonRequest(http.MethodPost, "/api/auth/login", map[string]string{"password": "correct horse"}, ""))
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", login.Code, login.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/media/asset-1", nil)
	req.RemoteAddr = "192.168.1.10:1234"
	req.Header.Set("Cookie", login.Result().Cookies()[0].String())
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("auth media status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestImmichValidationUsesUserSafeErrorsAndAlbums(t *testing.T) {
	immichServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "good-key" {
			http.Error(w, "nope", http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/server/version":
			_, _ = w.Write([]byte(`{"major":1,"minor":2,"patch":3}`))
		case "/api/api-keys/me":
			_, _ = w.Write([]byte(`{"name":"Frame key"}`))
		case "/api/albums":
			_, _ = w.Write([]byte(`[{"id":"a1","albumName":"Family","assetCount":42}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer immichServer.Close()

	server := newTestServer(t)
	session, err := server.Sessions.Create(auth.ScopeAdmin)
	if err != nil {
		t.Fatal(err)
	}
	cookie := (&http.Cookie{Name: "immich_frame_session", Value: session.Token}).String()

	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/immich/test", map[string]string{"url": immichServer.URL, "apiKey": "bad-key"}, cookie))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_key") {
		t.Fatalf("bad key response = %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/immich/test", map[string]string{"url": immichServer.URL, "apiKey": "good-key"}, cookie))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "1.2.3") {
		t.Fatalf("test response = %d %s", rec.Code, rec.Body.String())
	}

	server.Config.Immich.URL = immichServer.URL
	server.Secrets.ImmichAPIKey = "good-key"
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/immich/albums", nil)
	req.Header.Set("Cookie", cookie)
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Family") || strings.Contains(rec.Body.String(), "good-key") {
		t.Fatalf("albums response = %d %s", rec.Code, rec.Body.String())
	}
}

func TestCleanAssetPathRejectsTraversal(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "app.js", want: "app.js"},
		{name: "leading slash", in: "/app.js", want: "app.js"},
		{name: "nested", in: "nested/app.css", want: "nested/app.css"},
		{name: "parent", in: "../app.js", want: ""},
		{name: "windows parent", in: `..\app.js`, want: ""},
		{name: "current", in: ".", want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := cleanAssetPath(tc.in); got != tc.want {
				t.Fatalf("cleanAssetPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	paths := config.Paths{
		ConfigFile:  filepath.Join(dir, "config.toml"),
		SecretsFile: filepath.Join(dir, "secrets.json"),
		StateFile:   filepath.Join(dir, "state.json"),
		CacheDir:    filepath.Join(dir, "cache"),
	}
	cfg := config.DefaultConfig()
	cfg.Server.Hostname = "frame.local"
	state := config.State{SetupCode: "123456", SetupStatus: string(setupstate.StatusSetupCodeRequired), UpdatedAt: time.Now()}
	if err := config.SaveState(paths.StateFile, state); err != nil {
		t.Fatal(err)
	}
	store, err := cache.Open(paths.CacheDir)
	if err != nil {
		t.Fatal(err)
	}
	return &Server{
		Config:   cfg,
		State:    state,
		Paths:    paths,
		Cache:    store,
		Hub:      NewHub(),
		Setup:    setupstate.NewManager(paths.StateFile),
		Sessions: auth.NewManager(30 * time.Minute),
	}
}

func jsonRequest(method, target string, body any, cookie string) *http.Request {
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(method, target, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	return req
}

func ptr(value string) *string {
	return &value
}
