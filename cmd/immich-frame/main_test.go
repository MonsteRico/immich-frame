package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MonsteRico/immich-frame/internal/cache"
	"github.com/MonsteRico/immich-frame/internal/config"
	"github.com/MonsteRico/immich-frame/internal/source"
)

func TestStatusReportsSafeRuntimeDetails(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.toml")
	dataDir := filepath.Join(root, "data")
	cfg := config.DefaultConfig()
	cfg.Immich.URL = "https://immich.example.com"
	cfg.Source.Mode = "random"
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	secrets := config.Secrets{ImmichAPIKey: "super-secret-key"}
	if err := config.SaveSecrets(filepath.Join(dataDir, "secrets.json"), secrets); err != nil {
		t.Fatalf("SaveSecrets() error = %v", err)
	}
	state := config.State{
		SetupComplete:    true,
		SetupStatus:      "configured",
		ImmichValidation: config.NewImmichValidation(cfg.Immich.URL, secrets.ImmichAPIKey, "1.2.3", "Frame", time.Now()),
		LastError:        "Unable to reach Immich server",
		LastSync:         time.Date(2026, 5, 24, 1, 2, 3, 0, time.UTC),
	}
	if err := config.SaveState(filepath.Join(dataDir, "state.json"), state); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}
	store, err := cache.Open(filepath.Join(dataDir, "cache"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	sourcePath := filepath.Join(root, "one.jpg")
	if err := os.WriteFile(sourcePath, []byte("photo"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := store.Ensure([]source.Candidate{{ID: "one", SourcePath: sourcePath, MediaType: "image/jpeg"}}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	output := captureStdout(t, func() {
		if err := run([]string{"status", "-config", configPath, "-data-dir", dataDir}); err != nil {
			t.Fatalf("status error = %v", err)
		}
	})
	for _, want := range []string{
		"setup_complete=true",
		"config_valid=true",
		"source_mode=random",
		"immich_api_key_configured=true",
		"immich_validation_current=true",
		"cache_count=1",
		"last_error=Unable to reach Immich server",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("status output %q missing %q", output, want)
		}
	}
	if strings.Contains(output, secrets.ImmichAPIKey) {
		t.Fatalf("status output leaked API key: %q", output)
	}
}

func TestResetRemovesPrivateStateAndOptionalConfig(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	configPath := filepath.Join(root, "config.toml")
	for _, path := range []string{
		filepath.Join(dataDir, "secrets.json"),
		filepath.Join(dataDir, "state.json"),
		filepath.Join(dataDir, "cache", "one.jpg"),
		configPath,
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte("private"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	captureStdout(t, func() {
		if err := run([]string{"reset", "-data-dir", dataDir, "-config", configPath}); err != nil {
			t.Fatalf("reset error = %v", err)
		}
	})
	for _, path := range []string{
		filepath.Join(dataDir, "secrets.json"),
		filepath.Join(dataDir, "state.json"),
		filepath.Join(dataDir, "cache"),
		configPath,
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s still exists or stat error = %v", path, err)
		}
	}
}

func TestRendererPOCWritesPreview(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "source.png")
	outPath := filepath.Join(root, "preview.png")
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 80, B: 20, A: 255})
		}
	}
	file, err := os.Create(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		if err := run([]string{"renderer-poc", "-image", sourcePath, "-out", outPath, "-width", "64", "-height", "48"}); err != nil {
			t.Fatalf("renderer-poc error = %v", err)
		}
	})
	if !strings.Contains(output, "renderer proof-of-concept preview written") {
		t.Fatalf("unexpected renderer-poc output: %q", output)
	}
	preview, err := os.Open(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer preview.Close()
	decoded, err := png.Decode(preview)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != 64 || decoded.Bounds().Dy() != 48 {
		t.Fatalf("preview bounds = %v, want 64x48", decoded.Bounds())
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	os.Stdout = write
	fn()
	_ = write.Close()
	os.Stdout = original
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(read); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	return buf.String()
}
