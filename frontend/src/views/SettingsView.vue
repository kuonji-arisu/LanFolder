<script setup lang="ts">
import { ref, watch } from "vue";
import { Moon, Sun } from "lucide-vue-next";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { useTheme } from "@/composables/useTheme";
import { normalizeError } from "@/lib/errors";
import { useI18n, type Language } from "@/lib/i18n";
import { useAppStore } from "@/stores/app";
import { useNoticeStore } from "@/stores/notices";
import type { AppConfig } from "@/types/app";

const app = useAppStore();
const notices = useNoticeStore();
const { theme, setTheme } = useTheme();
const { language, languageOptions, t } = useI18n();
const portDraft = ref("");
const portError = ref("");

function syncPortDraft() {
  portDraft.value = String(app.config.port);
  portError.value = "";
}

watch(() => app.config.port, syncPortDraft, { immediate: true });

function handlePortInput(event: Event) {
  const input = event.target as HTMLInputElement;
  const digits = input.value.replace(/\D/g, "");
  if (input.value !== digits) input.value = digits;
  portDraft.value = digits;
  portError.value = "";
}

async function savePort() {
  const validation = validatePort(portDraft.value);
  if (!validation.ok) {
    portError.value = validation.message;
    return;
  }
  portError.value = "";
  const port = validation.port;
  if (port === app.config.port) {
    syncPortDraft();
    return;
  }
  const result = await app.saveConfig({ port });
  if (!result.ok) {
    if (normalizeError(result.error)?.code !== "invalid_port") syncPortDraft();
    portError.value = result.message;
  }
}

function validatePort(value: string): { ok: true; port: number } | { ok: false; message: string } {
  const trimmed = value.trim();
  if (!trimmed) return { ok: false, message: t("settings.port.empty") };
  const port = Number(trimmed);
  if (port <= 0 || port > 65535) return { ok: false, message: t("error.invalidPort") };
  return { ok: true, port };
}

async function saveSettingWithNotice(partial: Partial<AppConfig>) {
  notices.showTaskResult(await app.saveConfig(partial));
}

async function saveLanguage(value: Language) {
  if (value === app.config.language) return;
  await saveSettingWithNotice({ language: value });
}
</script>

<template>
  <main class="settings-view">
    <div class="settings-form">
      <div class="settings-row">
        <div>
          <label class="field-label" for="settings-port">{{ t("settings.port.label") }}</label>
          <p v-if="portError" class="field-error">{{ portError }}</p>
        </div>
        <Input
          id="settings-port"
          v-model="portDraft"
          class="port-input"
          type="text"
          inputmode="numeric"
          pattern="[0-9]*"
          maxlength="5"
          :aria-invalid="Boolean(portError)"
          @input="handlePortInput"
          @change="savePort"
          @keydown.enter="savePort"
        />
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">{{ t("settings.theme.label") }}</div>
        </div>
        <div class="segmented-toggle">
          <button :class="['theme-option', { 'theme-option--active': theme === 'light' }]" @click="setTheme('light')">
            <Sun class="h-[13px] w-[13px]" />{{ t("settings.theme.light") }}
          </button>
          <button :class="['theme-option', { 'theme-option--active': theme === 'dark' }]" @click="setTheme('dark')">
            <Moon class="h-[13px] w-[13px]" />{{ t("settings.theme.dark") }}
          </button>
        </div>
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">{{ t("settings.language.label") }}</div>
        </div>
        <div class="segmented-toggle">
          <button
            v-for="option in languageOptions"
            :key="option.value"
            :class="['theme-option', { 'theme-option--active': language === option.value }]"
            @click="saveLanguage(option.value)"
          >
            {{ option.label }}
          </button>
        </div>
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">{{ t("settings.accessApproval.label") }}</div>
        </div>
        <span class="switch-tooltip" :title="app.config.autoShare ? t('settings.turnOffAutoShare') : ''">
          <Switch :checked="app.config.accessApproval" :disabled="app.config.autoShare" @update:checked="saveSettingWithNotice({ accessApproval: $event })" />
        </span>
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">{{ t("settings.autoShare.label") }}</div>
        </div>
        <span class="switch-tooltip" :title="app.config.accessApproval ? '' : t('settings.autoShare.disabledTitle')">
          <Switch :checked="app.config.autoShare" :disabled="!app.config.accessApproval" @update:checked="saveSettingWithNotice({ autoShare: $event })" />
        </span>
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">{{ t("settings.startAtLogin.label") }}</div>
        </div>
        <Switch :checked="app.config.startAtLogin" :disabled="!app.state?.capabilities.startAtLogin" @update:checked="saveSettingWithNotice({ startAtLogin: $event })" />
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">{{ t("settings.keepInTray.label") }}</div>
        </div>
        <Switch :checked="app.config.keepInTray" @update:checked="saveSettingWithNotice({ keepInTray: $event })" />
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">{{ t("settings.hiddenFiles.label") }}</div>
        </div>
        <Switch :checked="app.config.showHiddenFiles" @update:checked="saveSettingWithNotice({ showHiddenFiles: $event })" />
      </div>
    </div>
  </main>
</template>

<style scoped>
.settings-view {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: var(--space-5);
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.settings-view::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.settings-form {
  max-width: 420px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.settings-row {
  min-height: 35px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.settings-row > div:first-child {
  min-width: 0;
  flex: 1;
}

.settings-row > :last-child {
  flex-shrink: 0;
}

.port-input {
  width: 82px;
  height: 35px;
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  text-align: center;
  font-size: var(--font-size-md);
}

.switch-tooltip {
  display: inline-flex;
}

.field-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.segmented-toggle {
  width: 180px;
  height: 35px;
  display: flex;
  overflow: hidden;
  flex-shrink: 0;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.theme-option {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 0 var(--space-3);
  color: var(--color-text-secondary);
  background: transparent;
  border: none;
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.theme-option + .theme-option {
  border-left: 1px solid var(--color-border);
}

.theme-option:hover {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.theme-option--active {
  background: var(--color-accent);
  color: var(--color-text-on-accent);
  font-weight: var(--font-weight-medium);
}

.theme-option--active:hover {
  background: var(--color-accent-hover);
  color: var(--color-text-on-accent);
}

.field-error {
  margin: 4px 0 0;
  color: var(--color-danger);
  font-size: var(--font-size-xs);
}
</style>
