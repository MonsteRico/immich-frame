package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/MonsteRico/immich-frame/internal/api"
	"github.com/MonsteRico/immich-frame/internal/auth"
	"github.com/MonsteRico/immich-frame/internal/cache"
	"github.com/MonsteRico/immich-frame/internal/config"
	"github.com/MonsteRico/immich-frame/internal/immich"
	"github.com/MonsteRico/immich-frame/internal/playback"
	setupstate "github.com/MonsteRico/immich-frame/internal/setup"
	"github.com/MonsteRico/immich-frame/internal/source"
)

type Options struct {
	ConfigPath string
	DataDir    string
	DevSource  string
	FrameDist  string
	SetupDist  string
}

type App struct {
	Config  config.Config
	Paths   config.Paths
	Secrets config.Secrets
	State   config.State
	Cache   *cache.Store
	Queue   *playback.Queue
	API     *api.Server
}

func New(opts Options) (*App, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	paths := config.DevPaths(root)
	if opts.DataDir != "" {
		paths.SecretsFile = filepath.Join(opts.DataDir, "secrets.json")
		paths.StateFile = filepath.Join(opts.DataDir, "state.json")
		paths.CacheDir = filepath.Join(opts.DataDir, "cache")
	}
	if opts.ConfigPath != "" {
		paths.ConfigFile = opts.ConfigPath
	}
	cfg, err := config.Load(paths.ConfigFile)
	if err != nil {
		return nil, err
	}
	if opts.DevSource != "" {
		cfg.Source.Mode = "local_folder"
		cfg.Source.LocalFolder.Path = opts.DevSource
	}
	store, err := cache.Open(paths.CacheDir)
	if err != nil {
		return nil, err
	}
	secrets, err := config.LoadSecrets(paths.SecretsFile)
	if err != nil {
		return nil, err
	}
	setupManager := setupstate.NewManager(paths.StateFile)
	state, err := setupManager.Ensure()
	if err != nil {
		return nil, err
	}
	entries, err := seedCache(context.Background(), cfg, secrets, state, store)
	if err != nil {
		entries = store.List()
	}
	queue := playback.NewQueue(entries)
	server := &api.Server{
		Config:    cfg,
		Secrets:   secrets,
		State:     state,
		Paths:     paths,
		Cache:     store,
		Queue:     queue,
		Hub:       api.NewHub(),
		Setup:     setupManager,
		Sessions:  auth.NewManager(30 * time.Minute),
		FrameDist: opts.FrameDist,
		SetupDist: opts.SetupDist,
	}
	application := &App{Config: cfg, Paths: paths, Secrets: secrets, State: state, Cache: store, Queue: queue, API: server}
	if err != nil {
		application.setDegraded(err)
	}
	return application, nil
}

func (a *App) Serve(ctx context.Context) error {
	srv := &http.Server{
		Addr:              a.Config.Addr(),
		Handler:           a.API.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go a.runSlideshow(ctx)
	go a.runCacheMaintenance(ctx)
	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("immich-frame serving http://%s/frame\n", a.Config.Addr())
		errCh <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (a *App) runCacheMaintenance(ctx context.Context) {
	backoff := time.Minute
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			changed, err := a.refreshCache(ctx)
			if err != nil {
				a.setDegraded(err)
				timer.Reset(backoff)
				backoff = nextRefreshBackoff(backoff)
				continue
			}
			backoff = time.Minute
			a.Queue.SetStatus("ready", "")
			a.API.RecordSyncStatus("", time.Now())
			if changed {
				a.Queue.Refresh(a.Cache.List())
				a.API.PublishState()
			}
			cfg, _, _ := a.API.RuntimeInputs()
			interval := time.Duration(cfg.Sync.RefreshIntervalMinutes) * time.Minute
			if interval <= 0 {
				interval = 6 * time.Hour
			}
			timer.Reset(interval)
		}
	}
}

func nextRefreshBackoff(current time.Duration) time.Duration {
	if current <= 0 {
		return time.Minute
	}
	next := current * 2
	if next > 30*time.Minute {
		return 30 * time.Minute
	}
	return next
}

func (a *App) runSlideshow(ctx context.Context) {
	interval := time.Duration(a.Config.Slideshow.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	a.API.PublishState()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if a.Queue.Paused() {
				continue
			}
			assetID, err := a.Queue.Next()
			if err == nil {
				_ = a.Cache.MarkShown(assetID)
				a.API.PublishState()
			}
		}
	}
}

func (a *App) refreshCache(ctx context.Context) (bool, error) {
	cfg, secrets, state := a.API.RuntimeInputs()
	if cfg.Source.Mode != "album" && cfg.Source.Mode != "random" {
		before := len(a.Cache.List())
		entries, err := seedCache(ctx, cfg, secrets, state, a.Cache)
		if err != nil {
			return false, err
		}
		a.Queue.Refresh(entries)
		return before != len(entries), nil
	}
	if !state.SetupComplete {
		return false, nil
	}
	client, err := immich.NewClient(cfg.Immich.URL, secrets.ImmichAPIKey)
	if err != nil {
		return false, err
	}
	candidates, err := immichCandidates(ctx, cfg, client)
	if err != nil {
		return false, err
	}
	sourceCandidates := sourceCandidates(candidates)
	sourceIDs := map[string]struct{}{}
	for _, candidate := range sourceCandidates {
		sourceIDs[candidate.ID] = struct{}{}
	}
	protectedIDs := a.Queue.ProtectedIDs(cfg.Cache.PrefetchItems)
	_, pruned, err := a.Cache.EvictSourceRemoved(sourceIDs, protectedIDs)
	if err != nil {
		return false, err
	}
	entries, toppedOff, err := a.Cache.TopOffFetched(ctx, sourceCandidates, cfg.Cache.TargetItems, func(ctx context.Context, candidate source.Candidate) (io.ReadCloser, string, error) {
		rendition, err := client.FetchRendition(ctx, candidate.ID, immich.Target{
			Width: cfg.Display.Width, Height: cfg.Display.Height, Format: cfg.Cache.Rendition,
		})
		if err != nil {
			return nil, "", err
		}
		return rendition.Body, rendition.ContentType, nil
	})
	if err != nil {
		return len(pruned) > 0 || toppedOff, err
	}
	entries, evicted, err := a.Cache.Evict(cache.EvictOptions{
		TargetItems:  cfg.Cache.TargetItems,
		SourceIDs:    sourceIDs,
		ProtectedIDs: protectedIDs,
	})
	if err != nil {
		return len(pruned) > 0 || toppedOff, err
	}
	_ = entries
	return len(pruned) > 0 || toppedOff || len(evicted) > 0, nil
}

func immichCandidates(ctx context.Context, cfg config.Config, client *immich.Client) ([]immich.AssetCandidate, error) {
	switch cfg.Source.Mode {
	case "album":
		return client.ListAlbumCandidates(ctx, cfg.Source.Album.ID)
	case "random":
		limit := cfg.Cache.TargetItems
		if limit <= 0 {
			limit = cfg.Cache.PrefetchItems
		}
		if limit <= 0 {
			limit = 100
		}
		return client.ListRandomCandidates(ctx, limit)
	default:
		return nil, nil
	}
}

func (a *App) setDegraded(err error) {
	if len(a.Cache.List()) == 0 {
		a.Queue.SetStatus("error", "Immich is unavailable and no cached photos are ready yet. The frame will keep retrying.")
	} else {
		a.Queue.SetStatus("degraded", "Immich is unavailable. Showing cached photos while retrying.")
	}
	a.API.RecordSyncStatus(err.Error(), time.Time{})
	a.API.PublishState()
}

func seedCache(ctx context.Context, cfg config.Config, secrets config.Secrets, state config.State, store *cache.Store) ([]cache.Entry, error) {
	if cfg.Source.Mode == "album" || cfg.Source.Mode == "random" {
		if !state.SetupComplete {
			return store.List(), nil
		}
		return seedImmichCache(ctx, cfg, secrets, store)
	}
	provider := source.LocalFolderProvider{Root: cfg.Source.LocalFolder.Path}
	candidates, err := provider.Candidates()
	if err != nil {
		if os.IsNotExist(err) {
			return store.List(), nil
		}
		return nil, err
	}
	return store.Ensure(candidates)
}

func seedImmichCache(ctx context.Context, cfg config.Config, secrets config.Secrets, store *cache.Store) ([]cache.Entry, error) {
	client, err := immich.NewClient(cfg.Immich.URL, secrets.ImmichAPIKey)
	if err != nil {
		return store.List(), err
	}
	var candidates []immich.AssetCandidate
	switch cfg.Source.Mode {
	case "album":
		candidates, err = client.ListAlbumCandidates(ctx, cfg.Source.Album.ID)
	case "random":
		limit := cfg.Cache.TargetItems
		if limit <= 0 {
			limit = cfg.Cache.PrefetchItems
		}
		candidates, err = client.ListRandomCandidates(ctx, limit)
	}
	if err != nil {
		return store.List(), err
	}
	entries, _, err := store.TopOffFetched(ctx, sourceCandidates(candidates), cfg.Cache.TargetItems, func(ctx context.Context, candidate source.Candidate) (io.ReadCloser, string, error) {
		rendition, err := client.FetchRendition(ctx, candidate.ID, immich.Target{
			Width: cfg.Display.Width, Height: cfg.Display.Height, Format: cfg.Cache.Rendition,
		})
		if err != nil {
			return nil, "", err
		}
		return rendition.Body, rendition.ContentType, nil
	})
	return entries, err
}

func sourceCandidates(candidates []immich.AssetCandidate) []source.Candidate {
	out := make([]source.Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, source.Candidate{
			ID:                candidate.ID,
			RenditionIdentity: candidate.RenditionIdentity,
			Title:             candidate.Title,
			SourceName:        candidate.SourceName,
			TakenAt:           candidate.TakenAt,
			MediaType:         candidate.MediaType,
			Width:             candidate.Width,
			Height:            candidate.Height,
			Orientation:       candidate.Orientation,
		})
	}
	return out
}
