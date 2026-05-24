# Immich Frame

A local digital picture frame app for Immich libraries, designed to become the software core for self-built HDMI frames.

## Status

Phase 0, the local Phase 1 mock slideshow loop, Phase 1.5 validation, Phase 2 Immich adapter, Phase 3 setup portal, and Phase 3.5 setup hardening are complete.

The current work is **Phase 5: Browser MVP Polish And Hardening**.

Phase 5 cache rotation and outage retry/backoff are implemented in the daemon. Remaining Phase 5 work is focused on degraded/offline frame UI polish, CLI status/reset hardening, browser MVP verification, and final docs alignment.

The Raspberry Pi Chromium kiosk direction is paused. The browser MVP should be finished first so future hardware work can focus on replacing the renderer instead of rebuilding setup, cache, source, and Immich integration behavior.

Start here:

- `AGENT_BRIEF.md` for future coding agents.
- `GOAL.md` for definition of done.
- `docs/implementation-plan.md` for build phases.
- `docs/architecture.md` for system design.
- `docs/development.md` for local development expectations.
- `docs/developer-guide.md` for the human developer workflow and project map.
- `docs/local-development.md` for desktop run, test, and manual verification steps.

## Product Direction

Immich Frame runs locally, connects outbound to a user's Immich server, caches display-appropriate photo renditions locally, and currently renders a fullscreen browser slideshow with subtle overlays.

It is not primarily a hosted web app, Docker dashboard, or cloud service.

## Current Target

- Local browser MVP.
- Go daemon.
- Preact/Vite frame and setup UIs.
- Embedded browser assets for release-style serving.
- Future lightweight renderer/hardware path after browser MVP behavior is solid.

## MVP Scope

- Local Go daemon.
- Preact frame UI and setup UI.
- pnpm frontend tooling.
- Same-Wi-Fi setup portal.
- Dedicated Immich API key per frame.
- Setup completion requires a successful Immich connection test for the saved URL/API key.
- Album and random-library modes.
- Display-targeted local image cache.
- Cache rotation that avoids looping one static cached seed forever.
- Clock, photo info, and operational status overlays.
- Hidden controls.
- Reboot recovery.
- Cache-first outage behavior.

## Out Of Scope For MVP

- Temporary setup Wi-Fi network.
- Weather provider.
- Favorites and on-this-day modes.
- GPIO buttons.
- Native renderer.
- Video playback.
- Flashable image.
- Auto-updates.
- Docker/LAN deployment mode.

## Local Development

For the complete desktop runbook, see `docs/local-development.md`.

Prerequisites:

- Go 1.22 or newer.
- Node.js from the active PowerShell environment.
- pnpm from the active PowerShell environment.

Install frontend dependencies:

```sh
pnpm install
```

Run checks:

```sh
go test ./...
pnpm typecheck
pnpm build
```

Prepare embedded UI assets before a release Go build:

```sh
pnpm build:embedded-ui
```

Run the desktop mock slideshow:

```sh
go run ./cmd/immich-frame serve -config config.dev.toml -dev-source dev/photos
```

Then open:

- `http://127.0.0.1:8787/frame`
- `http://127.0.0.1:8787/setup`

## License Plan

The intended public license is MIT once the reference Pi Zero 2 W install works.
