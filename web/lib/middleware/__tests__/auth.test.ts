import { NextRequest } from "next/server";
import { describe, expect, it } from "vitest";
import { accessTokenKey } from "@/lib/constants/cookies";
import { authMiddleware } from "../auth";

function buildRequest(url: string, accessToken?: string) {
  return new NextRequest(url, {
    headers: accessToken ? { cookie: `${accessTokenKey}=${accessToken}` } : {},
  });
}

describe("authMiddleware", () => {
  it("redirects an anonymous user on the root to /login", async () => {
    const result = await authMiddleware(buildRequest("http://localhost"));

    expect(result?.status).toBe(307);
    expect(result?.headers.get("location")).toContain("/login");
  });

  it("allows the login page through when unauthenticated", async () => {
    const result = await authMiddleware(buildRequest("http://localhost/login"));

    expect(result).toBeNull();
  });

  it("redirects a logged-in guest away from /login to /onboarding", async () => {
    const result = await authMiddleware(buildRequest("http://localhost/login", "token"));

    expect(result?.status).toBe(307);
    expect(result?.headers.get("location")).toContain("/onboarding");
  });

  it("redirects a logged-in guest away from /register", async () => {
    const result = await authMiddleware(buildRequest("http://localhost/register", "token"));

    expect(result?.headers.get("location")).toContain("/onboarding");
  });

  it("redirects an anonymous user from a protected path to /login", async () => {
    const result = await authMiddleware(buildRequest("http://localhost/onboarding"));

    expect(result?.status).toBe(307);
    expect(result?.headers.get("location")).toContain("/login");
  });

  it("redirects an anonymous user from an organization path to /login", async () => {
    const result = await authMiddleware(buildRequest("http://localhost/org/1/dashboard"));

    expect(result?.status).toBe(307);
    expect(result?.headers.get("location")).toContain("/login");
  });

  it("lets an authenticated user reach a protected path", async () => {
    const result = await authMiddleware(buildRequest("http://localhost/onboarding", "token"));

    expect(result).toBeNull();
  });

  it("redirects authenticated users to onboarding", async () => {
    const result = await authMiddleware(buildRequest("http://localhost/login", "token"));

    expect(result?.headers.get("location")).toContain("/onboarding");
  });

  it("sends an authenticated user on the root to onboarding", async () => {
    const result = await authMiddleware(buildRequest("http://localhost", "token"));

    expect(result?.status).toBe(307);
    expect(result?.headers.get("location")).toContain("/onboarding");
  });
});
