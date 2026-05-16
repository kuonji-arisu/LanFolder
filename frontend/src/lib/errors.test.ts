import { describe, expect, it } from "vitest";
import { createAppError, errorMessage, normalizeError } from "@/lib/errors";

describe("normalizeError", () => {
  it("translates stable API error codes", () => {
    expect(errorMessage(createAppError("permission_denied", { status: 403 }))).toBe("没有权限执行此操作");
    expect(errorMessage(createAppError("bad_origin", { status: 403 }))).toBe("请求来源不被允许");
    expect(errorMessage(createAppError("bad_host", { status: 403 }))).toBe("访问地址不被允许");
    expect(errorMessage(createAppError("network_not_allowed", { status: 403 }))).toBe("仅允许可信局域网访问");
    expect(errorMessage(createAppError("access_request_limited", { status: 429 }))).toBe("当前访问请求过多，请稍后再试");
  });

  it("uses params when translating error codes", () => {
    expect(errorMessage(createAppError("file_too_large", { params: { maxBytes: 512 } }))).toBe("文件不能超过 512 B");
    expect(errorMessage(createAppError("request_too_large", { params: { maxBytes: 16 << 10 } }))).toBe("请求不能超过 16.0 KB");
    expect(errorMessage(createAppError("invalid_filename", { params: { maxBytes: 255 } }))).toBe("名称不能超过 255 字节");
  });

  it("maps structured command errors from Wails runtime objects", () => {
    const err = { message: JSON.stringify({ error: "invalid_port" }), cause: {}, kind: "RuntimeError" };
    expect(normalizeError(err)?.code).toBe("invalid_port");
    expect(errorMessage(err)).toBe("端口必须在 1 到 65535 之间");
  });

  it("maps desktop command errors from Wails CallError causes", () => {
    const err = new Error(JSON.stringify({
      message: "shared_dir_required",
      cause: { error: "shared_dir_required" },
      kind: "RuntimeError",
    }));
    expect(normalizeError(err)?.code).toBe("shared_dir_required");
    expect(errorMessage(err)).toBe("请先选择共享目录");
  });

  it("does not infer codes from plain error messages", () => {
    expect(normalizeError(new Error("invalid_port"))).toBeNull();
    expect(errorMessage(new Error("invalid_port"))).toBe("操作失败，请重试");
  });
});
