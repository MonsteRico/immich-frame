package app

import (
	"errors"
	"os"
	"path/filepath"
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
