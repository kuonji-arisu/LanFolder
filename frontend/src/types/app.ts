import type { Permission } from "@/lib/constants";

export interface AppConfig {
  sharedDir: string;
  port: number;
  permission: Permission;
  accessApproval: boolean;
  autoShare: boolean;
  startAtLogin: boolean;
  keepInTray: boolean;
  showHiddenFiles: boolean;
}

export interface ServerRuntime {
  running: boolean;
  host: string;
  port: number;
  root: string;
  permission: Permission;
  accessApproval: boolean;
}

export interface AccessLog {
  time: string;
  method: string;
  path: string;
  remote: string;
  status: number;
  action: string;
  target?: string;
  targetPath?: string;
  detail?: string;
}

export interface AccessRequest {
  id: string;
  code: string;
  ip: string;
  userAgent: string;
  createdAt: string;
  expiresAt: string;
}

export interface AccessSession {
  id: string;
  ip: string;
  userAgent: string;
  createdAt: string;
}

export interface MessageEntry {
  id: string;
  createdAt: string;
  clientId: string;
  text: string;
}

export interface PermissionOption {
  value: Permission;
  label: string;
  description: string;
}

export interface WindowInfo {
  x: number;
  y: number;
  width: number;
  height: number;
  maximised: boolean;
  minimised: boolean;
  fullscreen: boolean;
}

export interface AppInfo {
  name: string;
  version: string;
  os: string;
  osName: string;
  osVersion: string;
  arch: string;
  locale: string;
  systemTheme: "light" | "dark";
  configPath: string;
  window: WindowInfo;
}

export interface Capabilities {
  startAtLogin: boolean;
  tray: boolean;
  openFolder: boolean;
  systemThemeEvents: boolean;
  windowState: boolean;
}

export interface AppState {
  config: AppConfig;
  server: ServerRuntime;
  appInfo: AppInfo;
  capabilities: Capabilities;
  addresses: string[];
  permissions: PermissionOption[];
}

export type NoticeLevel = "info" | "success" | "warning" | "error";
export type NoticeSource = "command" | "startup" | "system";

export interface ErrorPayload {
  error: string;
  params?: Record<string, unknown>;
}

export interface AppNotice {
  id: string;
  level: NoticeLevel;
  source: NoticeSource | string;
  error?: ErrorPayload | null;
  message?: string;
  createdAt: string;
}
