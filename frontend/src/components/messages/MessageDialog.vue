<script setup lang="ts">
import { ref } from "vue";
import { Loader2, RefreshCw, Trash2 } from "lucide-vue-next";
import MessagePanel from "@/components/messages/MessagePanel.vue";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import type { TaskResult } from "@/composables/useAsyncTask";
import { useI18n } from "@/lib/i18n";
import { useNoticeStore } from "@/stores/notices";
import type { MessageEntry } from "@/types/app";

type DialogLayout = "default" | "bounded";

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
  layout?: DialogLayout;
  loadMessages: () => Promise<TaskResult<unknown>>;
  sendMessage: () => Promise<TaskResult<unknown>>;
  clearMessages: () => Promise<TaskResult<unknown>>;
}>();
const notices = useNoticeStore();
const { t } = useI18n();
const clearConfirmOpen = ref(false);

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
    <DialogContent :size="layout === 'bounded' ? 'large' : undefined" :height="layout === 'bounded' ? 'bounded' : undefined">
      <div class="message-dialog-head">
        <DialogHeader>
          <DialogTitle>{{ t("message.title") }}</DialogTitle>
          <DialogDescription>{{ t("message.description") }}</DialogDescription>
        </DialogHeader>
        <div class="message-dialog-actions">
          <Button variant="secondary" size="icon" :disabled="loading" :aria-label="t('common.refresh')" @click="refreshMessages">
            <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
            <RefreshCw v-else class="h-4 w-4" />
          </Button>
          <Button v-if="canClear" variant="secondary" size="icon" :disabled="loading || !messages.length" :aria-label="t('message.clearMessages')" @click="clearConfirmOpen = true">
            <Trash2 class="h-4 w-4" />
          </Button>
        </div>
      </div>
      <MessagePanel
        v-model:draft="draft"
        :class="{ 'message-panel--bounded': layout === 'bounded' }"
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
    </DialogContent>
  </Dialog>

  <Dialog v-model:open="clearConfirmOpen">
    <DialogContent size="sm">
      <DialogHeader>
        <DialogTitle>{{ t("message.clearMessages") }}</DialogTitle>
        <DialogDescription>{{ t("message.clearDescription") }}</DialogDescription>
      </DialogHeader>
      <div class="message-clear-actions">
        <Button variant="secondary" :disabled="loading" @click="clearConfirmOpen = false">{{ t("common.cancel") }}</Button>
        <Button :disabled="loading" @click="clearMessages">
          <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
          {{ t("common.clear") }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.message-dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding-right: 44px;
}

.message-dialog-actions,
.message-clear-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.message-clear-actions {
  justify-content: flex-end;
}
</style>
