<script setup lang="ts">
import type { DialogContentEmits, DialogContentProps } from "radix-vue";
import type { HTMLAttributes } from "vue";
import { computed, unref, useAttrs } from "vue";
import { X } from "lucide-vue-next";
import {
  DialogClose,
  DialogContent,
  DialogOverlay,
  DialogPortal,
  useForwardPropsEmits,
} from "radix-vue";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";

defineOptions({
  inheritAttrs: false,
});

const props = defineProps<
  DialogContentProps & {
    class?: HTMLAttributes["class"];
    showClose?: boolean;
  }
>();
const emits = defineEmits<DialogContentEmits>();

const delegatedProps = computed(() => {
  const { class: _, showClose: _showClose, ...delegated } = props;
  return delegated;
});
const forwarded = useForwardPropsEmits(delegatedProps, emits);
const attrs = useAttrs();
const contentProps = computed(() => ({
  ...unref(forwarded),
  ...attrs,
}));
const { t } = useI18n();
</script>

<template>
  <DialogPortal>
    <DialogOverlay class="fixed inset-0 z-50 bg-black/45" />
    <DialogContent
      v-bind="contentProps"
      :class="
        cn(
          'fixed left-1/2 top-1/2 z-50 grid w-full max-w-lg -translate-x-1/2 -translate-y-1/2 gap-4 border bg-card p-4 text-card-foreground shadow-lg duration-200 sm:rounded-lg',
          props.class,
        )
      "
    >
      <slot />
      <DialogClose
        v-if="showClose !== false"
        class="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none"
      >
        <X class="h-4 w-4" />
        <span class="sr-only">{{ t("app.close") }}</span>
      </DialogClose>
    </DialogContent>
  </DialogPortal>
</template>
