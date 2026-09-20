// @vitest-environment node

import { HttpResponse, http } from "msw";
import { NextRequest, type NextResponse } from "next/server";
import { describe, expect, it } from "vitest";
import { csrfToken } from "@/lib/constants/security";
import { buildRequest, server } from "@/lib/tests";
import { proxyRequest } from "../proxy";

describe("proxyRequest", () => {
  it("forwards a GET and maps the upstream payload to camelCase", async () => {
    server.use(
      http.get("*/api/v1/me/organizations", () =>
        HttpResponse.json({
          success: true,
          message: "Organizations retrieved successfully.",
          data: {
            organizations: [{ id: 1, name: "Acme", legal_name: "Acme Inc" }],
          },
        }),
      ),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/me/organizations",
    });

    const response = await proxyRequest(request, ["me", "organizations"]);
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data.organizations[0]).toEqual({
      id: 1,
      name: "Acme",
      legalName: "Acme Inc",
    });
  });

  it("passes the access token as a bearer header", async () => {
    let capturedAuthorization: string | null = null;
    server.use(
      http.get("*/api/v1/me", ({ request }) => {
        capturedAuthorization = request.headers.get("authorization");
        return HttpResponse.json({
          success: true,
          message: "Profile retrieved.",
          data: { id: 1, email: "demo@swantara.local" },
        });
      }),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/me",
      headers: { cookie: "access_token=test-access-token" },
    });

    await proxyRequest(request, ["me"]);

    expect(capturedAuthorization).toBe("Bearer test-access-token");
  });

  it("snake_cases the request body before forwarding", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.post("*/api/v1/auth/login", async ({ request }) => {
        capturedBody = await request.json();
        return HttpResponse.json({
          success: true,
          message: "Authenticated.",
          data: {
            access_token: "new-access",
            refresh_token: "new-refresh",
          },
        });
      }),
    );

    const request = buildRequest({
      method: "POST",
      url: "http://localhost/api/v1/auth/login",
      headers: { "x-csrf-token": csrfToken, cookie: "csrf_token=test-csrf-token" },
      body: { email: "demo@swantara.local", password: "secret" },
    });

    await proxyRequest(request, ["auth", "login"]);

    expect(capturedBody).toEqual({
      email: "demo@swantara.local",
      password: "secret",
    });
  });

  it("sets session cookies from a successful login response", async () => {
    server.use(
      http.post("*/api/v1/auth/login", () =>
        HttpResponse.json({
          success: true,
          message: "Authenticated.",
          data: {
            access_token: "new-access",
            refresh_token: "new-refresh",
          },
        }),
      ),
    );

    const request = buildRequest({
      method: "POST",
      url: "http://localhost/api/v1/auth/login",
      headers: { "x-csrf-token": csrfToken, cookie: "csrf_token=test-csrf-token" },
      body: { email: "demo@swantara.local", password: "secret" },
    });

    const response = (await proxyRequest(request, ["auth", "login"])) as NextResponse;

    expect(response.cookies.get("access_token")?.value).toBe("new-access");
    expect(response.cookies.get("refresh_token")?.value).toBe("new-refresh");
  });

  it("injects the refresh token into a logout request and clears cookies", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.post("*/api/v1/auth/logout", async ({ request }) => {
        capturedBody = await request.json();
        return HttpResponse.json({
          success: true,
          message: "Logout success.",
        });
      }),
    );

    const request = buildRequest({
      method: "POST",
      url: "http://localhost/api/v1/auth/logout",
      headers: {
        cookie: "access_token=old-access; refresh_token=old-refresh; csrf_token=test-csrf-token",
        "x-csrf-token": csrfToken,
      },
      body: {},
    });

    const response = (await proxyRequest(request, ["auth", "logout"])) as NextResponse;

    expect(capturedBody).toEqual({ refresh_token: "old-refresh" });
    expect(response.cookies.get("access_token")?.value).toBe("");
    expect(response.cookies.get("refresh_token")?.value).toBe("");
  });

  it("returns a service unavailable response when the upstream is unreachable", async () => {
    server.use(http.get("*/api/v1/health", () => HttpResponse.error()));

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/health",
    });

    const response = await proxyRequest(request, ["health"]);
    const body = await response.json();

    expect(response.status).toBe(503);
    expect(body.success).toBe(false);
    expect(body.errorCode).toBe("ERR_SERVICE_UNAVAILABLE");
  });

  it("preserves the upstream status code and envelope on errors", async () => {
    server.use(
      http.get("*/api/v1/me", () =>
        HttpResponse.json({ success: false, message: "Unauthorized." }, { status: 401 }),
      ),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/me",
    });

    const response = await proxyRequest(request, ["me"]);
    const body = await response.json();

    expect(response.status).toBe(401);
    expect(body.message).toBe("Unauthorized.");
  });

  it("forwards multipart/form-data bodies verbatim (file uploads)", async () => {
    let capturedContentType: string | null = null;
    let capturedFileName: string | undefined;
    let capturedField: string | undefined;
    server.use(
      http.post("*/api/v1/attachments", async ({ request }) => {
        capturedContentType = request.headers.get("content-type");
        const form = await request.formData();
        const file = form.get("file");
        capturedField = form.get("kind") as string | undefined;
        if (file instanceof File) capturedFileName = file.name;
        return HttpResponse.json({
          success: true,
          message: "Uploaded.",
          data: { fileName: capturedFileName },
        });
      }),
    );

    const form = new FormData();
    form.append("file", new Blob(["binary-content"], { type: "text/plain" }), "note.txt");
    form.append("kind", "avatar");

    const request = new NextRequest("http://localhost/api/v1/attachments", {
      method: "POST",
      headers: { "x-csrf-token": csrfToken, cookie: "csrf_token=test-csrf-token" },
      body: form,
    });

    const response = await proxyRequest(request, ["attachments"]);
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(capturedContentType).toContain("multipart/form-data");
    expect(capturedFileName).toBe("note.txt");
    expect(capturedField).toBe("avatar");
    expect(body.data.fileName).toBe("note.txt");
  });

  it("forwards non-JSON response bodies and upstream metadata headers", async () => {
    server.use(
      http.get(
        "*/api/v1/reports/export",
        () =>
          new HttpResponse("id,name\n1,Acme", {
            headers: {
              "content-type": "text/csv",
              "x-total-count": "42",
              link: "<http://localhost/api/v1/reports/export?page=2>; rel=next",
            },
          }),
      ),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/reports/export",
    });

    const response = await proxyRequest(request, ["reports", "export"]);

    expect(response.headers.get("content-type")).toBe("text/csv");
    expect(response.headers.get("x-total-count")).toBe("42");
    expect(response.headers.get("link")).toContain("rel=next");
    expect(await response.text()).toBe("id,name\n1,Acme");
  });

  it("does not leak upstream set-cookie headers", async () => {
    server.use(
      http.get("*/api/v1/me", () =>
        HttpResponse.json(
          { success: true, message: "ok", data: {} },
          { headers: { "set-cookie": "evil=1; Path=/;" } },
        ),
      ),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/me",
    });

    const response = await proxyRequest(request, ["me"]);

    expect(response.headers.get("set-cookie")).toBeNull();
  });

  it("refreshes the session and retries once on a 401", async () => {
    let callCount = 0;
    let retriedWithToken: string | null = null;
    server.use(
      http.post("*/api/v1/auth/refresh", () =>
        HttpResponse.json({
          success: true,
          message: "",
          data: { access_token: "new-access", refresh_token: "new-refresh" },
        }),
      ),
      http.get("*/api/v1/me/organizations", ({ request }) => {
        callCount += 1;
        if (callCount === 1) {
          return new HttpResponse(null, { status: 401 });
        }
        retriedWithToken = request.headers.get("authorization");
        return HttpResponse.json({
          success: true,
          message: "",
          data: { organizations: [] },
        });
      }),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/me/organizations",
      headers: {
        cookie: "access_token=old; refresh_token=old-refresh; csrf_token=test-csrf-token",
        "x-csrf-token": csrfToken,
      },
    });

    const response = (await proxyRequest(request, ["me", "organizations"])) as NextResponse;

    expect(response.status).toBe(200);
    expect(retriedWithToken).toBe("Bearer new-access");
    expect(response.cookies.get("access_token")?.value).toBe("new-access");
    expect(response.cookies.get("refresh_token")?.value).toBe("new-refresh");
  });

  it("deduplicates concurrent 401 refreshes into a single call", async () => {
    let refreshCalls = 0;
    let callCount = 0;
    let retriedWithToken: string | null = null;
    server.use(
      http.post("*/api/v1/auth/refresh", () => {
        refreshCalls += 1;
        return HttpResponse.json({
          success: true,
          message: "",
          data: { access_token: "a2", refresh_token: "r2" },
        });
      }),
      http.get("*/api/v1/me", ({ request }) => {
        callCount += 1;
        if (callCount <= 2) {
          return new HttpResponse(null, { status: 401 });
        }
        retriedWithToken = request.headers.get("authorization");
        return HttpResponse.json({
          success: true,
          message: "",
          data: { id: 1 },
        });
      }),
    );

    const build = () =>
      buildRequest({
        method: "GET",
        url: "http://localhost/api/v1/me",
        headers: {
          cookie: "access_token=old; refresh_token=old; csrf_token=test-csrf-token",
          "x-csrf-token": csrfToken,
        },
      });

    const [r1, r2] = await Promise.all([
      proxyRequest(build(), ["me"]),
      proxyRequest(build(), ["me"]),
    ]);

    expect(refreshCalls).toBe(1);
    expect(r1.status).toBe(200);
    expect(r2.status).toBe(200);
    expect(retriedWithToken).toBe("Bearer a2");
  });

  it("clears cookies when the refresh fails", async () => {
    server.use(
      http.post("*/api/v1/auth/refresh", () => new HttpResponse(null, { status: 401 })),
      http.get("*/api/v1/me", () => new HttpResponse(null, { status: 401 })),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/me",
      headers: {
        cookie: "access_token=old; refresh_token=old; csrf_token=test-csrf-token",
        "x-csrf-token": csrfToken,
      },
    });

    const response = (await proxyRequest(request, ["me"])) as NextResponse;

    expect(response.status).toBe(401);
    expect(response.cookies.get("access_token")?.value).toBe("");
    expect(response.cookies.get("refresh_token")?.value).toBe("");
  });

  it("keeps cookies when the refresh token was already reused", async () => {
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
      http.get("*/api/v1/me", () => new HttpResponse(null, { status: 401 })),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/me",
      headers: {
        cookie: "access_token=old; refresh_token=stale; csrf_token=test-csrf-token",
        "x-csrf-token": csrfToken,
      },
    });

    const response = (await proxyRequest(request, ["me"])) as NextResponse;

    expect(response.status).toBe(401);
    expect(response.cookies.get("access_token")).toBeUndefined();
    expect(response.cookies.get("refresh_token")).toBeUndefined();
  });

  it("keeps cookies when the refresh call fails with a network error", async () => {
    server.use(
      http.post("*/api/v1/auth/refresh", () => HttpResponse.error()),
      http.get("*/api/v1/me", () => new HttpResponse(null, { status: 401 })),
    );

    const request = buildRequest({
      method: "GET",
      url: "http://localhost/api/v1/me",
      headers: {
        cookie: "access_token=old; refresh_token=old; csrf_token=test-csrf-token",
        "x-csrf-token": csrfToken,
      },
    });

    const response = (await proxyRequest(request, ["me"])) as NextResponse;

    expect(response.status).toBe(401);
    expect(response.cookies.get("access_token")).toBeUndefined();
    expect(response.cookies.get("refresh_token")).toBeUndefined();
  });

  it("rejects mutating requests missing the CSRF header", async () => {
    const request = new NextRequest("http://localhost/api/v1/products", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ name: "x" }),
    });

    const response = await proxyRequest(request, ["products"]);

    expect(response.status).toBe(403);
  });

  it("forwards unauthenticated register without CSRF tokens", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json({
          success: true,
          message: "Registered.",
          data: { id: 1 },
        }),
      ),
    );

    const request = buildRequest({
      method: "POST",
      url: "http://localhost/api/v1/auth/register",
      body: { email: "new@swantara.local", password: "secret123" },
    });

    const response = await proxyRequest(request, ["auth", "register"]);
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.message).toBe("Registered.");
  });

  it("forwards the refresh token cookie into the refresh request body", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.post("*/api/v1/auth/refresh", async ({ request }) => {
        capturedBody = await request.json();
        return HttpResponse.json({
          success: true,
          message: "",
          data: { access_token: "new-access", refresh_token: "new-refresh" },
        });
      }),
    );

    const request = buildRequest({
      method: "POST",
      url: "http://localhost/api/v1/auth/refresh",
      headers: {
        cookie: "refresh_token=old-refresh; csrf_token=test-csrf-token",
        "x-csrf-token": csrfToken,
      },
      body: {},
    });

    const response = (await proxyRequest(request, ["auth", "refresh"])) as NextResponse;

    expect(capturedBody).toEqual({ refresh_token: "old-refresh" });
    expect(response.cookies.get("access_token")?.value).toBe("new-access");
    expect(response.cookies.get("refresh_token")?.value).toBe("new-refresh");
  });
});
