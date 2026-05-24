package renderer

import (
	"strings"
	"time"

	"github.com/MonsteRico/immich-frame/internal/config"
	"github.com/MonsteRico/immich-frame/internal/playback"
)

type Snapshot struct {
	Status    string               `json:"status"`
	Message   string               `json:"message,omitempty"`
	Current   *Asset               `json:"current,omitempty"`
	Next      *Asset               `json:"next,omitempty"`
	Playback  Playback             `json:"playback"`
	Overlays  config.OverlayConfig `json:"overlays"`
	Display   Display              `json:"display"`
	Setup     Setup                `json:"setup"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

type Asset struct {
	ID          string    `json:"id"`
	MediaURL    string    `json:"mediaUrl"`
	LocalPath   string    `json:"localPath,omitempty"`
	Title       string    `json:"title,omitempty"`
	SourceName  string    `json:"sourceName,omitempty"`
	TakenAt     time.Time `json:"takenAt,omitempty"`
	ContentType string    `json:"contentType,omitempty"`
}

type Playback struct {
	IntervalSeconds int    `json:"intervalSeconds"`
	Paused          bool   `json:"paused"`
	Fit             string `json:"fit"`
}

type Display struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Orientation string `json:"orientation"`
}

type Setup struct {
	Configured bool `json:"configured"`
}

type CacheLookup func(assetID string) (localPath, contentType string, ok bool)

func FromFrameState(state playback.State, cfg config.Config, overlays config.OverlayConfig, setupConfigured bool, lookup CacheLookup) Snapshot {
	snapshot := Snapshot{
		Status:  nonEmpty(state.Status, "empty"),
		Message: state.Message,
		Playback: Playback{
			IntervalSeconds: cfg.Slideshow.IntervalSeconds,
			Paused:          state.Paused,
			Fit:             nonEmpty(cfg.Display.Fit, "contain"),
		},
		Overlays: overlays,
		Display: Display{
			Width:       cfg.Display.Width,
			Height:      cfg.Display.Height,
			Orientation: nonEmpty(cfg.Display.Orientation, "auto"),
		},
		Setup:     Setup{Configured: setupConfigured},
		UpdatedAt: state.UpdatedAt,
	}
	if snapshot.Display.Width <= 0 {
		snapshot.Display.Width = 1920
	}
	if snapshot.Display.Height <= 0 {
		snapshot.Display.Height = 1080
	}
	snapshot.Current = fromPlaybackAsset(state.Current, lookup)
	snapshot.Next = fromPlaybackAsset(state.Next, lookup)
	return snapshot
}

func fromPlaybackAsset(asset *playback.Asset, lookup CacheLookup) *Asset {
	if asset == nil {
		return nil
	}
	out := &Asset{
		ID:         asset.ID,
		MediaURL:   asset.MediaURL,
		Title:      asset.Title,
		SourceName: asset.SourceName,
		TakenAt:    asset.TakenAt,
	}
	if lookup != nil && strings.TrimSpace(asset.ID) != "" {
		localPath, contentType, ok := lookup(asset.ID)
		if ok {
			out.LocalPath = localPath
			out.ContentType = contentType
		}
	}
	return out
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
