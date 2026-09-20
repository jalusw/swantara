import type { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import { describe, expect, it, vi } from "vitest";
import { logger } from "@/lib/utils/logger";
import { finishErrorProfile, finishResponseProfile, startRequestProfile } from "../profiler";

function buildConfig(overrides: Partial<InternalAxiosRequestConfig> = {}) {
  return { method: "get", url: "/api/v1/me", ...overrides } as InternalAxiosRequestConfig;
}

describe("request profiler", () => {
  it("logs a completed response without throwing", () => {
    const info = vi.spyOn(logger, "info").mockImplementation(() => {});
    const config = buildConfig();
    startRequestProfile(config);
    finishResponseProfile({ config, status: 200 } as AxiosResponse);
    expect(info).toHaveBeenCalledOnce();
    info.mockRestore();
  });

  it("ignores responses without a started profile", () => {
    const info = vi.spyOn(logger, "info").mockImplementation(() => {});
    finishResponseProfile({ config: buildConfig(), status: 200 } as AxiosResponse);
    expect(info).not.toHaveBeenCalled();
    info.mockRestore();
  });

  it("logs a failed request as an error", () => {
    const error = vi.spyOn(logger, "error").mockImplementation(() => {});
    const config = buildConfig({ method: "post", url: "/api/v1/auth/login" });
    startRequestProfile(config);
    finishErrorProfile({ config, response: { status: 500 } } as unknown as AxiosError);
    expect(error).toHaveBeenCalledOnce();
    error.mockRestore();
  });

  it("ignores errors without a config or profile", () => {
    const error = vi.spyOn(logger, "error").mockImplementation(() => {});
    finishErrorProfile({ response: { status: 500 } } as unknown as AxiosError);
    finishErrorProfile({ config: buildConfig() } as unknown as AxiosError);
    expect(error).not.toHaveBeenCalled();
    error.mockRestore();
  });

  it("defaults missing method and url", () => {
    const info = vi.spyOn(logger, "info").mockImplementation(() => {});
    const config = {} as InternalAxiosRequestConfig;
    startRequestProfile(config);
    finishResponseProfile({ config, status: 200 } as AxiosResponse);
    expect(info).toHaveBeenCalledOnce();
    info.mockRestore();
  });
});
