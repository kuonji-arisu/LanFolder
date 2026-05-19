<script setup lang="ts">
import { computed } from "vue";
import { Send } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { formatDate } from "@/lib/format";
import { clientLabel } from "@/lib/clientId";
import { useI18n } from "@/lib/i18n";
import type { MessageEntry } from "@/types/app";

const props = defineProps<{
  messages: MessageEntry[];
  currentClientId: string;
  draft: string;
  loading?: boolean;
  disabled?: boolean;
  inputDisabled?: boolean;
  emptyText?: string;
  disabledText?: string;
  inputDisabledText?: string;
}>();

const emit = defineEmits<{
  (event: "update:draft", value: string): void;
  (event: "send"): void;
}>();
const { t } = useI18n();

const draftModel = computed({
  get: () => props.draft,
  set: (value: string) => emit("update:draft", value),
});

const canSend = computed(() => !props.disabled && !props.inputDisabled && !props.loading && props.draft.trim().length > 0);

function messageLabel(message: MessageEntry) {
  return clientLabel(message.clientId, props.currentClientId);
}

function isOwnMessage(message: MessageEntry) {
  return message.clientId === props.currentClientId;
}
</script>

<template>
  <section class="message-panel" :aria-label="t('message.aria')">
    <div class="message-list">
      <div v-if="disabled" class="message-empty">{{ disabledText || t("error.sharedDirRequired") }}</div>
      <div v-else-if="!messages.length" class="message-empty">{{ emptyText || t("message.empty") }}</div>
      <template v-else>
        <div v-for="message in messages" :key="message.id" class="message-item" :class="{ 'message-item--own': isOwnMessage(message) }">
          <div class="message-meta">
            <span>{{ messageLabel(message) }}</span>
            <span>{{ formatDate(message.createdAt) }}</span>
          </div>
          <div class="message-text">{{ message.text }}</div>
        </div>
      </template>
    </div>

    <form class="message-form" @submit.prevent="emit('send')">
      <textarea
        v-model="draftModel"
        class="message-input"
        rows="1"
        maxlength="2000"
        :disabled="disabled || inputDisabled || loading"
        :placeholder="inputDisabled && inputDisabledText ? inputDisabledText : t('message.inputPlaceholder')"
      />
      <Button class="message-send" type="submit" :disabled="!canSend">
        <Send class="h-4 w-4" />{{ t("message.send") }}
      </Button>
    </form>
  </section>
</template>

<style scoped>
.message-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  min-height: 0;
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  max-height: min(52dvh, 420px);
  overflow-y: auto;
  padding-right: 2px;
}

.message-empty {
  display: grid;
  min-height: 90px;
  place-items: center;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
  text-align: center;
}

.message-item {
  align-self: flex-start;
  max-width: min(100%, 420px);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-muted);
  padding: 9px 11px;
}

.message-item--own {
  align-self: flex-end;
  background: var(--color-accent-subtle);
  border-color: transparent;
}

.message-meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.message-text {
  margin-top: 5px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  line-height: 1.45;
}

.message-form {
  display: flex;
  gap: var(--space-2);
  align-items: flex-end;
}

.message-input {
  height: 36px;
  min-height: 36px;
  flex: 1;
  resize: none;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-control);
  padding: 8px 11px;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  line-height: 18px;
  outline: none;
  overflow-y: auto;
}

.message-input:focus {
  border-color: var(--color-accent);
}

.message-input:disabled {
  opacity: 0.55;
}

.message-send {
  height: 36px;
  min-height: 36px;
  gap: 6px;
}
</style>
