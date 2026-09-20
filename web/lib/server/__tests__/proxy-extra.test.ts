// @vitest-environment node
import { HttpResponse, http } from "msw";
import type { NextResponse } from "next/server";
import { describe, expect, it } from "vitest";
import { csrfToken } from "@/lib/constants/security";
import { buildRequest, server } from "@/lib/tests";
import { proxyRequest } from "../proxy";

describe("proxyRequest extra", () => {
  it("forwards query parameters to the upstream", async () => {
    let capturedUrl: string | null = null;
    server.use(
      http.get("*/api/v1/products", ({ request }) => {
        capturedUrl = request.url;
        return HttpResponse.json({ success: true, message: "", data: { items: [] } });
      }),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/products?page=2&limit=10",
    });

    await proxyRequest(request, ["products"]);

    expect(capturedUrl).toContain("page=2");
    expect(capturedUrl).toContain("limit=10");
  });

  it("forwards the accept-language header", async () => {
    let capturedLanguage: string | null = null;
    server.use(
      http.get("*/api/v1/me", ({ request }) => {
        capturedLanguage = request.headers.get("accept-language");
        return HttpResponse.json({ success: true, message: "", data: {} });
      }),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/me",
      headers: { "accept-language": "id-ID" },
    });

    await proxyRequest(request, ["me"]);

    expect(capturedLanguage).toBe("id-ID");
  });

  it("maps an invalid JSON upstream body to ERR_SERVICE", async () => {
    server.use(
      http.get(
        "*/api/v1/me",
        () => new HttpResponse("not-json{{{", { headers: { "content-type": "application/json" } }),
      ),
    );

    const request = buildRequest({ method: "GET", url: "http://localhost/api/v1/me" });
    const response = await proxyRequest(request, ["me"]);
    const body = await response.json();

    expect(body.success).toBe(false);
    expect(body.errorCode).toBe("ERR_SERVICE");
  });

  it("does not attempt a refresh for a 401 on the login path", async () => {
    let refreshCalls = 0;
    server.use(
      http.post("*/api/v1/auth/login", () => new HttpResponse(null, { status: 401 })),
      http.post("*/api/v1/auth/refresh", () => {
        refreshCalls += 1;
        return HttpResponse.json({ success: true, message: "", data: {} });
      }),
    );

    const request = buildRequest({
      method: "POST",
      url: "http://localhost/api/v1/auth/login",
      headers: {
        cookie: "access_token=old; refresh_token=old-refresh; csrf_token=test-csrf-token",
        "x-csrf-token": csrfToken,
      },
      body: { email: "demo@swantara.local", password: "secret" },
    });

    const response = await proxyRequest(request, ["auth", "login"]);

    expect(response.status).toBe(401);
    expect(refreshCalls).toBe(0);
  });

  it("sets no session cookies when login succeeds without tokens", async () => {
    server.use(
      http.post("*/api/v1/auth/login", () =>
        HttpResponse.json({ success: true, message: "", data: {} }),
      ),
    );

    const request = buildRequest({
      method: "POST",
      url: "http://localhost/api/v1/auth/login",
      headers: { "x-csrf-token": csrfToken, cookie: "csrf_token=test-csrf-token" },
      body: { email: "demo@swantara.local", password: "secret" },
    });

    const response = (await proxyRequest(request, ["auth", "login"])) as NextResponse;

    expect(response.status).toBe(200);
    expect(response.cookies.get("access_token")).toBeUndefined();
    expect(response.cookies.get("refresh_token")).toBeUndefined();
  });

  it("returns an empty body for an empty JSON upstream response", async () => {
    server.use(
      http.get(
        "*/api/v1/me",
        () => new HttpResponse("", { headers: { "content-type": "application/json" } }),
      ),
    );

    const request = buildRequest({ method: "GET", url: "http://localhost/api/v1/me" });
    const response = await proxyRequest(request, ["me"]);

    expect(response.status).toBe(200);
    expect(await response.text()).toBe("");
  });
});
