package cache

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MonsteRico/immich-frame/internal/source"
)

type Entry struct {
	AssetID           string    `json:"assetId"`
	SourcePath        string    `json:"sourcePath,omitempty"`
	RenditionIdentity string    `json:"renditionIdentity,omitempty"`
	CachePath         string    `json:"cachePath"`
	MediaType         string    `json:"mediaType"`
	Title             string    `json:"title"`
	SourceName        string    `json:"sourceName"`
	TakenAt           time.Time `json:"takenAt,omitempty"`
	Width             int       `json:"width,omitempty"`
	Height            int       `json:"height,omitempty"`
	Orientation       string    `json:"orientation,omitempty"`
	CachedAt          time.Time `json:"cachedAt"`
	LastShown         time.Time `json:"lastShown,omitempty"`
}

type Manifest struct {
	Version   int              `json:"version"`
	Entries   map[string]Entry `json:"entries"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

type Store struct {
	Dir          string
	ManifestPath string
	manifest     Manifest
}

type FetchFunc func(context.Context, source.Candidate) (io.ReadCloser, string, error)

func Open(dir string) (*Store, error) {
	store := &Store{
		Dir:          dir,
		ManifestPath: filepath.Join(dir, "manifest.json"),
		manifest:     Manifest{Version: 1, Entries: map[string]Entry{}, UpdatedAt: time.Now()},
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(store.ManifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return store, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &store.manifest); err != nil {
		return nil, err
	}
	if store.manifest.Entries == nil {
		store.manifest.Entries = map[string]Entry{}
	}
	return store, nil
}

func (s *Store) Ensure(candidates []source.Candidate) ([]Entry, error) {
	for _, candidate := range candidates {
		if _, ok := s.manifest.Entries[candidate.ID]; ok {
			continue
		}
		entry, err := s.cacheLocal(candidate)
		if err != nil {
			return nil, err
		}
		s.manifest.Entries[entry.AssetID] = entry
	}
	if err := s.Save(); err != nil {
		return nil, err
	}
	return s.List(), nil
}

func (s *Store) EnsureFetched(ctx context.Context, candidates []source.Candidate, fetch FetchFunc) ([]Entry, error) {
	for _, candidate := range candidates {
		if _, ok := s.manifest.Entries[candidate.ID]; ok {
			continue
		}
		entry, err := s.cacheFetched(ctx, candidate, fetch)
		if err != nil {
			return nil, err
		}
		s.manifest.Entries[entry.AssetID] = entry
	}
	if err := s.Save(); err != nil {
		return nil, err
	}
	return s.List(), nil
}

func (s *Store) Get(assetID string) (Entry, bool) {
	entry, ok := s.manifest.Entries[assetID]
	return entry, ok
}

func (s *Store) MarkShown(assetID string) error {
	entry, ok := s.manifest.Entries[assetID]
	if !ok {
		return nil
	}
	entry.LastShown = time.Now()
	s.manifest.Entries[assetID] = entry
	return s.Save()
}

func (s *Store) List() []Entry {
	entries := make([]Entry, 0, len(s.manifest.Entries))
	for _, entry := range s.manifest.Entries {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].LastShown.Equal(entries[j].LastShown) {
			return entries[i].AssetID < entries[j].AssetID
		}
		if entries[i].LastShown.IsZero() {
			return true
		}
		if entries[j].LastShown.IsZero() {
			return false
		}
		return entries[i].LastShown.Before(entries[j].LastShown)
	})
	return entries
}

func (s *Store) Save() error {
	s.manifest.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(s.manifest, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.ManifestPath + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.ManifestPath)
}

func (s *Store) cacheLocal(candidate source.Candidate) (Entry, error) {
	ext := filepath.Ext(candidate.SourcePath)
	cachePath := filepath.Join(s.Dir, candidate.ID+ext)
	if err := copyFile(cachePath, candidate.SourcePath); err != nil {
		return Entry{}, err
	}
	return Entry{
		AssetID:           candidate.ID,
		SourcePath:        candidate.SourcePath,
		RenditionIdentity: candidate.RenditionIdentity,
		CachePath:         cachePath,
		MediaType:         candidate.MediaType,
		Title:             candidate.Title,
		SourceName:        candidate.SourceName,
		TakenAt:           candidate.TakenAt,
		Width:             candidate.Width,
		Height:            candidate.Height,
		Orientation:       candidate.Orientation,
		CachedAt:          time.Now(),
	}, nil
}

func (s *Store) cacheFetched(ctx context.Context, candidate source.Candidate, fetch FetchFunc) (Entry, error) {
	body, mediaType, err := fetch(ctx, candidate)
	if err != nil {
		return Entry{}, err
	}
	defer body.Close()
	if mediaType == "" {
		mediaType = candidate.MediaType
	}
	ext := extensionForMediaType(mediaType)
	cachePath := filepath.Join(s.Dir, candidate.ID+ext)
	out, err := os.Create(cachePath)
	if err != nil {
		return Entry{}, err
	}
	defer out.Close()
	if _, err := io.Copy(out, body); err != nil {
		return Entry{}, err
	}
	if err := out.Close(); err != nil {
		return Entry{}, err
	}
	return Entry{
		AssetID:           candidate.ID,
		RenditionIdentity: candidate.RenditionIdentity,
		CachePath:         cachePath,
		MediaType:         mediaType,
		Title:             candidate.Title,
		SourceName:        candidate.SourceName,
		TakenAt:           candidate.TakenAt,
		Width:             candidate.Width,
		Height:            candidate.Height,
		Orientation:       candidate.Orientation,
		CachedAt:          time.Now(),
	}, nil
}

func extensionForMediaType(mediaType string) string {
	mediaType = strings.TrimSpace(strings.Split(mediaType, ";")[0])
	switch mediaType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	}
	if exts, err := mime.ExtensionsByType(mediaType); err == nil && len(exts) > 0 {
		return exts[0]
	}
	return ".bin"
}

func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
