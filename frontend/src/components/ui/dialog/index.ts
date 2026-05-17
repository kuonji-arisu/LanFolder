import type { VariantProps } from "class-variance-authority";
import { cva } from "class-variance-authority";

export { default as Dialog } from "./Dialog.vue";
export { default as DialogContent } from "./DialogContent.vue";
export { default as DialogDescription } from "./DialogDescription.vue";
export { default as DialogHeader } from "./DialogHeader.vue";
export { default as DialogTitle } from "./DialogTitle.vue";

export const dialogContentVariants = cva(
  "fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2 gap-4 overflow-hidden rounded-lg border bg-card p-4 text-card-foreground shadow-lg focus:outline-none",
  {
    variants: {
      size: {
        default: "w-[calc(100vw-28px)] max-w-[560px]",
        sm: "w-[calc(100vw-28px)] max-w-[420px]",
        large: "w-[calc(100vw_-_var(--space-6))] max-w-[var(--content-max-width)]",
      },
      height: {
        content: "grid max-h-[calc(100dvh-36px)]",
        bounded: "flex max-h-[calc(100dvh_-_var(--space-6))] flex-col",
        fill: "flex h-[calc(100dvh_-_var(--space-6))] flex-col",
      },
    },
    defaultVariants: {
      size: "default",
      height: "content",
    },
  },
);

export type DialogContentVariants = VariantProps<typeof dialogContentVariants>;
