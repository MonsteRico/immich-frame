package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFolderProviderDiscoversPhotoCandidates(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "First.JPG"), []byte("jpg"))
	writeFile(t, filepath.Join(root, "nested", "Second.svg"), []byte("<svg></svg>"))
	writeFile(t, filepath.Join(root, "notes.txt"), []byte("not a photo"))

	candidates, err := LocalFolderProvider{Root: root}.Candidates()
	if err != nil {
		t.Fatalf("Candidates() error = %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("Candidates() len = %d, want 2: %+v", len(candidates), candidates)
	}

	byTitle := map[string]Candidate{}
	for _, candidate := range candidates {
		byTitle[candidate.Title] = candidate
	}

	first := byTitle["First"]
	if first.ID != stableID("First.JPG") {
		t.Fatalf("First ID = %q, want stable ID for rel path", first.ID)
	}
	if first.MediaType != "image/jpeg" {
		t.Fatalf("First media type = %q", first.MediaType)
	}
	if first.SourceName != "Local folder" {
		t.Fatalf("First source name = %q", first.SourceName)
	}
	if !filepath.IsAbs(first.SourcePath) {
		t.Fatalf("SourcePath = %q, want absolute path", first.SourcePath)
	}

	second := byTitle["Second"]
	if second.ID != stableID(filepath.Join("nested", "Second.svg")) {
		t.Fatalf("Second ID = %q, want stable ID for nested rel path", second.ID)
	}
	if second.MediaType != "image/svg+xml" {
		t.Fatalf("Second media type = %q", second.MediaType)
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
