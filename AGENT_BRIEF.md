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

Continue with **Phase 2: Immich Adapter**.

Phase 0, the first local Phase 1 slice, and Phase 1.5 validation are complete on `master`.

Current validated base:

- Go daemon/CLI scaffold exists.
- Local folder mock source works.
- Cache manifest and playback queue exist.
- `/api/state`, `/api/events`, `/media/:assetID`, and playback commands exist.
- Frame/setup Preact bundles exist.
- Built Vite UI bundles are embedded and served by the Go daemon.
- Focused unit tests cover config, source, cache, playback, and embedded UI serving.
- Local mock slideshow has been verified in a desktop browser.

Next goal: implement the real Immich API adapter while preserving the local mock frame loop.

## Git Commit Guidance

Until the MVP/base is complete, work directly on `master` and commit there. Do not create feature branches for ordinary implementation slices unless the user explicitly asks for one.

Create meaningful commits throughout the work. Prefer commits at coherent feature or fix boundaries, not broad phase markers like `phase 2 done`.

For Phase 2, generally commit and push after each completed checklist feature or checklist item with its subitems. Do not commit after every tiny edit, but do commit/push once a distinct feature is implemented, tested, and documented.

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
- focused bug fixes.
- docs updates that record a changed decision or implemented behavior.

Avoid committing every tiny file edit. Also avoid waiting until an entire phase is complete if several distinct pieces are already working and verified.

## First Task Checklist

- [ ] Read `GOAL.md`, especially the Phase 2 checklist.
- [ ] Confirm the local branch is `master` and remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
- [ ] Run baseline checks: `go test ./...`, `pnpm typecheck`, and `pnpm build`.
- [ ] Verify current Immich API behavior through official docs/OpenAPI and, where needed, manual testing against Matthew's Immich instance.
- [ ] Implement the adapter behind `internal/immich` without leaking Immich details into playback/cache/setup code.
- [ ] Use mock HTTP unit tests for adapter behavior; do not add real-Immich integration tests to repo/CI for MVP.
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
7. `docs/development.md` before setting up local dev scripts or tests.
8. `docs/future.md` to preserve future extensibility without building it early.
