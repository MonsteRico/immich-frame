# Security

Immich Frame displays private photos. Treat it as a private-photo appliance.

## Core Rules

- Browser never receives Immich API key.
- Browser never receives direct authenticated Immich image URLs.
- Browser receives only local cached media URLs and sanitized metadata.
- Daemon owns Immich API access.
- LAN clients cannot view cached photos without authentication.
- Settings portal requires admin auth after setup.
- Use a dedicated Immich API key per frame.

## Setup Security

First boot:

1. HDMI screen shows setup URL and setup code.
2. User visits setup portal from same Wi-Fi.
3. Setup portal requires code.
4. User creates local admin password.
5. User enters Immich URL and dedicated API key.
6. Daemon validates the Immich URL/API key and stores settings.
7. Setup code is invalidated.

After setup:

- Admin password is required for settings.
- Admin password is stored as a hash in `secrets.json`.
- Immich API key is stored in `secrets.json`.
- Settings UI never reveals the saved Immich API key.
- Settings UI allows replacing the Immich API key by pasting a new raw key.
- Setup completion requires a successful Immich connection test for the saved URL/API key.
- Random-library mode has the same validation requirement as album mode.
- Admin sessions last about 30 minutes by default and renew on activity.
- MVP does not include a remember-me session.

Implemented route shape:

- `GET /api/setup/state` returns setup status. It does not reveal the setup code to LAN callers.
- `POST /api/setup/claim` accepts the setup code and creates a setup-scoped session.
- `POST /api/setup/admin-password` stores only a password hash and creates an admin session.
- `POST /api/auth/login` and `POST /api/auth/logout` manage admin sessions after setup.
- `GET /api/settings` and `PUT /api/settings` require a setup or admin session.
- `GET /api/status` requires a setup or admin session and returns setup/configuration status, Immich validation status, source mode, cache count, and last error without raw secrets.
- `POST /api/immich/test` and `GET /api/immich/albums` require a setup or admin session.
- `POST /api/setup/complete` invalidates the setup code only after admin password, saved Immich credentials, successful validation for those credentials, and source selection are present.

Runtime refresh errors are recorded as user-safe status text only. They must not include the raw Immich API key, direct authenticated Immich URLs, filesystem paths, or raw Immich response bodies.

## Localhost Trust Boundary

The local browser frame opens `http://127.0.0.1:8787/frame`.

Requests from localhost may access the frame and cached media so the device can boot into slideshow without login.

Requests from LAN clients must authenticate before accessing settings or photo media.

LAN callers without an admin session cannot fetch `/media/:assetID`. The localhost kiosk may fetch cached media without login so the appliance can boot directly into the slideshow.

Cache eviction preserves the current and near-upcoming playback entries before removing stale or over-target assets. Factory reset remains the privacy boundary for intentionally clearing cached photos.

## Local HTTP

MVP uses HTTP for the local setup/settings portal.

This assumes setup happens on a trusted same-Wi-Fi network. Local HTTPS with self-signed certificates is intentionally out of MVP because it creates certificate warnings and setup friction.

## Immich URL Scheme

Setup should prefer HTTPS Immich URLs.

HTTP Immich URLs may be allowed for local homelab deployments, but the setup UI must show a clear warning that the frame's Immich API key will be sent to Immich over HTTP and should only be used on trusted local networks.

## Physical Device Risk

MVP assumes basic practical protection, not strong protection against physical attackers.

Protections:

- API key stored in permissioned secrets file.
- Admin password hashed.
- Browser does not expose credentials.
- LAN photo access is authenticated.

Password hashes currently use PBKDF2-HMAC-SHA256 with a random salt and high iteration count. This avoids storing plain text and keeps the MVP dependency set small; bcrypt or Argon2id remain future hardening options.

Not planned:

- Full disk encryption.
- Secure boot.
- Hardware-backed secrets.
- Remote wipe.

## Lost Device Procedure

If a frame is lost, sold, stolen, or replaced:

1. Revoke that frame's dedicated Immich API key in Immich.
2. If possible, run `sudo immich-frame reset` before giving away the device.
3. Factory reset clears cached photos by default.

For troubleshooting only, a CLI reset may offer an explicit keep-cache option, but privacy-preserving reset is the default.

`immich-frame reset` clears secrets and state by default and removes cached media unless `--keep-cache` is explicitly set. Config removal is explicit with `--config /path/to/config.toml` so local development configs are not deleted accidentally.

## Metadata Minimization

Browser receives minimal display metadata by default:

- asset id.
- local media URL.
- taken/display date.
- source name if enabled.
- dimensions/orientation.
- favorite flag only if needed.

Avoid sending by default:

- original filename.
- raw EXIF.
- GPS coordinates.
- people names.
- full Immich asset JSON.
- filesystem paths.
- direct Immich URLs.

Richer metadata can be added later only through explicit overlay settings.
