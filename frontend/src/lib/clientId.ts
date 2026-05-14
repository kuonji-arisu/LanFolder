const storageKey = "lanfolder.clientId";

export function localClientId() {
  const stored = readStoredClientId();
  if (stored) return stored;

  const id = newClientId();
  try {
    window.localStorage.setItem(storageKey, id);
  } catch {
    // Storage can be unavailable in private or restricted contexts.
  }
  return id;
}

export function clientLabel(clientId: string, currentClientId: string) {
  if (clientId && clientId === currentClientId) return "我";
  if (!clientId) return "未知设备";
  return `设备 ${shortClientId(clientId)}`;
}

function readStoredClientId() {
  try {
    return window.localStorage.getItem(storageKey) || "";
  } catch {
    return "";
  }
}

function newClientId() {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID();
  const bytes = new Uint8Array(16);
  if (globalThis.crypto?.getRandomValues) {
    globalThis.crypto.getRandomValues(bytes);
    return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join("");
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`;
}

function shortClientId(clientId: string) {
  return clientId.replace(/[^a-zA-Z0-9]/g, "").slice(-4).toUpperCase() || "----";
}
