package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultHost     = "0.0.0.0"
	DefaultPort     = 8787
	DefaultHostname = "frame.local"
)

type Paths struct {
	ConfigFile  string
	SecretsFile string
	StateFile   string
	CacheDir    string
}

func AppliancePaths() Paths {
	return Paths{
		ConfigFile:  "/etc/immich-frame/config.toml",
		SecretsFile: "/var/lib/immich-frame/secrets.json",
		StateFile:   "/var/lib/immich-frame/state.json",
		CacheDir:    "/var/lib/immich-frame/cache",
	}
}

func DevPaths(root string) Paths {
	dataDir := filepath.Join(root, ".immich-frame")
	return Paths{
		ConfigFile:  filepath.Join(root, "config.dev.toml"),
		SecretsFile: filepath.Join(dataDir, "secrets.json"),
		StateFile:   filepath.Join(dataDir, "state.json"),
		CacheDir:    filepath.Join(dataDir, "cache"),
	}
}

type Config struct {
	Device    DeviceConfig    `json:"device"`
	Server    ServerConfig    `json:"server"`
	Immich    ImmichConfig    `json:"immich"`
	Source    SourceConfig    `json:"source"`
	Filters   FilterConfig    `json:"filters"`
	Display   DisplayConfig   `json:"display"`
	Slideshow SlideshowConfig `json:"slideshow"`
	Cache     CacheConfig     `json:"cache"`
	Sync      SyncConfig      `json:"sync"`
	Overlays  OverlayConfig   `json:"overlays"`
	Weather   WeatherConfig   `json:"weather"`
}

type DeviceConfig struct {
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
}

type ServerConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Hostname string `json:"hostname"`
}

type ImmichConfig struct {
	URL string `json:"url"`
}

type SourceConfig struct {
	Mode        string             `json:"mode"`
	Album       AlbumSourceConfig  `json:"album"`
	Random      RandomSourceConfig `json:"random"`
	LocalFolder LocalFolderConfig  `json:"localFolder"`
}

type AlbumSourceConfig struct {
	ID      string `json:"id"`
	Shuffle bool   `json:"shuffle"`
}

type RandomSourceConfig struct {
	Shuffle bool `json:"shuffle"`
}

type LocalFolderConfig struct {
	Path    string `json:"path"`
	Shuffle bool   `json:"shuffle"`
}

type FilterConfig struct {
	PhotosOnly      bool `json:"photosOnly"`
	ExcludeArchived bool `json:"excludeArchived"`
	ExcludeHidden   bool `json:"excludeHidden"`
	ExcludeTrashed  bool `json:"excludeTrashed"`
	ExcludeVideos   bool `json:"excludeVideos"`
}

type DisplayConfig struct {
	Orientation  string `json:"orientation"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Fit          string `json:"fit"`
	Background   string `json:"background"`
	Transition   string `json:"transition"`
	TransitionMS int    `json:"transitionMs"`
}

type SlideshowConfig struct {
	IntervalSeconds    int `json:"intervalSeconds"`
	RecentHistoryLimit int `json:"recentHistoryLimit"`
}

type CacheConfig struct {
	Preset        string `json:"preset"`
	MaxSizeMB     int    `json:"maxSizeMb"`
	MinFreeMB     int    `json:"minFreeMb"`
	TargetItems   int    `json:"targetItems"`
	PrefetchItems int    `json:"prefetchItems"`
	Rendition     string `json:"rendition"`
}

type SyncConfig struct {
	RefreshIntervalMinutes int `json:"refreshIntervalMinutes"`
}

type OverlayEnvelope struct {
	Enabled    bool   `json:"enabled"`
	Slot       string `json:"slot"`
	Visibility string `json:"visibility"`
}

type OverlayConfig struct {
	Clock     OverlayEnvelope `json:"clock"`
	PhotoInfo OverlayEnvelope `json:"photoInfo"`
	Status    OverlayEnvelope `json:"status"`
}

type WeatherConfig struct {
	Enabled        bool   `json:"enabled"`
	Provider       string `json:"provider"`
	Location       string `json:"location"`
	Units          string `json:"units"`
	RefreshMinutes int    `json:"refreshMinutes"`
}

type Secrets struct {
	ImmichAPIKey      string `json:"immichApiKey"`
	AdminPasswordHash string `json:"adminPasswordHash"`
}

type State struct {
	SetupComplete  bool      `json:"setupComplete"`
	SetupStatus    string    `json:"setupStatus,omitempty"`
	SetupCode      string    `json:"setupCode,omitempty"`
	CurrentAssetID string    `json:"currentAssetId,omitempty"`
	LastSync       time.Time `json:"lastSync,omitempty"`
	LastError      string    `json:"lastError,omitempty"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func DefaultConfig() Config {
	return Config{
		Device: DeviceConfig{Name: "Immich Frame", Timezone: "auto"},
		Server: ServerConfig{Host: DefaultHost, Port: DefaultPort, Hostname: DefaultHostname},
		Source: SourceConfig{
			Mode:        "local_folder",
			Album:       AlbumSourceConfig{Shuffle: true},
			Random:      RandomSourceConfig{Shuffle: true},
			LocalFolder: LocalFolderConfig{Path: "./dev/photos", Shuffle: true},
		},
		Filters: FilterConfig{
			PhotosOnly: true, ExcludeArchived: true, ExcludeHidden: true, ExcludeTrashed: true, ExcludeVideos: true,
		},
		Display: DisplayConfig{
			Orientation: "auto", Fit: "contain", Background: "blur", Transition: "crossfade", TransitionMS: 800,
		},
		Slideshow: SlideshowConfig{IntervalSeconds: 30, RecentHistoryLimit: 100},
		Cache: CacheConfig{
			Preset: "balanced", MaxSizeMB: 2048, MinFreeMB: 1024, TargetItems: 500, PrefetchItems: 20, Rendition: "auto",
		},
		Sync: SyncConfig{RefreshIntervalMinutes: 360},
		Overlays: OverlayConfig{
			Clock:     OverlayEnvelope{Enabled: true, Slot: "top-right", Visibility: "always"},
			PhotoInfo: OverlayEnvelope{Enabled: true, Slot: "bottom-left", Visibility: "on-photo-change"},
			Status:    OverlayEnvelope{Enabled: true, Slot: "bottom-center", Visibility: "when-degraded"},
		},
		Weather: WeatherConfig{Units: "imperial", RefreshMinutes: 60},
	}
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c Config) Validate() error {
	var issues []string
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		issues = append(issues, "server.port must be between 1 and 65535")
	}
	if c.Display.Fit != "contain" && c.Display.Fit != "cover" {
		issues = append(issues, "display.fit must be contain or cover")
	}
	if c.Display.Transition != "crossfade" && c.Display.Transition != "cut" {
		issues = append(issues, "display.transition must be crossfade or cut")
	}
	if c.Slideshow.IntervalSeconds <= 0 {
		issues = append(issues, "slideshow.interval_seconds must be positive")
	}
	switch c.Source.Mode {
	case "local_folder", "album", "random":
	default:
		issues = append(issues, "source.mode must be local_folder, album, or random")
	}
	if len(issues) > 0 {
		return errors.New(strings.Join(issues, "; "))
	}
	return nil
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	table := ""
	for _, raw := range strings.Split(string(data), "\n") {
		line := stripComment(strings.TrimSpace(raw))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			table = strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		assign(&cfg, table, key, value)
	}
	return cfg, cfg.Validate()
}

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data := []byte(cfg.TOML())
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (c Config) TOML() string {
	return fmt.Sprintf(`[device]
name = %q
timezone = %q

[server]
host = %q
port = %d
hostname = %q

[immich]
url = %q

[source]
mode = %q

[source.album]
id = %q
shuffle = %t

[source.random]
shuffle = %t

[source.local_folder]
path = %q
shuffle = %t

[filters]
photos_only = %t
exclude_archived = %t
exclude_hidden = %t
exclude_trashed = %t
exclude_videos = %t

[display]
orientation = %q
width = %d
height = %d
fit = %q
background = %q
transition = %q
transition_ms = %d

[slideshow]
interval_seconds = %d
recent_history_limit = %d

[cache]
preset = %q
max_size_mb = %d
min_free_mb = %d
target_items = %d
prefetch_items = %d
rendition = %q

[sync]
refresh_interval_minutes = %d

[overlays.clock]
enabled = %t
slot = %q
visibility = %q

[overlays.photo_info]
enabled = %t
slot = %q
visibility = %q

[overlays.status]
enabled = %t
slot = %q
visibility = %q
`,
		c.Device.Name, c.Device.Timezone,
		c.Server.Host, c.Server.Port, c.Server.Hostname,
		c.Immich.URL,
		c.Source.Mode,
		c.Source.Album.ID, c.Source.Album.Shuffle,
		c.Source.Random.Shuffle,
		c.Source.LocalFolder.Path, c.Source.LocalFolder.Shuffle,
		c.Filters.PhotosOnly, c.Filters.ExcludeArchived, c.Filters.ExcludeHidden, c.Filters.ExcludeTrashed, c.Filters.ExcludeVideos,
		c.Display.Orientation, c.Display.Width, c.Display.Height, c.Display.Fit, c.Display.Background, c.Display.Transition, c.Display.TransitionMS,
		c.Slideshow.IntervalSeconds, c.Slideshow.RecentHistoryLimit,
		c.Cache.Preset, c.Cache.MaxSizeMB, c.Cache.MinFreeMB, c.Cache.TargetItems, c.Cache.PrefetchItems, c.Cache.Rendition,
		c.Sync.RefreshIntervalMinutes,
		c.Overlays.Clock.Enabled, c.Overlays.Clock.Slot, c.Overlays.Clock.Visibility,
		c.Overlays.PhotoInfo.Enabled, c.Overlays.PhotoInfo.Slot, c.Overlays.PhotoInfo.Visibility,
		c.Overlays.Status.Enabled, c.Overlays.Status.Slot, c.Overlays.Status.Visibility,
	)
}

func stripComment(line string) string {
	if idx := strings.Index(line, "#"); idx >= 0 {
		return strings.TrimSpace(line[:idx])
	}
	return line
}

func assign(cfg *Config, table, key, value string) {
	stringValue := strings.Trim(value, `"`)
	boolValue := strings.EqualFold(stringValue, "true")
	intValue, _ := strconv.Atoi(stringValue)
	switch table + "." + key {
	case "device.name":
		cfg.Device.Name = stringValue
	case "device.timezone":
		cfg.Device.Timezone = stringValue
	case "server.host":
		cfg.Server.Host = stringValue
	case "server.port":
		cfg.Server.Port = intValue
	case "server.hostname":
		cfg.Server.Hostname = stringValue
	case "immich.url":
		cfg.Immich.URL = stringValue
	case "source.mode":
		cfg.Source.Mode = stringValue
	case "source.album.id":
		cfg.Source.Album.ID = stringValue
	case "source.album.shuffle":
		cfg.Source.Album.Shuffle = boolValue
	case "source.random.shuffle":
		cfg.Source.Random.Shuffle = boolValue
	case "source.local_folder.path":
		cfg.Source.LocalFolder.Path = stringValue
	case "source.local_folder.shuffle":
		cfg.Source.LocalFolder.Shuffle = boolValue
	case "filters.photos_only":
		cfg.Filters.PhotosOnly = boolValue
	case "filters.exclude_archived":
		cfg.Filters.ExcludeArchived = boolValue
	case "filters.exclude_hidden":
		cfg.Filters.ExcludeHidden = boolValue
	case "filters.exclude_trashed":
		cfg.Filters.ExcludeTrashed = boolValue
	case "filters.exclude_videos":
		cfg.Filters.ExcludeVideos = boolValue
	case "display.width":
		cfg.Display.Width = intValue
	case "display.height":
		cfg.Display.Height = intValue
	case "display.orientation":
		cfg.Display.Orientation = stringValue
	case "display.fit":
		cfg.Display.Fit = stringValue
	case "display.background":
		cfg.Display.Background = stringValue
	case "display.transition":
		cfg.Display.Transition = stringValue
	case "display.transition_ms":
		cfg.Display.TransitionMS = intValue
	case "slideshow.interval_seconds":
		cfg.Slideshow.IntervalSeconds = intValue
	case "slideshow.recent_history_limit":
		cfg.Slideshow.RecentHistoryLimit = intValue
	case "cache.preset":
		cfg.Cache.Preset = stringValue
	case "cache.max_size_mb":
		cfg.Cache.MaxSizeMB = intValue
	case "cache.min_free_mb":
		cfg.Cache.MinFreeMB = intValue
	case "cache.target_items":
		cfg.Cache.TargetItems = intValue
	case "cache.prefetch_items":
		cfg.Cache.PrefetchItems = intValue
	case "cache.rendition":
		cfg.Cache.Rendition = stringValue
	case "sync.refresh_interval_minutes":
		cfg.Sync.RefreshIntervalMinutes = intValue
	case "overlays.clock.enabled":
		cfg.Overlays.Clock.Enabled = boolValue
	case "overlays.clock.slot":
		cfg.Overlays.Clock.Slot = stringValue
	case "overlays.clock.visibility":
		cfg.Overlays.Clock.Visibility = stringValue
	case "overlays.photo_info.enabled":
		cfg.Overlays.PhotoInfo.Enabled = boolValue
	case "overlays.photo_info.slot":
		cfg.Overlays.PhotoInfo.Slot = stringValue
	case "overlays.photo_info.visibility":
		cfg.Overlays.PhotoInfo.Visibility = stringValue
	case "overlays.status.enabled":
		cfg.Overlays.Status.Enabled = boolValue
	case "overlays.status.slot":
		cfg.Overlays.Status.Slot = stringValue
	case "overlays.status.visibility":
		cfg.Overlays.Status.Visibility = stringValue
	}
}

func LoadState(path string) (State, error) {
	state := State{UpdatedAt: time.Now()}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state, nil
		}
		return state, err
	}
	return state, json.Unmarshal(data, &state)
}

func SaveState(path string, state State) error {
	state.UpdatedAt = time.Now()
	return writeJSONAtomic(path, state, 0o644)
}

func LoadSecrets(path string) (Secrets, error) {
	var secrets Secrets
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return secrets, nil
		}
		return secrets, err
	}
	return secrets, json.Unmarshal(data, &secrets)
}

func SaveSecrets(path string, secrets Secrets) error {
	return writeJSONAtomic(path, secrets, 0o600)
}

func writeJSONAtomic(path string, value any, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
