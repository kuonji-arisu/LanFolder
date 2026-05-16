<script setup lang="ts">
import { ref, watch } from "vue";
import { Moon, Sun, Trash2 } from "lucide-vue-next";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { useTheme } from "@/composables/useTheme";
import { normalizeError } from "@/lib/errors";
import { useAppStore } from "@/stores/app";
import { useNoticeStore } from "@/stores/notices";
import type { AppConfig } from "@/types/app";

const app = useAppStore();
const notices = useNoticeStore();
const { theme, setTheme } = useTheme();
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
  const result = await app.saveConfig({ port });
  if (!result.ok) {
    if (normalizeError(result.error)?.code !== "invalid_port") syncPortDraft();
    portError.value = result.message;
  }
}

function validatePort(value: string): { ok: true; port: number } | { ok: false; message: string } {
  const trimmed = value.trim();
  if (!trimmed) return { ok: false, message: "请输入端口" };
  const port = Number(trimmed);
  if (port <= 0 || port > 65535) return { ok: false, message: "端口必须在 1 到 65535 之间" };
  return { ok: true, port };
}

async function saveSettingWithNotice(partial: Partial<AppConfig>) {
  notices.showTaskResult(await app.saveConfig(partial));
}
</script>

<template>
  <main class="settings-view">
    <div class="settings-form">
      <label class="settings-field">
        <span class="field-label">端口</span>
        <Input v-model="portDraft" type="text" inputmode="numeric" pattern="[0-9]*" :aria-invalid="Boolean(portError)" @input="handlePortInput" @change="savePort" />
        <span v-if="portError" class="field-error">{{ portError }}</span>
        <span v-else class="field-hint">默认 8899，修改后自动重启共享服务</span>
      </label>

      <div class="settings-row">
        <div>
          <div class="field-label">外观主题</div>
          <p class="field-hint">更改界面颜色</p>
        </div>
        <div class="theme-toggle">
          <button :class="['theme-option', { 'theme-option--active': theme === 'light' }]" @click="setTheme('light')">
            <Sun class="h-[13px] w-[13px]" />浅色
          </button>
          <button :class="['theme-option', { 'theme-option--active': theme === 'dark' }]" @click="setTheme('dark')">
            <Moon class="h-[13px] w-[13px]" />深色
          </button>
        </div>
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">启动应用后自动共享</div>
          <p class="field-hint">打开应用后使用上次目录自动运行</p>
        </div>
        <Switch :checked="app.config.autoShare" @update:checked="saveSettingWithNotice({ autoShare: $event })" />
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">开机自动启动</div>
          <p class="field-hint">{{ app.state?.capabilities.startAtLogin ? "登录系统后自动打开 LanFolder" : "当前平台暂不支持" }}</p>
        </div>
        <Switch :checked="app.config.startAtLogin" :disabled="!app.state?.capabilities.startAtLogin" @update:checked="saveSettingWithNotice({ startAtLogin: $event })" />
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">关闭窗口后保持后台运行</div>
          <p class="field-hint">关闭窗口时隐藏到系统托盘</p>
        </div>
        <Switch :checked="app.config.keepInTray" @update:checked="saveSettingWithNotice({ keepInTray: $event })" />
      </div>

      <div class="settings-row">
        <div>
          <div class="field-label">显示隐藏文件</div>
          <p class="field-hint">显示点号文件，受管目录仍会隐藏</p>
        </div>
        <Switch :checked="app.config.showHiddenFiles" @update:checked="saveSettingWithNotice({ showHiddenFiles: $event })" />
      </div>

      <div class="delete-note">
        <Trash2 class="h-4 w-4" />
        <span>删除会移入共享目录下的 .lanfolder/trash</span>
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

.settings-field {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.settings-field :deep(input) {
  height: var(--input-height);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  font-size: var(--font-size-md);
}

.settings-row {
  min-height: 46px;
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

.field-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.field-hint {
  margin: 4px 0 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.theme-toggle {
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

.delete-note {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  min-height: 58px;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: 0 var(--space-4);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.delete-note svg {
  flex-shrink: 0;
  color: var(--color-text-tertiary);
}

.delete-note span {
  min-width: 0;
}

.field-error {
  margin-top: 4px;
  color: var(--color-danger);
  font-size: var(--font-size-xs);
}
</style>
