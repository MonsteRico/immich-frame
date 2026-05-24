# Implementation Plan

## Phase 0: Scaffold

Goal: create the project shape and enough build plumbing that future phases have a stable foundation.

Checklist:

- [ ] Initialize repo under `D:/Coding/immich-frame`.
- [ ] Add MIT license.
- [ ] Initialize single Go module at repo root.
- [ ] Add `cmd/immich-frame/main.go`.
- [ ] Add CLI commands: `serve`, `status`, `reset`, `config validate`, `version`.
- [ ] Add internal package directories.
- [ ] Add config/secrets/state models and default paths.
- [ ] Add Preact/Vite `ui/frame` bundle.
- [ ] Add Preact/Vite `ui/setup` bundle.
- [ ] Add `ui/shared` package.
- [ ] Use pnpm for frontend tooling.
- [ ] Use plain CSS/CSS modules and minimal inline SVG icons.
- [ ] Add embedded UI build path for release.
- [ ] Add development config example.
- [ ] Add initial docs links from README.

## Phase 1: Local Frame Loop

Goal: run a mock/local-folder slideshow locally in a desktop browser without Immich.

Checklist:

- [ ] Add local folder source provider for dev only.
- [ ] Add asset candidate model.
- [ ] Add cache manifest model.
- [ ] Add local media serving from cache/dev source.
- [ ] Add playback queue with current, next, previous, pause/resume.
- [ ] Add SSE event stream.
- [ ] Add `/api/state`.
- [ ] Add `/api/playback/*` commands.
- [ ] Add `/frame` UI slideshow.
- [ ] Add cut/crossfade transitions.
- [ ] Add hidden controls and keyboard shortcuts.
- [ ] Add clock/photo-info/status overlay placeholders.
- [ ] Add browser display-size reporting endpoint.

## Phase 1.5: Base Validation And Embedding Cleanup

Goal: tighten the local foundation before starting the real Immich adapter.

Checklist:

- [ ] Verify Go toolchain availability and run `go test ./...`.
- [ ] Add unit tests for config defaults/validation.
- [ ] Add unit tests for local folder candidate discovery.
- [ ] Add unit tests for cache manifest ensure/list/mark-shown behavior.
- [ ] Add unit tests for playback next/previous/pause/resume behavior.
- [ ] Reconcile embedded release UI behavior.
- [ ] If release embedding uses Vite output, serve embedded `/assets/*` from `embed.FS`.
- [ ] If release embedding uses a deliberate inline fallback, document that and keep `build:embedded-ui` from producing broken asset references.
- [ ] Run `pnpm typecheck`.
- [ ] Run `pnpm build`.
- [ ] Manually verify the local mock slideshow in a desktop browser when possible.

## Phase 2: Immich Adapter

Goal: connect to real Immich through an isolated adapter.

Checklist:

- [ ] Verify current Immich API through official docs/OpenAPI.
- [ ] Manually test required behavior against Matthew's Immich instance.
- [ ] Implement connection test.
- [ ] Implement album listing.
- [ ] Implement album asset candidate listing.
- [ ] Implement random-library candidate listing.
- [ ] Implement display-targeted rendition fetch.
- [ ] Normalize Immich metadata into minimal display metadata.
- [ ] Add mock HTTP unit tests for adapter request/response behavior.
- [ ] Keep all endpoint assumptions isolated in `internal/immich`.

No built-in real-Immich integration tests for MVP.

## Phase 3: Setup Portal

Goal: configure the frame from a phone/laptop on same Wi-Fi.

Checklist:

- [x] Add setup state model and backend boundary.
- [x] Generate fixed first-boot setup code until setup completes.
- [x] Render HDMI setup state with URL, IP fallback, code, and status.
- [x] Add `/setup` portal.
- [x] Claim setup code.
- [x] Create local admin password.
- [x] Store password hash in `secrets.json`.
- [x] Add admin login/logout/session behavior.
- [x] Add settings read/write API backed by `config.toml`.
- [x] Accept raw Immich URL and raw Immich API key fields.
- [x] Validate Immich connection.
- [x] Save Immich URL to config and API key to secrets.
- [x] Select source mode: one album or random library.
- [x] Provide a searchable album picker for album mode, showing album name and item count when available.
- [x] Configure interval, display fit, cache preset, and overlays.
- [x] Add lightweight status page.
- [x] Require admin session after setup.
- [x] Update setup/security/config/developer docs alongside behavior changes.

## Phase 3.5: Setup Portal Hardening

Goal: close setup portal gaps before Raspberry Pi appliance work starts.

Checklist:

- [x] Require a successful Immich validation for the saved URL/API key before setup completion.
- [x] Prevent random-library setup from bypassing Immich validation.
- [x] Add clear setup UI guardrails and messages for missing validation or required fields.
- [x] Add a lightweight status API/settings surface.
- [x] Reconcile overlay configuration docs with implemented backend behavior.
- [x] Update README, agent handoff docs, and developer docs.
- [x] Run `go test ./...`, `pnpm typecheck`, and `pnpm build`.

## Phase 4: Pi Appliance Experiment - Reverted

Goal was to make the reference Pi Zero 2 W boot directly into the frame.

Status: reverted/paused. The Chromium kiosk path looks too heavy for the Pi Zero 2 W, and hardware setup work is premature until Phase 6 chooses a lighter renderer direction.

Checklist:

- [x] Remove installer/systemd/kiosk assets from the active codebase.
- [x] Pause Raspberry Pi Chromium kiosk work.
- [ ] Revisit hardware after Phase 6 chooses a lighter rendering strategy.

## Phase 5: Browser MVP Polish And Hardening

Goal: make the browser-based MVP reliable enough to use daily before swapping or replacing the renderer.

Checklist:

- [x] Implement cache rotation and eviction policy:
  - [x] Refresh Immich candidate pools periodically.
  - [x] Top off cached display-targeted renditions toward `target_items`.
  - [x] Maintain a near-term prefetch buffer from `prefetch_items`.
  - [x] Prefer never-shown and least-recently-shown candidates.
  - [x] Evict entries that left the source before evicting valid offline fallback photos.
  - [x] Avoid evicting current and near-upcoming playback entries.
  - [x] Refresh playback queue when cache contents change so the frame does not loop one static seed forever.
- [x] Implement outage retry/backoff.
- [x] Implement subtle degraded status overlay.
- [x] Implement reset flows.
- [x] Implement CLI status details.
- [x] Add unit tests for playback/cache/config/auth.
- [x] Keep tests unit-only; do not add real-Immich integration tests to repo/CI for MVP.
- [x] Add frontend typecheck/build verification.
- [x] Update all docs.
- [x] Confirm no Immich secrets reach browser or logs.
- [x] Confirm LAN clients cannot view cached photos unauthenticated.

## Phase 5.5: Browser MVP Acceptance Fix Pass

Goal: close PM verification gaps found after Phase 5 was reported complete, without restarting hardware/appliance work.

Checklist:

- [x] Rotate/churn stable album caches even when already at `cache.target_items`.
- [x] Preserve current and near-upcoming playback entries during album churn using `cache.prefetch_items`.
- [x] Add an `extra-small` local testing cache preset with roughly 10 cached photos.
- [x] Publish recovered `ready` status after outage retry succeeds, even if cache contents did not change.
- [x] Align docs to show Phase 5.5 as the current fix pass until final verification is complete.
- [x] Run final `go test ./...`, `pnpm typecheck`, `pnpm build`, and `pnpm build:embedded-ui` because setup UI assets changed.

## Post-Phase 5.5: Rolling Cache Strategy Adjustment

Goal: apply the approved product change that the frame should steadily progress through the wider album/library instead of gently adding one new cached image per timer refresh.

Checklist:

- [x] Add playback-driven rolling cache refresh after a configurable number of shown images.
- [x] Add configurable rolling refresh batch size.
- [x] Swap shown, non-protected cached entries for new album/random-library candidates.
- [x] Preserve current and near-upcoming playback entries.
- [x] Request immediate cache maintenance after first setup completes with an Immich source.
- [x] Document that this strategy change was requested and approved after Phase 5.5.

## Phase 6: Renderer And Hardware Re-evaluation

Goal: choose and prototype the appliance renderer after the browser/reference MVP proved the daemon, setup, Immich adapter, cache, and playback system.

Context:

- Phase 5.5 is complete.
- Matthew verified the app with a personal Immich instance/API key.
- Random mode rotates, `extra-small` cache rotation is visible, restart/reboot simulation resumes without setup, and embedded release-style browser serving works.
- During a network outage, the daemon appears to keep cache-first playback and reconnect correctly, but the Chromium tab does not keep the visual slideshow moving or resume cleanly after reconnect.
- Because Chromium kiosk is likely too heavy or fragile for Pi Zero 2 W-class hardware, do not make the next phase a browser outage-hardening phase.

Checklist:

- [x] Define renderer boundary between daemon state/media APIs and presentation layer.
- [x] Decide whether the appliance renderer consumes `/api/state` and `/media/:assetID`, a local-only renderer API, direct cache file paths, or another narrow contract.
- [x] Prefer a resilient polling or hybrid update loop over SSE-only rendering.
- [x] Evaluate at least three lightweight renderer options for Pi Zero 2 W-class hardware.
- [x] Score renderer options on footprint, image quality, overlay feasibility, packaging complexity, Go integration, and testability without hardware.
- [x] Recommend one primary renderer path and one fallback path.
- [x] Build a narrow proof of concept for the recommended path.
- [x] Show at least one cached/local image and one simple overlay in the proof of concept.
- [x] Keep the current image visible if state/media refresh fails.
- [ ] Reuse existing daemon, setup portal, Immich adapter, cache, playback, and settings behavior.
- [ ] Avoid restarting installer/systemd work until a renderer direction is chosen.
- [ ] Re-plan hardware install steps around the chosen renderer.

### Phase 6 Renderer Contract Decision

The appliance renderer should be a presentation client for daemon-owned state. It should not call Immich, manage cache refresh, choose playback order, write setup/settings, or receive secrets. The browser `/frame` path remains useful as a development renderer, but the appliance renderer should start from a narrow local-only contract rather than inheriting the browser's SSE-first behavior.

Contract decision:

- Keep setup/settings browser-based.
- Keep `/api/state`, `/api/events`, and `/media/:assetID` for the browser reference renderer.
- Add a future local-only renderer snapshot endpoint, tentatively `GET /api/renderer/state`, for native renderers.
- Shape the snapshot around current media, optional next media, status/message, overlay config, playback interval/paused state, display fit, and display target.
- Prefer snapshot polling with optional event wake-ups. SSE may reduce latency, but missed events must not be fatal.
- Allow direct local cache file references only on a renderer-local API or process boundary. Do not expose cache paths to LAN/browser callers.
- Require the renderer to keep the last successfully decoded image visible when state or media refresh fails.

### Phase 6 Renderer Option Evaluation

External reference points reviewed during the spike:

- `go-sdl2` wraps SDL2 for Go, requires the C SDL2 installation, and documents that first builds can take several minutes on weaker machines such as Raspberry Pi.
- Rust SDL2 offers current Cargo integration and optional image/ttf/gfx features, but would add a second implementation language.
- Qt for Embedded Linux can run fullscreen without X11/Wayland through EGLFS/LinuxFB, but Qt's own docs warn that high-resolution Qt Quick/OpenGL paths can need at least 128 MB of GPU memory.
- WPE WebKit is designed for embedded, low-consumption devices and Raspberry Pi-class backends, but it remains a web engine and keeps much of the browser-runtime complexity Phase 6 is trying to avoid.
- Framebuffer image viewers such as `fbi` are available on Raspberry Pi OS-family systems and are very light, but they are process-oriented image display tools rather than app UI renderers.

Score scale: 1 is poor, 5 is strong for this project.

| Option | Footprint | Image quality | Overlays/crossfade | Packaging | Go integration | Windows/WSL testability | Pi availability | Total | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Go + SDL2 renderer | 4 | 4 | 4 | 3 | 5 | 3 | 4 | 27 | Best fit for a first native prototype: same repo/language, direct polling loop, good enough image/text primitives. Main risk is cgo/SDL2 setup on Windows and target images. |
| Rust + SDL2 helper | 4 | 4 | 4 | 3 | 2 | 4 | 4 | 25 | Similar runtime profile with a stronger Rust desktop story, but adds Cargo, FFI/process protocol, and a second production language. |
| Framebuffer/image-viewer process (`fbi`/similar) | 5 | 3 | 1 | 4 | 3 | 1 | 4 | 21 | Very light fallback for showing photos on weak boards. Poor native overlay/crossfade story unless daemon pre-composites images. |
| Qt/QML or Qt Widgets | 2 | 5 | 5 | 2 | 2 | 4 | 3 | 23 | Capable visuals and overlays, but heavier packaging/runtime than the product needs for Pi Zero 2 W-class hardware. |
| WPE WebKit/lightweight browser | 2 | 4 | 5 | 2 | 2 | 3 | 3 | 21 | Lighter than Chromium in embedded contexts but still a web renderer with browser recovery and packaging concerns. |

Recommended primary path:

- Prototype a Go + SDL2 appliance renderer behind a clearly named experimental command/package.
- Keep renderer logic split so state adaptation, outage behavior, and image-selection decisions are unit-testable without SDL.
- Consume a fixture first, then the future local-only renderer snapshot contract.
- Render one display-sized cached image and one text overlay, with the SDL shell responsible for taking that loop fullscreen on target hardware.
- Implement the core loop as "snapshot, decode next, then swap"; on any state/media failure, keep showing the last decoded image.

Recommended fallback path:

- If Go + SDL2 proves too painful on the target or Windows development machine, fall back to an ultra-light framebuffer/image-viewer path where the daemon or a small helper pre-composites the selected photo plus minimal status/clock text into a display-sized image and an external viewer displays it.
- This fallback intentionally scopes down transitions and rich overlays. It preserves the most important appliance behavior: keep showing cached photos through network/daemon refresh trouble on very weak hardware.

Rejected for the first proof of concept:

- Qt/QML: excellent UI capability, but too much runtime and packaging weight for the current question.
- WPE/WebKit: credible embedded web path, but it keeps the project in a browser-engine failure mode before proving a truly lighter renderer.
- Deep Chromium hardening: explicitly out of scope for Phase 6 because the browser renderer is now a reference/development renderer.

### Phase 6 Proof-Of-Concept Slice

Implemented proof-of-concept surface:

- `internal/renderer` contains the local renderer snapshot contract and a small retained-frame loop.
- `GET /api/renderer/state` is local-only and adapts daemon playback/config/cache state into renderer input. It can include local cache file paths for the renderer process, but rejects non-loopback callers.
- `immich-frame renderer-poc` renders a cached/local image through the renderer contract into a PNG preview with a simple status/clock overlay. This keeps the contract and image fit behavior testable on Windows without requiring SDL to be installed.
- Unit tests prove renderer snapshot adaptation, local-only API behavior, preview generation, and the outage invariant that a previously decoded image remains visible when a later state fetch or media decode fails.

Remaining after this slice:

- Add the SDL display shell around `internal/renderer` after confirming the contract/loop behavior is acceptable.
- Re-plan Pi install/systemd/autostart around the chosen renderer only after SDL packaging and target-device behavior are tested.

## MVP Exclusions

- [ ] Temporary Wi-Fi AP setup.
- [ ] Weather provider.
- [ ] Favorites source.
- [ ] On-this-day source.
- [ ] GPIO/remote input.
- [ ] Video playback.
- [ ] Flashable image.
- [ ] Auto-updates.
- [ ] Docker/LAN mode.
- [ ] Runtime third-party overlay plugins.
