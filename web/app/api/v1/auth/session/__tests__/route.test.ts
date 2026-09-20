// @vitest-environment node
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { GET } from "@/app/api/v1/auth/session/route";
import { buildRequest, server } from "@/lib/tests";

describe("session route GET", () => {
  it("redirects to login without mutating cookies", async () => {
    const response = await GET(buildRequest({ url: "http://localhost/api/v1/auth/session" }));

    expect(response.status).toBe(307);
    expect(response.headers.get("location")).toBe("http://localhost/login");
    expect(response.cookies.get("access_token")).toBeUndefined();
    expect(response.cookies.get("refresh_token")).toBeUndefined();
  });

  it("redirects to login", async () => {
    const response = await GET(buildRequest({ url: "http://localhost/api/v1/auth/session" }));

    expect(response.headers.get("location")).toBe("http://localhost/login");
  });

  it("refreshes the session and redirects to next when the refresh token is valid", async () => {
    server.use(
      http.post("*/api/v1/auth/refresh", () =>
        HttpResponse.json({
          success: true,
          data: { access_token: "new-access", refresh_token: "new-refresh" },
        }),
      ),
    );

    const response = await GET(
      buildRequest({
        url: "http://localhost/api/v1/auth/session?next=/onboarding",
        headers: { cookie: "refresh_token=old-refresh" },
      }),
    );

    expect(response.headers.get("location")).toBe("http://localhost/onboarding");
    expect(response.cookies.get("access_token")?.value).toBe("new-access");
    expect(response.cookies.get("refresh_token")?.value).toBe("new-refresh");
  });

  it("redirects to login when refresh is rejected without mutating cookies", async () => {
    server.use(http.post("*/api/v1/auth/refresh", () => new HttpResponse(null, { status: 401 })));

    const response = await GET(
      buildRequest({
        url: "http://localhost/api/v1/auth/session?next=/onboarding",
        headers: { cookie: "refresh_token=reused-token" },
      }),
    );

    expect(response.headers.get("location")).toBe("http://localhost/login");
    expect(response.cookies.get("access_token")).toBeUndefined();
  });

  it("ignores absolute next URLs", async () => {
    server.use(
      http.post("*/api/v1/auth/refresh", () =>
        HttpResponse.json({
          success: true,
          data: { access_token: "new-access", refresh_token: "new-refresh" },
        }),
      ),
    );

    const response = await GET(
      buildRequest({
        url: "http://localhost/api/v1/auth/session?next=https://evil.test/phish",
        headers: { cookie: "refresh_token=old-refresh" },
      }),
    );

    expect(response.headers.get("location")).toBe("http://localhost/onboarding");
  });
});
