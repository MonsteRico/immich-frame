# Immich API Adapter Notes

Phase 2 implements the Immich API boundary inside `internal/immich`. Browser-facing routes do not receive Immich API keys, raw Immich JSON, direct Immich URLs, EXIF blobs, GPS coordinates, people data, or original file paths.

## Reviewed API Surface

Reviewed on 2026-05-21 against official Immich API documentation and OpenAPI guidance:

- Immich publishes API docs from its OpenAPI-generated contract. The generated contract is maintained from `immich-openapi-specs.json`.
- The currently published `immich.app/docs/api` page observed during this work identified API version `1.136.0`.
- The newer official API documentation site at `api.immich.app` documents the REST API surface and confirms the current thumbnail path shape used by the generated SDK.

The adapter assumes these endpoints:

- `GET /api/server/version` for server compatibility probing.
- `GET /api/api-keys/me` for API key validation.
- `GET /api/albums` for album listing.
- `GET /api/albums/{id}` for selected album asset candidates.
- `POST /api/search/random` for random library candidates.
- `GET /api/assets/{id}/thumbnail?size=preview&format=WEBP` for display-targeted non-original renditions.

Older archived API pages used `GET /api/asset/thumbnail/{id}`. The MVP adapter intentionally uses the current plural `/assets/{id}/thumbnail` path and requests `size=preview` so it does not download originals by default.

## Filtering And Metadata

Candidate listing is conservative:

- random search requests `type=IMAGE`, `visibility=timeline`, `withDeleted=false`, `withExif=true`, `withPeople=false`, and `withStacked=false`;
- album candidates are filtered locally to images only, timeline visibility only, and not archived or trashed;
- videos, archived assets, hidden assets, locked assets, trashed assets, raw people data, GPS, original paths, and full Immich asset blobs are excluded from normalized output.

Normalized metadata is limited to asset id, rendition identity, title, source name, taken date, media type, dimensions, and orientation.

## Manual Verification

No real Immich credentials or Matthew-instance context were available in this session, so no live Immich integration test was run. MVP repo/CI coverage remains mock HTTP unit tests only.
