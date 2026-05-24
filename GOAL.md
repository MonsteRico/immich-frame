# Immich Frame Goal

This document defines when an agent can stop working and what counts as done.

## Long-Term Product Goal

Immich Frame is complete enough for the local frame MVP when it can run as a reliable digital picture frame connected to Immich, using a renderer that is practical for Pi Zero 2 W-class hardware.

The current browser renderer and setup portal proved the daemon, setup, Immich adapter, cache, playback, and embedded UI foundation. The browser frame should now be treated as a development/reference renderer, not necessarily the final appliance runtime.

Hardware/appliance setup is intentionally paused until the renderer direction is chosen. The Raspberry Pi Zero 2 W Chromium kiosk experiment showed enough performance and outage-recovery risk that the next phase should evaluate a lighter rendering engine and reuse the existing daemon/cache/setup behavior.

## MVP Definition Of Done

The local frame MVP is done when all items below are complete:

- [x] Unconfigured HDMI/browser reference screen shows setup instructions, `frame.local:8787`, IP fallback, and first-boot setup code.
- [x] Setup portal accepts setup code.
- [x] User can create local admin password.
- [x] Setup portal accepts Immich URL and dedicated Immich API key.
- [x] Daemon validates Immich connection.
- [x] User can choose one album or random library mode.
- [x] Daemon caches first display-targeted photo renditions locally.
- [x] Cache rotation refreshes candidates, tops off target cache, and avoids cycling one static cache forever.
- [x] Random mode continues rotating against a personal Immich instance.
- [x] Extra-small cache preset makes rotations visible during testing.
- [x] Restart/reboot simulation resumes configured frame without setup.
- [x] Embedded release-style browser serving works.
- [x] Settings portal remains available after setup and requires admin auth.
- [x] `immich-frame status`, `reset`, `config validate`, `version`, and `serve --logs` work.
- [x] Unit tests pass.
- [x] Frontend typecheck/build passes.
- [x] Docs match implemented behavior through Phase 5.5.
- [ ] Replacement renderer direction is chosen for Pi Zero 2 W-class hardware.
- [ ] Replacement renderer displays cached slideshow from daemon state/media.
- [ ] Replacement renderer supports clock/photo info/degraded status overlays or explicitly scopes the MVP overlay subset.
- [ ] Replacement renderer continues showing cached photos through Immich/network outage.
- [ ] Replacement renderer recovers cleanly after network reconnect and shows/clears degraded state when daemon state changes.
- [ ] LAN media/settings security is verified from another device, or explicitly deferred with rationale.

## Current Session Goal

Complete **Phase 6: Renderer Replacement Spike**.

Phase 5.5 is complete on `master`. Matthew verified a personal Immich instance/API key, random-mode rotation, extra-small cache rotation visibility, restart/reboot resume, and embedded release-style serving. A network outage showed the daemon continues cache/playback/reconnect behavior correctly, but the Chromium tab does not visually recover. Because Chromium kiosk is already suspect on Pi Zero 2 W-class hardware, do not spend the next phase deeply hardening the browser renderer. Treat the browser frame as a reference/development renderer and use Phase 6 to choose and prototype a lighter appliance renderer that reuses the existing daemon/cache/setup system.

### Phase 0 Done Checklist

- [x] Repo structure exists.
- [x] Go module initialized.
- [x] CLI skeleton compiles.
- [x] Config/secrets/state file paths and data types exist.
- [x] Frame UI bundle scaffold exists.
- [x] Setup UI bundle scaffold exists.
- [x] Shared UI package exists.
- [x] Release embedding strategy is represented in code/build files.
- [x] Basic CI commands are documented or configured.
- [x] CI scope is Go unit tests plus frontend typecheck/build.

### Phase 1 Partial Done Checklist

- [x] Local folder/mock source can produce photo candidates.
- [x] Cache manifest can track cached local media.
- [x] Playback queue can advance current/next/previous in memory.
- [x] `/api/state` returns frame state.
- [x] `/api/events` streams state updates with SSE.
- [x] `/media/:assetID` serves local cached media.
- [x] `/frame` renders slideshow in desktop browser.
- [x] Clock/photo-info/status overlay placeholders render.

### Phase 1.5 Base Validation And Embedding Cleanup Checklist

- [x] Verify the Go toolchain is available in the development environment, or document the exact blocker.
- [x] Run `go test ./...` successfully once Go is available.
- [x] Add focused Go unit tests for the existing local base:
  - [x] config defaults/validation.
  - [x] local folder candidate discovery.
  - [x] cache manifest ensure/list/mark-shown behavior.
  - [x] playback next/previous/pause/resume behavior.
- [x] Reconcile release UI embedding:
  - [x] Decide whether embedded release UI should use the built Vite bundles or a deliberately separate inline fallback.
  - [x] If using Vite bundles, make embedded `/assets/*` serving work from `embed.FS`.
  - [x] If keeping inline fallback, document that choice and ensure `build:embedded-ui` does not create broken embedded asset references.
- [x] Run `pnpm typecheck`.
- [x] Run `pnpm build`.
- [x] Run the local mock slideshow manually in a desktop browser if the Go toolchain is available.
- [x] Record verification results and any remaining environment limitations in docs or commit notes.
- [x] Commit fixes in meaningful slices, not as one broad phase commit.

### Phase 1.5 Verification Notes - 2026-05-20

- Go is available as `go1.26.3 windows/386`.
- `go test ./...` passed after adding focused tests for config, source, cache, playback, and embedded UI serving.
- Release embedding now uses built Vite bundles. `pnpm build:embedded-ui` copies `ui/frame/dist` and `ui/setup/dist` into `internal/api/static`, and `/assets/*` falls back to `embed.FS` when external dist directories are absent.
- `pnpm typecheck` passed.
- `pnpm build` passed.
- `pnpm build:embedded-ui` passed.
- Local mock daemon was run with missing external dist paths to force embedded UI serving: `go run ./cmd/immich-frame serve -config config.dev.toml -dev-source dev/photos -data-dir .immich-frame-verify -frame-dist missing-frame-dist -setup-dist missing-setup-dist`.
- HTTP smoke checks against the running daemon passed: `/frame` 200, embedded frame JS asset 200, embedded frame CSS asset 200, `/api/state` returned ready local-folder state, `/media/:assetID` returned 200, and `POST /api/playback/pause` returned paused state.
- Desktop browser verification passed on 2026-05-21 using the Codex in-app Browser plugin against `http://127.0.0.1:8787/frame`. The daemon was run with missing external dist paths to force embedded UI serving. Browser evidence showed title `Immich Frame`, visible mock photo `sample-dawn`, clock/photo-info overlays, playback controls, local image source `/media/a0007e962ab0bb36`, embedded asset URLs `/assets/index-DW__ochF.js` and `/assets/index-CiPW52pQ.css`, and no console warnings/errors. HTTP checks also confirmed `/api/state`, the embedded JS/CSS assets, and `/media/a0007e962ab0bb36` returned `200`.

### Phase 1.5 PM Readiness Confirmation - 2026-05-21

- Repo is clean on `master`.
- `go test ./...` passed.
- Matthew's PowerShell profile initializes fnm, and agent shells should use the active Node/pnpm from that PowerShell environment.
- As of the cleanup after Phase 1.5, this resolves to Node `v24.12.0` and pnpm `11.x`.
- Frontend package typechecks passed with plain pnpm:
  - `pnpm --filter @immich-frame/shared typecheck`
  - `pnpm --filter @immich-frame/frame typecheck`
  - `pnpm --filter @immich-frame/setup typecheck`
- Frontend package builds passed with plain pnpm:
  - `pnpm --filter @immich-frame/shared build`
  - `pnpm --filter @immich-frame/frame build`
  - `pnpm --filter @immich-frame/setup build`
- The repo intentionally does not pin pnpm with `packageManager`; use the machine's active PowerShell `pnpm`.

### Phase 2 Immich Adapter Checklist

- [x] Baseline verification before changes:
  - [x] Confirm branch is `master`.
  - [x] Confirm remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
  - [x] Run `go test ./...`.
  - [x] Run `pnpm typecheck` and `pnpm build`.
- [x] Verify current Immich API behavior:
  - [x] Review official Immich docs/OpenAPI for required endpoints.
  - [x] Manually verify required behavior against Matthew's Immich instance when credentials/context are available.
  - [x] Record endpoint assumptions and version notes in docs or adapter comments.
- [x] Define the Immich adapter boundary in `internal/immich`:
  - [x] Connection/auth test method.
  - [x] Album listing method.
  - [x] Asset candidate listing for a selected album.
  - [x] Asset candidate listing for random library mode.
  - [x] Display-targeted rendition fetch method.
  - [x] Error normalization that hides low-level HTTP details from callers.
- [x] Implement Immich connection testing:
  - [x] Use the dedicated API key from secrets/config inputs.
  - [x] Return user-safe errors for invalid URL, invalid key, network failure, and incompatible response.
  - [x] Add mock HTTP unit tests.
- [x] Implement Immich album listing:
  - [x] Normalize album id, name, and item count when available.
  - [x] Keep setup/search UI needs in mind without implementing the full setup portal yet.
  - [x] Add mock HTTP unit tests.
- [x] Implement Immich asset candidate listing:
  - [x] Album source candidates.
  - [x] Random library candidates.
  - [x] Conservative default filters where the API supports them: photos only, exclude archived/hidden/trashed/videos.
  - [x] Avoid leaking raw Immich asset JSON outside `internal/immich`.
  - [x] Add mock HTTP unit tests.
- [x] Implement display-targeted rendition fetching:
  - [x] Prefer the best Immich-provided non-original rendition appropriate for the requested target.
  - [x] Preserve returned content type.
  - [x] Do not download originals by default.
  - [x] Add mock HTTP unit tests.
- [x] Normalize display metadata:
  - [x] Minimal asset id, media/rendition identity, taken date, source name, dimensions/orientation if available.
  - [x] Do not expose raw EXIF, GPS coordinates, file paths, people names, direct Immich URLs, or full Immich asset blobs to browser-facing code.
  - [x] Add unit tests for metadata normalization.
- [x] Integrate adapter with existing source/cache seams lightly:
  - [x] Preserve local folder source for development.
  - [x] Do not build full setup portal behavior yet unless needed for adapter verification.
  - [x] Keep playback/cache APIs source-agnostic.
- [x] Update docs:
  - [x] Record supported/tested Immich version assumptions.
  - [x] Record any manual verification steps.
  - [x] Update `GOAL.md` as checklist items are completed.
- [x] Commit and push after each coherent checklist feature or checklist item with subitems is completed.

### Phase 2 Verification Notes - 2026-05-21

- Baseline before code changes passed on `master`: `go test ./...`, `pnpm typecheck`, and `pnpm build`.
- Branch was `master`; `origin` was `https://github.com/MonsteRico/immich-frame.git`.
- Official Immich API docs/OpenAPI guidance were reviewed. Endpoint assumptions are recorded in `docs/immich-api.md`.
- No Matthew-instance credentials/context were available in this session, so manual live Immich verification was not run. Adapter behavior is covered with mock HTTP unit tests only, per MVP scope.
- Implemented `internal/immich` client methods for connection testing, album listing, album candidates, random candidates, preview rendition fetching, user-safe error normalization, and metadata normalization.
- Integrated Immich candidates with the cache through a source-agnostic fetch path. Local folder source remains the default development loop.

### Phase 2 PM Readiness Confirmation - 2026-05-21

- Repo is clean on `master`.
- `go test ./...` passed.
- `pnpm typecheck` passed.
- `pnpm build` passed.
- Phase 2 commit history used meaningful slices:
  - adapter boundary.
  - rendition/cache integration.
  - Immich endpoint assumption docs.
- Phase 2 is accepted as complete. No Phase 2 fix phase is required before starting setup portal work.

### Developer Documentation Baseline - 2026-05-21

- Added `docs/developer-guide.md` as the human-oriented workflow and project map.
- Existing command-level runbook remains in `docs/local-development.md`.
- Future agents must update developer-facing docs alongside behavior changes, especially setup flow, configuration, verification, and Immich assumptions.

### Phase 3 Setup Portal Checklist

- [x] Baseline verification before changes:
  - [x] Confirm branch is `master`.
  - [x] Confirm remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
  - [x] Run `go test ./...`.
  - [x] Run `pnpm typecheck`.
  - [x] Run `pnpm build`.
- [x] Define setup state model and backend boundary:
  - [x] Represent unconfigured, setup-code-required, configured, and error/degraded states.
  - [x] Keep setup state compatible with future Wi-Fi AP/captive portal work without implementing that future path.
  - [x] Add unit tests for setup-state transitions.
- [x] Implement first-boot setup code flow:
  - [x] Generate a fixed setup code until setup completes.
  - [x] Persist setup code/state in `state.json`.
  - [x] Expose setup state to the HDMI `/frame` UI and setup portal.
  - [x] Invalidate setup code after setup completes.
  - [x] Add unit tests.
- [x] Implement local admin auth:
  - [x] Create admin password during setup.
  - [x] Store only a password hash in `secrets.json`.
  - [x] Add 30-minute admin sessions that renew on activity.
  - [x] Add login/logout routes.
  - [x] Add unit tests for password/session behavior.
- [x] Implement settings/config API:
  - [x] `GET /api/settings`.
  - [x] `PUT /api/settings`.
  - [x] Read/write `config.toml` as the source of truth for non-secret settings.
  - [x] Preserve the rule that Immich API key is replace-only and never revealed.
  - [x] Add unit tests for config write/validation behavior.
- [x] Implement Immich setup validation routes:
  - [x] `POST /api/immich/test` using the Phase 2 adapter.
  - [x] `GET /api/immich/albums` for authenticated/setup-authorized callers.
  - [x] Return user-safe errors for invalid URL, invalid key, network failure, permission failure, and incompatible response.
  - [x] Add mock HTTP unit tests.
- [x] Implement source/settings selection support:
  - [x] Album mode with one selected album.
  - [x] Random library mode.
  - [x] Searchable album picker data shape with name and item count when available.
  - [x] Slide interval, display fit, cache preset, and overlay settings needed by MVP.
- [x] Implement phone-first setup/settings UI:
  - [x] Setup-code claim screen.
  - [x] Admin password creation.
  - [x] Immich URL/API key entry.
  - [x] Clear warning for HTTP Immich URLs.
  - [x] Connection test feedback.
  - [x] Source selection with searchable album picker.
  - [x] Lightweight status page.
  - [x] Existing setup scaffold should evolve into ongoing settings, not a full admin dashboard.
- [x] Update `/frame` unconfigured behavior:
  - [x] Show polished HDMI setup screen with `frame.local:8787/setup`, IP fallback if available, setup code, and status.
  - [x] Do not require keyboard/touch on the frame.
  - [x] Preserve the existing local slideshow behavior when configured or when using local dev source.
- [x] Security checks:
  - [x] Browser never receives Immich API key.
  - [x] Settings UI never reveals saved Immich API key.
  - [x] LAN clients cannot access cached media/settings without proper setup/admin authorization.
  - [x] Localhost kiosk access remains appliance-friendly.
- [x] Documentation updates:
  - [x] Update `docs/configuration.md` for any config/state/secrets shape changes.
  - [x] Update `docs/security.md` for setup/auth/session behavior.
  - [x] Update `docs/local-development.md` with setup portal run/verification steps.
  - [x] Update `docs/developer-guide.md` if daily workflow changes.
  - [x] Update `GOAL.md` as checklist items are completed.
- [x] Commit and push after each coherent checklist feature or checklist item with subitems is completed.

### Phase 3 PM Review Notes - 2026-05-21

- Repo is clean on `master` at commit `3e0b0c2 Build setup portal`.
- `go test ./...` passed.
- `pnpm typecheck` passed.
- `pnpm build` passed.
- Phase 3 added setup state, first-boot code flow, admin password/session auth, settings read/write, Immich validation routes, setup/settings UI, and unconfigured frame setup screen.
- Phase 3 is not accepted as fully ready for Phase 4 yet. A focused Phase 3.5 hardening pass is needed before Pi appliance installer work.

### Phase 3.5 Setup Portal Hardening Checklist

- [x] Baseline verification before changes:
  - [x] Confirm branch is `master`.
  - [x] Confirm remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
  - [x] Run `go test ./...`.
  - [x] Run `pnpm typecheck`.
  - [x] Run `pnpm build`.
- [x] Require successful Immich validation before setup completion:
  - [x] Track whether the saved Immich URL/API key have passed validation.
  - [x] Prevent finishing setup when the user skipped validation or changed URL/key after validation.
  - [x] Preserve the saved-key replace-only rule.
  - [x] Add backend unit tests for validation-required setup completion.
  - [x] Add UI feedback that clearly tells the user why setup cannot finish yet.
- [x] Add a real lightweight status surface:
  - [x] Implement `GET /api/status` or document and intentionally remove it from planned API docs.
  - [x] Surface setup/configuration status, Immich connection status when known, source mode, cache count, and last error without leaking secrets.
  - [x] Show this status in the settings portal after setup.
  - [x] Add unit tests for the status response.
- [x] Tighten setup/settings UI behavior:
  - [x] Disable or guard source/finish actions until required fields and validation state are ready.
  - [x] Make random-library mode as validation-dependent as album mode.
  - [x] Keep phone-first layout and avoid turning the portal into a broad admin dashboard.
- [x] Reconcile overlay configuration docs with implementation:
  - [x] Either remove unimplemented overlay-specific TOML fields from current config docs, or implement preservation for raw overlay-specific fields.
  - [x] Keep MVP settings focused on implemented overlay controls.
- [x] Update stale project status docs:
  - [x] Update `README.md`.
  - [x] Update `AGENT_BRIEF.md`.
  - [x] Update `docs/implementation-plan.md` with Phase 3 done state and Phase 3.5 current work.
  - [x] Update `docs/configuration.md`, `docs/security.md`, `docs/local-development.md`, and `docs/developer-guide.md` for any behavior changes.
  - [x] Update `GOAL.md` with Phase 3.5 verification notes.
- [x] Commit and push after each coherent checklist feature or feature plus subitems is complete.

### Phase 3.5 Verification Notes - 2026-05-21

- Baseline before code changes passed on `master`: `go test ./...`, `pnpm typecheck`, and `pnpm build`.
- `origin` was `https://github.com/MonsteRico/immich-frame.git`.
- `state.json` now records Immich validation metadata for the saved URL/API key fingerprint without storing the raw key.
- `POST /api/setup/complete` now requires the admin password, saved Immich URL/key, validation matching those saved credentials, and a valid album or random-library source.
- Random-library setup can no longer bypass Immich validation.
- `GET /api/status` and the `GET /api/settings` response expose setup status, Immich validation status, source mode, cache count, and last error without raw secrets.
- Setup/settings UI now disables or guards Save/Finish actions until required fields and validation state are ready.
- Overlay docs now match the backend's implemented generic envelope fields: `enabled`, `slot`, and `visibility`.
- Final verification passed: `go test ./...`, `pnpm typecheck`, `pnpm build`, and `pnpm build:embedded-ui`.

### Phase 4 Appliance Work Reverted - 2026-05-23

- Raspberry Pi/Chromium kiosk work was intentionally reverted after hardware exploration showed likely performance risk on the Pi Zero 2 W.
- Hardware install scripts, packaging assets, and Pi appliance docs are not part of the current renderer spike.
- Future hardware work should start after Phase 6 chooses a lighter renderer rather than assuming Chromium kiosk is the final target.

### Phase 5 Browser MVP Polish And Hardening Checklist

- [x] Baseline verification before changes:
  - [x] Confirm branch is `master`.
  - [x] Confirm remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
  - [x] Run `go test ./...`.
  - [x] Run `pnpm typecheck`.
  - [x] Run `pnpm build`.
- [x] Implement cache rotation and eviction:
  - [x] Periodically refresh Immich candidates for album and random-library sources.
  - [x] Top off cached display-targeted renditions toward `cache.target_items`.
  - [x] Maintain a near-term prefetch buffer using `cache.prefetch_items`.
  - [x] Prefer never-shown and least-recently-shown candidates.
  - [x] Evict assets that left the selected source before evicting valid offline fallback photos.
  - [x] Avoid evicting current and near-upcoming playback entries.
  - [x] Refresh playback queue when cache contents change so the frame does not loop one static seed forever.
  - [x] Add focused unit tests.
- [x] Implement outage retry/backoff:
  - [x] Continue slideshow from cache when Immich/network is unavailable.
  - [x] Retry Immich refresh with bounded backoff.
  - [x] Preserve useful last-error details for status surfaces without noisy bright failures.
  - [x] Add focused unit tests.
- [x] Tighten degraded/offline UI states:
  - [x] Show operational status overlay only when degraded/error conditions exist.
  - [x] Keep the current photo visible when the next fetch fails.
  - [x] Show calm empty-cache plus unavailable-Immich state when no cached media can play.
- [x] Finish reset/status CLI behavior:
  - [x] Ensure `immich-frame status` reports setup/config/source/cache/last-error details without secrets.
  - [x] Ensure `reset` behavior is documented and privacy-preserving.
  - [x] Ensure `config validate` covers the settings needed by the browser MVP.
- [x] Re-verify browser MVP manually:
  - [x] Local mock source path still works.
  - [x] Setup portal flow still works with mocked/unit-tested Immich behavior.
  - [x] `/frame` slideshow, overlays, controls, and SSE state remain stable.
  - [x] Embedded UI serving still works after `pnpm build:embedded-ui`.
- [x] Update docs:
  - [x] Update `README.md`.
  - [x] Update `AGENT_BRIEF.md`.
  - [x] Update `docs/implementation-plan.md`.
  - [x] Update `docs/architecture.md`, `docs/configuration.md`, `docs/security.md`, and `docs/local-development.md` for changed behavior.
  - [x] Update `docs/future.md` with renderer/hardware follow-up notes.
  - [x] Update `GOAL.md` with Phase 5 verification notes.
- [x] Commit and push after each coherent checklist feature or feature plus subitems is complete.

### Phase 5 Cache Rotation And Outage Slice Notes - 2026-05-24

- Baseline before code changes passed on `master`: `go test ./...`, `pnpm typecheck`, and `pnpm build`.
- `origin` was `https://github.com/MonsteRico/immich-frame.git`.
- Added daemon background cache maintenance for configured Immich album and random-library sources.
- Immich candidates now refresh periodically using `sync.refresh_interval_minutes`; failures retry with bounded backoff while cached playback continues.
- The cache tops off display-targeted renditions toward `cache.target_items`.
- Cache eviction prefers assets no longer present in the selected source before removing valid fallback photos.
- Source-removed cache entries are pruned after a successful candidate refresh, while protected current and near-upcoming entries survive.
- Current and near-upcoming playback entries are protected according to `cache.prefetch_items`.
- Playback queues refresh when cache contents change and preserve the current photo when possible.
- Runtime refresh failures set calm degraded/error queue messages and persist a user-safe last error for status surfaces.
- Focused unit tests cover top-off, source-aware eviction/pruning, protected playback entries, bounded backoff, degraded cache-first status, and queue refresh preserving the current photo.
- Verification after this slice passed: `go test ./...`, `pnpm typecheck`, and `pnpm build`.

### Phase 5 Degraded Frame UI Slice Notes - 2026-05-24

- The frame UI now shows backend operational status text only for degraded/error states, plus local fetch/command errors.
- Configured frames with no playable cached media and an Immich outage now show a calm offline screen with the setup URL instead of local-development photo instructions.
- The current photo remains the primary view during degraded cached playback; outage status appears as the subtle operational overlay.
- Verification after code changes passed: `pnpm typecheck` and `pnpm build`.
- Browser rendering verification passed in the Codex in-app Browser against `http://127.0.0.1:8787/frame` for configured empty-cache outage state and local mock slideshow state. Evidence showed the offline screen without duplicate status overlay, no console warnings/errors, the mock slideshow with photo overlays, and ArrowRight advancing playback from one cached local photo to the next.

### Phase 5 CLI Hardening Slice Notes - 2026-05-24

- `immich-frame status` now accepts `-config`, reports setup/config/source/cache/Immich validation/last-error details, and reports whether an API key is configured without printing the key.
- `immich-frame reset` keeps clearing secrets and state by default, clears cached media unless `--keep-cache` is set, and can remove a config file when `--config` is explicitly provided.
- `config validate` now checks browser MVP source requirements, Immich URL requirements for album/random modes, cache values, sync refresh interval, and overlay slots/visibility.
- Focused unit tests cover secret-safe status output, reset removal behavior, and the expanded config validation surface.

### Phase 5 Browser MVP Verification Notes - 2026-05-24

- `pnpm build:embedded-ui` passed and synced the updated frame UI assets into `internal/api/static/frame`.
- Embedded UI smoke verification ran the daemon with missing external dist paths: `go run ./cmd/immich-frame serve -config config.dev.toml -dev-source dev/photos -data-dir .immich-frame-embedded-verify -frame-dist missing-frame-dist -setup-dist missing-setup-dist`.
- Browser verification passed in the Codex in-app Browser for `http://127.0.0.1:8787/frame`: title `Immich Frame`, visible local mock photo, clock/photo-info overlays, no status overlay while ready, no console warnings/errors, and embedded frame asset URLs `/assets/index-Djr5t2pk.js` and `/assets/index-2C9x_Gpy.css`.
- HTTP smoke checks returned `200` for `/api/state`, the embedded frame JS asset, and the embedded frame CSS asset.
- Browser verification passed for `http://127.0.0.1:8787/setup`: title `Immich Frame Setup`, setup code claim screen rendered from embedded setup assets, and no console warnings/errors. Setup/auth/settings/Immich validation behavior remains covered by mock HTTP unit tests; no real-Immich integration tests were added.
- CLI smoke checks passed: `immich-frame version`, `config validate -config config.dev.toml`, and `status -config config.dev.toml -data-dir .immich-frame`.

### Phase 5.5 Browser MVP Acceptance Fix Pass Checklist

Phase 5.5 acceptance notes:

- Full album caches now churn one safe slot per refresh for stable album sources larger than the cache target, using manifest history so never-cached candidates are introduced before recently evicted candidates are re-cached.
- `cache.prefetch_items` continues to protect the current and near-upcoming playback entries during rotation/eviction.
- `cache.preset = "extra-small"` is available for local testing at roughly 10 cached photos; `balanced` remains the default.
- Cache maintenance publishes a recovered `ready` SSE state after a successful outage retry, even when cache contents did not change.
- Phase 6 renderer/hardware work remains future-only; installer/systemd/kiosk work is still paused.

Post-Phase 5.5 requested and approved cache strategy change:

- The frame should progress through the broader selected album/library over time instead of looping the same warm cache and adding only one new image per timer refresh.
- Album/random-library sources now use playback-driven rolling refresh: after `cache.refresh_after_shown_items` photos are shown, cache maintenance is requested immediately.
- A rolling refresh can replace up to `cache.refresh_batch_items` shown, unprotected cache entries with new candidates while preserving current and near-upcoming playback entries.
- Completing first setup with an Immich source requests immediate cache maintenance so an empty cache can begin fetching images and start playback without restarting the server.

Post-Phase 5.5 logging addition:

- `immich-frame serve --logs` enables opt-in operational logs for cache refresh and playback activity.
- Logs should stay count/status focused and avoid Immich API keys, direct Immich URLs, filenames, titles, raw response bodies, and filesystem paths unless a future debug mode explicitly reconsiders that privacy tradeoff.

- [x] Baseline verification before changes:
  - [x] Confirm branch is `master`.
  - [x] Confirm remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
  - [x] Run `go test ./...`.
  - [x] Run `pnpm typecheck`.
  - [x] Run `pnpm build`.
- [x] Fix stable album cache rotation:
  - [x] Rotate/churn album-mode cache even when cache is already at `cache.target_items`.
  - [x] Prefer never-cached or least-recently-shown album candidates from sources larger than the cache target.
  - [x] Preserve current and near-upcoming playback entries using `cache.prefetch_items`.
  - [x] Keep the rotation deterministic enough for unit tests.
  - [x] Add focused unit tests proving a full cache can rotate in new album candidates.
- [x] Add an extra-small cache preset for local testing:
  - [x] Add a preset intended for roughly 10 cached photos.
  - [x] Keep production defaults unchanged.
  - [x] Document that the preset exists to make cache rotation visible during development.
- [x] Fix ready-status publication after outage recovery:
  - [x] Publish recovered ready state when refresh succeeds after degraded/error status, even if cache contents do not change.
  - [x] Add or update a focused unit test where practical.
- [x] Align docs and status:
  - [x] Update `AGENT_BRIEF.md` to describe Phase 5.5 as current until complete.
  - [x] Update `README.md` current status.
  - [x] Update `docs/implementation-plan.md` with Phase 5.5 current work and Phase 6 remaining future work.
  - [x] Update `GOAL.md` with Phase 5.5 acceptance notes.
  - [x] Keep Phase 6 renderer/hardware re-evaluation as future work only.
- [x] Final verification:
  - [x] Run `go test ./...`.
  - [x] Run `pnpm typecheck`.
  - [x] Run `pnpm build`.
  - [x] Run `pnpm build:embedded-ui` because setup UI assets changed.
- [x] Commit and push coherent fix slices to `master`.

### Phase 5.5 / Browser Reference Verification Notes - 2026-05-24

- Matthew tested against a personal Immich instance with a real API key.
- Random mode continues rotating.
- The `extra-small` cache preset makes cache rotations visible.
- Restart/reboot simulation works: the frame resumes without repeating setup.
- Embedded release-style serving works.
- Longer soak testing is intentionally deferred for now.
- LAN security from another device is not yet verified and should remain a future verification item.
- Network outage behavior exposed a renderer limitation: the daemon appears to continue cache-first playback and reconnects correctly for the next Immich request, but the Chromium tab does not keep the visual slideshow moving and does not resume cleanly after reconnect.
- Because the browser renderer is likely not the final Pi Zero 2 W appliance runtime, this Chromium outage gap should be documented as a reference-renderer limitation rather than becoming the next deep browser-hardening phase.

### Phase 6 Renderer Replacement Spike Checklist

Goal: choose and prototype the appliance renderer direction while preserving the existing daemon, setup portal, Immich adapter, cache, playback, settings, and docs foundation.

- [ ] Baseline verification before changes:
  - [x] Confirm branch is `master`.
  - [x] Confirm remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
  - [x] Review the current `immich-config.dev.toml` diff before touching config files; Matthew may have local testing settings there.
  - [x] Run `go test ./...`.
  - [x] Run `pnpm typecheck`.
  - [x] Run `pnpm build`.
- [x] Define the renderer contract:
  - [x] Document the minimal daemon APIs/state a renderer needs: current asset, next asset if useful, status/message, overlay config, media path, playback interval, and any display target information.
  - [x] Decide whether the renderer should consume `/api/state`, `/media/:assetID`, direct local cache file paths exposed by a renderer-only API, or another local-only contract.
  - [x] Prefer a pull/polling or resilient hybrid model over SSE-only behavior for the appliance renderer.
  - [x] Preserve the setup portal as browser-based unless there is a concrete reason to change it.
- [x] Evaluate renderer options for Pi Zero 2 W-class hardware:
  - [x] Compare at least three options such as SDL2/Go, SDL2/Rust, framebuffer/image viewer process, Qt/QML, lightweight WebKit, or another credible local renderer.
  - [x] Score options on memory/CPU footprint, image scaling quality, crossfade/overlay feasibility, packaging complexity, Go integration, testability without hardware, and Pi availability.
  - [x] Recommend one primary path and one fallback path.
  - [x] Record rejected options and why.
- [x] Build a narrow proof of concept for the recommended renderer:
  - [x] Keep the proof of concept behind a clearly named command, package, or `experiments/` path.
  - [x] Render a display-sized cached/local image from daemon-compatible state or a documented fixture.
  - [x] Render at least one simple overlay, such as clock or status text, enough to prove overlays are feasible.
  - [x] Use a resilient update loop that keeps the current image visible when state/media refresh fails.
  - [x] Avoid restarting installer/systemd/kiosk work in this phase.
- [x] Add testable seams:
  - [x] Unit-test renderer contract parsing/state adaptation where practical.
  - [x] Add fixture-driven tests for outage/reconnect state handling if a full renderer cannot run in CI.
  - [x] Keep tests unit/mock only; no real-Immich integration tests in repo/CI.
- [ ] Update docs:
  - [ ] Update `docs/architecture.md` with browser-as-reference and new renderer boundary.
  - [ ] Update `docs/implementation-plan.md` with the Phase 6 decision/prototype outcome.
  - [ ] Update `docs/future.md` and hardware notes with what remains after the spike.
  - [ ] Update `AGENT_BRIEF.md` and `GOAL.md` as work progresses.
- [ ] Final verification:
  - [ ] Run `go test ./...`.
  - [ ] Run `pnpm typecheck`.
  - [ ] Run `pnpm build`.
  - [ ] Run any renderer-specific unit/build checks introduced by the phase.
- [ ] Commit and push coherent slices to `master`.

### Phase 6 Contract And Option Notes - 2026-05-24

- Baseline passed on `master`: `go test ./...`, `pnpm typecheck`, and `pnpm build`.
- `origin` is `https://github.com/MonsteRico/immich-frame.git`.
- Existing local `immich-config.dev.toml` changes are Matthew's test settings: `extra-small` cache, smaller cache sizes, lower prefetch, and one-minute sync refresh. They were reviewed and not edited.
- The appliance renderer contract is documented as a daemon-owned local snapshot/presentation boundary.
- The current browser `/frame` renderer remains the reference/development renderer using `/api/state`, `/api/events`, and `/media/:assetID`.
- The appliance renderer should use a local-only snapshot endpoint, tentatively `GET /api/renderer/state`, with optional event wake-ups rather than SSE-only behavior.
- The renderer must keep the last successfully decoded image visible when state or media refresh fails.
- Evaluated Go + SDL2, Rust + SDL2, framebuffer/image-viewer process, Qt/QML, and WPE/WebKit.
- Recommended primary path: Go + SDL2 native renderer.
- Recommended fallback path: pre-composited framebuffer/image-viewer renderer for very weak hardware or SDL packaging failures.
- Proof-of-concept work is intentionally paused until the path is discussed.

### Phase 6 Proof-Of-Concept Notes - 2026-05-24

- Added `internal/renderer` with the local renderer snapshot contract, frame-retention loop, and PNG preview renderer.
- Added local-only `GET /api/renderer/state`, which adapts daemon playback/config/cache state for the appliance renderer and rejects non-loopback callers.
- The renderer snapshot can include local cache file paths only across the loopback renderer boundary; `/api/state` remains the browser reference contract.
- Added `immich-frame renderer-poc`, a Windows-friendly fixture/prototype command that renders a cached/local image through the renderer contract into a PNG preview with a status/clock overlay.
- Added tests for renderer state adaptation, local-only renderer API behavior, preview generation, and keeping the previously decoded image visible when snapshot fetch or media decode fails.
- No installer, systemd, autostart, kiosk, or OS-image work was restarted.
- The remaining hardware-facing work is to put an SDL display shell around the tested renderer contract/loop and then verify SDL packaging/runtime behavior on target hardware.

## Stop Conditions

An agent can stop when:

- [ ] The requested phase checklist is complete.
- [ ] Tests/builds for touched areas pass, or failures are clearly documented.
- [ ] Docs are updated for any changed behavior or decision.
- [ ] Work has been committed in meaningful increments.
- [ ] No required dev servers, long-running commands, or install sessions are left running.
- [ ] Remaining work is captured in a clear checklist.

## Commit Expectations

Keep git history useful and reviewable.

Until the MVP/base is complete, agents should work directly on `master` and commit there. Do not create feature branches for ordinary implementation slices unless the user explicitly asks for one.

Commit and push at coherent feature or fix boundaries rather than only at phase boundaries. A good history should explain what was built in practical slices, such as:

- project scaffold and build tooling.
- CLI skeleton.
- config/secrets/state models.
- frontend frame scaffold.
- setup/settings scaffold.
- embedded UI assets.
- playback queue.
- cache manifest.
- SSE/API routes.
- mock source support.
- targeted bug fixes.
- documentation updates for changed behavior.

Do not commit after every tiny edit. Do not wait until all of Phase 0 or Phase 1 is complete if several distinct working pieces can be committed separately.

For Phase 6, generally commit and push after each coherent decision, prototype, or documentation slice. Examples: renderer contract docs, renderer option evaluation, proof-of-concept scaffold, outage-resilient renderer state loop, and final phase handoff docs.

Before each commit:

- [ ] Review `git status`.
- [ ] Review the diff for the intended scope.
- [ ] Run the relevant tests/build checks when practical.
- [ ] Use a specific commit message describing the completed slice.
- [ ] Push `master` after the commit unless the user has asked you to hold local commits.

## Out Of Scope For MVP

- [ ] Temporary setup Wi-Fi access point.
- [ ] Captive portal.
- [ ] Weather provider integration.
- [ ] Favorites source mode.
- [ ] On-this-day source mode.
- [ ] People/location/date smart rules.
- [ ] Video playback.
- [ ] GPIO buttons, IR remotes, or Bluetooth remote integration.
- [ ] Raspberry Pi install/systemd/kiosk setup.
- [ ] Flashable custom OS image.
- [ ] Automatic updates.
- [ ] Docker/LAN deployment mode.
- [ ] Runtime third-party overlay plugins.
- [ ] Full log dashboard or cache management dashboard.
