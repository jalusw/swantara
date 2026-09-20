// @vitest-environment node

import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";
import { buildRequest, server } from "@/lib/tests";
import { refreshSessionInMiddleware } from "../refresh";

function makeJwt(exp: number): string {
  const enc = (obj: unknown) =>
    btoa(JSON.stringify(obj)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
  return `${enc({ alg: "HS256" })}.${enc({ exp })}.sig`;
}

describe("refreshSessionInMiddleware", () => {
  it("refreshes and updates the request cookies when the access token is expired", async () => {
    server.use(
      http.post("*/api/v1/auth/refresh", () =>
        HttpResponse.json({
          success: true,
          data: { access_token: "new-access", refresh_token: "new-refresh" },
        }),
      ),
    );

    const request = buildRequest({
      url: "http://localhost/onboarding",
      headers: {
        cookie: `access_token=${makeJwt(1)}; refresh_token=old-refresh`,
      },
    });

    const result = await refreshSessionInMiddleware(request);

    expect(result.refreshed).toBe(true);
    expect(result.failed).toBe(false);
    expect(request.cookies.get("access_token")?.value).toBe("new-access");
    expect(request.cookies.get("refresh_token")?.value).toBe("new-refresh");
    expect(request.headers.get("cookie")).toContain("access_token=new-access");
  });

  it("does not refresh when the access token is still valid", async () => {
    const spy = vi.fn();
    server.use(http.post("*/api/v1/auth/refresh", spy));

    const request = buildRequest({
      url: "http://localhost/onboarding",
      headers: {
        cookie: `access_token=${makeJwt(Math.floor(Date.now() / 1000) + 3600)}; refresh_token=old`,
      },
    });

    const result = await refreshSessionInMiddleware(request);

    expect(result).toEqual({ refreshed: false, failed: false });
    expect(spy).not.toHaveBeenCalled();
  });

  it("clears the request cookies when the refresh endpoint rejects the token", async () => {
    server.use(http.post("*/api/v1/auth/refresh", () => new HttpResponse(null, { status: 401 })));

    const request = buildRequest({
      url: "http://localhost/onboarding",
      headers: {
        cookie: `access_token=${makeJwt(1)}; refresh_token=old`,
      },
    });

    const result = await refreshSessionInMiddleware(request);

    expect(result.refreshed).toBe(false);
    expect(result.failed).toBe(true);
    expect(request.cookies.get("access_token")).toBeUndefined();
    expect(request.cookies.get("refresh_token")).toBeUndefined();
  });
});
