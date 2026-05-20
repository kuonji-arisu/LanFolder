<script setup lang="ts">
import { ShieldCheck } from "lucide-vue-next";
import AppDialogContent from "@/components/app/AppDialogContent.vue";
import AppDialogHeader from "@/components/app/AppDialogHeader.vue";
import { Button } from "@/components/ui/button";
import { Dialog, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import { formatDate } from "@/lib/format";
import { useI18n } from "@/lib/i18n";
import { userAgentLabel } from "@/lib/userAgent";
import { useAppStore } from "@/stores/app";
import { useNoticeStore } from "@/stores/notices";

const open = defineModel<boolean>("open", { required: true });
const app = useAppStore();
const notices = useNoticeStore();
const { t } = useI18n();

async function approve(id: string) {
  notices.showTaskResult(await app.approveAccessRequest(id));
}

async function deny(id: string) {
  notices.showTaskResult(await app.denyAccessRequest(id));
}

async function revoke(id: string) {
  notices.showTaskResult(await app.revokeAccessSession(id));
}

function requestStats(request: { requestCount: number; lastSeenAt: string }) {
  const parts: string[] = [];
  if (request.requestCount > 1) parts.push(`${request.requestCount} ${t("access.times")}`, `${t("access.recent")} ${formatDate(request.lastSeenAt)}`);
  return parts;
}

function sessionStats(session: { createdAt: string; expiresAt?: string | null }) {
  const parts = [t("access.authorizedAt", { time: formatDate(session.createdAt) })];
  parts.push(session.expiresAt ? t("access.expiresAt", { time: formatDate(session.expiresAt) }) : t("access.noAutoExpire"));
  return parts.join(" · ");
}
</script>

<template>
  <Dialog v-model:open="open">
    <AppDialogContent width="calc(100vw - var(--space-6))" max-width="var(--content-max-width)" max-height="calc(100dvh - var(--space-6))">
      <AppDialogHeader>
        <DialogTitle>{{ t("access.manage") }}</DialogTitle>
        <DialogDescription>{{ t("access.description") }}</DialogDescription>
      </AppDialogHeader>

      <div class="access-dialog-scroll">
        <section class="access-section">
          <div class="section-label">{{ t("access.pending") }}</div>
          <div v-if="!app.pendingAccessRequests.length" class="empty-row">{{ t("access.emptyPending") }}</div>
          <div v-for="request in app.pendingAccessRequests" :key="request.id" class="access-row">
            <div class="access-copy">
              <div class="access-title access-ip">{{ t("access.from") }} {{ request.ip }}</div>
              <div class="access-agent" :title="request.userAgent || t('common.unknownBrowser')">{{ userAgentLabel(request.userAgent) }}</div>
              <div v-if="requestStats(request).length" class="access-meta">{{ requestStats(request).join(" · ") }}</div>
            </div>
            <div class="access-actions">
              <Button size="sm" variant="secondary" @click="deny(request.id)">{{ t("access.reject") }}</Button>
              <Button size="sm" @click="approve(request.id)">{{ t("access.approve") }}</Button>
            </div>
          </div>
        </section>

        <section class="access-section">
          <div class="section-label">{{ t("access.authorized") }}</div>
          <div v-if="!app.accessSessions.length" class="empty-row">{{ t("access.emptyAuthorized") }}</div>
          <div v-for="session in app.accessSessions" :key="session.id" class="access-row">
            <div class="access-copy">
              <div class="access-title session-title">
                <ShieldCheck class="session-icon h-4 w-4" />
                {{ session.ip }}
              </div>
              <div class="access-agent" :title="session.userAgent || t('common.unknownBrowser')">{{ userAgentLabel(session.userAgent) }}</div>
              <div class="access-meta">{{ sessionStats(session) }}</div>
            </div>
            <Button size="sm" variant="secondary" @click="revoke(session.id)">{{ t("access.revoke") }}</Button>
          </div>
        </section>
      </div>
    </AppDialogContent>
  </Dialog>
</template>

<style scoped>
.access-dialog-scroll {
  flex: 0 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;
}

.access-dialog-scroll::-webkit-scrollbar {
  display: none;
}

.access-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  min-width: 0;
}

.section-label {
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

.empty-row,
.access-row {
  min-height: 52px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
}

.empty-row {
  display: flex;
  align-items: center;
  padding: 0 var(--space-3);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}

.access-row {
  min-height: 76px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
}

.access-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.access-title {
  display: flex;
  align-items: center;
  min-width: 0;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0;
}

.access-ip {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.session-title {
  gap: var(--space-2);
  font-size: var(--font-size-sm);
}

.session-icon {
  flex-shrink: 0;
  color: var(--color-accent);
}

.access-agent,
.access-meta {
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.access-agent {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow-wrap: anywhere;
  line-height: 1.45;
}

.access-actions {
  display: flex;
  gap: var(--space-2);
  flex-shrink: 0;
  align-self: center;
}
</style>
