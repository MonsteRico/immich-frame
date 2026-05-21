package immich

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClientValidatesInputs(t *testing.T) {
	if _, err := NewClient("", "key"); !IsKind(err, ErrorInvalidURL) {
		t.Fatalf("expected invalid URL error, got %v", err)
	}
	if _, err := NewClient("ftp://example.test", "key"); !IsKind(err, ErrorInvalidURL) {
		t.Fatalf("expected invalid URL scheme error, got %v", err)
	}
	if _, err := NewClient("https://example.test", ""); !IsKind(err, ErrorInvalidKey) {
		t.Fatalf("expected invalid key error, got %v", err)
	}
}

func TestTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "secret" {
			http.Error(w, "nope", http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/server/version":
			writeTestJSON(t, w, map[string]int{"major": 1, "minor": 136, "patch": 0})
		case "/api/api-keys/me":
			writeTestJSON(t, w, map[string]string{"name": "frame"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "secret")
	info, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if info.Version != "1.136.0" || info.KeyName != "frame" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestConnectionInvalidKeyIsUserSafe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "secret details", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "bad")
	_, err := client.TestConnection(context.Background())
	if !IsKind(err, ErrorInvalidKey) {
		t.Fatalf("expected invalid key error, got %v", err)
	}
	if strings.Contains(err.Error(), "secret details") {
		t.Fatalf("leaked response body in user error: %v", err)
	}
}

func TestListAlbumsNormalizesAndSorts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/albums" {
			http.NotFound(w, r)
			return
		}
		writeTestJSON(t, w, []map[string]any{
			{"id": "b", "albumName": "Vacation", "assetCount": 7},
			{"id": "a", "albumName": "Family", "assetCount": 3},
			{"albumName": "ignored", "assetCount": 1},
		})
	}))
	defer server.Close()

	albums, err := newTestClient(t, server.URL, "secret").ListAlbums(context.Background())
	if err != nil {
		t.Fatalf("ListAlbums() error = %v", err)
	}
	if len(albums) != 2 {
		t.Fatalf("expected 2 albums, got %d", len(albums))
	}
	if albums[0] != (Album{ID: "a", Name: "Family", AssetCount: 3}) {
		t.Fatalf("unexpected first album: %+v", albums[0])
	}
}

func TestListAlbumCandidatesFiltersAndNormalizesMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/albums/album-1" {
			http.NotFound(w, r)
			return
		}
		writeTestJSON(t, w, map[string]any{
			"id": "album-1", "albumName": "Kitchen Wall", "assetCount": 4,
			"assets": []map[string]any{
				{
					"id": "img-1", "type": "IMAGE", "originalFileName": "sunset.jpg",
					"originalMimeType": "image/jpeg", "localDateTime": "2026-05-01T12:30:00",
					"visibility": "timeline",
					"exifInfo":   map[string]any{"exifImageWidth": 4000, "exifImageHeight": 3000},
				},
				{"id": "video-1", "type": "VIDEO", "visibility": "timeline"},
				{"id": "hidden-1", "type": "IMAGE", "visibility": "hidden"},
				{"id": "trashed-1", "type": "IMAGE", "isTrashed": true, "visibility": "timeline"},
			},
		})
	}))
	defer server.Close()

	candidates, err := newTestClient(t, server.URL, "secret").ListAlbumCandidates(context.Background(), "album-1")
	if err != nil {
		t.Fatalf("ListAlbumCandidates() error = %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	got := candidates[0]
	if got.ID != "img-1" || got.SourceName != "Kitchen Wall" || got.Width != 4000 || got.Height != 3000 || got.Orientation != "landscape" {
		t.Fatalf("unexpected candidate: %+v", got)
	}
	if got.TakenAt.IsZero() {
		t.Fatal("expected taken date to be normalized")
	}
}

func TestListRandomCandidatesUsesConservativeFilters(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search/random" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		writeTestJSON(t, w, []map[string]any{
			{"id": "img-1", "type": "IMAGE", "visibility": "timeline"},
		})
	}))
	defer server.Close()

	candidates, err := newTestClient(t, server.URL, "secret").ListRandomCandidates(context.Background(), 25)
	if err != nil {
		t.Fatalf("ListRandomCandidates() error = %v", err)
	}
	if len(candidates) != 1 || candidates[0].SourceName != "Immich library" {
		t.Fatalf("unexpected candidates: %+v", candidates)
	}
	if body["type"] != "IMAGE" || body["visibility"] != "timeline" || body["withDeleted"] != false || body["withPeople"] != false || body["size"] != float64(25) {
		t.Fatalf("unexpected random request filters: %#v", body)
	}
}

func TestFetchRenditionUsesThumbnailAndPreservesContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/assets/img-1/thumbnail" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("format") != "WEBP" {
			t.Fatalf("expected WEBP format, got %q", r.URL.Query().Get("format"))
		}
		if r.URL.Query().Get("size") != "preview" {
			t.Fatalf("expected preview size, got %q", r.URL.Query().Get("size"))
		}
		w.Header().Set("Content-Type", "image/webp")
		_, _ = w.Write([]byte("webp bytes"))
	}))
	defer server.Close()

	rendition, err := newTestClient(t, server.URL, "secret").FetchRendition(context.Background(), "img-1", Target{Width: 1920, Height: 1080})
	if err != nil {
		t.Fatalf("FetchRendition() error = %v", err)
	}
	defer rendition.Body.Close()
	data, err := io.ReadAll(rendition.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if rendition.ContentType != "image/webp" || string(data) != "webp bytes" || rendition.Identity != "preview-webp" {
		t.Fatalf("unexpected rendition: %+v body=%q", rendition, data)
	}
}

func newTestClient(t *testing.T, baseURL, apiKey string) *Client {
	t.Helper()
	client, err := NewClient(baseURL, apiKey, WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}
