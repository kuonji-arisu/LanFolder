<script setup lang="ts">
import type { CSSProperties, HTMLAttributes } from "vue";
import { computed } from "vue";
import { DialogContent } from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

const props = withDefaults(
  defineProps<{
    class?: HTMLAttributes["class"];
    width?: string;
    height?: string;
    maxWidth?: string;
    maxHeight?: string;
  }>(),
  {
    width: "calc(100vw - 28px)",
    maxWidth: "560px",
    maxHeight: "calc(100dvh - 36px)",
  },
);

const contentStyle = computed<CSSProperties>(() => ({
  width: props.width,
  height: props.height,
  maxWidth: props.maxWidth,
  maxHeight: props.maxHeight,
}));
</script>

<template>
  <DialogContent
    :show-close="false"
    :class="cn('flex min-h-0 flex-col overflow-hidden rounded-lg', props.class)"
    :style="contentStyle"
  >
    <slot />
  </DialogContent>
</template>
