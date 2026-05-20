import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi, type Mock } from "vitest";
import { appApi } from "@/lib/appApi";
import { useDesktopMessagesStore } from "@/stores/desktopMessages";
import type { MessageEntry } from "@/types/app";

vi.mock("@/lib/appApi", () => ({
  appApi: {
    clearMessages: vi.fn(),
    messages: vi.fn(),
    sendMessage: vi.fn(),
  },
}));

const api = appApi as unknown as Record<"clearMessages" | "messages" | "sendMessage", Mock>;

describe("useDesktopMessagesStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("does not let a reset in-flight load restore stale messages", async () => {
    const pending = deferred<MessageEntry[]>();
    api.messages.mockReturnValue(pending.promise);
    const store = useDesktopMessagesStore();
    store.messages = [message("old", "old")];

    const loading = store.load();
    store.reset();
    pending.resolve([message("stale", "stale")]);
    await loading;

    expect(store.messages).toEqual([]);
  });

  it("does not let a reset in-flight load commit stale errors", async () => {
    const pending = deferred<MessageEntry[]>();
    api.messages.mockReturnValue(pending.promise);
    const store = useDesktopMessagesStore();

    const loading = store.load();
    store.reset();
    pending.reject(new Error("share_not_running"));
    const result = await loading;

    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.stale).toBe(true);
    expect(store.error).toBe("");
  });

  it("keeps the latest load when loads finish out of order", async () => {
    const first = deferred<MessageEntry[]>();
    const second = deferred<MessageEntry[]>();
    api.messages.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const store = useDesktopMessagesStore();

    const firstLoad = store.load();
    const secondLoad = store.load();
    second.resolve([message("new", "new")]);
    await secondLoad;
    first.resolve([message("stale", "stale")]);
    await firstLoad;

    expect(store.messages).toEqual([message("new", "new")]);
  });

  it("does not let an older load overwrite a sent message", async () => {
    const pending = deferred<MessageEntry[]>();
    api.messages.mockReturnValue(pending.promise);
    api.sendMessage.mockResolvedValue(message("sent", "sent"));
    const store = useDesktopMessagesStore();
    store.draft = " sent ";

    const loading = store.load();
    await store.send();
    pending.resolve([message("stale", "stale")]);
    await loading;

    expect(store.messages).toEqual([message("sent", "sent")]);
    expect(store.draft).toBe("");
  });

  it("does not let a reset in-flight send append stale messages", async () => {
    const pending = deferred<MessageEntry>();
    api.sendMessage.mockReturnValue(pending.promise);
    const store = useDesktopMessagesStore();
    store.draft = " sent ";

    const sending = store.send();
    store.reset();
    pending.resolve(message("sent", "sent"));
    await sending;

    expect(store.messages).toEqual([]);
    expect(store.draft).toBe("");
  });

  it("does not let a reset in-flight send commit stale errors", async () => {
    const pending = deferred<MessageEntry>();
    api.sendMessage.mockReturnValue(pending.promise);
    const store = useDesktopMessagesStore();
    store.draft = " sent ";

    const sending = store.send();
    store.reset();
    pending.reject(new Error("share_not_running"));
    const result = await sending;

    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.stale).toBe(true);
    expect(store.error).toBe("");
  });

  it("does not let an older clear remove newer messages", async () => {
    const pending = deferred<void>();
    api.clearMessages.mockReturnValue(pending.promise);
    const store = useDesktopMessagesStore();
    store.messages = [message("old", "old")];

    const clearing = store.clear();
    api.messages.mockResolvedValue([message("new", "new")]);
    await store.load();
    pending.resolve();
    await clearing;

    expect(store.messages).toEqual([message("new", "new")]);
  });

  it("does not let a reset in-flight clear commit stale errors", async () => {
    const pending = deferred<void>();
    api.clearMessages.mockReturnValue(pending.promise);
    const store = useDesktopMessagesStore();

    const clearing = store.clear();
    store.reset();
    pending.reject(new Error("share_not_running"));
    const result = await clearing;

    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.stale).toBe(true);
    expect(store.error).toBe("");
  });
});

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
}

function message(id: string, text: string): MessageEntry {
  return {
    id,
    text,
    clientId: "host",
    createdAt: "2026-05-14T00:00:00Z",
  };
}
