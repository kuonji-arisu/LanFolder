import { describe, expect, it, vi } from "vitest";
import { useLatestAsyncTask } from "@/composables/useLatestAsyncTask";

describe("useLatestAsyncTask", () => {
  it("does not call onSuccess after invalidate", async () => {
    const task = useLatestAsyncTask();
    const pending = deferred<string>();
    const onSuccess = vi.fn();

    const result = task.runLatest(() => pending.promise, onSuccess);
    task.invalidate();
    pending.resolve("stale");
    await result;

    expect(onSuccess).not.toHaveBeenCalled();
  });

  it("marks runLatest failures stale after invalidate", async () => {
    const task = useLatestAsyncTask();
    const pending = deferred<string>();

    const resultPromise = task.runLatest(() => pending.promise, () => undefined);
    task.invalidate();
    pending.reject(new Error(JSON.stringify({ error: "permission_denied" })));
    const result = await resultPromise;

    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.stale).toBe(true);
    expect(task.error.value).toBe("");
    expect(task.failure.value).toBe(null);
  });

  it("does not let earlier runLatest results commit after a later runLatest finishes", async () => {
    const task = useLatestAsyncTask();
    const slow = deferred<string>();
    const fast = deferred<string>();
    const committed: string[] = [];

    const slowResult = task.runLatest(() => slow.promise, (value) => {
      committed.push(value);
    });
    const fastResult = task.runLatest(() => fast.promise, (value) => {
      committed.push(value);
    });

    fast.resolve("current");
    await fastResult;
    slow.resolve("stale");
    await slowResult;

    expect(committed).toEqual(["current"]);
  });

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

  it("does not apply latest invalidation to ordinary run tasks", async () => {
    const task = useLatestAsyncTask();
    const pending = deferred<string>();
    const onSuccess = vi.fn();

    const resultPromise = task.run(() => pending.promise, onSuccess);
    task.invalidate();
    pending.resolve("ordinary");
    const result = await resultPromise;

    expect(result).toEqual({ ok: true, value: "ordinary" });
    expect(onSuccess).toHaveBeenCalledWith("ordinary");
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
