<script setup lang="ts">
import type { DialogTitleProps } from "radix-vue";
import type { HTMLAttributes } from "vue";
import { computed } from "vue";
import { DialogTitle } from "radix-vue";
import { cn } from "@/lib/utils";

const props = defineProps<DialogTitleProps & { class?: HTMLAttributes["class"] }>();
const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props;
  return delegated;
});
</script>

<template>
  <DialogTitle v-bind="delegatedProps" :class="cn('text-lg font-semibold leading-none tracking-normal', props.class)">
    <slot />
  </DialogTitle>
</template>
