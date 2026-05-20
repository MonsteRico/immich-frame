# Development

## Goal

Development should be possible on a normal desktop without a Raspberry Pi and without a real Immich server.

The reference runtime is still Raspberry Pi Zero 2 W, but contributors should be able to build overlays, frame UI, setup UI, playback behavior, and cache logic locally.

## Tooling

- Go for daemon/core.
- Preact + Vite for frontend bundles.
- pnpm for frontend tooling.
- Plain CSS/CSS modules.
- Minimal inline SVG icons.
- No Tailwind.
- No component library.
- No icon library.

## Local Development Modes

Supported development modes:

```text
desktop mock mode
  daemon runs locally
  frame opens in regular browser
  source uses local folder photos

desktop Immich mode
  daemon runs locally
  frame opens in regular browser
  source uses a real Immich URL/API key from local dev config
```

Mock/local folder source is a dev-only tool, not a user-facing MVP mode.

Example config shape:

```toml
[source]
mode = "local_folder"

[source.local_folder]
path = "./dev/photos"
shuffle = true
```

## Expected Local URLs

```text
http://localhost:8787/frame
http://localhost:8787/setup
```

Kiosk hardware later uses:

```text
http://127.0.0.1:8787/frame
```

## Frontend Development

Use separate frame and setup bundles:

```text
ui/frame
ui/setup
ui/shared
```

Frame bundle should stay lean:

- slideshow.
- overlays.
- hidden controls.
- display-size reporting.

Setup bundle can include forms:

- setup code claim.
- Immich URL/API key entry.
- album search picker.
- settings editor.
- overlay config editor.
- status page.

## Testing Policy

MVP tests are unit tests only.

Do not add built-in real-Immich integration tests to repo/CI for MVP. Real Immich verification can be done manually during development through the setup portal and local status commands.

Default CI scope:

```text
go test ./...
frontend typecheck
frontend build
```

Adapter tests should use mock HTTP servers/fixtures.

Useful test areas:

- config parsing/writing.
- secrets permissions behavior where practical.
- auth/session behavior.
- setup code lifecycle.
- Immich request construction and response normalization with mocks.
- cache manifest and eviction.
- playback queue/history/repeat avoidance.
- overlay schema/default behavior.

## Release Build Expectation

Release builds embed built UI assets into the Go binary.

Development may optionally serve Vite bundles or external UI assets, but appliance releases should not depend on separate UI files.

