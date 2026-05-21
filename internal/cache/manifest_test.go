package cache

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MonsteRico/immich-frame/internal/source"
)

func TestStoreEnsureListAndMarkShown(t *testing.T) {
	root := t.TempDir()
	sourceOne := filepath.Join(root, "one.jpg")
	sourceTwo := filepath.Join(root, "two.png")
	writeFile(t, sourceOne, []byte("one"))
	writeFile(t, sourceTwo, []byte("two"))

	store, err := Open(filepath.Join(root, "cache"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	entries, err := store.Ensure([]source.Candidate{
		{ID: "asset-one", SourcePath: sourceOne, Title: "One", SourceName: "Local folder", MediaType: "image/jpeg"},
		{ID: "asset-two", SourcePath: sourceTwo, Title: "Two", SourceName: "Local folder", MediaType: "image/png"},
	})
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Ensure() len = %d, want 2", len(entries))
	}
	for _, id := range []string{"asset-one", "asset-two"} {
		entry, ok := store.Get(id)
		if !ok {
			t.Fatalf("Get(%q) ok = false", id)
		}
		if _, err := os.Stat(entry.CachePath); err != nil {
			t.Fatalf("cached file for %q missing: %v", id, err)
		}
	}

	if err := store.MarkShown("asset-one"); err != nil {
		t.Fatalf("MarkShown() error = %v", err)
	}
	if err := store.MarkShown("missing-asset"); err != nil {
		t.Fatalf("MarkShown() missing asset error = %v", err)
	}

	list := store.List()
	if list[0].AssetID != "asset-two" {
		t.Fatalf("List()[0] = %q, want unshown asset first", list[0].AssetID)
	}
	if list[1].AssetID != "asset-one" || list[1].LastShown.IsZero() {
		t.Fatalf("List()[1] = %+v, want shown asset with LastShown", list[1])
	}

	reopened, err := Open(filepath.Join(root, "cache"))
	if err != nil {
		t.Fatalf("Open() reopened error = %v", err)
	}
	entry, ok := reopened.Get("asset-one")
	if !ok || entry.LastShown.IsZero() {
		t.Fatalf("reopened manifest did not preserve LastShown: %+v ok=%t", entry, ok)
	}
}

func TestStoreEnsureFetchedCachesRemoteRenditionMetadata(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "cache"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	entries, err := store.EnsureFetched(context.Background(), []source.Candidate{
		{
			ID: "immich-one", RenditionIdentity: "thumbnail-webp", Title: "One", SourceName: "Immich",
			MediaType: "image/jpeg", Width: 1920, Height: 1080, Orientation: "landscape",
		},
	}, func(ctx context.Context, candidate source.Candidate) (io.ReadCloser, string, error) {
		return io.NopCloser(strings.NewReader("remote image")), "image/webp", nil
	})
	if err != nil {
		t.Fatalf("EnsureFetched() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("EnsureFetched() len = %d, want 1", len(entries))
	}
	entry := entries[0]
	if entry.CachePath == "" || filepath.Ext(entry.CachePath) != ".webp" {
		t.Fatalf("cache path = %q, want .webp", entry.CachePath)
	}
	if entry.SourcePath != "" || entry.RenditionIdentity != "thumbnail-webp" || entry.MediaType != "image/webp" {
		t.Fatalf("unexpected entry source metadata: %+v", entry)
	}
	if entry.Width != 1920 || entry.Height != 1080 || entry.Orientation != "landscape" {
		t.Fatalf("unexpected display metadata: %+v", entry)
	}
	data, err := os.ReadFile(entry.CachePath)
	if err != nil {
		t.Fatalf("read cached file: %v", err)
	}
	if string(data) != "remote image" {
		t.Fatalf("cached data = %q", data)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
