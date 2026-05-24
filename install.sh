#!/usr/bin/env bash
set -euo pipefail

APP_NAME="immich-frame"
SERVICE_USER="immich-frame"
SERVICE_GROUP="immich-frame"
KIOSK_USER="immich-frame-kiosk"
CONFIG_DIR="/etc/immich-frame"
DATA_DIR="/var/lib/immich-frame"
BIN_PATH="/usr/local/bin/immich-frame"
LIB_DIR="/usr/local/lib/immich-frame"
DRY_RUN=0
SKIP_BUILD=0

usage() {
  cat <<'USAGE'
Usage: sudo ./install.sh [--dry-run] [--skip-build]

Installs Immich Frame for Raspberry Pi OS Lite:
  - builds the embedded UI and Go binary unless --skip-build is set
  - creates immich-frame service user/group
  - creates runtime config and data directories
  - installs systemd daemon and Chromium kiosk units
  - installs Avahi mDNS dependency for frame.local

The installer is idempotent and preserves existing config, secrets, state, and cache.
USAGE
}

log() {
  printf '%s\n' "$*"
}

run() {
  if [ "$DRY_RUN" -eq 1 ]; then
    printf '[dry-run] %q' "$1"
    shift
    for arg in "$@"; do
      printf ' %q' "$arg"
    done
    printf '\n'
    return 0
  fi
  "$@"
}

require_root() {
  if [ "$(id -u)" -ne 0 ]; then
    log "error: run as root, for example: sudo ./install.sh"
    exit 1
  fi
}

require_repo_root() {
  if [ ! -f "go.mod" ] || [ ! -f "package.json" ] || [ ! -d "cmd/immich-frame" ]; then
    log "error: run this script from the immich-frame repository root"
    exit 1
  fi
}

install_packages() {
  export DEBIAN_FRONTEND=noninteractive
  chromium_package="chromium-browser"
  if [ "$DRY_RUN" -eq 0 ] && ! apt-cache show chromium-browser >/dev/null 2>&1; then
    chromium_package="chromium"
  fi
  run apt-get update
  run apt-get install -y --no-install-recommends \
    avahi-daemon \
    ca-certificates \
    "$chromium_package" \
    dbus-x11 \
    fonts-dejavu-core \
    openbox \
    unclutter \
    x11-xserver-utils \
    xinit \
    xserver-xorg
}

build_binary() {
  if [ "$SKIP_BUILD" -eq 1 ]; then
    log "Skipping local build; expecting an existing immich-frame binary at $BIN_PATH"
    return 0
  fi
  run pnpm build:embedded-ui
  run go build -trimpath -ldflags "-s -w" -o ".dist/immich-frame" ./cmd/immich-frame
  run install -D -m 0755 ".dist/immich-frame" "$BIN_PATH"
}

ensure_users() {
  if ! getent group "$SERVICE_GROUP" >/dev/null; then
    run groupadd --system "$SERVICE_GROUP"
  fi
  if ! id -u "$SERVICE_USER" >/dev/null 2>&1; then
    run useradd --system --gid "$SERVICE_GROUP" --home-dir "$DATA_DIR" --shell /usr/sbin/nologin "$SERVICE_USER"
  fi
  if ! id -u "$KIOSK_USER" >/dev/null 2>&1; then
    run useradd --system --create-home --home-dir "$DATA_DIR/kiosk" --shell /bin/bash "$KIOSK_USER"
  fi
  for group in audio video input render; do
    if getent group "$group" >/dev/null; then
      run usermod -aG "$group" "$KIOSK_USER"
    fi
  done
}

ensure_filesystem() {
  run install -d -m 0775 -o root -g "$SERVICE_GROUP" "$CONFIG_DIR"
  run install -d -m 0750 -o "$SERVICE_USER" -g "$SERVICE_GROUP" "$DATA_DIR"
  run install -d -m 0750 -o "$SERVICE_USER" -g "$SERVICE_GROUP" "$DATA_DIR/cache"
  run install -d -m 0750 -o "$KIOSK_USER" -g "$KIOSK_USER" "$DATA_DIR/kiosk"
  run install -d -m 0755 -o root -g root "$LIB_DIR"

  if [ ! -f "$CONFIG_DIR/config.toml" ]; then
    run install -m 0664 -o root -g "$SERVICE_GROUP" "packaging/config/appliance-config.toml" "$CONFIG_DIR/config.toml"
  else
    run chown root:"$SERVICE_GROUP" "$CONFIG_DIR/config.toml"
    run chmod 0664 "$CONFIG_DIR/config.toml"
  fi

  if [ -f "$DATA_DIR/secrets.json" ]; then
    run chown "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR/secrets.json"
    run chmod 0600 "$DATA_DIR/secrets.json"
  fi
  if [ -f "$DATA_DIR/state.json" ]; then
    run chown "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR/state.json"
    run chmod 0644 "$DATA_DIR/state.json"
  fi
}

install_units() {
  run install -m 0644 -o root -g root "packaging/systemd/immich-frame.service" "/etc/systemd/system/immich-frame.service"
  run install -m 0644 -o root -g root "packaging/systemd/immich-frame-kiosk.service" "/etc/systemd/system/immich-frame-kiosk.service"
  if [ ! -f "$CONFIG_DIR/kiosk.env" ]; then
    run install -m 0644 -o root -g root "packaging/kiosk/immich-frame-kiosk.env" "$CONFIG_DIR/kiosk.env"
  else
    run chown root:root "$CONFIG_DIR/kiosk.env"
    run chmod 0644 "$CONFIG_DIR/kiosk.env"
  fi
  run install -m 0755 -o root -g root "packaging/kiosk/start-kiosk.sh" "$LIB_DIR/start-kiosk.sh"
  run systemctl daemon-reload
  run systemctl enable avahi-daemon.service
  run systemctl enable immich-frame.service
  run systemctl enable immich-frame-kiosk.service
}

validate_assets() {
  for path in \
    packaging/config/appliance-config.toml \
    packaging/systemd/immich-frame.service \
    packaging/systemd/immich-frame-kiosk.service \
    packaging/kiosk/immich-frame-kiosk.env \
    packaging/kiosk/start-kiosk.sh; do
    if [ ! -f "$path" ]; then
      log "error: missing $path"
      exit 1
    fi
  done
}

main() {
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --dry-run) DRY_RUN=1 ;;
      --skip-build) SKIP_BUILD=1 ;;
      -h|--help) usage; exit 0 ;;
      *) log "error: unknown argument $1"; usage; exit 1 ;;
    esac
    shift
  done

  require_repo_root
  validate_assets
  if [ "$DRY_RUN" -eq 0 ]; then
    require_root
  fi

  install_packages
  build_binary
  ensure_users
  ensure_filesystem
  install_units

  log "Immich Frame install prepared."
  log "Next on the Pi:"
  log "  sudo systemctl start immich-frame.service"
  log "  sudo systemctl start immich-frame-kiosk.service"
  log "  systemctl status immich-frame.service --no-pager"
  log "  systemctl status immich-frame-kiosk.service --no-pager"
}

main "$@"
