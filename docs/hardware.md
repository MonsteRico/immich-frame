# Hardware

## Hardware Status

Hardware/appliance setup is paused while Phase 6 chooses and prototypes the renderer direction.

The previous Raspberry Pi Zero 2 W + Chromium kiosk direction is no longer the active implementation path. Keep this document as hardware context, but do not add new installer/systemd/kiosk work until the browser MVP is reliable and a lighter renderer direction is chosen.

## Reference Device

The intended physical frame class is still:

- Raspberry Pi Zero 2 W.
- Raspberry Pi OS Lite.
- HDMI display.
- Wi-Fi.
- Lightweight renderer to be chosen by the Phase 6 spike.
- No touch required.
- No keyboard required after setup.

## Setup Assumptions For MVP

Wi-Fi is configured before Immich Frame setup, likely through Raspberry Pi Imager.

The frame and setup phone/laptop must be on the same network.

The frame displays:

```text
http://frame.local:8787/setup
http://<frame-ip>:8787/setup
setup code
```

## Display Detection

MVP uses browser-reported display size:

```text
window.innerWidth
window.innerHeight
devicePixelRatio
screen.width/screen.height if useful
```

The frame UI reports this to the daemon. The daemon uses it as the cache rendition target.

Fallback target:

```text
1920x1080
```

Manual override should be available in config/setup.

## Orientation

Default orientation is auto from viewport:

```text
width >= height -> landscape
height > width -> portrait
```

Config override:

```toml
[display]
orientation = "auto" # auto | landscape | portrait
```

MVP does not manage OS-level display rotation.

## Browser/Kiosk Experiment

Chromium kiosk was the first explored path, but it is likely too heavy for the Pi Zero 2 W. Future hardware work should evaluate a lighter renderer before reviving install scripts.

Kiosk URL:

```text
http://127.0.0.1:8787/frame
```

Kiosk config may live in an env file such as:

```sh
KIOSK_BROWSER=chromium-browser
KIOSK_URL=http://127.0.0.1:8787/frame
KIOSK_FLAGS="--kiosk --noerrdialogs --disable-infobars"
```

## Phase 6 Renderer Direction

Current recommendation before proof-of-concept work:

- Primary path: a Go + SDL2 native appliance renderer that consumes daemon-owned state and cached media through a narrow local renderer contract.
- Fallback path: an ultra-light framebuffer/image-viewer process where the daemon or helper pre-composites the selected photo plus minimal overlay text into a display-sized image.

The primary path is preferred because it keeps implementation in Go, avoids a web engine, supports fullscreen image scaling and simple overlays, and can keep the "decode next image before swap" outage behavior inside a deterministic renderer loop.

The fallback path is intentionally smaller. It may sacrifice crossfade and richer overlays, but it gives very weak boards a credible way to keep showing cached photos if SDL packaging or drivers are not acceptable on the target.

Do not add installer, systemd, autostart, kiosk, or OS-image work until the renderer proof of concept has been discussed and accepted.

## Display Server

Document one tested Raspberry Pi OS Lite path. Keep display-server specifics isolated in installer/systemd files.

Do not tie the Go daemon to X11, Wayland, Labwc, Openbox, or any specific display server.

## Future Hardware Features

Not MVP:

- Temporary setup Wi-Fi access point.
- Captive portal.
- GPIO buttons.
- IR remote.
- Bluetooth remote.
- Display sleep/wake schedule.
- Brightness/dimming.
- HDMI CEC.
- Motion sensor wake.
- Renderer packaging for weaker boards after the Phase 6 proof of concept is accepted.
