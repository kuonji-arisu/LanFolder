<script setup lang="ts">
import { ref } from "vue";
import { ArrowUp, Folder, MessageSquareText, Plus, RefreshCw, UploadCloud } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { TaskResult } from "@/composables/useAsyncTask";
import { useNoticeStore } from "@/stores/notices";
import { useWebFilesStore } from "@/stores/webFiles";

const files = useWebFilesStore();
const notices = useNoticeStore();
const uploadInput = ref<HTMLInputElement | null>(null);

defineProps<{
  messagesOpen?: boolean;
}>();

const emit = defineEmits<{
  (event: "toggleMessages"): void;
}>();

async function runWithNotice(action: () => Promise<TaskResult<unknown>>) {
  notices.showTaskResult(await action());
}

async function handleUpload(fileList: FileList | null) {
  const result = await files.uploadFiles(fileList);
  notices.showTaskResult(result);
  if (result.ok && result.value) notices.showSuccess("上传成功");
  if (uploadInput.value) uploadInput.value.value = "";
}

async function createFolder() {
  notices.showTaskResult(await files.createFolder());
}
</script>

<template>
  <header class="web-toolbar">
    <div class="toolbar-main">
      <div class="title-block">
        <div class="title-row">
          <span class="title-mark"><Folder class="h-4 w-4" /></span>
          <span>LanFolder</span>
        </div>
        <span class="permission-pill">{{ files.permissionLabel }}</span>
      </div>
      <button class="icon-button" aria-label="刷新" @click="runWithNotice(() => files.load())">
        <RefreshCw class="h-5 w-5" />
      </button>
      <button class="icon-button" :class="{ 'icon-button--active': messagesOpen }" aria-label="传递字符" @click="emit('toggleMessages')">
        <MessageSquareText class="h-5 w-5" />
      </button>
    </div>

    <nav class="path-bar" aria-label="当前路径">
      <button class="crumb" @click="runWithNotice(() => files.load(''))">根目录</button>
      <template v-for="(crumb, index) in files.crumbs" :key="`${crumb}-${index}`">
        <span class="slash">/</span>
        <button class="crumb crumb-nested" @click="runWithNotice(() => files.jumpToCrumb(index))">{{ crumb }}</button>
      </template>
    </nav>

    <div class="action-row">
      <Button variant="secondary" class="touch-button" :disabled="!files.listing?.path || files.loading" @click="runWithNotice(() => files.load(files.listing?.parentPath ?? ''))">
        <ArrowUp class="h-4 w-4" />上级
      </Button>
      <template v-if="files.canUpload">
        <input ref="uploadInput" type="file" multiple class="hidden" @change="handleUpload(($event.target as HTMLInputElement).files)" />
        <Button class="touch-button" :disabled="files.loading" @click="uploadInput?.click()">
          <UploadCloud class="h-4 w-4" />上传
        </Button>
      </template>
    </div>

    <form v-if="files.canUpload" class="mkdir-row" @submit.prevent="createFolder">
      <Input v-model="files.newFolderName" placeholder="新建文件夹名称" />
      <Button variant="secondary" class="touch-button" :disabled="!files.newFolderName.trim() || files.loading" type="submit">
        <Plus class="h-4 w-4" />新建
      </Button>
    </form>
  </header>
</template>

<style scoped>
.web-toolbar {
  --web-toolbar-padding-x: 14px;
  --web-toolbar-padding-x-wide: 22px;
  --web-toolbar-padding-top: 14px;
  --web-toolbar-padding-bottom: 12px;
  --brand-mark-size: 30px;
  --crumb-max-width: 150px;

  flex: 0 0 auto;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-bg-chrome);
  padding: max(var(--web-toolbar-padding-top), env(safe-area-inset-top)) var(--web-toolbar-padding-x) var(--web-toolbar-padding-bottom);
  z-index: 1;
}

.toolbar-main {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: var(--space-3);
  max-width: var(--content-max-width);
  margin: 0 auto;
}

.title-block {
  flex: 1;
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-lg);
  font-weight: 700;
}

.title-mark {
  width: var(--brand-mark-size);
  height: var(--brand-mark-size);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--color-accent);
  color: var(--color-text-on-accent);
}

.permission-pill {
  display: inline-flex;
  margin-top: 7px;
  border-radius: 999px;
  background: var(--color-bg-muted);
  padding: 3px var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.path-bar {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  max-width: var(--content-max-width);
  margin: 13px auto 0;
  overflow-x: auto;
  white-space: nowrap;
  -webkit-overflow-scrolling: touch;
}

.crumb {
  flex: 0 0 auto;
  border-radius: 999px;
  background: var(--color-bg-control);
  padding: 7px var(--space-3);
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
}

.crumb-nested {
  max-width: var(--crumb-max-width);
  overflow: hidden;
  text-overflow: ellipsis;
}

.slash {
  color: var(--color-text-secondary);
}

.action-row,
.mkdir-row {
  display: flex;
  gap: var(--space-2);
  max-width: var(--content-max-width);
  margin: 12px auto 0;
}

.mkdir-row :deep(input) {
  min-height: var(--touch-target);
}

.touch-button {
  min-height: var(--touch-target);
  flex: 1;
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

.icon-button--active {
  background: var(--color-accent-subtle);
  color: var(--color-accent);
}

@media (min-width: 760px) {
  .web-toolbar {
    padding-left: var(--web-toolbar-padding-x-wide);
    padding-right: var(--web-toolbar-padding-x-wide);
  }

  .touch-button {
    flex: 0 0 auto;
  }
}
</style>
