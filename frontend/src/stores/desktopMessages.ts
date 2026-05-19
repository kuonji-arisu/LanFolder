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

  async function load() {
    return run(async () => {
      messages.value = await appApi.messages();
    });
  }

  async function send() {
    const text = draft.value.trim();
    if (!text) return taskSuccess(undefined);
    return run(async () => {
      const message = await appApi.sendMessage(text);
      messages.value = [...messages.value, message];
      draft.value = "";
    });
  }

  async function clear() {
    return run(async () => {
      await appApi.clearMessages();
      messages.value = [];
    });
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
  };
});
