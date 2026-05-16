# Security Policy

LanFolder is intended for trusted LAN use only.

It has no passwords, accounts, cloud identity, or internet-facing authentication. When new-device access approval is enabled, unknown browsers must be approved in the desktop app before they receive a temporary in-memory session cookie. When access approval is disabled, anyone who can reach the displayed LAN address can use the currently selected permission level.

Access approval is an HTTP LAN gate, not encrypted transport security. A device that can observe or tamper with traffic on the LAN may still see files or session cookies.

Do not expose LanFolder to the internet. Do not use it on public Wi-Fi, shared office networks with untrusted devices, or any network where unknown devices may be able to connect.

## Recommended Use

- Use LanFolder on home Wi-Fi, a private hotspot, or a temporary LAN between your own devices.
- Keep the permission mode at `readonly` unless you intentionally need uploads or deletion.
- Switch to `upload` or `manage` only when you trust the network and nearby devices.
- Enable new-device access approval before using automatic sharing.
- Stop sharing when you are done.

## Security Model

LanFolder relies on a trusted-network model. Access approval limits casual or accidental access from unknown browsers, but it does not make HTTP safe on untrusted networks.

It includes filesystem safety checks for path confinement, symlink escape, reserved `.lanfolder` paths, upload filenames, overwrite avoidance, hidden files, and trash-based deletion.

These checks reduce local file-operation risk, but they do not make LanFolder safe for untrusted networks.

## Reporting Security Issues

If you find a security issue, please do not open a public issue with exploit details.

Report it privately by contacting the maintainer or using GitHub's private vulnerability reporting if it is enabled for this repository.
