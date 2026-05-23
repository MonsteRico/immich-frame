#!/usr/bin/env bash
set -euo pipefail

KIOSK_BROWSER="${KIOSK_BROWSER:-chromium-browser}"
KIOSK_URL="${KIOSK_URL:-http://127.0.0.1:8787/frame}"
KIOSK_FLAGS="${KIOSK_FLAGS:---kiosk --noerrdialogs --disable-infobars}"

xset -dpms || true
xset s off || true
xset s noblank || true
unclutter -idle 0.2 -root >/dev/null 2>&1 &
openbox-session >/dev/null 2>&1 &

exec "$KIOSK_BROWSER" $KIOSK_FLAGS "$KIOSK_URL"
