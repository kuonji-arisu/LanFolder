# LanFolder

[English](README.md) | [中文](README.zh-CN.md)

LanFolder is a small desktop app for sharing one local folder with nearby devices on the same trusted LAN.

It started as a personal vibe-coding project: I wanted a quick way to move files between my computer, phone, and tablet without setting up SMB, plugging in a cable, or sending files through a chat app.

It is not meant to be a full file server, NAS, sync service, or internet-facing app. It is a simple local-network utility with a desktop control panel and a browser-based file manager.

## Features

- Share one selected local folder over HTTP on your LAN
- Show available LAN addresses in the desktop app
- Open the shared folder from a phone, tablet, or another computer with a browser
- Browse and download files
- Upload files when upload permission is enabled
- Create folders when upload permission is enabled
- Delete files when manage permission is enabled
- Move deleted files into `.lanfolder/trash`
- Send short text messages between LAN devices
- Store messages in `.lanfolder/messages.jsonl`
- Show recent access logs in the desktop app
- Require desktop approval for new browser access when enabled
- Hide dotfiles by default
- Keep running in the tray if enabled
- Start sharing automatically when access approval is enabled

## Permission Levels

LanFolder has three permission modes.

### `readonly`

Devices on the LAN can browse and download files.

This is the safest default mode.

### `upload`

Devices on the LAN can browse, download, upload files, and create folders.

Use this when you want to receive files from trusted devices.

### `manage`

Devices on the LAN can browse, download, upload, create folders, and delete files.

Deleted files are moved into `.lanfolder/trash` instead of being removed directly.

Use this only on a trusted network.

## Safety Model

LanFolder is designed for trusted local networks, such as home Wi-Fi, a private hotspot, or a temporary LAN between your own devices.

When new-device access approval is enabled, an unknown browser can only request access. The request is tracked with a short-lived HttpOnly cookie, and the desktop app must approve it before that browser receives a temporary in-memory session cookie.

Access sessions are kept in memory and are cleared when sharing stops, the shared folder or permission changes, the server restarts, or the app exits. Automatic sharing requires new-device access approval to be enabled first.

When access approval is disabled, anyone who can reach the displayed LAN address can use the current permission level.

The server rejects non-private remote addresses by default and blocks cross-site write requests, but HTTP access is still not encrypted. Do not expose LanFolder to the internet or use it on networks with untrusted devices.

Use `readonly` by default, and switch to `upload` or `manage` only when the network and devices around you are trusted.

## Managed Directory

LanFolder creates a reserved `.lanfolder` directory inside the shared folder for app-managed data:

```text
.lanfolder/
├── trash/
└── messages.jsonl
````

The `.lanfolder` directory is hidden from the LAN file API and cannot be accessed or deleted through the browser interface.

## Project Scope

LanFolder is a personal tool first. The goal is to keep it small, direct, and easy to use.

The code still pays attention to the risky parts, especially filesystem behavior: path traversal, symlink escape, reserved managed paths, upload filenames, accidental overwrite, hidden files, and delete behavior.

## Tech Stack

* Go
* Wails 3
* Vue 3
* TypeScript
* Pinia
* Tailwind CSS
* shadcn-vue style components

## Development

Requirements:

* Go
* pnpm
* Wails 3

Install frontend dependencies:

```bash
cd frontend
pnpm install
```

Run in development mode:

```bash
cd ..
wails3 dev
```

Build:

```bash
wails3 build
```

Translations live in `locales/*.json`. After editing them, regenerate the Go catalog:

```bash
go generate ./internal/i18n
```

Run checks:

```bash
go test ./...
cd frontend
pnpm test
pnpm build
```
