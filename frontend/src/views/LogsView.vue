<script setup lang="ts">
import { ArrowLeft } from "lucide-vue-next";
import { useRouter } from "vue-router";
import { useAppStore } from "@/stores/app";
import { formatDate } from "@/lib/format";
import type { AccessLog } from "@/types/app";

const app = useAppStore();
const router = useRouter();

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
  <main class="logs-view">
    <div class="logs-header">
      <button class="back-button" title="返回" @click="router.push({ name: 'share' })">
        <ArrowLeft class="h-4 w-4" />
      </button>
      <p class="field-hint">{{ app.logs.length }} 条记录</p>
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
  </main>
</template>

<style scoped>
.logs-view {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: var(--space-3);
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.logs-view::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.logs-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: 0 0 var(--space-2);
}

.back-button {
  width: var(--icon-button-size);
  height: var(--icon-button-size);
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  background: var(--color-bg-control);
}

.back-button:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-hover);
}

.field-hint {
  margin: 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.empty-state {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  display: flex;
  min-height: 110px;
  align-items: center;
  justify-content: center;
  margin-top: var(--space-3);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}

.log-list {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  overflow: hidden;
}

.log-row {
  padding: 13px var(--space-3);
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
  margin-top: 6px;
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
  padding: 3px 8px;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.status-code--error {
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  color: var(--color-danger);
}
</style>
