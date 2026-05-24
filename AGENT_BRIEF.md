# Immich Frame Agent Brief

This is the entry point for future coding agents. Read this file first, then follow the linked docs.

## Project Intent

Build **Immich Frame**, a local digital picture frame app that can eventually power self-built HDMI frames.

The frame runs local software, connects outbound to a user's Immich server, caches display-appropriate photo renditions locally, and currently renders a fullscreen browser slideshow with subtle overlays.

This is not primarily a hosted web app, Docker/LAN dashboard, or cloud service.

## Current Reference Platform

- Local desktop/browser development.
- Go daemon.
- Preact/Vite browser renderer.
- Same-machine or same-LAN setup portal.
- Future frame hardware remains important, but hardware/appliance setup is paused until the browser MVP behavior is solid.

## Non-Negotiable Decisions

- Current runtime target is the local browser MVP.
- Raspberry Pi appliance/install work is paused.
- Core daemon is Go.
- Frontend uses Preact + Vite, with separate frame and setup bundles.
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
- Future hardware work should evaluate a lighter renderer instead of assuming Chromium kiosk on Pi Zero 2 W.

## Current Phase

**Phase 5: Browser MVP Polish And Hardening** is the current goal on `master`.

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

The next planned phase is **Phase 5: Browser MVP Polish And Hardening**. Finish the browser-based product behavior first so a future renderer swap only replaces the rendering layer rather than the daemon/setup/cache/source system.

Phase 5 priorities:

- cache rotation and eviction so the same static cache seed is not shown forever.
- Immich outage retry/backoff with cache-first playback.
- calm degraded/offline UI states.
- CLI status/reset/config validation hardening.
- browser MVP verification and docs.
- preserve renderer boundaries so a lighter renderer can replace the browser later.

## Git Commit Guidance

Until the MVP/base is complete, work directly on `master` and commit there. Do not create feature branches for ordinary implementation slices unless the user explicitly asks for one.

Create meaningful commits throughout the work. Prefer commits at coherent feature or fix boundaries, not broad phase markers like `phase 3 done`.

For Phase 5, generally commit and push after each completed checklist feature or checklist item with its subitems. Do not commit after every tiny edit, but do commit/push once a distinct feature is implemented, tested, and documented.

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

For Phase 5, pay special attention to:

- `docs/architecture.md` for cache/playback/source behavior.
- `docs/configuration.md` for cache, sync, status, and reset settings.
- `docs/security.md` for secret-safe status/reset/media behavior.
- `docs/local-development.md` for browser MVP verification steps.
- `docs/developer-guide.md` for durable human workflow notes.
- `docs/future.md` for renderer/hardware follow-up notes.
- `GOAL.md` for checklist progress and handoff status.

## First Task Checklist

- [ ] Read `GOAL.md`, especially the Phase 5 checklist.
- [ ] Confirm the local branch is `master` and remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
- [ ] Run baseline checks: `go test ./...`, `pnpm typecheck`, and `pnpm build`.
- [ ] Implement cache rotation and eviction.
- [ ] Implement outage retry/backoff.
- [ ] Tighten degraded/offline frame UI states.
- [ ] Harden CLI status/reset/config validation behavior.
- [ ] Re-verify browser MVP behavior.
- [ ] Update developer-facing docs as behavior changes.
- [ ] Update `GOAL.md` as items are completed.
- [ ] Commit and push coherent slices as checklist features are completed.

## Do Not Build Yet

- [ ] Temporary Wi-Fi access point setup.
- [ ] Weather provider integration.
- [ ] Favorites mode.
- [ ] On-this-day mode.
- [ ] GPIO buttons or IR remote.
- [ ] Native renderer or other lightweight renderer.
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
