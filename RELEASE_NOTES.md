# Cross-Platform Update Support

## New

- **macOS support** — the updater now embeds and runs the `.dmg` installer on macOS: it mounts the image with `hdiutil`, copies the `.app` to `/Applications` via `ditto`, then unmounts.
- **Linux support** — the updater now embeds and runs the `.deb` package on Linux using `pkexec dpkg -i` and relaunches the app from `/usr/bin/void-presence` on completion.

## Improved

- **Platform-aware binary selection** — `extractInstaller` now looks for the correct embedded asset prefix and extension per OS (`Void.Presence.Setup.*.exe` on Windows, `Void.Presence.*.dmg` on macOS, `Void.Presence.*.deb` on Linux).
- **Cross-platform process tracking** — `isProcessAlive` uses the Windows API on Windows and `syscall.Kill(pid, 0)` on Unix, keeping the goroutine-based install watcher working on all platforms.
- **Cross-platform log path** — log files are written to `%LOCALAPPDATA%\voidupdates` on Windows, `~/Library/Logs/voidupdates` on macOS, and `~/.local/share/voidupdates` on Linux.
- **GitHub Actions matrix build** — `build-installer.yml` and `test.yml` now build the updater for all three platforms in parallel, each embedding the matching platform artifact from void-presence.
