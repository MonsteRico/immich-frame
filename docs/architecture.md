# Architecture

## System Shape

Immich Frame runs locally on a frame device.

```text
Raspberry Pi Zero 2 W
  -> immich-frame Go daemon
  -> embedded static Preact UI
  -> Chromium kiosk at localhost
  -> outbound HTTPS to Immich
  -> local display-sized image cache
```

The browser is a renderer, not the owner of app state. The daemon owns configuration, secrets, Immich access, cache, playback, and state updates.

## Main Components

```text
cmd/immich-frame/
  CLI entry point

internal/api/
  HTTP routes, auth middleware, SSE, static UI serving

internal/app/
  application wiring and lifecycle

internal/auth/
  setup code, admin login, password hashing, sessions

internal/cache/
  media cache, manifest, eviction, rendition tracking

internal/config/
  config.toml, secrets.json, state.json loading/writing/validation

internal/display/
  browser-reported viewport/display target handling

internal/immich/
  isolated Immich API adapter

internal/media/
  local media serving and content types

internal/playback/
  slideshow scheduler, queue, history, commands

internal/setup/
  setup stages and setup completion flow

internal/source/
  album/random source providers, local folder dev source

internal/weather/
  future provider interface placeholder only

ui/frame/
  tiny Preact kiosk slideshow bundle

ui/setup/
  Preact setup/settings bundle

ui/shared/
  shared API types, overlay schemas, overlay registry helpers
```

## Runtime Flow

```text
Boot
  -> systemd starts immich-frame serve
  -> daemon loads config/secrets/state
  -> daemon serves local HTTP on :8787
  -> kiosk opens http://127.0.0.1:8787/frame
  -> UI requests /api/state and subscribes to /api/events
```

If unconfigured:

```text
/frame shows setup instructions and code
/setup from phone/laptop claims code and configures frame
```

If configured:

```text
daemon builds/refreshes candidate pool
daemon caches display-targeted renditions
daemon maintains playback queue
daemon emits frame state over SSE
browser renders current cached image and overlays
```

## Local API Surface

Public UI routes:

```text
GET /frame
GET /setup
GET /assets/*
```

Frame runtime API:

```text
GET  /api/events
GET  /api/state
POST /api/playback/next
POST /api/playback/previous
POST /api/playback/pause
POST /api/playback/resume
GET  /media/:assetID
```

Setup/settings API:

```text
POST /api/setup/claim
POST /api/setup/complete
POST /api/auth/login
POST /api/auth/logout
GET  /api/settings
PUT  /api/settings
POST /api/immich/test
GET  /api/immich/albums
GET  /api/status
```

`/media/:assetID` serves cached local files only. It must not proxy arbitrary Immich URLs.

## Auth Boundary

- Requests from `127.0.0.1` are trusted for kiosk slideshow/media access.
- LAN clients require setup/admin session for settings and photo/media access.
- The browser never receives Immich API keys.
- The browser never receives direct authenticated Immich image URLs.

## State Updates

Use SSE for meaningful state changes.

Daemon-owned loops:

```text
slideshow tick       every configured interval
weather refresh      future, about hourly
Immich refresh       default about every 6 hours
cache maintenance    background/idle
settings watcher     on change
```

Browser-owned display timer:

```text
clock text updates locally every minute
```

## Renderer

Primary renderer is Chromium kiosk with static Preact UI.

Rules:

- Keep `/frame` bundle lean.
- Separate setup/settings bundle from frame bundle.
- No full SPA framework patterns beyond simple Preact composition.
- No client-side image processing.
- Use two image elements for crossfade.
- Avoid animation libraries and large UI kits.
- Use plain CSS files or CSS modules.
- Do not use Tailwind or a component library for MVP.
- Use minimal inline SVG icons for controls where helpful.
- Do not use a full icon library for MVP.

## Visual Direction

The frame UI should be restrained and photo-first:

- full-bleed photo presentation.
- blurred or black background for aspect-ratio gaps.
- small low-contrast overlay surfaces.
- controls hidden unless interacting.
- no dashboard-heavy layout.
- no decorative chrome competing with photos.
- no theme system in MVP.
- status messaging stays quiet unless the daemon reports a degraded/error condition or the frame UI has a local runtime error.
- configured empty-cache outages use a calm offline screen instead of development placeholder text.

## Setup UI

The setup/settings portal should be phone-first and desktop-compatible.

Setup usually happens near the physical frame from a phone, so prioritize:

- single-column forms.
- large tap targets.
- raw Immich URL and raw Immich API key fields.
- easy paste behavior for Immich URL/API key.
- clear connection test feedback.
- searchable album picker usable on mobile.
- lightweight settings, not a dense admin dashboard.

## Overlay System

Overlay contribution model is source-level, reviewed, and compiled into the app. No runtime third-party plugin loading for MVP.

Overlay layout is slot-based:

```text
top-left
top-center
top-right
middle-left
center
middle-right
bottom-left
bottom-center
bottom-right
```

MVP overlays:

- Clock.
- Photo info.
- Operational status.

Near-future overlay:

- Weather.

MVP overlay config is limited to the generic envelope fields currently read and written by the backend: `enabled`, `slot`, and `visibility`. The backend does not currently preserve arbitrary overlay-specific TOML fields. Future overlay-specific settings should be added explicitly to the config model or through a deliberate raw-options preservation structure before docs describe them as supported.

## Source Model

Do not hardcode source behavior as only an album ID. Use source providers.

MVP providers:

- One Immich album.
- Random library photos.
- Local folder source for development only.

Future providers:

- Favorites.
- On this day.
- People/person.
- Location.
- Date range.
- Smart rules.
- Mixed/weighted albums.

## Cache Model

Cache display-targeted renditions, not originals by default.

Default behavior:

- Detect display size from browser report.
- Default fallback is 1920x1080.
- Fetch the best Immich-provided non-original rendition for the display target.
- Store whatever content type Immich returns.
- Do not re-encode in MVP.
- Originals are opt-in fallback only.

Cache default target:

```toml
[cache]
max_size_mb = 2048
min_free_mb = 1024
target_items = 500
prefetch_items = 20
```

Use one JSON manifest with atomic write/rename.

Implemented rotation behavior:

- The cache manifest tracks `LastShown`.
- Cached media is listed with unshown and least-recently-shown entries first.
- The playback queue is seeded from the cache at startup and refreshed when cache contents change.
- `target_items` is the desired warm cache size, not a fixed forever playlist.
- `prefetch_items` defines near-upcoming playback entries that are protected from eviction.
- The daemon periodically refreshes the Immich candidate pool for album and random-library sources.
- The daemon tops off display-targeted renditions toward `target_items`, preferring never-cached candidates first.
- Stable album sources larger than `target_items` also churn one unprotected cache slot per refresh once the cache is full. Rotation uses manifest history to prefer candidates that have never been cached, then candidates least recently shown/cached, so the initial album seed does not remain fixed forever.
- Eviction prefers cached assets that have left the selected source, then recently shown valid assets, while preserving current and near-upcoming playback entries.
- Existing cached media remains playable during Immich/network outages while the daemon retries refresh with bounded backoff.

## Playback Model

The daemon owns playback state.

State:

- current asset.
- in-memory previous history, roughly 20-50 items.
- upcoming queue.
- candidate pool.

Behavior:

- Start slideshow as soon as first few cached images are ready.
- If the frame boots while Immich is unavailable but cached photos exist, start slideshow from cache immediately and retry Immich in the background.
- If configured but cache is empty and Immich is unavailable, show a calm retry screen with settings URL, no cached photos message, retry status, and a small last-error detail.
- Avoid recent repeats.
- Prefer least-recently-shown candidates.
- Rebuild or refresh the queue as the cache changes so playback does not loop one static seed forever.
- Keep the current photo stable when a background Immich refresh or rendition fetch fails.
- No strict full-library cycle guarantee.
- Resume current/last cached photo after reboot if possible.
- Previous history does not need to survive reboot.
