# Developer Guide

This guide is for humans working on Immich Frame without relying on an AI agent to remember project context.

For command-by-command local run and browser verification steps, also read `docs/local-development.md`.

## Toolchain

Use the active PowerShell environment on Matthew's machine.

Expected tools:

```powershell
go version
node -v
pnpm -v
```

Current expected shape:

- Go 1.22 or newer.
- Node from the PowerShell/fnm environment.
- pnpm from the PowerShell environment.

The repo intentionally does not pin pnpm in `package.json`. If `node` or `pnpm` resolve to an unexpected version, fix the shell environment instead of changing project scripts.

## First-Time Setup

From the repo root:

```powershell
pnpm install
go test ./...
pnpm typecheck
pnpm build
```

pnpm 11 uses build-script approvals. The repo approves `esbuild` in `pnpm-workspace.yaml` because Vite needs it.

## Daily Development Loop

Before changing code:

```powershell
git status --short
go test ./...
pnpm typecheck
pnpm build
```

For local frame work:

```powershell
go run ./cmd/immich-frame serve -config config.dev.toml -dev-source dev/photos
```

Then open:

```text
http://127.0.0.1:8787/frame
http://127.0.0.1:8787/setup
```

## Release-Style Embedded UI Check

Before testing the embedded UI path:

```powershell
pnpm build:embedded-ui
```

Then run the daemon with missing external dist paths so the embedded assets are forced:

```powershell
go run ./cmd/immich-frame serve -config config.dev.toml -dev-source dev/photos -data-dir .immich-frame-verify -frame-dist missing-frame-dist -setup-dist missing-setup-dist
```

Verify:

- `/frame` loads.
- `/setup` loads.
- `/assets/index-*.js` returns `200`.
- `/assets/index-*.css` returns `200`.
- `/api/state` returns a ready local-folder state.
- `/media/:assetID` returns local cached media.

## Project Map

```text
cmd/immich-frame
  CLI entry point.

internal/api
  HTTP routes, SSE, static UI serving, media route.

internal/app
  Application wiring and source/cache/playback startup.

internal/cache
  Cache manifest and local/fetched media storage.

internal/config
  Config, secrets, state models and file IO.

internal/immich
  Immich REST API adapter. Keep Immich endpoint details here.

internal/playback
  Slideshow queue and playback state.

internal/source
  Source-neutral candidates, including local folder dev source.

ui/frame
  Kiosk slideshow UI.

ui/setup
  Phone-first setup/settings UI.

ui/shared
  Shared frontend API/types.
```

## Immich Development Notes

The repo should not include real Immich credentials or real-Immich integration tests.

Use mock HTTP tests for repo/CI coverage. Manual Immich checks are allowed during development when credentials are available, but record only endpoint behavior and version notes, not secrets.

Relevant docs:

- `docs/immich-api.md` for endpoint assumptions.
- `docs/security.md` for credential handling.
- `docs/configuration.md` for config/secrets/state layout.

## Documentation Expectations

Documentation is part of each feature, not a cleanup chore at the end.

When changing behavior, update the relevant docs in the same feature slice:

- setup or run commands: update `docs/local-development.md` and/or this guide.
- configuration shape: update `docs/configuration.md`.
- Immich endpoints or assumptions: update `docs/immich-api.md`.
- security/auth/media exposure behavior: update `docs/security.md`.
- phase status or handoff state: update `GOAL.md` and `AGENT_BRIEF.md`.
- future decisions or intentionally deferred work: update `docs/future.md`.

Agents should commit docs with the feature they describe whenever practical.

## Git Workflow During MVP/Base Work

Until the MVP/base is complete:

- work directly on `master`.
- commit coherent feature/fix slices.
- push `master` after each completed checklist feature unless told otherwise.

Do not make one broad phase commit if several independently useful pieces are complete.

