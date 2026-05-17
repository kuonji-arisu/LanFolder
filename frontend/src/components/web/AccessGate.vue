<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { Loader2, ShieldCheck } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { accessApi, type AccessPollResult, type AccessStatus } from "@/lib/api";
import { errorMessage } from "@/lib/errors";

const emit = defineEmits<{
  authorized: [];
}>();

const accessStatus = ref<AccessStatus | null>(null);
const hasRequestedAccess = ref(false);
const accessState = ref<AccessPollResult["state"] | "idle">("idle");
const accessError = ref("");
let pollTimer: number | undefined;

const needsAccess = computed(() => accessStatus.value?.required && !accessStatus.value.authorized);

onMounted(() => {
  void loadAccess();
});

onBeforeUnmount(stopPolling);

async function loadAccess() {
  accessError.value = "";
  try {
    const status = await accessApi.status();
    accessStatus.value = status;
    if (!status.required || status.authorized) {
      stopPolling();
      emit("authorized");
    }
  } catch (err) {
    accessError.value = errorMessage(err);
  }
}

async function requestAccess() {
  accessError.value = "";
  accessState.value = "pending";
  try {
    await accessApi.request();
    hasRequestedAccess.value = true;
    startPolling();
  } catch (err) {
    accessState.value = "idle";
    accessError.value = errorMessage(err);
  }
}

function startPolling() {
  stopPolling();
  pollTimer = window.setInterval(() => void pollAccess(), 1500);
  void pollAccess();
}

function stopPolling() {
  if (pollTimer !== undefined) {
    window.clearInterval(pollTimer);
    pollTimer = undefined;
  }
}

async function pollAccess() {
  try {
    const result = await accessApi.poll();
    accessState.value = result.state;
    if (result.state === "approved") {
      stopPolling();
      accessStatus.value = { required: true, authorized: true };
      emit("authorized");
    } else if (result.state === "denied" || result.state === "expired") {
      stopPolling();
    }
  } catch (err) {
    stopPolling();
    accessError.value = errorMessage(err);
  }
}
</script>

<template>
  <main v-if="needsAccess || accessError" class="access-page">
    <Card class="access-panel">
      <div class="access-icon">
        <ShieldCheck class="h-6 w-6" />
      </div>
      <div>
        <h1 class="access-title">请求访问 LanFolder</h1>
        <p class="access-copy">这台电脑批准后，当前浏览器才能访问共享内容。</p>
      </div>

      <p v-if="accessState === 'pending'" class="access-status">
        <Loader2 class="h-4 w-4 animate-spin" />
        等待电脑端批准
      </p>
      <p v-else-if="accessState === 'denied'" class="access-status">请求已被拒绝</p>
      <p v-else-if="accessState === 'expired'" class="access-status">请求已过期，请重新发起</p>
      <p v-else-if="accessError" class="access-status access-status--error">{{ accessError }}</p>

      <Button class="access-button" :disabled="accessState === 'pending'" @click="requestAccess">
        {{ hasRequestedAccess && accessState !== "pending" ? "重新请求" : "请求访问" }}
      </Button>
    </Card>
  </main>
</template>

<style scoped>
.access-page {
  flex: 1;
  display: grid;
  place-items: center;
  padding: var(--space-5);
}

.access-panel {
  width: min(100%, 360px);
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: var(--space-4);
  padding: var(--space-5);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-elevated);
}

.access-icon {
  width: 48px;
  height: 48px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  color: var(--color-accent);
  background: var(--color-accent-subtle);
}

.access-title {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.access-copy,
.access-status {
  margin: 6px 0 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.access-status {
  min-height: 22px;
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

.access-status--error {
  color: var(--color-danger);
}

.access-button {
  width: 100%;
}
</style>
