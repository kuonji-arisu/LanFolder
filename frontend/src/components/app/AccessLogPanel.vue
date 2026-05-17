<script setup lang="ts">
import { computed } from "vue";
import { ArrowRight } from "lucide-vue-next";
import { Card } from "@/components/ui/card";
import { useI18n } from "@/lib/i18n";
import type { AccessLog } from "@/types/app";

const props = defineProps<{
  logs: AccessLog[];
}>();
const emit = defineEmits<{
  open: [];
}>();
const { t } = useI18n();

const totalCount = computed(() => props.logs.length);
const errorCount = computed(() => props.logs.filter((log) => log.status >= 400).length);
</script>

<template>
  <Card class="logs-summary">
    <div class="panel-title-row">
      <span class="field-label">{{ t("log.recent") }}</span>
      <button class="view-link" type="button" @click="emit('open')">
        {{ t("log.view") }}
        <ArrowRight class="h-[13px] w-[13px]" />
      </button>
    </div>

    <div class="summary-grid">
      <div class="summary-item">
        <span class="summary-value">{{ totalCount }}</span>
        <span class="summary-label">{{ t("log.times") }}</span>
      </div>
      <div class="summary-item">
        <span class="summary-value" :class="{ 'summary-value--danger': errorCount > 0 }">{{ errorCount }}</span>
        <span class="summary-label">{{ t("log.errorRequests") }}</span>
      </div>
    </div>
  </Card>
</template>

<style scoped>
.logs-summary {
  padding: var(--space-3);
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.panel-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.field-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.view-link {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  color: var(--color-accent);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  text-decoration: none;
}

.view-link:hover {
  color: var(--color-accent-hover);
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-2);
  margin-top: var(--space-3);
}

.summary-item {
  display: flex;
  min-height: 58px;
  flex-direction: column;
  justify-content: center;
  border-radius: var(--radius-md);
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border);
  padding: 0 var(--space-3);
}

.summary-value {
  color: var(--color-text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.summary-value--danger {
  color: var(--color-danger);
}

.summary-label {
  margin-top: var(--space-1);
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
}

</style>
