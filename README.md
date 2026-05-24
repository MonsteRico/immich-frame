# Immich Frame

A local digital picture frame app for Immich libraries, designed to become the software core for self-built HDMI frames.

## Status

Phase 0, the local Phase 1 mock slideshow loop, Phase 1.5 validation, Phase 2 Immich adapter, Phase 3 setup portal, and Phase 3.5 setup hardening are complete.

**Phase 6: Renderer Replacement Spike** is the current work on `master`.

Phase 5 browser MVP polish was reported complete, and Phase 5.5 closed PM verification gaps: stable album cache churn when full, an extra-small local testing cache preset, recovered ready-status publication after outages, and final status/doc alignment.

Matthew verified the daemon/cache/setup path against a personal Immich instance. The Raspberry Pi Chromium kiosk direction is paused: Chromium showed enough performance and outage-recovery risk that Phase 6 is choosing/prototyping a lighter renderer while reusing the daemon, cache, source, and Immich integration behavior.

Current Phase 6 direction:

- Primary path: Go + SDL2 native appliance renderer.
- Fallback path: pre-composited framebuffer/image-viewer renderer for very weak hardware or SDL packaging trouble.
- Implemented prototype foundation: local-only `GET /api/renderer/state`, `internal/renderer` retained-frame loop, and `immich-frame renderer-poc` for a Windows-friendly fixture preview.

Start here:

- `AGENT_BRIEF.md` for future coding agents.
- `GOAL.md` for definition of done.
- `docs/implementation-plan.md` for build phases.
- `docs/architecture.md` for system design.
- `docs/development.md` for local development expectations.
- `docs/developer-guide.md` for the human developer workflow and project map.
- `docs/local-development.md` for desktop run, test, and manual verification steps.

## Product Direction

Immich Frame runs locally, connects outbound to a user's Immich server, caches display-appropriate photo renditions locally, and currently includes a fullscreen browser reference slideshow with subtle overlays.

It is not primarily a hosted web app, Docker dashboard, or cloud service.

## Current Target

- Local frame MVP.
- Go daemon.
- Preact/Vite reference frame UI and setup UI.
- Embedded browser assets for release-style serving.
- Lightweight renderer/hardware path after Phase 6 selects a direction.

## MVP Scope

- Local Go daemon.
- Preact reference frame UI and setup UI.
- pnpm frontend tooling.
- Same-Wi-Fi setup portal.
- Dedicated Immich API key per frame.
- Setup completion requires a successful Immich connection test for the saved URL/API key.
- Album and random-library modes.
- Display-targeted local image cache.
- Playback-driven rolling cache refresh that avoids looping one static cached seed forever.
- Clock, photo info, and operational status overlays.
- Hidden controls.
- Reboot recovery.
- Cache-first outage behavior.

## Out Of Scope For MVP

- Temporary setup Wi-Fi network.
- Weather provider.
- Favorites and on-this-day modes.
- GPIO buttons.
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

Render the appliance renderer proof-of-concept preview:

```sh
go run ./cmd/immich-frame renderer-poc -image dev/photos/indy.jpg -out .immich-frame/renderer-poc.png -width 800 -height 480
```

Add `--logs` to print cache/playback activity such as refresh triggers, candidate counts, cache before/after counts, fetched/rotated/evicted counts, and slideshow cache hits:

```sh
go run ./cmd/immich-frame serve -config config.dev.toml -dev-source dev/photos --logs
```

Then open:

- `http://127.0.0.1:8787/frame`
- `http://127.0.0.1:8787/setup`

## License Plan

The intended public license is MIT once the reference Pi Zero 2 W install works.
