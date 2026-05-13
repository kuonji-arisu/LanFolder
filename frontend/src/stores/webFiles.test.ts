import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi, type Mock } from "vitest";
import { fileApi } from "@/lib/api";
import { createAppError } from "@/lib/errors";
import { useWebFilesStore } from "@/stores/webFiles";

vi.mock("@/lib/api", () => ({
  fileApi: {
    status: vi.fn(),
    list: vi.fn(),
    upload: vi.fn(),
    delete: vi.fn(),
    mkdir: vi.fn(),
    downloadUrl: vi.fn((path: string) => `/api/download?path=${encodeURIComponent(path)}`),
  },
}));

const api = fileApi as unknown as Record<keyof typeof fileApi, Mock>;

describe("useWebFilesStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    api.status.mockResolvedValue({
      running: true,
      port: 8899,
      permission: "manage",
      permissions: [
        { value: "readonly", label: "Read", description: "" },
        { value: "upload", label: "Upload", description: "" },
        { value: "manage", label: "Manage", description: "" },
      ],
    });
  });

  it("loads status and listing for the requested path", async () => {
    api.list.mockResolvedValue(listing("docs", [{ name: "note.txt", path: "docs/note.txt", isDir: false }]));
    const store = useWebFilesStore();

    await store.load("docs");

    expect(api.status).toHaveBeenCalledTimes(1);
    expect(api.list).toHaveBeenCalledWith("docs");
    expect(store.currentPath).toBe("docs");
    expect(store.canUpload).toBe(true);
    expect(store.canDelete).toBe(true);
    expect(store.permissionLabel).toBe("Manage");
  });

  it("keeps the latest listing when navigation requests finish out of order", async () => {
    const slow = deferred<ReturnType<typeof listing>>();
    const fast = deferred<ReturnType<typeof listing>>();
    api.list.mockImplementation((path: string) => (path === "slow" ? slow.promise : fast.promise));
    const store = useWebFilesStore();

    const slowLoad = store.load("slow");
    const fastLoad = store.load("fast");

    fast.resolve(listing("fast", [{ name: "fast.txt", path: "fast/fast.txt", isDir: false }]));
    await fastLoad;
    expect(store.currentPath).toBe("fast");

    slow.resolve(listing("slow", [{ name: "slow.txt", path: "slow/slow.txt", isDir: false }]));
    await slowLoad;
    expect(store.currentPath).toBe("fast");
    expect(store.listing?.entries[0].name).toBe("fast.txt");
  });

  it("refreshes the current folder after upload", async () => {
    api.list.mockResolvedValue(listing(""));
    api.upload.mockResolvedValue({ entries: [] });
    const store = useWebFilesStore();
    await store.load("");

    const file = new File(["hello"], "note.txt", { type: "text/plain" });
    await store.uploadFiles([file] as unknown as FileList);

    expect(api.upload).toHaveBeenCalledWith("", [file]);
    expect(api.list).toHaveBeenCalledTimes(2);
  });

  it("confirms deletion before deleting and refreshing", async () => {
    vi.spyOn(window, "confirm").mockReturnValue(true);
    api.list.mockResolvedValue(listing("", [{ name: "note.txt", path: "note.txt", isDir: false }]));
    api.delete.mockResolvedValue({ ok: true });
    const store = useWebFilesStore();
    await store.load("");

    await store.deleteEntry({ name: "note.txt", path: "note.txt", isDir: false, size: 1, modTime: new Date().toISOString() });

    expect(api.delete).toHaveBeenCalledWith("note.txt");
    expect(api.list).toHaveBeenCalledTimes(2);
  });

  it("creates a folder and clears the input", async () => {
    api.list.mockResolvedValue(listing(""));
    api.mkdir.mockResolvedValue({ name: "docs", path: "docs", isDir: true, size: 0, modTime: new Date().toISOString() });
    const store = useWebFilesStore();
    await store.load("");
    store.newFolderName = "docs";

    await store.createFolder();

    expect(api.mkdir).toHaveBeenCalledWith("", "docs");
    expect(store.newFolderName).toBe("");
  });

  it("keeps input and returns failure when a file operation fails", async () => {
    api.list.mockResolvedValue(listing(""));
    api.mkdir.mockRejectedValue(createAppError("permission_denied", { status: 403 }));
    const store = useWebFilesStore();
    await store.load("");
    store.newFolderName = "docs";

    const result = await store.createFolder();

    expect(result.ok).toBe(false);
    expect(store.error).toBe("没有权限执行此操作");
    expect(store.newFolderName).toBe("docs");
  });
});

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((promiseResolve, promiseReject) => {
    resolve = promiseResolve;
    reject = promiseReject;
  });
  return { promise, resolve, reject };
}

function listing(path: string, entries: Array<{ name: string; path: string; isDir: boolean }> = []) {
  return {
    path,
    parentPath: "",
    entries: entries.map((entry) => ({
      ...entry,
      size: 0,
      modTime: new Date("2026-05-09T00:00:00Z").toISOString(),
    })),
  };
}
