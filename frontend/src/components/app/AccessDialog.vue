<script setup lang="ts">
import { ShieldCheck } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { formatDate } from "@/lib/format";
import { useAppStore } from "@/stores/app";
import { useNoticeStore } from "@/stores/notices";

const open = defineModel<boolean>("open", { required: true });
const app = useAppStore();
const notices = useNoticeStore();

async function approve(id: string) {
  notices.showTaskResult(await app.approveAccessRequest(id));
}

async function deny(id: string) {
  notices.showTaskResult(await app.denyAccessRequest(id));
}

async function revoke(id: string) {
  notices.showTaskResult(await app.revokeAccessSession(id));
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="access-dialog">
      <div class="access-dialog-head">
        <DialogHeader>
          <DialogTitle>访问管理</DialogTitle>
          <DialogDescription>批准新浏览器，或撤销已经授权的浏览器 session。</DialogDescription>
        </DialogHeader>
      </div>

      <section class="access-section">
        <div class="section-label">待批准</div>
        <div v-if="!app.pendingAccessRequests.length" class="empty-row">暂无新设备请求</div>
        <div v-for="request in app.pendingAccessRequests" :key="request.id" class="access-row">
          <div class="access-copy">
            <div class="access-title">{{ request.code }}</div>
            <div class="access-meta">{{ request.ip }} · {{ request.userAgent || "未知浏览器" }}</div>
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
              <ShieldCheck class="h-4 w-4" />
              {{ session.ip }}
            </div>
            <div class="access-meta">{{ session.userAgent || "未知浏览器" }} · {{ formatDate(session.createdAt) }}</div>
          </div>
          <Button size="sm" variant="secondary" @click="revoke(session.id)">撤销</Button>
        </div>
      </section>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.access-dialog {
  max-width: 560px;
}

.access-dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding-right: 44px;
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
}

.access-copy {
  min-width: 0;
}

.access-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0;
}

.session-title {
  font-size: var(--font-size-sm);
}

.access-meta {
  max-width: 300px;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.access-actions {
  display: flex;
  gap: var(--space-2);
  flex-shrink: 0;
}
</style>
