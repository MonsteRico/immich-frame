package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigIsValidAndUsesDevelopmentLocalSource(t *testing.T) {
	cfg := DefaultConfig()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}
	if cfg.Addr() != "0.0.0.0:8787" {
		t.Fatalf("Addr() = %q, want %q", cfg.Addr(), "0.0.0.0:8787")
	}
	if cfg.Source.Mode != "local_folder" {
		t.Fatalf("Source.Mode = %q, want local_folder", cfg.Source.Mode)
	}
	if cfg.Source.LocalFolder.Path != "./dev/photos" {
		t.Fatalf("Source.LocalFolder.Path = %q, want ./dev/photos", cfg.Source.LocalFolder.Path)
	}
	if !cfg.Filters.PhotosOnly || !cfg.Filters.ExcludeVideos {
		t.Fatalf("default filters should prefer photos-only media: %+v", cfg.Filters)
	}
}

func TestLoadAppliesKnownOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	data := []byte(`
[device]
name = "Kitchen Frame"

[server]
host = "127.0.0.1"
port = 9999

[source]
mode = "random"

[display]
fit = "cover"
transition = "cut"

[slideshow]
interval_seconds = 5
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Device.Name != "Kitchen Frame" {
		t.Fatalf("Device.Name = %q", cfg.Device.Name)
	}
	if cfg.Addr() != "127.0.0.1:9999" {
		t.Fatalf("Addr() = %q", cfg.Addr())
	}
	if cfg.Source.Mode != "random" {
		t.Fatalf("Source.Mode = %q", cfg.Source.Mode)
	}
	if cfg.Display.Fit != "cover" || cfg.Display.Transition != "cut" {
		t.Fatalf("display overrides not applied: %+v", cfg.Display)
	}
	if cfg.Slideshow.IntervalSeconds != 5 {
		t.Fatalf("Slideshow.IntervalSeconds = %d", cfg.Slideshow.IntervalSeconds)
	}
}

func TestValidateReportsInvalidBaseSettings(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.Port = 0
	cfg.Display.Fit = "stretch"
	cfg.Display.Transition = "wipe"
	cfg.Slideshow.IntervalSeconds = 0
	cfg.Source.Mode = "favorites"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want validation issues")
	}
	for _, want := range []string{
		"server.port",
		"display.fit",
		"display.transition",
		"slideshow.interval_seconds",
		"source.mode",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Validate() error %q does not contain %q", err.Error(), want)
		}
	}
}

func TestSaveRoundTripsNonSecretConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := DefaultConfig()
	cfg.Device.Name = "Hall Frame"
	cfg.Immich.URL = "https://immich.example.com"
	cfg.Source.Mode = "album"
	cfg.Source.Album.ID = "album-1"
	cfg.Display.Fit = "cover"
	cfg.Slideshow.IntervalSeconds = 45

	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Device.Name != cfg.Device.Name || loaded.Immich.URL != cfg.Immich.URL || loaded.Source.Album.ID != cfg.Source.Album.ID {
		t.Fatalf("loaded config did not round trip: %+v", loaded)
	}
	if loaded.Display.Fit != "cover" || loaded.Slideshow.IntervalSeconds != 45 {
		t.Fatalf("loaded settings did not round trip: %+v", loaded)
	}
}
