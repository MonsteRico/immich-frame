package playback

import (
	"testing"

	"github.com/MonsteRico/immich-frame/internal/cache"
)

func TestQueueNextPreviousPauseAndResume(t *testing.T) {
	queue := NewQueue([]cache.Entry{
		entry("asset-one"),
		entry("asset-two"),
		entry("asset-three"),
	})

	state := queue.State()
	if state.Current == nil || state.Current.ID != "asset-one" {
		t.Fatalf("initial current = %+v, want asset-one", state.Current)
	}
	if state.Next == nil || state.Next.ID != "asset-two" {
		t.Fatalf("initial next = %+v, want asset-two", state.Next)
	}

	next, err := queue.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if next != "asset-two" {
		t.Fatalf("Next() = %q, want asset-two", next)
	}
	state = queue.State()
	if state.Previous == nil || state.Previous.ID != "asset-one" {
		t.Fatalf("previous after Next() = %+v, want asset-one", state.Previous)
	}

	previous, err := queue.Previous()
	if err != nil {
		t.Fatalf("Previous() error = %v", err)
	}
	if previous != "asset-one" {
		t.Fatalf("Previous() = %q, want asset-one", previous)
	}

	queue.Pause()
	if !queue.Paused() || !queue.State().Paused {
		t.Fatal("queue should be paused")
	}
	queue.Resume()
	if queue.Paused() || queue.State().Paused {
		t.Fatal("queue should be resumed")
	}
}

func TestQueueEmptyStateAndCommands(t *testing.T) {
	queue := NewQueue(nil)
	state := queue.State()

	if state.Configured {
		t.Fatal("empty queue should not be configured")
	}
	if state.Status != "empty" {
		t.Fatalf("empty status = %q, want empty", state.Status)
	}
	if state.Current != nil {
		t.Fatalf("empty current = %+v, want nil", state.Current)
	}
	if _, err := queue.Next(); err == nil {
		t.Fatal("Next() error = nil, want no cached media")
	}
	if _, err := queue.Previous(); err == nil {
		t.Fatal("Previous() error = nil, want no cached media")
	}
}

func entry(id string) cache.Entry {
	return cache.Entry{
		AssetID:    id,
		MediaType:  "image/jpeg",
		Title:      id,
		SourceName: "Local folder",
	}
}
