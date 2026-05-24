# Immich Frame Goal

This document defines when an agent can stop working and what counts as done.

## Long-Term Product Goal

Immich Frame is complete enough for the browser MVP when it can run as a reliable local browser-based digital picture frame connected to Immich.

The browser MVP should show a polished setup screen if unconfigured, accept setup from a same-machine or same-LAN browser session, cache display-appropriate Immich photos, rotate cached photos without looping a static seed forever, and run a fullscreen slideshow with subtle overlays.

Hardware/appliance setup is intentionally paused until the browser MVP behavior is solid. The Raspberry Pi Zero 2 W Chromium kiosk experiment showed enough performance risk that the next hardware phase should first evaluate a lighter rendering engine and reuse the browser MVP daemon/cache/setup behavior.

## MVP Definition Of Done

The browser MVP is done when all items below are complete in local/browser development:

- [ ] Unconfigured HDMI screen shows setup instructions, `frame.local:8787`, IP fallback, and first-boot setup code.
- [ ] Setup portal accepts setup code.
- [ ] User can create local admin password.
- [ ] Setup portal accepts Immich URL and dedicated Immich API key.
- [ ] Daemon validates Immich connection.
- [ ] User can choose one album or random library mode.
- [ ] Daemon caches first display-targeted photo renditions locally.
- [ ] Cache rotation refreshes candidates, tops off target cache, and avoids cycling one static cache forever.
- [ ] Slideshow starts as soon as first few cached images are ready.
- [ ] Hidden controls work: previous, pause/play, next, info toggle.
- [ ] Clock overlay works.
- [ ] Photo info overlay works with minimal metadata.
- [ ] Operational status overlay appears only for degraded/error states.
- [ ] Reboot resumes slideshow without manual intervention.
- [ ] Immich/network outage continues from cache when possible.
- [ ] Empty cache plus unavailable Immich shows a calm retry/error state.
- [ ] Settings portal remains available after setup and requires admin auth.
- [ ] `immich-frame status`, `reset`, `config validate`, and `version` work.
- [ ] Unit tests pass.
- [ ] Frontend typecheck/build passes.
- [ ] Docs match implemented behavior.

## Current Session Goal

Complete **Phase 5: Browser MVP Polish And Hardening**.

Phase 3.5 is complete. Phase 4 appliance installer work was reverted and hardware setup is paused. The next agent should finish the browser-based MVP behavior first, especially cache rotation, outage behavior, status overlays, reset/status details, and docs.

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
- Hardware install scripts, packaging assets, and Pi appliance docs are not part of the current browser MVP path.
- Future hardware work should start after Phase 5 and should evaluate a lighter renderer rather than assuming Chromium kiosk is the final target.

### Phase 5 Browser MVP Polish And Hardening Checklist

- [ ] Baseline verification before changes:
  - [ ] Confirm branch is `master`.
  - [ ] Confirm remote is `origin` at `https://github.com/MonsteRico/immich-frame.git`.
  - [ ] Run `go test ./...`.
  - [ ] Run `pnpm typecheck`.
  - [ ] Run `pnpm build`.
- [ ] Implement cache rotation and eviction:
  - [ ] Periodically refresh Immich candidates for album and random-library sources.
  - [ ] Top off cached display-targeted renditions toward `cache.target_items`.
  - [ ] Maintain a near-term prefetch buffer using `cache.prefetch_items`.
  - [ ] Prefer never-shown and least-recently-shown candidates.
  - [ ] Evict assets that left the selected source before evicting valid offline fallback photos.
  - [ ] Avoid evicting current and near-upcoming playback entries.
  - [ ] Refresh playback queue when cache contents change so the frame does not loop one static seed forever.
  - [ ] Add focused unit tests.
- [ ] Implement outage retry/backoff:
  - [ ] Continue slideshow from cache when Immich/network is unavailable.
  - [ ] Retry Immich refresh with bounded backoff.
  - [ ] Preserve useful last-error details for status surfaces without noisy bright failures.
  - [ ] Add focused unit tests.
- [ ] Tighten degraded/offline UI states:
  - [ ] Show operational status overlay only when degraded/error conditions exist.
  - [ ] Keep the current photo visible when the next fetch fails.
  - [ ] Show calm empty-cache plus unavailable-Immich state when no cached media can play.
- [ ] Finish reset/status CLI behavior:
  - [ ] Ensure `immich-frame status` reports setup/config/source/cache/last-error details without secrets.
  - [ ] Ensure `reset` behavior is documented and privacy-preserving.
  - [ ] Ensure `config validate` covers the settings needed by the browser MVP.
- [ ] Re-verify browser MVP manually:
  - [ ] Local mock source path still works.
  - [ ] Setup portal flow still works with mocked/unit-tested Immich behavior.
  - [ ] `/frame` slideshow, overlays, controls, and SSE state remain stable.
  - [ ] Embedded UI serving still works after `pnpm build:embedded-ui`.
- [ ] Update docs:
  - [ ] Update `README.md`.
  - [ ] Update `AGENT_BRIEF.md`.
  - [ ] Update `docs/implementation-plan.md`.
  - [ ] Update `docs/architecture.md`, `docs/configuration.md`, `docs/security.md`, and `docs/local-development.md` for changed behavior.
  - [ ] Update `docs/future.md` with renderer/hardware follow-up notes.
  - [ ] Update `GOAL.md` with Phase 5 verification notes.
- [ ] Commit and push after each coherent checklist feature or feature plus subitems is complete.

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

For Phase 5 specifically, generally commit and push after each completed checklist feature or checklist item with its subitems. Examples: cache rotation, eviction policy, outage retry/backoff, degraded UI states, CLI status/reset hardening, browser verification, and docs updates.

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
- [ ] Native framebuffer/SDL renderer or other lightweight renderer.
- [ ] Raspberry Pi install/systemd/kiosk setup.
- [ ] Flashable custom OS image.
- [ ] Automatic updates.
- [ ] Docker/LAN deployment mode.
- [ ] Runtime third-party overlay plugins.
- [ ] Full log dashboard or cache management dashboard.
