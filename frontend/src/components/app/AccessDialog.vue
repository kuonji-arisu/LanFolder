<script setup lang="ts">
import { ShieldCheck } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { formatDate } from "@/lib/format";
import { userAgentLabel } from "@/lib/userAgent";
import { useAppStore } from "@/stores/app";
import { useNoticeStore } from "@/stores/notices";

const open = defineModel<boolean>("open", { required: true });
const app = useAppStore();
const notices = useNoticeStore();
const dialogContentClass =
  "w-[calc(100vw_-_var(--space-6))] max-w-[var(--content-max-width)] max-h-[calc(100dvh_-_var(--space-6))] flex flex-col overflow-hidden";

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
  if (request.requestCount > 1) parts.push(`${request.requestCount} 次`, `最近 ${formatDate(request.lastSeenAt)}`);
  return parts;
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="dialogContentClass">
      <div class="access-dialog-head">
        <DialogHeader>
          <DialogTitle>访问管理</DialogTitle>
          <DialogDescription>批准新浏览器，或撤销已经授权的浏览器 session。</DialogDescription>
        </DialogHeader>
      </div>

      <div class="access-dialog-scroll">
        <section class="access-section">
          <div class="section-label">待批准</div>
          <div v-if="!app.pendingAccessRequests.length" class="empty-row">暂无新设备请求</div>
          <div v-for="request in app.pendingAccessRequests" :key="request.id" class="access-row">
            <div class="access-copy">
              <div class="access-title access-ip">来自 {{ request.ip }}</div>
              <div class="access-agent" :title="request.userAgent || '未知浏览器'">{{ userAgentLabel(request.userAgent) }}</div>
              <div v-if="requestStats(request).length" class="access-meta">{{ requestStats(request).join(" · ") }}</div>
            </div>
            <div class="access-actions">
              <Button size="sm" variant="secondary" @click="deny(request.id)">拒绝</Button>
              <Button size="sm" @click="approve(request.id)">允许</Button>
            </div>
          </div>
        </section>

        <section class="access-section">
          <div class="section-label">已授权</div>
          <div v-if="!app.accessSessions.length" class="empty-row">暂无已授权浏览器</div>
          <div v-for="session in app.accessSessions" :key="session.id" class="access-row">
            <div class="access-copy">
              <div class="access-title session-title">
                <ShieldCheck class="session-icon h-4 w-4" />
                {{ session.ip }}
              </div>
              <div class="access-agent" :title="session.userAgent || '未知浏览器'">{{ userAgentLabel(session.userAgent) }}</div>
              <div class="access-meta">{{ formatDate(session.createdAt) }}</div>
            </div>
            <Button size="sm" variant="secondary" @click="revoke(session.id)">撤销</Button>
          </div>
        </section>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.access-dialog-head {
  padding-right: calc(var(--icon-button-size) + var(--space-2));
}

.access-dialog-scroll {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  overflow-y: auto;
  overscroll-behavior: contain;
}

.access-dialog-scroll::-webkit-scrollbar {
  width: 0;
  height: 0;
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
