<script setup lang="ts">
import { onMounted } from "vue";
import { RouterView } from "vue-router";
import AppTitleBar from "@/components/app/AppTitleBar.vue";
import { Toaster } from "@/components/ui/sonner";
import { useTheme } from "@/composables/useTheme";
import { useAppStore } from "@/stores/app";
import { useNoticeStore } from "@/stores/notices";

const app = useAppStore();
const notices = useNoticeStore();
const { initTheme } = useTheme();

onMounted(async () => {
  initTheme();
  await notices.startListening();
  await app.loadSnapshot();
  await notices.drain();
  app.startAutoRefresh();
});
</script>

<template>
  <div class="app-shell">
    <AppTitleBar />
    <RouterView />
    <Toaster position="top-right" />
  </div>
</template>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
  overflow: hidden;
  background: var(--color-bg-base);
  color: var(--color-text-primary);
}
</style>
