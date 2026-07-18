# Post-Install Cleanup

## Bug Fixes

- **Automatic installer cleanup on Windows** — after Void Presence finishes installing, the bundled Windows installer now removes its temporary setup binary and clears update/cache directories under AppData, so updater runs no longer leave behind stray files.
