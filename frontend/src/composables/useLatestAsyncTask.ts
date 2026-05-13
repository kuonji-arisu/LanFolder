import { useAsyncTask } from "@/composables/useAsyncTask";

export function useLatestAsyncTask(initialBusy = false) {
  const { busy, error, failure, run } = useAsyncTask(initialBusy);
  let latestId = 0;

  async function runLatest<T>(task: () => Promise<T>, onSuccess: (value: T) => Promise<void> | void) {
    const requestId = ++latestId;
    return run(async () => {
      const value = await task();
      if (requestId === latestId) {
        await onSuccess(value);
      }
      return value;
    }, undefined, {
      commit: () => requestId === latestId,
      stale: () => requestId !== latestId,
    });
  }

  return { busy, error, failure, run, runLatest };
}
