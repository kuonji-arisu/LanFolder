<script setup lang="ts">
import { ArrowLeft } from "lucide-vue-next";
import { useRouter } from "vue-router";
import { useAppStore } from "@/stores/app";
import { formatDate } from "@/lib/format";

const app = useAppStore();
const router = useRouter();
</script>

<template>
  <main class="logs-view">
    <div class="logs-header">
      <button class="back-button" title="返回" @click="router.push({ name: 'share' })">
        <ArrowLeft class="h-4 w-4" />
      </button>
      <div>
        <div class="field-label">访问日志</div>
        <p class="field-hint">{{ app.logs.length }} 条记录</p>
      </div>
    </div>

    <div v-if="!app.logs.length" class="empty-state">暂无访问记录</div>

    <div v-else class="log-list">
      <div v-for="log in app.logs" :key="`${log.time}-${log.remote}-${log.path}`" class="log-row">
        <div class="log-main">
          <span class="log-path">{{ log.method }} {{ log.path }}</span>
          <span :class="['status-code', log.status >= 400 && 'status-code--error']">{{ log.status }}</span>
        </div>
        <div class="log-meta">
          <span class="mono">{{ log.remote }}</span>
          <span>{{ formatDate(log.time) }}</span>
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

.logs-header,
.empty-state,
.log-list {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.logs-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
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

.field-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.field-hint {
  margin: var(--space-1) 0 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.empty-state {
  display: flex;
  min-height: 110px;
  align-items: center;
  justify-content: center;
  margin-top: var(--space-3);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}

.log-list {
  margin-top: var(--space-3);
  overflow: hidden;
}

.log-row {
  padding: var(--space-3);
  border-bottom: 1px solid var(--color-border);
}

.log-row:last-child {
  border-bottom: 0;
}

.log-main,
.log-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.log-path {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.log-meta {
  margin-top: var(--space-2);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.status-code {
  flex-shrink: 0;
  border-radius: 999px;
  padding: 3px 8px;
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
  color: color-mix(in srgb, var(--color-success) 78%, var(--color-text-primary));
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.status-code--error {
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  color: var(--color-danger);
}

.mono {
  font-family: "Cascadia Mono", "SFMono-Regular", Consolas, monospace;
}
</style>
