import type { NextRequest, NextResponse } from "next/server";
import { accessTokenKey, refreshTokenKey } from "@/lib/constants/cookies";
import { serverConfig } from "@/lib/server/config";
import { dedupRefresh } from "@/lib/server/refresh-shared";
import {
  clearSessionCookies,
  getAccessToken,
  getRefreshToken,
  REFRESH_TOKEN_REUSED_CODE,
  type SessionTokens,
  setSessionCookies,
} from "@/lib/server/session";
import { toCamelCase } from "@/lib/utils/case";

const AUTH_REFRESH_PATH = "/api/v1/auth/refresh";
const UPSTREAM_TIMEOUT_MS = 10_000;
const EXPIRY_SKEW_SECONDS = 30;

export type MiddlewareRefreshResult = {
  refreshed: boolean;
  failed: boolean;
  tokens?: SessionTokens;
};

function decodeExpiry(token: string): number | null {
  const segments = token.split(".");
  if (segments.length < 2) return null;
  try {
    let normalized = segments[1]!.replace(/-/g, "+").replace(/_/g, "/");
    while (normalized.length % 4 !== 0) normalized += "=";
    const json = JSON.parse(atob(normalized));
    return typeof json.exp === "number" ? json.exp : null;
  } catch {
    return null;
  }
}

function needsRefresh(request: NextRequest): boolean {
  const accessToken = getAccessToken(request);
  const refreshToken = getRefreshToken(request);
  if (!accessToken) return Boolean(refreshToken);
  const expiry = decodeExpiry(accessToken);
  if (expiry === null) return false;
  return Math.floor(Date.now() / 1000) >= expiry - EXPIRY_SKEW_SECONDS;
}

type RefreshCallResult =
  | { ok: true; tokens: SessionTokens }
  | { ok: false; networkError: boolean; reused: boolean };

async function callUpstreamRefresh(refreshTokenValue: string): Promise<RefreshCallResult> {
  const target = new URL(AUTH_REFRESH_PATH, serverConfig.serviceApiUrl);
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), UPSTREAM_TIMEOUT_MS);
  try {
    const response = await fetch(target, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshTokenValue }),
      signal: controller.signal,
    });
    if (!response.ok) return { ok: false, networkError: false, reused: await isReuse(response) };
    let json: unknown;
    try {
      json = await response.json();
    } catch {
      return { ok: false, networkError: false, reused: false };
    }
    const raw = toCamelCase(json) as { data?: Record<string, unknown> };
    const rawData = raw.data as
      | {
          accessToken?: string;
          refreshToken?: string;
          access_token?: string;
          refresh_token?: string;
        }
      | undefined;
    const accessToken = rawData?.accessToken ?? rawData?.access_token;
    const refreshToken = rawData?.refreshToken ?? rawData?.refresh_token;
    if (accessToken && refreshToken) {
      return { ok: true, tokens: { accessToken, refreshToken } };
    }
    return { ok: false, networkError: false, reused: false };
  } catch {
    return { ok: false, networkError: true, reused: false };
  } finally {
    clearTimeout(timeout);
  }
}

async function isReuse(response: Response): Promise<boolean> {
  try {
    const raw = (await response.json()) as { errorCode?: string; error_code?: string };
    const code = raw.errorCode ?? raw.error_code;
    return code === REFRESH_TOKEN_REUSED_CODE;
  } catch {
    return false;
  }
}

function callSharedRefresh(refreshTokenValue: string): Promise<RefreshCallResult> {
  return dedupRefresh(`proxy:${refreshTokenValue}`, () => callUpstreamRefresh(refreshTokenValue));
}

function applyRequestCookies(request: NextRequest, tokens: SessionTokens | null): void {
  if (tokens) {
    request.cookies.set(accessTokenKey, tokens.accessToken);
    request.cookies.set(refreshTokenKey, tokens.refreshToken);
  } else {
    request.cookies.delete(accessTokenKey);
    request.cookies.delete(refreshTokenKey);
  }
  request.headers.set("cookie", request.cookies.toString());
}

export async function refreshSessionInMiddleware(
  request: NextRequest,
): Promise<MiddlewareRefreshResult> {
  if (!needsRefresh(request)) return { refreshed: false, failed: false };
  const refreshTokenValue = getRefreshToken(request);
  if (!refreshTokenValue) return { refreshed: false, failed: false };

  const result = await callSharedRefresh(refreshTokenValue);
  if (result.ok) {
    applyRequestCookies(request, result.tokens);
    return { refreshed: true, failed: false, tokens: result.tokens };
  }

  if (result.networkError || result.reused) {
    return { refreshed: false, failed: false };
  }

  applyRequestCookies(request, null);
  return { refreshed: false, failed: true };
}

export function applySessionCookiesToResponse(
  response: NextResponse,
  result: MiddlewareRefreshResult,
): void {
  if (result.failed) {
    clearSessionCookies(response);
  } else if (result.refreshed && result.tokens) {
    setSessionCookies(response, result.tokens);
  }
}
