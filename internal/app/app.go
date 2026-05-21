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
	Config config.Config
	Paths  config.Paths
	Cache  *cache.Store
	Queue  *playback.Queue
	API    *api.Server
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
		return nil, err
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
	return &App{Config: cfg, Paths: paths, Cache: store, Queue: queue, API: server}, nil
}

func (a *App) Serve(ctx context.Context) error {
	srv := &http.Server{
		Addr:              a.Config.Addr(),
		Handler:           a.API.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go a.runSlideshow(ctx)
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
		limit := cfg.Cache.PrefetchItems
		if limit <= 0 {
			limit = cfg.Cache.TargetItems
		}
		candidates, err = client.ListRandomCandidates(ctx, limit)
	}
	if err != nil {
		return store.List(), err
	}
	return store.EnsureFetched(ctx, sourceCandidates(candidates), func(ctx context.Context, candidate source.Candidate) (io.ReadCloser, string, error) {
		rendition, err := client.FetchRendition(ctx, candidate.ID, immich.Target{
			Width: cfg.Display.Width, Height: cfg.Display.Height, Format: cfg.Cache.Rendition,
		})
		if err != nil {
			return nil, "", err
		}
		return rendition.Body, rendition.ContentType, nil
	})
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
