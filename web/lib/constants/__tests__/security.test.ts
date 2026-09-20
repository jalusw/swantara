import { describe, expect, it } from "vitest";
import { csrfHeader, csrfToken } from "@/lib/constants/security";

describe("security constants", () => {
  it("defines the CSRF header name", () => {
    expect(csrfHeader).toBe("x-csrf-token");
  });

  it("defines the CSRF token value", () => {
    expect(csrfToken).toBe("test-csrf-token");
  });
});
