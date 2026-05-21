# Immich Frame

A self-hosted digital picture frame for Immich libraries, designed for self-built HDMI frames powered by Raspberry Pi Zero 2 W-class hardware.

## Status

Phase 0, the local Phase 1 mock slideshow loop, Phase 1.5 validation, and Phase 2 Immich adapter are complete. The next major work is Phase 3: the phone-first setup/settings portal and supporting backend routes.

Start here:

- `AGENT_BRIEF.md` for future coding agents.
- `GOAL.md` for definition of done.
- `docs/implementation-plan.md` for build phases.
- `docs/architecture.md` for system design.
- `docs/development.md` for local development expectations.
- `docs/developer-guide.md` for the human developer workflow and project map.
- `docs/local-development.md` for desktop run, test, and manual verification steps.

## Product Direction

Immich Frame runs locally on the frame device. It connects outbound to a user's Immich server, caches display-appropriate photo renditions locally, and renders a fullscreen Chromium kiosk slideshow with subtle overlays.

It is not primarily a hosted web app, Docker dashboard, or cloud service.

## Reference Hardware

- Raspberry Pi Zero 2 W.
- Raspberry Pi OS Lite.
- HDMI display.
- Wi-Fi configured before setup.
- Chromium kiosk.

## MVP Scope

- Local Go daemon.
- Preact frame UI and setup UI.
- pnpm frontend tooling.
- Same-Wi-Fi setup portal.
- Dedicated Immich API key per frame.
- Album and random-library modes.
- Display-targeted local image cache.
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
