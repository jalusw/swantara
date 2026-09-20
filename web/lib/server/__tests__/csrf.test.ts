// @vitest-environment node
import { NextRequest } from "next/server";
import { describe, expect, it } from "vitest";
import { isCsrfValid } from "@/lib/server/csrf";

function makeRequest(
  method: string,
  headers: Record<string, string> = {},
  url = "http://localhost/api/test",
) {
  return new NextRequest(url, { method, headers });
}

describe("isCsrfValid", () => {
  it("allows GET requests without validation", () => {
    expect(isCsrfValid(makeRequest("GET"))).toBe(true);
  });

  it("rejects POST without CSRF header", () => {
    expect(isCsrfValid(makeRequest("POST"))).toBe(false);
  });

  it("rejects POST when origin matches host but token missing", () => {
    expect(
      isCsrfValid(
        makeRequest("POST", {
          origin: "http://localhost",
          host: "localhost",
        }),
      ),
    ).toBe(false);
  });

  it.each([
    "/api/v1/auth/login",
    "/api/v1/auth/register",
    "/api/v1/auth/email/check",
    "/api/v1/auth/email-verification/request",
    "/api/v1/auth/email-verification/verify",
    "/api/v1/auth/password-reset/request",
    "/api/v1/auth/password-reset",
  ])("allows unauthenticated POST to %s without CSRF tokens", (pathname) => {
    expect(isCsrfValid(makeRequest("POST", {}, `http://localhost${pathname}`))).toBe(true);
  });

  it("still rejects POST to refresh and logout without CSRF tokens", () => {
    expect(isCsrfValid(makeRequest("POST", {}, "http://localhost/api/v1/auth/refresh"))).toBe(
      false,
    );
    expect(isCsrfValid(makeRequest("POST", {}, "http://localhost/api/v1/auth/logout"))).toBe(false);
  });
});
