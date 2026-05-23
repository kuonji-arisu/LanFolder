<script setup lang="ts">
import type { Permission } from "@/lib/constants";
import type { PermissionOption } from "@/types/app";

defineProps<{
  modelValue: Permission;
  options: PermissionOption[];
}>();

defineEmits<{
  (event: "update:modelValue", value: Permission): void;
}>();
</script>

<template>
  <div class="segmented">
    <button
      v-for="item in options"
      :key="item.value"
      class="segment-option"
      :class="[`segment-option--${item.value}`, { 'segment-option--active': modelValue === item.value }]"
      @click="$emit('update:modelValue', item.value)"
    >
      {{ item.label }}
    </button>
  </div>
</template>

<style scoped>
.segmented {
  display: flex;
  overflow: hidden;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.segment-option {
  --segment-active-bg: var(--color-accent);
  --segment-active-bg-hover: var(--color-accent-hover);
  --segment-active-fg: var(--color-text-on-accent);

  flex: 1;
  min-height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  background: transparent;
  font-size: var(--font-size-sm);
  transition: background 0.15s, color 0.15s;
}

.segment-option--manage {
  --segment-active-bg: var(--color-danger);
  --segment-active-bg-hover: var(--color-danger-hover);
  --segment-active-fg: var(--color-text-on-danger);
}

.segment-option + .segment-option {
  border-left: 1px solid var(--color-border);
}

.segment-option:hover {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.segment-option--active {
  background: var(--segment-active-bg);
  color: var(--segment-active-fg);
  font-weight: var(--font-weight-medium);
}

.segment-option--active:hover {
  background: var(--segment-active-bg-hover);
  color: var(--segment-active-fg);
}
</style>
