<script setup lang="ts">
import type { DialogContentEmits, DialogContentProps } from "radix-vue";
import type { HTMLAttributes } from "vue";
import { computed } from "vue";
import { X } from "lucide-vue-next";
import {
  DialogClose,
  DialogContent,
  DialogOverlay,
  DialogPortal,
  useForwardPropsEmits,
} from "radix-vue";
import type { DialogContentVariants } from ".";
import { cn } from "@/lib/utils";
import { dialogContentVariants } from ".";

const props = defineProps<
  DialogContentProps & {
    class?: HTMLAttributes["class"];
    size?: DialogContentVariants["size"];
    height?: DialogContentVariants["height"];
  }
>();
const emits = defineEmits<DialogContentEmits>();

const delegatedProps = computed(() => {
  const { class: _, size: _size, height: _height, ...delegated } = props;
  return delegated;
});
const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
  <DialogPortal>
    <DialogOverlay class="fixed inset-0 z-50 bg-black/45" />
    <DialogContent
      v-bind="forwarded"
      :class="
        cn(
          dialogContentVariants({ size: props.size, height: props.height }),
          props.class,
        )
      "
    >
      <slot />
      <DialogClose class="absolute right-4 top-4 inline-flex h-9 w-9 items-center justify-center gap-2 whitespace-nowrap rounded-md border border-[color:var(--color-secondary-border)] bg-secondary text-sm font-medium text-secondary-foreground shadow-none ring-offset-background transition-colors hover:bg-[color:var(--color-secondary-hover)] active:bg-[color:var(--color-secondary-active)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0">
        <X class="h-4 w-4" />
        <span class="sr-only">关闭</span>
      </DialogClose>
    </DialogContent>
  </DialogPortal>
</template>
