<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import { X } from "lucide-vue-next";
import { DialogClose } from "radix-vue";
import { Button } from "@/components/ui/button";
import { DialogHeader } from "@/components/ui/dialog";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";

const props = withDefaults(
  defineProps<{
    class?: HTMLAttributes["class"];
    showClose?: boolean;
  }>(),
  {
    showClose: true,
  },
);

const { t } = useI18n();
</script>

<template>
  <div :class="cn('app-dialog-header', props.class)">
    <DialogHeader class="app-dialog-heading">
      <slot />
    </DialogHeader>

    <div class="app-dialog-actions">
      <slot name="actions" />
      <DialogClose v-if="showClose" as-child>
        <Button variant="secondary" size="icon" :aria-label="t('app.close')">
          <X class="h-4 w-4" />
          <span class="sr-only">{{ t("app.close") }}</span>
        </Button>
      </DialogClose>
    </div>
  </div>
</template>

<style scoped>
.app-dialog-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.app-dialog-heading {
  min-width: 0;
  flex: 1 1 auto;
}

.app-dialog-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: var(--space-2);
}
</style>
