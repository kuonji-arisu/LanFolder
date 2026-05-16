<script setup lang="ts">
import { Download, File, Folder, Loader2, Trash2 } from "lucide-vue-next";
import { Card } from "@/components/ui/card";
import { fileApi } from "@/lib/api";
import { formatBytes, formatDate } from "@/lib/format";
import { useNoticeStore } from "@/stores/notices";
import { useWebFilesStore } from "@/stores/webFiles";

const files = useWebFilesStore();
const notices = useNoticeStore();

async function openEntry(entry: Parameters<typeof files.openEntry>[0]) {
  notices.showTaskResult(await files.openEntry(entry));
}

async function deleteEntry(entry: Parameters<typeof files.deleteEntry>[0]) {
  const result = await files.deleteEntry(entry);
  notices.showTaskResult(result);
  if (result.ok && result.value) notices.showSuccess("删除成功");
}
</script>

<template>
  <main class="web-content">
    <Card class="file-panel">
      <div v-if="files.loading" class="state-view">
        <Loader2 class="h-5 w-5 animate-spin" />
        <span>加载中</span>
      </div>

      <div v-else-if="files.error && !files.listing" class="state-view">
        <Folder class="h-8 w-8 text-muted-foreground" />
        <span>{{ files.error || "加载失败" }}</span>
      </div>

      <div v-else-if="files.listing && !files.listing.entries.length" class="state-view">
        <Folder class="h-8 w-8 text-muted-foreground" />
        <span>当前文件夹为空</span>
      </div>

      <div v-else class="file-list">
        <div v-for="entry in files.listing?.entries ?? []" :key="entry.path" class="file-row">
          <button class="file-main" @click="openEntry(entry)">
            <span class="file-icon">
              <Folder v-if="entry.isDir" class="h-5 w-5 text-primary" />
              <File v-else class="h-5 w-5 text-muted-foreground" />
            </span>
            <span class="file-copy">
              <span class="file-name">{{ entry.name }}</span>
              <span class="file-meta">{{ entry.isDir ? "文件夹" : formatBytes(entry.size) }} · {{ formatDate(entry.modTime) }}</span>
            </span>
          </button>

          <div class="file-actions">
            <a v-if="!entry.isDir" :href="fileApi.downloadUrl(entry.path)" class="icon-button" aria-label="下载">
              <Download class="h-5 w-5" />
            </a>
            <button v-if="files.canDelete" class="icon-button danger-button" aria-label="删除" @click="deleteEntry(entry)">
              <Trash2 class="h-5 w-5" />
            </button>
          </div>
        </div>
      </div>
    </Card>
  </main>
</template>

<style scoped>
.web-content {
  --web-content-padding-x: 14px;
  --web-content-padding-x-wide: 22px;
  --web-content-padding-top: 16px;

  flex: 1 1 auto;
  width: 100%;
  min-height: 0;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  max-width: var(--content-max-width);
  margin: 0 auto;
  padding: var(--web-content-padding-top) var(--web-content-padding-x) calc(var(--space-safe-bottom) + env(safe-area-inset-bottom));
}

.file-panel {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-elevated);
  overflow: hidden;
}

.state-view {
  display: grid;
  min-height: 220px;
  place-items: center;
  align-content: center;
  gap: var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-md);
}

.file-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  min-height: 68px;
  border-bottom: 1px solid var(--color-border);
  padding: 10px 10px 10px var(--space-3);
}

.file-row:last-child {
  border-bottom: 0;
}

.file-main {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  gap: var(--space-3);
  text-align: left;
}

.file-icon {
  width: var(--touch-target);
  height: var(--touch-target);
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--color-bg-muted);
}

.file-copy {
  min-width: 0;
}

.file-name,
.file-meta {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-name {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
}

.file-meta {
  margin-top: 4px;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.file-actions {
  display: flex;
  flex: 0 0 auto;
  gap: 6px;
}

.icon-button {
  display: inline-flex;
  width: var(--touch-target);
  height: var(--touch-target);
  align-items: center;
  justify-content: center;
  border-radius: 9px;
  background: var(--color-bg-control);
  color: var(--color-text-primary);
}

.danger-button {
  color: var(--color-danger);
}

@media (min-width: 760px) {
  .web-content {
    padding-left: var(--web-content-padding-x-wide);
    padding-right: var(--web-content-padding-x-wide);
  }
}
</style>
