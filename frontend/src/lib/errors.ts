import { formatBytes } from "@/lib/format";

export type ErrorParams = Record<string, unknown>;

export type AppError = Error & {
  code: string;
  params?: ErrorParams;
  status?: number;
  cause?: unknown;
};

type CreateAppErrorOptions = {
  params?: ErrorParams;
  status?: number;
  cause?: unknown;
};

type ErrorPayload = {
  error: string;
  params?: ErrorParams;
};

type ErrorTranslator = (params?: ErrorParams) => string;

const fallbackMessage = "操作失败，请重试";

const messages: Record<string, ErrorTranslator> = {
  invalid_request: () => "请求格式无效",
  invalid_path: () => "路径无效",
  invalid_message: () => "消息不能为空，且不能超过 2000 个字符",
  cannot_delete_root: () => "不能删除共享根目录",
  permission_denied: () => "没有权限执行此操作",
  not_found: () => "文件或目录不存在",
  no_files_uploaded: () => "没有选择要上传的文件",
  invalid_port: () => "端口必须在 1 到 65535 之间",
  shared_dir_required: () => "请先选择共享目录",
  file_too_large: (params) => {
    const maxBytes = typeof params?.maxBytes === "number" ? params.maxBytes : undefined;
    return maxBytes ? `文件不能超过 ${formatBytes(maxBytes)}` : "文件太大";
  },
  server_error: () => fallbackMessage,
};

export function createAppError(code: string, options: CreateAppErrorOptions = {}): AppError {
  const error = new Error(code) as AppError;
  error.code = code;
  error.params = options.params;
  error.status = options.status;
  error.cause = options.cause;
  return error;
}

export function normalizeError(err: unknown): AppError | null {
  if (isAppError(err)) return err;

  const payload = readPayload(err);
  if (!payload) return null;
  return createAppError(payload.error, {
    params: payload.params,
    status: isRecord(err) && typeof err.status === "number" ? err.status : undefined,
    cause: err,
  });
}

export function errorMessage(err: unknown) {
  const error = normalizeError(err);
  return error ? translateError(error) : fallbackMessage;
}

export function translateError(error: AppError) {
  return (messages[error.code] ?? (() => fallbackMessage))(error.params);
}

function readPayload(value: unknown): ErrorPayload | null {
  if (isRecord(value) && typeof value.error === "string") {
    return {
      error: value.error,
      params: isRecord(value.params) ? value.params : undefined,
    };
  }

  if (value instanceof Error) return parsePayload(value.message);
  if (isRecord(value) && typeof value.message === "string") return parsePayload(value.message);
  return null;
}

function parsePayload(value: string): ErrorPayload | null {
  try {
    const parsed = JSON.parse(value) as unknown;
    if (isRecord(parsed) && typeof parsed.error === "string") {
      return {
        error: parsed.error,
        params: isRecord(parsed.params) ? parsed.params : undefined,
      };
    }
  } catch {
    return null;
  }
  return null;
}

function isAppError(value: unknown): value is AppError {
  return value instanceof Error && typeof (value as AppError).code === "string";
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
