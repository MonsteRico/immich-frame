package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MonsteRico/immich-frame/internal/api"
	"github.com/MonsteRico/immich-frame/internal/cache"
	"github.com/MonsteRico/immich-frame/internal/config"
	"github.com/MonsteRico/immich-frame/internal/playback"
	"github.com/MonsteRico/immich-frame/internal/source"
)

func TestNextRefreshBackoffIsBounded(t *testing.T) {
	tests := []struct {
		current time.Duration
		want    time.Duration
	}{
		{0, time.Minute},
		{time.Minute, 2 * time.Minute},
		{16 * time.Minute, 30 * time.Minute},
		{30 * time.Minute, 30 * time.Minute},
	}
	for _, test := range tests {
		if got := nextRefreshBackoff(test.current); got != test.want {
			t.Fatalf("nextRefreshBackoff(%s) = %s, want %s", test.current, got, test.want)
		}
	}
}

func TestSetDegradedRecordsCacheFirstStatusAndLastError(t *testing.T) {
	root := t.TempDir()
	store, err := cache.Open(filepath.Join(root, "cache"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	cachedPath := filepath.Join(root, "cached-one.jpg")
	if err := os.WriteFile(cachedPath, []byte("cached image"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	entries, err := store.Ensure([]source.Candidate{{
		ID:         "cached-one",
		SourcePath: cachedPath,
		MediaType:  "image/jpeg",
		Title:      "Cached One",
		SourceName: "Immich",
	}})
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	queue := playback.NewQueue(entries)
	server := &api.Server{
		Config: config.DefaultConfig(),
		State:  config.State{SetupComplete: true},
		Paths:  config.Paths{StateFile: filepath.Join(root, "state.json")},
		Cache:  store,
		Queue:  queue,
		Hub:    api.NewHub(),
	}
	application := &App{Cache: store, Queue: queue, API: server}

	application.setDegraded(errors.New("Unable to reach Immich server"))

	state := queue.State()
	if state.Status != "degraded" {
		t.Fatalf("status = %q, want degraded", state.Status)
	}
	if state.Current == nil || state.Current.ID != "cached-one" {
		t.Fatalf("current = %+v, want cached-one", state.Current)
	}
	saved, err := config.LoadState(server.Paths.StateFile)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if saved.LastError != "Unable to reach Immich server" {
		t.Fatalf("LastError = %q", saved.LastError)
	}
}

func TestSuccessfulRefreshPublishesRecoveredReadyStatusWithoutCacheChanges(t *testing.T) {
	root := t.TempDir()
	store, err := cache.Open(filepath.Join(root, "cache"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	queue := playback.NewQueue([]cache.Entry{{
		AssetID:    "cached-one",
		MediaType:  "image/jpeg",
		Title:      "Cached One",
		SourceName: "Immich",
	}})
	queue.SetStatus("degraded", "Immich is unavailable. Showing cached photos while retrying.")
	server := &api.Server{
		Config: config.DefaultConfig(),
		State:  config.State{SetupComplete: true, LastError: "previous outage"},
		Paths:  config.Paths{StateFile: filepath.Join(root, "state.json")},
		Cache:  store,
		Queue:  queue,
		Hub:    api.NewHub(),
	}
	application := &App{Cache: store, Queue: queue, API: server}
	ch, unsubscribe := server.Hub.Subscribe()
	defer unsubscribe()

	application.finishSuccessfulRefresh(false)

	select {
	case state := <-ch:
		if state.Status != "ready" || state.Message != "" {
			t.Fatalf("published status = %q message %q, want ready with empty message", state.Status, state.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("expected recovered ready state to be published")
	}
	saved, err := config.LoadState(server.Paths.StateFile)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if saved.LastError != "" || saved.LastSync.IsZero() {
		t.Fatalf("saved recovery state = %+v, want cleared LastError and LastSync", saved)
	}
}

func TestRecordShownRequestsRollingCacheRefreshAtThreshold(t *testing.T) {
	root := t.TempDir()
	store, err := cache.Open(filepath.Join(root, "cache"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	cfg := config.DefaultConfig()
	cfg.Source.Mode = "album"
	cfg.Cache.TargetItems = 50
	cfg.Cache.PrefetchItems = 10
	cfg.Cache.RefreshAfterShownItems = 3
	server := &api.Server{
		Config: cfg,
		State:  config.State{SetupComplete: true},
		Paths:  config.Paths{StateFile: filepath.Join(root, "state.json")},
		Cache:  store,
		Queue:  playback.NewQueue([]cache.Entry{entry("cached-one")}),
		Hub:    api.NewHub(),
	}
	application := &App{Cache: store, Queue: server.Queue, API: server, refreshNow: make(chan struct{}, 1)}

	application.recordShownAndMaybeRequestRefresh()
	application.recordShownAndMaybeRequestRefresh()
	select {
	case <-application.refreshNow:
		t.Fatal("refresh requested before threshold")
	default:
	}
	application.recordShownAndMaybeRequestRefresh()
	select {
	case <-application.refreshNow:
	case <-time.After(time.Second):
		t.Fatal("refresh was not requested at threshold")
	}
}

func TestDerivedCacheRefreshWindowUsesHalfTargetWithPrefetchFloor(t *testing.T) {
	cfg := config.CacheConfig{TargetItems: 50, PrefetchItems: 10}
	if got := cacheRefreshAfterShownItems(cfg); got != 25 {
		t.Fatalf("cacheRefreshAfterShownItems() = %d, want 25", got)
	}
	cfg = config.CacheConfig{TargetItems: 10, PrefetchItems: 8}
	if got := cacheRefreshBatchItems(cfg); got != 8 {
		t.Fatalf("cacheRefreshBatchItems() = %d, want 8", got)
	}
}

func TestVerboseLogsAreOptIn(t *testing.T) {
	var buf bytes.Buffer
	application := &App{logs: true, logWriter: &buf}
	application.logf("cache refresh summary cache_after=%d", 12)
	if got := buf.String(); !strings.Contains(got, "cache refresh summary cache_after=12") {
		t.Fatalf("log output = %q, want cache summary", got)
	}

	buf.Reset()
	application.logs = false
	application.logf("hidden")
	if buf.Len() != 0 {
		t.Fatalf("disabled logs wrote %q", buf.String())
	}
}

func entry(id string) cache.Entry {
	return cache.Entry{AssetID: id, MediaType: "image/jpeg", Title: id, SourceName: "Immich"}
}
