import type { NextRequest, NextResponse } from "next/server";
import { accessTokenKey, csrfKey, refreshTokenKey } from "@/lib/constants/cookies";

export const ACCESS_TOKEN_MAX_AGE = 30 * 60;
export const REFRESH_TOKEN_MAX_AGE = 48 * 60 * 60;
export const REFRESH_TOKEN_REUSED_CODE = "ERR_TOKEN_REUSED";

export const baseCookieOptions = {
  httpOnly: true,
  secure: process.env.NODE_ENV === "production",
  sameSite: "lax",
  path: "/",
} as const;

export type SessionTokens = {
  accessToken: string;
  refreshToken: string;
};

export function getAccessToken(request: NextRequest): string | undefined {
  return request.cookies.get(accessTokenKey)?.value;
}

export function getRefreshToken(request: NextRequest): string | undefined {
  return request.cookies.get(refreshTokenKey)?.value;
}

export function setSessionCookies(response: NextResponse, tokens: SessionTokens): void {
  response.cookies.set(accessTokenKey, tokens.accessToken, {
    ...baseCookieOptions,
    maxAge: ACCESS_TOKEN_MAX_AGE,
  });
  response.cookies.set(refreshTokenKey, tokens.refreshToken, {
    ...baseCookieOptions,
    maxAge: REFRESH_TOKEN_MAX_AGE,
  });
}

export function clearSessionCookies(response: NextResponse): void {
  response.cookies.set(accessTokenKey, "", {
    ...baseCookieOptions,
    maxAge: 0,
  });
  response.cookies.set(refreshTokenKey, "", {
    ...baseCookieOptions,
    maxAge: 0,
  });
  response.cookies.set(csrfKey, "", {
    ...baseCookieOptions,
    httpOnly: false,
    maxAge: 0,
  });
}
