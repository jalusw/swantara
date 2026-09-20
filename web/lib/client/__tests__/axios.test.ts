import { describe, expect, it } from "vitest";
import { axiosInstance } from "../axios";

describe("axiosInstance", () => {
  it("is defined", () => {
    expect(axiosInstance).toBeDefined();
  });

  it("has baseURL set to /api/v1", () => {
    expect(axiosInstance.defaults.baseURL).toBe("/api/v1");
  });

  it("has withCredentials enabled", () => {
    expect(axiosInstance.defaults.withCredentials).toBe(true);
  });

  it("has a response interceptor", () => {
    expect(axiosInstance.interceptors.response.handlers?.length).toBeGreaterThan(0);
  });
});
