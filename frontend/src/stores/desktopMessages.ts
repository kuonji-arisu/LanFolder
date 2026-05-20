import { ref } from "vue";
import { defineStore } from "pinia";
import { taskSuccess } from "@/composables/useAsyncTask";
import { useLatestAsyncTask } from "@/composables/useLatestAsyncTask";
import { appApi } from "@/lib/appApi";
import { HOST_CLIENT_ID } from "@/lib/clientId";
import type { MessageEntry } from "@/types/app";

export const useDesktopMessagesStore = defineStore("desktopMessages", () => {
  const messages = ref<MessageEntry[]>([]);
  const draft = ref("");
  const clientId = HOST_CLIENT_ID;
  const { busy: loading, error, runLatest, invalidate } = useLatestAsyncTask();

  async function load() {
    return runLatest(appApi.messages, (loaded) => {
      messages.value = loaded;
    });
  }

  async function send() {
    const text = draft.value.trim();
    if (!text) return taskSuccess(undefined);
    return runLatest(() => appApi.sendMessage(text), (message) => {
      messages.value = [...messages.value, message];
      draft.value = "";
    });
  }

  async function clear() {
    return runLatest(appApi.clearMessages, () => {
      messages.value = [];
    });
  }

  function reset() {
    invalidate();
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
