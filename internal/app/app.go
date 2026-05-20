package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/MonsteRico/immich-frame/internal/api"
	"github.com/MonsteRico/immich-frame/internal/cache"
	"github.com/MonsteRico/immich-frame/internal/config"
	"github.com/MonsteRico/immich-frame/internal/playback"
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
	entries, err := seedCache(cfg, store)
	if err != nil {
		return nil, err
	}
	queue := playback.NewQueue(entries)
	server := &api.Server{
		Config:    cfg,
		Cache:     store,
		Queue:     queue,
		Hub:       api.NewHub(),
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

func seedCache(cfg config.Config, store *cache.Store) ([]cache.Entry, error) {
	if cfg.Source.Mode != "local_folder" {
		return store.List(), nil
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
