import { ref } from "vue";
import { defineStore } from "pinia";
import { taskSuccess } from "@/composables/useAsyncTask";
import { useAsyncTask } from "@/composables/useAsyncTask";
import { messageApi } from "@/lib/api";
import { localClientId } from "@/lib/clientId";
import type { MessageEntry } from "@/types/app";

export const useWebMessagesStore = defineStore("webMessages", () => {
  const messages = ref<MessageEntry[]>([]);
  const draft = ref("");
  const clientId = localClientId();
  const { busy: loading, error, run } = useAsyncTask();

  async function load() {
    return run(async () => {
      messages.value = await messageApi.list();
    });
  }

  async function send() {
    const text = draft.value.trim();
    if (!text) return taskSuccess(undefined);
    return run(async () => {
      const message = await messageApi.send(text, clientId);
      messages.value = [...messages.value, message];
      draft.value = "";
    });
  }

  async function clear() {
    return run(async () => {
      await messageApi.clear();
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
