import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi, type Mock } from "vitest";
import { appApi } from "@/lib/appApi";
import { useAppStore } from "@/stores/app";
import type { AppState } from "@/types/app";

vi.mock("@wailsio/runtime", () => ({
  Events: {
    On: vi.fn(() => vi.fn()),
  },
}));

vi.mock("@/lib/appApi", () => ({
  appApi: {
    state: vi.fn(),
    logs: vi.fn(),
    chooseFolder: vi.fn(),
    openSharedFolder: vi.fn(),
    startSharing: vi.fn(),
    stopSharing: vi.fn(),
    saveSettings: vi.fn(),
  },
}));

const api = appApi as unknown as Record<keyof typeof appApi, Mock>;

describe("useAppStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    api.logs.mockResolvedValue([]);
  });

  it("commits the snapshot returned by saveSettings", async () => {
    api.state.mockResolvedValue(appState({ permission: "readonly", running: true, port: 8899 }));
    const store = useAppStore();
    await store.loadSnapshot();

    api.saveSettings.mockResolvedValue(appState({ permission: "manage", running: true, port: 9000 }));
    await store.saveConfig({ permission: "manage", port: 9000 });

    expect(api.saveSettings).toHaveBeenCalledWith(expect.objectContaining({ permission: "manage", port: 9000 }));
    expect(store.state?.config.permission).toBe("manage");
    expect(store.state?.server.permission).toBe("manage");
    expect(store.primaryAddress).toBe("http://192.168.1.20:9000");
  });

  it("keeps the previous snapshot when saveSettings fails", async () => {
    api.state.mockResolvedValue(appState({ permission: "readonly", running: true, port: 8899 }));
    const store = useAppStore();
    await store.loadSnapshot();

    api.saveSettings.mockRejectedValue({ message: JSON.stringify({ error: "invalid_port" }), kind: "RuntimeError" });
    const saved = await store.saveConfig({ port: 0 });

    expect(saved.ok).toBe(false);
    expect(store.state?.config.port).toBe(8899);
    expect(store.state?.server.port).toBe(8899);
    expect(store.error).toBe("端口必须在 1 到 65535 之间");
    if (!saved.ok) expect(saved.error).toEqual(expect.objectContaining({ kind: "RuntimeError" }));
  });

  it("uses the same snapshot commit path for sharing commands", async () => {
    api.state.mockResolvedValue(appState({ running: false }));
    const store = useAppStore();
    await store.loadSnapshot();

    api.startSharing.mockResolvedValue(appState({ running: true }));
    await store.toggleSharing();

    expect(api.startSharing).toHaveBeenCalledTimes(1);
    expect(store.isRunning).toBe(true);

    api.stopSharing.mockResolvedValue(appState({ running: false }));
    await store.toggleSharing();

    expect(api.stopSharing).toHaveBeenCalledTimes(1);
    expect(store.isRunning).toBe(false);
  });

  it("runs non-state commands without replacing the snapshot", async () => {
    const initial = appState({ permission: "upload" });
    api.state.mockResolvedValue(initial);
    api.openSharedFolder.mockResolvedValue(undefined);
    const store = useAppStore();
    await store.loadSnapshot();

    await store.openSharedFolder();

    expect(api.openSharedFolder).toHaveBeenCalledTimes(1);
    expect(store.state).toEqual(initial);
  });

});

function appState(overrides: { permission?: "readonly" | "upload" | "manage"; running?: boolean; port?: number } = {}): AppState {
  const permission = overrides.permission ?? "readonly";
  const port = overrides.port ?? 8899;
  return {
    config: {
      sharedDir: "C:/Share",
      port,
      permission,
      autoShare: false,
      startAtLogin: false,
      keepInTray: false,
      showHiddenFiles: false,
    },
    server: {
      running: overrides.running ?? false,
      host: "0.0.0.0",
      port,
      root: "C:/Share",
      permission,
    },
    appInfo: {
      name: "LanFolder",
      version: "test",
      os: "windows",
      osName: "Windows",
      osVersion: "11",
      arch: "amd64",
      locale: "zh-CN",
      systemTheme: "light",
      configPath: "C:/config.json",
      window: { x: 0, y: 0, width: 350, height: 600, maximised: false, minimised: false, fullscreen: false },
    },
    capabilities: {
      startAtLogin: true,
      tray: true,
      openFolder: true,
      systemThemeEvents: true,
      windowState: true,
    },
    addresses: [`http://192.168.1.20:${port}`],
    permissions: [
      { value: "readonly", label: "Read", description: "Read only" },
      { value: "upload", label: "Upload", description: "Upload files" },
      { value: "manage", label: "Manage", description: "Manage files" },
    ],
  };
}
