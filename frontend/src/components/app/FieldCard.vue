<script setup lang="ts">
import { Card } from "@/components/ui/card";

defineProps<{
  label: string;
  value: string;
  mono?: boolean;
}>();
</script>

<template>
  <Card class="field-card">
    <div class="field-main">
      <div class="field-label">{{ label }}</div>
      <div class="field-value" :class="{ mono }">{{ value }}</div>
    </div>
    <div class="field-actions">
      <slot />
    </div>
  </Card>
</template>

<style scoped>
.field-card {
  min-height: 80px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3);
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.field-main {
  min-width: 0;
}

.field-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.field-value {
  max-width: 350px;
  margin-top: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-secondary);
  font-size: var(--font-size-base);
}

.field-actions {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
}

.mono {
  font-family: "Cascadia Mono", "SFMono-Regular", Consolas, monospace;
}
</style>
