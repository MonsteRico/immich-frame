# Implementation Plan

## Phase 0: Scaffold

Goal: create the project shape and enough build plumbing that future phases have a stable foundation.

Checklist:

- [ ] Initialize repo under `D:/Coding/immich-frame`.
- [ ] Add MIT license.
- [ ] Initialize single Go module at repo root.
- [ ] Add `cmd/immich-frame/main.go`.
- [ ] Add CLI commands: `serve`, `status`, `reset`, `config validate`, `version`.
- [ ] Add internal package directories.
- [ ] Add config/secrets/state models and default paths.
- [ ] Add Preact/Vite `ui/frame` bundle.
- [ ] Add Preact/Vite `ui/setup` bundle.
- [ ] Add `ui/shared` package.
- [ ] Use pnpm for frontend tooling.
- [ ] Use plain CSS/CSS modules and minimal inline SVG icons.
- [ ] Add embedded UI build path for release.
- [ ] Add development config example.
- [ ] Add initial docs links from README.

## Phase 1: Local Frame Loop

Goal: run a mock/local-folder slideshow locally in a desktop browser without Immich.

Checklist:

- [ ] Add local folder source provider for dev only.
- [ ] Add asset candidate model.
- [ ] Add cache manifest model.
- [ ] Add local media serving from cache/dev source.
- [ ] Add playback queue with current, next, previous, pause/resume.
- [ ] Add SSE event stream.
- [ ] Add `/api/state`.
- [ ] Add `/api/playback/*` commands.
- [ ] Add `/frame` UI slideshow.
- [ ] Add cut/crossfade transitions.
- [ ] Add hidden controls and keyboard shortcuts.
- [ ] Add clock/photo-info/status overlay placeholders.
- [ ] Add browser display-size reporting endpoint.

## Phase 1.5: Base Validation And Embedding Cleanup

Goal: tighten the local foundation before starting the real Immich adapter.

Checklist:

- [ ] Verify Go toolchain availability and run `go test ./...`.
- [ ] Add unit tests for config defaults/validation.
- [ ] Add unit tests for local folder candidate discovery.
- [ ] Add unit tests for cache manifest ensure/list/mark-shown behavior.
- [ ] Add unit tests for playback next/previous/pause/resume behavior.
- [ ] Reconcile embedded release UI behavior.
- [ ] If release embedding uses Vite output, serve embedded `/assets/*` from `embed.FS`.
- [ ] If release embedding uses a deliberate inline fallback, document that and keep `build:embedded-ui` from producing broken asset references.
- [ ] Run `pnpm typecheck`.
- [ ] Run `pnpm build`.
- [ ] Manually verify the local mock slideshow in a desktop browser when possible.

## Phase 2: Immich Adapter

Goal: connect to real Immich through an isolated adapter.

Checklist:

- [ ] Verify current Immich API through official docs/OpenAPI.
- [ ] Manually test required behavior against Matthew's Immich instance.
- [ ] Implement connection test.
- [ ] Implement album listing.
- [ ] Implement album asset candidate listing.
- [ ] Implement random-library candidate listing.
- [ ] Implement display-targeted rendition fetch.
- [ ] Normalize Immich metadata into minimal display metadata.
- [ ] Add mock HTTP unit tests for adapter request/response behavior.
- [ ] Keep all endpoint assumptions isolated in `internal/immich`.

No built-in real-Immich integration tests for MVP.

## Phase 3: Setup Portal

Goal: configure the frame from a phone/laptop on same Wi-Fi.

Checklist:

- [x] Add setup state model and backend boundary.
- [x] Generate fixed first-boot setup code until setup completes.
- [x] Render HDMI setup state with URL, IP fallback, code, and status.
- [x] Add `/setup` portal.
- [x] Claim setup code.
- [x] Create local admin password.
- [x] Store password hash in `secrets.json`.
- [x] Add admin login/logout/session behavior.
- [x] Add settings read/write API backed by `config.toml`.
- [x] Accept raw Immich URL and raw Immich API key fields.
- [x] Validate Immich connection.
- [x] Save Immich URL to config and API key to secrets.
- [x] Select source mode: one album or random library.
- [x] Provide a searchable album picker for album mode, showing album name and item count when available.
- [x] Configure interval, display fit, cache preset, and overlays.
- [x] Add lightweight status page.
- [x] Require admin session after setup.
- [x] Update setup/security/config/developer docs alongside behavior changes.

## Phase 3.5: Setup Portal Hardening

Goal: close setup portal gaps before Raspberry Pi appliance work starts.

Checklist:

- [x] Require a successful Immich validation for the saved URL/API key before setup completion.
- [x] Prevent random-library setup from bypassing Immich validation.
- [x] Add clear setup UI guardrails and messages for missing validation or required fields.
- [x] Add a lightweight status API/settings surface.
- [x] Reconcile overlay configuration docs with implemented backend behavior.
- [x] Update README, agent handoff docs, and developer docs.
- [x] Run `go test ./...`, `pnpm typecheck`, and `pnpm build`.

## Phase 4: Pi Appliance Experiment - Reverted

Goal was to make the reference Pi Zero 2 W boot directly into the frame.

Status: reverted/paused. The Chromium kiosk path looks too heavy for the Pi Zero 2 W, and hardware setup work is premature until Phase 6 chooses a lighter renderer direction.

Checklist:

- [x] Remove installer/systemd/kiosk assets from the active codebase.
- [x] Pause Raspberry Pi Chromium kiosk work.
- [ ] Revisit hardware after Phase 6 chooses a lighter rendering strategy.

## Phase 5: Browser MVP Polish And Hardening

Goal: make the browser-based MVP reliable enough to use daily before swapping or replacing the renderer.

Checklist:

- [x] Implement cache rotation and eviction policy:
  - [x] Refresh Immich candidate pools periodically.
  - [x] Top off cached display-targeted renditions toward `target_items`.
  - [x] Maintain a near-term prefetch buffer from `prefetch_items`.
  - [x] Prefer never-shown and least-recently-shown candidates.
  - [x] Evict entries that left the source before evicting valid offline fallback photos.
  - [x] Avoid evicting current and near-upcoming playback entries.
  - [x] Refresh playback queue when cache contents change so the frame does not loop one static seed forever.
- [x] Implement outage retry/backoff.
- [x] Implement subtle degraded status overlay.
- [x] Implement reset flows.
- [x] Implement CLI status details.
- [x] Add unit tests for playback/cache/config/auth.
- [x] Keep tests unit-only; do not add real-Immich integration tests to repo/CI for MVP.
- [x] Add frontend typecheck/build verification.
- [x] Update all docs.
- [x] Confirm no Immich secrets reach browser or logs.
- [x] Confirm LAN clients cannot view cached photos unauthenticated.

## Phase 5.5: Browser MVP Acceptance Fix Pass

Goal: close PM verification gaps found after Phase 5 was reported complete, without restarting hardware/appliance work.

Checklist:

- [x] Rotate/churn stable album caches even when already at `cache.target_items`.
- [x] Preserve current and near-upcoming playback entries during album churn using `cache.prefetch_items`.
- [x] Add an `extra-small` local testing cache preset with roughly 10 cached photos.
- [x] Publish recovered `ready` status after outage retry succeeds, even if cache contents did not change.
- [x] Align docs to show Phase 5.5 as the current fix pass until final verification is complete.
- [x] Run final `go test ./...`, `pnpm typecheck`, `pnpm build`, and `pnpm build:embedded-ui` because setup UI assets changed.

## Post-Phase 5.5: Rolling Cache Strategy Adjustment

Goal: apply the approved product change that the frame should steadily progress through the wider album/library instead of gently adding one new cached image per timer refresh.

Checklist:

- [x] Add playback-driven rolling cache refresh after a configurable number of shown images.
- [x] Add configurable rolling refresh batch size.
- [x] Swap shown, non-protected cached entries for new album/random-library candidates.
- [x] Preserve current and near-upcoming playback entries.
- [x] Request immediate cache maintenance after first setup completes with an Immich source.
- [x] Document that this strategy change was requested and approved after Phase 5.5.

## Phase 6: Renderer And Hardware Re-evaluation

Goal: choose and prototype the appliance renderer after the browser/reference MVP proved the daemon, setup, Immich adapter, cache, and playback system.

Context:

- Phase 5.5 is complete.
- Matthew verified the app with a personal Immich instance/API key.
- Random mode rotates, `extra-small` cache rotation is visible, restart/reboot simulation resumes without setup, and embedded release-style browser serving works.
- During a network outage, the daemon appears to keep cache-first playback and reconnect correctly, but the Chromium tab does not keep the visual slideshow moving or resume cleanly after reconnect.
- Because Chromium kiosk is likely too heavy or fragile for Pi Zero 2 W-class hardware, do not make the next phase a browser outage-hardening phase.

Checklist:

- [ ] Define renderer boundary between daemon state/media APIs and presentation layer.
- [ ] Decide whether the appliance renderer consumes `/api/state` and `/media/:assetID`, a local-only renderer API, direct cache file paths, or another narrow contract.
- [ ] Prefer a resilient polling or hybrid update loop over SSE-only rendering.
- [ ] Evaluate at least three lightweight renderer options for Pi Zero 2 W-class hardware.
- [ ] Score renderer options on footprint, image quality, overlay feasibility, packaging complexity, Go integration, and testability without hardware.
- [ ] Recommend one primary renderer path and one fallback path.
- [ ] Build a narrow proof of concept for the recommended path.
- [ ] Show at least one cached/local image and one simple overlay in the proof of concept.
- [ ] Keep the current image visible if state/media refresh fails.
- [ ] Reuse existing daemon, setup portal, Immich adapter, cache, playback, and settings behavior.
- [ ] Avoid restarting installer/systemd work until a renderer direction is chosen.
- [ ] Re-plan hardware install steps around the chosen renderer.

## MVP Exclusions

- [ ] Temporary Wi-Fi AP setup.
- [ ] Weather provider.
- [ ] Favorites source.
- [ ] On-this-day source.
- [ ] GPIO/remote input.
- [ ] Video playback.
- [ ] Flashable image.
- [ ] Auto-updates.
- [ ] Docker/LAN mode.
- [ ] Runtime third-party overlay plugins.
