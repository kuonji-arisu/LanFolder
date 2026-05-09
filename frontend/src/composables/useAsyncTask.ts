import { computed, ref } from "vue";

export function errorMessage(err: unknown) {
  return err instanceof Error ? err.message : String(err);
}

export function useAsyncTask(initialBusy = false) {
  const activeTasks = ref(0);
  const hasStarted = ref(false);
  const busy = computed(() => activeTasks.value > 0 || (initialBusy && !hasStarted.value));
  const error = ref("");

  async function run<T>(task: () => Promise<T>, onSuccess?: (value: T) => Promise<void> | void) {
    hasStarted.value = true;
    activeTasks.value += 1;
    error.value = "";
    try {
      const value = await task();
      await onSuccess?.(value);
      return value;
    } catch (err) {
      error.value = errorMessage(err);
      return undefined;
    } finally {
      activeTasks.value = Math.max(0, activeTasks.value - 1);
    }
  }

  return { busy, error, run };
}
