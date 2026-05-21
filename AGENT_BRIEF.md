# Immich Frame Agent Brief

This is the entry point for future coding agents. Read this file first, then follow the linked docs.

## Project Intent

Build **Immich Frame**, a local appliance-style digital picture frame for self-built HDMI frames powered by Raspberry Pi Zero 2 W-class hardware.

The frame runs its own local software on the device, connects outbound to a user's Immich server, caches display-appropriate photo renditions locally, and renders a fullscreen Chromium kiosk slideshow with subtle overlays.

This is not primarily a hosted web app, Docker/LAN dashboard, or cloud service.

## Reference Platform

- Raspberry Pi Zero 2 W.
- Raspberry Pi OS Lite.
- HDMI display.
- Wi-Fi already configured for MVP.
- No touch or keyboard required after setup.
- Chromium kiosk renderer.

## Non-Negotiable Decisions

- Runtime target is a local Raspberry Pi appliance.
- Core daemon is Go.
- Frontend uses Preact + Vite, with separate frame and setup bundles.
- Frontend tooling uses the active Node/pnpm from the PowerShell environment.
- Matthew's PowerShell profile initializes fnm, so `node -v` and `pnpm -v` should resolve directly in agent shells.
- Styling uses plain CSS/CSS modules, not Tailwind or a component library.
- Icons are minimal inline SVGs, not an icon library.
- Release Go binary embeds built UI assets.
- Kiosk browser opens `http://127.0.0.1:8787/frame`.
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

## Current Phase

**Phase 4: Pi Appliance** is the current goal on `master`.

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

PM review accepted Phase 3.5 as complete. The next planned phase is **Phase 4: Pi Appliance**.

Phase 4 should make the reference Raspberry Pi Zero 2 W boot directly into Immich Frame. Build the install/runtime layer around the existing daemon and UI:

- idempotent Raspberry Pi install script.
- `immich-frame` service user and filesystem permissions.
- `/etc/immich-frame/config.toml` and `/var/lib/immich-frame` runtime data.
- systemd service for `immich-frame serve`.
- Chromium kiosk startup opening `http://127.0.0.1:8787/frame`.
- `frame.local` mDNS through the chosen Raspberry Pi OS path.
- appliance install/runbook docs.
- clear physical Pi verification notes, especially if hardware verification remains pending.

## Git Commit Guidance

Until the MVP/base is complete, work directly on `master` and commit there. Do not create feature branches for ordinary implementation slices unless the user explicitly asks for one.

Create meaningful commits throughout the work. Prefer commits at coherent feature or fix boundaries, not broad phase markers like `phase 3 done`.

For Phase 4, generally commit and push after each completed checklist feature or checklist item with its subitems. Do not commit after every tiny edit, but do commit/push once a distinct feature is implemented, tested, and documented.

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
- Pi installer script.
- service user and filesystem permissions.
- systemd daemon service.
- Chromium kiosk startup.
- mDNS/frame.local setup.
- appliance install/runbook docs.
- setup UI screens.
- focused bug fixes.
- docs updates that record a changed decision or implemented behavior.

Avoid committing every tiny file edit. Also avoid waiting until an entire phase is complete if several distinct pieces are already working and verified.

## Documentation Guidance

Developer-facing docs must move with the code. When a feature changes how a human runs, configures, tests, or debugs the project, update the relevant docs in the same feature slice.

For Phase 4, pay special attention to:

- `docs/hardware.md` for Raspberry Pi OS, kiosk, display-server, and mDNS behavior.
- `docs/local-development.md` for setup portal verification steps.
- `docs/developer-guide.md` for durable human workflow notes.
- appliance install/runbook docs for human setup without AI help.
- `GOAL.md` for checklist progress and handoff status.

## First Task Checklist

- [ ] Read `GOAL.md`, especially the Phase 4 checklist.
- [ ] Confirm the local branch is `master` and remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
- [ ] Run baseline checks: `go test ./...`, `pnpm typecheck`, and `pnpm build`.
- [ ] Add an idempotent Raspberry Pi install script.
- [ ] Add service user, filesystem permission, and runtime directory handling.
- [ ] Add systemd daemon service for `immich-frame serve`.
- [ ] Add Chromium kiosk startup opening `http://127.0.0.1:8787/frame`.
- [ ] Configure or document `frame.local` mDNS.
- [ ] Add/update appliance install and operation docs.
- [ ] Update `GOAL.md` as items are completed.
- [ ] Commit and push coherent slices as checklist features are completed.

## Do Not Build Yet

- [ ] Temporary Wi-Fi access point setup.
- [ ] Weather provider integration.
- [ ] Favorites mode.
- [ ] On-this-day mode.
- [ ] GPIO buttons or IR remote.
- [ ] Native renderer.
- [ ] Video playback.
- [ ] Flashable OS image.
- [ ] Auto-updater.
- [ ] Docker deployment mode.
- [ ] Runtime third-party overlay plugins.

## Read Next

1. `GOAL.md` for definition of done and stop conditions.
2. `docs/implementation-plan.md` for phase order.
3. `docs/architecture.md` for system design.
4. `docs/security.md` before handling auth, secrets, or media routes.
5. `docs/configuration.md` before writing config/state code.
6. `docs/hardware.md` before writing installer/systemd/kiosk code.
7. `docs/developer-guide.md` for human developer workflow expectations.
8. `docs/development.md` before setting up local dev scripts or tests.
9. `docs/future.md` to preserve future extensibility without building it early.
