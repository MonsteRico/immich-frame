# Local Development And Verification

This runbook is for developing and manually verifying Immich Frame on a desktop
without a Raspberry Pi and without a real Immich server.

The local mock mode uses `dev/photos` as the photo source. It is a development
tool only; the MVP user-facing sources remain Immich album and random-library
modes.

## Prerequisites

- Go 1.22 or newer.
- Node.js from the active PowerShell environment.
- pnpm from the active PowerShell environment.
- A desktop browser. Chromium or Chrome is closest to the target kiosk runtime.

Check tool availability:

```sh
go version
node --version
pnpm --version
```

Matthew's PowerShell profile initializes fnm, so `node` and `pnpm` should work directly in agent shells.

## Install Frontend Dependencies

From the repository root:

```sh
pnpm install
```

If dependencies are already installed, this should be a no-op or a quick
lockfile-confirming run.

## Run The Standard Checks

Run these before committing code that touches the daemon or frontend:

```sh
go test ./...
pnpm typecheck
pnpm build
```

The default CI scope for the MVP base is:

- Go unit tests.
- Frontend TypeScript typecheck.
- Frontend production build.

## Verify Embedded Release UI Assets

Release binaries embed the built Vite assets from `internal/api/static`.

Before a release-style Go build or embedded UI smoke test, run:

```sh
pnpm build:embedded-ui
```

That command:

1. Builds `ui/shared`, `ui/frame`, and `ui/setup`.
2. Copies `ui/frame/dist` to `internal/api/static/frame`.
3. Copies `ui/setup/dist` to `internal/api/static/setup`.

The embedded `index.html` files intentionally reference root-relative
`/assets/*` URLs. The Go server first checks external development dist
directories, then falls back to embedded assets from `embed.FS`.

## Run The Mock Slideshow

From the repository root:

```sh
go run ./cmd/immich-frame serve -config config.dev.toml -dev-source dev/photos
```

Open these URLs in a desktop browser:

- `http://127.0.0.1:8787/frame`
- `http://127.0.0.1:8787/setup`

Expected `/frame` behavior:

- A full-window mock slideshow renders from `dev/photos`.
- The clock overlay appears.
- Photo info appears for the current mock photo.
- The operational status overlay stays quiet while state is ready.
- The current image URL is a local `/media/:assetID` URL, not an external URL.
- Background cache maintenance may refresh the queue when local or Immich cache contents change.
- A configured frame with no cached media and an unavailable Immich source shows the offline/no cached photos screen rather than the local development placeholder.

Expected `/setup` behavior:

- If `state.json` is unconfigured, the portal asks for the first-boot setup code shown by the frame.
- After claiming the code, the portal creates a local admin password, accepts an Immich URL/API key, validates Immich, and lets you choose album or random-library mode.
- The Immich step disables saving until the current URL/API key has passed the connection test.
- The source step disables finish for both album and random-library mode until the saved URL/API key validation is current.
- After setup is complete, `/setup` becomes the ongoing settings page and requires the admin password.
- The settings page shows lightweight status: setup state, Immich validation state, source mode, cache count, and last error.
- The settings page never displays a saved Immich API key. Paste a new key only when replacing it.

For local mock slideshow work, `-dev-source dev/photos` keeps `/frame` on the local folder source even if setup is incomplete.

Stop the server with `Ctrl+C`.

## Verify Embedded Assets Locally

To prove the server is using embedded Vite assets instead of external dist
folders, first prepare the embedded UI:

```sh
pnpm build:embedded-ui
```

Then run the daemon with intentionally missing dist paths:

```sh
go run ./cmd/immich-frame serve -config config.dev.toml -dev-source dev/photos -data-dir .immich-frame-verify -frame-dist missing-frame-dist -setup-dist missing-setup-dist
```

Open:

- `http://127.0.0.1:8787/frame`
- `http://127.0.0.1:8787/setup`

Expected result:

- Both pages load.
- Browser devtools Network shows `/assets/index-*.js` and
  `/assets/index-*.css` returning `200`.
- `/frame` shows cached local mock media.

Useful HTTP smoke checks:

```sh
curl -i http://127.0.0.1:8787/frame
curl -i http://127.0.0.1:8787/api/state
curl -i http://127.0.0.1:8787/api/setup/state
```

On PowerShell:

```powershell
Invoke-WebRequest -UseBasicParsing http://127.0.0.1:8787/frame
Invoke-WebRequest -UseBasicParsing http://127.0.0.1:8787/api/state
Invoke-WebRequest -UseBasicParsing http://127.0.0.1:8787/api/setup/state
```

After the embedded smoke test, remove the temporary verification data:

```sh
rm -rf .immich-frame-verify
```

On PowerShell:

```powershell
Remove-Item -LiteralPath .immich-frame-verify -Recurse -Force
```

## Manual Browser Verification Checklist

Use a desktop Chromium-family browser when possible.

1. Run the mock slideshow command.
2. Open `http://127.0.0.1:8787/frame`.
3. Confirm the page is not blank and no framework error overlay appears.
4. Confirm at least one mock photo from `dev/photos` is visible.
5. Confirm the clock and photo info overlays render.
6. Open browser devtools and check the Console for relevant errors.
7. Open browser devtools Network and confirm:
   - `/api/state` returns `200`.
   - `/api/events` is open as an event stream.
   - `/media/:assetID` returns `200`.
   - If testing embedded mode, `/assets/index-*.js` and
     `/assets/index-*.css` return `200`.
8. Test playback commands with HTTP requests:

```sh
curl -X POST http://127.0.0.1:8787/api/playback/pause
curl -X POST http://127.0.0.1:8787/api/playback/resume
curl -X POST http://127.0.0.1:8787/api/playback/next
curl -X POST http://127.0.0.1:8787/api/playback/previous
```

On PowerShell:

```powershell
Invoke-WebRequest -UseBasicParsing -Method Post http://127.0.0.1:8787/api/playback/pause
Invoke-WebRequest -UseBasicParsing -Method Post http://127.0.0.1:8787/api/playback/resume
Invoke-WebRequest -UseBasicParsing -Method Post http://127.0.0.1:8787/api/playback/next
Invoke-WebRequest -UseBasicParsing -Method Post http://127.0.0.1:8787/api/playback/previous
```

The response JSON should reflect the updated playback state.

## Setup Portal Smoke Checks

Use a temporary data directory when testing first boot repeatedly. To see the HDMI setup screen instead of the local mock slideshow, point `-config` at a temporary config whose `[source] mode` is `album` or `random`; `config.dev.toml` intentionally uses `local_folder` for the mock slideshow loop. One quick option is to copy `config.dev.toml` to `.setup-verify.toml` and change only `source.mode` to `album`.

```powershell
go run ./cmd/immich-frame serve -config .setup-verify.toml -data-dir .immich-frame-setup-verify -frame-dist ui/frame/dist -setup-dist ui/setup/dist
```

Open `http://127.0.0.1:8787/frame` and confirm the setup URL/code screen appears when setup is incomplete and the source is not `local_folder`.

Open `http://127.0.0.1:8787/setup` and confirm:

- the setup code claim screen renders.
- wrong codes fail.
- the code from the HDMI frame advances setup.
- admin password creation requires at least 8 characters.
- HTTP Immich URLs show a warning.
- the Immich settings step explains missing URL/key and missing validation before Save is enabled.
- random-library mode cannot finish setup without a successful saved-credential validation.
- `/api/status` requires the setup/admin session and does not expose the raw Immich API key.
- saved settings do not reveal the Immich API key.

Mock HTTP unit tests cover setup/auth/settings/Immich validation behavior. Do not add live Immich integration tests to repo or CI for the MVP.

## Cache Rotation And Outage Checks

Unit tests cover the cache rotation rules that are hard to verify deterministically in a browser:

- top-off stops at `cache.target_items`.
- uncached candidates are preferred when filling the cache.
- stale source entries are evicted before valid fallback photos.
- current and near-upcoming playback entries are protected by `cache.prefetch_items`.
- queue refresh preserves the current photo when cache contents change.

For manual browser checks, use local mock source to confirm `/frame` stays calm and playable while the daemon runs. Immich outage behavior should be tested with mocked/unit-tested adapter failures for MVP development, not live Immich CI tests.

## Useful Local Commands

Validate config:

```sh
go run ./cmd/immich-frame config validate -config config.dev.toml
```

Print version:

```sh
go run ./cmd/immich-frame version
```

Inspect runtime state when using a custom data directory:

```sh
go run ./cmd/immich-frame status -data-dir .immich-frame
```

Reset local generated state and cache:

```sh
go run ./cmd/immich-frame reset -data-dir .immich-frame
```

## Troubleshooting

If `go test ./...` fails on Windows with access errors under the Go build cache,
try setting `GOCACHE` to a writable local directory for the command:

```powershell
$env:GOCACHE = "$PWD\.immich-frame-go-build"
go test ./...
```

If `pnpm` or `node` resolve to an unexpected version in an agent shell, check the active PowerShell environment before changing project scripts.

If the slideshow page is blank:

- Confirm the daemon is still running.
- Confirm `dev/photos` contains the sample SVG files.
- Check `http://127.0.0.1:8787/api/state`.
- Check browser devtools Console and Network.
- Re-run `pnpm build:embedded-ui` if testing embedded assets.

If a browser automation tool cannot render locally, manually verify in a normal
desktop browser window. Headless browser modes may behave differently from the
target Chromium kiosk runtime on some Windows GPU/driver setups.
