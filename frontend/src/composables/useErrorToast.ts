import { errorMessage } from "@/lib/errors";
import type { TaskResult } from "@/composables/useAsyncTask";
import { toast } from "vue-sonner";

export function useErrorToast() {
  function showError(err: unknown) {
    toast.error(errorMessage(err));
  }

  function showResultError(result: TaskResult<unknown>) {
    if (!result.ok && !result.stale) showError(result.error);
  }

  return { showError, showResultError };
}
