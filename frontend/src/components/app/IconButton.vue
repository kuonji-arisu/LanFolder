<script setup lang="ts">
defineProps<{
  active?: boolean;
  accent?: boolean;
}>();
</script>

<template>
  <button class="icon-button" :class="{ 'icon-button--active': active, 'icon-button--accent': accent }">
    <slot />
  </button>
</template>

<style scoped>
.icon-button {
  width: var(--icon-button-size);
  height: var(--icon-button-size);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  transition: background 0.15s, color 0.15s;
}

.icon-button:hover {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.icon-button:disabled:hover {
  background: transparent;
  color: var(--color-text-secondary);
}

.icon-button--accent {
  color: var(--color-accent);
}

.icon-button--accent:hover {
  background: var(--color-accent-subtle);
  color: var(--color-accent);
}
</style>
