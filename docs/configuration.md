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

`state.json` stores runtime recovery state such as current asset, setup stage, first-boot setup code, Immich validation status, last sync, and last errors.

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
preset = "balanced" # extra-small | light | balanced | large | custom
max_size_mb = 2048
min_free_mb = 1024
target_items = 500
prefetch_items = 20
rendition = "auto" # auto | webp | jpeg

[sync]
refresh_interval_minutes = 360

[overlays.clock]
enabled = true
slot = "top-right"
visibility = "always"

[overlays.photo_info]
enabled = true
slot = "bottom-left"
visibility = "on-photo-change"

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

## Immich Validation State

Setup completion requires a successful Immich connection test for the saved Immich URL and API key. The daemon stores validation metadata in `state.json`:

- validation boolean.
- normalized Immich URL.
- SHA-256 fingerprint of the API key, never the raw key.
- validation timestamp.
- Immich version and API key name when returned by Immich.

Changing the saved Immich URL or replacing the API key clears the previous validation unless the new URL/key pair already matches a successful validation. This keeps random-library mode from bypassing the same connection check required by album mode.

## Settings API

The setup/settings portal treats `config.toml` as the source of truth for non-secret settings and `secrets.json` as the source of truth for credentials.

Implemented browser-facing behavior:

- `GET /api/settings` returns sanitized config plus `hasImmichApiKey` and the lightweight status payload.
- `PUT /api/settings` writes non-secret settings to `config.toml`.
- `GET /api/status` returns setup/configuration status, Immich validation status, source mode, cache count, and last error for setup/admin sessions.
- `immichApiKey` is replace-only. The raw saved key is never returned.
- The server preserves service/network-only fields such as listen host and local development source path when settings are written from the portal.
- Album mode stores one selected album id in `[source.album]`.
- Random mode stores `source.mode = "random"`.
- Display fit, slide interval, cache preset, and overlay enabled flags are editable from the portal.

## Config Validation

`immich-frame config validate` checks the browser MVP settings needed for a usable frame:

- server port.
- source mode and required source fields.
- Immich URL for album/random-library modes.
- display fit and transition.
- positive slideshow interval.
- cache size, target, prefetch, and rendition values.
- sync refresh interval.
- overlay slots and visibility values.

## Reset Behavior

Factory reset deletes cached photos by default because the frame may be leaving the owner's control.

Troubleshooting flows may explicitly keep the cache:

```text
sudo immich-frame reset --keep-cache
```

The CLI always clears `secrets.json` and `state.json` under the selected data directory. It clears cached media unless `--keep-cache` is set. Pass `--config /path/to/config.toml` when the reset should also remove the non-secret config file, such as a factory reset on a dedicated frame device.

Settings UI should distinguish:

```text
Reset Immich connection
  clears connection/source configuration
  may keep cached photos

Factory reset
  clears secrets, state, cached photos, and config when --config is provided
```

## Overlay Config

MVP overlay settings are intentionally limited to the generic envelope fields currently read and written by the backend:

- `enabled`.
- `slot`.
- `visibility`.

Future overlay-specific settings such as clock format, timezone, color, or photo-info field selection should either be explicitly added to the config model or stored in a preserved raw overlay options structure. Do not document new overlay-specific TOML fields as implemented until the backend can round-trip them.

## Cache Presets

Expose plain-language presets first:

```text
Extra small
  local testing preset, roughly 10 cached photos so cache rotation is easy to see

Light
  less SD card usage, less offline history

Balanced
  recommended, about 2 GB default

Large
  better offline resilience, more storage/background downloads
```

Advanced config can expose raw cache values.

When a TOML file sets `cache.preset`, the daemon applies that preset as the starting cache shape while parsing the file. Explicit raw values such as `target_items` or `prefetch_items` later in the same `[cache]` table still override the preset. The `extra-small` preset is for development visibility only and is not the production default.

## Cache Rotation And Sync

`cache.target_items` is the warm-cache goal for display-targeted renditions. The daemon periodically refreshes Immich candidates for album and random-library sources and downloads missing renditions until the cache approaches that target.

`cache.prefetch_items` protects the current photo and near-upcoming playback entries from eviction. This keeps the frame from deleting the next few items it is likely to show.

`sync.refresh_interval_minutes` controls the normal Immich candidate refresh interval. When Immich is unavailable, the daemon keeps playing from cache when possible and retries refresh with bounded backoff. The last user-safe refresh error is stored in `state.json` and exposed through status surfaces without secrets.

Eviction favors assets that are no longer in the selected Immich source before removing valid fallback photos. If more valid photos still need to be removed, recently shown entries are less valuable than never-shown or least-recently-shown entries.
