# AI Agent Constraints

## Product Boundary

- LanFolder is a personal trusted-LAN utility, not an enterprise file server, NAS, sync service, cloud app, or public web app.
- Do not add authentication, passwords, tokens, accounts, cloud sync, or internet exposure unless the product direction explicitly changes.
- Default safety posture is conservative: `readonly` first; `upload` and `manage` only for trusted networks and devices.

## Architecture Constraints

- Go is the source of truth for platform/runtime state, filesystem behavior, LAN addresses, capabilities, and permission metadata; frontend may use permission values but must not duplicate permission labels/descriptions.
- Frontend stores must call local API adapters, not generated Wails bindings directly.
- State-changing desktop commands must commit frontend state from the backend `AppState` snapshot returned by the command.
- Preserve the split between desktop Wails commands and LAN HTTP APIs.
- Keep settings as single-setting commits; do not replace this with a whole-form workflow or optimistic shared-store updates.
- Keep `AppService.SaveSettings()` simple and non-transactional: validate, apply platform side effects, save JSON, update memory, restart if needed, then return the current snapshot plus any error.
- Keep app notifications on the AppNotice pipeline; do not add parallel event or toast paths.

## Safety Constraints

- Keep LAN filesystem and message safety rules inside `internal/share`, not in frontend code.
- Preserve the reserved `.lanfolder` model: it must not be accessible through the LAN file API, and trash belongs under `.lanfolder/trash`.
- Preserve stale-response protection for overlapping frontend async loads.

## Wails And Docs

- When Go service models exposed to the frontend change, regenerate Wails TypeScript bindings.