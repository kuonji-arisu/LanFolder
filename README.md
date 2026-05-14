# LanFolder

LanFolder is a small desktop app for sharing one local folder with nearby devices on the same trusted LAN.

Pick a folder, choose a permission level, start sharing, and open the shown address from another phone, tablet, or computer. The other device only needs a browser.

## What It Does

- Shares a selected local folder over HTTP on your local network
- Shows LAN access addresses in the desktop app
- Provides a mobile-friendly browser file manager
- Supports browsing, downloading, uploading, creating folders, and deleting files based on permission
- Provides a lightweight browser message panel for passing short text between LAN devices
- Keeps recent access logs in the desktop app
- Moves deleted files into the shared folder's reserved `.lanfolder/trash` directory
- Stores message history in the shared folder's reserved `.lanfolder` directory
- Sanitizes uploaded filenames and avoids overwriting existing files
- Blocks path traversal so requests stay inside the shared folder

## Permission Levels

`readonly`
: Browse and download files.

`upload`
: Browse, download, upload files, and create folders.

`manage`
: Browse, download, upload, create folders, and delete files.

## Safety Model

LanFolder is built for trusted local networks, such as your home Wi-Fi or a temporary private hotspot. It does not include accounts, passwords, or token-based access control. Anyone who can reach the displayed server address can use the currently selected permission level.

The LAN server rejects non-private remote addresses by default and blocks cross-site write requests, but it does not provide authentication.

The app focuses on filesystem safety instead:

- Request paths are constrained to the selected shared folder
- Hidden files are blocked unless explicitly enabled
- Uploaded names are sanitized
- Existing files are not overwritten by uploads
- Deletes are moved into `.lanfolder/trash`
- `.lanfolder` is reserved and cannot be accessed through the file API

Use `readonly` when you only need to send files out, and switch to `upload` or `manage` only when you trust the devices on the LAN.

## Desktop App

The desktop app has two main screens:

- Share: choose the shared folder, start or stop sharing, copy the access address, and view recent requests.
- Settings: change the port, permission level, auto-share behavior, tray behavior, startup behavior, theme, and hidden-file visibility.

## Browser File Manager

Devices on the same LAN can open the displayed address in a browser. The web interface adapts to the selected permission level:

- `readonly`: file list and download actions
- `upload`: upload and new-folder controls
- `manage`: delete controls in addition to upload features

The browser interface also includes a manual-refresh message panel for short text. Messages are stored as JSONL under `.lanfolder/messages.jsonl` and use a per-browser local client ID only to distinguish devices on the trusted LAN.

## Server Mode

LanFolder also has a server build for running without the desktop window. It reads configuration from the saved desktop config and environment variables such as:

- `LANFOLDER_ROOT`
- `LANFOLDER_HOST`
- `LANFOLDER_PORT`
- `LANFOLDER_PERMISSION`
- `LANFOLDER_SHOW_HIDDEN`

This mode is useful for lightweight LAN sharing on a trusted machine where a GUI is not needed.

## From Source

Requirements:

- Go
- pnpm
- Wails 3

Install frontend dependencies:

```powershell
cd frontend
pnpm install
```

Run the desktop app:

```powershell
cd ..
wails3 dev
```

Build:

```powershell
wails3 build
```

Run checks:

```powershell
go test ./...
cd frontend
pnpm test
pnpm build
```
