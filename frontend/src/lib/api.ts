import type { Permission } from "./constants";
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
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, options);
  if (!response.ok) {
    let message = `${response.status} ${response.statusText}`;
    try {
      const body = (await response.json()) as ApiError;
      if (body.error) message = body.error;
    } catch {
      // Keep the HTTP status as the fallback message.
    }
    throw new Error(message);
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
