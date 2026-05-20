<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { Check, Copy, FolderOpen, FolderPen, HardDrive, MessageSquareText, Play, ShieldCheck, Square } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import AccessLogDialog from "@/components/app/AccessLogDialog.vue";
import AccessLogPanel from "@/components/app/AccessLogPanel.vue";
import AccessDialog from "@/components/app/AccessDialog.vue";
import FieldCard from "@/components/app/FieldCard.vue";
import IconButton from "@/components/app/IconButton.vue";
import MessageDialog from "@/components/messages/MessageDialog.vue";
import PermissionSegment from "@/components/app/PermissionSegment.vue";
import { useClipboard } from "@/composables/useClipboard";
import { useToast } from "@/composables/useToast";
import { useI18n } from "@/lib/i18n";
import { useAppStore } from "@/stores/app";
import { useDesktopMessagesStore } from "@/stores/desktopMessages";
import { useNoticeStore } from "@/stores/notices";
import type { Permission } from "@/lib/constants";
import type { TaskResult } from "@/composables/useAsyncTask";

const app = useAppStore();
const messages = useDesktopMessagesStore();
const notices = useNoticeStore();
const toasts = useToast();
const { t } = useI18n();
const { copied, copy } = useClipboard();
const accessDialogOpen = ref(false);
const accessLogDialogOpen = ref(false);
const messagesOpen = ref(false);
const accessCount = computed(() => app.pendingAccessRequests.length + app.accessSessions.length);
const messagesEnabled = computed(() => app.isRunning && Boolean(app.config.sharedDir));

watch(messagesOpen, (open) => {
  if (open) void messages.load();
});

async function runWithNotice(action: () => Promise<TaskResult<unknown>>) {
  notices.showTaskResult(await action());
}

async function setPermission(permission: Permission) {
  if (permission === "manage") {
    toasts.warning(t("share.manageWarning"));
  }
  await runWithNotice(() => app.setPermission(permission));
}
</script>

<template>
  <main class="view-body">
    <Card class="share-hero" :class="{ 'share-hero--running': app.isRunning }">
      <div class="hero-left">
        <span class="hero-icon">
          <HardDrive class="h-5 w-5" />
        </span>
        <div class="hero-copy">
          <div class="hero-title">{{ app.isRunning ? t("share.running") : t("share.stopped") }}</div>
          <p class="hero-hint">{{ t("share.accessPermission") }}: {{ app.activePermission.label }}</p>
        </div>
      </div>

      <Button class="hero-action" size="sm" :variant="app.isRunning ? 'destructive' : 'default'" :disabled="app.busy" @click="runWithNotice(app.toggleSharing)">
        <Square v-if="app.isRunning" class="h-4 w-4" />
        <Play v-else class="h-4 w-4" />
        {{ app.isRunning ? t("share.stop") : t("share.start") }}
      </Button>
    </Card>

    <FieldCard :label="t('share.accessAddress')" :value="app.primaryAddress" mono>
      <span v-if="app.config.accessApproval" class="access-button-wrap">
        <IconButton :title="t('access.manage')" :accent="Boolean(app.pendingAccessRequests.length)" @click="accessDialogOpen = true">
          <ShieldCheck class="h-4 w-4" />
        </IconButton>
        <span v-if="accessCount" class="access-badge">{{ accessCount }}</span>
      </span>
      <IconButton :title="t('share.copyAddress')" accent @click="copy(app.primaryAddress)">
        <Check v-if="copied" class="h-4 w-4" />
        <Copy v-else class="h-4 w-4" />
      </IconButton>
    </FieldCard>

    <FieldCard :label="t('share.directory')" :value="app.config.sharedDir || t('share.noFolder')" :value-title="app.config.sharedDir">
      <IconButton v-if="!app.isRunning" :title="t('share.changeFolder')" @click="runWithNotice(app.chooseFolder)">
        <FolderPen class="h-4 w-4" />
      </IconButton>
      <IconButton v-else :title="t('message.title')" :disabled="!messagesEnabled" :accent="messagesOpen" @click="messagesOpen = true">
        <MessageSquareText class="h-4 w-4" />
      </IconButton>
      <IconButton :title="t('common.open')" :disabled="!app.config.sharedDir" @click="runWithNotice(app.openSharedFolder)">
        <FolderOpen class="h-4 w-4" />
      </IconButton>
    </FieldCard>

    <Card class="panel">
      <div class="field-label">{{ t("share.accessPermission") }}</div>
      <PermissionSegment :model-value="app.config.permission" :options="app.state?.permissions ?? []" @update:model-value="setPermission" />
      <p class="field-hint">{{ app.activePermission.description }}</p>
    </Card>

    <AccessLogPanel :logs="app.logs" @open="accessLogDialogOpen = true" />
    <AccessLogDialog v-model:open="accessLogDialogOpen" />
    <AccessDialog v-model:open="accessDialogOpen" />
    <MessageDialog
      v-model:open="messagesOpen"
      v-model:draft="messages.draft"
      :messages="messages.messages"
      :current-client-id="messages.clientId"
      :loading="messages.loading"
      :can-send="messagesEnabled"
      :can-clear="messagesEnabled"
      :disabled="!messagesEnabled"
      :disabled-text="t('error.shareNotRunning')"
      layout="bounded"
      :load-messages="messages.load"
      :send-message="messages.send"
      :clear-messages="messages.clear"
    />
  </main>
</template>

<style scoped>
.view-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  overflow-y: auto;
  padding: var(--space-3);
}

.view-body::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.share-hero,
.panel {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.share-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3);
  transition: border-color 0.15s, box-shadow 0.15s;
}

.share-hero--running {
  border-color: var(--color-accent);
  box-shadow: var(--shadow-sm);
}

.hero-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
  flex: 1;
}

.hero-icon {
  width: 52px;
  height: 52px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  color: var(--color-accent);
  background: var(--color-accent-subtle);
}

.hero-copy {
  min-width: 0;
  flex: 1;
}

.hero-action {
  min-width: 76px;
  height: 34px;
  flex-shrink: 0;
  gap: 6px;
  padding: 0 13px;
}

.hero-title {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.hero-hint,
.field-hint {
  margin: 3px 0 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
}

.panel {
  padding: var(--space-3);
}

.panel :deep(.segmented) {
  margin-top: var(--space-3);
}

.field-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.access-button-wrap {
  position: relative;
  display: inline-flex;
}

.access-badge {
  position: absolute;
  top: -5px;
  right: -5px;
  min-width: 16px;
  height: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  border-radius: 999px;
  background: var(--color-accent);
  color: var(--color-text-on-accent);
  font-size: 10px;
  line-height: 1;
}

</style>
