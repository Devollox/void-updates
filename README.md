<img width="3844" height="793" alt="484064966-2c662772-bca231-4de4-988f-5304d7dfd87d" src="https://github.com/user-attachments/assets/eea692df-b300-45de-8acb-03ab75cfdf3c" />

##

Modern, high‑performance **updater** for Void Presence built with **Go (Wails)** and **React**.

## Features

- **Custom updater flow**: Launches the bundled Void Presence NSIS installer and passes the required arguments.
- **Automatic relaunch**: After installation, it automatically starts the updated `Void Presence.exe` from `%LOCALAPPDATA%\Programs\voidpresence`.
- **Hands-off updates**: Minimal interaction — the update is installed and the app is restarted automatically.
- **Version-aware installer**: A dedicated `Void.Presence.Updates.<version>.exe` is built for each Void Presence release.

## Built With

- [Go](https://go.dev) + [Wails](https://wails.io) (Backend & OS Integration)
- [React](https://reactjs.org) + [TypeScript](https://typescriptlang.org) (Frontend)
- [Tailwind CSS](https://tailwindcss.com) (Styling)

## How It Works

1. Void Presence downloads `Void.Presence.Updates.<version>.exe` from the [void-updates Releases](https://github.com/Devollox/void-updates/releases) page.
2. The updater launches and checks the installation path in `%LOCALAPPDATA%\Programs\voidpresence`.
3. The embedded NSIS installer `Void.Presence.Setup.<version>.exe` is extracted and run silently.
4. After a successful installation, the updater relaunches Void Presence and exits.

## Download

Download the latest `Void.Presence.Updates.<version>.exe` from the [Releases](https://github.com/Devollox/void-updates/releases) page and use it as the external updater for Void Presence.

## Author

Made with ❤️ by [Devollox](https://github.com/Devollox)

<p align="left">
  <img width="128" height="128" alt="выфвфы" src="https://github.com/user-attachments/assets/32b65183-a39c-4871-bb37-5fbe01ecaade" />
</p>

**Void Presence** – Control your Discord presence. Own your story.
