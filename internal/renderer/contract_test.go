package renderer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MonsteRico/immich-frame/internal/config"
	"github.com/MonsteRico/immich-frame/internal/playback"
)

func TestFromFrameStateBuildsLocalRendererSnapshot(t *testing.T) {
	takenAt := time.Date(2026, 5, 24, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 24, 11, 0, 0, 0, time.UTC)
	cfg := config.DefaultConfig()
	cfg.Display.Width = 800
	cfg.Display.Height = 480
	cfg.Display.Fit = "cover"
	cfg.Slideshow.IntervalSeconds = 12
	state := playback.State{
		Paused:    true,
		Status:    "ready",
		UpdatedAt: updatedAt,
		Current: &playback.Asset{
			ID:         "asset-1",
			MediaURL:   "/media/asset-1",
			Title:      "Backyard",
			SourceName: "Family",
			TakenAt:    takenAt,
		},
		Next: &playback.Asset{ID: "asset-2", MediaURL: "/media/asset-2"},
	}
	snapshot := FromFrameState(state, cfg, cfg.Overlays, true, func(assetID string) (string, string, bool) {
		if assetID == "asset-1" {
			return "/cache/asset-1.jpg", "image/jpeg", true
		}
		return "", "", false
	})

	if snapshot.Status != "ready" || !snapshot.Playback.Paused || snapshot.Playback.IntervalSeconds != 12 || snapshot.Playback.Fit != "cover" {
		t.Fatalf("unexpected playback snapshot: %+v", snapshot)
	}
	if snapshot.Display.Width != 800 || snapshot.Display.Height != 480 || snapshot.Display.Orientation != "auto" {
		t.Fatalf("unexpected display snapshot: %+v", snapshot.Display)
	}
	if snapshot.Current == nil || snapshot.Current.LocalPath != "/cache/asset-1.jpg" || snapshot.Current.ContentType != "image/jpeg" {
		t.Fatalf("current asset did not include local cache details: %+v", snapshot.Current)
	}
	if snapshot.Next == nil || snapshot.Next.LocalPath != "" {
		t.Fatalf("next asset should be present without cache path: %+v", snapshot.Next)
	}
	if !snapshot.Setup.Configured || !snapshot.UpdatedAt.Equal(updatedAt) || !snapshot.Current.TakenAt.Equal(takenAt) {
		t.Fatalf("snapshot lost setup/time fields: %+v", snapshot)
	}
}

func TestFromFrameStateFallsBackDisplayTarget(t *testing.T) {
	cfg := config.DefaultConfig()
	snapshot := FromFrameState(playback.State{}, cfg, cfg.Overlays, false, nil)
	if snapshot.Display.Width != 1920 || snapshot.Display.Height != 1080 {
		t.Fatalf("display fallback = %+v, want 1920x1080", snapshot.Display)
	}
	if snapshot.Status != "empty" {
		t.Fatalf("status = %q, want empty", snapshot.Status)
	}
}

func TestLoopKeepsVisibleAssetWhenFetchFails(t *testing.T) {
	snapshots := []Snapshot{
		{Status: "ready", Current: &Asset{ID: "asset-1", LocalPath: "one.jpg"}},
	}
	loop := Loop{
		Fetch: func(context.Context) (Snapshot, error) {
			if len(snapshots) == 0 {
				return Snapshot{}, errors.New("network sleeping")
			}
			next := snapshots[0]
			snapshots = snapshots[1:]
			return next, nil
		},
		Decode: func(context.Context, Asset) error { return nil },
	}

	first := loop.Step(context.Background())
	if first.Visible == nil || first.Visible.ID != "asset-1" || first.Err != nil {
		t.Fatalf("first step = %+v", first)
	}
	second := loop.Step(context.Background())
	if second.Visible == nil || second.Visible.ID != "asset-1" {
		t.Fatalf("visible asset was not retained after fetch failure: %+v", second)
	}
	if second.Err == nil || second.Err.Error() != "network sleeping" {
		t.Fatalf("failure was not surfaced: %+v", second.Err)
	}
}

func TestLoopKeepsVisibleAssetWhenDecodeFails(t *testing.T) {
	snapshots := []Snapshot{
		{Status: "ready", Current: &Asset{ID: "asset-1", LocalPath: "one.jpg"}},
		{Status: "ready", Current: &Asset{ID: "asset-2", LocalPath: "two.jpg"}},
	}
	loop := Loop{
		Fetch: func(context.Context) (Snapshot, error) {
			next := snapshots[0]
			snapshots = snapshots[1:]
			return next, nil
		},
		Decode: func(_ context.Context, asset Asset) error {
			if asset.ID == "asset-2" {
				return errors.New("decode failed")
			}
			return nil
		},
	}

	first := loop.Step(context.Background())
	if first.Visible == nil || first.Visible.ID != "asset-1" {
		t.Fatalf("first step = %+v", first)
	}
	second := loop.Step(context.Background())
	if second.Visible == nil || second.Visible.ID != "asset-1" {
		t.Fatalf("visible asset was not retained after decode failure: %+v", second)
	}
	if second.Err == nil || second.Err.Error() != "decode failed" {
		t.Fatalf("decode failure was not surfaced: %+v", second.Err)
	}
}
