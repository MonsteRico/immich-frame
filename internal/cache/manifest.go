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
	"sync"
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
	mu           sync.Mutex
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return s.listLocked(), nil
}

func (s *Store) EnsureFetched(ctx context.Context, candidates []source.Candidate, fetch FetchFunc) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return s.listLocked(), nil
}

func (s *Store) TopOffFetched(ctx context.Context, candidates []source.Candidate, targetItems int, fetch FetchFunc) ([]Entry, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if targetItems <= 0 {
		targetItems = len(candidates)
	}
	ordered := s.prioritizedCandidatesLocked(candidates)
	changed := false
	for _, candidate := range ordered {
		if len(s.manifest.Entries) >= targetItems {
			break
		}
		if _, ok := s.manifest.Entries[candidate.ID]; ok {
			continue
		}
		entry, err := s.cacheFetched(ctx, candidate, fetch)
		if err != nil {
			return s.listLocked(), changed, err
		}
		s.manifest.Entries[entry.AssetID] = entry
		changed = true
	}
	if changed {
		if err := s.saveLocked(); err != nil {
			return nil, changed, err
		}
	}
	return s.listLocked(), changed, nil
}

type EvictOptions struct {
	TargetItems  int
	SourceIDs    map[string]struct{}
	ProtectedIDs map[string]struct{}
}

func (s *Store) EvictSourceRemoved(sourceIDs, protectedIDs map[string]struct{}) ([]Entry, []string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var evicted []string
	for _, entry := range s.evictionCandidatesLocked(EvictOptions{SourceIDs: sourceIDs, ProtectedIDs: protectedIDs}) {
		if _, inSource := sourceIDs[entry.AssetID]; inSource {
			continue
		}
		if _, protected := protectedIDs[entry.AssetID]; protected {
			continue
		}
		delete(s.manifest.Entries, entry.AssetID)
		_ = os.Remove(entry.CachePath)
		evicted = append(evicted, entry.AssetID)
	}
	if len(evicted) > 0 {
		if err := s.saveLocked(); err != nil {
			return nil, evicted, err
		}
	}
	return s.listLocked(), evicted, nil
}

func (s *Store) Evict(opts EvictOptions) ([]Entry, []string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if opts.TargetItems <= 0 || len(s.manifest.Entries) <= opts.TargetItems {
		return s.listLocked(), nil, nil
	}
	victims := s.evictionCandidatesLocked(opts)
	var evicted []string
	for _, entry := range victims {
		if len(s.manifest.Entries) <= opts.TargetItems {
			break
		}
		if _, protected := opts.ProtectedIDs[entry.AssetID]; protected {
			continue
		}
		delete(s.manifest.Entries, entry.AssetID)
		_ = os.Remove(entry.CachePath)
		evicted = append(evicted, entry.AssetID)
	}
	if len(evicted) > 0 {
		if err := s.saveLocked(); err != nil {
			return nil, evicted, err
		}
	}
	return s.listLocked(), evicted, nil
}

func (s *Store) Get(assetID string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.manifest.Entries[assetID]
	return entry, ok
}

func (s *Store) MarkShown(assetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.manifest.Entries[assetID]
	if !ok {
		return nil
	}
	entry.LastShown = time.Now()
	s.manifest.Entries[assetID] = entry
	return s.saveLocked()
}

func (s *Store) List() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listLocked()
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *Store) listLocked() []Entry {
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

func (s *Store) saveLocked() error {
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

func (s *Store) prioritizedCandidatesLocked(candidates []source.Candidate) []source.Candidate {
	ordered := append([]source.Candidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, leftCached := s.manifest.Entries[ordered[i].ID]
		right, rightCached := s.manifest.Entries[ordered[j].ID]
		if leftCached != rightCached {
			return !leftCached
		}
		if !leftCached {
			return false
		}
		if left.LastShown.Equal(right.LastShown) {
			return left.AssetID < right.AssetID
		}
		if left.LastShown.IsZero() {
			return true
		}
		if right.LastShown.IsZero() {
			return false
		}
		return left.LastShown.Before(right.LastShown)
	})
	return ordered
}

func (s *Store) evictionCandidatesLocked(opts EvictOptions) []Entry {
	entries := make([]Entry, 0, len(s.manifest.Entries))
	for _, entry := range s.manifest.Entries {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		left := entries[i]
		right := entries[j]
		_, leftInSource := opts.SourceIDs[left.AssetID]
		_, rightInSource := opts.SourceIDs[right.AssetID]
		if leftInSource != rightInSource {
			return !leftInSource
		}
		if left.LastShown.IsZero() != right.LastShown.IsZero() {
			return !left.LastShown.IsZero()
		}
		if left.LastShown.Equal(right.LastShown) {
			if left.CachedAt.Equal(right.CachedAt) {
				return left.AssetID < right.AssetID
			}
			return left.CachedAt.Before(right.CachedAt)
		}
		return left.LastShown.After(right.LastShown)
	})
	return entries
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
