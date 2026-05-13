import type { Permission } from "./constants";
import { createAppError } from "@/lib/errors";
import type { PermissionOption } from "@/types/app";

export interface FileEntry {
  name: string;
  path: string;
  isDir: boolean;
  size: number;
  modTime: string;
  extension?: string;
}

export interface ListResult {
  path: string;
  parentPath: string;
  entries: FileEntry[];
}

export interface ServerStatus {
  running: boolean;
  port: number;
  permission: Permission;
  permissions: PermissionOption[];
}

interface ApiError {
  error: string;
  params?: Record<string, unknown>;
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, options);
  if (!response.ok) {
    let code = "server_error";
    let params: Record<string, unknown> | undefined;
    try {
      const body = (await response.json()) as ApiError;
      if (body.error) code = body.error;
      if (body.params) params = body.params;
    } catch {
      // Keep the generic server error fallback.
    }
    throw createAppError(code, { params, status: response.status });
  }
  return (await response.json()) as T;
}

export const fileApi = {
  status: () => request<ServerStatus>("/api/status"),
  list: (path = "") => request<ListResult>(`/api/list?path=${encodeURIComponent(path)}`),
  upload: async (path: string, files: FileList | File[]) => {
    const form = new FormData();
    Array.from(files).forEach((file) => form.append("files", file));
    return request<{ entries: FileEntry[] }>(`/api/upload?path=${encodeURIComponent(path)}`, {
      method: "POST",
      body: form,
    });
  },
  delete: (path: string) =>
    request<{ ok: true }>("/api/delete", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path }),
    }),
  mkdir: (path: string, name: string) =>
    request<FileEntry>("/api/mkdir", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path, name }),
    }),
  downloadUrl: (path: string) => `/api/download?path=${encodeURIComponent(path)}`,
};
