# Immich Frame

A self-hosted digital picture frame for Immich libraries, designed for self-built HDMI frames powered by Raspberry Pi Zero 2 W-class hardware.

## Status

Planning scaffold. Implementation has not started yet.

Start here:

- `AGENT_BRIEF.md` for future coding agents.
- `GOAL.md` for definition of done.
- `docs/implementation-plan.md` for build phases.
- `docs/architecture.md` for system design.
- `docs/development.md` for local development expectations.

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

## License Plan

The intended public license is MIT once the reference Pi Zero 2 W install works.
