#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

bash -n install.sh
bash -n packaging/kiosk/start-kiosk.sh

go run ./cmd/immich-frame config validate -config packaging/config/appliance-config.toml

if command -v systemd-analyze >/dev/null 2>&1; then
  systemd-analyze verify \
    packaging/systemd/immich-frame.service \
    packaging/systemd/immich-frame-kiosk.service
else
  echo "systemd-analyze not available; skipped unit verification"
fi

./install.sh --dry-run --skip-build >/dev/null
echo "Pi install assets validated"
