# Refactor Name Updates

## Improvements

- **Renamed the updater app to “Void Presence Updates”** — clearly separates the updater helper from the main Void Presence client while keeping branding consistent.
- **Added a version sync script** — a Node-based helper now reads the version from `wails.json` and propagates it to `frontend/package.json` and `frontend/package-lock.json`, keeping Go and frontend versions in lockstep.

## Technical

- **Single source of truth for version** — the script trims `wails.json.version`, normalizes it for Go (`vX.Y.Z` when needed) and uses the raw value for the frontend package version.
- **Safe lockfile updates** — if `frontend/package-lock.json` exists, both `version` at the root and `packages[""].version` are updated to match, preventing mismatched versions in the npm ecosystem.
- **Fail-fast on errors** — any read/parse/update error causes the script to exit with code `1`, failing the build instead of producing a release with inconsistent version metadata.
