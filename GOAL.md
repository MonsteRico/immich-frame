# Immich Frame Goal

This document defines when an agent can stop working and what counts as done.

## Long-Term Product Goal

Immich Frame is complete enough for an MVP when a Raspberry Pi Zero 2 W can act as a reliable self-built HDMI digital picture frame connected to Immich.

The frame should boot into a local Chromium kiosk, show a polished setup screen if unconfigured, accept setup from a phone/laptop on the same Wi-Fi, cache display-appropriate Immich photos, and run a fullscreen slideshow with subtle overlays.

## MVP Definition Of Done

The MVP is done when all items below are complete on the reference Pi Zero 2 W hardware:

- [ ] Raspberry Pi OS Lite is flashed and Wi-Fi is configured through Raspberry Pi Imager.
- [ ] `install.sh` installs Immich Frame idempotently.
- [ ] `immich-frame serve` runs as a systemd service.
- [ ] Chromium kiosk starts automatically on boot.
- [ ] Kiosk opens `http://127.0.0.1:8787/frame`.
- [ ] Unconfigured HDMI screen shows setup instructions, `frame.local:8787`, IP fallback, and first-boot setup code.
- [ ] Setup portal accepts setup code.
- [ ] User can create local admin password.
- [ ] Setup portal accepts Immich URL and dedicated Immich API key.
- [ ] Daemon validates Immich connection.
- [ ] User can choose one album or random library mode.
- [ ] Daemon caches first display-targeted photo renditions locally.
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

For the first implementation session, stop when **Phase 0** is complete and enough of **Phase 1** exists to run a local mock slideshow in a desktop browser.

### Phase 0 Done Checklist

- [ ] Repo structure exists.
- [ ] Go module initialized.
- [ ] CLI skeleton compiles.
- [ ] Config/secrets/state file paths and data types exist.
- [ ] Frame UI bundle scaffold exists.
- [ ] Setup UI bundle scaffold exists.
- [ ] Shared UI package exists.
- [ ] Release embedding strategy is represented in code/build files.
- [ ] Basic CI commands are documented or configured.
- [ ] CI scope is Go unit tests plus frontend typecheck/build.

### Phase 1 Partial Done Checklist

- [ ] Local folder/mock source can produce photo candidates.
- [ ] Cache manifest can track cached local media.
- [ ] Playback queue can advance current/next/previous in memory.
- [ ] `/api/state` returns frame state.
- [ ] `/api/events` streams state updates with SSE.
- [ ] `/media/:assetID` serves local cached media.
- [ ] `/frame` renders slideshow in desktop browser.
- [ ] Clock/photo-info/status overlay placeholders render.

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

Commit at coherent feature or fix boundaries rather than only at phase boundaries. A good history should explain what was built in practical slices, such as:

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

Before each commit:

- [ ] Review `git status`.
- [ ] Review the diff for the intended scope.
- [ ] Run the relevant tests/build checks when practical.
- [ ] Use a specific commit message describing the completed slice.

## Out Of Scope For MVP

- [ ] Temporary setup Wi-Fi access point.
- [ ] Captive portal.
- [ ] Weather provider integration.
- [ ] Favorites source mode.
- [ ] On-this-day source mode.
- [ ] People/location/date smart rules.
- [ ] Video playback.
- [ ] GPIO buttons, IR remotes, or Bluetooth remote integration.
- [ ] Native framebuffer/SDL renderer.
- [ ] Flashable custom OS image.
- [ ] Automatic updates.
- [ ] Docker/LAN deployment mode.
- [ ] Runtime third-party overlay plugins.
- [ ] Full log dashboard or cache management dashboard.
