# Process-Safe Installer Update

## Improvements

- **Force-close running client before update** — the installer now terminates all running `Void Presence.exe` processes via `taskkill` before launching the NSIS installer, preventing file lock errors and half-applied updates.
