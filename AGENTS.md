# Agent Constraints

- Treat Go as the source of truth for platform, runtime, filesystem, LAN address, and capability state.
- Frontend stores must call local API adapters, not generated Wails bindings directly.
- State-changing desktop commands must commit frontend state from the backend `AppState` snapshot returned by the command.
- Keep settings as single-setting commits. Do not replace this with a whole-form workflow or optimistic shared-store updates.
- Keep `AppService.SaveSettings()` simple and non-transactional: validate, apply required platform side effects, save JSON, update memory, restart the server if needed, then return the current snapshot plus any error.
- Keep permission metadata in Go; frontend code may use literal permission types but must not duplicate labels or descriptions.
- Preserve the split between desktop Wails commands and LAN HTTP file APIs.
- Keep path safety and file-operation rules inside `internal/share`.
- Preserve the trusted-LAN security model. Do not add authentication, password, or token layers unless the product direction changes.
- Treat `.trash` as a reserved managed directory that cannot be deleted through the LAN web API.
- Preserve server-mode build separation with `//go:build server` and `//go:build !server`.
- When changing Go service models exposed to the frontend, regenerate Wails TypeScript bindings.
