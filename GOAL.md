# Immich Frame Goal

This document defines when an agent can stop working and what counts as done.

## Long-Term Product Goal

Immich Frame is complete enough for an MVP when a Raspberry Pi Zero 2 W can act as a reliable self-built HDMI digital picture frame connected to Immich.

The frame should boot into a local Chromium kiosk, show a polished setup screen if unconfigured, accept setup from a phone/laptop on the same Wi-Fi, cache display-appropriate Immich photos, and run a fullscreen slideshow with subtle overlays.

## MVP Definition Of Done

The MVP is done when all items below are complete on the reference Pi Zero 2 W hardware:

- [ ] Raspberry Pi OS Lite is flashed and Wi-Fi is configured through Raspberry Pi Imager.
- [ ] `install.sh` installs Immich Frame idempotently.
- [ ] `immich-frame serve` runs as a systemd service.
- [ ] Chromium kiosk starts automatically on boot.
- [ ] Kiosk opens `http://127.0.0.1:8787/frame`.
- [ ] Unconfigured HDMI screen shows setup instructions, `frame.local:8787`, IP fallback, and first-boot setup code.
- [ ] Setup portal accepts setup code.
- [ ] User can create local admin password.
- [ ] Setup portal accepts Immich URL and dedicated Immich API key.
- [ ] Daemon validates Immich connection.
- [ ] User can choose one album or random library mode.
- [ ] Daemon caches first display-targeted photo renditions locally.
- [ ] Slideshow starts as soon as first few cached images are ready.
- [ ] Hidden controls work: previous, pause/play, next, info toggle.
- [ ] Clock overlay works.
- [ ] Photo info overlay works with minimal metadata.
- [ ] Operational status overlay appears only for degraded/error states.
- [ ] Reboot resumes slideshow without manual intervention.
- [ ] Immich/network outage continues from cache when possible.
- [ ] Empty cache plus unavailable Immich shows a calm retry/error state.
- [ ] Settings portal remains available after setup and requires admin auth.
- [ ] `immich-frame status`, `reset`, `config validate`, and `version` work.
- [ ] Unit tests pass.
- [ ] Frontend typecheck/build passes.
- [ ] Docs match implemented behavior.

## Current Session Goal

Complete **Phase 2: Immich Adapter**.

The local scaffold, mock frame loop, and Phase 1.5 validation are complete. The next agent should implement the real Immich adapter behind `internal/immich` while preserving the local mock development loop.

### Phase 0 Done Checklist

- [x] Repo structure exists.
- [x] Go module initialized.
- [x] CLI skeleton compiles.
- [x] Config/secrets/state file paths and data types exist.
- [x] Frame UI bundle scaffold exists.
- [x] Setup UI bundle scaffold exists.
- [x] Shared UI package exists.
- [x] Release embedding strategy is represented in code/build files.
- [x] Basic CI commands are documented or configured.
- [x] CI scope is Go unit tests plus frontend typecheck/build.

### Phase 1 Partial Done Checklist

- [x] Local folder/mock source can produce photo candidates.
- [x] Cache manifest can track cached local media.
- [x] Playback queue can advance current/next/previous in memory.
- [x] `/api/state` returns frame state.
- [x] `/api/events` streams state updates with SSE.
- [x] `/media/:assetID` serves local cached media.
- [x] `/frame` renders slideshow in desktop browser.
- [x] Clock/photo-info/status overlay placeholders render.

### Phase 1.5 Base Validation And Embedding Cleanup Checklist

- [x] Verify the Go toolchain is available in the development environment, or document the exact blocker.
- [x] Run `go test ./...` successfully once Go is available.
- [x] Add focused Go unit tests for the existing local base:
  - [x] config defaults/validation.
  - [x] local folder candidate discovery.
  - [x] cache manifest ensure/list/mark-shown behavior.
  - [x] playback next/previous/pause/resume behavior.
- [x] Reconcile release UI embedding:
  - [x] Decide whether embedded release UI should use the built Vite bundles or a deliberately separate inline fallback.
  - [x] If using Vite bundles, make embedded `/assets/*` serving work from `embed.FS`.
  - [x] If keeping inline fallback, document that choice and ensure `build:embedded-ui` does not create broken embedded asset references.
- [x] Run `pnpm typecheck`.
- [x] Run `pnpm build`.
- [x] Run the local mock slideshow manually in a desktop browser if the Go toolchain is available.
- [x] Record verification results and any remaining environment limitations in docs or commit notes.
- [x] Commit fixes in meaningful slices, not as one broad phase commit.

### Phase 1.5 Verification Notes - 2026-05-20

- Go is available as `go1.26.3 windows/386`.
- `go test ./...` passed after adding focused tests for config, source, cache, playback, and embedded UI serving.
- Release embedding now uses built Vite bundles. `pnpm build:embedded-ui` copies `ui/frame/dist` and `ui/setup/dist` into `internal/api/static`, and `/assets/*` falls back to `embed.FS` when external dist directories are absent.
- `pnpm typecheck` passed.
- `pnpm build` passed.
- `pnpm build:embedded-ui` passed.
- Local mock daemon was run with missing external dist paths to force embedded UI serving: `go run ./cmd/immich-frame serve -config config.dev.toml -dev-source dev/photos -data-dir .immich-frame-verify -frame-dist missing-frame-dist -setup-dist missing-setup-dist`.
- HTTP smoke checks against the running daemon passed: `/frame` 200, embedded frame JS asset 200, embedded frame CSS asset 200, `/api/state` returned ready local-folder state, `/media/:assetID` returned 200, and `POST /api/playback/pause` returned paused state.
- Desktop browser verification passed on 2026-05-21 using the Codex in-app Browser plugin against `http://127.0.0.1:8787/frame`. The daemon was run with missing external dist paths to force embedded UI serving. Browser evidence showed title `Immich Frame`, visible mock photo `sample-dawn`, clock/photo-info overlays, playback controls, local image source `/media/a0007e962ab0bb36`, embedded asset URLs `/assets/index-DW__ochF.js` and `/assets/index-CiPW52pQ.css`, and no console warnings/errors. HTTP checks also confirmed `/api/state`, the embedded JS/CSS assets, and `/media/a0007e962ab0bb36` returned `200`.

### Phase 1.5 PM Readiness Confirmation - 2026-05-21

- Repo is clean on `master`.
- `go test ./...` passed.
- Matthew's PowerShell profile initializes fnm, and agent shells should use the active Node/pnpm from that PowerShell environment.
- As of the cleanup after Phase 1.5, this resolves to Node `v24.12.0` and pnpm `11.x`.
- Frontend package typechecks passed with plain pnpm:
  - `pnpm --filter @immich-frame/shared typecheck`
  - `pnpm --filter @immich-frame/frame typecheck`
  - `pnpm --filter @immich-frame/setup typecheck`
- Frontend package builds passed with plain pnpm:
  - `pnpm --filter @immich-frame/shared build`
  - `pnpm --filter @immich-frame/frame build`
  - `pnpm --filter @immich-frame/setup build`
- The repo intentionally does not pin pnpm with `packageManager`; use the machine's active PowerShell `pnpm`.

### Phase 2 Immich Adapter Checklist

- [ ] Baseline verification before changes:
  - [ ] Confirm branch is `master`.
  - [ ] Confirm remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
  - [ ] Run `go test ./...`.
  - [ ] Run `pnpm typecheck` and `pnpm build`.
- [ ] Verify current Immich API behavior:
  - [ ] Review official Immich docs/OpenAPI for required endpoints.
  - [ ] Manually verify required behavior against Matthew's Immich instance when credentials/context are available.
  - [ ] Record endpoint assumptions and version notes in docs or adapter comments.
- [ ] Define the Immich adapter boundary in `internal/immich`:
  - [ ] Connection/auth test method.
  - [ ] Album listing method.
  - [ ] Asset candidate listing for a selected album.
  - [ ] Asset candidate listing for random library mode.
  - [ ] Display-targeted rendition fetch method.
  - [ ] Error normalization that hides low-level HTTP details from callers.
- [ ] Implement Immich connection testing:
  - [ ] Use the dedicated API key from secrets/config inputs.
  - [ ] Return user-safe errors for invalid URL, invalid key, network failure, and incompatible response.
  - [ ] Add mock HTTP unit tests.
- [ ] Implement Immich album listing:
  - [ ] Normalize album id, name, and item count when available.
  - [ ] Keep setup/search UI needs in mind without implementing the full setup portal yet.
  - [ ] Add mock HTTP unit tests.
- [ ] Implement Immich asset candidate listing:
  - [ ] Album source candidates.
  - [ ] Random library candidates.
  - [ ] Conservative default filters where the API supports them: photos only, exclude archived/hidden/trashed/videos.
  - [ ] Avoid leaking raw Immich asset JSON outside `internal/immich`.
  - [ ] Add mock HTTP unit tests.
- [ ] Implement display-targeted rendition fetching:
  - [ ] Prefer the best Immich-provided non-original rendition appropriate for the requested target.
  - [ ] Preserve returned content type.
  - [ ] Do not download originals by default.
  - [ ] Add mock HTTP unit tests.
- [ ] Normalize display metadata:
  - [ ] Minimal asset id, media/rendition identity, taken date, source name, dimensions/orientation if available.
  - [ ] Do not expose raw EXIF, GPS coordinates, file paths, people names, direct Immich URLs, or full Immich asset blobs to browser-facing code.
  - [ ] Add unit tests for metadata normalization.
- [ ] Integrate adapter with existing source/cache seams lightly:
  - [ ] Preserve local folder source for development.
  - [ ] Do not build full setup portal behavior yet unless needed for adapter verification.
  - [ ] Keep playback/cache APIs source-agnostic.
- [ ] Update docs:
  - [ ] Record supported/tested Immich version assumptions.
  - [ ] Record any manual verification steps.
  - [ ] Update `GOAL.md` as checklist items are completed.
- [ ] Commit and push after each coherent checklist feature or checklist item with subitems is completed.

## Stop Conditions

An agent can stop when:

- [ ] The requested phase checklist is complete.
- [ ] Tests/builds for touched areas pass, or failures are clearly documented.
- [ ] Docs are updated for any changed behavior or decision.
- [ ] Work has been committed in meaningful increments.
- [ ] No required dev servers, long-running commands, or install sessions are left running.
- [ ] Remaining work is captured in a clear checklist.

## Commit Expectations

Keep git history useful and reviewable.

Until the MVP/base is complete, agents should work directly on `master` and commit there. Do not create feature branches for ordinary implementation slices unless the user explicitly asks for one.

Commit and push at coherent feature or fix boundaries rather than only at phase boundaries. A good history should explain what was built in practical slices, such as:

- project scaffold and build tooling.
- CLI skeleton.
- config/secrets/state models.
- frontend frame scaffold.
- setup/settings scaffold.
- embedded UI assets.
- playback queue.
- cache manifest.
- SSE/API routes.
- mock source support.
- targeted bug fixes.
- documentation updates for changed behavior.

Do not commit after every tiny edit. Do not wait until all of Phase 0 or Phase 1 is complete if several distinct working pieces can be committed separately.

For Phase 2 specifically, generally commit and push after each completed checklist feature or checklist item with its subitems. Examples: Immich connection test, album listing, asset candidate listing, rendition fetching, metadata normalization, adapter/source integration, and docs updates.

Before each commit:

- [ ] Review `git status`.
- [ ] Review the diff for the intended scope.
- [ ] Run the relevant tests/build checks when practical.
- [ ] Use a specific commit message describing the completed slice.
- [ ] Push `master` after the commit unless the user has asked you to hold local commits.

## Out Of Scope For MVP

- [ ] Temporary setup Wi-Fi access point.
- [ ] Captive portal.
- [ ] Weather provider integration.
- [ ] Favorites source mode.
- [ ] On-this-day source mode.
- [ ] People/location/date smart rules.
- [ ] Video playback.
- [ ] GPIO buttons, IR remotes, or Bluetooth remote integration.
- [ ] Native framebuffer/SDL renderer.
- [ ] Flashable custom OS image.
- [ ] Automatic updates.
- [ ] Docker/LAN deployment mode.
- [ ] Runtime third-party overlay plugins.
- [ ] Full log dashboard or cache management dashboard.
