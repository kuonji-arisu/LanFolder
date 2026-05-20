<script setup lang="ts">
import AppDialogContent from "@/components/app/AppDialogContent.vue";
import AppDialogHeader from "@/components/app/AppDialogHeader.vue";
import { Dialog, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import { formatDate } from "@/lib/format";
import { useI18n } from "@/lib/i18n";
import { useAppStore } from "@/stores/app";
import type { AccessLog } from "@/types/app";

const open = defineModel<boolean>("open", { required: true });
const app = useAppStore();
const { t } = useI18n();

function logTitle(log: AccessLog) {
  const action = log.action || t("log.action.shareAccess");
  if (!log.target) return action;
  return t("log.actionTarget", { action, target: log.target });
}

function logTitleHint(log: AccessLog) {
  if (!log.targetPath) return logTitle(log);
  return t("log.actionTarget", { action: log.action || t("log.action.shareAccess"), target: log.targetPath });
}

function remoteLabel(log: AccessLog) {
  const parts = [log.remote];
  if (log.detail) parts.push(log.detail);
  return parts.filter(Boolean).join(" · ");
}

function logDescription(count: number) {
  return t("log.description", { count });
}
</script>

<template>
  <Dialog v-model:open="open">
    <AppDialogContent width="calc(100vw - var(--space-6))" max-width="var(--content-max-width)" max-height="calc(100dvh - var(--space-6))">
      <AppDialogHeader>
        <DialogTitle>{{ t("log.title") }}</DialogTitle>
        <DialogDescription>{{ logDescription(app.logs.length) }}</DialogDescription>
      </AppDialogHeader>

      <div v-if="!app.logs.length" class="empty-state">{{ t("log.empty") }}</div>

      <div v-else class="log-list">
        <div v-for="log in app.logs" :key="`${log.time}-${log.remote}-${log.path}`" class="log-row">
          <div class="log-main">
            <span class="log-title" :title="logTitleHint(log)">{{ logTitle(log) }}</span>
            <span v-if="log.status >= 400" class="status-code status-code--error">{{ log.status }}</span>
          </div>
          <div class="log-meta">
            <span class="log-remote">{{ remoteLabel(log) }}</span>
            <span class="log-time">{{ formatDate(log.time) }}</span>
          </div>
        </div>
      </div>
    </AppDialogContent>
  </Dialog>
</template>

<style scoped>
.empty-state,
.log-list {
  --log-entry-min-height: 64px;
}

.empty-state {
  min-height: var(--log-entry-min-height);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}

.log-list {
  flex: 0 1 auto;
  min-height: var(--log-entry-min-height);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;
}

.log-list::-webkit-scrollbar {
  display: none;
}

.log-row {
  min-height: var(--log-entry-min-height);
  padding: var(--space-3);
  border-bottom: 1px solid var(--color-border);
}

.log-row:last-child {
  border-bottom: 0;
}

.log-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.log-title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.log-meta {
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  margin-top: var(--space-1);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.log-remote {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.log-time {
  flex-shrink: 0;
}

.status-code {
  flex-shrink: 0;
  border-radius: 999px;
  padding: var(--space-1) var(--space-2);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.status-code--error {
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  color: var(--color-danger);
}
</style>
