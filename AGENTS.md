# AI Agent Constraints

## Product Boundary

- LanFolder is a personal trusted-LAN utility, not an enterprise file server, NAS, sync service, cloud app, or public web app.
- Do not add user authentication, passwords, accounts, cloud sync, internet exposure, or user-facing/API bearer tokens unless the product direction explicitly changes.
- Default safety posture is conservative: `readonly` first; `upload` and `manage` only for trusted networks and devices.

## Architecture Constraints

- Go is the source of truth for platform/runtime state, filesystem behavior, LAN addresses, capabilities, and permission metadata; frontend may use permission values but must not duplicate permission labels/descriptions.
- Frontend stores must call local API adapters, not generated Wails bindings directly.
- State-changing desktop commands must commit frontend state from the backend `AppState` snapshot returned by the command.
- Preserve the split between desktop Wails commands and LAN HTTP APIs.
- Keep settings as single-setting commits; do not replace this with a whole-form workflow or optimistic shared-store updates.
- Keep `AppService.SaveSettings()` simple and non-transactional: validate, apply platform side effects, save JSON, update memory, restart if needed, then return the current snapshot plus any error.
- Keep app notifications on the AppNotice pipeline; do not add parallel event or toast paths.
- Keep LAN access approval small and Go-owned: a browser receives an opaque `lf_request` HttpOnly cookie for the request phase, the desktop approves an internal request ID, and the approved browser receives an opaque in-memory `lf_session` cookie.
- `lf_request` only identifies the pending approval request; `lf_session` only means "allowed to enter". Current share permission remains the source of truth for what the browser can do.
- Do not expose desktop request IDs, request codes, PINs, passwords, secrets, or login mechanisms to LAN browsers. Access polling must use the `lf_request` cookie, not a URL request ID.
- Treat IP as LAN boundary, display, logging, and rate-limit metadata only. User-Agent is display-only metadata. Neither IP nor User-Agent may affect authorization or request identity.
- Aggregate repeated access requests by `lf_request` with counts/last-seen metadata; do not notify or log every duplicate request.

## Safety Constraints

- Keep LAN filesystem and message safety rules inside `internal/share`, not in frontend code.
- Preserve the reserved `.lanfolder` model: it must not be accessible through the LAN file API, and trash belongs under `.lanfolder/trash`.
- Preserve stale-response protection for overlapping frontend async loads.

## Wails And Docs

- When Go service models exposed to the frontend change, regenerate Wails TypeScript bindings.
