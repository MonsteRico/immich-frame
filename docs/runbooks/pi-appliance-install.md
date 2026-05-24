# Raspberry Pi Appliance Install Runbook

This runbook is for installing Immich Frame on the reference Raspberry Pi Zero 2 W.

Physical Pi verification is pending until Matthew runs these commands on the hardware and reports the output.

## Assumptions

- Raspberry Pi Zero 2 W.
- Ubuntu Server 24.04 LTS.
- Wi-Fi was configured before first boot, for example with cloud-init or the Ubuntu Raspberry Pi image setup flow.
- The Pi and setup phone/laptop are on the same Wi-Fi.
- The repository is available on the Pi.
- Compatible Node.js and pnpm are installed before running `install.sh`.
- The installer can install Ubuntu-packaged Go and runtime packages.

## Install

From the repository root on the Pi:

```sh
sudo ./install.sh
```

Do not run `pnpm setup` as part of this installer. Install Node.js and pnpm first, then let `install.sh` use the existing `node` and `pnpm` commands from `PATH`.

Check the prerequisite commands before installing:

```sh
node --version
pnpm --version
```

The installer is designed to be idempotent. Re-running it should:

- preserve `/etc/immich-frame/config.toml` if it already exists.
- preserve `/var/lib/immich-frame/secrets.json`.
- preserve `/var/lib/immich-frame/state.json`.
- preserve `/var/lib/immich-frame/cache/`.
- preserve `/etc/immich-frame/kiosk.env` if it already exists.
- refresh the binary, systemd units, and kiosk launcher.

To preview the actions without changing the Pi:

```sh
./install.sh --dry-run --skip-build
```

To validate install assets from a development shell with Bash:

```sh
./scripts/validate-pi-install-assets.sh
```

## Start Services

After install:

```sh
sudo systemctl start immich-frame.service
sudo systemctl start immich-frame-kiosk.service
```

Check daemon status:

```sh
systemctl status immich-frame.service --no-pager
journalctl -u immich-frame.service -n 80 --no-pager
```

Check kiosk status:

```sh
systemctl status immich-frame-kiosk.service --no-pager
journalctl -u immich-frame-kiosk.service -n 120 --no-pager
```

Check listening HTTP:

```sh
curl -I http://127.0.0.1:8787/frame
curl -I http://127.0.0.1:8787/setup
```

Check mDNS from another same-Wi-Fi device:

```text
http://frame.local:8787/setup
```

If `frame.local` does not resolve, use the IP shown on the HDMI setup screen or inspect the Pi:

```sh
hostname -I
systemctl status avahi-daemon.service --no-pager
```

## Kiosk Configuration

The kiosk opens:

```text
http://127.0.0.1:8787/frame
```

The browser command, URL, and flags live in:

```text
/etc/immich-frame/kiosk.env
```

After editing kiosk flags:

```sh
sudo systemctl restart immich-frame-kiosk.service
```

## Runtime Files

```text
/usr/local/bin/immich-frame
/etc/immich-frame/config.toml
/etc/immich-frame/kiosk.env
/var/lib/immich-frame/secrets.json
/var/lib/immich-frame/state.json
/var/lib/immich-frame/cache/
/usr/local/lib/immich-frame/start-kiosk.sh
```

Permissions are managed by `install.sh`:

- daemon service user: `immich-frame`.
- kiosk user: `immich-frame-kiosk`.
- `config.toml`: group-writable by `immich-frame`.
- `secrets.json`: owned by `immich-frame`, mode `0600`.
- cache/state directory: writable by `immich-frame`.

## Reset And Recovery

Show CLI status:

```sh
sudo -u immich-frame immich-frame status -data-dir /var/lib/immich-frame
```

Factory reset secrets, state, and cache:

```sh
sudo systemctl stop immich-frame.service
sudo -u immich-frame immich-frame reset -data-dir /var/lib/immich-frame
sudo systemctl start immich-frame.service
```

Troubleshooting reset while keeping cached media:

```sh
sudo systemctl stop immich-frame.service
sudo -u immich-frame immich-frame reset -data-dir /var/lib/immich-frame --keep-cache
sudo systemctl start immich-frame.service
```

## Hardware Verification Gate

After running the install on the physical Pi, send back:

```sh
cat /etc/os-release
uname -a
/snap/bin/chromium --version
systemctl status immich-frame.service --no-pager
systemctl status immich-frame-kiosk.service --no-pager
journalctl -u immich-frame.service -n 80 --no-pager
journalctl -u immich-frame-kiosk.service -n 120 --no-pager
curl -I http://127.0.0.1:8787/frame
curl -I http://127.0.0.1:8787/setup
hostname -I
```

Also report what appears on HDMI after:

```sh
sudo reboot
```

Do not mark boot, reboot, kiosk display, or `frame.local` behavior complete until this hardware output confirms it.
