<script setup lang="ts">
import { computed, ref } from "vue";
import { Loader2, RefreshCw, Trash2 } from "lucide-vue-next";
import AppDialogContent from "@/components/app/AppDialogContent.vue";
import AppDialogHeader from "@/components/app/AppDialogHeader.vue";
import MessagePanel from "@/components/messages/MessagePanel.vue";
import { Button } from "@/components/ui/button";
import { Dialog, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import type { TaskResult } from "@/composables/useAsyncTask";
import { useI18n } from "@/lib/i18n";
import { useNoticeStore } from "@/stores/notices";
import type { MessageEntry } from "@/types/app";

const open = defineModel<boolean>("open", { required: true });
const draft = defineModel<string>("draft", { required: true });
const props = defineProps<{
  messages: MessageEntry[];
  currentClientId: string;
  loading?: boolean;
  canSend?: boolean;
  canClear?: boolean;
  disabled?: boolean;
  disabledText?: string;
  sendDisabledText?: string;
  emptyText?: string;
  width?: string;
  height?: string;
  maxWidth?: string;
  maxHeight?: string;
  loadMessages: () => Promise<TaskResult<unknown>>;
  sendMessage: () => Promise<TaskResult<unknown>>;
  clearMessages: () => Promise<TaskResult<unknown>>;
}>();
const notices = useNoticeStore();
const { t } = useI18n();
const clearConfirmOpen = ref(false);
const fillAvailableHeight = computed(() => Boolean(props.height || props.maxHeight));

async function refreshMessages() {
  notices.showTaskResult(await props.loadMessages());
}

async function sendMessage() {
  notices.showTaskResult(await props.sendMessage());
}

async function clearMessages() {
  const result = await props.clearMessages();
  notices.showTaskResult(result);
  if (result.ok) {
    clearConfirmOpen.value = false;
    notices.showSuccess(t("message.clearSuccess"));
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <AppDialogContent :width="props.width" :height="props.height" :max-width="props.maxWidth" :max-height="props.maxHeight">
      <AppDialogHeader>
        <DialogTitle>{{ t("message.title") }}</DialogTitle>
        <DialogDescription>{{ t("message.description") }}</DialogDescription>

        <template #actions>
          <Button variant="secondary" size="icon" :disabled="loading" :aria-label="t('common.refresh')" @click="refreshMessages">
            <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
            <RefreshCw v-else class="h-4 w-4" />
          </Button>
          <Button v-if="canClear" variant="secondary" size="icon" :disabled="loading || !messages.length" :aria-label="t('message.clearMessages')" @click="clearConfirmOpen = true">
            <Trash2 class="h-4 w-4" />
          </Button>
        </template>
      </AppDialogHeader>

      <MessagePanel
        v-model:draft="draft"
        :class="{ 'message-panel--fill': fillAvailableHeight }"
        :messages="messages"
        :current-client-id="currentClientId"
        :loading="loading"
        :disabled="disabled"
        :disabled-text="disabledText"
        :input-disabled="!canSend"
        :input-disabled-text="sendDisabledText"
        :empty-text="emptyText || t('message.emptyRelay')"
        @send="sendMessage"
      />
    </AppDialogContent>
  </Dialog>

  <Dialog v-model:open="clearConfirmOpen">
    <AppDialogContent max-width="420px">
      <AppDialogHeader>
        <DialogTitle>{{ t("message.clearMessages") }}</DialogTitle>
        <DialogDescription>{{ t("message.clearDescription") }}</DialogDescription>
      </AppDialogHeader>
      <div class="message-clear-actions">
        <Button variant="secondary" :disabled="loading" @click="clearConfirmOpen = false">{{ t("common.cancel") }}</Button>
        <Button :disabled="loading" @click="clearMessages">
          <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
          {{ t("common.clear") }}
        </Button>
      </div>
    </AppDialogContent>
  </Dialog>
</template>

<style scoped>
.message-clear-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.message-clear-actions {
  justify-content: flex-end;
}
</style>
