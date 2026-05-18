import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi, type Mock } from "vitest";
import { createAppError } from "@/lib/errors";
import { noticeApi } from "@/lib/noticeApi";
import { useNoticeStore } from "@/stores/notices";
import type { AppNotice } from "@/types/app";

vi.mock("@/lib/noticeApi", () => ({
  noticeApi: {
    drainNotices: vi.fn(),
    listenNotices: vi.fn(),
    presentNotice: vi.fn(),
  },
}));

vi.mock("vue-sonner", () => ({
  toast: Object.assign(vi.fn(), {
    error: vi.fn(),
    success: vi.fn(),
    warning: vi.fn(),
  }),
}));

const api = noticeApi as unknown as Record<keyof typeof noticeApi, Mock>;

describe("useNoticeStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    api.drainNotices.mockResolvedValue([]);
    api.presentNotice.mockResolvedValue("toast");
  });

  it("drains pending backend notices and translates error payloads", async () => {
    const { toast } = await import("vue-sonner");
    api.drainNotices.mockResolvedValue([notice({ error: { error: "invalid_port" } })]);
    const store = useNoticeStore();

    await store.drain();

    expect(store.notices).toHaveLength(1);
    await vi.waitFor(() => expect(toast.error).toHaveBeenCalledWith("端口必须在 1 到 65535 之间"));
  });

  it("listens for realtime backend notices", async () => {
    const { toast } = await import("vue-sonner");
    let handler: ((notice: AppNotice) => void) | undefined;
    api.listenNotices.mockImplementation(async (next) => {
      handler = next;
      return vi.fn();
    });
    const store = useNoticeStore();

    await store.startListening();
    handler?.(notice({ id: "live", level: "warning", message: "warning" }));

    expect(store.notices[0].id).toBe("live");
    await vi.waitFor(() => expect(toast.warning).toHaveBeenCalledWith("warning"));
  });

  it("deduplicates notices by ID", async () => {
    const { toast } = await import("vue-sonner");
    const store = useNoticeStore();
    const item = notice({ id: "same", message: "one" });

    store.show(item);
    store.show(item);

    expect(store.notices).toHaveLength(1);
    await vi.waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));
  });

  it("shows task failures through the same notice path", async () => {
    const { toast } = await import("vue-sonner");
    const store = useNoticeStore();

    store.showTaskResult({
      ok: false,
      error: createAppError("shared_dir_required"),
      message: "请先选择共享目录",
    });

    expect(store.notices[0].error?.error).toBe("shared_dir_required");
    await vi.waitFor(() => expect(toast.error).toHaveBeenCalledWith("请先选择共享目录"));
  });

  it("shows fallback toasts for unmapped file operation failures", async () => {
    const { toast } = await import("vue-sonner");
    const store = useNoticeStore();

    store.showTaskResult({
      ok: false,
      error: new Error("rename failed"),
      message: "操作失败，请重试",
    });

    expect(store.notices[0].message).toBe("操作失败，请重试");
    await vi.waitFor(() => expect(toast.error).toHaveBeenCalledWith("操作失败，请重试"));
  });

  it("uses the common fallback for backend notices without payloads", async () => {
    const { toast } = await import("vue-sonner");
    const store = useNoticeStore();

    store.show(notice());

    await vi.waitFor(() => expect(toast.error).toHaveBeenCalledWith("操作失败，请重试"));
  });

  it("delegates presentation through the notice API", async () => {
    const store = useNoticeStore();
    const item = notice({ id: "present", message: "hello" });

    store.show(item);

    await vi.waitFor(() => expect(api.presentNotice).toHaveBeenCalledWith(item, "hello"));
  });

  it("does not show a page toast when the backend handles a system notification", async () => {
    const { toast } = await import("vue-sonner");
    api.presentNotice.mockResolvedValue("system");
    const store = useNoticeStore();

    store.show(notice({ message: "background" }));

    await vi.waitFor(() => expect(api.presentNotice).toHaveBeenCalled());
    expect(toast.error).not.toHaveBeenCalled();
  });
});

function notice(overrides: Partial<AppNotice> = {}): AppNotice {
  return {
    id: "notice",
    level: "error",
    source: "startup",
    createdAt: "2026-05-16T00:00:00Z",
    ...overrides,
  };
}
