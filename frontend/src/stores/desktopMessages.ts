import { ref } from "vue";
import { defineStore } from "pinia";
import { taskSuccess } from "@/composables/useAsyncTask";
import { useAsyncTask } from "@/composables/useAsyncTask";
import { appApi } from "@/lib/appApi";
import { HOST_CLIENT_ID } from "@/lib/clientId";
import type { MessageEntry } from "@/types/app";

export const useDesktopMessagesStore = defineStore("desktopMessages", () => {
  const messages = ref<MessageEntry[]>([]);
  const draft = ref("");
  const clientId = HOST_CLIENT_ID;
  const { busy: loading, error, run } = useAsyncTask();
  let version = 0;

  async function load() {
    const requestVersion = ++version;
    return run(
      async () => {
        const loaded = await appApi.messages();
        if (requestVersion === version) {
          messages.value = loaded;
        }
      },
      undefined,
      {
        commit: () => requestVersion === version,
        stale: () => requestVersion !== version,
      },
    );
  }

  async function send() {
    const text = draft.value.trim();
    if (!text) return taskSuccess(undefined);
    const requestVersion = ++version;
    return run(async () => {
      const message = await appApi.sendMessage(text);
      if (requestVersion === version) {
        messages.value = [...messages.value, message];
        draft.value = "";
      }
    });
  }

  async function clear() {
    const requestVersion = ++version;
    return run(async () => {
      await appApi.clearMessages();
      if (requestVersion === version) {
        messages.value = [];
      }
    });
  }

  function reset() {
    version += 1;
    messages.value = [];
    draft.value = "";
  }

  return {
    messages,
    draft,
    clientId,
    loading,
    error,
    load,
    send,
    clear,
    reset,
  };
});
