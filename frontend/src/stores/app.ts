import { computed, ref } from "vue";
import { defineStore } from "pinia";
import { Events } from "@wailsio/runtime";
import { useAsyncTask, type TaskResult } from "@/composables/useAsyncTask";
import { appApi } from "@/lib/appApi";
import { DEFAULT_PORT, type Permission } from "@/lib/constants";
import type { AccessLog, AccessRequest, AccessSession, AppConfig, AppState, PermissionOption } from "@/types/app";

const defaultConfig: AppConfig = {
  sharedDir: "",
  port: DEFAULT_PORT,
  permission: "readonly",
  accessApproval: false,
  autoShare: false,
  startAtLogin: false,
  keepInTray: false,
  showHiddenFiles: false,
};

const defaultPermission: PermissionOption = { value: "readonly", label: "只读", description: "" };

export const useAppStore = defineStore("app", () => {
  const state = ref<AppState | null>(null);
  const logs = ref<AccessLog[]>([]);
  const pendingAccessRequests = ref<AccessRequest[]>([]);
  const accessSessions = ref<AccessSession[]>([]);
  const { busy, error, run: runTask } = useAsyncTask();
  let refreshTimer: number | undefined;
  let stopStateChangedListener: (() => void) | undefined;
  let settingsSaveQueue = Promise.resolve();

  const config = computed(() => state.value?.config ?? defaultConfig);
  const isRunning = computed(() => Boolean(state.value?.server.running));
  const primaryAddress = computed(() => state.value?.addresses?.[0] ?? `http://127.0.0.1:${config.value.port}`);
  const activePermission = computed(() => state.value?.permissions.find((item) => item.value === config.value.permission) ?? state.value?.permissions[0] ?? defaultPermission);

  async function loadState() {
    state.value = await appApi.state();
  }

  async function loadLogs() {
    logs.value = await appApi.logs();
  }

  async function loadPendingAccessRequests() {
    pendingAccessRequests.value = await appApi.pendingAccessRequests();
  }

  async function loadAccess() {
    const [requests, sessions] = await Promise.all([appApi.pendingAccessRequests(), appApi.accessSessions()]);
    pendingAccessRequests.value = requests;
    accessSessions.value = sessions;
  }

  async function loadSnapshot() {
    await Promise.all([loadState(), loadLogs(), loadAccess()]);
  }

  async function commitSnapshot(task: () => Promise<AppState>): Promise<TaskResult<AppState>> {
    const result = await runTask(task);
    if (!result.ok) return result;
    state.value = result.value;
    await Promise.all([loadLogs(), loadAccess()]);
    return result;
  }

  async function runCommand<T>(task: () => Promise<T>): Promise<TaskResult<T>> {
    return runTask(task);
  }

  function nextSettings(partial: Partial<AppConfig>) {
    return { ...config.value, ...partial, port: Number(partial.port ?? config.value.port) };
  }

  function enqueueSettingsSave(task: () => Promise<TaskResult<AppState>>) {
    const run = settingsSaveQueue.then(task, task);
    settingsSaveQueue = run.then(
      () => undefined,
      () => undefined,
    );
    return run;
  }

  function saveConfig(partial: Partial<AppConfig> = {}) {
    return enqueueSettingsSave(() => commitSnapshot(() => appApi.saveSettings(nextSettings(partial))));
  }

  async function setPermission(permission: Permission) {
    return saveConfig({ permission });
  }

  async function chooseFolder() {
    return commitSnapshot(() => appApi.chooseFolder());
  }

  async function openSharedFolder() {
    return runCommand(() => appApi.openSharedFolder());
  }

  async function toggleSharing() {
    return commitSnapshot(() => (isRunning.value ? appApi.stopSharing() : appApi.startSharing()));
  }

  async function approveAccessRequest(id: string) {
    const result = await runTask(() => appApi.approveAccessRequest(id));
    await loadAccess();
    return result;
  }

  async function denyAccessRequest(id: string) {
    const result = await runTask(() => appApi.denyAccessRequest(id));
    await loadAccess();
    return result;
  }

  async function revokeAccessSession(id: string) {
    const result = await runTask(() => appApi.revokeAccessSession(id));
    await loadAccess();
    return result;
  }

  function startAutoRefresh() {
    if (refreshTimer !== undefined) return;
    refreshTimer = window.setInterval(() => {
      if (state.value?.server.running) void Promise.all([loadLogs(), loadAccess()]);
    }, 2500);
    stopStateChangedListener = Events.On("app:state-changed", () => void loadSnapshot());
  }

  function stopAutoRefresh() {
    if (refreshTimer !== undefined) {
      window.clearInterval(refreshTimer);
      refreshTimer = undefined;
    }
    stopStateChangedListener?.();
    stopStateChangedListener = undefined;
  }

  return {
    state,
    logs,
    pendingAccessRequests,
    accessSessions,
    busy,
    error,
    config,
    isRunning,
    primaryAddress,
    activePermission,
    loadState,
    loadLogs,
    loadAccess,
    loadSnapshot,
    saveConfig,
    setPermission,
    chooseFolder,
    openSharedFolder,
    toggleSharing,
    approveAccessRequest,
    denyAccessRequest,
    revokeAccessSession,
    startAutoRefresh,
    stopAutoRefresh,
  };
});
