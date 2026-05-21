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
- [ ] Add lightweight status page.
- [x] Require admin session after setup.
- [x] Update setup/security/config/developer docs alongside behavior changes.

## Phase 3.5: Setup Portal Hardening

Goal: close setup portal gaps before Raspberry Pi appliance work starts.

Checklist:

- [ ] Require a successful Immich validation for the saved URL/API key before setup completion.
- [ ] Prevent random-library setup from bypassing Immich validation.
- [ ] Add clear setup UI guardrails and messages for missing validation or required fields.
- [ ] Add a lightweight status API/settings surface, or intentionally revise API docs if it remains out of scope.
- [ ] Reconcile overlay configuration docs with implemented backend behavior.
- [ ] Update README, agent handoff docs, and developer docs.
- [ ] Run `go test ./...`, `pnpm typecheck`, and `pnpm build`.

## Phase 4: Pi Appliance

Goal: make the reference Pi Zero 2 W boot directly into the frame.

Checklist:

- [ ] Add idempotent install script.
- [ ] Create `immich-frame` system user/group.
- [ ] Create `/etc/immich-frame` and `/var/lib/immich-frame`.
- [ ] Install daemon binary.
- [ ] Install systemd service for daemon.
- [ ] Install or configure Chromium kiosk service.
- [ ] Configure `frame.local` mDNS.
- [ ] Use `http://127.0.0.1:8787/frame` as kiosk URL.
- [ ] Document tested Raspberry Pi OS Lite version.
- [ ] Verify boot/reboot recovery on Pi Zero 2 W.

## Phase 5: Polish And Hardening

Goal: make MVP reliable enough to use daily.

Checklist:

- [ ] Implement cache eviction policy.
- [ ] Implement outage retry/backoff.
- [ ] Implement subtle degraded status overlay.
- [ ] Implement reset flows.
- [ ] Implement CLI status details.
- [ ] Add unit tests for playback/cache/config/auth.
- [ ] Keep tests unit-only; do not add real-Immich integration tests to repo/CI for MVP.
- [ ] Add frontend typecheck/build CI.
- [ ] Update all docs.
- [ ] Confirm no Immich secrets reach browser or logs.
- [ ] Confirm LAN clients cannot view cached photos unauthenticated.

## MVP Exclusions

- [ ] Temporary Wi-Fi AP setup.
- [ ] Weather provider.
- [ ] Favorites source.
- [ ] On-this-day source.
- [ ] GPIO/remote input.
- [ ] Native renderer.
- [ ] Video playback.
- [ ] Flashable image.
- [ ] Auto-updates.
- [ ] Docker/LAN mode.
- [ ] Runtime third-party overlay plugins.
