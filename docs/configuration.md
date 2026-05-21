# Configuration

## Files

```text
/etc/immich-frame/config.toml
/var/lib/immich-frame/secrets.json
/var/lib/immich-frame/state.json
/var/lib/immich-frame/cache/
```

`config.toml` is the source of truth for non-secret settings. The setup/settings portal reads and writes this file.

`secrets.json` stores Immich API key and admin password hash. It must be permissioned `0600` and owned by the `immich-frame` service user.

`state.json` stores runtime recovery state such as current asset, setup stage, first-boot setup code, last sync, and last errors.

## Permissions

```text
config.toml:
  contains no secrets
  readable/writable by immich-frame service as needed

secrets.json:
  owner immich-frame
  mode 0600
```

## Example Config

```toml
[device]
name = "Kitchen Frame"
timezone = "auto"

[server]
host = "0.0.0.0"
port = 8787
hostname = "frame.local"

[immich]
url = "https://immich.example.com"

[source]
mode = "album"

[source.album]
id = ""
shuffle = true

[source.random]
shuffle = true

[filters]
photos_only = true
exclude_archived = true
exclude_hidden = true
exclude_trashed = true
exclude_videos = true

[display]
orientation = "auto"
width = 0
height = 0
fit = "contain" # contain | cover
background = "blur" # blur | black
transition = "crossfade" # crossfade | cut
transition_ms = 800

[slideshow]
interval_seconds = 30
recent_history_limit = 100

[cache]
preset = "balanced" # light | balanced | large | custom
max_size_mb = 2048
min_free_mb = 1024
target_items = 500
prefetch_items = 20
rendition = "auto" # auto | thumbnail | preview | original

[sync]
refresh_interval_minutes = 360

[overlays.clock]
enabled = true
slot = "top-right"
visibility = "always"
timezone = "auto"
format = "12h"
color = "auto"

[overlays.photo_info]
enabled = true
slot = "bottom-left"
visibility = "on-photo-change"
show_date = true
show_source = true
show_location = false
color = "auto"

[overlays.status]
enabled = true
slot = "bottom-center"
visibility = "when-degraded"

[weather]
enabled = false
provider = ""
location = ""
units = "imperial"
refresh_minutes = 60
```

## Secrets Example

```json
{
  "immichApiKey": "",
  "adminPasswordHash": ""
}
```

Admin password must be stored as a hash, not plain text.

Current MVP code uses PBKDF2-HMAC-SHA256 with a per-password salt because it is available from the Go standard library without adding a native/runtime dependency. A future hardening slice may move this to bcrypt or Argon2id.

## Setup Code

First-boot setup code is generated once and remains fixed until setup completes. After setup completes, it is invalidated. Reset generates a new setup code.

The setup code and setup status are persisted in `state.json`. The localhost kiosk `/frame` state may include the active setup code so the HDMI screen can display it. LAN callers can read setup status but do not receive the code from API state or SSE.

## Settings API

The setup/settings portal treats `config.toml` as the source of truth for non-secret settings and `secrets.json` as the source of truth for credentials.

Implemented browser-facing behavior:

- `GET /api/settings` returns sanitized config plus `hasImmichApiKey`.
- `PUT /api/settings` writes non-secret settings to `config.toml`.
- `immichApiKey` is replace-only. The raw saved key is never returned.
- The server preserves service/network-only fields such as listen host and local development source path when settings are written from the portal.
- Album mode stores one selected album id in `[source.album]`.
- Random mode stores `source.mode = "random"`.
- Display fit, slide interval, cache preset, and overlay enabled flags are editable from the portal.

## Reset Behavior

Factory reset deletes cached photos by default because the frame may be leaving the owner's control.

Troubleshooting flows may explicitly keep the cache:

```text
sudo immich-frame reset --keep-cache
```

Settings UI should distinguish:

```text
Reset Immich connection
  clears connection/source configuration
  may keep cached photos

Factory reset
  clears config, secrets, state, and cached photos by default
```

## Overlay Config

Overlay-specific schema/defaults live in the frontend for MVP. Backend validates only generic envelope fields:

- `enabled`.
- `slot`.
- `visibility`.

The backend should preserve overlay-specific options where possible.

## Cache Presets

Expose plain-language presets first:

```text
Light
  less SD card usage, less offline history

Balanced
  recommended, about 2 GB default

Large
  better offline resilience, more storage/background downloads
```

Advanced config can expose raw cache values.
