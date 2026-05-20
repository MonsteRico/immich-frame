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
- Frontend tooling uses pnpm.
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

Start with **Phase 0: Scaffold**, then enough of **Phase 1: Local Frame Loop** to run a mock slideshow locally in a desktop browser.

## Git Commit Guidance

Create meaningful commits throughout the work. Prefer commits at coherent feature or fix boundaries, not just broad phase markers like `phase 0 done`.

Good commit boundaries include:

- repository/build scaffold.
- CLI skeleton.
- config/secrets/state models.
- frame UI scaffold.
- setup UI scaffold.
- embedded UI build path.
- mock source/playback/cache loop.
- SSE/API route group.
- focused bug fixes.
- docs updates that record a changed decision or implemented behavior.

Avoid committing every tiny file edit. Also avoid waiting until an entire phase is complete if several distinct pieces are already working and verified.

## First Task Checklist

- [ ] Create repo scaffold under `D:/Coding/immich-frame` without moving these planning docs.
- [ ] Initialize a single Go module at repo root.
- [ ] Add `cmd/immich-frame/main.go` with CLI skeleton.
- [ ] Add internal package directories from `docs/architecture.md`.
- [ ] Add Preact/Vite frame bundle under `ui/frame`.
- [ ] Add Preact/Vite setup bundle under `ui/setup`.
- [ ] Add shared frontend package under `ui/shared`.
- [ ] Use pnpm for frontend tooling.
- [ ] Add config/secrets/state model skeletons.
- [ ] Add static UI embedding path for release builds.
- [ ] Add local dev mode plan or placeholders for mock photo source.
- [ ] Update `GOAL.md` with completed checklist items.

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
