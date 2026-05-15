<script lang="ts" setup>
import type { ToasterProps } from "vue-sonner"
import "vue-sonner/style.css"
import { reactiveOmit } from "@vueuse/core"
import { CircleCheckIcon, InfoIcon, Loader2Icon, OctagonXIcon, TriangleAlertIcon } from "lucide-vue-next"
import { Toaster as Sonner } from "vue-sonner"

const props = defineProps<ToasterProps>()
const delegatedProps = reactiveOmit(props, "toastOptions", "offset", "mobileOffset")
const toastOptions: ToasterProps["toastOptions"] = {
  classes: {
    toast: "lan-toast",
    icon: "lan-toast-icon",
    content: "lan-toast-content",
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
    :offset="{ top: 'calc(var(--titlebar-height) + var(--space-3))', right: 'var(--space-4)' }"
    :mobile-offset="{ top: 'calc(var(--titlebar-height) + var(--space-3))', right: 'var(--space-4)', left: 'var(--space-4)' }"
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
.lan-toaster {
  --lan-toast-info: #0078d4;
  --lan-toast-warning: #f9a825;
}

.lan-toaster .lan-toast {
  --lan-toast-tone: var(--lan-toast-info);
  --lan-toast-fill: color-mix(in srgb, var(--lan-toast-tone) 4%, var(--color-bg-elevated));
  --lan-toast-border: color-mix(in srgb, var(--lan-toast-tone) 22%, var(--color-border));
  min-height: 50px;
  align-items: center;
  gap: 10px;
  padding: 11px 14px 11px 15px;
  border-color: var(--lan-toast-border);
  border-radius: var(--radius-md);
  background: var(--lan-toast-fill);
  color: var(--color-text-primary);
  box-shadow:
    inset 4px 0 0 var(--lan-toast-tone),
    0 6px 14px rgba(0, 0, 0, 0.07),
    0 1px 2px rgba(0, 0, 0, 0.06);
  font-family: inherit;
}

.lan-toaster .lan-toast[data-type="default"],
.lan-toaster .lan-toast[data-type="info"] {
  --lan-toast-tone: var(--lan-toast-info);
}

.lan-toaster .lan-toast[data-type="success"] {
  --lan-toast-tone: var(--color-success);
}

.lan-toaster .lan-toast[data-type="warning"] {
  --lan-toast-tone: var(--lan-toast-warning);
}

.lan-toaster .lan-toast[data-type="error"] {
  --lan-toast-tone: var(--color-danger);
}

.lan-toaster [data-icon].lan-toast-icon {
  width: 18px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 18px;
  margin: 0;
  color: var(--lan-toast-tone);
}

.lan-toaster [data-content].lan-toast-content {
  min-width: 0;
  flex: 1 1 auto;
  gap: 2px;
}

.lan-toaster [data-title].lan-toast-title {
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  line-height: 1.35;
}

.lan-toaster [data-description].lan-toast-description {
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  line-height: 1.38;
}

.lan-toaster [data-button].lan-toast-action,
.lan-toaster [data-button].lan-toast-cancel {
  height: 28px;
  border: 1px solid var(--color-secondary-border);
  border-radius: var(--radius-sm);
  padding: 0 11px;
  background: var(--color-bg-control);
  color: var(--color-text-primary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  box-shadow: var(--shadow-sm);
}

.lan-toaster [data-button].lan-toast-action {
  border-color: color-mix(in srgb, var(--lan-toast-tone) 42%, var(--color-secondary-border));
  background: color-mix(in srgb, var(--lan-toast-tone) 10%, var(--color-bg-control));
}

.lan-toaster [data-button].lan-toast-cancel {
  color: var(--color-text-secondary);
}

.dark .lan-toaster .lan-toast,
[data-theme="dark"] .lan-toaster .lan-toast {
  --lan-toast-fill: color-mix(in srgb, var(--lan-toast-tone) 8%, var(--color-bg-elevated));
  --lan-toast-border: color-mix(in srgb, var(--lan-toast-tone) 28%, var(--color-border));
  box-shadow:
    inset 4px 0 0 var(--lan-toast-tone),
    0 10px 22px rgba(0, 0, 0, 0.28),
    0 1px 2px rgba(0, 0, 0, 0.22);
}
</style>
