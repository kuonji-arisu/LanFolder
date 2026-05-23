<script setup lang="ts">
import { Card } from "@/components/ui/card";

withDefaults(
  defineProps<{
    label: string;
    value?: string;
    valueTitle?: string;
    mono?: boolean;
    variant?: "field" | "panel";
  }>(),
  {
    value: undefined,
    valueTitle: undefined,
    mono: false,
    variant: "panel",
  },
);
</script>

<template>
  <Card class="info-card" :class="`info-card--${variant}`">
    <div class="info-card-header">
      <div class="info-card-label">{{ label }}</div>
      <div v-if="$slots.actions" class="info-card-actions">
        <slot name="actions" />
      </div>
    </div>

    <div class="info-card-body">
      <div
        v-if="value !== undefined"
        class="info-card-value"
        :class="{ 'info-card-value--mono': mono }"
        :title="valueTitle || value"
      >
        {{ value }}
      </div>
      <slot />
    </div>
  </Card>
</template>

<style scoped>
.info-card {
  --info-card-padding-block: var(--layout-card-padding);
  --info-card-padding-inline: var(--layout-card-padding);
  --info-card-field-height: var(--layout-field-card-height);
  --info-card-action-icon-size: 16px;
  --info-card-label-line-height: 18px;
  --info-card-action-inline-offset: calc((var(--icon-button-size) - var(--info-card-action-icon-size)) / 2);
  --info-card-action-block-offset: calc((var(--icon-button-size) - var(--info-card-label-line-height)) / 2);

  display: grid;
  min-width: 0;
  gap: var(--space-2);
  padding: var(--info-card-padding-block) var(--info-card-padding-inline);
  overflow: hidden;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.info-card--field {
  height: var(--info-card-field-height);
  min-height: var(--info-card-field-height);
  grid-template-rows: var(--icon-button-size) minmax(0, 1fr);
  gap: 0;
}

.info-card--panel {
  --info-card-padding-block: var(--layout-panel-card-padding-block);
  --info-card-label-line-height: var(--font-size-sm);
}

.info-card-header {
  display: grid;
  min-width: 0;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: flex-start;
  gap: var(--space-2);
}

.info-card-label {
  min-width: 0;
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  line-height: var(--info-card-label-line-height);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-card-actions {
  display: inline-flex;
  min-width: 0;
  align-items: flex-start;
  align-self: flex-start;
  justify-content: flex-end;
  gap: var(--space-2);
}

.info-card--field .info-card-actions {
  margin-block-start: calc(var(--info-card-action-block-offset) * -1);
  margin-inline-end: calc(var(--info-card-action-inline-offset) * -1);
}

.info-card-body {
  min-width: 0;
  min-height: 0;
}

.info-card--field .info-card-body {
  display: flex;
  align-items: flex-end;
}

.info-card-value {
  width: 100%;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  color: var(--color-text-secondary);
  font-size: var(--font-size-base);
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-card-value--mono {
  font-family: "Cascadia Mono", "SFMono-Regular", Consolas, monospace;
}
</style>
