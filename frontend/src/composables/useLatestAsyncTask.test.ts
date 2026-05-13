import { describe, expect, it } from "vitest";
import { useLatestAsyncTask } from "@/composables/useLatestAsyncTask";

describe("useLatestAsyncTask", () => {
  it("does not let stale failures overwrite the latest task state", async () => {
    const task = useLatestAsyncTask();
    const slow = deferred<string>();
    const fast = deferred<string>();

    const slowResult = task.runLatest(() => slow.promise, () => undefined);
    const fastResult = task.runLatest(() => fast.promise, () => undefined);

    fast.resolve("current");
    await expect(fastResult).resolves.toEqual({ ok: true, value: "current" });

    slow.reject(new Error(JSON.stringify({ error: "permission_denied" })));
    const result = await slowResult;

    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.stale).toBe(true);
    expect(task.error.value).toBe("");
    expect(task.failure.value).toBe(null);
  });
});

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((promiseResolve, promiseReject) => {
    resolve = promiseResolve;
    reject = promiseReject;
  });
  return { promise, resolve, reject };
}
