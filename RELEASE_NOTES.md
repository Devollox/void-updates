# Installer Pattern Update

## Changed

- **Installer filename pattern for macOS and Linux** — updated `tempInstallerPattern()` to align with the new void‑presence release naming:
  - macOS: pattern changed from `Void.Presence.*.dmg` to `Void.Presence.Setup.*.dmg`.
  - Linux: pattern changed from `Void.Presence.*.deb` to `Void.Presence.Setup.*.deb`.
  - Windows: pattern remains `Void.Presence.Setup.*.exe`.
    This ensures the updater correctly selects the embedded installer for each platform using the unified `Setup.` prefix across all OS builds.
