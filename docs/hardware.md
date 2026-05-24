# Hardware

## Reference Device

Immich Frame MVP targets:

- Raspberry Pi Zero 2 W.
- Raspberry Pi OS Lite.
- HDMI display.
- Wi-Fi.
- Chromium kiosk.
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

## Browser/Kiosk

Default browser is Raspberry Pi OS Chromium. Browser command should be configurable.

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

The Phase 4 installer writes the editable kiosk env file to:

```text
/etc/immich-frame/kiosk.env
```

The installed kiosk systemd service starts X with `startx` and runs:

```text
/usr/local/lib/immich-frame/start-kiosk.sh
```

That launcher disables screen blanking with `xset`, hides the pointer with `unclutter`, starts `openbox-session`, and opens Chromium at the kiosk URL.

## Display Server

The prepared Raspberry Pi OS Lite path is:

```text
systemd service -> startx -> Xorg -> openbox -> Chromium kiosk
```

Required packages are installed by `install.sh`:

```text
xserver-xorg
xinit
openbox
chromium-browser or chromium
unclutter
x11-xserver-utils
dbus-x11
fonts-dejavu-core
```

Physical Pi verification of this display path is pending.

Keep display-server specifics isolated in installer/systemd files.

Do not tie the Go daemon to X11, Wayland, Labwc, Openbox, or any specific display server.

## mDNS

Phase 4 uses Avahi for `frame.local` on Raspberry Pi OS Lite.

The installer installs and enables:

```text
avahi-daemon.service
```

Expected setup URL on same Wi-Fi:

```text
http://frame.local:8787/setup
```

The HDMI setup screen and runbook keep IP fallback as the required backup path because local mDNS behavior depends on the client OS and network.

Physical Pi and same-Wi-Fi client verification of `frame.local` is pending.

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
- Native renderer for weaker boards.
