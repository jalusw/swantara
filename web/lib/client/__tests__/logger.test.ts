import { describe, expect, it } from "vitest";
import { logger } from "@/lib/utils/logger";
import { axiosInstance } from "../axios";

describe("axios instance", () => {
  it("is configured to talk to the API prefix with JSON", () => {
    expect(axiosInstance.defaults.baseURL).toBe("/api/v1");
    expect(axiosInstance.defaults.headers["Content-Type"]).toBe("application/json");
  });

  it("sends credentials and keeps a generous timeout", () => {
    expect(axiosInstance.defaults.withCredentials).toBe(true);
    expect(axiosInstance.defaults.timeout).toBe(1000 * 45);
  });
});

describe("client logger", () => {
  it("defaults to the info level when NEXT_PUBLIC_LOG_LEVEL is unset", () => {
    expect(logger.level).toBe("info");
  });

  it("can emit log lines without throwing", () => {
    expect(() => {
      logger.info("hello");
      logger.warn("careful");
      logger.error("boom");
    }).not.toThrow();
  });
});
