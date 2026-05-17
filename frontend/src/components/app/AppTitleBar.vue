<script setup lang="ts">
import { computed } from "vue";
import { Window } from "@wailsio/runtime";
import { Settings, Wifi, X } from "lucide-vue-next";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "@/lib/i18n";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const title = computed(() => (route.meta.titleKey ? t(String(route.meta.titleKey)) : String(route.meta.title ?? "LanFolder")));
</script>

<template>
  <header class="titlebar">
    <span class="titlebar-title">{{ title }}</span>

    <div class="titlebar-actions">
      <button class="titlebar-btn" :class="{ 'titlebar-btn--active': route.name === 'share' }" :title="t('app.share')" @click="router.push({ name: 'share' })">
        <Wifi class="h-[15px] w-[15px]" />
      </button>
      <button class="titlebar-btn" :class="{ 'titlebar-btn--active': route.name === 'settings' }" :title="t('app.settings')" @click="router.push({ name: 'settings' })">
        <Settings class="h-[15px] w-[15px]" />
      </button>
      <button class="titlebar-btn titlebar-btn--close" :title="t('app.close')" @click="Window.Close()">
        <X class="h-[14px] w-[14px]" />
      </button>
    </div>
  </header>
</template>

<style scoped>
.titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--titlebar-height);
  padding: 0 var(--space-3);
  background: var(--color-bg-titlebar);
  border-bottom: 1px solid var(--color-border);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  user-select: none;
  flex-shrink: 0;
  --wails-draggable: drag;
}

.titlebar-title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  pointer-events: none;
}

.titlebar-actions {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  --wails-draggable: no-drag;
}

.titlebar-btn {
  width: var(--titlebar-button-size);
  height: var(--titlebar-button-size);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  color: var(--color-text-tertiary);
  transition: background 0.15s, color 0.15s;
}

.titlebar-btn:hover {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.titlebar-btn--active {
  background: var(--color-accent);
  color: var(--color-text-on-accent);
}

.titlebar-btn--active:hover {
  background: var(--color-accent-hover);
  color: var(--color-text-on-accent);
}

.titlebar-btn--close:hover {
  background: var(--color-danger);
  color: var(--color-text-on-accent);
}
</style>
