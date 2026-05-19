import { formatBytes } from "@/lib/format";
import { translate } from "@/lib/i18n";

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

type WailsCallErrorPayload = {
  message?: string;
  cause?: unknown;
  kind?: string;
};

type ErrorTranslator = (params?: ErrorParams) => string;

const messages: Record<string, ErrorTranslator> = {
  invalid_request: () => translate("error.invalidRequest"),
  bad_origin: () => translate("error.badOrigin"),
  bad_host: () => translate("error.badHost"),
  network_not_allowed: () => translate("error.networkNotAllowed"),
  access_required: () => translate("error.accessRequired"),
  access_not_required: () => translate("error.accessNotRequired"),
  access_approval_required: () => translate("error.accessApprovalRequired"),
  access_request_unavailable: () => translate("error.accessRequestUnavailable"),
  access_request_limited: () => translate("error.accessRequestLimited"),
  request_too_large: (params) => {
    const maxBytes = typeof params?.maxBytes === "number" ? params.maxBytes : undefined;
    return maxBytes ? translate("error.requestTooLarge", { maxBytes: formatBytes(maxBytes) }) : translate("error.fallback");
  },
  invalid_path: () => translate("error.invalidPath"),
  invalid_filename: (params) => {
    const maxBytes = typeof params?.maxBytes === "number" ? params.maxBytes : undefined;
    return maxBytes ? translate("error.invalidFilename", { maxBytes }) : translate("error.fallback");
  },
  invalid_message: () => translate("error.invalidMessage"),
  cannot_delete_root: () => translate("error.cannotDeleteRoot"),
  permission_denied: () => translate("error.permissionDenied"),
  not_found: () => translate("error.notFound"),
  no_files_uploaded: () => translate("error.noFilesUploaded"),
  multi_upload_fail: (params) => {
    const failed = typeof params?.failed === "number" ? params.failed : undefined;
    return failed ? translate("error.multiUploadFail", { failed }) : translate("error.fallback");
  },
  invalid_port: () => translate("error.invalidPort"),
  shared_dir_required: () => translate("error.sharedDirRequired"),
  file_too_large: (params) => {
    const maxBytes = typeof params?.maxBytes === "number" ? params.maxBytes : undefined;
    return maxBytes ? translate("error.fileTooLarge", { maxBytes: formatBytes(maxBytes) }) : translate("error.fallback");
  },
  server_error: () => translate("error.fallback"),
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
  return error ? translateError(error) : translate("error.fallback");
}

export function translateError(error: AppError) {
  return (messages[error.code] ?? (() => translate("error.fallback")))(error.params);
}

function readPayload(value: unknown): ErrorPayload | null {
  if (typeof value === "string") return parsePayload(value);

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
    const wailsPayload = readWailsCallErrorPayload(parsed);
    if (wailsPayload) return wailsPayload;
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

function readWailsCallErrorPayload(value: unknown): ErrorPayload | null {
  if (!isWailsCallErrorPayload(value)) return null;
  return readPayload(value.cause);
}

function isAppError(value: unknown): value is AppError {
  return value instanceof Error && typeof (value as AppError).code === "string";
}

function isWailsCallErrorPayload(value: unknown): value is WailsCallErrorPayload {
  return isRecord(value) && typeof value.kind === "string" && "cause" in value;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
