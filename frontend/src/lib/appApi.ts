import { AppService } from "../../bindings/lanfolder/internal/appservice";
import { Config as WailsConfig } from "../../bindings/lanfolder/internal/config/models";
import { Permission as WailsPermission } from "../../bindings/lanfolder/internal/share/models";
import type { Permission } from "@/lib/constants";
import { normalizeError } from "@/lib/errors";
import type { AccessLog, AccessRequest, AccessSession, AppConfig, AppState } from "@/types/app";

function toWailsPermission(permission: Permission) {
  switch (permission) {
    case "upload":
      return WailsPermission.PermissionUpload;
    case "manage":
      return WailsPermission.PermissionManage;
    default:
      return WailsPermission.PermissionReadOnly;
  }
}

function toWailsConfig(config: AppConfig) {
  return new WailsConfig({
    ...config,
    permission: toWailsPermission(config.permission),
    language: config.language,
  });
}

async function callDesktop<T>(command: () => Promise<T>) {
  try {
    return await command();
  } catch (err) {
    throw normalizeError(err) ?? err;
  }
}

export const appApi = {
  state: async () => (await callDesktop(() => AppService.State())) as AppState,
  logs: async () => (await callDesktop(() => AppService.Logs())) as AccessLog[],
  pendingAccessRequests: async () => (await callDesktop(() => AppService.PendingAccessRequests())) as AccessRequest[],
  accessSessions: async () => (await callDesktop(() => AppService.AccessSessions())) as AccessSession[],
  approveAccessRequest: (id: string) => callDesktop(() => AppService.ApproveAccessRequest(id)),
  denyAccessRequest: (id: string) => callDesktop(() => AppService.DenyAccessRequest(id)),
  revokeAccessSession: (id: string) => callDesktop(() => AppService.RevokeAccessSession(id)),
  chooseFolder: async () => (await callDesktop(() => AppService.ChooseFolder())) as AppState,
  openSharedFolder: () => callDesktop(() => AppService.OpenSharedFolder()),
  startSharing: async () => (await callDesktop(() => AppService.StartSharing())) as AppState,
  stopSharing: async () => (await callDesktop(() => AppService.StopSharing())) as AppState,
  saveSettings: async (config: AppConfig) => (await callDesktop(() => AppService.SaveSettings(toWailsConfig(config)))) as AppState,
};
