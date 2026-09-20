import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  finishErrorProfile,
  finishResponseProfile,
  startRequestProfile,
} from "@/lib/utils/profiler";

const { logger } = vi.hoisted(() => ({
  logger: {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("@/lib/utils/logger", () => ({ logger }));

function makeConfig(url: string) {
  return {
    method: "get",
    url,
    data: undefined,
  } as RenderableRequestConfig;
}

type RenderableRequestConfig = Parameters<typeof startRequestProfile>[0];

function makeResponse(config: RenderableRequestConfig, status: number) {
  return {
    config,
    status,
    data: {},
  } as Parameters<typeof finishResponseProfile>[0];
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("startRequestProfile", () => {
  it("records a start time for the request", () => {
    const config = makeConfig("/api/v1/me");

    startRequestProfile(config);

    expect(() => finishResponseProfile(makeResponse(config, 200))).not.toThrow();
    expect(logger.info).toHaveBeenCalledOnce();
  });
});

describe("finishResponseProfile", () => {
  it("logs a fast success response as info", () => {
    const config = makeConfig("/api/v1/me");
    startRequestProfile(config);

    finishResponseProfile(makeResponse(config, 200));

    expect(logger.info).toHaveBeenCalledOnce();
    expect(logger.warn).not.toHaveBeenCalled();
    expect(logger.error).not.toHaveBeenCalled();
    const payload = logger.info.mock.calls[0]![0];
    expect(payload).toEqual(
      expect.objectContaining({
        method: "GET",
        url: "/api/v1/me",
        status: 200,
      }),
    );
    expect(payload.duration).toMatch(/(ms|s)$/);
  });

  it("warns when the request took longer than one second", () => {
    vi.useFakeTimers();
    try {
      const config = makeConfig("/api/v1/slow");
      startRequestProfile(config);
      vi.advanceTimersByTime(1500);

      finishResponseProfile(makeResponse(config, 200));

      expect(logger.warn).toHaveBeenCalledOnce();
      expect(logger.info).not.toHaveBeenCalled();
    } finally {
      vi.useRealTimers();
    }
  });

  it("errors when the response status is 5xx", () => {
    const config = makeConfig("/api/v1/broken");
    startRequestProfile(config);

    finishResponseProfile(makeResponse(config, 500));

    expect(logger.error).toHaveBeenCalledOnce();
    expect(logger.info).not.toHaveBeenCalled();
  });

  it("logs nothing when no profile was started", () => {
    finishResponseProfile(makeResponse(makeConfig("/api/v1/me"), 200));

    expect(logger.info).not.toHaveBeenCalled();
    expect(logger.warn).not.toHaveBeenCalled();
    expect(logger.error).not.toHaveBeenCalled();
  });
});

describe("finishErrorProfile", () => {
  it("logs the error payload", () => {
    const config = makeConfig("/api/v1/boom");
    startRequestProfile(config);

    finishErrorProfile({
      config,
      response: { status: 503 },
    } as Parameters<typeof finishErrorProfile>[0]);

    expect(logger.error).toHaveBeenCalledOnce();
    const payload = logger.error.mock.calls[0]![0];
    expect(payload).toEqual(
      expect.objectContaining({
        method: "GET",
        url: "/api/v1/boom",
        status: 503,
      }),
    );
    expect(payload.error).toEqual(expect.objectContaining({ config }));
  });

  it("logs nothing when the config is missing", () => {
    finishErrorProfile({} as Parameters<typeof finishErrorProfile>[0]);

    expect(logger.error).not.toHaveBeenCalled();
  });
});
