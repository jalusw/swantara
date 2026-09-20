// @vitest-environment node
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createProfiler, finishProfiler, finishProfilerWithError } from "../profiler";

const { logger } = vi.hoisted(() => ({
  logger: {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("../logger", () => ({ logger }));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("createProfiler", () => {
  it("captures the method, path and a start time", () => {
    const profiler = createProfiler("POST", "/api/things");

    expect(profiler.method).toBe("POST");
    expect(profiler.path).toBe("/api/things");
    expect(profiler.startTime).toEqual(expect.any(Number));
  });
});

describe("finishProfiler", () => {
  it("logs a fast, non-error response as info and returns it unchanged", () => {
    const response = new Response(JSON.stringify({ ok: true }), {
      status: 200,
      headers: { "content-length": "2" },
    });

    const result = finishProfiler(createProfiler("GET", "/api/health"), response);

    expect(result).toBe(response);
    expect(logger.info).toHaveBeenCalledOnce();
    expect(logger.warn).not.toHaveBeenCalled();
    expect(logger.error).not.toHaveBeenCalled();
    const payload = logger.info.mock.calls[0]![0];
    expect(payload).toMatchObject({
      method: "GET",
      path: "/api/health",
      statusCode: 200,
      contentLength: "2 B",
      memory: expect.objectContaining({
        rss: expect.stringMatching(/MB$/),
      }),
    });
    expect(payload.duration).toMatch(/(ms|s)$/);
  });

  it("warns when the request took longer than one second", () => {
    const ctx = createProfiler("GET", "/api/slow");
    ctx.startTime = performance.now() - 1500;

    finishProfiler(ctx, new Response(null, { status: 200 }));

    expect(logger.warn).toHaveBeenCalledOnce();
    expect(logger.info).not.toHaveBeenCalled();
  });

  it("errors when the response status is 5xx", () => {
    finishProfiler(createProfiler("GET", "/api/broken"), new Response(null, { status: 500 }));

    expect(logger.error).toHaveBeenCalledOnce();
    expect(logger.info).not.toHaveBeenCalled();
  });

  it("omits content length when the header is absent", () => {
    finishProfiler(createProfiler("GET", "/api/empty"), new Response(null, { status: 204 }));

    expect(logger.info).toHaveBeenCalledOnce();
    const payload = logger.info.mock.calls[0]![0];
    expect(payload.contentLength).toBeNull();
  });
});

describe("finishProfilerWithError", () => {
  it("logs the error payload", () => {
    finishProfilerWithError(createProfiler("POST", "/api/boom"), new Error("x"));

    expect(logger.error).toHaveBeenCalledOnce();
    const payload = logger.error.mock.calls[0]![0];
    expect(payload).toMatchObject({ method: "POST", path: "/api/boom" });
    expect(payload.error).toBeInstanceOf(Error);
  });
});
