import { AppService } from "../../bindings/lanfolder";
import { Config as WailsConfig } from "../../bindings/lanfolder/internal/config/models";
import { Permission as WailsPermission } from "../../bindings/lanfolder/internal/share/models";
import type { Permission } from "@/lib/constants";
import type { AccessLog, AppConfig, AppState } from "@/types/app";

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
  });
}

export const appApi = {
  state: async () => (await AppService.State()) as AppState,
  logs: async () => (await AppService.Logs()) as AccessLog[],
  chooseFolder: async () => (await AppService.ChooseFolder()) as AppState,
  openSharedFolder: () => AppService.OpenSharedFolder(),
  startSharing: async () => (await AppService.StartSharing()) as AppState,
  stopSharing: async () => (await AppService.StopSharing()) as AppState,
  saveSettings: async (config: AppConfig) => (await AppService.SaveSettings(toWailsConfig(config))) as AppState,
};
