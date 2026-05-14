import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi, type Mock } from "vitest";
import { messageApi } from "@/lib/api";
import { useWebMessagesStore } from "@/stores/webMessages";

vi.mock("@/lib/api", () => ({
  messageApi: {
    clear: vi.fn(),
    list: vi.fn(),
    send: vi.fn(),
  },
}));

vi.mock("@/lib/clientId", () => ({
  localClientId: vi.fn(() => "client-1234"),
}));

const api = messageApi as unknown as Record<keyof typeof messageApi, Mock>;

describe("useWebMessagesStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("loads messages", async () => {
    api.list.mockResolvedValue([message("1", "hello")]);
    const store = useWebMessagesStore();

    await store.load();

    expect(api.list).toHaveBeenCalledTimes(1);
    expect(store.messages[0].text).toBe("hello");
  });

  it("sends and appends a message", async () => {
    api.send.mockResolvedValue(message("2", "sent"));
    const store = useWebMessagesStore();
    store.draft = " sent ";

    await store.send();

    expect(api.send).toHaveBeenCalledWith("sent", "client-1234");
    expect(store.messages).toHaveLength(1);
    expect(store.draft).toBe("");
  });

  it("clears messages", async () => {
    api.clear.mockResolvedValue({ ok: true });
    const store = useWebMessagesStore();
    store.messages = [message("1", "hello")];

    await store.clear();

    expect(api.clear).toHaveBeenCalledTimes(1);
    expect(store.messages).toEqual([]);
  });
});

function message(id: string, text: string) {
  return {
    id,
    text,
    clientId: "client-1234",
    createdAt: "2026-05-14T00:00:00Z",
  };
}
