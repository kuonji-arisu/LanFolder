<script setup lang="ts">
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { formatDate } from "@/lib/format";
import { useAppStore } from "@/stores/app";
import type { AccessLog } from "@/types/app";

const open = defineModel<boolean>("open", { required: true });
const app = useAppStore();
const dialogContentClass =
  "w-[calc(100vw_-_var(--space-6))] max-w-[var(--content-max-width)] h-[calc(100dvh_-_var(--space-6))] flex flex-col overflow-hidden";

function logTitle(log: AccessLog) {
  if (!log.target) return log.action || "访问共享";
  if (log.action === "上传") return `上传到 ${log.target}`;
  return `${log.action || "访问共享"} ${log.target}`;
}

function logTitleHint(log: AccessLog) {
  if (!log.targetPath) return logTitle(log);
  if (log.action === "上传") return `上传到 ${log.targetPath}`;
  return `${log.action || "访问共享"} ${log.targetPath}`;
}

function remoteLabel(log: AccessLog) {
  const parts = [log.remote];
  if (log.detail) parts.push(log.detail);
  return parts.filter(Boolean).join(" · ");
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="dialogContentClass">
      <div class="log-dialog-body">
        <div class="log-dialog-head">
          <DialogHeader>
            <DialogTitle>访问日志</DialogTitle>
            <DialogDescription>{{ app.logs.length }} 条最近访问记录</DialogDescription>
          </DialogHeader>
        </div>

        <div v-if="!app.logs.length" class="empty-state">暂无访问记录</div>

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
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.log-dialog-body {
  --log-dialog-head-padding: calc(var(--icon-button-size) + var(--space-2));

  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.log-dialog-head {
  padding-right: var(--log-dialog-head-padding);
}

.empty-state {
  flex: 1;
  min-height: 0;
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
  flex: 1;
  min-height: 0;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-color: var(--color-border) transparent;
  scrollbar-width: thin;
}

.log-list::-webkit-scrollbar {
  width: var(--space-2);
  height: 0;
}

.log-list::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: var(--color-border);
}

.log-row {
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
