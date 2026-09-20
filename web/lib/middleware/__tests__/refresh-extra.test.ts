// @vitest-environment node
import { HttpResponse, http } from "msw";
import { NextResponse } from "next/server";
import { describe, expect, it, vi } from "vitest";
import { buildRequest, server } from "@/lib/tests";
import { applySessionCookiesToResponse, refreshSessionInMiddleware } from "../refresh";

function makeJwt(exp: number): string {
  const enc = (obj: unknown) =>
    btoa(JSON.stringify(obj)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
  return `${enc({ alg: "HS256" })}.${enc({ exp })}.sig`;
}

describe("refreshSessionInMiddleware extra", () => {
  it("does nothing when no tokens are present", async () => {
    const spy = vi.fn();
    server.use(http.post("*/api/v1/auth/refresh", spy));

    const request = buildRequest({ url: "http://localhost/onboarding" });
    const result = await refreshSessionInMiddleware(request);

    expect(result).toEqual({ refreshed: false, failed: false });
    expect(spy).not.toHaveBeenCalled();
  });

  it("refreshes when the access token is missing but a refresh token exists", async () => {
    server.use(
      http.post("*/api/v1/auth/refresh", () =>
        HttpResponse.json({
          success: true,
          data: { access_token: "fresh-access", refresh_token: "fresh-refresh" },
        }),
      ),
    );

    const request = buildRequest({
      url: "http://localhost/onboarding",
      headers: { cookie: "refresh_token=old-refresh" },
    });
    const result = await refreshSessionInMiddleware(request);

    expect(result).toEqual({
      refreshed: true,
      failed: false,
      tokens: { accessToken: "fresh-access", refreshToken: "fresh-refresh" },
    });
  });

  it("does not eagerly refresh when the access token is malformed", async () => {
    const spy = vi.fn();
    server.use(http.post("*/api/v1/auth/refresh", spy));

    const request = buildRequest({
      url: "http://localhost/onboarding",
      headers: { cookie: "access_token=not-a-jwt; refresh_token=old-refresh" },
    });
    const result = await refreshSessionInMiddleware(request);

    expect(result).toEqual({ refreshed: false, failed: false });
    expect(spy).not.toHaveBeenCalled();
    expect(request.cookies.get("access_token")?.value).toBe("not-a-jwt");
  });

  it("treats a network failure as neither refreshed nor failed", async () => {
    server.use(http.post("*/api/v1/auth/refresh", () => HttpResponse.error()));

    const request = buildRequest({
      url: "http://localhost/onboarding",
      headers: { cookie: `access_token=${makeJwt(1)}; refresh_token=old-refresh` },
    });
    const result = await refreshSessionInMiddleware(request);

    expect(result).toEqual({ refreshed: false, failed: false });
  });

  it("treats an invalid refresh payload as failed and clears cookies", async () => {
    server.use(
      http.post("*/api/v1/auth/refresh", () =>
        HttpResponse.json({ success: true, data: { access_token: "only-access" } }),
      ),
    );

    const request = buildRequest({
      url: "http://localhost/onboarding",
      headers: { cookie: `access_token=${makeJwt(1)}; refresh_token=old-refresh` },
    });
    const result = await refreshSessionInMiddleware(request);

    expect(result.failed).toBe(true);
    expect(request.cookies.get("access_token")).toBeUndefined();
  });

  it("shares one upstream refresh between concurrent requests", async () => {
    let refreshCalls = 0;
    server.use(
      http.post("*/api/v1/auth/refresh", async () => {
        refreshCalls += 1;
        await new Promise((resolve) => setTimeout(resolve, 20));
        return HttpResponse.json({
          success: true,
          data: { access_token: "shared-access", refresh_token: "shared-refresh" },
        });
      }),
    );

    const build = () =>
      buildRequest({
        url: "http://localhost/onboarding",
        headers: { cookie: `access_token=${makeJwt(1)}; refresh_token=shared-token` },
      });
    const [first, second] = await Promise.all([
      refreshSessionInMiddleware(build()),
      refreshSessionInMiddleware(build()),
    ]);

    expect(refreshCalls).toBe(1);
    expect(first.refreshed).toBe(true);
    expect(second.refreshed).toBe(true);
    expect(second.tokens).toEqual(first.tokens);
  });

  it("leaves cookies untouched when the token was already reused", async () => {
    server.use(
      http.post(
        "*/api/v1/auth/refresh",
        () =>
          new HttpResponse(
            JSON.stringify({
              success: false,
              message: "Refresh token already used.",
              error_code: "ERR_TOKEN_REUSED",
            }),
            { status: 401, headers: { "content-type": "application/json" } },
          ),
      ),
    );

    const request = buildRequest({
      url: "http://localhost/onboarding",
      headers: { cookie: `access_token=${makeJwt(1)}; refresh_token=stale-refresh` },
    });
    const result = await refreshSessionInMiddleware(request);

    expect(result).toEqual({ refreshed: false, failed: false });
    expect(request.cookies.get("access_token")?.value).toBe(makeJwt(1));
    expect(request.cookies.get("refresh_token")?.value).toBe("stale-refresh");
  });
});

describe("applySessionCookiesToResponse", () => {
  it("sets session cookies after a refresh", () => {
    const response = new NextResponse("ok");
    applySessionCookiesToResponse(response, {
      refreshed: true,
      failed: false,
      tokens: { accessToken: "a", refreshToken: "r" },
    });
    expect(response.cookies.get("access_token")?.value).toBe("a");
    expect(response.cookies.get("refresh_token")?.value).toBe("r");
  });

  it("clears session cookies after a failure", () => {
    const response = new NextResponse("ok");
    applySessionCookiesToResponse(response, { refreshed: false, failed: true });
    expect(response.cookies.get("access_token")?.value).toBe("");
    expect(response.cookies.get("refresh_token")?.value).toBe("");
  });

  it("leaves cookies untouched when nothing happened", () => {
    const response = new NextResponse("ok");
    applySessionCookiesToResponse(response, { refreshed: false, failed: false });
    expect(response.cookies.get("access_token")).toBeUndefined();
  });
});
