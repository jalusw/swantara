import { StatusCodes } from "http-status-codes";
import { type NextRequest, NextResponse } from "next/server";
import { csrfKey } from "@/lib/constants/cookies";
import { toCamelCase, toSnakeCase } from "@/lib/utils/case";
import { isJson, jsonContentType } from "@/lib/utils/http";
import { serverConfig } from "./config";
import { isCsrfValid } from "./csrf";
import { logger } from "./logger";
import { dedupRefresh } from "./refresh-shared";
import { createForbiddenResponse, createServiceUnavailableResponse } from "./response";
import {
  clearSessionCookies,
  getAccessToken,
  getRefreshToken,
  REFRESH_TOKEN_MAX_AGE,
  REFRESH_TOKEN_REUSED_CODE,
  type SessionTokens,
  setSessionCookies,
} from "./session";

const AUTH_REFRESH_PATH = "/api/v1/auth/refresh";
const AUTH_LOGOUT_PATH = "/api/v1/auth/logout";
const FORWARDED_REQUEST_HEADERS = [
  "accept",
  "accept-language",
  "authorization",
  "content-type",
  "idempotency-key",
  "x-idempotency-key",
  "x-request-id",
  "x-correlation-id",
];
const FORWARDED_RESPONSE_HEADERS = [
  "content-type",
  "content-disposition",
  "content-language",
  "cache-control",
  "etag",
  "expires",
  "last-modified",
  "location",
  "x-total-count",
  "link",
  "x-ratelimit-limit",
  "x-ratelimit-remaining",
  "x-ratelimit-reset",
  "retry-after",
  "x-request-id",
  "x-correlation-id",
];
const UPSTREAM_TIMEOUT_MS = 10_000;

type RefreshOutcome =
  | { ok: true; tokens: SessionTokens }
  | { ok: false; shouldClearCookies: boolean };

async function requestRefresh(refreshToken: string): Promise<RefreshOutcome> {
  const target = new URL(AUTH_REFRESH_PATH, serverConfig.serviceApiUrl);
  let upstream: Response;
  try {
    upstream = await fetch(target, {
      method: "POST",
      headers: { "content-type": jsonContentType },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
  } catch {
    return { ok: false, shouldClearCookies: false };
  }
  if (upstream.status !== StatusCodes.OK) {
    return { ok: false, shouldClearCookies: !(await isReuse(upstream)) };
  }
  try {
    const raw = (await upstream.json()) as {
      data?: { access_token?: string; refresh_token?: string };
    };
    const accessToken = raw.data?.access_token;
    const rotated = raw.data?.refresh_token;
    if (!accessToken || !rotated) return { ok: false, shouldClearCookies: true };
    return { ok: true, tokens: { accessToken, refreshToken: rotated } };
  } catch {
    return { ok: false, shouldClearCookies: true };
  }
}

async function isReuse(upstream: Response): Promise<boolean> {
  try {
    const raw = (await upstream.json()) as { error_code?: string };
    return raw.error_code === REFRESH_TOKEN_REUSED_CODE;
  } catch {
    return false;
  }
}

function refreshSession(refreshToken: string): Promise<RefreshOutcome> {
  return dedupRefresh(`proxy:${refreshToken}`, () => requestRefresh(refreshToken));
}

function buildTarget(path: string[], request: NextRequest): URL {
  const encoded = path.map((segment) => encodeURIComponent(segment)).join("/");
  const fullPath = `/api/v1/${encoded}`;
  const target = new URL(fullPath, serverConfig.serviceApiUrl);
  request.nextUrl.searchParams.forEach((value, key) => {
    target.searchParams.append(key, value);
  });
  return target;
}

function buildRequestHeaders(request: NextRequest): Headers {
  const headers = new Headers();
  for (const name of FORWARDED_REQUEST_HEADERS) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }
  const accessToken = getAccessToken(request);
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);
  return headers;
}

async function buildRequestBody(
  request: NextRequest,
  fullPath: string,
  refreshToken: string | undefined,
): Promise<BodyInit | undefined> {
  const needsRefreshToken =
    (fullPath === AUTH_REFRESH_PATH || fullPath === AUTH_LOGOUT_PATH) && !!refreshToken;
  if (!request.body) {
    if (needsRefreshToken) {
      return JSON.stringify(toSnakeCase({ refreshToken }));
    }
    return undefined;
  }
  if (!isJson(request.headers.get("content-type"))) {
    if (needsRefreshToken) {
      return JSON.stringify({ refresh_token: refreshToken });
    }
    return request.body;
  }
  const raw = await request.text();
  if (!raw) {
    if (needsRefreshToken) {
      return JSON.stringify({ refresh_token: refreshToken });
    }
    return undefined;
  }
  let parsed: unknown = raw;
  try {
    parsed = JSON.parse(raw);
  } catch {
    if (needsRefreshToken) {
      return JSON.stringify({ refresh_token: refreshToken });
    }
    return raw;
  }
  if (needsRefreshToken) {
    if (parsed && typeof parsed === "object" && !Array.isArray(parsed) && parsed !== null) {
      parsed = { ...(parsed as Record<string, unknown>), refreshToken };
    } else {
      parsed = { refreshToken };
    }
  }
  return JSON.stringify(toSnakeCase(parsed));
}

function buildResponseHeaders(upstream: Response): Headers {
  const headers = new Headers();
  for (const name of FORWARDED_RESPONSE_HEADERS) {
    const value = upstream.headers.get(name);
    if (value) headers.set(name, value);
  }
  return headers;
}

async function send(
  target: URL,
  method: string,
  headers: Headers,
  body: BodyInit | undefined,
): Promise<Response> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), UPSTREAM_TIMEOUT_MS);
  try {
    const init: RequestInit & { duplex?: string } = {
      method,
      headers,
      signal: controller.signal,
      body,
    };
    if (body instanceof ReadableStream) {
      init.duplex = "half";
    }
    return await fetch(target, init);
  } catch (error: unknown) {
    const isAbort = error instanceof DOMException && error.name === "AbortError";
    const detail = error instanceof Error ? `${error.name}: ${error.message}` : String(error);
    logger.error(
      { err: error, target: target.pathname, method },
      `proxy failed ${method} ${target.pathname}${isAbort ? " (timeout)" : ""} — ${detail}`,
    );
    return createServiceUnavailableResponse("Service is unavailable. Please try again later.");
  } finally {
    clearTimeout(timeout);
  }
}

export async function proxyRequest(request: NextRequest, path: string[]): Promise<Response> {
  const fullPath = `/api/v1/${path.join("/")}`;
  const refreshToken = getRefreshToken(request);
  if (!isCsrfValid(request)) {
    return createForbiddenResponse("CSRF token missing.");
  }
  const target = buildTarget(path, request);
  const headers = buildRequestHeaders(request);
  const body = await buildRequestBody(request, fullPath, refreshToken);
  if (
    body !== undefined &&
    (fullPath === AUTH_REFRESH_PATH || fullPath === AUTH_LOGOUT_PATH) &&
    refreshToken
  ) {
    headers.set("content-type", jsonContentType);
  }
  let upstream = await send(target, request.method, headers, body);
  let refreshed: SessionTokens | null = null;
  if (
    upstream.status === StatusCodes.UNAUTHORIZED &&
    fullPath !== AUTH_REFRESH_PATH &&
    fullPath !== "/api/v1/auth/login" &&
    refreshToken
  ) {
    const outcome = await refreshSession(refreshToken);
    if (outcome.ok) {
      refreshed = outcome.tokens;
      headers.set("Authorization", `Bearer ${refreshed.accessToken}`);
      upstream = await send(target, request.method, headers, body);
    } else if (outcome.shouldClearCookies) {
      const cleared = new NextResponse(upstream.body, {
        status: upstream.status,
        headers: buildResponseHeaders(upstream),
      });
      clearSessionCookies(cleared);
      return cleared;
    }
  }
  const contentTypeOut = upstream.headers.get("content-type") ?? "";
  let payload: unknown;
  let responseBody: BodyInit | null;
  if (isJson(contentTypeOut)) {
    const rawBody = await upstream.text().catch(() => "");
    if (!rawBody) {
      payload = null;
      responseBody = null;
    } else {
      try {
        payload = toCamelCase(JSON.parse(rawBody));
        responseBody = JSON.stringify(payload);
      } catch {
        logger.debug(
          { contentType: contentTypeOut, status: upstream.status },
          "non-JSON upstream response discarded",
        );
        payload = {
          success: false,
          message: "Upstream returned an invalid response.",
          errorCode: "ERR_SERVICE",
        };
        responseBody = JSON.stringify(payload);
      }
    }
  } else {
    responseBody = upstream.body;
  }
  const response = new NextResponse(responseBody, {
    status: upstream.status,
    headers: buildResponseHeaders(upstream),
  });
  if (fullPath === "/api/v1/auth/login" || fullPath === "/api/v1/auth/refresh") {
    if (upstream.status === StatusCodes.OK) {
      const rawData = (payload as { data?: Record<string, unknown> })?.data as
        | {
            accessToken?: string;
            refreshToken?: string;
            access_token?: string;
            refresh_token?: string;
          }
        | undefined;
      const accessToken = rawData?.accessToken ?? rawData?.access_token;
      const rotatedRefreshToken = rawData?.refreshToken ?? rawData?.refresh_token;
      if (accessToken && rotatedRefreshToken) {
        setSessionCookies(response, {
          accessToken,
          refreshToken: rotatedRefreshToken,
        });
        response.cookies.set(csrfKey, crypto.randomUUID(), {
          httpOnly: false,
          secure: process.env.NODE_ENV === "production",
          sameSite: "lax",
          path: "/",
          maxAge: REFRESH_TOKEN_MAX_AGE,
        });
      }
    }
  } else if (fullPath === "/api/v1/auth/logout") {
    clearSessionCookies(response);
  }
  if (refreshed && upstream.status === StatusCodes.OK) {
    setSessionCookies(response, refreshed);
  }
  return response;
}
