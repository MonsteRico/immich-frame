package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/MonsteRico/immich-frame/internal/app"
	"github.com/MonsteRico/immich-frame/internal/cache"
	"github.com/MonsteRico/immich-frame/internal/config"
)

const version = "0.1.0-dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usage()
	}
	switch args[0] {
	case "serve":
		return serve(args[1:])
	case "status":
		return status(args[1:])
	case "reset":
		return reset(args[1:])
	case "config":
		return configCommand(args[1:])
	case "version":
		fmt.Println(version)
		return nil
	case "-h", "--help", "help":
		return usage()
	default:
		return usage()
	}
}

func serve(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "", "path to config.toml")
	dataDir := fs.String("data-dir", "", "runtime data directory")
	devSource := fs.String("dev-source", "", "local folder of mock photos")
	frameDist := fs.String("frame-dist", filepath.FromSlash("ui/frame/dist"), "built frame UI directory")
	setupDist := fs.String("setup-dist", filepath.FromSlash("ui/setup/dist"), "built setup UI directory")
	logs := fs.Bool("logs", false, "log cache refresh and playback activity")
	if err := fs.Parse(args); err != nil {
		return err
	}
	application, err := app.New(app.Options{
		ConfigPath: *configPath,
		DataDir:    *dataDir,
		DevSource:  *devSource,
		FrameDist:  *frameDist,
		SetupDist:  *setupDist,
		Logs:       *logs,
	})
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return application.Serve(ctx)
}

func status(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	configPath := fs.String("config", "config.dev.toml", "path to config.toml")
	dataDir := fs.String("data-dir", ".immich-frame", "runtime data directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, configErr := config.Load(*configPath)
	configOK := configErr == nil
	statePath := filepath.Join(*dataDir, "state.json")
	state, err := config.LoadState(statePath)
	if err != nil {
		return err
	}
	secrets, err := config.LoadSecrets(filepath.Join(*dataDir, "secrets.json"))
	if err != nil {
		return err
	}
	store, err := cache.Open(filepath.Join(*dataDir, "cache"))
	if err != nil {
		return err
	}
	validationCurrent := state.ImmichValidation.Matches(cfg.Immich.URL, secrets.ImmichAPIKey)
	fmt.Printf("configured=%t\n", state.SetupComplete)
	fmt.Printf("setup_complete=%t\n", state.SetupComplete)
	fmt.Printf("setup_status=%s\n", nonEmpty(state.SetupStatus, "unknown"))
	fmt.Printf("config_valid=%t\n", configOK)
	if !configOK {
		fmt.Printf("config_error=%s\n", configErr)
	}
	fmt.Printf("source_mode=%s\n", cfg.Source.Mode)
	if cfg.Source.Mode == "album" {
		fmt.Printf("album_configured=%t\n", cfg.Source.Album.ID != "")
	}
	fmt.Printf("immich_url_configured=%t\n", cfg.Immich.URL != "")
	fmt.Printf("immich_api_key_configured=%t\n", secrets.ImmichAPIKey != "")
	fmt.Printf("immich_validation_current=%t\n", validationCurrent)
	fmt.Printf("cache_count=%d\n", len(store.List()))
	if state.CurrentAssetID != "" {
		fmt.Printf("current_asset=%s\n", state.CurrentAssetID)
	}
	if !state.LastSync.IsZero() {
		fmt.Printf("last_sync=%s\n", state.LastSync.Format(time.RFC3339))
	}
	if state.LastError != "" {
		fmt.Printf("last_error=%s\n", state.LastError)
	}
	return nil
}

func reset(args []string) error {
	fs := flag.NewFlagSet("reset", flag.ExitOnError)
	configPath := fs.String("config", "", "optional config.toml path to remove")
	dataDir := fs.String("data-dir", ".immich-frame", "runtime data directory")
	keepCache := fs.Bool("keep-cache", false, "preserve cached media")
	if err := fs.Parse(args); err != nil {
		return err
	}
	targets := []string{
		filepath.Join(*dataDir, "secrets.json"),
		filepath.Join(*dataDir, "state.json"),
	}
	if !*keepCache {
		targets = append(targets, filepath.Join(*dataDir, "cache"))
	}
	if *configPath != "" {
		targets = append(targets, *configPath)
	}
	for _, target := range targets {
		if err := os.RemoveAll(target); err != nil {
			return err
		}
	}
	fmt.Println("reset complete")
	return nil
}

func configCommand(args []string) error {
	if len(args) == 0 || args[0] != "validate" {
		return fmt.Errorf("usage: immich-frame config validate [-config path]")
	}
	fs := flag.NewFlagSet("config validate", flag.ExitOnError)
	configPath := fs.String("config", "config.dev.toml", "path to config.toml")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	fmt.Println("config valid")
	return nil
}

func usage() error {
	fmt.Println(`immich-frame commands:
  serve             run the local frame daemon
  status            print runtime status without secrets
  reset             clear secrets, state, and cache unless --keep-cache is set
  config validate   validate config.toml
  version           print version`)
	return nil
}

func nonEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
