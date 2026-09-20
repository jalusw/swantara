// @vitest-environment node
import { NextRequest, NextResponse } from "next/server";
import { describe, expect, it } from "vitest";
import {
  ACCESS_TOKEN_MAX_AGE,
  baseCookieOptions,
  clearSessionCookies,
  getAccessToken,
  getRefreshToken,
  REFRESH_TOKEN_MAX_AGE,
  setSessionCookies,
} from "@/lib/server/session";

function makeRequestWithCookie(name: string, value: string) {
  return new NextRequest("http://localhost", {
    headers: { cookie: `${name}=${value}` },
  });
}

describe("session constants", () => {
  it("access token max age is 30 minutes", () => {
    expect(ACCESS_TOKEN_MAX_AGE).toBe(30 * 60);
  });

  it("refresh token max age is 48 hours", () => {
    expect(REFRESH_TOKEN_MAX_AGE).toBe(48 * 60 * 60);
  });

  it("base cookie options are httpOnly and sameSite lax", () => {
    expect(baseCookieOptions.httpOnly).toBe(true);
    expect(baseCookieOptions.sameSite).toBe("lax");
  });
});

describe("getAccessToken / getRefreshToken", () => {
  it("reads access_token from cookies", () => {
    const req = makeRequestWithCookie("access_token", "tok123");
    expect(getAccessToken(req)).toBe("tok123");
  });

  it("reads refresh_token from cookies", () => {
    const req = makeRequestWithCookie("refresh_token", "ref456");
    expect(getRefreshToken(req)).toBe("ref456");
  });
});

describe("setSessionCookies", () => {
  it("sets both access and refresh cookies", () => {
    const response = NextResponse.next();
    setSessionCookies(response, { accessToken: "a", refreshToken: "r" });
    expect(response.cookies.get("access_token")?.value).toBe("a");
    expect(response.cookies.get("refresh_token")?.value).toBe("r");
  });
});

describe("clearSessionCookies", () => {
  it("clears both cookies by setting maxAge to 0", () => {
    const response = NextResponse.next();
    clearSessionCookies(response);
    expect(response.cookies.get("access_token")?.value).toBe("");
    expect(response.cookies.get("refresh_token")?.value).toBe("");
  });

  it("clears the csrf hint so the client stops firing authenticated queries", () => {
    const response = NextResponse.next();
    clearSessionCookies(response);
    expect(response.cookies.get("csrf_token")?.value).toBe("");
  });
});
