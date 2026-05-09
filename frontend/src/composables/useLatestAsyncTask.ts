import { useAsyncTask } from "@/composables/useAsyncTask";

export function useLatestAsyncTask(initialBusy = false) {
  const { busy, error, run } = useAsyncTask(initialBusy);
  let latestId = 0;

  async function runLatest<T>(task: () => Promise<T>, onSuccess: (value: T) => Promise<void> | void) {
    const requestId = ++latestId;
    return run(async () => {
      const value = await task();
      if (requestId === latestId) {
        await onSuccess(value);
      }
      return value;
    });
  }

  return { busy, error, run, runLatest };
}
