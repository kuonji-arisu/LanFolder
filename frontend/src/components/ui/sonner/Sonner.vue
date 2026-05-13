<script lang="ts" setup>
import type { ToasterProps } from "vue-sonner"
import "vue-sonner/style.css"
import { reactiveOmit } from "@vueuse/core"
import { CircleCheckIcon, InfoIcon, Loader2Icon, OctagonXIcon, TriangleAlertIcon } from "lucide-vue-next"
import { Toaster as Sonner } from "vue-sonner"

const props = defineProps<ToasterProps>()
const delegatedProps = reactiveOmit(props, "toastOptions")
const toastOptions: ToasterProps["toastOptions"] = {
  unstyled: true,
  classes: {
    toast: "lan-toast",
    title: "lan-toast-title",
    description: "lan-toast-description",
    actionButton: "lan-toast-action",
    cancelButton: "lan-toast-cancel",
  },
}
</script>

<template>
  <Sonner
    class="lan-toaster"
    :toast-options="toastOptions"
    v-bind="delegatedProps"
  >
    <template #success-icon>
      <CircleCheckIcon class="size-4" />
    </template>
    <template #info-icon>
      <InfoIcon class="size-4" />
    </template>
    <template #warning-icon>
      <TriangleAlertIcon class="size-4" />
    </template>
    <template #error-icon>
      <OctagonXIcon class="size-4" />
    </template>
    <template #loading-icon>
      <div>
        <Loader2Icon class="size-4 animate-spin" />
      </div>
    </template>
  </Sonner>
</template>

<style>
.lan-toaster .lan-toast {
  width: min(360px, calc(100vw - 28px));
  min-height: 54px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 11px 14px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-overlay);
  color: var(--color-text-primary);
  box-shadow: var(--shadow-md);
  backdrop-filter: blur(20px);
  font-family: inherit;
}

.lan-toaster .lan-toast[data-type="error"] {
  border-color: color-mix(in srgb, var(--color-danger) 30%, var(--color-border));
}

.lan-toaster .lan-toast[data-type="success"] {
  border-color: color-mix(in srgb, var(--color-success) 28%, var(--color-border));
}

.lan-toaster .lan-toast[data-type="warning"] {
  border-color: color-mix(in srgb, #f4b740 36%, var(--color-border));
}

.lan-toaster .lan-toast [data-icon] {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 28px;
  border-radius: var(--radius-sm);
  background: var(--color-bg-muted);
  color: var(--color-accent);
}

.lan-toaster .lan-toast[data-type="error"] [data-icon] {
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  color: var(--color-danger);
}

.lan-toaster .lan-toast[data-type="success"] [data-icon] {
  background: color-mix(in srgb, var(--color-success) 14%, transparent);
  color: var(--color-success);
}

.lan-toaster .lan-toast[data-type="warning"] [data-icon] {
  background: color-mix(in srgb, #f4b740 16%, transparent);
  color: #b87800;
}

.lan-toaster .lan-toast-title {
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  line-height: 1.35;
}

.lan-toaster .lan-toast-description {
  margin-top: 2px;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  line-height: 1.35;
}

.lan-toaster .lan-toast-action,
.lan-toaster .lan-toast-cancel {
  height: 30px;
  border-radius: var(--radius-sm);
  padding: 0 12px;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

.lan-toaster .lan-toast-action {
  background: var(--color-accent);
  color: var(--color-text-on-accent);
}

.lan-toaster .lan-toast-cancel {
  background: var(--color-bg-control);
  color: var(--color-text-primary);
}
</style>
