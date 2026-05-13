import { computed, ref } from "vue";
import { errorMessage as appErrorMessage } from "@/lib/errors";

export type TaskSuccess<T> = {
  ok: true;
  value: T;
};

export type TaskFailure = {
  ok: false;
  error: unknown;
  message: string;
  stale?: boolean;
};

export type TaskResult<T> = TaskSuccess<T> | TaskFailure;

type RunOptions = {
  commit?: () => boolean;
  stale?: () => boolean;
};

export function taskSuccess<T>(value: T): TaskSuccess<T> {
  return { ok: true, value };
}

export function taskFailure(error: unknown, options: { stale?: boolean } = {}): TaskFailure {
  return { ok: false, error, message: appErrorMessage(error), stale: options.stale };
}

export function errorMessage(err: unknown) {
  return appErrorMessage(err);
}

export function useAsyncTask(initialBusy = false) {
  const activeTasks = ref(0);
  const hasStarted = ref(false);
  const busy = computed(() => activeTasks.value > 0 || (initialBusy && !hasStarted.value));
  const error = ref("");
  const failure = ref<unknown>(null);

  async function run<T>(task: () => Promise<T>, onSuccess?: (value: T) => Promise<void> | void, options: RunOptions = {}): Promise<TaskResult<T>> {
    hasStarted.value = true;
    activeTasks.value += 1;
    const shouldCommit = () => options.commit?.() ?? true;
    if (shouldCommit()) {
      error.value = "";
      failure.value = null;
    }
    try {
      const value = await task();
      await onSuccess?.(value);
      return taskSuccess(value);
    } catch (err) {
      const result = taskFailure(err, { stale: options.stale?.() });
      if (shouldCommit()) {
        failure.value = err;
        error.value = result.message;
      }
      return result;
    } finally {
      activeTasks.value = Math.max(0, activeTasks.value - 1);
    }
  }

  return { busy, error, failure, run };
}
