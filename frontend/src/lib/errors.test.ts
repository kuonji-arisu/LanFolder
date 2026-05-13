import { describe, expect, it } from "vitest";
import { createAppError, errorMessage, normalizeError } from "@/lib/errors";

describe("normalizeError", () => {
  it("translates stable API error codes", () => {
    expect(errorMessage(createAppError("permission_denied", { status: 403 }))).toBe("没有权限执行此操作");
  });

  it("uses params when translating error codes", () => {
    expect(errorMessage(createAppError("file_too_large", { params: { maxBytes: 512 } }))).toBe("文件不能超过 512 B");
  });

  it("maps structured command errors from Wails runtime objects", () => {
    const err = { message: JSON.stringify({ error: "invalid_port" }), cause: {}, kind: "RuntimeError" };
    expect(normalizeError(err)?.code).toBe("invalid_port");
    expect(errorMessage(err)).toBe("端口必须在 1 到 65535 之间");
  });

  it("does not infer codes from plain error messages", () => {
    expect(normalizeError(new Error("invalid_port"))).toBeNull();
    expect(errorMessage(new Error("invalid_port"))).toBe("操作失败，请重试");
  });
});
