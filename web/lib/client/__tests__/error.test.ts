import { AxiosError } from "axios";
import { describe, expect, it } from "vitest";
import { isNetworkError, isServerError } from "../error";

describe("isServerError", () => {
  it("returns true for 500+ axios errors", () => {
    const error = new AxiosError("fail", "500", undefined, undefined, {
      status: 500,
    } as never);
    expect(isServerError(error)).toBe(true);
  });

  it("returns false for 4xx errors", () => {
    const error = new AxiosError("fail", "400", undefined, undefined, {
      status: 400,
    } as never);
    expect(isServerError(error)).toBe(false);
  });

  it("returns false for non-axios errors", () => {
    expect(isServerError(new Error("plain"))).toBe(false);
  });
});

describe("isNetworkError", () => {
  it("returns true for non-axios errors", () => {
    expect(isNetworkError(new Error("plain"))).toBe(true);
  });

  it("returns false for axios errors", () => {
    const error = new AxiosError("fail");
    expect(isNetworkError(error)).toBe(false);
  });
});
