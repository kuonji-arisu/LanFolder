# Security Policy

LanFolder is intended for trusted LAN use only.

It has no authentication, no passwords, no accounts, and no access tokens. Anyone who can reach the displayed LAN address can use the currently selected permission level.

Do not expose LanFolder to the internet. Do not use it on public Wi-Fi, shared office networks with untrusted devices, or any network where unknown devices may be able to connect.

## Recommended Use

- Use LanFolder on home Wi-Fi, a private hotspot, or a temporary LAN between your own devices.
- Keep the permission mode at `readonly` unless you intentionally need uploads or deletion.
- Switch to `upload` or `manage` only when you trust the network and nearby devices.
- Stop sharing when you are done.

## Security Model

LanFolder relies on a trusted-network model, not user authentication.

It includes filesystem safety checks for path confinement, symlink escape, reserved `.lanfolder` paths, upload filenames, overwrite avoidance, hidden files, and trash-based deletion.

These checks reduce local file-operation risk, but they do not make LanFolder safe for untrusted networks.

## Reporting Security Issues

If you find a security issue, please do not open a public issue with exploit details.

Report it privately by contacting the maintainer or using GitHub's private vulnerability reporting if it is enabled for this repository.