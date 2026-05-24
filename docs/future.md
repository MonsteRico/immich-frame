# Future Roadmap

These features are intentionally out of MVP but should influence extension points.

## Setup And Distribution

- Temporary Wi-Fi access point setup.
- Captive portal for nontechnical users.
- QR code setup URL.
- Flashable Raspberry Pi image.
- Explicit `immich-frame update` command with checksum/signature verification.
- Public open-source release after the reference Pi Zero 2 W install works.
- MIT license unless a later concrete reason changes it.

## Source Modes

- Favorites.
- On this day.
- People/person.
- Location.
- Date range.
- Smart rules.
- Multiple albums.
- Weighted/mixed albums.

The source provider interface should make these additions possible without rewriting playback/cache.

## Overlays

Near-future:

- Weather overlay.

Possible later overlays:

- Photo location label.
- Camera/lens info.
- Album/source label.
- Sync health/debug stats.
- Calendar/holiday-aware overlays.

Overlay contribution model remains source-level reviewed contributions, compiled into the app. Runtime plugin loading is not planned for MVP or near future.

## Hardware

- Hardware setup is paused until Phase 6 chooses a renderer direction.
- The Pi Zero 2 W Chromium kiosk path appears likely to struggle and showed reference-renderer outage recovery issues, so the next hardware pass should evaluate a lighter renderer first.
- Future install/systemd work should reuse the daemon, Immich adapter, setup portal, cache, playback, and settings behavior rather than rebuilding the whole app.
- GPIO buttons.
- IR remote.
- Bluetooth remote.
- Display sleep/wake schedule.
- Brightness/dimming.
- Overlay burn-in mitigation such as subtle position nudging or dimming.
- HDMI CEC.
- Motion sensor wake.
- Renderer contract hardening after Phase 6 chooses the appliance renderer.

## Media

Video is not important for MVP. Future optional behaviors could include:

- show first frame only.
- muted short clips.
- skip long videos.
- opt-in videos from selected albums only.

## Renditions

MVP uses the best Immich-provided non-original rendition appropriate for display target.

The daemon now keeps cache rotation and outage handling out of the renderer, so the Phase 6 renderer should only need to consume current/next local media URLs and status state instead of reimplementing Immich refresh or cache policy.

Future options:

- Higher-quality rendition selection for 4K displays.
- Optional original download and local resize.
- Optional companion service near Immich to create frame-sized renditions for multiple/high-res frames.

The companion service is out of scope for near-term work.

## Modes Not Planned

- Docker/LAN deployment mode.
- Hosted web service for single frames.
- Full admin dashboard.
- Runtime third-party overlay plugin sandbox.
- Strong physical anti-theft security.
