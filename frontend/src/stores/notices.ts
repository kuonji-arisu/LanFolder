import { ref } from "vue";
import { defineStore } from "pinia";
import { toast } from "vue-sonner";
import type { TaskResult } from "@/composables/useAsyncTask";
import { createAppError, errorMessage, normalizeError } from "@/lib/errors";
import { noticeApi } from "@/lib/noticeApi";
import type { AppNotice, NoticeLevel, NoticePresentation, NoticeSource } from "@/types/app";

const maxNotices = 50;

export const useNoticeStore = defineStore("notices", () => {
  const notices = ref<AppNotice[]>([]);
  const seen = new Set<string>();
  let stopListener: (() => void) | undefined;

  function show(notice: AppNotice) {
    if (seen.has(notice.id)) return;
    seen.add(notice.id);
    notices.value.unshift(notice);
    if (notices.value.length > maxNotices) {
      const removed = notices.value.splice(maxNotices);
      removed.forEach((item) => seen.delete(item.id));
    }

    const message = noticeMessage(notice);
    void presentNotice(notice, message);
  }

  function showTaskResult(result: TaskResult<unknown>, source: NoticeSource = "command") {
    if (!result.ok && !result.stale) showError(result.error, source);
  }

  function showError(error: unknown, source: NoticeSource = "command") {
    const appError = normalizeError(error);
    show({
      id: localNoticeID(),
      level: "error",
      source,
      error: appError ? { error: appError.code, params: appError.params } : undefined,
      message: appError ? undefined : errorMessage(error),
      createdAt: new Date().toISOString(),
    });
  }

  function showSuccess(message: string, source: NoticeSource = "command") {
    showLocal("success", message, source);
  }

  function showLocal(level: NoticeLevel, message: string, source: NoticeSource = "command") {
    show({
      id: localNoticeID(),
      level,
      source,
      message,
      createdAt: new Date().toISOString(),
    });
  }

  async function drain() {
    const pending = await noticeApi.drainNotices();
    pending.forEach(show);
  }

  async function startListening() {
    if (stopListener) return;
    stopListener = await noticeApi.listenNotices(show);
  }

  function stopListening() {
    stopListener?.();
    stopListener = undefined;
  }

  return { notices, show, showTaskResult, showError, showSuccess, drain, startListening, stopListening };
});

async function presentNotice(notice: AppNotice, message: string) {
  let presentation: NoticePresentation = "toast";
  try {
    presentation = await noticeApi.presentNotice(notice, message);
  } catch {
    presentation = "toast";
  }
  if (presentation === "toast") showToast(notice.level, message);
}

function showToast(level: NoticeLevel, message: string) {
  if (level === "error") toast.error(message);
  else if (level === "warning") toast.warning(message);
  else if (level === "success") toast.success(message);
  else toast(message);
}

function noticeMessage(notice: AppNotice) {
  if (notice.error) {
    return errorMessage(createAppError(notice.error.error, { params: notice.error.params }));
  }
  return notice.message || errorMessage(undefined);
}

function localNoticeID() {
  return globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}
