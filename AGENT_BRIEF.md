# Immich Frame Agent Brief

This is the entry point for future coding agents. Read this file first, then follow the linked docs.

## Project Intent

Build **Immich Frame**, a local digital picture frame app that can eventually power self-built HDMI frames.

The frame runs local software, connects outbound to a user's Immich server, caches display-appropriate photo renditions locally, and currently has a fullscreen browser reference renderer with subtle overlays.

This is not primarily a hosted web app, Docker/LAN dashboard, or cloud service.

## Current Reference Platform

- Local desktop/browser development for daemon/setup/reference UI work.
- Go daemon.
- Preact/Vite browser reference renderer.
- Same-machine or same-LAN setup portal.
- Future frame hardware remains important, but hardware/appliance setup is paused until a lighter renderer direction is chosen.

## Non-Negotiable Decisions

- Current product target is a local frame MVP that can run on Pi Zero 2 W-class hardware.
- The browser frame UI is now a development/reference renderer, not the assumed final appliance runtime.
- Raspberry Pi appliance/install work is paused.
- Core daemon is Go.
- Frontend uses Preact + Vite, with separate frame and setup bundles for the reference frame UI and setup portal.
- Frontend tooling uses the active Node/pnpm from the PowerShell environment.
- Matthew's PowerShell profile initializes fnm, so `node -v` and `pnpm -v` should resolve directly in agent shells.
- Styling uses plain CSS/CSS modules, not Tailwind or a component library.
- Icons are minimal inline SVGs, not an icon library.
- Release Go binary embeds built UI assets.
- Browser renderer opens `http://127.0.0.1:8787/frame`.
- Setup portal is available at `http://frame.local:8787/setup` and IP fallback.
- MVP networking assumes same Wi-Fi already configured.
- Future public setup may add temporary Wi-Fi AP/captive portal.
- Browser never receives Immich API keys or direct authenticated Immich image URLs.
- Daemon owns Immich API access, cache, playback queue, settings, and SSE state.
- Browser renders local cached media URLs and sanitized metadata only.
- Cache stores display-targeted renditions, not originals by default.
- Videos are out of MVP.
- Weather is near-future, not MVP.
- Docker/LAN/server-hosted modes are out of scope.
- Phase 6 should evaluate a lighter renderer instead of assuming Chromium kiosk on Pi Zero 2 W.

## Current Phase

**Phase 6: Renderer Replacement Spike** is the current goal on `master`.

Phase 0, the first local Phase 1 slice, Phase 1.5 validation, Phase 2 Immich adapter work, Phase 3 setup portal, and Phase 3.5 setup hardening are complete on `master`.

Current validated base:

- Go daemon/CLI scaffold exists.
- Local folder mock source works.
- Cache manifest and playback queue exist.
- `/api/state`, `/api/events`, `/media/:assetID`, and playback commands exist.
- Frame/setup Preact bundles exist.
- Built Vite UI bundles are embedded and served by the Go daemon.
- Focused unit tests cover config, source, cache, playback, and embedded UI serving.
- Local mock slideshow has been verified in a desktop browser.
- Immich adapter exists behind `internal/immich` with mock HTTP tests.
- Immich connection testing, album listing, album/random candidate listing, preview rendition fetching, and metadata normalization are implemented.

Current base now includes the phone-first setup/settings portal flow and supporting backend routes, using the existing Immich adapter and preserving the local mock frame loop.

Phase 3.5 closed the PM review gaps found on 2026-05-21:

- setup completion now requires a successful Immich validation for the saved URL/key, including random-library mode.
- changing the saved Immich URL or API key clears stale validation unless the new pair already matches a successful validation.
- `GET /api/status` and `GET /api/settings` surface setup/configuration status, Immich validation status, source mode, cache count, and last error without raw secrets.
- setup/settings UI guards unavailable actions and explains required validation or missing fields.
- overlay configuration docs match the generic envelope fields the backend currently reads/writes.

Phase 4 Raspberry Pi appliance work was reverted after hardware exploration showed likely Chromium kiosk performance risk on the Pi Zero 2 W. Do not restart installer/systemd/kiosk work in the current phase.

Phase 5 browser MVP polish was reported complete, but PM verification found acceptance gaps around stable album cache churn, a very small local testing cache preset, recovered ready-state publication after outages, and stale phase-status docs. Phase 5.5 closed those gaps on 2026-05-24.

Phase 5.5 completed:

- stable album cache rotation when the cache is already full, preserving current and near-upcoming playback entries.
- an `extra-small` cache preset for local testing with roughly 10 cached photos.
- recovered `ready` status publication after outage retry succeeds even if cache contents did not change.
- documentation/status alignment so Phase 6 renderer/hardware re-evaluation remains future work only.
- final verification with `go test ./...`, `pnpm typecheck`, `pnpm build`, and `pnpm build:embedded-ui`.

Post-Phase 5.5 product feedback requested and approved a stronger cache rotation strategy: the frame should progress through the broader album/library over time rather than looping the same warm cache. The app now uses playback-driven rolling cache refresh for album/random-library sources, with `cache.refresh_after_shown_items` triggering an immediate maintenance request and `cache.refresh_batch_items` controlling how many shown, unprotected entries can be swapped for new candidates. Completing setup with an Immich source also requests immediate cache maintenance so an empty first-run frame can fetch images without a daemon restart.

`immich-frame serve --logs` enables opt-in cache/playback diagnostics. Keep these logs operational and count-focused: refresh triggers, source mode, candidate counts, cache before/after counts, protected/fetched/rotated/evicted/pruned counts, threshold events, and cache hits are useful. Do not log Immich API keys, direct authenticated URLs, filenames, titles, raw responses, or other secret-heavy details.

Matthew verified Phase 5.5 against a personal Immich instance/API key. Random mode rotates, `extra-small` cache rotation is visible, restart/reboot simulation resumes without setup, and embedded release-style serving works.

Known reference-renderer limitation: during a network outage, the daemon appears to continue cache-first playback and reconnects correctly, but the Chromium tab does not keep the visual slideshow moving and does not resume cleanly after reconnect. Do not make the next phase a deep Chromium/SSE/browser hardening pass. Treat this as evidence that the appliance renderer should be replaced.

Phase 6 should choose and prototype a lighter renderer while preserving the existing daemon, setup portal, Immich adapter, cache, playback, settings, and docs foundation. Do not restart installer/systemd/kiosk work until the renderer direction is selected.

Phase 6 current decision state:

- The renderer contract is now documented as a daemon-owned snapshot/presentation boundary.
- The browser `/frame` path remains the reference/development renderer.
- The appliance renderer should use resilient polling or hybrid polling plus event wake-ups, not SSE-only state.
- Recommended primary path before prototype discussion: Go + SDL2 native renderer.
- Recommended fallback path before prototype discussion: pre-composited framebuffer/image-viewer renderer.
- Proof-of-concept work has not started yet; discuss the renderer paths before continuing.

## Git Commit Guidance

Until the MVP/base is complete, work directly on `master` and commit there. Do not create feature branches for ordinary implementation slices unless the user explicitly asks for one.

Create meaningful commits throughout the work. Prefer commits at coherent feature or fix boundaries, not broad phase markers like `phase 3 done`.

For Phase 6, generally commit and push after each coherent decision, prototype, or documentation slice. Do not commit after every tiny edit, but do commit/push once a distinct piece is implemented, tested, and documented.

Good commit boundaries include:

- repository/build scaffold.
- CLI skeleton.
- config/secrets/state models.
- frame UI scaffold.
- setup UI scaffold.
- embedded UI build path.
- mock source/playback/cache loop.
- SSE/API route group.
- Immich connection test.
- Immich album listing.
- Immich album/random candidate listing.
- Immich rendition fetching.
- Immich metadata normalization.
- setup code and setup state.
- admin password/session auth.
- settings read/write.
- Immich setup validation and album picker API.
- validation-required setup completion.
- lightweight status API/settings surface.
- setup UI guardrails.
- cache rotation.
- cache eviction policy.
- outage retry/backoff.
- degraded/offline frame states.
- CLI status/reset hardening.
- setup UI screens.
- focused bug fixes.
- docs updates that record a changed decision or implemented behavior.

Avoid committing every tiny file edit. Also avoid waiting until an entire phase is complete if several distinct pieces are already working and verified.

## Documentation Guidance

Developer-facing docs must move with the code. When a feature changes how a human runs, configures, tests, or debugs the project, update the relevant docs in the same feature slice.

For Phase 6, pay special attention to:

- `docs/architecture.md` for renderer boundaries and daemon-owned behavior.
- `docs/implementation-plan.md` for renderer option evaluation and phase status.
- `docs/security.md` for secret-safe status/reset/media behavior.
- `docs/local-development.md` for local renderer spike verification steps.
- `docs/developer-guide.md` for durable human workflow notes.
- `docs/future.md` for renderer/hardware follow-up notes.
- `GOAL.md` for checklist progress and handoff status.

## First Task Checklist

- [ ] Read `GOAL.md`, especially the Phase 6 checklist.
- [x] Confirm the local branch is `master` and remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
- [x] Review the current `immich-config.dev.toml` diff before touching config files; Matthew may have local testing settings there.
- [x] Run baseline checks: `go test ./...`, `pnpm typecheck`, and `pnpm build`.
- [x] Define the renderer contract between daemon state/media and presentation.
- [x] Evaluate lightweight renderer options for Pi Zero 2 W-class hardware.
- [x] Recommend one primary renderer path and one fallback path.
- [ ] Build a narrow proof of concept for the recommended path.
- [ ] Prove or document how the new renderer keeps cached photos visible through state/media refresh failures.
- [ ] Add unit/fixture tests for renderer contract logic where practical.
- [ ] Update developer-facing docs as behavior changes.
- [ ] Update `GOAL.md` as items are completed.
- [ ] Commit and push coherent slices as checklist features are completed.

## Do Not Build Yet

- [ ] Temporary Wi-Fi access point setup.
- [ ] Weather provider integration.
- [ ] Favorites mode.
- [ ] On-this-day mode.
- [ ] GPIO buttons or IR remote.
- [ ] Video playback.
- [ ] Flashable OS image.
- [ ] Auto-updater.
- [ ] Docker deployment mode.
- [ ] Runtime third-party overlay plugins.
- [ ] Raspberry Pi install/systemd/kiosk setup.

## Read Next

1. `GOAL.md` for definition of done and stop conditions.
2. `docs/implementation-plan.md` for phase order.
3. `docs/architecture.md` for system design.
4. `docs/security.md` before handling auth, secrets, or media routes.
5. `docs/configuration.md` before writing config/state code.
6. `docs/future.md` for renderer and hardware follow-up notes.
7. `docs/developer-guide.md` for human developer workflow expectations.
8. `docs/development.md` before setting up local dev scripts or tests.
9. `docs/future.md` to preserve future extensibility without building it early.
