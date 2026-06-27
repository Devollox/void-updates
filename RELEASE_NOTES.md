# Custom Updater

## Improvements

- **Added a custom updater flow** — the app now runs a bundled NSIS installer to apply updates directly from within Void Presence.
- **Automatic app relaunch** — after the installer finishes, the updater starts the installed `Void Presence.exe` from the user’s `%LOCALAPPDATA%\Programs\voidpresence` directory.
- **Smoother, hands-off updates** — the process runs silently and restarts the app without requiring any manual user actions.
